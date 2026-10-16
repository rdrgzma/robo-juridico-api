package http

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator/collectors"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cnj"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/prazos"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/service"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/webhooks"
)

//go:embed openapi.json
var openAPIJSON []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <title>Robô Jurídico — Documentação Swagger</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

var ufsValidas = map[string]bool{
	"AC": true, "AL": true, "AM": true, "AP": true, "BA": true, "CE": true,
	"DF": true, "ES": true, "GO": true, "MA": true, "MG": true, "MS": true,
	"MT": true, "PA": true, "PB": true, "PE": true, "PI": true, "PR": true,
	"RJ": true, "RN": true, "RO": true, "RR": true, "RS": true, "SC": true,
	"SE": true, "SP": true, "TO": true,
}

type Handlers struct {
	orquestrador *service.Orquestrador
	database     storage.BancoDados
	calcEngine   *calculator.EngineCalculo
	worker       *collectors.WorkerColetor
	webhooks     *webhooks.GerenciadorWebhooks
	versao       string
}

func NovosHandlers(orq *service.Orquestrador, db storage.BancoDados, versao string, calcEngine ...*calculator.EngineCalculo) *Handlers {
	if versao == "" {
		versao = "1.0.0"
	}
	var engine *calculator.EngineCalculo
	if len(calcEngine) > 0 && calcEngine[0] != nil {
		engine = calcEngine[0]
	} else {
		engine = calculator.NovoEngineCalculo(nil)
	}
	return &Handlers{
		orquestrador: orq,
		database:     db,
		calcEngine:   engine,
		webhooks:     webhooks.NovoGerenciadorWebhooks(),
		versao:       versao,
	}
}

func (h *Handlers) SetWorkerColetor(w *collectors.WorkerColetor) {
	h.worker = w
}

// Saude verifica se a API está online (/saude)
func (h *Handlers) Saude(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"versao": h.versao,
	})
}

// OpenAPIJSON serve a especificação OpenAPI 3.1.0 em JSON (/openapi.json)
func (h *Handlers) OpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPIJSON)
}

// SwaggerUI serve a interface interativa do Swagger UI (/docs)
func (h *Handlers) SwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; connect-src 'self' http://localhost:* http://127.0.0.1:* https:; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data: https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

// DecodificarCNJ decodifica o número CNJ sem consultar fontes externas (/v1/cnj/{numero})
func (h *Handlers) DecodificarCNJ(w http.ResponseWriter, r *http.Request) {
	numero := chi.URLParam(r, "numero")
	parsed := cnj.Parse(numero)

	if !parsed.Valido {
		JSON(w, http.StatusOK, domain.RespostaCNJ{
			OK:            true,
			Numero:        numero,
			NumeroDigitos: cnj.ApenasDigitos(numero),
			Valido:        false,
		})
		return
	}

	JSON(w, http.StatusOK, domain.RespostaCNJ{
		OK:            true,
		Numero:        parsed.Formatado,
		NumeroDigitos: parsed.Digitos,
		Valido:        true,
		Tribunal:      parsed.Tribunal,
		Segmento:      parsed.NomeSegmento,
		Ano:           parsed.Ano,
		OrigemCodigo:  parsed.OrigemCodigo,
	})
}

// ProcessoExiste verifica a existência do processo nas fontes (/v1/processos/{numero}/existe)
func (h *Handlers) ProcessoExiste(w http.ResponseWriter, r *http.Request) {
	numero := chi.URLParam(r, "numero")
	parsed := cnj.Parse(numero)

	if !parsed.Valido {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "numero_invalido",
			"Número fora do padrão CNJ (NNNNNNN-DD.AAAA.J.TR.OOOO) ou com dígito verificador errado.")
		return
	}

	achou, fonte, err := h.orquestrador.Existe(r.Context(), parsed.Formatado)
	if err != nil {
		if err == domain.ErrFonteIndisponivel {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel",
				"A fonte do CNJ não respondeu. Tente de novo em instantes.")
			return
		}
		EnviarErro(w, r, http.StatusBadGateway, "falha_na_fonte", "A fonte do CNJ devolveu uma resposta inesperada.")
		return
	}

	JSON(w, http.StatusOK, domain.RespostaExiste{
		OK:     true,
		Numero: parsed.Formatado,
		Existe: achou,
		Fonte:  fonte,
	})
}

