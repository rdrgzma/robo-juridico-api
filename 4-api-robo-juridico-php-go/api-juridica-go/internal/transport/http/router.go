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
	r.Use(chimiddleware.Recoverer)
	r.Use(MiddlewareEnvelope(db))

	// Tratamento uniforme para 404 e 405
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		EnviarErro(w, req, http.StatusNotFound, "rota_inexistente", "Rota não encontrada.")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		EnviarErro(w, req, http.StatusMethodNotAllowed, "metodo_nao_permitido", "Método HTTP não permitido.")
	})

	// Rota pública de saúde
	r.Get("/saude", handlers.Saude)

	// Rotas autenticadas SEM custo de cota (sem_custo)
	r.Group(func(sub chi.Router) {
		sub.Use(MiddlewareAuth(servicoAuth, db, false))

		sub.Get("/v1/cnj/{numero}", handlers.DecodificarCNJ)
		sub.Get("/v1/processos/{numero}/existe", handlers.ProcessoExiste)
		sub.Get("/v1/conta", handlers.Conta)
	})

	// Rotas autenticadas COM dedução de cota (cobrar)
	r.Group(func(sub chi.Router) {
		sub.Use(MiddlewareAuth(servicoAuth, db, true))

		sub.Get("/v1/processos/{numero}", handlers.Processo)
		sub.Get("/v1/oab/{uf}/{numero}/processos", handlers.OABProcessos)
		sub.Get("/v1/oab/{uf}/{numero}/publicacoes", handlers.OABPublicacoes)
	})

	return r
}
