# Guia de Deploy e Operação em Produção: Robô Jurídico API

> **Documento:** DEPLOY-RJ-001  
> **Versão:** 1.0.0  
> **Aplica-se a:** `api-juridica-go` (Go 1.24+ / Alpine Container / Bare-Metal Linux)  
> **Data:** Outubro de 2026  

---

## 1. Visão Geral da Arquitetura de Implantação

A API em Go foi projetada como um **binário estático puro sem dependência de CGO** (`CGO_ENABLED=0`), gerando uma aplicação de altíssima eficiência, baixíssimo consumo de memória (geralmente entre **18 MB e 45 MB** de RAM em execução) e tempo de inicialização inferior a **100 milissegundos**.

### 1.1. Topologia de Produção Recomendada

```
                         [ Clientes da API / Painel Web ]
                                       │
                                       ▼ HTTPS (Porta 443)
                         ┌─────────────────────────────┐
                         │      Cloudflare / WAF       │ (Opcional: Proteção DDoS)
                         └─────────────┬───────────────┘
                                       │
                                       ▼
                         ┌─────────────────────────────┐
                         │   Proxy Reverso & TLS       │
                         │   (Caddy ou Nginx + SSL)    │
                         └─────────────┬───────────────┘
                                       │ HTTP Local (127.0.0.1:8000)
                                       ▼
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │ Servidor Linux (Docker ou Bare-Metal)                                       │
 │                                                                             │
 │   ┌─────────────────────────────────────────────────────────────────────┐   │
 │   │ Robô Jurídico API (Go)                                             │   │
 │   │  - Roteamento Chi v5 & Middlewares de Segurança                     │   │
 │   │  - Autenticação SHA-256 e Rate Limiter em Memória                   │   │
 │   │  - Singleflight Cache Anti-Thundering Herd                          │   │
 │   │  - Orquestrador Concorrente de Consultas                            │   │
 │   └───────────────┬──────────────────────────────┬──────────────────────┘   │
 │                   │                              │                          │
 │                   ▼                              ▼                          │
 │       ┌────────────────────────┐     ┌────────────────────────┐             │
 │       │ Banco SQLite (Modo WAL)│     │ Pool Proxies / Gateway │             │
 │       │ /dados/api.db          │     │ Residencial Rotativo   │             │
 │       └────────────────────────┘     └───────────┬────────────┘             │
 └──────────────────────────────────────────────────┼──────────────────────────┘
                                                    │
                                     ┌──────────────┴──────────────┐
                                     ▼                             ▼
                            [ DataJud (CNJ) ]            [ DJEN (ComunicaPJe) ]
```

---

## 2. Requisitos Mínimos de Infraestrutura

| Componente | Requisito Mínimo (Homologação / MVP) | Recomendado para Produção (Alta Carga) |
|---|---|---|
| **Sistema Operacional** | Linux Ubuntu 22.04/24.04 LTS ou Debian 12 | Ubuntu 24.04 LTS / Debian 12 |
| **Processador (vCPU)** | 1 vCPU (x86_64 ou arm64) | 2 ou 4 vCPUs |
| **Memória RAM** | 512 MB | 2 GB ou 4 GB |
| **Armazenamento** | 10 GB SSD (NVMe recomendado) | 40 GB NVMe com backups automáticos |
| **Rede** | 100 Mbps com IP Público IPv4 | 1 Gbps com tráfego ilimitado |
| **Portas de Entrada** | `80/TCP` (HTTP) e `443/TCP` (HTTPS) | `80/TCP`, `443/TCP` e `22/TCP` (SSH restrito) |

---

## 3. Método 1: Deploy com Docker & Docker Compose (Recomendado)

O método baseado em Docker Compose oferece isolamento de processos, portabilidade imediata e gerenciamento simples de volumes persistentes.

### 3.1. Estrutura de Arquivos no Servidor

Crie a estrutura de diretórios em `/opt/robo-juridico`:

```bash
sudo mkdir -p /opt/robo-juridico/dados
sudo chown -R 1000:1000 /opt/robo-juridico
cd /opt/robo-juridico
```

