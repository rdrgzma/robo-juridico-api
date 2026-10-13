package http

import (
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cnj"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/service"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
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
	versao       string
}

func NovosHandlers(orq *service.Orquestrador, db storage.BancoDados, versao string) *Handlers {
	if versao == "" {
		versao = "1.0.0"
	}
	return &Handlers{
		orquestrador: orq,
		database:     db,
		versao:       versao,
	}
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
	w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data: https://cdn.jsdelivr.net; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;")
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
