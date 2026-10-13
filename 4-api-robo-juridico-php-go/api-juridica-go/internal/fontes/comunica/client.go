package comunica

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/network"
)

const (
	URLAPIComunica = "https://comunicaapi.pje.jus.br/api/v1/comunicacao"
	OrigemComunica = "https://comunica.pje.jus.br"
	RefererComunica = "https://comunica.pje.jus.br/"
)

var regexSomenteDigitos = regexp.MustCompile(`\D`)

type ClienteComunica struct {
	httpCliente   *network.ClienteHTTP
	proxyPool     *network.ProxyPool
	timeout       time.Duration
	tamanhoPagina int
	tentativas    int
}

func NovoClienteComunica(httpCliente *network.ClienteHTTP, proxyPool *network.ProxyPool, timeout time.Duration, tamanhoPagina, tentativas int) *ClienteComunica {
	if tamanhoPagina <= 0 {
		tamanhoPagina = 100
	}
	if tentativas <= 0 {
		tentativas = 3
	}
	return &ClienteComunica{
		httpCliente:   httpCliente,
		proxyPool:     proxyPool,
		timeout:       timeout,
		tamanhoPagina: tamanhoPagina,
		tentativas:    tentativas,
	}
}

type respostaDJEN struct {
	Count int              `json:"count"`
	Items []map[string]any `json:"items"`
}

func (c *ClienteComunica) buscar(ctx context.Context, params url.Values) ([]domain.Publicacao, int, error) {
	cabecalhos := c.httpCliente.CabecalhosNavegador(OrigemComunica, RefererComunica)

	var resp *network.RespostaHTTP
	var err error

	for tentativa := 1; tentativa <= c.tentativas; tentativa++ {
		ctxReq, cancel := context.WithTimeout(ctx, c.timeout)
		if c.proxyPool != nil && c.proxyPool.TemProxies() {
			resp, err = c.proxyPool.ExecutarComCorrida(ctxReq, c.httpCliente, URLAPIComunica, params, cabecalhos)
		} else {
			resp, err = c.httpCliente.Get(ctxReq, URLAPIComunica, params, cabecalhos, "")
		}
		cancel()

		if err == nil && resp != nil {
			if resp.StatusCode == 200 {
				var dados respostaDJEN
				if err := json.Unmarshal(resp.Body, &dados); err == nil {
					var pubs []domain.Publicacao
					for _, item := range dados.Items {
						pubs = append(pubs, c.montarPublicacao(item))
					}
					return pubs, dados.Count, nil
				}
			}
			if resp.StatusCode == 400 || resp.StatusCode == 404 {
				return []domain.Publicacao{}, 0, nil
			}
		}

		if tentativa < c.tentativas {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}
	}

	return nil, 0, domain.NovoErroAPI(503, "fonte_indisponivel", "Comunica/DJEN sem resposta definitiva")
}