// Processo consulta dados completos do processo (/v1/processos/{numero})
func (h *Handlers) Processo(w http.ResponseWriter, r *http.Request) {
	numero := chi.URLParam(r, "numero")
	parsed := cnj.Parse(numero)

	if !parsed.Valido {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "numero_invalido",
			"Número fora do padrão CNJ (NNNNNNN-DD.AAAA.J.TR.OOOO) ou com dígito verificador errado.")
		return
	}

	// Parâmetros de query
	modo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("modo")))
	if modo == "" {
		modo = "completo"
	}
	if modo != "completo" && modo != "rapido" {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida",
			"Modo inválido. Valores aceitos: 'completo', 'rapido'.")
		return
	}

	comPublicacoes := true
	pubParam := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("publicacoes")))
	if pubParam == "false" || pubParam == "0" || pubParam == "nao" || pubParam == "não" {
		comPublicacoes = false
	}

	inicio := time.Now()
	proc, avisos, err := h.orquestrador.ConsultarProcessoDetalhado(r.Context(), parsed.Formatado, comPublicacoes, modo)
	if err != nil {
		if err == domain.ErrFonteIndisponivel {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel",
				"A fonte do CNJ não respondeu. Tente de novo em instantes.")
			return
		}
		if err == domain.ErrNumeroInvalido {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "numero_invalido", "Número fora do padrão CNJ.")
			return
		}
		EnviarErro(w, r, http.StatusBadGateway, "falha_na_fonte", "A fonte do CNJ devolveu uma resposta inesperada.")
		return
	}

	if proc == nil {
		if len(avisos) > 0 {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel", strings.Join(avisos, " "))
			return
		}
		EnviarErro(w, r, http.StatusNotFound, "processo_nao_encontrado",
			"O processo não aparece no DataJud nem no DJEN. Se foi distribuído agora, tente de novo mais tarde; processo em segredo de justiça não é publicado.")
		return
	}

	ms := time.Since(inicio).Milliseconds()
	JSON(w, http.StatusOK, domain.RespostaProcesso{
		OK:           true,
		ConsultadoEm: time.Now().UTC().Format(time.RFC3339),
		TempoMs:      ms,
		Avisos:       avisos,
		Processo:     *proc,
	})
}

// RequisicaoConsultarProcesso representa o payload de consulta via POST JSON
type RequisicaoConsultarProcesso struct {
	NumeroProcesso string `json:"numero_processo"`
	Numero         string `json:"numero"`
	Modo           string `json:"modo"`
	Publicacoes    *bool  `json:"publicacoes"`
}

// ConsultarProcessoJSON consulta dados do processo via corpo JSON (POST /v1/processos/consultar)
func (h *Handlers) ConsultarProcessoJSON(w http.ResponseWriter, r *http.Request) {
	var reqBody RequisicaoConsultarProcesso
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida",
			"Corpo da requisição deve ser um JSON válido contendo 'numero_processo'.")
		return
	}

	numero := reqBody.NumeroProcesso
	if numero == "" {
		numero = reqBody.Numero
	}

	parsed := cnj.Parse(numero)
	if !parsed.Valido {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "numero_invalido",
			"Número fora do padrão CNJ (NNNNNNN-DD.AAAA.J.TR.OOOO) ou com dígito verificador errado.")
		return
	}

	modo := strings.ToLower(strings.TrimSpace(reqBody.Modo))
	if modo == "" {
		modo = "completo"
	}
	if modo != "completo" && modo != "rapido" {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida",
			"Modo inválido. Valores aceitos: 'completo', 'rapido'.")
		return
	}

	comPublicacoes := true
	if reqBody.Publicacoes != nil {
		comPublicacoes = *reqBody.Publicacoes
	}

	inicio := time.Now()
	proc, avisos, err := h.orquestrador.ConsultarProcessoDetalhado(r.Context(), parsed.Formatado, comPublicacoes, modo)
	if err != nil {
		if err == domain.ErrFonteIndisponivel {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel",
				"A fonte do CNJ não respondeu. Tente de novo em instantes.")
			return
		}
		if err == domain.ErrNumeroInvalido {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "numero_invalido", "Número fora do padrão CNJ.")
			return
		}
		EnviarErro(w, r, http.StatusBadGateway, "falha_na_fonte", "A fonte do CNJ devolveu uma resposta inesperada.")
		return
	}

	if proc == nil {
		if len(avisos) > 0 {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel", strings.Join(avisos, " "))
			return
		}
		EnviarErro(w, r, http.StatusNotFound, "processo_nao_encontrado",
			"O processo não aparece no DataJud nem no DJEN. Se foi distribuído agora, tente de novo mais tarde; processo em segredo de justiça não é publicado.")
		return
	}

	ms := time.Since(inicio).Milliseconds()
	JSON(w, http.StatusOK, domain.RespostaProcesso{
		OK:           true,
		ConsultadoEm: time.Now().UTC().Format(time.RFC3339),
		TempoMs:      ms,
		Avisos:       avisos,
		Processo:     *proc,
	})
}

