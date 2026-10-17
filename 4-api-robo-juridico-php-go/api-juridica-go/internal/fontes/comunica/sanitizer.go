package comunica

import (
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
)

var (
	regexStyle  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	regexScript = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	regexTag    = regexp.MustCompile(`(?s)<[^>]+>`)
	regexEspacos = regexp.MustCompile(`\s+`)
	regexDivBody     = regexp.MustCompile(`(?is)<div[^>]+id=["']divBody["'][^>]*>(.*?)</div>`)
	regexComarcaPub  = regexp.MustCompile(`(?i)Comarca\s+de\s+([A-ZÀ-Úa-zà-ú\s]+?)(?:\s*[/\-\n,.]|\s+Avenida|\s+Rua|\s+Praça|$)`)
	regexPoloAtivo   = regexp.MustCompile(`(?i)(?:AUTOR(?:A|ES)?|REQUERENTE(?:S)?|EXEQUENTE(?:S)?|RECORRENTE(?:S)?|IMPETRANTE(?:S)?|POLO\s+ATIVO)\s*:\s*([A-ZÀ-Ú0-9][^\n\r;|]{2,80}?)(?:\s+(?:CPF|CNPJ|R[ÉE]U|REQUERID|EXECUTAD|RECORRID|IMPETRAD|POLO\s+PASSIVO|ADVOGAD|REPRESENTANTE|DESPACHO|DECIS[ÃA]O|SENTEN[ÇC]A)|$)`)
	regexPoloPassivo = regexp.MustCompile(`(?i)(?:R[ÉE]U(?:S)?|REQUERID[OA](?:S)?|EXECUTAD[OA](?:S)?|RECORRID[OA](?:S)?|IMPETRAD[OA](?:S)?|POLO\s+PASSIVO)\s*:\s*([A-ZÀ-Ú0-9][^\n\r;|]{2,80}?)(?:\s+(?:CPF|CNPJ|AUTOR|REQUERENT|EXEQUENT|ADVOGAD|TERCEIR|REPRESENTANTE|DESPACHO|DECIS[ÃA]O|SENTEN[ÇC]A)|$)`)
	regexParteDoc    = regexp.MustCompile(`(?:\(\d+\)|PROCESSO\s+N[º°]?:\s*[\d\.\-]+.*?)?\s*([A-ZÁÉÍÓÚÂÊÔÃÕÇ][A-ZÁÉÍÓÚÂÊÔÃÕÇ\s\.\-&]{3,80}?)\s+(?:CPF|CNPJ)\s*:\s*[\d\.\-\/]+`)
	regexValorCausa  = regexp.MustCompile(`(?i)(?:valor\s+(?:da\s+causa|atribu[íi]do\s+[àa]\s+causa|da\s+a[çc][ãa]o|da\s+execu[çc][ãa]o|da\s+condena[çc][ãa]o|do\s+d[ée]bito)|causa\s+avaliada\s+em)\s*[:\-]?\s*R\$\s*([\d\.]+\,\d{2})`)
	regexClasseCodigoPrefixo   = regexp.MustCompile(`(?i)^.*?\(\s*\d{2,6}\s*\)\s*`)
	regexRepresentantesSufixo = regexp.MustCompile(`(?i)\s+(?:REPRESENTANTES?\s+POLO\s+(?:ATIVO|PASSIVO)\s*:|REPRESENTANTES?\s*:|ADVOGAD[OA]S?\s*:|PROCURADOR(?:ES)?\s*:).*$`)
)

