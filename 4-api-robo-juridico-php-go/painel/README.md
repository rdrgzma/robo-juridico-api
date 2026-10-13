# Painéis Web — Robô Jurídico (Laravel 13 + Filament v5)

Sistema web corporativo construído em **Laravel 13** e **Filament v5** sob arquitetura **Multi-Panel** nativa, contendo o **Painel do Administrador** e o **Portal do Desenvolvedor / Cliente**, totalmente integrado ao banco de dados SQLite WAL (`dados/api.db`) compartilhado com a API em Go.

---

## 🏛️ Arquitetura Multi-Panel

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            APLICAÇÃO LARAVEL 13                             │
├──────────────────────────────────────┬──────────────────────────────────────┤
│    PAINEL ADMINISTRATIVO (/admin)    │       PORTAL DO CLIENTE (/portal)    │
│  - Guard: 'web'                      │  - Guard: 'cliente'                  │
│  - Model: Admin                      │  - Model: Cliente                    │
│  - Gestão global de clientes e MRR   │  - Autoatendimento do desenvolvedor  │
│  - Visão irrestrita do sistema       │  - Isolamento estrito por cliente_id │
└──────────────────────────────────────┴──────────────────────────────────────┘
                                       │
                                       ▼
                     BANCO DE DADOS COMPARTILHADO COM GO
                              (dados/api.db)
```

---

## 👥 1. Painel do Administrador (`/admin`)

Voltado exclusivamente para a equipe interna de gestão da plataforma:
- **Gestão de Clientes:** Cadastro de empresas, definição de cotas mensais, limites por minuto, valores contratuais e liberação de senha para o portal.
- **Gestão de Chaves de API:** Emissão, inspeção de prefixos e revogação imediata de tokens.
- **Telemetria Global:** Histórico completo de todas as requisições atendidas pela API em Go.
- **Trilha de Auditoria:** Registro de ações administrativas (criação de contas, alterações de plano, revogações).
- **Widgets de Dashboard:** Faturamento MRR, volume de consultas globais, gráfico de 14 dias e bloqueios por cota (HTTP 402).

**Acesso:** `http://localhost:8080/admin`
- **E-mail:** `admin@empresa.com.br`
- **Senha:** `admin123`

---

## 💻 2. Portal do Desenvolvedor / Cliente (`/portal`)

Ambiente de autoatendimento para o cliente/desenvolvedor contratante, com **isolamento estrito de dados**:
- **Minhas Chaves de API (`MinhaChaveResource`):**
  - Geração autônoma de tokens no padrão `rj_live_<random_bytes>`.
  - Exibição do token em texto puro uma única vez com persistência segura do SHA-256.
  - Revogação instantânea de chaves antigas ou comprometidas.
- **Meu Consumo & Quotas (`MinhaCotaWidget` & `MeuConsumoChartWidget`):**
  - Barra de progresso da cota mensal (ex: *1.250 / 5.000 consultas utilizadas*).
  - Alerta visual em caso de consumo elevado (>80%) ou franquia esgotada (100%).
  - Gráfico diário de requisições próprias (2xx Sucesso vs 4xx/5xx Erros).
- **Histórico de Requisições (`MeuUsoResource`):**
  - Logs detalhados apenas das requisições disparadas pelas chaves do cliente (endpoint, status code, latência em ms, IP).
- **Guia Rápido da API & Swagger (`DocumentacaoApi`):**
  - Exemplos prontos em cURL, Python e Node.js.
  - Link direto para o Swagger UI da API Go (`http://localhost:8000/docs`).
- **Meu Plano & Segurança (`MeuPlano`):**
  - Visualização de detalhes contratuais e alteração de senha de acesso.

**Acesso:** `http://localhost:8080/portal`
- **E-mail:** `cliente@empresa.com.br`
- **Senha:** `cliente123`

---

## 🛠️ Comandos de Linha de Comando (CLI Artisan)

O sistema inclui comandos para criar e gerenciar acessos diretamente pelo terminal:

```bash
# Criar ou redefinir Administrador
php artisan admin:criar seu@email.com --nome="Administrador" --senha="suasenha"

# Criar ou atualizar Cliente com acesso ao Portal
php artisan cliente:criar cliente@empresa.com.br --nome="Empresa XPTO" --empresa="XPTO Ltda" --plano=pro --senha="senha-portal"
```

---

## 🧪 Suíte de Testes Automatizados

A suíte cobre autenticação, isolamento multi-tenant e geração segura de chaves:

```bash
php artisan test
```

- `tests/Feature/AdminAuthTest.php`: Autenticação e restrição de acesso ao painel admin.
- `tests/Feature/ClientePortalAuthTest.php`: Login com guard `cliente` e bloqueio cruzado de painéis.
- `tests/Feature/ClientePortalIsolationTest.php`: Isolamento estrito de chaves e logs entre clientes.
- `tests/Feature/ClienteChaveSelfServiceTest.php`: Autoemissão e revogação de chaves pelo desenvolvedor.
- `tests/Feature/ChaveGeracaoTest.php`: Compatibilidade do SHA-256 entre PHP e Go.
- `tests/Feature/ClienteResourceTest.php`: CRUD administrativo de clientes.
- `tests/Unit/ConsumoCalculoTest.php`: Cálculo exato de consumo no fuso de Brasília (`America/Sao_Paulo`).