// OABProcessos lista processos agrupados com publicações da OAB (/v1/oab/{uf}/{numero}/processos)
func (h *Handlers) OABProcessos(w http.ResponseWriter, r *http.Request) {
	uf := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "uf")))
	numero := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "numero")))

	if !ufsValidas[uf] {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "uf_invalida", fmt.Sprintf("UF desconhecida: %s", uf))
		return
	}

	dias := 30
	if diasParam := r.URL.Query().Get("dias"); diasParam != "" {
		if v, err := strconv.Atoi(diasParam); err == nil && v >= 1 && v <= 365 {
			dias = v
		} else {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida", "Parâmetro 'dias' deve ser um inteiro entre 1 e 365.")
			return
		}
	}

	tribunal := strings.TrimSpace(r.URL.Query().Get("tribunal"))

	resumos, err := h.orquestrador.BuscarPorOAB(r.Context(), numero, uf, tribunal, dias)
	if err != nil {
		if err == domain.ErrFonteIndisponivel {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel",
				"A fonte do DJEN não respondeu. Tente de novo em instantes.")
			return
		}
		EnviarErro(w, r, http.StatusBadGateway, "falha_na_fonte", "A fonte do CNJ devolveu uma resposta inesperada.")
		return
	}

	JSON(w, http.StatusOK, domain.RespostaOABProcessos{
		OK:          true,
		OAB:         fmt.Sprintf("OAB/%s %s", uf, numero),
		PeriodoDias: dias,
		Total:       len(resumos),
		Processos:   resumos,
	})
}

// OABPublicacoes retorna a lista detalhada de publicações da OAB (/v1/oab/{uf}/{numero}/publicacoes)
func (h *Handlers) OABPublicacoes(w http.ResponseWriter, r *http.Request) {
	uf := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "uf")))
	numero := strings.ToUpper(strings.TrimSpace(chi.URLParam(r, "numero")))

	if !ufsValidas[uf] {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "uf_invalida", fmt.Sprintf("UF desconhecida: %s", uf))
		return
	}

	dias := 7
	if diasParam := r.URL.Query().Get("dias"); diasParam != "" {
		if v, err := strconv.Atoi(diasParam); err == nil && v >= 1 && v <= 90 {
			dias = v
		} else {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida", "Parâmetro 'dias' deve ser um inteiro entre 1 e 90.")
			return
		}
	}

	paginas := 1
	if pagParam := r.URL.Query().Get("paginas"); pagParam != "" {
		if v, err := strconv.Atoi(pagParam); err == nil && v >= 1 && v <= 20 {
			paginas = v
		} else {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "entrada_invalida", "Parâmetro 'paginas' deve ser um inteiro entre 1 e 20.")
			return
		}
	}

	fim := time.Now().UTC()
	inicio := fim.AddDate(0, 0, -dias)

	pubs, err := h.orquestrador.PublicacoesPorOAB(r.Context(), numero, uf, inicio, fim, paginas)
	if err != nil {
		if err == domain.ErrFonteIndisponivel {
			w.Header().Set("Retry-After", "30")
			EnviarErro(w, r, http.StatusServiceUnavailable, "fonte_indisponivel",
				"A fonte do DJEN não respondeu. Tente de novo em instantes.")
			return
		}
		EnviarErro(w, r, http.StatusBadGateway, "falha_na_fonte", "A fonte do CNJ devolveu uma resposta inesperada.")
		return
	}

	JSON(w, http.StatusOK, domain.RespostaOABPublicacoes{
		OK:          true,
		OAB:         fmt.Sprintf("OAB/%s %s", uf, numero),
		Inicio:      inicio.Format("2006-01-02"),
		Fim:         fim.Format("2006-01-02"),
		Total:       len(pubs),
		Publicacoes: pubs,
	})
}

