package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// WebhookSubscription representa uma inscrição de monitoramento Push (OAB ou Processo CNJ)
type WebhookSubscription struct {
	ID         string   `json:"id"`
	ClienteID  int      `json:"cliente_id"`
	TipoAlvo   string   `json:"tipo_alvo"`   // "OAB" ou "PROCESSO"
	Alvo       string   `json:"alvo"`        // Ex: "SP/123456" ou "0000001-23.2024.8.26.0100"
	URLDestino string   `json:"url_destino"` // Endpoint HTTPS do cliente
	Segredo    string   `json:"segredo,omitempty"`
	Eventos    []string `json:"eventos"` // Ex: ["nova_publicacao", "nova_movimentacao"]
	Ativo      bool     `json:"ativo"`
	CriadoEm   string   `json:"criado_em"`
}

// PayloadEvento representa o envelope disparado via POST para o endpoint do assinante
type PayloadEvento struct {
	IDEvento     string `json:"id_evento"`
	Evento       string `json:"evento"`
	DisparadoEm  string `json:"disparado_em"`
	TipoAlvo     string `json:"tipo_alvo"`
	Alvo         string `json:"alvo"`
	Dados        any    `json:"dados"`
	Assinatura   string `json:"assinatura_hmac_sha256"`
}

// GerenciadorWebhooks gerencia inscrições e disparos assinados com HMAC-SHA256
type GerenciadorWebhooks struct {
	mu         sync.RWMutex
	inscricoes map[string]WebhookSubscription
	seq        int
	httpClient *http.Client
}

func NovoGerenciadorWebhooks() *GerenciadorWebhooks {
	return &GerenciadorWebhooks{
		inscricoes: make(map[string]WebhookSubscription),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *GerenciadorWebhooks) Registrar(clienteID int, tipoAlvo, alvo, urlDestino, segredo string, eventos []string) (*WebhookSubscription, error) {
	tipoAlvo = strings.ToUpper(strings.TrimSpace(tipoAlvo))
	if tipoAlvo == "" {
		tipoAlvo = "PROCESSO"
	}
	if tipoAlvo != "PROCESSO" && tipoAlvo != "OAB" {
		return nil, fmt.Errorf("tipo_alvo deve ser 'PROCESSO' ou 'OAB'")
	}

	alvo = strings.TrimSpace(alvo)
	if alvo == "" {
		return nil, fmt.Errorf("alvo (numero do processo ou OAB) e obrigatorio")
	}

	urlDestino = strings.TrimSpace(urlDestino)
	if !strings.HasPrefix(urlDestino, "http://") && !strings.HasPrefix(urlDestino, "https://") {
		return nil, fmt.Errorf("url_destino deve iniciar com http:// ou https://")
	}

	if len(eventos) == 0 {
		eventos = []string{"nova_publicacao", "nova_movimentacao"}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.seq++
	id := fmt.Sprintf("whk_%d_%d", time.Now().Unix(), g.seq)

	sub := WebhookSubscription{
		ID:         id,
		ClienteID:  clienteID,
		TipoAlvo:   tipoAlvo,
		Alvo:       alvo,
		URLDestino: urlDestino,
		Segredo:    segredo,
		Eventos:    eventos,
		Ativo:      true,
		CriadoEm:   time.Now().UTC().Format(time.RFC3339),
	}

	g.inscricoes[id] = sub
	return &sub, nil
}

func (g *GerenciadorWebhooks) Listar(clienteID int) []WebhookSubscription {
	g.mu.RLock()
	defer g.mu.RUnlock()

	lista := make([]WebhookSubscription, 0)
	for _, sub := range g.inscricoes {
		if sub.ClienteID == clienteID || clienteID == 0 {
			// Oculta segredo na listagem pública
			copia := sub
			copia.Segredo = ""
			lista = append(lista, copia)
		}
	}
	return lista
}

func (g *GerenciadorWebhooks) Remover(clienteID int, id string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	sub, ok := g.inscricoes[id]
	if !ok {
		return false
	}
	if clienteID != 0 && sub.ClienteID != clienteID {
		return false
	}
	delete(g.inscricoes, id)
	return true
}

// CalcularAssinaturaHMAC gera a assinatura HMAC-SHA256 no padrão "sha256=<hex>"
func CalcularAssinaturaHMAC(payload []byte, segredo string) string {
	mac := hmac.New(sha256.New, []byte(segredo))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// DispararTeste monta um evento assinado com HMAC-SHA256 e opcionalmente envia ao endpoint
func (g *GerenciadorWebhooks) DispararTeste(ctx context.Context, urlDestino, segredo, evento, alvo string, enviarHTTP bool) (*PayloadEvento, int, error) {
	if evento == "" {
		evento = "nova_publicacao"
	}
	if alvo == "" {
		alvo = "0000001-23.2024.8.26.0100"
	}
	if segredo == "" {
		segredo = "segredo_padrao_webhook"
	}

	evt := &PayloadEvento{
		IDEvento:    fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Evento:      evento,
		DisparadoEm: time.Now().UTC().Format(time.RFC3339),
		TipoAlvo:    "PROCESSO",
		Alvo:        alvo,
		Dados: map[string]any{
			"numero_processo": alvo,
			"tribunal":        "TJSP",
			"resumo":          "Intimação disponibilizada no Diário da Justiça Eletrônico Nacional (DJEN).",
		},
	}

	rawBytes, _ := json.Marshal(evt)
	assinatura := CalcularAssinaturaHMAC(rawBytes, segredo)
	evt.Assinatura = assinatura

	if !enviarHTTP || urlDestino == "" {
		return evt, http.StatusOK, nil
	}

	finalBytes, _ := json.Marshal(evt)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlDestino, bytes.NewReader(finalBytes))
	if err != nil {
		return evt, 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", "Robo-Juridico-Webhook/1.0")
	req.Header.Set("X-Webhook-Signature", assinatura)
	req.Header.Set("X-Webhook-Event", evento)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return evt, 0, err
	}
	defer resp.Body.Close()

	return evt, resp.StatusCode, nil
}