var comarcasTJMG = map[int]string{
	0:   "Belo Horizonte",
	24:  "Belo Horizonte",
	27:  "Betim",
	40:  "Araxá",
	79:  "Contagem",
	105: "Governador Valadares",
	114: "Ibirité",
	133: "Carangola",
	134: "Caratinga",
	144: "Carmo do Rio Claro",
	145: "Juiz de Fora",
	188: "Nova Lima",
	210: "Pedro Leopoldo",
	223: "Divinópolis",
	261: "Formiga",
	313: "Ipatinga",
	392: "Malacacheta",
	433: "Montes Claros",
	439: "Muriaé",
	461: "Ouro Preto",
	479: "Passos",
	480: "Patos de Minas",
	508: "Pirapora",
	517: "Poços de Caldas",
	518: "Poços de Caldas",
	520: "Pompéu",
	525: "Pouso Alegre",
	546: "Ribeirão das Neves",
	567: "Sabará",
	578: "Santa Luzia",
	672: "Sete Lagoas",
	686: "Teófilo Otoni",
	701: "Uberaba",
	702: "Uberlândia",
	707: "Varginha",
	713: "Viçosa",
	878: "Camanducaia",
}

func FormatarID(raw any) string {
	switch v := raw.(type) {
	case float64:
		return strconvFormatFloat(v)
	case string:
		return strings.TrimSpace(v)
	default:
		if raw == nil {
			return ""
		}
		return strings.TrimSpace(html.UnescapeString(strings.ReplaceAll(
			func() string { return "" }(), "", "")))
	}
}

func strconvFormatFloat(v float64) string {
	// Formata inteiro sem notação científica
	const digits = "0123456789"
	n := int64(v)
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n%10]
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

func ExtrairComarcaDoTexto(texto string) string {
	if m := regexComarcaPub.FindStringSubmatch(texto); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

// InferirComarcaPorOrigem tenta extrair a comarca do texto/órgão e, se vazio, usa o código de origem CNJ (OOOO)
func InferirComarcaPorOrigem(tribunal string, origemCodigo int, orgao, texto string) string {
	if c := ExtrairComarcaDoTexto(texto); c != "" {
		return c
	}
	if c := ExtrairComarcaDoTexto(orgao); c != "" {
		return c
	}
	if strings.EqualFold(tribunal, "TJMG") {
		if nome, ok := comarcasTJMG[origemCodigo]; ok {
			return nome
		}
	}
	return ""
}

// InferirGrau deduz o grau de jurisdição (G1, G2 ou TR) a partir do nome do órgão julgador e classe
func InferirGrau(orgao, classe string) string {
	combinado := strings.ToUpper(orgao + " " + classe)
	if strings.Contains(combinado, "TURMA RECURSAL") || strings.Contains(combinado, "RECURSO INOMINADO") || strings.Contains(combinado, " TR ") {
		return "TR"
	}
	if strings.Contains(combinado, "CÂMARA") || strings.Contains(combinado, "CAMARA") ||
		strings.Contains(combinado, "VICE-PRESID") || strings.Contains(combinado, "PRESIDÊNCIA") ||
		strings.Contains(combinado, "CACIV") || strings.Contains(combinado, "CACRIM") ||
		strings.Contains(combinado, "RECURSO ESPECIAL") || strings.Contains(combinado, "RECURSO EXTRAORDIN") ||
		strings.Contains(combinado, "AGRAVO DE INSTRUMENTO") || strings.Contains(combinado, "APELAÇÃO") {
		return "G2"
	}
	if strings.TrimSpace(combinado) != "" {
		return "G1"
	}
	return ""
}

// ExtrairValorCausaDoTexto captura valores monetários explícitos de valor da causa/execução no texto da publicação
func ExtrairValorCausaDoTexto(texto string) *float64 {
	m := regexValorCausa.FindStringSubmatch(texto)
	if len(m) < 2 {
		return nil
	}
	limpo := strings.ReplaceAll(m[1], ".", "")
	limpo = strings.ReplaceAll(limpo, ",", ".")
	var val float64
	var inteiro, frac int64
	partes := strings.Split(limpo, ".")
	if len(partes) == 2 {
		for _, ch := range partes[0] {
			if ch >= '0' && ch <= '9' {
				inteiro = inteiro*10 + int64(ch-'0')
			}
		}
		for _, ch := range partes[1] {
			if ch >= '0' && ch <= '9' {
				frac = frac*10 + int64(ch-'0')
			}
		}
		val = float64(inteiro) + float64(frac)/100.0
		if val > 0 {
			return &val
		}
	}
	return nil
}

// NormalizarCaixaAlta corrige textos em CAIXA ALTA que vieram do tribunal com acentos em minúsculo (ex: "APELAçãO CíVEL" -> "APELAÇÃO CÍVEL")
func NormalizarCaixaAlta(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	maiusculas := 0
	minusculasASCII := 0
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			maiusculas++
		} else if r >= 'a' && r <= 'z' {
			minusculasASCII++
		}
	}
	if maiusculas > 2 && minusculasASCII == 0 {
		return strings.ToUpper(s)
	}
	return s
}