// Conta retorna dados da conta do cliente e cota mensal (/v1/conta)
func (h *Handlers) Conta(w http.ResponseWriter, r *http.Request) {
	identidade := ObterIdentidade(r)
	if identidade == nil {
		EnviarErro(w, r, http.StatusUnauthorized, "nao_autorizado", "Envie uma chave válida no cabeçalho X-API-Key.")
		return
	}

	reiniciaEm := storage.InicioDoProximoMesBrasil().Format(time.RFC3339)

	if identidade.EhInterna || h.database == nil || identidade.LimiteMensal == nil {
		JSON(w, http.StatusOK, domain.RespostaConta{
			OK:              true,
			Cliente:         identidade.NomeCliente,
			Plano:           "interno",
			UsadoNoMes:      0,
			ReiniciaEm:      reiniciaEm,
			LimitePorMinuto: identidade.LimitePorMinuto,
		})
		return
	}

	usado, err := h.database.ObterConsumoMes(r.Context(), identidade.ClienteID)
	if err != nil {
		EnviarErro(w, r, http.StatusInternalServerError, "erro_interno", "Erro ao obter consumo da conta.")
		return
	}

	restante := *identidade.LimiteMensal - usado
	if restante < 0 {
		restante = 0
	}

	JSON(w, http.StatusOK, domain.RespostaConta{
		OK:              true,
		Cliente:         identidade.NomeCliente,
		Plano:           identidade.Plano,
		LimiteMensal:    identidade.LimiteMensal,
		UsadoNoMes:      usado,
		Restante:        &restante,
		ReiniciaEm:      reiniciaEm,
		LimitePorMinuto: identidade.LimitePorMinuto,
	})
}

// ListarIndices retorna a relação de indexadores cadastrados (/v1/calculos/indices)
func (h *Handlers) ListarIndices(w http.ResponseWriter, r *http.Request) {
	indices := h.calcEngine.Repositorio().ListarIndices()
	JSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"indices": indices,
	})
}

// ObterSerieIndice retorna os pontos históricos de um índice específico (/v1/calculos/indices/{codigo}/serie)
func (h *Handlers) ObterSerieIndice(w http.ResponseWriter, r *http.Request) {
	codigo := chi.URLParam(r, "codigo")
	serie, err := h.calcEngine.Repositorio().ObterSerieCompleta(strings.ToUpper(codigo))
	if err != nil {
		EnviarErro(w, r, http.StatusNotFound, "indice_nao_encontrado", fmt.Sprintf("Índice '%s' não encontrado.", codigo))
		return
	}
	JSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"codigo": strings.ToUpper(codigo),
		"serie":  serie,
	})
}

// CalcularJudicial executa a liquidação judicial completa de débitos/créditos (/v1/calculos/judicial)
func (h *Handlers) CalcularJudicial(w http.ResponseWriter, r *http.Request) {
	var req calculator.CalculoJudicialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "Corpo da requisição JSON inválido ou malformado.")
		return
	}

	resp, err := h.calcEngine.Executar(req)
	if err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "erro_calculo", err.Error())
		return
	}

	JSON(w, http.StatusOK, resp)
}

// SincronizarIndices aciona manualmente a coleta de índices do Bacen, IBGE e Tribunais (/v1/calculos/coletar)
func (h *Handlers) SincronizarIndices(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		EnviarErro(w, r, http.StatusServiceUnavailable, "worker_nao_disponivel", "Worker de coleta não configurado.")
		return
	}

	rel := h.worker.Executar(r.Context())
	JSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"relatorio": rel,
	})
}

