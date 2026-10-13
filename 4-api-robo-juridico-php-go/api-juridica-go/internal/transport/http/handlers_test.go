package http_test

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