func ExtrairPartesDoTexto(texto string, existentes []domain.Parte) []domain.Parte {
	resultado := make([]domain.Parte, 0, len(existentes))
	conhecidos := make(map[string]bool)
	temAtivo := false
	temPassivo := false

	for _, p := range existentes {
		if p.Advogados == nil {
			p.Advogados = []domain.Advogado{}
		}
		resultado = append(resultado, p)
		conhecidos[strings.ToUpper(strings.TrimSpace(p.Nome))] = true
		if p.Polo == domain.PoloAtivo {
			temAtivo = true
		}
		if p.Polo == domain.PoloPassivo {
			temPassivo = true
		}
	}

	adicionar := func(nome string, polo domain.Polo) {
		nome = LimparNomeParte(nome)
		if !EhNomeParteValido(nome) {
			return
		}
		chave := strings.ToUpper(nome)
		if conhecidos[chave] {
			return
		}
		conhecidos[chave] = true
		resultado = append(resultado, domain.Parte{
			Nome:      nome,
			Polo:      polo,
			Advogados: []domain.Advogado{},
		})
		if polo == domain.PoloAtivo {
			temAtivo = true
		} else if polo == domain.PoloPassivo {
			temPassivo = true
		}
	}

	if !temAtivo {
		if m := regexPoloAtivo.FindStringSubmatch(texto); len(m) > 1 {
			adicionar(m[1], domain.PoloAtivo)
		}
	}
	if !temPassivo {
		if m := regexPoloPassivo.FindStringSubmatch(texto); len(m) > 1 {
			adicionar(m[1], domain.PoloPassivo)
		}
	}

	// Padrão PJe (ex: TJMG): "<NOME> CPF: 000.000.000-00 <NOME RÉU> CPF/CNPJ: 00.000.000/0001-00"
	matches := regexParteDoc.FindAllStringSubmatch(texto, -1)
	for idx, m := range matches {
		if len(m) < 2 {
			continue
		}
		candidato := LimparNomeParte(m[1])
		if idx == 0 && !temAtivo {
			adicionar(candidato, domain.PoloAtivo)
		} else if idx >= 1 && !temPassivo {
			adicionar(candidato, domain.PoloPassivo)
		}
	}

	return resultado
}

func EhNomeParteValido(nome string) bool {
	nome = strings.TrimSpace(nome)
	if len(nome) < 3 || len(nome) > 85 {
		return false
	}
	upper := strings.ToUpper(nome)
	// Termos puramente de sufixo empresarial isolado (que não formam um nome por si só)
	sufixosIsolados := map[string]bool{
		"LTDA": true, "LTDA.": true, "S.A.": true, "S/A": true, "SA": true,
		"ME": true, "EPP": true, "EIRELI": true, "MEI": true, "S/S": true, "SS": true,
	}
	if sufixosIsolados[upper] || sufixosIsolados[strings.Trim(upper, " ./-")] {
		return false
	}

	palavrasProibidas := []string{
		"DESPACHO", "DECISÃO", "DECISAO", "SENTENÇA", "SENTENCA",
		"JUIZ FEDERAL", "JUIZ DE DIREITO", "JUIZA FEDERAL", "JUÍZA",
		"CUMPRA-SE", "INTIMEM-SE", "INTIME-SE", "RETORNEM CONCLUSOS",
		"CONCLUSOS", "MANIFESTEM", "SEGUIMENTO DO FEITO", "PRAZO DE",
		"SUBSEÇÃO JUDICIÁRIA", "SUBSECAO JUDICIARIA", "VARA FEDERAL",
		"VARA CÍVEL", "VARA CIVEL", "COMARCA DE", "DIAS, E REQUEIRAM",
		"CONSIDERANDO A REMESSA", "ASSINADO ELETRONICAMENTE", "ASSINATURA ELETRÔNICA",
		"PODER JUDICIÁRIO", "PODER JUDICIARIO", "JUSTIÇA FEDERAL", "JUSTICA FEDERAL",
		"MANDADO DE SEGURANÇA", "MANDADO DE SEGURANCA",
	}
	for _, p := range palavrasProibidas {
		if strings.Contains(upper, p) {
			return false
		}
	}
	return true
}

