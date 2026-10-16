# Plano Diretor de Desenvolvimento: API Jurídica em Go (`api-juridica-go`)

> **Projeto:** Robô Jurídico — Core Engine de Alta Performance em Go  
> **Localização de Destino:** `4-api-robo-juridico-php-go/api-juridica-go`  
> **Status:** 100% Implementado, Integrado e Validado por Testes Automatizados  
> **Versão do Documento:** 2.0.0  
> **Data:** Outubro de 2026  

---

## 1. Escopo e Diretrizes do Projeto

Este documento consolida a especificação de engenharia, arquitetura de pacotes, modelos de dados, fluxo de concorrência e suíte de testes do *Core Engine* da **API Jurídica em Go (Golang)**.

### 1.1. Premissas Arquiteturais
1. **Foco no Core da API:** Este projeto em Go implementa a **API REST pública de alta performance** (`/v1/...`), integrada ao painel administrativo e portal do cliente em PHP (`painel/`).
2. **Compatibilidade Rigorosa de Contrato:** Todas as rotas, query params, schemas JSON de entrada e saída, códigos de status HTTP e cabeçalhos (`X-API-Key`, `X-Cota-Limite`, `X-Cota-Restante`, `X-Request-ID`) mantêm interoperabilidade transparente.
3. **Padrão Idiomático em Go:** Estruturação seguindo as convenções canônicas da comunidade Go (*Standard Go Project Layout*), com tipagem estrita, cancelamento por `context.Context` e zero dependência de CGO (`CGO_ENABLED=0`).

---

## 2. Especificação Completa dos Endpoints da API