### 3.2. Arquivo de Ambiente de Produção (`.env.prod`)

Crie o arquivo `/opt/robo-juridico/.env.prod`:

```env
# Configurações do Servidor
HOST=0.0.0.0
PORT=8000
API_BANCO=/app/dados/api.db
API_LIMITE_POR_MINUTO=120

# Timeouts das Fontes Oficiais
DATAJUD_TIMEOUT=90
DATAJUD_TENTATIVAS=3
COMUNICA_TIMEOUT=12
COMUNICA_TENTATIVAS=3

# Chaves Internas (cliente:chave_com_mais_de_32_chars)
API_CHAVES=admin_master:c2VjcmV0LWtleS1yb2JvLWp1cmlkaWNvLTIwMjYtZ28=,painel_interno:cGFpbmVsLWludGVybm8tc2VncmVkb3NvLXByb2R1Y2Fv

# Gateway de Proxies (Opcional, para contornar bloqueios do DJEN)
CONSULTA_PROXIES=
CONSULTA_GATEWAY_HOST=
CONSULTA_GATEWAY_LOGIN=
CONSULTA_GATEWAY_SENHA=
CONSULTA_GATEWAY_PORTA_INICIAL=10000
CONSULTA_GATEWAY_PORTAS=0
```

### 3.3. Arquivo `docker-compose.yml`

```yaml
services:
  api:
    image: ghcr.io/rdrgzma/robo-juridico-api:latest
    # Ou para build local a partir do código-fonte:
    # build:
    #   context: .
    #   dockerfile: Dockerfile
    container_name: robo-juridico-api
    restart: unless-stopped
    env_file:
      - .env.prod
    ports:
      - "127.0.0.1:8000:8000"
    volumes:
      - ./dados:/app/dados
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/saude"]
      interval: 15s
      timeout: 3s
      retries: 3
      start_period: 5s
    security_opt:
      - no-new-privileges:true
```

### 3.4. Inicialização e Comandos de Operação

```bash
# Subir o container em segundo plano
docker compose up -d

# Visualizar logs em tempo real
docker compose logs -f api

# Verificar status do healthcheck
docker compose ps

# Atualizar para a versão mais recente
docker compose pull
docker compose up -d --force-recreate
```

---

## 4. Método 2: Deploy Bare-Metal com Systemd (Sem Docker)

Para servidores que exigem performance máxima e menor overhead possível, execute o binário nativo Go gerenciado pelo `systemd`.

### 4.1. Compilação do Binário Estático

Na sua máquina de desenvolvimento ou esteira de CI/CD:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -trimpath \
  -ldflags="-s -w -X main.version=1.0.0" \
  -o api-juridica-linux ./cmd/api
```

Transfira o binário para o servidor:
```bash
scp api-juridica-linux usuario@seu-servidor:/tmp/api-juridica
```

### 4.2. Preparação do Servidor Linux

```bash
# Criação do usuário de sistema sem permissão de login
sudo useradd -r -s /bin/false -d /opt/robo-juridico robojuridico

# Estrutura de diretórios
sudo mkdir -p /opt/robo-juridico/bin
sudo mkdir -p /opt/robo-juridico/dados
sudo mkdir -p /etc/robo-juridico

# Instalação do binário
sudo mv /tmp/api-juridica /opt/robo-juridico/bin/api
sudo chmod +x /opt/robo-juridico/bin/api
sudo chown -R robojuridico:robojuridico /opt/robo-juridico
```

### 4.3. Criação do Serviço Systemd

Crie o arquivo `/etc/systemd/system/robo-juridico.service`:

```ini
[Unit]
Description=Robô Jurídico — API de Processos em Go
After=network.target network-online.target
Wants=network-online.target

[Service]
Type=simple
User=robojuridico
Group=robojuridico
WorkingDirectory=/opt/robo-juridico
ExecStart=/opt/robo-juridico/bin/api

# Variáveis de ambiente
EnvironmentFile=/etc/robo-juridico/api.env

