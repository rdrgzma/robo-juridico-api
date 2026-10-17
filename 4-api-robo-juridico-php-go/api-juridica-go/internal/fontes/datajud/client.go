package datajud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cnj"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/network"
)

const URLBaseDataJud = "https://api-publica.datajud.cnj.jus.br"

type ClienteDataJud struct {
	httpCliente *network.ClienteHTTP
	apiKey      string
	timeout     time.Duration
	tentativas  int
}

func NovoClienteDataJud(httpCliente *network.ClienteHTTP, apiKey string, timeout time.Duration, tentativas int) *ClienteDataJud {
	if tentativas <= 0 {
		tentativas = 3
	}
	return &ClienteDataJud{
		httpCliente: httpCliente,
		apiKey:      apiKey,
		timeout:     timeout,
		tentativas:  tentativas,
	}
}

func (c *ClienteDataJud) ConsultarProcesso(ctx context.Context, numero string) (*domain.Processo, error) {
	procCNJ := cnj.Parse(numero)
	if !procCNJ.Valido {
		return nil, domain.ErrNumeroInvalido
	}

	payload := map[string]any{
		"size":             1,
		"track_total_hits": false,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []map[string]any{
					{
						"term": map[string]any{
							"numeroProcesso": procCNJ.Digitos,
						},
					},
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%s/_search", URLBaseDataJud, procCNJ.IndiceDatajud)

	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("ApiKey %s", c.apiKey))

	var ultimaResp *network.RespostaHTTP
	espera := 500 * time.Millisecond

	timeoutTentativa := c.timeout
	if timeoutTentativa > 15*time.Second {
		timeoutTentativa = 15 * time.Second
	}

	for tentativa := 1; tentativa <= c.tentativas; tentativa++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		ctxReq, cancel := context.WithTimeout(ctx, timeoutTentativa)
		resp, err := c.httpCliente.PostJSON(ctxReq, endpoint, payloadBytes, headers)
		cancel()

		if err == nil && resp != nil {
			if resp.StatusCode == http.StatusOK {
				return c.montarProcesso(resp.Body, procCNJ)
			}
			if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
				return nil, domain.NovoErroAPI(http.StatusServiceUnavailable, "fonte_indisponivel", "DataJud: chave de API recusada pelo CNJ")
			}
			ultimaResp = resp
		}

		if tentativa < c.tentativas {
			pausa := espera
			if ultimaResp != nil && ultimaResp.RetryAfter() > 0 {
				pausa = time.Duration(ultimaResp.RetryAfter()) * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(pausa):
			}
			espera *= 2
		}
	}

	return nil, domain.NovoErroAPI(http.StatusServiceUnavailable, "fonte_indisponivel", "DataJud não respondeu após tentativas")
}