func (c *ClienteComunica) PublicacoesDoProcesso(ctx context.Context, numero string, inicio, fim time.Time) ([]domain.Publicacao, error) {
	digitos := regexSomenteDigitos.ReplaceAllString(numero, "")
	if inicio.IsZero() {
		inicio = time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if fim.IsZero() {
		fim = time.Now().UTC()
	}

	params := url.Values{}
	params.Set("numeroProcesso", digitos)
	params.Set("dataDisponibilizacaoInicio", inicio.Format("2006-01-02"))
	params.Set("dataDisponibilizacaoFim", fim.Format("2006-01-02"))
	params.Set("tamanhoPagina", fmt.Sprintf("%d", c.tamanhoPagina))

	pubs, _, err := c.buscar(ctx, params)
	if err == nil && len(pubs) > 0 {
		return pubs, nil
	}

	// Segunda tentativa com número formatado caso digitos puros volte vazio
	if len(pubs) == 0 && numero != digitos {
		params.Set("numeroProcesso", numero)
		pubs, _, err = c.buscar(ctx, params)
	}

	return pubs, err
}

func (c *ClienteComunica) PublicacoesPorOAB(ctx context.Context, oabNumero, oabUF string, inicio, fim time.Time, paginas int) ([]domain.Publicacao, error) {
	digitosOAB := regexSomenteDigitos.ReplaceAllString(oabNumero, "")
	uf := strings.ToUpper(strings.TrimSpace(oabUF))

	params := url.Values{}
	params.Set("numeroOab", digitosOAB)
	params.Set("ufOab", uf)
	params.Set("tamanhoPagina", fmt.Sprintf("%d", c.tamanhoPagina))

	if !inicio.IsZero() {
		params.Set("dataDisponibilizacaoInicio", inicio.Format("2006-01-02"))
	}
	if !fim.IsZero() {
		params.Set("dataDisponibilizacaoFim", fim.Format("2006-01-02"))
	}

	var todas []domain.Publicacao
	pagina := 1

	for {
		params.Set("pagina", fmt.Sprintf("%d", pagina))
		pubs, total, err := c.buscar(ctx, params)
		if err != nil && pagina == 1 {
			return nil, err
		}

		todas = append(todas, pubs...)
		vistos := pagina * c.tamanhoPagina
		acabou := len(pubs) == 0 || vistos >= total
		if acabou || (paginas > 0 && pagina >= paginas) {
			break
		}
		pagina++
	}

	return todas, nil
}

func (c *ClienteComunica) montarPublicacao(item map[string]any) domain.Publicacao {
	id := fmt.Sprintf("%v", item["id"])
	dataDisp, _ := item["data_disponibilizacao"].(string)
	if dataDisp == "" {
		dataDisp, _ = item["datadisponibilizacao"].(string)
	}
	_, dataFormatada := ParseDataDJEN(dataDisp)

	tipoCom, _ := item["tipoComunicacao"].(string)
	tipoDoc, _ := item["tipoDocumento"].(string)
	orgao, _ := item["nomeOrgao"].(string)
	numProc, _ := item["numero_processo"].(string)
	if numProc == "" {
		numProc, _ = item["numeroprocessocommascara"].(string)
	}
	classe, _ := item["nomeClasse"].(string)
	textoBruto, _ := item["texto"].(string)
	link, _ := item["link"].(string)

	var partes []domain.Parte
	if dests, ok := item["destinatarios"].([]any); ok {
		for _, d := range dests {
			if dmap, ok := d.(map[string]any); ok {
				nome, _ := dmap["nome"].(string)
				poloStr, _ := dmap["polo"].(string)
				var polo domain.Polo
				switch strings.ToUpper(poloStr) {
				case "A":
					polo = domain.PoloAtivo
				case "P":
					polo = domain.PoloPassivo
				default:
					polo = domain.PoloTerceiro
				}
				if strings.TrimSpace(nome) != "" {
					partes = append(partes, domain.Parte{
						Nome: strings.TrimSpace(nome),
						Polo: polo,
					})
				}
			}
		}
	}

	var advs []domain.Advogado
	if destAdvs, ok := item["destinatarioadvogados"].([]any); ok {
		for _, entry := range destAdvs {
			if emap, ok := entry.(map[string]any); ok {
				if advMap, ok := emap["advogado"].(map[string]any); ok {
					nome, _ := advMap["nome"].(string)
					num, _ := advMap["numero_oab"].(string)
					uf, _ := advMap["uf_oab"].(string)
					if strings.TrimSpace(nome) != "" {
						adv := domain.Advogado{
							Nome:      strings.TrimSpace(nome),
							OABNumero: strings.TrimSpace(num),
							OABUF:     strings.ToUpper(strings.TrimSpace(uf)),
						}
						adv.OAB = adv.Formatada()
						advs = append(advs, adv)
					}
				}
			}
		}
	}

	return domain.Publicacao{
		ID:                   id,
		DataDisponibilizacao: dataDisp,
		DataFormatada:        dataFormatada,
		TipoComunicacao:      tipoCom,
		TipoDocumento:        tipoDoc,
		Orgao:                orgao,
		NumeroProcesso:       numProc,
		Classe:               classe,
		Texto:                ExtrairTeor(textoBruto),
		Link:                 link,
		Partes:               partes,
		Advogados:            advs,
	}
}