# Política de reinicialização
Restart=always
RestartSec=5s

# Limites de recursos e segurança
LimitNOFILE=65535
ProtectSystem=full
ProtectHome=true
NoNewPrivileges=true
PrivateTmp=true

# Logs redirecionados ao journald
StandardOutput=journal
StandardError=journal
SyslogIdentifier=robo-juridico

[Install]
WantedBy=multi-user.target
```

### 4.4. Ativação e Gerenciamento do Serviço

```bash
# Recarrega o daemon do systemd
sudo systemctl daemon-reload

# Habilita para iniciar no boot
sudo systemctl enable robo-juridico

# Inicia o serviço
sudo systemctl start robo-juridico

# Verifica o status
sudo systemctl status robo-juridico

# Acompanhar logs
sudo journalctl -u robo-juridico -f
```

---

## 5. Configuração de Proxy Reverso e SSL/HTTPS

A API deve ser exposta na Internet através de um proxy reverso que forneça terminação TLS, compressão e proteção de cabeçalhos.

### 5.1. Opção A: Caddy Server (Recomendada — SSL Automático)

O **Caddy** gerencia certificados Let's Encrypt de forma 100% automática e sem necessidade de scripts cron.

Instale o Caddy (`sudo apt install caddy`) e configure `/etc/caddy/Caddyfile`:

```caddy
api.seudominio.com.br {
    encode zstd gzip

    # Redireciona para a API local em Go
    reverse_proxy 127.0.0.1:8000 {
        header_up X-Real-IP {remote_host}
        header_up X-Forwarded-For {remote_host}
        header_up X-Forwarded-Proto {scheme}
    }

    # Cabeçalhos de segurança adicionais
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains; preload"
        X-Robots-Tag "noindex, nofollow"
    }

    log {
        output file /var/log/caddy/api_access.log {
            roll_size 50mb
            roll_keep 10
        }
    }
}
```

Recarregue com:
```bash
sudo systemctl reload caddy
```

---

### 5.2. Opção B: Nginx + Certbot

Se preferir o **Nginx**, crie `/etc/nginx/sites-available/api-juridica.conf`:

```nginx
server {
    listen 80;
    server_name api.seudominio.com.br;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.seudominio.com.br;

    ssl_certificate /etc/letsencrypt/live/api.seudominio.com.br/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.seudominio.com.br/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Otimização de buffers
    client_max_body_size 10M;
    keepalive_timeout 65;

    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts longos para acomodar queries profundas do DataJud
        proxy_connect_timeout 30s;
        proxy_send_timeout 90s;
        proxy_read_timeout 90s;
    }
}
```

Ative e gere o certificado com o Certbot:
```bash
sudo ln -s /etc/nginx/sites-available/api-juridica.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
sudo certbot --nginx -d api.seudominio.com.br
```

---

## 6. Persistência, WAL e Rotina de Backups do SQLite

O banco SQLite é configurado em modo **WAL (Write-Ahead Logging)** com timeout de espera de transação (`busy_timeout=15000ms`), garantindo leituras concorrentes ilimitadas sem bloqueio por escrita.

### 6.1. Script de Backup a Quente (`backup-sqlite.sh`)

Crie o utilitário de backup em `/opt/robo-juridico/scripts/backup-sqlite.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

ORIGEM="/opt/robo-juridico/dados/api.db"
DESTINO="/opt/robo-juridico/backups"
DATA=$(date +"%Y%m%d_%H%M%S")
ARQUIVO_DESTINO="${DESTINO}/api_${DATA}.db.gz"

mkdir -p "${DESTINO}"

# Realiza backup online consistente sem travar leituras/escritas
sqlite3 "${ORIGEM}" ".backup /tmp/backup_temporario.db"

# Compacta o arquivo de backup
gzip -c /tmp/backup_temporario.db > "${ARQUIVO_DESTINO}"
rm -f /tmp/backup_temporario.db

echo "[BACKUP SUCESSO] Backup gerado: ${ARQUIVO_DESTINO}"

