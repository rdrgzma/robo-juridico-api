package http

import (
	"encoding/json"
	"net/http"
)

// DetalheErro descreve um erro de validação específico de campo
type DetalheErro struct {
	Campo    string `json:"campo"`
	Mensagem string `json:"mensagem"`
}

// CorpoErro descreve a estrutura interna de erro
type CorpoErro struct {
	Codigo   string        `json:"codigo"`
	Mensagem string        `json:"mensagem"`
	Detalhes []DetalheErro `json:"detalhes,omitempty"`
}

// RespostaErro representa a resposta padrão para erros da API
type RespostaErro struct {
	Ok           bool      `json:"ok"`
	Erro         CorpoErro `json:"erro"`
	IDRequisicao string    `json:"id_requisicao,omitempty"`
}

// RespostaGenerica representa a resposta de sucesso padrão
type RespostaGenerica struct {
	Ok     bool   `json:"ok"`
	Versao string `json:"versao,omitempty"`
}

// JSON envia uma resposta JSON com o status HTTP informado
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// EnviarErro envia uma resposta de erro no padrão canônico da API
func EnviarErro(w http.ResponseWriter, r *http.Request, status int, codigo, mensagem string, detalhes ...DetalheErro) {
	reqID := ObterRequestID(r)
	resp := RespostaErro{
		Ok: false,
		Erro: CorpoErro{
			Codigo:   codigo,
			Mensagem: mensagem,
			Detalhes: detalhes,
		},
		IDRequisicao: reqID,
	}
	JSON(w, status, resp)
}