func (c *ClienteDataJud) montarProcesso(body []byte, procCNJ *cnj.ProcessoCNJ) (*domain.Processo, error) {
	var respES struct {
		Hits struct {
			Hits []struct {
				Source map[string]any `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.Unmarshal(body, &respES); err != nil {
		return nil, err
	}

	if len(respES.Hits.Hits) == 0 {
		return nil, nil // Não encontrado no DataJud
	}

	src := respES.Hits.Hits[0].Source

	// Extração de partes e advogados
	var partesBrutas []map[string]any
	if pList, ok := src["partes"].([]any); ok {
		for _, item := range pList {
			if m, ok := item.(map[string]any); ok {
				partesBrutas = append(partesBrutas, m)
			}
		}
	}
	partes := ParsePartes(partesBrutas)

	var advs []domain.Advogado
	for _, p := range partes {
		advs = append(advs, p.Advogados...)
	}

	// Extração de movimentos
	var movsBrutas []map[string]any
	if mList, ok := src["movimentos"].([]any); ok {
		for _, item := range mList {
			if m, ok := item.(map[string]any); ok {
				movsBrutas = append(movsBrutas, m)
			}
		}
	}
	movimentacoes := ParseMovimentos(movsBrutas)

	// Extração de órgão e comarca
	orgao := ""
	if oMap, ok := src["orgaoJulgador"].(map[string]any); ok {
		orgao, _ = oMap["nome"].(string)
		orgao = CorrigirMojibake(strings.TrimSpace(orgao))
	}

	// Garante que cada movimentação do DataJud herde o órgão da capa quando não vier no movimento individual
	for i := range movimentacoes {
		if movimentacoes[i].Orgao == "" && orgao != "" {
			movimentacoes[i].Orgao = orgao
		}
	}

	comarca, _ := src["municipioInicioFormatado"].(string)
	comarca = CorrigirMojibake(strings.TrimSpace(comarca))
	if comarca == "" && orgao != "" {
		if m := regexComarca.FindStringSubmatch(orgao); len(m) > 1 {
			comarca = strings.TrimSpace(m[1])
		} else if m := regexJurisdicaoCidades.FindStringSubmatch(orgao); len(m) > 1 {
			comarca = strings.TrimSpace(m[1])
		}
	}

	// Classe e assunto
	classe := ""
	if cMap, ok := src["classe"].(map[string]any); ok {
		classe, _ = cMap["nome"].(string)
		classe = CorrigirMojibake(strings.TrimSpace(classe))
	}

	assunto := ""
	if aList, ok := src["assuntos"].([]any); ok && len(aList) > 0 {
		if aMap, ok := aList[0].(map[string]any); ok {
			assunto, _ = aMap["nome"].(string)
			assunto = CorrigirMojibake(strings.TrimSpace(assunto))
		}
	}

	// Valor da causa
	var valorCausa *float64
	if v, ok := src["valorCausa"].(float64); ok {
		valorCausa = &v
	}

	// Data de ajuizamento (suporta ISO e formato compacto YYYYMMDDHHmmss do DataJud, com fallback para 1ª movimentação)
	var dataAjuizamento *time.Time
	if dStr, ok := src["dataAjuizamento"].(string); ok && dStr != "" {
		if t, okParse := ParseDataDataJud(dStr); okParse {
			dataAjuizamento = &t
		}
	}
	if dataAjuizamento == nil && len(movimentacoes) > 0 {
		maisAntiga := movimentacoes[len(movimentacoes)-1].Data
		dataAjuizamento = &maisAntiga
	}

	var dataUltimaMov string
	var dataUltimaMovFormatada string
	if len(movimentacoes) > 0 {
		dataUltimaMov = movimentacoes[0].Data.Format("2006-01-02")
		dataUltimaMovFormatada = movimentacoes[0].Data.Format("02/01/2006")
	}

	tribunalFinal, _ := src["tribunal"].(string)
	if tribunalFinal == "" {
		tribunalFinal = procCNJ.Tribunal
	}

	grau, _ := src["grau"].(string)

	if advs == nil {
		advs = []domain.Advogado{}
	}

	linkTribunal := cnj.GerarLinkTribunal(procCNJ.Formatado, tribunalFinal, grau, procCNJ.OrigemCodigo)
	for i := range movimentacoes {
		if movimentacoes[i].Link == "" {
			movimentacoes[i].Link = linkTribunal
		}
	}

	proc := &domain.Processo{
		Numero:                         procCNJ.Formatado,
		NumeroDigitos:                  procCNJ.Digitos,
		Tribunal:                       tribunalFinal,
		Segmento:                       procCNJ.NomeSegmento,
		Grau:                           grau,
		Classe:                         classe,
		Assunto:                        assunto,
		OrgaoJulgador:                  orgao,
		Comarca:                        comarca,
		Link:                           linkTribunal,
		ValorCausa:                     valorCausa,
		ValorCausaFormatado:            FormatarDinheiroBR(valorCausa),
		DataAjuizamento:                dataAjuizamento,
		DataUltimaMovimentacao:         dataUltimaMov,
		DataUltimaMovimentacaoFormatada: dataUltimaMovFormatada,
		Partes:                         partes,
		Advogados:                      advs,
		Movimentacoes:                  movimentacoes,
		Publicacoes:                    []domain.Publicacao{},
		Fontes:                         []string{string(domain.FonteDataJud)},
		Polos: domain.Polos{
			Ativo:     filtrarPolo(partes, domain.PoloAtivo),
			Passivo:   filtrarPolo(partes, domain.PoloPassivo),
			Terceiros: filtrarPolo(partes, domain.PoloTerceiro),
		},
		Totais: domain.Totais{
			Partes:        len(partes),
			Advogados:     len(advs),
			Movimentacoes: len(movimentacoes),
			Publicacoes:   0,
		},
	}

	if dataAjuizamento != nil {
		proc.DataAjuizamentoFormatada = dataAjuizamento.Format("02/01/2006")
	}

	proc.Titulo = montarTitulo(proc.Polos)

	return proc, nil
}

func filtrarPolo(partes []domain.Parte, polo domain.Polo) []domain.Parte {
	filtradas := []domain.Parte{}
	for _, p := range partes {
		if p.Polo == polo {
			filtradas = append(filtradas, p)
		}
	}
	return filtradas
}

func montarTitulo(polos domain.Polos) string {
	var ativos []string
	for _, p := range polos.Ativo {
		ativos = append(ativos, p.Nome)
	}
	var passivos []string
	for _, p := range polos.Passivo {
		passivos = append(passivos, p.Nome)
	}

	ativoStr := "—"
	if len(ativos) > 0 {
		ativoStr = strings.Join(ativos, ", ")
	}
	passivoStr := "—"
	if len(passivos) > 0 {
		passivoStr = strings.Join(passivos, ", ")
	}

	return fmt.Sprintf("%s × %s", ativoStr, passivoStr)
}