// StatusColeta retorna o status da última rotina de coleta (/v1/calculos/coletar/status)
func (h *Handlers) StatusColeta(w http.ResponseWriter, r *http.Request) {
	if h.worker == nil {
		EnviarErro(w, r, http.StatusServiceUnavailable, "worker_nao_disponivel", "Worker de coleta não configurado.")
		return
	}

	rel, ultima, emExec := h.worker.UltimoRelatorio()
	var ultimaStr string
	if !ultima.IsZero() {
		ultimaStr = ultima.Format(time.RFC3339)
	}

	JSON(w, http.StatusOK, map[string]any{
		"ok":               true,
		"em_execucao":      emExec,
		"ultima_execucao":  ultimaStr,
		"ultimo_relatorio": rel,
	})
}

// ImportarPlanilha extrai parcelas de arquivos Excel (.xlsx), CSV (.csv) ou texto colado (/v1/calculos/importar-planilha)
func (h *Handlers) ImportarPlanilha(w http.ResponseWriter, r *http.Request) {
	nomeArquivo := "planilha.csv"
	var conteudo []byte

	// 1. Tenta multipart form
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err == nil { // Max 10MB
			file, header, err := r.FormFile("planilha")
			if err != nil {
				file, header, err = r.FormFile("arquivo")
			}
			if err == nil {
				defer file.Close()
				nomeArquivo = header.Filename
				conteudo, _ = io.ReadAll(file)
			}
		}
	}

	// 2. Fallback para corpo raw ou texto colado
	if len(conteudo) == 0 {
		if fn := r.Header.Get("X-Filename"); fn != "" {
			nomeArquivo = fn
		} else if fn := r.URL.Query().Get("arquivo"); fn != "" {
			nomeArquivo = fn
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil || len(bodyBytes) == 0 {
			EnviarErro(w, r, http.StatusBadRequest, "arquivo_vazio", "Nenhum arquivo ou texto enviado para importação.")
			return
		}
		conteudo = bodyBytes
	}

	parcelas, err := calculator.ImportarPlanilha(nomeArquivo, conteudo)
	if err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "erro_importacao", err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"ok":             true,
		"total_parcelas": len(parcelas),
		"arquivo":        nomeArquivo,
		"parcelas":       parcelas,
	})
}

// CalcularFinanciamento executa a revisão contratual bancária nos sistemas Price, SAC ou Gauss (/v1/calculos/financiamento)
func (h *Handlers) CalcularFinanciamento(w http.ResponseWriter, r *http.Request) {
	var req calculator.CalculoFinanciamentoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "Corpo da requisição JSON inválido ou malformado.")
		return
	}

	resp, err := calculator.CalcularRevisaoFinanciamento(req)
	if err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "erro_calculo_financiamento", err.Error())
		return
	}

	JSON(w, http.StatusOK, resp)
}

// ExportarPlanilha gera e retorna arquivo Excel (.xlsx) ou CSV (.csv) da memória de cálculo (/v1/calculos/exportar-planilha)
func (h *Handlers) ExportarPlanilha(w http.ResponseWriter, r *http.Request) {
	formato := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("formato")))
	if formato == "" {
		formato = "xlsx"
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil || len(bodyBytes) == 0 {
		EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "Envie o JSON do cálculo judicial ou os parâmetros das parcelas.")
		return
	}

	// Aceita tanto um CalculoJudicialResponse pronto quanto um CalculoJudicialRequest para calcular na hora
	var respCalculo *calculator.CalculoJudicialResponse
	var jaCalculado calculator.CalculoJudicialResponse
	if err := json.Unmarshal(bodyBytes, &jaCalculado); err == nil && (len(jaCalculado.ItensLiquidados) > 0 || len(jaCalculado.Itens) > 0) {
		respCalculo = &jaCalculado
	} else {
		var reqCalculo calculator.CalculoJudicialRequest
		if err := json.Unmarshal(bodyBytes, &reqCalculo); err != nil {
			EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "JSON inválido para exportação de planilha.")
			return
		}
		calculado, err := h.calcEngine.Executar(reqCalculo)
		if err != nil {
			EnviarErro(w, r, http.StatusUnprocessableEntity, "erro_calculo", err.Error())
			return
		}
		respCalculo = calculado
	}

	if formato == "csv" {
		csvBytes, err := calculator.ExportarMemoriaCSV(respCalculo)
		if err != nil {
			EnviarErro(w, r, http.StatusInternalServerError, "erro_exportacao", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="memoria-calculo-judicial.csv"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(csvBytes)
		return
	}

	xlsxBytes, err := calculator.ExportarMemoriaXLSX(respCalculo)
	if err != nil {
		EnviarErro(w, r, http.StatusInternalServerError, "erro_exportacao", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="memoria-calculo-judicial.xlsx"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(xlsxBytes)
}

// CalcularPrazo executa a contagem de prazos processuais fatais conforme CPC/2015 e CLT (/v1/prazos/calcular)
func (h *Handlers) CalcularPrazo(w http.ResponseWriter, r *http.Request) {
	var req prazos.CalculoPrazoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "Corpo da requisição JSON inválido.")
		return
	}

	resp, err := prazos.CalcularPrazoProcessual(req)
	if err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "erro_calculo_prazo", err.Error())
		return
	}

	JSON(w, http.StatusOK, resp)
}

