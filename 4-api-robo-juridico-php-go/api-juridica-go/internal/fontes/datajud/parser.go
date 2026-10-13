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

var regexComarca = regexp.MustCompile(`(?i)da Comarca de (.+?)$`)

func ClassificarTipo(descricao string) domain.TipoMovimentacao {
	d := strings.ToLower(descricao)
	if strings.Contains(d, "sentença") || strings.Contains(d, "sentenca") {
		return domain.TipoSentenca
	}
	if strings.Contains(d, "decisão") || strings.Contains(d, "decisao") {
		return domain.TipoDecisao
	}
	if strings.Contains(d, "despacho") {
		return domain.TipoDespacho
	}
	if strings.Contains(d, "audiência") || strings.Contains(d, "audiencia") {
		return domain.TipoAudiencia
	}
	if strings.Contains(d, "intimação") || strings.Contains(d, "intimacao") {
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
		if dataStr == "" || desc == "" {
			continue
		}

		// Parse de data ISO
		t, err := time.Parse(time.RFC3339, dataStr)
		if err != nil {
			t, err = time.Parse("2006-01-02T15:04:05", strings.Split(dataStr, ".")[0])
			if err != nil {
				continue
			}
		}
		t = t.UTC()

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
					if nome != "" || descC != "" {
						complementos = append(complementos, strings.Trim(fmt.Sprintf("%s: %s", nome, descC), ": "))
					}
				}
			}
		}

		movs = append(movs, domain.Movimentacao{
			Data:          t,
			DataFormatada: t.Format("02/01/2006"),
			Tipo:          string(ClassificarTipo(desc)),
			Descricao:     truncate(desc, 2000),
			Complemento:   strings.Join(complementos, "; "),
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

		var advogados []domain.Advogado
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
