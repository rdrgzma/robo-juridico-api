# Painéis Web & Calculadora Judicial — Robô Jurídico (Laravel 11 + Filament v3)

Sistema web corporativo construído em **Laravel 11** e **Filament v3** sob arquitetura **Multi-Panel** nativa, contendo o **Painel do Administrador**, o **Portal do Desenvolvedor / Cliente** e a **Calculadora Judicial Pública com Captação de Leads (Lead Magnet)**, totalmente integrado ao banco de dados SQLite WAL (`dados/api.db`) compartilhado com a API em Go.

---

## 🏛️ Arquitetura do Sistema

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                   APLICAÇÃO WEB EM LARAVEL                                  │
├──────────────────────────────────────┬──────────────────────────────────────┬───────────────┤
│    PAINEL ADMINISTRATIVO (/admin)    │       PORTAL DO CLIENTE (/portal)    │  LEAD MAGNET  │
│  - Guard: 'web'                      │  - Guard: 'cliente'                  │  (GET /)      │
│  - Model: Admin                      │  - Model: Cliente                    │  - Calculadora│
│  - Gestão global de clientes e MRR   │  - Autoatendimento do desenvolvedor  │    Pública    │
│  - Calculadora Judicial Integrada    │  - Emissão e revogação de tokens     │  - Importador │
│  - Importador de Planilhas Excel     │  - Calculadora Judicial de Débitos   │  - Laudo A4   │
│  - Emissão de Laudos Certificados    │  - Isolamento estrito por cliente_id │  - QR Code    │
└──────────────────────────────────────┴──────────────────────────────────────┴───────────────┘
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
- **Calculadora de Débitos Judiciais (`/admin/calculadora-judicial`):**
  - Liquidação contábil com paridade com o DrCalc e aplicação da **Lei 14.905/2024**.
  - Ação de cabeçalho **"Importar Planilha / Colar"** para subir arquivos `.xlsx`, `.csv` ou colar linhas do Excel (Ctrl+V).
  - Emissão de Laudo Pericial Certificado com QR Code de integridade.
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
- **Calculadora Judicial para Clientes (`/portal/calculadora-judicial`):**
  - Acesso irrestrito ao liquidor judicial de alta precisão.
  - Importação de planilhas e emissão de laudos oficiais para processos do escritório.
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

## 🧮 3. Calculadora Judicial Pública & Lead Magnet (`GET /` e `GET /calculadora`)

Página de conversão de alto impacto e utilidade real para captação de advogados, peritos e escritórios:
- **Interface Pública:** Construída com Tailwind CSS, sem necessidade de login.
- **Liquidação em Tempo Real:** Chamadas assíncronas (AJAX) para o motor em Go.
- **26+ Tabelas Oficiais:** TJSP, TJRJ, TJMG, TJRS, TJPR, TRF, TST, IPCA, INPC, Selic e Taxa Legal Bacen.
- **Importador de Planilha:** Botões para importar arquivos `.xlsx`, `.csv` ou colar linhas diretamente do Excel.
- **Modal Lead Magnet:** Ao solicitar o download do Laudo em PDF, solicita os dados do profissional (Nome, E-mail, WhatsApp, OAB/Empresa) e registra na tabela `calculos_salvos`.
- **Laudo Pericial Certificado (`/laudo/{uuid}/imprimir`):** Layout formal A4 para peticionamento, com resumo metodológico, planilha analítica, totais consolidados e **QR Code de autenticidade**.
- **Selo Digital de Autenticidade (`/validar/{uuid}`):** Verificação criptográfica via hash SHA-256 acessível por qualquer celular ao ler o QR Code do laudo.

---

## 🛠️ Comandos de Linha de Comando (CLI Artisan)

O sistema inclui comandos para criar e gerenciar acessos diretamente pelo terminal:

```bash
# Executar migrações do banco de dados (incluindo calculos_salvos)
php artisan migrate

# Criar ou redefinir Administrador
php artisan admin:criar seu@email.com --nome="Administrador" --senha="suasenha"

# Criar ou atualizar Cliente com acesso ao Portal
php artisan cliente:criar cliente@empresa.com.br --nome="Empresa XPTO" --empresa="XPTO Ltda" --plano=pro --senha="senha-portal"
```

---

## 🧪 Suíte de Testes Automatizados

A suíte cobre autenticação, isolamento multi-tenant, motor de cálculos judiciais e certificação de laudos:

```bash
vendor/bin/phpunit
```

- `tests/Feature/CalculoJudicialFeatureTest.php`: Carregamento da página pública, importação, emissão de laudo com lead, integridade de hash SHA-256 e validação de autenticidade.
- `tests/Feature/AdminAuthTest.php`: Autenticação e restrição de acesso ao painel admin.
- `tests/Feature/ClientePortalAuthTest.php`: Login com guard `cliente` e bloqueio cruzado de painéis.
- `tests/Feature/ClientePortalIsolationTest.php`: Isolamento estrito de chaves e logs entre clientes.
- `tests/Feature/ClienteChaveSelfServiceTest.php`: Autoemissão e revogação de chaves pelo desenvolvedor.
- `tests/Feature/ChaveGeracaoTest.php`: Compatibilidade do SHA-256 entre PHP e Go.
- `tests/Feature/ClienteResourceTest.php`: CRUD administrativo de clientes.
- `tests/Unit/ConsumoCalculoTest.php`: Cálculo exato de consumo no fuso de Brasília (`America/Sao_Paulo`).
