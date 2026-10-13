package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/auth"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
)

type contextKey string

const (
	requestIDKey  contextKey = "request_id"
	identidadeKey contextKey = "identidade"
	usoIDKey      contextKey = "uso_id"
	cotaInfoKey   contextKey = "cota_info"
)

// CotaInfo armazena dados de consumo de cota durante o ciclo da requisição
type CotaInfo struct {
	Limite   int
	Usado    int
	Cobravel bool
}

// ObterRequestID retorna o ID da requisição salvo no contexto
func ObterRequestID(r *http.Request) string {
	if val, ok := r.Context().Value(requestIDKey).(string); ok && val != "" {
		return val
	}
	return ""
}

// ObterIdentidade retorna a identidade do cliente autenticado
func ObterIdentidade(r *http.Request) *auth.Identidade {
	if val, ok := r.Context().Value(identidadeKey).(*auth.Identidade); ok {
		return val
	}
	return nil
}

// ObterIP extrai o IP real do cliente considerando cabeçalhos de proxy
func ObterIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		partes := strings.Split(xff, ",")
		if len(partes) > 0 {
			ip := strings.TrimSpace(partes[0])
			if ip != "" {
				return ip
			}
		}
	}
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
	r             *http.Request
}

func (rec *statusRecorder) WriteHeader(code int) {
	if rec.headerWritten {
		return
	}
	rec.statusCode = code
	rec.headerWritten = true

	// Se houver cota registrada, atualiza os cabeçalhos X-Cota-*
	if val := rec.r.Context().Value(cotaInfoKey); val != nil {
		if cota, ok := val.(*CotaInfo); ok {
			usado := cota.Usado
			if cota.Cobravel && code >= 400 {
				usado-- // Não foi cobrada devido a falha
			}
			limite := cota.Limite
			restante := limite - usado
			if restante < 0 {
				restante = 0
			}
			rec.Header().Set("X-Cota-Limite", strconv.Itoa(limite))
			rec.Header().Set("X-Cota-Usada", strconv.Itoa(usado))
			rec.Header().Set("X-Cota-Restante", strconv.Itoa(restante))
		}
	}

	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if !rec.headerWritten {
		rec.WriteHeader(http.StatusOK)
	}
	return rec.ResponseWriter.Write(b)
}

// MiddlewareEnvelope adiciona RequestID, segurança, auditoria e finalização de uso
func MiddlewareEnvelope(db storage.BancoDados) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			reqID := hex.EncodeToString(b)

			ctx := context.WithValue(r.Context(), requestIDKey, reqID)
			r = r.WithContext(ctx)

			// Cabeçalhos de Segurança Padrão
			w.Header().Set("X-Request-ID", reqID)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
			w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")
			w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			inicio := time.Now()
			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				r:              r,
			}

			next.ServeHTTP(rec, r)

			duracaoMs := int(time.Since(inicio).Milliseconds())
			ip := ObterIP(r)
			clienteNome := "-"
			if id := ObterIdentidade(r); id != nil {
				clienteNome = id.NomeCliente
			}

			log.Printf("[AUDITORIA] id=%s cliente=%s ip=%s %s %s status=%d ms=%d",
				reqID, clienteNome, ip, r.Method, r.URL.Path, rec.statusCode, duracaoMs)

			// Finaliza uso no banco de dados se houve reserva
			if db != nil {
				if usoIDVal := r.Context().Value(usoIDKey); usoIDVal != nil {
					if usoID, ok := usoIDVal.(int64); ok && usoID > 0 {
						_ = db.FinalizarUso(context.Background(), usoID, rec.statusCode, duracaoMs, r.URL.Path)
					}
				}
			}
		})
	}
}

// MiddlewareAuth cria o validador de autenticação e controle de cotas
func MiddlewareAuth(servicoAuth *auth.ServicoAuth, db storage.BancoDados, cobravel bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ObterIP(r)

			chave := strings.TrimSpace(r.Header.Get("X-API-Key"))
			identidade, err := servicoAuth.Autenticar(r.Context(), chave, ip)
			if err != nil {
				if appErr, ok := err.(*domain.ErroAPI); ok {
					if appErr.TentarEm > 0 {
						w.Header().Set("Retry-After", strconv.Itoa(appErr.TentarEm))
					}
					EnviarErro(w, r, appErr.Status, appErr.Codigo, appErr.Mensagem)
					return
				}
				EnviarErro(w, r, http.StatusUnauthorized, "nao_autorizado", "Envie uma chave válida no cabeçalho X-API-Key.")
				return
			}

			// 2. Limite de requisições por minuto
			permitidoTaxa := servicoAuth.VerificarLimiteTaxa(identidade)
			if !permitidoTaxa {
				w.Header().Set("Retry-After", "60")
				EnviarErro(w, r, http.StatusTooManyRequests, "limite_excedido",
					fmt.Sprintf("Limite de %d requisições por minuto atingido.", identidade.LimitePorMinuto))
				return
			}

			ctx := context.WithValue(r.Context(), identidadeKey, identidade)

			// 3. Cota mensal e auditoria no banco
			if db != nil && !identidade.EhInterna && identidade.LimiteMensal != nil {
				permitido, usoID, usado, errRes := db.ReservarCota(r.Context(), identidade.ClienteID,
					identidade.ChaveID, *identidade.LimiteMensal, r.Method, r.URL.Path, ip, cobravel)

				cotaInfo := &CotaInfo{
					Limite:   *identidade.LimiteMensal,
					Usado:    usado,
					Cobravel: cobravel && permitido,
				}
				ctx = context.WithValue(ctx, cotaInfoKey, cotaInfo)

				if errRes != nil || !permitido {
					w.Header().Set("X-Cota-Limite", strconv.Itoa(*identidade.LimiteMensal))
					w.Header().Set("X-Cota-Usada", strconv.Itoa(usado))
					w.Header().Set("X-Cota-Restante", strconv.Itoa(max(0, *identidade.LimiteMensal-usado)))

					rComCtx := r.WithContext(ctx)
					EnviarErro(w, rComCtx, http.StatusPaymentRequired, "cota_esgotada",
						fmt.Sprintf("A cota de %d consultas do mês acabou. Ela renova no dia 1º; para ampliar agora, fale com o suporte.",
							*identidade.LimiteMensal))
					return
				}

				ctx = context.WithValue(ctx, usoIDKey, usoID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
