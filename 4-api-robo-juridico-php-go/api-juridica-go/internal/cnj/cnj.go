package cnj

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

var (
	regexCNJ     = regexp.MustCompile(`^(\d{7})-(\d{2})\.(\d{4})\.(\d)\.(\d{1,2})\.(\d{4})$`)
	regexDigitos = regexp.MustCompile(`\D`)

	SegmentosJustica = map[int]string{
		1: "Supremo Tribunal Federal",
		2: "Conselho Nacional de Justiça",
		3: "Superior Tribunal de Justiça",
		4: "Justiça Federal",
		5: "Justiça do Trabalho",
		6: "Justiça Eleitoral",
		7: "Justiça Militar da União",
		8: "Justiça Estadual",
		9: "Justiça Militar Estadual",
	}

	TribunaisTJ = map[int]string{
		1: "TJAC", 2: "TJAL", 3: "TJAP", 4: "TJAM", 5: "TJBA",
		6: "TJCE", 7: "TJCE", 8: "TJES", 9: "TJGO", 10: "TJMA",
		11: "TJMT", 12: "TJMS", 13: "TJMG", 14: "TJPA", 15: "TJPB",
		16: "TJPR", 17: "TJPE", 18: "TJPI", 19: "TJRJ", 20: "TJRN",
		21: "TJRS", 22: "TJRO", 23: "TJRR", 24: "TJSC", 25: "TJSE",
		26: "TJSP", 27: "TJTO",
	}

	TribunaisTRF = map[int]string{
		1: "TRF1", 2: "TRF2", 3: "TRF3", 4: "TRF4", 5: "TRF5", 6: "TRF6",
	}

	UFsPorCodigo = map[int]string{
		1: "AC", 2: "AL", 3: "AP", 4: "AM", 5: "BA", 6: "CE", 7: "DF", 8: "ES", 9: "GO",
		10: "MA", 11: "MT", 12: "MS", 13: "MG", 14: "PA", 15: "PB", 16: "PR", 17: "PE",
		18: "PI", 19: "RJ", 20: "RN", 21: "RS", 22: "RO", 23: "RR", 24: "SC", 25: "SE",
		26: "SP", 27: "TO",
	}
)

func init() {
	// Correção do código 7 da Justiça Estadual
	TribunaisTJ[7] = "TJDFT"
}

// ProcessoCNJ representa os dados extraídos do número de um processo no padrão CNJ.
type ProcessoCNJ struct {
	NumeroOriginal string `json:"numero_original"`
	Formatado      string `json:"formatado"`
	Digitos        string `json:"digitos"`
	Valido         bool   `json:"valido"`
	Sequencial     string `json:"sequencial,omitempty"`
	Digito         string `json:"digito,omitempty"`
	Ano            int    `json:"ano,omitempty"`
	Segmento       int    `json:"segmento,omitempty"`
	TribunalCodigo int    `json:"tribunal_codigo,omitempty"`
	OrigemCodigo   int    `json:"origem_codigo,omitempty"`
	Tribunal       string `json:"tribunal,omitempty"`
	NomeSegmento   string `json:"segmento_nome,omitempty"`
	IndiceDatajud  string `json:"indice_datajud,omitempty"`
}

// ApenasDigitos remove tudo que não for dígito numérico da string.
func ApenasDigitos(numero string) string {
	return regexDigitos.ReplaceAllString(numero, "")
}

// Formatar garante o formato NNNNNNN-DD.AAAA.J.TR.OOOO a partir de 20 dígitos.
func Formatar(numero string) string {
	d := ApenasDigitos(numero)
	if len(d) == 20 {
		return fmt.Sprintf("%s-%s.%s.%s.%s.%s",
			d[0:7], d[7:9], d[9:13], d[13:14], d[14:16], d[16:20])
	}
	return strings.TrimSpace(numero)
}

// DigitoConfere calcula o dígito verificador módulo 97 conforme Resolução CNJ 65/2008.
func DigitoConfere(digitos string) bool {
	if len(digitos) != 20 {
		return false
	}
	dvInformado, err := strconv.Atoi(digitos[7:9])
	if err != nil {
		return false
	}

	// Base = sequencial (7 dígitos) + ano/segmento/tribunal/origem (11 dígitos) + "00"
	baseStr := digitos[0:7] + digitos[9:20] + "00"
	num := new(big.Int)
	num.SetString(baseStr, 10)

	rem := new(big.Int).Mod(num, big.NewInt(97)).Int64()
	dvCalculado := int(98 - rem)

	return dvCalculado == dvInformado
}

// Parse analisa e valida o número CNJ informado.
func Parse(numero string) *ProcessoCNJ {
	formatado := Formatar(numero)
	d := regexDigitos.ReplaceAllString(formatado, "")

	p := &ProcessoCNJ{
		NumeroOriginal: strings.TrimSpace(numero),
		Formatado:      formatado,
		Digitos:        d,
		Valido:         false,
	}

	m := regexCNJ.FindStringSubmatch(formatado)
	if m == nil {
		return p
	}

	ano, _ := strconv.Atoi(m[3])
	segmento, _ := strconv.Atoi(m[4])
	tribunalCod, _ := strconv.Atoi(m[5])
	origem, _ := strconv.Atoi(m[6])

	p.Sequencial = m[1]
	p.Digito = m[2]
	p.Ano = ano
	p.Segmento = segmento
	p.TribunalCodigo = tribunalCod
	p.OrigemCodigo = origem
	p.NomeSegmento = SegmentosJustica[segmento]
	if p.NomeSegmento == "" {
		p.NomeSegmento = "Desconhecido"
	}

	p.Tribunal = identificarTribunal(segmento, tribunalCod)
	p.IndiceDatajud = fmt.Sprintf("api_publica_%s", strings.ToLower(p.Tribunal))
	p.Valido = DigitoConfere(d)

	return p
}

func identificarTribunal(segmento, cod int) string {
	switch segmento {
	case 1:
		return "STF"
	case 2:
		return "CNJ"
	case 3:
		return "STJ"
	case 4:
		if t, ok := TribunaisTRF[cod]; ok {
			return t
		}
		return fmt.Sprintf("TRF%d", cod)
	case 5:
		if cod == 0 {
			return "TST"
		}
		return fmt.Sprintf("TRT%d", cod)
	case 6:
		if cod == 0 {
			return "TSE"
		}
		if uf, ok := UFsPorCodigo[cod]; ok {
			return fmt.Sprintf("TRE-%s", uf)
		}
		return fmt.Sprintf("TRE-%d", cod)
	case 7:
		return "STM"
	case 8:
		if t, ok := TribunaisTJ[cod]; ok {
			return t
		}
		return fmt.Sprintf("TJ%02d", cod)
	case 9:
		if uf, ok := UFsPorCodigo[cod]; ok {
			return fmt.Sprintf("TJM-%s", uf)
		}
		return fmt.Sprintf("TJM-%d", cod)
	default:
		return fmt.Sprintf("SEG%d-TR%d", segmento, cod)
	}
}
