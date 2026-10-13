# ==============================================================================
# Multi-stage Dockerfile para Robô Jurídico API (Go)
# Binário estático puro sem dependência de CGO (usando modernc.org/sqlite)
# ==============================================================================

# --- Estágio 1: Build ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instala certificados necessários para baixar módulos
RUN apk add --no-cache ca-certificates git

# Cache de dependências
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Código-fonte
COPY . .

# Compilação estática pura sem CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=1.0.0" \
    -o /app/bin/api ./cmd/api

# --- Estágio 2: Runtime Seguro e Leve ---
FROM alpine:3.21

# Certificados de autoridade e fuso horário do Brasil
RUN apk add --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && \
    echo "America/Sao_Paulo" > /etc/timezone

# Cria usuário não-root por segurança
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Cria diretório para persistência de dados do SQLite
RUN mkdir -p /app/dados && chown -R appuser:appgroup /app

# Copia binário do builder
COPY --from=builder /app/bin/api /app/api
RUN chmod +x /app/api

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/saude || exit 1

ENTRYPOINT ["/app/api"]
