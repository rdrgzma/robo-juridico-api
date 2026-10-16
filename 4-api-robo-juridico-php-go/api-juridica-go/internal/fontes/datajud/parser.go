package datajud

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
)

var (
	regexComarca          = regexp.MustCompile(`(?i)(?:da\s+)?Comarca\s+de\s+([^/\-\n,]+)`)
	regexJurisdicaoCidades = regexp.MustCompile(`(?i)Jurisdição\s+Exclusiva,\s+de\s+(.+?)$`)
)

// CorrigirMojibake corrige strings que sofreram dupla codificação UTF-8 / ISO-8859-1 no envio do tribunal
func CorrigirMojibake(s string) string {
	if !strings.Contains(s, "Ã") && !strings.Contains(s, "Â") {
		return s
	}
	replacer := strings.NewReplacer(
		"Ã¡", "á", "Ã ", "à", "Ã¢", "â", "Ã£", "ã",
		"Ã©", "é", "Ãª", "ê", "Ã­", "í", "Ã³", "ó",
		"Ã´", "ô", "Ãµ", "õ", "Ãº", "ú", "Ã§", "ç",
		"Ã\u0081", "Á", "Ã\u0089", "É", "Ã\u008d", "Í", "Ã\u0093", "Ó",
		"Ã\u009a", "Ú", "Ã\u0083", "Ã", "Ã\u0095", "Õ", "Ã\u0087", "Ç",
	)
	return replacer.Replace(s)
}

