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
	regexDivBody = regexp.MustCompile(`(?is)<div[^>]+id=["']divBody["'][^>]*>(.*?)</div>`)
)

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
