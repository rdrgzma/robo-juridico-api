package http_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rdrgzma/robo-juridico-desktop/internal/auth"
	"github.com/rdrgzma/robo-juridico-desktop/internal/cache"
	"github.com/rdrgzma/robo-juridico-desktop/internal/domain"
	"github.com/rdrgzma/robo-juridico-desktop/internal/fontes/comunica"
	"github.com/rdrgzma/robo-juridico-desktop/internal/fontes/datajud"
	"github.com/rdrgzma/robo-juridico-desktop/internal/network"
	"github.com/rdrgzma/robo-juridico-desktop/internal/service"
	"github.com/rdrgzma/robo-juridico-desktop/internal/storage"
	transporthttp "github.com/rdrgzma/robo-juridico-desktop/internal/transport/http"
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

func TestRotasCalculoJudicial(t *testing.T) {
	roteador := setupRoteadorTeste()

	// 1. GET /v1/calculos/indices
	reqIndices := httptest.NewRequest("GET", "/v1/calculos/indices", nil)
	reqIndices.Header.Set("X-API-Key", chaveTesteValida)
	wIndices := httptest.NewRecorder()
	roteador.ServeHTTP(wIndices, reqIndices)

	if wIndices.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/calculos/indices, obtido %d", wIndices.Code)
	}

	var resIndices map[string]any
	if err := json.Unmarshal(wIndices.Body.Bytes(), &resIndices); err != nil {
		t.Fatalf("Erro ao decodificar JSON dos indices: %v", err)
	}
	if resIndices["ok"] != true {
		t.Errorf("Esperado ok=true em indices: %v", resIndices)
	}

	// 2. GET /v1/calculos/indices/TJSP/serie
	reqSerie := httptest.NewRequest("GET", "/v1/calculos/indices/TJSP/serie", nil)
	reqSerie.Header.Set("X-API-Key", chaveTesteValida)
	wSerie := httptest.NewRecorder()
	roteador.ServeHTTP(wSerie, reqSerie)

	if wSerie.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/calculos/indices/TJSP/serie, obtido %d", wSerie.Code)
	}

	// 3. POST /v1/calculos/judicial
	payload := `{
		"data_atualizacao": "2026-10-14",
		"indexador_codigo": "TJSP",
		"aplicar_deflacao": true,
		"juros_mora": {
			"tipo": "CONVENCIONAL_MENSAL",
			"percentual_mensal": 1.0,
			"modo": "COMPOSTO",
			"aplicar_prorata_die": true
		},
		"fase_execucao": {
			"aplicar_art_523_multa": true,
			"aplicar_art_523_honorarios": true,
			"multa_moratoria_percentual": 10.0
		},
		"parcelas": [
			{
				"numero_item": 1,
				"descricao": "Parcela Inicial",
				"tipo": "DEBITO",
				"data_vencimento": "2023-01-15",
				"valor_original": 10000.00
			}
		]
	}`

	reqCalc := httptest.NewRequest("POST", "/v1/calculos/judicial", strings.NewReader(payload))
	reqCalc.Header.Set("Content-Type", "application/json")
	reqCalc.Header.Set("X-API-Key", chaveTesteValida)
	wCalc := httptest.NewRecorder()
	roteador.ServeHTTP(wCalc, reqCalc)

	if wCalc.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/calculos/judicial, obtido %d: %s", wCalc.Code, wCalc.Body.String())
	}

	var resCalc map[string]any
	if err := json.Unmarshal(wCalc.Body.Bytes(), &resCalc); err != nil {
		t.Fatalf("Erro ao decodificar JSON de calculo judicial: %v", err)
	}

	if resCalc["sucesso"] != true {
		t.Errorf("Esperado sucesso=true no calculo: %v", resCalc)
	}

	// Verifica presença dos campos unificados para o painel PHP e API Go
	if resCalc["itens"] == nil || resCalc["itens_liquidados"] == nil {
		t.Errorf("Esperava ambos 'itens' e 'itens_liquidados' no JSON de resposta")
	}

	// 4. POST /v1/calculos/exportar-planilha?formato=xlsx
	reqExport := httptest.NewRequest("POST", "/v1/calculos/exportar-planilha?formato=xlsx", strings.NewReader(payload))
	reqExport.Header.Set("Content-Type", "application/json")
	reqExport.Header.Set("X-API-Key", chaveTesteValida)
	wExport := httptest.NewRecorder()
	roteador.ServeHTTP(wExport, reqExport)

	if wExport.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/calculos/exportar-planilha, obtido %d", wExport.Code)
	}
	if !strings.Contains(wExport.Header().Get("Content-Type"), "spreadsheetml.sheet") {
		t.Errorf("Content-Type inesperado na exportacao XLSX: %s", wExport.Header().Get("Content-Type"))
	}

	// 5. POST /v1/calculos/financiamento (Método de Gauss)
	payloadFin := `{
		"valor_financiado": 40000,
		"numero_parcelas": 24,
		"taxa_contratada_mensal": 2.5,
		"taxa_mercado_mensal": 1.5,
		"sistema_amortizacao": "GAUSS",
		"restituicao_em_dobro": true
	}`
	reqFin := httptest.NewRequest("POST", "/v1/calculos/financiamento", strings.NewReader(payloadFin))
	reqFin.Header.Set("Content-Type", "application/json")
	reqFin.Header.Set("X-API-Key", chaveTesteValida)
	wFin := httptest.NewRecorder()
	roteador.ServeHTTP(wFin, reqFin)

	if wFin.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/calculos/financiamento, obtido %d: %s", wFin.Code, wFin.Body.String())
	}

	// 6. POST /v1/prazos/calcular e GET /v1/prazos/feriados
	payloadPrazo := `{
		"data_disponibilizacao": "2026-10-08",
		"dias_prazo": 15,
		"rito": "CPC_UTEIS",
		"tipo_intimacao": "DIARIO_ELETRONICO"
	}`
	reqPrazo := httptest.NewRequest("POST", "/v1/prazos/calcular", strings.NewReader(payloadPrazo))
	reqPrazo.Header.Set("Content-Type", "application/json")
	reqPrazo.Header.Set("X-API-Key", chaveTesteValida)
	wPrazo := httptest.NewRecorder()
	roteador.ServeHTTP(wPrazo, reqPrazo)

	if wPrazo.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/prazos/calcular, obtido %d: %s", wPrazo.Code, wPrazo.Body.String())
	}

	reqFeriados := httptest.NewRequest("GET", "/v1/prazos/feriados?ano=2026", nil)
	reqFeriados.Header.Set("X-API-Key", chaveTesteValida)
	wFeriados := httptest.NewRecorder()
	roteador.ServeHTTP(wFeriados, reqFeriados)
	if wFeriados.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em /v1/prazos/feriados, obtido %d", wFeriados.Code)
	}

	// 7. POST /v1/webhooks, GET /v1/webhooks e POST /v1/webhooks/testar
	payloadWh := `{
		"tipo_alvo": "OAB",
		"alvo": "SP/123456",
		"url_destino": "https://exemplo.adv.br/webhook",
		"segredo": "meu-segredo-hmac"
	}`
	reqWh := httptest.NewRequest("POST", "/v1/webhooks", strings.NewReader(payloadWh))
	reqWh.Header.Set("Content-Type", "application/json")
	reqWh.Header.Set("X-API-Key", chaveTesteValida)
	wWh := httptest.NewRecorder()
	roteador.ServeHTTP(wWh, reqWh)
	if wWh.Code != http.StatusCreated {
		t.Fatalf("Esperado 201 em POST /v1/webhooks, obtido %d: %s", wWh.Code, wWh.Body.String())
	}

	reqWhTest := httptest.NewRequest("POST", "/v1/webhooks/testar", strings.NewReader(`{"segredo":"meu-segredo-hmac","enviar_http":false}`))
	reqWhTest.Header.Set("Content-Type", "application/json")
	reqWhTest.Header.Set("X-API-Key", chaveTesteValida)
	wWhTest := httptest.NewRecorder()
	roteador.ServeHTTP(wWhTest, reqWhTest)
	if wWhTest.Code != http.StatusOK {
		t.Fatalf("Esperado 200 em POST /v1/webhooks/testar, obtido %d", wWhTest.Code)
	}
}

