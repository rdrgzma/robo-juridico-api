# Stack Integrada: API Robô Jurídico (Go) + Painel Administrativo e Portal do Cliente (PHP/Filament)

Solução corporativa completa para consulta automatizada de processos jurídicos, combinando a alta velocidade de processamento de um microsserviço nativo em **Go** com a produtividade e elegância do **Laravel 13 + Filament v5** para gestão administrativa e autoatendimento do cliente.

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
                                            │
                     ┌──────────────────────┴───────────────────────┐
                     │       painel-admin (Porta 8080 - Multi-Panel)│
                     │   - PHP 8.4 + Laravel 13 + Filament v5       │
                     ├──────────────────────────────────────────────┤
                     │   A) PAINEL ADMINISTRATIVO (/admin)          │
                     │      - Gestão de clientes e planos           │
                     │      - Métricas de faturamento (MRR)         │
                     │      - Auditoria global do sistema           │
                     ├──────────────────────────────────────────────┤
                     │   B) PORTAL DO CLIENTE (/portal)             │
                     │      - Autoatendimento do desenvolvedor      │
                     │      - Emissão e revogação de chaves         │
                     │      - Gráfico de consumo e cota restante    │
                     │      - Logs de requisições do próprio cliente│
                     └──────────────────────────────────────────────┘
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
- **Painel do Administrador:** [http://localhost:8080/admin](http://localhost:8080/admin)
  - **Login:** `admin@empresa.com.br` | **Senha:** `admin123`
- **Portal do Desenvolvedor / Cliente:** [http://localhost:8080/portal](http://localhost:8080/portal)
  - **Login:** `cliente@empresa.com.br` | **Senha:** `cliente123`

---

## 🔑 Fluxo Completo de Autoatendimento do Cliente

1. O cliente faz login no **Portal do Desenvolvedor** em `http://localhost:8080/portal`.
2. No menu **Minhas Chaves de API**, clica em **Gerar Nova Chave de API** e copia seu token `rj_live_...`.
3. Executa chamadas à **API Jurídica (Go)** na porta `8000`:
   ```bash
   curl -X POST http://localhost:8000/api/v1/processos/consultar \
     -H "Authorization: Bearer rj_live_SEU_TOKEN_AQUI" \
     -H "Content-Type: application/json" \
     -d '{"numero_processo": "0001234-56.2023.8.26.0100"}'
   ```
4. No **Dashboard do Portal**, acompanha em tempo real:
   - Quantidade de requisições consumidas e saldo da cota restante.
   - Gráfico diário com requisições 2xx (sucesso) vs 4xx/5xx (erros).
   - Histórico detalhado no menu **Histórico de Consultas**.
5. Se uma chave for exposta ou precisar ser substituída, o próprio cliente clica em **Revogar** com efeito instantâneo.
