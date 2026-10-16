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
- **Motor de Liquidação Judicial (Paridade DrCalc / PJe-Calc):** Precisão arbitrária com `decimal.Decimal`, conformidade plena com a **Lei 14.905/2024** ($\max(0, \text{Selic} - \text{IPCA})$), juros moratórios e compensatórios/remuneratórios com *pro rata die* diário, correção monetária *pro rata die* linear, imputação pelo **Art. 354 do Código Civil** e fase de execução (Art. 523 CPC 10%+10%).
- **Revisão de Contratos Bancários (`PRICE`, `SAC` e `GAUSS`):** Comparativo contratado vs. mercado com expurgo de anatocismo e apuração de repetição de indébito simples e em dobro (Art. 42 CDC).
- **Calculadora de Prazos Processuais Fatais (CPC/2015 e CLT):** Apuração de publicação no DJe, início de contagem e vencimento fatal em dias úteis ou corridos (Art. 219 e 224 CPC), recesso forense (Art. 220 CPC) e feriados móveis (Páscoa, Carnaval, Corpus Christi).
- **Monitoramento Push via Webhooks (`HMAC-SHA256`):** Cadastro e disparo assinado de notificações push para processos CNJ e inscrições OAB.
- **Worker Coletor Automático Bacen/IBGE:** Sincronização automática contínua de índices econômicos (SGS Bacen, SIDRA IBGE, janela de 60 meses) e 27 tabelas oficiais de Tribunais Estaduais, Federais e do Trabalho.
- **Importador e Exportador Nativo de Planilhas:** Parser e gerador nativos em Go para `.xlsx` (OpenXML) e `.csv` sem bibliotecas externas de C.

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
│   ├── calculator/
│   │   ├── collectors/
│   │   │   ├── bacen.go           # Coletor Bacen SGS (Selic, CDI, TR, Poupança, IGP-M, etc. - 60 meses)
│   │   │   ├── ibge.go            # Coletor IBGE SIDRA (IPCA, INPC, IPCA-15 - 60 meses)
│   │   │   ├── tribunais.go       # Sincronizador de 27 tabelas de tribunais (TJSP, TJRJ, TRF, TST)
│   │   │   └── worker.go          # Daemon de coleta automática diária (ticker 24h)
│   │   ├── correction.go          # Motor de correção monetária (fatores acumulados, pro rata e deflação)
│   │   ├── engine.go              # Orquestrador contábil da liquidação judicial (inclui Art. 354 CC)
│   │   ├── exporter.go            # Exportador nativo de memória de cálculo em .xlsx e .csv
│   │   ├── financing.go           # Motor de revisão bancária (PRICE, SAC, GAUSS e Art. 42 CDC)
│   │   ├── importer.go            # Importador nativo de planilhas Excel (.xlsx) e CSV
│   │   ├── interest.go            # Motor de juros moratórios e compensatórios (Lei 14.905, Pro Rata)
│   │   ├── models.go              # Modelos de domínio com precisão arbitrária decimal.Decimal
│   │   ├── penalties.go           # Multas moratórias, honorários e encargos do Art. 523 CPC
│   │   ├── repository.go          # Repositório thread-safe de índices e séries temporais
│   │   └── seed.go                # Carga inicial histórica (2023-2026) para todos os 27 índices
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
│   ├── prazos/
│   │   └── prazos.go              # Calculadora de prazos processuais (CPC Art. 224, CLT Art. 775 e feriados)
│   ├── service/
│   │   └── orchestrator.go        # Orquestrador de fontes, unificação e ordenação cronológica
│   ├── storage/
│   │   ├── database.go            # Interface de persistência e cálculo de períodos (BRT)
│   │   └── sqlite.go              # Driver SQLite puro (WAL, reservas atômicas de cotas)
│   ├── transport/
│   │   └── http/
│   │       ├── handlers.go        # Handlers HTTP de todas as rotas (processos, cálculos, prazos, webhooks)
│   │       ├── handlers_test.go   # Testes automatizados da camada HTTP
│   │       ├── middleware.go      # Middlewares de auditoria, Request-ID, segurança e cotas
│   │       ├── response.go        # Padronização de respostas JSON e envelopes de erro
│   │       └── router.go          # Roteador Chi v5 com rotas públicas e autenticadas
│   └── webhooks/
│       └── webhooks.go            # Gerenciador de webhooks push com assinatura HMAC-SHA256
├── Dockerfile                     # Multi-stage build otimizado (Alpine 3.21)
├── docker-compose.yml             # Orquestração local do container
└── go.mod                         # Definição do módulo Go
```

---

## 📡 Endpoints da API

### Módulo de Processos e Diários Oficiais (CNJ)

| Método | Rota | Autenticação | Custo de Cota | Descrição |
|---|---|---|---|---|
| `GET` | `/saude` | Não | Não | Liveness e readiness check |
| `GET` | `/v1/cnj/{numero}` | Sim (`X-API-Key`) | Não | Valida e decodifica o número CNJ sem consultar fontes externas |
| `GET` | `/v1/processos/{numero}/existe` | Sim (`X-API-Key`) | Não | Checagem rápida de existência no DataJud ou DJEN |
| `GET` | `/v1/processos/{numero}` | Sim (`X-API-Key`) | **Sim (1)** | Consulta completa consolidada (DataJud + DJEN) ou `modo=rapido` |
| `POST` | `/v1/processos/consultar` | Sim (`X-API-Key`) | **Sim (1)** | Consulta consolidada via payload JSON |
| `GET` | `/v1/oab/{uf}/{numero}/processos` | Sim (`X-API-Key`) | **Sim (1)** | Processos com publicações da OAB no período informado |
| `GET` | `/v1/oab/{uf}/{numero}/publicacoes` | Sim (`X-API-Key`) | **Sim (1)** | Diário oficial na íntegra para controle de prazos |
| `GET` | `/v1/conta` | Sim (`X-API-Key`) | Não | Consulta plano, cota mensal usada/restante e data de renovação |

### Módulo de Cálculos Judiciais, Revisão Bancária, Prazos e Webhooks

| Método | Rota | Autenticação | Custo de Cota | Descrição |
|---|---|---|---|---|
| `GET` | `/v1/calculos/indices` | Sim (`X-API-Key`) | Não | Lista todos os 27 indexadores econômicos e tabelas dos tribunais |
| `GET` | `/v1/calculos/indices/{codigo}/serie` | Sim (`X-API-Key`) | Não | Retorna a série histórica mensal completa do índice solicitado |
| `POST` | `/v1/calculos/judicial` | Sim (`X-API-Key`) | Não | Executa liquidação completa de débito (correção, juros, Art. 354 CC e Art. 523 CPC) |
| `POST` | `/v1/calculos/financiamento` | Sim (`X-API-Key`) | Não | Revisão de contratos bancários (`PRICE`, `SAC`, `GAUSS` e Art. 42 CDC) |
| `POST` | `/v1/calculos/importar-planilha` | Sim (`X-API-Key`) | Não | Extrai parcelas de arquivo Excel (`.xlsx`), CSV (`.csv`) ou linhas de texto |
| `POST` | `/v1/calculos/exportar-planilha` | Sim (`X-API-Key`) | Não | Exporta memória de cálculo liquidada em `.xlsx` ou `.csv` |
| `POST` | `/v1/calculos/coletar` | Sim (`X-API-Key`) | Não | Dispara imediatamente a rotina de coleta no Bacen (SGS) e IBGE (SIDRA) |
| `GET` | `/v1/calculos/coletar/status` | Sim (`X-API-Key`) | Não | Retorna o status e relatório da última coleta automática de índices |
| `POST` | `/v1/prazos/calcular` | Sim (`X-API-Key`) | Não | Calcula prazos processuais fatais (CPC/2015 Art. 224 e CLT Art. 775) |
| `GET` | `/v1/prazos/feriados` | Sim (`X-API-Key`) | Não | Lista feriados nacionais, móveis e forenses do ano |
| `GET` | `/v1/webhooks` | Sim (`X-API-Key`) | Não | Lista webhooks de monitoramento push cadastrados |
| `POST` | `/v1/webhooks` | Sim (`X-API-Key`) | Não | Registra novo webhook de monitoramento de processo ou OAB |
| `DELETE` | `/v1/webhooks/{id}` | Sim (`X-API-Key`) | Não | Remove um webhook cadastrado |
| `POST` | `/v1/webhooks/testar` | Sim (`X-API-Key`) | Não | Simula disparo de evento push com assinatura `HMAC-SHA256` |

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