type mockBancoComAuditoria struct {
	reservado      bool
	finalizado     bool
	finalizadoID   int64
	statusGravado  int
	rotaGravada    string
	msGravado      int
}

func (m *mockBancoComAuditoria) ValidarChave(ctx context.Context, hashChave string) (*storage.ChaveCliente, error) {
	return &storage.ChaveCliente{
		ChaveID:         10,
		ClienteID:       1,
		NomeCliente:     "Cliente Teste",
		Ativo:           true,
		Plano:           "pro",
		LimiteMensal:    1000,
		LimitePorMinuto: 60,
	}, nil
}

func (m *mockBancoComAuditoria) ReservarCota(ctx context.Context, clienteID, chaveID, limiteMensal int, metodo, caminho, ip string, cobravel bool) (bool, int64, int, error) {
	m.reservado = true
	return true, 12345, 1, nil
}

func (m *mockBancoComAuditoria) FinalizarUso(ctx context.Context, usoID int64, status, ms int, rota string) error {
	m.finalizado = true
	m.finalizadoID = usoID
	m.statusGravado = status
	m.msGravado = ms
	m.rotaGravada = rota
	return nil
}

func (m *mockBancoComAuditoria) ObterConsumoMes(ctx context.Context, clienteID int) (int, error) {
	return 1, nil
}

func (m *mockBancoComAuditoria) Fechar() error {
	return nil
}

func TestFinalizarUsoGravadoCorretamente(t *testing.T) {
	mockDB := &mockBancoComAuditoria{}
	httpCli := network.NovoClienteHTTP(30, "teste")
	dj := datajud.NovoClienteDataJud(httpCli, "key-fake", 5, 1)
	com := comunica.NovoClienteComunica(httpCli, nil, 5, 10, 1)
	c := cache.NovoCacheTTL()
	orq := service.NovoOrquestrador(dj, com, c)

	servicoAuth := auth.NovoServicoAuth(mockDB, nil, 60)
	handlers := transporthttp.NovosHandlers(orq, nil, "1.0.0")
	roteador := transporthttp.NovoRoteador(handlers, servicoAuth, mockDB)

	req := httptest.NewRequest("GET", "/v1/processos/5053283-30.2024.8.13.0079/existe", nil)
	req.Header.Set("X-API-Key", "qualquer-chave-cliente")
	w := httptest.NewRecorder()

	roteador.ServeHTTP(w, req)

	if !mockDB.reservado {
		t.Fatal("Esperava que ReservarCota tivesse sido chamado")
	}
	if !mockDB.finalizado {
		t.Fatal("Esperava que FinalizarUso tivesse sido chamado")
	}
	if mockDB.finalizadoID != 12345 {
		t.Errorf("Esperado usoID 12345, obtido %d", mockDB.finalizadoID)
	}
	if mockDB.statusGravado != w.Code {
		t.Errorf("Esperado status %d gravado, obtido %d", w.Code, mockDB.statusGravado)
	}
	if mockDB.rotaGravada != "/v1/processos/5053283-30.2024.8.13.0079/existe" {
		t.Errorf("Esperada rota '/v1/processos/5053283-30.2024.8.13.0079/existe', obtida '%s'", mockDB.rotaGravada)
	}
}