// ParseDataDataJud interpreta datas do DataJud nos formatos ISO/RFC3339 e compacto (YYYYMMDDHHmmss)
func ParseDataDataJud(dataStr string) (time.Time, bool) {
	dataStr = strings.TrimSpace(dataStr)
	if dataStr == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05",
		"20060102150405",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"20060102",
	}
	for _, layout := range layouts {
		alvo := dataStr
		if layout == "2006-01-02T15:04:05" && strings.Contains(alvo, ".") {
			alvo = strings.Split(alvo, ".")[0]
		}
		if t, err := time.Parse(layout, alvo); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func ClassificarTipo(descricao string) domain.TipoMovimentacao {
	d := strings.ToLower(descricao)
	if strings.Contains(d, "sentença") || strings.Contains(d, "sentenca") || strings.Contains(d, "julgamento") || strings.Contains(d, "provimento") {
		return domain.TipoSentenca
	}
	if strings.Contains(d, "decisão") || strings.Contains(d, "decisao") || strings.Contains(d, "mérito") || strings.Contains(d, "merito") {
		return domain.TipoDecisao
	}
	if strings.Contains(d, "despacho") || strings.Contains(d, "conclusão") || strings.Contains(d, "conclusao") {
		return domain.TipoDespacho
	}
	if strings.Contains(d, "audiência") || strings.Contains(d, "audiencia") {
		return domain.TipoAudiencia
	}
	if strings.Contains(d, "intimação") || strings.Contains(d, "intimacao") || strings.Contains(d, "citação") || strings.Contains(d, "citacao") {
		return domain.TipoIntimacao
	}
	if strings.Contains(d, "diário") || strings.Contains(d, "publicação") || strings.Contains(d, "djen") {
		return domain.TipoPublicacaoDJE
	}
	return domain.TipoOutro
}

func FormatarDinheiroBR(v *float64) string {
	if v == nil {
		return ""
	}
	val := fmt.Sprintf("%.2f", *v)
	partes := strings.Split(val, ".")
	inteiro := partes[0]
	decimal := partes[1]

	var res []string
	for i, c := range reverse(inteiro) {
		if i > 0 && i%3 == 0 {
			res = append(res, ".")
		}
		res = append(res, string(c))
	}
	return fmt.Sprintf("R$ %s,%s", reverse(strings.Join(res, "")), decimal)
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ParseMovimentos converte e deduplica movimentos com hash MD5 de descrição e data
func ParseMovimentos(brutos []map[string]any) []domain.Movimentacao {
	var movs []domain.Movimentacao
	usados := make(map[string]int)

	for _, m := range brutos {
		dataStr, _ := m["dataHora"].(string)
		desc, _ := m["nome"].(string)
		if desc == "" {
			desc, _ = m["descricao"].(string)
		}
		desc = CorrigirMojibake(strings.TrimSpace(desc))
		if dataStr == "" || desc == "" {
			continue
		}

		t, ok := ParseDataDataJud(dataStr)
		if !ok {
			continue
		}

		cod := "0"
		if c, ok := m["codigo"]; ok {
			cod = fmt.Sprintf("%v", c)
		}

		h := md5.Sum([]byte(truncate(desc, 100)))
		resumo := hex.EncodeToString(h[:])[:8]

		prefixoData := dataStr
		if len(prefixoData) > 19 {
			prefixoData = prefixoData[:19]
		}
		idExterno := fmt.Sprintf("%s_%s_%s", cod, prefixoData, resumo)

		usados[idExterno]++
		if vezes := usados[idExterno]; vezes > 1 {
			idExterno = fmt.Sprintf("%s_%d", idExterno, vezes)
		}

		var complementos []string
		if comps, ok := m["complementosTabelados"].([]any); ok {
			for _, c := range comps {
				if cmap, ok := c.(map[string]any); ok {
					nome, _ := cmap["nome"].(string)
					descC, _ := cmap["descricao"].(string)
					nome = CorrigirMojibake(strings.TrimSpace(nome))
					descC = CorrigirMojibake(strings.TrimSpace(descC))

					// Se descC for apenas um identificador interno em snake_case (ex: tipo_de_peticao), usa apenas o valor legível (nome)
					if nome != "" && (descC == "" || strings.Contains(descC, "_") || strings.EqualFold(nome, descC)) {
						complementos = append(complementos, nome)
					} else if nome != "" && descC != "" {
						complementos = append(complementos, fmt.Sprintf("%s (%s)", nome, descC))
					} else if descC != "" && !strings.Contains(descC, "_") {
						complementos = append(complementos, descC)
					}
				}
			}
		}

		compFinal := strings.Join(complementos, "; ")
		if compFinal == "" {
			compFinal = desc
		}

		orgaoMov := ""
		if oMap, ok := m["orgaoJulgador"].(map[string]any); ok {
			if nomeOrgao, _ := oMap["nome"].(string); nomeOrgao != "" {
				orgaoMov = CorrigirMojibake(strings.TrimSpace(nomeOrgao))
			} else if nomeOrgao, _ := oMap["nomeOrgao"].(string); nomeOrgao != "" {
				orgaoMov = CorrigirMojibake(strings.TrimSpace(nomeOrgao))
			}
		}

		movs = append(movs, domain.Movimentacao{
			Data:          t,
			DataFormatada: t.Format("02/01/2006"),
			Tipo:          string(ClassificarTipo(desc + " " + compFinal)),
			Descricao:     truncate(desc, 2000),
			Complemento:   compFinal,
			Orgao:         orgaoMov,
			Fonte:         string(domain.FonteDataJud),
			IDExterno:     idExterno,
		})
	}

	sort.Slice(movs, func(i, j int) bool {
		return movs[i].Data.After(movs[j].Data)
	})

	return movs
}

// ParsePartes extrai polos e advogados
func ParsePartes(brutas []map[string]any) []domain.Parte {
	var partes []domain.Parte

	for _, p := range brutas {
		nome, _ := p["nome"].(string)
		nome = strings.TrimSpace(nome)
		if nome == "" {
			continue
		}

		poloStr, _ := p["polo"].(string)
		var polo domain.Polo
		switch strings.ToUpper(poloStr) {
		case "ATIVO":
			polo = domain.PoloAtivo
		case "PASSIVO":
			polo = domain.PoloPassivo
		default:
			polo = domain.PoloTerceiro
		}

		advogados := []domain.Advogado{}
		if advs, ok := p["advogados"].([]any); ok {
			for _, a := range advs {
				if amap, ok := a.(map[string]any); ok {
					nomeAdv, _ := amap["nome"].(string)
					nomeAdv = strings.TrimSpace(nomeAdv)
					if nomeAdv == "" {
						continue
					}
					numOAB := fmt.Sprintf("%v", amap["numero"])
					if numOAB == "" || numOAB == "<nil>" {
						numOAB = fmt.Sprintf("%v", amap["oab"])
					}
					ufOAB, _ := amap["uf"].(string)

					adv := domain.Advogado{
						Nome:      nomeAdv,
						OABNumero: strings.TrimSpace(numOAB),
						OABUF:     strings.ToUpper(strings.TrimSpace(ufOAB)),
					}
					adv.OAB = adv.Formatada()
					advogados = append(advogados, adv)
				}
			}
		}

		partes = append(partes, domain.Parte{
			Nome:      nome,
			Polo:      polo,
			Advogados: advogados,
		})
	}

	return partes
}

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
