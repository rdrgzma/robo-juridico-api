# ==============================================================================
# Dockerfile Multi-stage para o Painel Administrativo (PHP 8.4 + Filament / Laravel)
# ==============================================================================

# --- Estágio 1: Instalação de Dependências via Composer ---
FROM composer:2 AS vendor

WORKDIR /app

COPY composer.json ./

RUN composer install \
    --no-dev \
    --no-interaction \
    --prefer-dist \
    --ignore-platform-reqs \
    --optimize-autoloader

# --- Estágio 2: Runtime PHP 8.4 ---
FROM php:8.4-cli-alpine

# Instala dependências de compilação/execução e fuso horário oficial de Brasília
RUN apk add --no-cache \
    tzdata \
    sqlite-dev \
    icu-dev \
    libzip-dev \
    curl \
    bash && \
    cp /usr/share/zoneinfo/America/Sao_Paulo /etc/localtime && \
    echo "America/Sao_Paulo" > /etc/timezone && \
    docker-php-ext-install \
    pdo_sqlite \
    bcmath \
    intl \
    zip \
    opcache

WORKDIR /app

# Copia dependências instaladas pelo Composer
COPY --from=vendor /app/vendor /app/vendor

# Copia arquivos da aplicação
COPY . .

# Garante permissões em storage, cache e script de inicialização
RUN mkdir -p /app/storage/framework/cache \
             /app/storage/framework/sessions \
             /app/storage/framework/views \
             /app/storage/logs \
             /app/bootstrap/cache \
             /app/dados && \
    chmod -R 777 /app/storage /app/bootstrap/cache /app/dados && \
    chmod +x /app/docker-entrypoint.sh

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/up || exit 1

ENTRYPOINT ["/app/docker-entrypoint.sh"]
