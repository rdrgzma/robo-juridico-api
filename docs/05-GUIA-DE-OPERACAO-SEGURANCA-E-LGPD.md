# Guia de Operações, Infraestrutura, Segurança e LGPD: Robô Jurídico

> **Identificador:** OPS-SEC-RJ-001  
> **Versão:** 1.0.0  
> **Ambiente de Destino:** Produção / Cloud / On-Premise  
> **Status:** Ativo  
> **Data:** Outubro de 2026  

---

## 1. Arquitetura de Implantação em Produção

Para colocar a API em produção de forma profissional e segura para consumo por clientes e sistemas externos, o servidor Uvicorn não deve ser exposto diretamente na internet sem uma camada de terminação TLS (HTTPS) e proteção de borda.

### 1.1. Topologia Recomendada

```
[Internet: Clientes / Apps / ERPs]
               │ HTTPS (Porta 443)
               ▼
┌──────────────────────────────────────────────┐
│  Proxy Reverso (Nginx / HAProxy / Traefik)    │
│  • Certificado SSL/TLS (Let's Encrypt)       │
│  • Cabeçalho X-Forwarded-For                 │
│  • Bloqueio de IP / VPN para /admin          │
└──────────────────────────────────────────────┘
               │ HTTP Local (127.0.0.1:8000)
               ▼
┌──────────────────────────────────────────────┐
│  Uvicorn + FastAPI (Robô Jurídico)          │
│  • 1 Worker (Mantém cache e rate-limit)      │
│  • Validação de API Keys e Cotas             │
│  • SQLite WAL (dados/api.db)                 │
└──────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
 [DataJud / CNJ]           [DJEN / Comunica]
 (Elasticsearch)       (Se no exterior: via Proxy BR)
```

> [!IMPORTANT]
> **Por que rodar com 1 único worker Uvicorn?**  
> Na versão atual (v0.1.0), as estruturas `CacheTTL` e `LimiteDeTaxa` operam na memória do processo Python. Executar múltiplos workers Uvicorn (`--workers 4`) particionaria o cache e os contadores de requisições por minuto entre os processos. Um único worker assíncrono com Uvicorn é capaz de atender centenas de requisições por segundo, pois as chamadas externas pesadas são delegadas a pools de threads não bloqueantes.

---

## 2. Configuração de Proxy Reverso (Nginx)

