package calculator

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ImportarPlanilha extrai parcelas de um arquivo Excel (.xlsx) ou CSV (.csv)
func ImportarPlanilha(nomeArquivo string, conteudo []byte) ([]ParcelaItem, error) {
	nomeLower := strings.ToLower(nomeArquivo)

	if strings.HasSuffix(nomeLower, ".xlsx") {
		return importarXLSX(conteudo)
	}

	return importarCSV(conteudo)
}

// importarCSV processa arquivos delimitados por ponto-e-vírgula, vírgula ou tabulação
func importarCSV(conteudo []byte) ([]ParcelaItem, error) {
	// Detecta delimitador comum
	delimitador := ';'
	primeiraLinha := string(bytes.Split(conteudo, []byte("\n"))[0])
	if strings.Count(primeiraLinha, "\t") > strings.Count(primeiraLinha, ";") {
		delimitador = '\t'
	} else if strings.Count(primeiraLinha, ",") > strings.Count(primeiraLinha, ";") {
		delimitador = ','
	}

	r := csv.NewReader(bytes.NewReader(conteudo))
	r.Comma = delimitador
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	linhas, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo CSV: %w", err)
	}

	if len(linhas) == 0 {
		return nil, fmt.Errorf("arquivo CSV vazio")
	}

	return processarLinhasTabela(linhas)
}

// importarXLSX extrai os dados diretamente do XML do arquivo .xlsx usando archive/zip nativo do Go
func importarXLSX(conteudo []byte) ([]ParcelaItem, error) {
	zr, err := zip.NewReader(bytes.NewReader(conteudo), int64(len(conteudo)))
	if err != nil {
		return nil, fmt.Errorf("arquivo .xlsx invalido ou corrompido: %w", err)
	}

	// 1. Carrega sharedStrings.xml (textos compartilhados do Excel)
	var sharedStrings []string
	for _, f := range zr.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err == nil {
				sharedStrings = extrairSharedStrings(rc)
				rc.Close()
			}
			break
		}
	}

	// 2. Localiza a primeira planilha (xl/worksheets/sheet1.xml)
	var sheetFile *zip.File
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			sheetFile = f
			break
		}
	}

	if sheetFile == nil {
		return nil, fmt.Errorf("nenhuma planilha encontrada no arquivo .xlsx")
	}

	rc, err := sheetFile.Open()
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir planilha: %w", err)
	}
	defer rc.Close()

	linhas := extrairLinhasSheetXML(rc, sharedStrings)
	return processarLinhasTabela(linhas)
}

// processarLinhasTabela mapeia matriz de strings para slice de ParcelaItem
func processarLinhasTabela(linhas [][]string) ([]ParcelaItem, error) {
	if len(linhas) == 0 {
		return nil, fmt.Errorf("planilha sem linhas de dados")
	}

	// Identifica cabeçalhos e colunas
	idxData := -1
	idxDescricao := -1
	idxValor := -1
	idxTipo := -1

	startRow := 0
	cabecalho := linhas[0]

	for i, c := range cabecalho {
		cNorm := strings.ToLower(strings.TrimSpace(c))
		if strings.Contains(cNorm, "data") || strings.Contains(cNorm, "vencto") || strings.Contains(cNorm, "vencimento") {
			idxData = i
		} else if strings.Contains(cNorm, "desc") || strings.Contains(cNorm, "historico") || strings.Contains(cNorm, "parcela") {
			idxDescricao = i
		} else if strings.Contains(cNorm, "valor") || strings.Contains(cNorm, "original") || strings.Contains(cNorm, "nominal") {
			idxValor = i
		} else if strings.Contains(cNorm, "tipo") || strings.Contains(cNorm, "natureza") {
			idxTipo = i
		}
	}

	// Se encontrou cabeçalho, começa na linha 1; senão assume posições padrão (0: Data, 1: Descrição, 2: Valor)
	if idxData != -1 || idxValor != -1 {
		startRow = 1
	}

	if idxData == -1 {
		idxData = 0
	}
	if idxDescricao == -1 {
		idxDescricao = 1
	}
	if idxValor == -1 {
		idxValor = 2
	}

	var parcelas []ParcelaItem
	numItem := 1

	for r := startRow; r < len(linhas); r++ {
		linha := linhas[r]
		if len(linha) == 0 {
			continue
		}

		// Garante índices válidos
		strData := ""
		if idxData < len(linha) {
			strData = strings.TrimSpace(linha[idxData])
		}

		strDesc := fmt.Sprintf("Item %d", numItem)
		if idxDescricao < len(linha) && strings.TrimSpace(linha[idxDescricao]) != "" {
			strDesc = strings.TrimSpace(linha[idxDescricao])
		}

		strValor := ""
		if idxValor < len(linha) {
			strValor = strings.TrimSpace(linha[idxValor])
		}

		strTipo := "DEBITO"
		if idxTipo != -1 && idxTipo < len(linha) {
			t := strings.ToUpper(strings.TrimSpace(linha[idxTipo]))
			if strings.Contains(t, "CUSTA") {
				strTipo = "CUSTA_PROCESSUAL"
			} else if strings.Contains(t, "DESPESA") {
				strTipo = "DESPESA_PROCESSUAL"
			} else if strings.Contains(t, "PAG") || strings.Contains(t, "ABAT") {
				strTipo = "PAGAMENTO_PARCIAL"
			}
		}

		if strData == "" && strValor == "" {
			continue
		}

		dtFormatada := normalizarDataString(strData)
		if dtFormatada == "" {
			continue
		}

		valDecimal, err := normalizarValorString(strValor)
		if err != nil || valDecimal.IsZero() {
			continue
		}

		parcelas = append(parcelas, ParcelaItem{
			NumeroItem:     numItem,
			Descricao:      strDesc,
			Tipo:           TipoParcela(strTipo),
			DataVencimento: dtFormatada,
			ValorOriginal:  valDecimal,
		})
		numItem++
	}

	if len(parcelas) == 0 {
		return nil, fmt.Errorf("nenhuma parcela valida identificada na planilha")
	}

	return parcelas, nil
}

