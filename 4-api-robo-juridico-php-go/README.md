# Stack Integrada: API Robô Jurídico (Go) + Painel Administrativo (PHP/Filament)

Solução corporativa completa para consulta automatizada de processos jurídicos, combinando a alta velocidade de processamento de um microsserviço nativo em **Go** com a produtividade e elegância do **Laravel 13 + Filament v5** para gestão administrativa.

---

## 🏛️ Visão Geral da Arquitetura

```
                     ┌──────────────────────────────────────────────┐
                     │                 Clientes /                   │
                     │          Sistemas Consumidores               │
                     └──────────────────────┬───────────────────────┘
                                            │ Requisições HTTP (Bearer Token)
                                            ▼
                     ┌──────────────────────────────────────────────┐
                     │           api-juridica-go (Porta 8000)       │
                     │   - Go 1.24 HTTP nativo (Sem frameworks)     │
                     │   - Validação de Bearer Token & SHA-256      │
                     │   - Verificação e bloqueio de cotas (402)    │
                     │   - Consulta de processos TJ/TRF/TST         │
                     │   - Telemetria de latência e consumo         │
                     │   - Documentação Swagger interativa (/docs)  │
                     └──────────────────────┬───────────────────────┘
                                            │
                                            │ Leitura / Escrita Concorrente
                                            │ (SQLite WAL + busy_timeout=15s)
                                            ▼
                             ┌──────────────────────────────┐
                             │    /app/dados/api.db (WAL)   │
                             └──────────────┬───────────────┘
                                            ▲
                                            │ Leitura / Escrita Administrativa
                                            │
                     ┌──────────────────────┴───────────────────────┐
                     │            painel-admin (Porta 8080)         │
                     │   - PHP 8.4 + Laravel 13 + Filament v5       │
                     │   - Gestão de Clientes, Contratos e Planos   │
                     │   - Emissão e revogação de chaves (rj_live_) │
                     │   - Dashboard KPI & Gráficos de consumo      │
                     │   - Histórico de telemetria e auditoria      │
                     └──────────────────────────────────────────────┘
```

---

## 📁 Estrutura de Diretórios

```
4-api-robo-juridico-php-go/
├── api-juridica-go/            # Microsserviço de Alta Performance em Go
│   ├── cmd/api/main.go         # Ponto de entrada do servidor Go
│   ├── internal/               # Camadas de domínio, dados e regras de negócio
│   ├── docs/                   # Especificações OpenAPI 3.1.0 (YAML/JSON)
│   ├── Dockerfile              # Imagem estática Alpine (~20MB)
│   └── README.md               # Documentação detalhada da API Go
│
├── painel/                     # Painel Administrativo em Laravel 13 + Filament v5
│   ├── app/Filament/           # Resources, Widgets e Páginas do Filament
│   ├── app/Models/             # Modelos Eloquent integrados ao schema SQLite
│   ├── app/Services/           # Serviços de chaves e auditoria
│   ├── database/migrations/    # Migrações compatíveis com o banco compartilhado
│   ├── tests/                  # Suíte de testes automatizados (Feature & Unit)
│   ├── Dockerfile              # Container PHP 8.4 Alpine com extensões
│   └── README.md               # Documentação detalhada do painel PHP
│
├── docker-compose.yml          # Orquestrador dos dois serviços com volume compartilhado
└── README.md                   # Esta documentação
```

---

## ⚡ Como Rodar com Docker Compose

Suba toda a infraestrutura com apenas um comando:

```bash
docker compose up -d --build
```

### Serviços Disponíveis:
- **API Jurídica (Go):** [http://localhost:8000](http://localhost:8000)
- **Documentação Swagger UI:** [http://localhost:8000/docs](http://localhost:8000/docs)
- **Especificação OpenAPI JSON:** [http://localhost:8000/openapi.json](http://localhost:8000/openapi.json)
- **Healthcheck da API:** [http://localhost:8000/saude](http://localhost:8000/saude)
- **Painel Administrativo (Filament):** [http://localhost:8080/admin](http://localhost:8080/admin)

### Credenciais Padrão do Painel Admin:
- **E-mail:** `admin@empresa.com.br`
- **Senha:** `admin123`

---

## 🔑 Fluxo de Uso Rápido

1. Acesse o **Painel Administrativo** em `http://localhost:8080/admin` e efetue login.
2. Crie um novo cliente na aba **Clientes**.
3. Na aba **Chaves de API** (ou no detalhe do cliente), gere uma nova chave de acesso. Copie o token `rj_live_...` gerado.
4. Faça uma chamada à **API Jurídica em Go**:
   ```bash
   curl -X POST http://localhost:8000/api/v1/processos/consultar \
     -H "Authorization: Bearer rj_live_SEU_TOKEN_AQUI" \
     -H "Content-Type: application/json" \
     -d '{"numero_processo": "0001234-56.2023.8.26.0100"}'
   ```
5. Atualize o Painel Administrativo:
   - O gráfico diário registrará o volume de chamadas imediatamente.
   - O widget de consumo indicará o decremento da cota mensal do cliente.
   - A tabela de **Histórico de Uso** registrará a rota, latência em milissegundos e IP solicitante.
