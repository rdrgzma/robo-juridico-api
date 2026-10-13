# Robô Jurídico — API de Processos em Go (`api-juridica-go`)

Implementação de alta performance e baixa latência da API REST do **Robô Jurídico**, escrita em Go moderno (1.24+).

Esta API consulta e consolida processos judiciais nas fontes oficiais do Poder Judiciário brasileiro (**DataJud** e **DJEN/ComunicaPJe**), mantendo **100% de compatibilidade** com o contrato da versão original em Python.

---

## 🚀 Características e Diferenciais

- **Desempenho Extremo:** Tempo de resposta na casa de sub-milissegundos para rotas internas e ~300ms–800ms para consultas federadas.
- **Concorrência Nativa:** Consulta paralela simultânea ao DataJud e DJEN utilizando goroutines com cancelamento via `context.Context`.
- **Zero CGO (`CGO_ENABLED=0`):** Utiliza driver SQLite puro em Go (`modernc.org/sqlite`), permitindo compilação de binários estáticos enxutos para containers Scratch ou Alpine.
- **Cache Anti-Thundering Herd:** TTL em memória com sincronização via `golang.org/x/sync/singleflight` para evitar requisições duplicadas concorrentes para o mesmo número de processo.
- **Proxy Pool Residencial:** Rotação e disparo concorrente (*race*) em múltiplos proxies HTTP/SOCKS5 para contornar bloqueios do DJEN.
- **Proteção & Cotas:** Autenticação por SHA-256 (`X-API-Key`), rate limiting deslizante em memória e reserva atômica de cota mensal com estorno em caso de falha $\ge 400$.

---

## 📁 Estrutura do Projeto

```
api-juridica-go/
├── cmd/
│   └── api/
│       └── main.go                # Ponto de entrada, injeção de dependências e graceful shutdown
├── configs/
│   └── config.go                  # Carregamento e validação de variáveis de ambiente
├── internal/
│   ├── auth/
│   │   └── auth.go                # Autenticação SHA-256, rate limiting e proteção contra força bruta
│   ├── cache/
│   │   └── cache.go               # Cache TTL em memória com singleflight
│   ├── cnj/
│   │   ├── cnj.go                 # Validador módulo 97 com big.Int, formatação e metadados
│   │   └── cnj_test.go            # Testes unitários do validador CNJ
│   ├── domain/
│   │   ├── erros.go               # Erros canônicos da API
│   │   └── modelos.go             # Structs de domínio e envelopes JSON
│   ├── fontes/
│   │   ├── comunica/
│   │   │   ├── client.go          # Cliente HTTP do DJEN / ComunicaPJe
│   │   │   ├── sanitizer.go       # Sanitização de HTML (divBody) e classificação de atos
│   │   │   └── sanitizer_test.go  # Testes de sanitização e regex
│   │   └── datajud/
│   │       ├── client.go          # Cliente Elasticsearch da API Pública do CNJ
│   │       ├── parser.go          # Parser de movimentos, deduplicação MD5 e moedas
│   │       └── parser_test.go     # Testes do parser e formatação monetária
│   ├── network/
│   │   ├── httpclient.go          # Cliente HTTP com pool de conexões e rotação de User-Agents
│   │   └── proxy_pool.go          # Pool de proxies com disparo concorrente
│   ├── service/
│   │   └── orchestrator.go        # Orquestrador de fontes, unificação e ordenação cronológica
│   ├── storage/
│   │   ├── database.go            # Interface de persistência e cálculo de períodos (BRT)
│   │   └── sqlite.go              # Driver SQLite puro (WAL, reservas atômicas de cotas)
│   └── transport/
│       └── http/
│           ├── handlers.go        # Handlers HTTP de todas as rotas
│           ├── handlers_test.go   # Testes automatizados da camada HTTP
│           ├── middleware.go      # Middlewares de auditoria, Request-ID, segurança e cotas
│           ├── response.go        # Padronização de respostas JSON e envelopes de erro
│           └── router.go          # Roteador Chi v5 com rotas públicas e autenticadas
├── Dockerfile                     # Multi-stage build otimizado (Alpine 3.21)
├── docker-compose.yml             # Orquestração local do container
└── go.mod                         # Definição do módulo Go
```

---

## 📡 Endpoints da API

| Método | Rota | Autenticação | Custo de Cota | Descrição |
|---|---|---|---|---|
| `GET` | `/saude` | Não | Não | Liveness e readiness check |
| `GET` | `/v1/cnj/{numero}` | Sim (`X-API-Key`) | Não | Valida e decodifica o número CNJ sem consultar fontes externas |
| `GET` | `/v1/processos/{numero}/existe` | Sim (`X-API-Key`) | Não | Checagem rápida de existência no DataJud ou DJEN |
| `GET` | `/v1/processos/{numero}` | Sim (`X-API-Key`) | **Sim (1)** | Consulta completa consolidada (DataJud + DJEN) ou `modo=rapido` |
| `GET` | `/v1/oab/{uf}/{numero}/processos` | Sim (`X-API-Key`) | **Sim (1)** | Processos com publicações da OAB no período informado |
| `GET` | `/v1/oab/{uf}/{numero}/publicacoes` | Sim (`X-API-Key`) | **Sim (1)** | Diário oficial na íntegra para controle de prazos |
| `GET` | `/v1/conta` | Sim (`X-API-Key`) | Não | Consulta plano, cota mensal usada/restante e data de renovação |

---

## ⚙️ Variáveis de Ambiente

| Variável | Padrão | Descrição |
|---|---|---|
| `HOST` | `0.0.0.0` | Endereço de escuta |
| `PORT` | `8000` | Porta HTTP |
| `API_BANCO` | `dados/api.db` | Caminho do arquivo SQLite compartilhado com o Admin |
| `API_CHAVES` | `""` | Chaves internas no formato `cliente:chave_secreta_min_32_chars` |
| `API_LIMITE_POR_MINUTO` | `60` | Taxa máxima padrão por chave |
| `DATAJUD_API_KEY` | *(chave pública CNJ)* | API Key da API Pública do CNJ |
| `DATAJUD_TIMEOUT` | `90` | Timeout em segundos para o DataJud |
| `COMUNICA_TIMEOUT` | `12` | Timeout em segundos para o DJEN |
| `CONSULTA_PROXIES` | `""` | Lista separada por vírgula de proxies HTTP |
| `CONSULTA_GATEWAY_HOST` | `""` | Host do gateway de proxies residenciais |
| `CONSULTA_GATEWAY_LOGIN` | `""` | Usuário do gateway |
| `CONSULTA_GATEWAY_SENHA` | `""` | Senha do gateway |
| `CONSULTA_GATEWAY_PORTAS`| `0` | Quantidade de portas a expandir no gateway |

---

## 🛠️ Como Executar

### 1. Localmente (Go 1.22+)

```bash
# Rodar os testes automatizados
go test -v ./...

# Iniciar o servidor
go run ./cmd/api
```

### 2. Com Docker Compose

```bash
docker compose up -d --build
```

### 3. Verificar Saúde

```bash
curl http://localhost:8000/saude
```
Retorno:
```json
{
  "ok": true,
  "versao": "1.0.0"
}
```