# Remove backups com mais de 30 dias
find "${DESTINO}" -type f -name "api_*.db.gz" -mtime +30 -delete
```

Dê permissão de execução:
```bash
chmod +x /opt/robo-juridico/scripts/backup-sqlite.sh
```

### 6.2. Agendamento Cron Diário

Adicione ao cron com `sudo crontab -e`:

```cron
# Executa o backup às 03:00 da manhã diariamente
0 3 * * * /opt/robo-juridico/scripts/backup-sqlite.sh >> /var/log/backup-sqlite.log 2>&1
```

---

## 7. Pipeline de CI/CD com GitHub Actions

Crie o arquivo `.github/workflows/deploy.yml` no seu repositório para automatizar testes e deploy:

```yaml
name: CI/CD — Robô Jurídico API

on:
  push:
    branches: [ "main" ]

jobs:
  test:
    name: Testes Automatizados
    runs-on: ubuntu-latest
    steps:
      - name: Checkout do Código
        uses: actions/checkout@v4

      - name: Configurar Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache-dependency-path: '4-api-robo-juridico-php-go/api-juridica-go/go.sum'

      - name: Executar Testes
        run: |
          cd 4-api-robo-juridico-php-go/api-juridica-go
          go test -v -race ./...

  build-and-deploy:
    name: Build Docker & Deploy em Produção
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Login no GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build e Push da Imagem Docker
        uses: docker/build-push-action@v5
        with:
          context: ./4-api-robo-juridico-php-go/api-juridica-go
          file: ./4-api-robo-juridico-php-go/api-juridica-go/Dockerfile
          push: true
          tags: ghcr.io/${{ github.repository }}/api:latest

      - name: Deploy via SSH no Servidor
        uses: appleboy/ssh-action@v1.0.3
        with:
          host: ${{ secrets.PROD_SERVER_HOST }}
          username: ${{ secrets.PROD_SERVER_USER }}
          key: ${{ secrets.PROD_SERVER_SSH_KEY }}
          script: |
            cd /opt/robo-juridico
            docker compose pull api
            docker compose up -d --remove-orphans
            docker image prune -f
```

---

## 8. Monitoramento e Métricas

### 8.1. Checagem de Saúde (Liveness / Readiness)
Configure ferramentas de monitoramento externo (ex.: Uptime Kuma, Better Stack, Datadog) para monitorar:
- **URL:** `https://api.seudominio.com.br/saude`
- **Frequência esperada:** A cada 30 segundos
- **Código de Sucesso:** `200 OK`
- **Payload esperado:** `{"ok":true,"versao":"1.0.0"}`

### 8.2. Auditoria e Rastreabilidade
Cada requisição processada gera uma linha de auditoria estruturada:
```
[AUDITORIA] id=a81d2f70364e09cb cliente=empresa_x ip=200.189.12.3 GET /v1/processos/5053283-30.2024.8.13.0079 status=200 ms=342
```
Para investigar chamadas de suporte relatadas por clientes, basta pesquisar no journal pelo `id_requisicao` informado no erro.

---

## 9. Checklist de Segurança Pré-Go-Live

- [x] **Acesso Externo Restrito:** A porta `8000` está fechada no firewall (`ufw deny 8000`), sendo acessada apenas pelo proxy reverso local.
- [x] **SSL/TLS Ativado:** Certificado HTTPS ativo com nota A/A+ no SSL Labs.
- [x] **Isolamento não-root:** Binário Go executado como usuário `appuser` (Docker) ou `robojuridico` (Systemd).
- [x] **Chaves Secretas Fortes:** Todas as chaves em `API_CHAVES` e chaves geradas para clientes possuem no mínimo 32 caracteres criptograficamente aleatórios.
- [x] **Proteção de Força Bruta:** Rate limiter ativado barrando tentativas com chaves erradas por IP.
- [x] **Backup do SQLite Agendado:** Script de backup diário configurado no cron.
- [x] **Documentação Swagger Acessível:** Rota `/docs` disponível para os desenvolvedores parceiros.