// normalizarDataString converte DD/MM/AAAA, AAAA-MM-DD ou dias seriais do Excel para AAAA-MM-DD
func normalizarDataString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	// Caso serial do Excel (ex: "44941" para 15/01/2023)
	if num, err := strconv.Atoi(s); err == nil && num > 30000 && num < 60000 {
		// Marco inicial do Excel: 30/12/1899
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		d := base.AddDate(0, 0, num)
		return d.Format("2006-01-02")
	}

	dt, err := ParseDataFlexivel(s)
	if err == nil && !dt.IsZero() {
		return dt.Format("2006-01-02")
	}

	return ""
}

// normalizarValorString limpa R$, pontos e converte vírgula para ponto decimal
func normalizarValorString(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "R$", "")
	s = strings.ReplaceAll(s, " ", "")

	// Se tiver formato brasileiro 1.500,50
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}

	return decimal.NewFromString(s)
}

// extrairSharedStrings lê as strings da tabela de compartilhamento do XLSX
func extrairSharedStrings(r io.Reader) []string {
	var stringsList []string
	decoder := xml.NewDecoder(r)
	var inT bool
	var sb strings.Builder

	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "t" {
				inT = true
				sb.Reset()
			}
		case xml.EndElement:
			if se.Name.Local == "t" {
				inT = false
				stringsList = append(stringsList, sb.String())
			}
		case xml.CharData:
			if inT {
				sb.Write(se)
			}
		}
	}
	return stringsList
}

// extrairLinhasSheetXML decodifica as linhas e células do XML da planilha
func extrairLinhasSheetXML(r io.Reader, sharedStrings []string) [][]string {
	var matriz [][]string
	decoder := xml.NewDecoder(r)

	var linhaAtual []string
	var celulaAtual strings.Builder
	var tipoCelula string
	var inV bool

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch el := tok.(type) {
		case xml.StartElement:
			if el.Name.Local == "row" {
				linhaAtual = []string{}
			} else if el.Name.Local == "c" {
				celulaAtual.Reset()
				tipoCelula = ""
				for _, attr := range el.Attr {
					if attr.Name.Local == "t" {
						tipoCelula = attr.Value
					}
				}
			} else if el.Name.Local == "v" {
				inV = true
			}
		case xml.EndElement:
			if el.Name.Local == "v" {
				inV = false
				val := celulaAtual.String()
				if tipoCelula == "s" {
					if idx, err := strconv.Atoi(val); err == nil && idx >= 0 && idx < len(sharedStrings) {
						val = sharedStrings[idx]
					}
				}
				linhaAtual = append(linhaAtual, val)
			} else if el.Name.Local == "row" {
				matriz = append(matriz, linhaAtual)
			}
		case xml.CharData:
			if inV {
				celulaAtual.Write(el)
			}
		}
	}
	return matriz
}
