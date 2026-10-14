package http_test

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/auth"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cache"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/comunica"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/fontes/datajud"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/network"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/service"
	transporthttp "github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/transport/http"
)

const chaveTesteValida = "teste-chave-secreta-longa-com-mais-de-32-caracteres"

func setupRoteadorTeste() http.Handler {
	httpCli := network.NovoClienteHTTP(30, "teste")
	dj := datajud.NovoClienteDataJud(httpCli, "key-fake", 5, 1)
	com := comunica.NovoClienteComunica(httpCli, nil, 5, 10, 1)
	c := cache.NovoCacheTTL()
	orq := service.NovoOrquestrador(dj, com, c)

	chavesInternas := map[string][32]byte{
		"cliente-teste": sha256.Sum256([]byte(chaveTesteValida)),
	}
	servicoAuth := auth.NovoServicoAuth(nil, chavesInternas, 60)
	handlers := transporthttp.NovosHandlers(orq, nil, "1.0.0")

	return transporthttp.NovoRoteador(handlers, servicoAuth, nil)
}

func TestRotaSaude(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/saude", nil)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado status 200, obtido %d", w.Code)
	}

	var corpo map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("Erro ao decodificar JSON: %v", err)
	}

	if corpo["ok"] != true || corpo["versao"] != "1.0.0" {
		t.Errorf("Resposta inesperada em /saude: %v", corpo)
	}

	// Verifica cabeçalhos de segurança padrão
	if w.Header().Get("X-Request-ID") == "" {
		t.Errorf("X-Request-ID ausente")
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("X-Content-Type-Options incorreto")
	}
}

func TestAutenticacaoSemChave(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/v1/conta", nil)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Esperado status 401, obtido %d", w.Code)
	}

	var resp transporthttp.RespostaErro
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Erro ao decodificar erro JSON: %v", err)
	}

	if resp.Ok != false || resp.Erro.Codigo != "nao_autorizado" {
		t.Errorf("Erro inesperado: %+v", resp)
	}
}

func TestDecodificarCNJ(t *testing.T) {
	roteador := setupRoteadorTeste()

	// 1. Número válido
	reqValido := httptest.NewRequest("GET", "/v1/cnj/5053283-30.2024.8.13.0079", nil)
	reqValido.Header.Set("X-API-Key", chaveTesteValida)
	wValido := httptest.NewRecorder()
	roteador.ServeHTTP(wValido, reqValido)

	if wValido.Code != http.StatusOK {
		t.Fatalf("Esperado 200 para CNJ válido, obtido %d", wValido.Code)
	}

	var respValida domain.RespostaCNJ
	if err := json.Unmarshal(wValido.Body.Bytes(), &respValida); err != nil {
		t.Fatalf("Erro ao deserializar resposta CNJ: %v", err)
	}

	if !respValida.OK || !respValida.Valido || respValida.Tribunal != "TJMG" || respValida.Ano != 2024 {
		t.Errorf("Dados incorretos do CNJ decodificado: %+v", respValida)
	}

	// 2. Número inválido (dígito errado)
	reqInvalido := httptest.NewRequest("GET", "/v1/cnj/5053283-99.2024.8.13.0079", nil)
	reqInvalido.Header.Set("X-API-Key", chaveTesteValida)
	wInvalido := httptest.NewRecorder()
	roteador.ServeHTTP(wInvalido, reqInvalido)

	if wInvalido.Code != http.StatusOK {
		t.Fatalf("Esperado 200 com valido=false, obtido %d", wInvalido.Code)
	}

	var respInvalida domain.RespostaCNJ
	_ = json.Unmarshal(wInvalido.Body.Bytes(), &respInvalida)
	if respInvalida.Valido {
		t.Errorf("Esperava valido=false para dígito errado")
	}
}

func TestRotaContaAutenticada(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/v1/conta", nil)
	req.Header.Set("X-API-Key", chaveTesteValida)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado status 200, obtido %d", w.Code)
	}

	var resp domain.RespostaConta
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Erro ao decodificar JSON conta: %v", err)
	}

	if !resp.OK || resp.Cliente != "cliente-teste" || resp.Plano != "interno" {
		t.Errorf("Resposta de conta inesperada: %+v", resp)
	}
}