A API em Go atende aos seguintes endpoints:

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                 CATÁLOGO DE ENDPOINTS                                  │
├────────────────────────────────────────┬─────────┬──────────────┬──────────────────────┤
│ Rota                                   │ Método  │ Cobra Cota?  │ Descrição Resumida   │
├────────────────────────────────────────┼─────────┼──────────────┼──────────────────────┤
│ /saude                                 │ GET     │ Não (Livre)  │ Healthcheck e versão │
│ /v1/cnj/{numero}                       │ GET     │ Não (0 cota) │ Decodificação de DV  │
│ /v1/processos/{numero}/existe          │ GET     │ Não (0 cota) │ Checagem preliminar  │
│ /v1/conta                              │ GET     │ Não (0 cota) │ Saldo do mês e plano │
│ /v1/processos/{numero}                 │ GET     │ Sim (1 cota) │ Processo unificado   │
│ /v1/processos/consultar                │ POST    │ Sim (1 cota) │ Consulta via JSON    │
│ /v1/oab/{uf}/{numero}/processos        │ GET     │ Sim (1 cota) │ Processos por OAB    │
│ /v1/oab/{uf}/{numero}/publicacoes      │ GET     │ Sim (1 cota) │ Diário DJEN por OAB  │
│ /v1/calculos/indices                   │ GET     │ Não (0 cota) │ 27 índices oficiais  │
│ /v1/calculos/indices/{codigo}/serie    │ GET     │ Não (0 cota) │ Série histórica      │
│ /v1/calculos/judicial                  │ POST    │ Não (0 cota) │ Liquidação judicial  │
│ /v1/calculos/financiamento             │ POST    │ Não (0 cota) │ Revisão Price/Gauss  │
│ /v1/calculos/importar-planilha         │ POST    │ Não (0 cota) │ Importador .xlsx/csv │
│ /v1/calculos/exportar-planilha         │ POST    │ Não (0 cota) │ Exportador .xlsx/csv │
│ /v1/calculos/coletar                   │ POST    │ Não (0 cota) │ Coleta Bacen/IBGE    │
│ /v1/calculos/coletar/status            │ GET     │ Não (0 cota) │ Status do coletor    │
│ /v1/prazos/calcular                    │ POST    │ Não (0 cota) │ Prazos CPC e CLT     │
│ /v1/prazos/feriados                    │ GET     │ Não (0 cota) │ Feriados forenses    │
│ /v1/webhooks                           │ GET/POST│ Não (0 cota) │ Gestão de Webhooks   │
│ /v1/webhooks/testar                    │ POST    │ Não (0 cota) │ Teste HMAC-SHA256    │
│ /docs e /openapi.json                  │ GET     │ Não (Livre)  │ Swagger interativo   │
└────────────────────────────────────────┴─────────┴──────────────┴──────────────────────┘
```

### Detalhamento das Rotas

#### 1. `GET /v1/processos/{numero}`
* **Headers:** `X-API-Key: rj_live_...`
* **Parâmetros de Rota:** `numero` (string com ou sem pontuação).
* **Query Params:**
  * `modo` (`completo` ou `rapido` — padrão `completo`).
  * `publicacoes` (`true` ou `false` — padrão `true`).
* **Lógica Interna em Go:**
  1. Valida DV CNJ via Módulo 97 em memória.
  2. Reserva 1 cota no banco/Redis atomicamente.
  3. Dispara goroutines concorrentes:
     * Goroutine A: Busca DataJud (Elasticsearch `/api_publica_{tribunal}/_search`).
     * Goroutine B: Busca DJEN (`/api/v1/comunicacao`).
  4. Junta as árvores, deduplica movimentos com hash MD5 e gera o JSON formatado.
  5. Se status < 400, consolida o débito da cota. Se $\ge 400$, realiza o estorno.

#### 2. `GET /v1/processos/{numero}/existe`
* **Lógica:** Executa chamada leve com timeout curto (10s) ao DataJud e ao DJEN para confirmar existência de processo recém-distribuído sem gastar cota.

#### 3. `GET /v1/cnj/{numero}`
* **Lógica:** Puro cálculo matemático de Módulo 97 e decomposição de strings. Zero chamadas de rede externas e custo zero.

#### 4. `GET /v1/oab/{uf}/{numero}/processos`
* **Query Params:** `dias` (int, default 30), `tribunal` (string opcional).
* **Lógica:** Consulta publicações no DJEN pelo número e UF da OAB, agrupa por número de processo e formata os resumos.

#### 5. `GET /v1/oab/{uf}/{numero}/publicacoes`
* **Query Params:** `dias` (int, default 7), `paginas` (int, default 1).
* **Lógica:** Retorna o histórico de comunicações do diário com paginação e texto já sanitizado (sem tags HTML).

#### 6. `GET /v1/conta`
* **Lógica:** Retorna dados da chave autenticada: nome do cliente, plano, limite mensal, usado no mês, saldo restante e data exata de renovação (dia 1º do próximo mês às 00:00 no fuso de Brasília).

---

## 3. Estrutura Proposta de Diretórios e Arquivos (Go Layout)

A estrutura será criada dentro de `4-api-robo-juridico-php-go/api-juridica-go`:

```
4-api-robo-juridico-php-go/api-juridica-go/
├── cmd/
│   └── api/
│       └── main.go                    # Ponto de entrada, injeção de dependências e graceful shutdown
│
├── internal/
│   ├── cnj/
│   │   ├── cnj.go                     # Validador de máscara, Módulo 97 e mapeamento de tribunais
│   │   └── cnj_test.go                # Testes unitários de validação e DV
│   │
│   ├── domain/
│   │   ├── modelos.go                 # Structs canônicas: Processo, Parte, Advogado, Movimentacao
│   │   └── erros.go                   # Tipos de erros customizados (ErroAPI, FonteIndisponivel)
│   │
│   ├── fontes/
│   │   ├── datajud/
│   │   │   ├── client.go              # Driver Elasticsearch do DataJud com retries
│   │   │   ├── parser.go              # Parser de JSON bruto para structs com deduplicação MD5
│   │   │   └── client_test.go         # Testes com payload gravado do DataJud
│   │   └── comunica/
│   │       ├── client.go              # Driver do DJEN / Diário da Justiça
│   │       ├── sanitizer.go           # Limpeza de HTML (divBody) e classificação de tipos
│   │       └── client_test.go         # Testes com payload gravado do DJEN
│   │
│   ├── network/
│   │   ├── httpclient.go              # Cliente HTTP com suporte a impersonate TLS (Chrome 124)
│   │   └── proxy_pool.go              # Gestão de pool de proxies residenciais e corrida concorrente
│   │
│   ├── cache/
│   │   ├── cache.go                   # Cache TTL in-memory thread-safe
│   │   ├── singleflight.go            # Prevenção de Thundering Herd usando golang.org/x/sync/singleflight
│   │   └── cache_test.go              # Testes concorrentes de colisão de chaves
│   │
│   ├── storage/
│   │   ├── database.go                # Interface de banco de dados (SQLite WAL e PostgreSQL)
│   │   ├── sqlite.go                  # Implementação SQLite compartilhada com o painel PHP
│   │   └── quota.go                   # Lógica atômica de reserva, débito e estorno de cota
│   │
│   ├── service/
│   │   ├── orchestrator.go            # Motor de fusão paralela DataJud + DJEN
│   │   └── orchestrator_test.go       # Testes da lógica de mesclagem e fallbacks
│   │
│   └── transport/
│       └── http/
│           ├── router.go              # Configuração de rotas (Chi ou Gin)
│           ├── handlers.go            # Handlers REST para /v1/processos, /v1/oab, etc.
│           ├── middleware_auth.go     # Validação de X-API-Key, rate-limit por IP/cliente
│           ├── middleware_audit.go    # Cálculo de latência ms, ID de requisição e headers de cota
│           └── response.go            # Formatador padronizado de respostas e erros JSON
│
├── configs/
│   └── config.go                      # Leitura de variáveis de ambiente (.env)
│
├── testdata/                          # Fixtures de teste JSON (respostas reais gravadas do CNJ)
│   ├── datajud_processo.json
│   └── djen_publicacao.json
│
├── Dockerfile                         # Build multi-stage gerando binário estático Scratch/Alpine
├── .dockerignore
├── go.mod
└── go.sum
```

---

## 4. Stack Tecnológica e Bibliotecas Selecionadas

| Função Técnica | Biblioteca Escolhida | Justificativa |
|---|---|---|
| **Roteador HTTP** | `github.com/go-chi/chi/v5` | Ultraleve, 100% idiomático, compatível com `net/http` e zero alocações desnecessárias. |
| **Evasão TLS / Anti-WAF** | `github.com/bogdanfinn/tls-client` | A melhor biblioteca open-source do mundo para replicar o fingerprint JA3/JA4 do Chrome 124 e contornar bloqueios do gov.br. |
| **Concorrência & Deduplicação** | `golang.org/x/sync/singleflight` | Pacote oficial de Go para garantir que chamadas simultâneas ao mesmo processo resultem em uma única requisição ao CNJ. |
| **Banco de Dados Local** | `modernc.org/sqlite` | Implementação Pure Go de SQLite (não exige compilador C / CGO), garantindo compilação estática simples para Docker. |
| **Rate Limiter** | `golang.org/x/time/rate` | Token bucket oficial de alta performance para controle de limite por minuto por cliente e IP. |

---

## 5. Estratégia de Testes Automatizados

Seguindo o padrão de alta confiabilidade estabelecido na versão Python, a suite em Go será desenhada para **rodar completamente offline sem tocar na rede**:

1. **Testes Unitários de Domínio:**
   * Validação de dígito verificador CNJ com tabela de casos de teste válidos e inválidos (*Table-Driven Tests*).
   * Sanitização de texto: remoção correta de tags HTML e preservação do teor da intimação.
2. **Testes de Concorrência e Singleflight:**
   * Disparo de 50 goroutines simultâneas solicitando o mesmo número para validar se apenas 1 chamada ao mock é executada e todas as 50 recebem a resposta idêntica.
3. **Testes de Integração com Mocks HTTP:**
   * Utilização de `net/http/httptest` simulando o servidor do CNJ retornando payloads gravados em `testdata/`.
   * Simulação de falhas 429, 502 e 503 para validar as regras de retry exponencial e estorno de cota.
4. **Testes de Reserva e Estorno de Cotas:**
   * Validação de que erros 404 e 422 não reduzem o saldo mensal do cliente na tabela `uso`.

---

## 6. Fases de Execução da Implementação

Quando você autorizar o início do código, o desenvolvimento seguirá a ordem cronológica abaixo:

```
[FASE 1: Fundação & Modelos]
  ├── Inicialização de go.mod e setup de dependências
  ├── Implementação de internal/cnj (parser e módulo 97) + cnj_test.go
  └── Implementação de internal/domain/modelos.go e erros.go

[FASE 2: Camada de Rede & Drivers do CNJ]
  ├── Implementação do cliente HTTP resiliente (tls-client e proxy pool)
  ├── Driver do DataJud (Elasticsearch) com parser de movimentos MD5
  └── Driver do DJEN (Comunica) com sanitização de HTML

[FASE 3: Cache, Concorrência & Orquestração]
  ├── Implementação do CacheTTL in-memory com sync.singleflight
  └── Implementação do Orquestrador (fusão paralela e modo rápido)

[FASE 4: Persistência & Autenticação]
  ├── Conexão SQLite (compatível com a base dados/api.db)
  └── Middleware de autenticação X-API-Key, controle de cotas e estorno

[FASE 5: Rotas REST, Handlers & Docker]
  ├── Roteador Chi, endpoints /v1/... e respostas formatadas
  ├── Dockerfile multi-stage com compilação estática
  └── Execução completa da suite de testes automatizados com "go test -v ./..."
```