Exemplo completo de configuração do Nginx atuando como proxy reverso seguro:

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

    # Encaminhamento para a API Robô Jurídico
    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts longos para acomodar consultas pesadas do DataJud
        proxy_connect_timeout 30s;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }

    # Proteção adicional do painel administrativo por IP (opcional ao Nginx)
    location /admin {
        # allow 187.100.20.0/24; # Exemplo: IP do escritório
        # deny all;

        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

No `.env` da aplicação, configure:
```env
API_PROXY_CONFIAVEL=127.0.0.1
API_COOKIE_SEGURO=1
```

---

## 3. Implantação com Docker e Dockerfile de Produção

### 3.1. `Dockerfile` Otimizado
```dockerfile
FROM python:3.12-slim

# Dependências de sistema para compilação eventual e curl-cffi
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libffi-dev \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Criação de usuário não-root por segurança
RUN useradd -m -u 1000 appuser

# Copia metadados de dependência primeiro para cache do Docker
COPY pyproject.toml README.md LICENSE ./
COPY consulta_processos ./consulta_processos

# Instala a aplicação com extensões de API e evasão de navegador
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -e ".[api,navegador]"

# Cria diretório de banco com permissão para o usuário da aplicação
RUN mkdir -p /app/dados && chown -R appuser:appuser /app

USER appuser

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8000/saude || exit 1

CMD ["consulta-processos", "api", "--host", "0.0.0.0", "--porta", "8000"]
```

### 3.2. `docker-compose.yml`
```yaml
version: '3.8'

services:
  robo-juridico:
    build: .
    container_name: robo-juridico-api
    restart: unless-stopped
    ports:
      - "127.0.0.1:8000:8000"
    volumes:
      - ./dados:/app/dados
      - ./.env:/app/.env:ro
    environment:
      - API_BANCO=/app/dados/api.db
      - API_COOKIE_SEGURO=1
      - API_PROXY_CONFIAVEL=127.0.0.1,172.16.0.0/12
    logging:
      driver: "json-file"
      options:
        max-size: "20m"
        max-file: "5"
```

---

## 4. Gestão e Configuração de Proxies Residenciais

### 4.1. Quando é Mandatório Usar Proxy?
1. **Servidor Hospedado no Exterior:** Provedores como AWS (us-east-1), DigitalOcean (NYC), Hetzner (Alemanha) ou Linode receberão **HTTP 403 Forbidden** na API do Comunica/DJEN. Nesses casos, o proxy residencial brasileiro é **obrigatório**.
2. **Consultas em Larga Escala (Milhares/hora):** Pools de datacenters comuns são barrados por rate limits no gateway do gov.br. O pool de IPs domésticos rotativos distribui o tráfego de maneira imperceptível.

### 4.2. Configuração no `.env`
```env
# Gateway residencial brasileiro que abre 1 IP por porta:
CONSULTA_GATEWAY_HOST=gw.provedor-residencial.com
CONSULTA_GATEWAY_LOGIN=usuario_res
CONSULTA_GATEWAY_SENHA=senha_res
CONSULTA_GATEWAY_PORTA_INICIAL=10000
CONSULTA_GATEWAY_PORTAS=50

# Concorrência: dispara 4 proxies em paralelo e fica com o mais rápido
CONSULTA_PROXIES_PARALELO=4
CONSULTA_PROXIES_MAX=16
```

---

## 5. Estratégia de Backup do SQLite WAL

Como o arquivo `dados/api.db` armazena o histórico financeiro, clientes, chaves e trilha de auditoria, ele deve receber backup diário automatizado. Nunca copie o arquivo diretamente enquanto a aplicação estiver em execução sem usar a API de backup online do SQLite.

### 5.1. Script de Backup Quente (`backup.sh`)
```bash
#!/usr/bin/env bash
set -e

DATA=$(date +%Y%m%d_%H%M%S)
PASTA_ORIGEM="/app/dados"
PASTA_DESTINO="/var/backups/robo_juridico"

mkdir -p "$PASTA_DESTINO"

# Executa backup atômico e consistente usando o utilitário sqlite3
sqlite3 "$PASTA_ORIGEM/api.db" ".backup '$PASTA_DESTINO/api_$DATA.db'"

# Compacta com gzip
gzip "$PASTA_DESTINO/api_$DATA.db"

# Remove backups com mais de 30 dias
find "$PASTA_DESTINO" -name "api_*.db.gz" -mtime +30 -delete

echo "Backup concluído com sucesso: api_$DATA.db.gz"
```

Adicione ao `crontab` do servidor:
```cron
0 3 * * * /usr/local/bin/backup.sh > /var/log/backup_api.log 2>&1
```

---

## 6. Governança, LGPD e Uso Ético dos Dados

### 6.1. Enquadramento Legal da LGPD (Lei 13.709/2018)
- **Natureza dos Dados:** Dados processuais judiciais (nomes das partes, advogados, órgãos julgadores e teores de despachos).
- **Base Legal:** **Art. 7º, § 4º da LGPD:** É dispensada a exigência do consentimento para os dados tornados manifestamente públicos pelo titular, resguardados os direitos do titular e os princípios da boa-fé e da finalidade.
- **Processos em Segredo de Justiça:** A API consome exclusivamente as fontes públicas do CNJ. Processos que tramitam sob segredo de justiça (Direito de Família, menores, etc.) **não são indexados nem fornecidos** pelas APIs públicas utilizadas.
- **Princípio da Necessidade e Segurança:**
  - O operador da API não deve reter dados pessoais indefinidamente além do estritamente necessário para fins de auditoria ou entrega do serviço contratado.
  - A tabela `uso` registra metadados de requisição (IP, rota, código HTTP, tempo gasto), garantindo rastreabilidade jurídica de quem consultou cada processo.

### 6.2. Boas Práticas de Convivência com o CNJ
1. **Não Quebrar Automações:** A aplicação consome os endpoints oficiais em conformidade com as diretrizes do CNJ, sem violar mecanismos de captcha.
2. **Utilizar o Cache:** Não desative o cache TTL. Consultas frequentes a um mesmo processo devem ser respondidas pela memória da API para aliviar a infraestrutura pública.
3. **Agendamento de Varreduras:** Varreduras massivas por OAB devem ser executadas preferencialmente fora dos horários de pico judiciário (após as 20h e antes das 08h).
4. **Identificação Responsável:** Configure a variável `CONSULTA_USER_AGENT_EXTRA` com o nome da sua aplicação e um e-mail de contato técnico.

---

## 7. Runbook de Troubleshooting (Solução de Problemas)

| Sintoma Observado | Causa Mais Provável | Procedimento de Resolução |
|---|---|---|
| **Erro no PowerShell: "A execução de scripts foi desabilitada"** | Política de segurança restritiva padrão do Windows. | Execute no PowerShell: `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned` e reative o ambiente `.venv`. |
| **Todas as chamadas retornam HTTP 401** | Cabeçalho incorreto ou chave revogada. | Certifique-se de enviar `X-API-Key: sua_chave`. No PowerShell, use sempre `curl.exe` (o comando nativo `curl` é alias e não repassa headers da mesma forma). |
| **Todas as consultas ao DJEN voltam 403** | Servidor hospedado em IP estrangeiro bloqueado pelo gov.br. | Configure um pool de proxies residenciais brasileiros ou utilize uma máquina hospedada em território nacional. |
| **Consulta ao DataJud demora mais de 60 segundos** | Horário de pico no Elasticsearch do CNJ. | Utilize `modo=rapido` na chamada para obter os dados do DJEN imediatamente; a segunda consulta sairá do cache em milissegundos. |
| **Administrador bloqueado na tela de login** | Mais de 5 tentativas erradas em 15 minutos. | Redefina a senha imediatamente pelo terminal com: `consulta-processos admin senha seu_email@dominio.com`. |
| **Erro `Address already in use` (Porta 8000)** | Outra instância da API ou serviço rodando na porta. | Suba a aplicação em outra porta: `consulta-processos api --porta 8010`. |