func TestProcessoNumeroInvalidoRetorna422(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/v1/processos/123-numero-invalido", nil)
	req.Header.Set("X-API-Key", chaveTesteValida)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Esperado 422 para processo com formato inválido, obtido %d", w.Code)
	}

	var resp transporthttp.RespostaErro
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Erro.Codigo != "numero_invalido" {
		t.Errorf("Esperava código 'numero_invalido', obtido '%s'", resp.Erro.Codigo)
	}
}

func TestRotaNaoEncontradaUniforme404(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/rota-que-nao-existe", nil)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Esperado 404, obtido %d", w.Code)
	}

	var resp transporthttp.RespostaErro
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Erro.Codigo != "rota_inexistente" {
		t.Errorf("Esperava código 'rota_inexistente', obtido '%s'", resp.Erro.Codigo)
	}
}

func TestSwaggerUIERotasDocs(t *testing.T) {
	roteador := setupRoteadorTeste()

	// 1. Testa /openapi.json
	reqJSON := httptest.NewRequest("GET", "/openapi.json", nil)
	wJSON := httptest.NewRecorder()
	roteador.ServeHTTP(wJSON, reqJSON)

	if wJSON.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /openapi.json, obtido %d", wJSON.Code)
	}
	var openapiSpec map[string]any
	if err := json.Unmarshal(wJSON.Body.Bytes(), &openapiSpec); err != nil {
		t.Fatalf("Erro ao decodificar /openapi.json: %v", err)
	}
	if openapiSpec["openapi"] != "3.1.0" {
		t.Errorf("Versão OpenAPI incorreta: %v", openapiSpec["openapi"])
	}

	// 2. Testa /docs (Swagger UI HTML)
	reqDocs := httptest.NewRequest("GET", "/docs", nil)
	wDocs := httptest.NewRecorder()
	roteador.ServeHTTP(wDocs, reqDocs)

	if wDocs.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /docs, obtido %d", wDocs.Code)
	}
	if wDocs.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("Content-Type incorreto em /docs: %s", wDocs.Header().Get("Content-Type"))
	}
}

func TestAutenticacaoComBearerToken(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("GET", "/v1/conta", nil)
	req.Header.Set("Authorization", "Bearer "+chaveTesteValida)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado 200 via Bearer Token, obtido %d", w.Code)
	}

	var corpo map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &corpo); err != nil {
		t.Fatalf("Erro ao decodificar JSON: %v", err)
	}
	if corpo["cliente"] != "cliente-teste" {
		t.Errorf("Cliente incorreto na resposta: %v", corpo["cliente"])
	}
}

func TestMiddlewareCORS_OptionsPreflight(t *testing.T) {
	roteador := setupRoteadorTeste()

	req := httptest.NewRequest("OPTIONS", "/v1/processos/5053283-30.2024.8.13.0079", nil)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("Esperado 204 No Content para OPTIONS preflight, obtido %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS Allow-Origin incorreto: %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("CORS Allow-Methods ausente")
	}
}

func TestAliasRotaApiV1(t *testing.T) {
	roteador := setupRoteadorTeste()

	// Testa chamada através do alias /api/v1/cnj/...
	req := httptest.NewRequest("GET", "/api/v1/cnj/5053283-30.2024.8.13.0079", nil)
	req.Header.Set("Authorization", "Bearer "+chaveTesteValida)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Esperado 200 no alias /api/v1/cnj/..., obtido %d", w.Code)
	}

	var resp domain.RespostaCNJ
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Erro ao decodificar resposta: %v", err)
	}
	if !resp.Valido || resp.Tribunal != "TJMG" {
		t.Errorf("Decodificação no alias falhou: %+v", resp)
	}
}

func TestConsultarProcessoJSON_Validacao(t *testing.T) {
	roteador := setupRoteadorTeste()

	// 1. JSON inválido (número CNJ inválido)
	payloadInvalido := `{"numero_processo": "123-numero-invalido"}`
	req := httptest.NewRequest("POST", "/v1/processos/consultar", strings.NewReader(payloadInvalido))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", chaveTesteValida)
	w := httptest.NewRecorder()
	roteador.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Esperado 422 para número CNJ inválido no POST, obtido %d", w.Code)
	}

	// 2. JSON malformado
	reqBad := httptest.NewRequest("POST", "/api/v1/processos/consultar", strings.NewReader("nao-e-json"))
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.Header.Set("X-API-Key", chaveTesteValida)
	wBad := httptest.NewRecorder()
	roteador.ServeHTTP(wBad, reqBad)

	if wBad.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Esperado 422 para JSON malformado no POST, obtido %d", wBad.Code)
	}
}