// ListarFeriados retorna os feriados nacionais e forenses de um determinado ano (/v1/prazos/feriados)
func (h *Handlers) ListarFeriados(w http.ResponseWriter, r *http.Request) {
	ano := time.Now().Year()
	if anoStr := r.URL.Query().Get("ano"); anoStr != "" {
		if v, err := strconv.Atoi(anoStr); err == nil && v >= 1990 && v <= 2100 {
			ano = v
		}
	}

	feriados := prazos.ListarFeriadosAno(ano)
	JSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"ano":      ano,
		"feriados": feriados,
	})
}

// ListarWebhooks lista as inscrições de monitoramento Push ativas do cliente (/v1/webhooks)
func (h *Handlers) ListarWebhooks(w http.ResponseWriter, r *http.Request) {
	clienteID := 0
	if id := ObterIdentidade(r); id != nil {
		clienteID = id.ClienteID
	}

	lista := h.webhooks.Listar(clienteID)
	JSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"total":    len(lista),
		"webhooks": lista,
	})
}

// RegistrarWebhook cadastra um novo webhook para monitoramento de OAB ou Processo CNJ (/v1/webhooks)
func (h *Handlers) RegistrarWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TipoAlvo   string   `json:"tipo_alvo"`
		Alvo       string   `json:"alvo"`
		URLDestino string   `json:"url_destino"`
		Segredo    string   `json:"segredo"`
		Eventos    []string `json:"eventos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		EnviarErro(w, r, http.StatusBadRequest, "requisicao_invalida", "Corpo JSON inválido.")
		return
	}

	clienteID := 0
	if id := ObterIdentidade(r); id != nil {
		clienteID = id.ClienteID
	}

	sub, err := h.webhooks.Registrar(clienteID, body.TipoAlvo, body.Alvo, body.URLDestino, body.Segredo, body.Eventos)
	if err != nil {
		EnviarErro(w, r, http.StatusUnprocessableEntity, "webhook_invalido", err.Error())
		return
	}

	JSON(w, http.StatusCreated, map[string]any{
		"ok":      true,
		"webhook": sub,
	})
}

// RemoverWebhook exclui uma inscrição de webhook (/v1/webhooks/{id})
func (h *Handlers) RemoverWebhook(w http.ResponseWriter, r *http.Request) {
	idWebhook := chi.URLParam(r, "id")
	clienteID := 0
	if id := ObterIdentidade(r); id != nil {
		clienteID = id.ClienteID
	}

	if !h.webhooks.Remover(clienteID, idWebhook) {
		EnviarErro(w, r, http.StatusNotFound, "webhook_nao_encontrado", "Inscrição de webhook não encontrada.")
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"removido": idWebhook,
	})
}

// TestarWebhook gera e assina (HMAC-SHA256) um payload de evento de teste (/v1/webhooks/testar)
func (h *Handlers) TestarWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URLDestino string `json:"url_destino"`
		Segredo    string `json:"segredo"`
		Evento     string `json:"evento"`
		Alvo       string `json:"alvo"`
		EnviarHTTP bool   `json:"enviar_http"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	evt, statusHTTP, err := h.webhooks.DispararTeste(r.Context(), body.URLDestino, body.Segredo, body.Evento, body.Alvo, body.EnviarHTTP)
	if err != nil {
		EnviarErro(w, r, http.StatusBadGateway, "falha_envio_webhook", err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"status_http": statusHTTP,
		"payload":     evt,
	})
}