func LimparNomeParte(s string) string {
	s = strings.TrimSpace(s)
	// Se contiver código numérico de classe entre parênteses antes do nome (ex: "...(14695) MARIA..."), pega após o código
	if m := regexClasseCodigoPrefixo.FindStringIndex(s); len(m) == 2 {
		s = strings.TrimSpace(s[m[1]:])
	}
	// Trunca sufixo de representantes ou advogados (ex: "... REPRESENTANTES POLO ATIVO: GISELE...")
	if m := regexRepresentantesSufixo.FindStringIndex(s); len(m) == 2 {
		s = strings.TrimSpace(s[:m[0]])
	}
	s = strings.TrimPrefix(s, "e outros ")
	s = strings.TrimPrefix(s, "E OUTROS ")
	s = strings.TrimSuffix(s, " e outros")
	s = strings.TrimSuffix(s, " E OUTROS")
	s = strings.TrimSuffix(s, " -")
	s = strings.TrimSuffix(s, ":")
	s = strings.TrimSpace(s)
	return s
}

func SemHTML(textoHTML string) string {
	limpo := regexStyle.ReplaceAllString(textoHTML, " ")
	limpo = regexScript.ReplaceAllString(limpo, " ")
	limpo = regexTag.ReplaceAllString(limpo, " ")
	limpo = html.UnescapeString(limpo)
	return strings.TrimSpace(regexEspacos.ReplaceAllString(limpo, " "))
}

func ExtrairTeor(textoHTML string) string {
	if textoHTML == "" {
		return ""
	}
	m := regexDivBody.FindStringSubmatch(textoHTML)
	if len(m) > 1 {
		return SemHTML(m[1])
	}
	return SemHTML(textoHTML)
}

func ClassificarTipo(tipoComunicacao string) domain.TipoMovimentacao {
	t := strings.ToLower(strings.TrimSpace(tipoComunicacao))
	switch {
	case strings.Contains(t, "intimação") || strings.Contains(t, "intimacao"):
		return domain.TipoIntimacao
	case strings.Contains(t, "despacho"):
		return domain.TipoDespacho
	case strings.Contains(t, "decisão") || strings.Contains(t, "decisao"):
		return domain.TipoDecisao
	case strings.Contains(t, "sentença") || strings.Contains(t, "sentenca") || strings.Contains(t, "acórdão") || strings.Contains(t, "acordao"):
		return domain.TipoSentenca
	case strings.Contains(t, "edital") || strings.Contains(t, "publicação") || strings.Contains(t, "publicacao"):
		return domain.TipoPublicacaoDJE
	default:
		return domain.TipoOutro
	}
}

func ParseDataDJEN(dataStr string) (time.Time, string) {
	if dataStr == "" {
		return time.Time{}, ""
	}
	if len(dataStr) > 10 {
		dataStr = dataStr[:10]
	}

	for _, layout := range []string{"2006-01-02", "02/01/2006"} {
		if t, err := time.Parse(layout, dataStr); err == nil {
			return t, t.Format("02/01/2006")
		}
	}
	return time.Time{}, ""
}
