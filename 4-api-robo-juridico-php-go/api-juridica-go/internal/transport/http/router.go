package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/auth"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
)

// NovoRoteador monta a árvore de rotas da API com middlewares e proteção de cotas
func NovoRoteador(handlers *Handlers, servicoAuth *auth.ServicoAuth, db storage.BancoDados) http.Handler {
	r := chi.NewRouter()

	// Middlewares globais
	r.Use(MiddlewareCORS)
	r.Use(chimiddleware.Recoverer)
	r.Use(MiddlewareEnvelope(db))

	// Tratamento uniforme para 404 e 405
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		EnviarErro(w, req, http.StatusNotFound, "rota_inexistente", "Rota não encontrada.")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		EnviarErro(w, req, http.StatusMethodNotAllowed, "metodo_nao_permitido", "Método HTTP não permitido.")
	})

	// Rotas públicas (saúde e documentação OpenAPI / Swagger UI)
	r.Get("/saude", handlers.Saude)
	r.Get("/openapi.json", handlers.OpenAPIJSON)
	r.Get("/docs", handlers.SwaggerUI)

	// Registra rotas para ambos os prefixos: /v1 e /api/v1
	registrarRotas := func(prefixo string) {
		// Rotas autenticadas SEM custo de cota (sem_custo)
		r.Group(func(sub chi.Router) {
			sub.Use(MiddlewareAuth(servicoAuth, db, false))

			sub.Get(prefixo+"/cnj/{numero}", handlers.DecodificarCNJ)
			sub.Get(prefixo+"/processos/{numero}/existe", handlers.ProcessoExiste)
			sub.Get(prefixo+"/conta", handlers.Conta)
			sub.Get(prefixo+"/calculos/indices", handlers.ListarIndices)
			sub.Get(prefixo+"/calculos/indices/{codigo}/serie", handlers.ObterSerieIndice)
			sub.Get(prefixo+"/calculos/coletar/status", handlers.StatusColeta)
			sub.Post(prefixo+"/calculos/coletar", handlers.SincronizarIndices)
			sub.Post(prefixo+"/calculos/importar-planilha", handlers.ImportarPlanilha)
			sub.Post(prefixo+"/calculos/exportar-planilha", handlers.ExportarPlanilha)
			sub.Post(prefixo+"/calculos/judicial", handlers.CalcularJudicial)
			sub.Post(prefixo+"/calculos/financiamento", handlers.CalcularFinanciamento)

			// Prazos Processuais Fatais (CPC/2015 e CLT)
			sub.Post(prefixo+"/prazos/calcular", handlers.CalcularPrazo)
			sub.Get(prefixo+"/prazos/feriados", handlers.ListarFeriados)

			// Webhooks de Monitoramento Push (OAB / Processos)
			sub.Get(prefixo+"/webhooks", handlers.ListarWebhooks)
			sub.Post(prefixo+"/webhooks", handlers.RegistrarWebhook)
			sub.Delete(prefixo+"/webhooks/{id}", handlers.RemoverWebhook)
			sub.Post(prefixo+"/webhooks/testar", handlers.TestarWebhook)
		})

		// Rotas autenticadas COM dedução de cota (cobrar)
		r.Group(func(sub chi.Router) {
			sub.Use(MiddlewareAuth(servicoAuth, db, true))

			sub.Get(prefixo+"/processos/{numero}", handlers.Processo)
			sub.Post(prefixo+"/processos/consultar", handlers.ConsultarProcessoJSON)
			sub.Get(prefixo+"/oab/{uf}/{numero}/processos", handlers.OABProcessos)
			sub.Get(prefixo+"/oab/{uf}/{numero}/publicacoes", handlers.OABPublicacoes)
		})
	}

	registrarRotas("/v1")
	registrarRotas("/api/v1")

	return r
}
