# Painel Administrativo — Robô Jurídico (Laravel 13 + Filament v5)

Painel administrativo corporativo para gestão de clientes, planos contratuais, chaves de API, monitoramento de cotas de consumo em tempo real e auditoria da API Robô Jurídico.

---

## 🏛️ Arquitetura e Decisões Técnicas

- **Framework Web:** Laravel 13 (PHP 8.2–8.4)
- **Admin Panel:** Filament v5 (`filament/filament:^5.0` ou `^3.0`/`^4.0`)
- **Banco de Dados Compartilhado:** SQLite 3 em modo **WAL (Write-Ahead Logging)** com `busy_timeout = 15000` (15 segundos).
- **Interoperabilidade Total com Go:** O painel PHP compartilha exatamente o mesmo arquivo de banco de dados (`dados/api.db`) que o microsserviço de alta performance em Go (`api-juridica-go`). O Go atende milhares de consultas jurídicas concorrentes enquanto o PHP executa operações de retaguarda sem contenção de lock.
- **Fuso Horário:** `America/Sao_Paulo` (Horário Oficial de Brasília) padronizado para fechamento de cotas e cálculos mensais.

---

## 🔐 Segurança e Autenticação

- **Tabela de Administradores:** Utiliza a coluna `senha_hash` com algoritmo Bcrypt nativo.
- **Tokens de API:** Gerados no formato `rj_live_<32_bytes_base64url>`.
  - O segredo completo é exibido **uma única vez** no modal do Filament após a criação.
  - Apenas o prefixo (`rj_live_...`) e o **hash SHA-256** de 64 caracteres são persistidos no banco.
  - O cálculo do hash SHA-256 em PHP coincide 100% com o algoritmo utilizado pelo middleware de autenticação da API em Go.

---

## 📊 Recursos do Painel (Filament Resources)

1. **Clientes (`ClienteResource`):**
   - Cadastro e edição de clientes com planos (`starter`, `pro`, `enterprise`).
   - Definição de limites mensais de consultas e limites de taxa por minuto.
   - Cálculo dinâmico do consumo mensal no horário de Brasília, saldo restante e barra de progresso da cota.
   - Relações vinculadas: Chaves de API e Histórico de Uso do cliente.

2. **Chaves de API (`ChaveResource`):**
   - Listagem com indicação de prefixo, data de emissão, última utilização e status (Ativa / Revogada).
   - Ação de revogação imediata com modal de confirmação.

3. **Histórico de Uso (`UsoResource`):**
   - Telemetria detalhada de cada requisição processada pela API em Go: endpoint, método HTTP, status code, tempo de resposta (ms), IP de origem e sinalização se a consulta foi faturada (`cobravel`/`cobrado`).

4. **Eventos e Auditoria (`EventoResource`):**
   - Registro cronológico imutável de ações administrativas (criação de cliente, emissão de chaves, revogação, alterações de cota).

5. **Administradores (`AdminResource`):**
   - Gestão de acessos à interface administrativa com controle de ativação/desativação.

---

## 📈 Widgets do Dashboard

- **StatsOverviewWidget:**
  - Total de consultas consumidas no mês corrente.
  - Quantidade de empresas/clientes ativas.
  - MRR estimado (Receita Recorrente Mensal contratada).
  - Bloqueios por estouro de cota (status HTTP 402 no mês).
- **ConsultasChartWidget:**
  - Gráfico temporal dos últimos 14 dias diferenciando requisições com sucesso (2xx/3xx) de falhas/bloqueios (4xx/5xx).
- **TopClientesWidget:**
  - Tabela com os 5 maiores consumidores do mês corrente, exibindo percentual de consumo da cota com alertas visuais.

---

## 🚀 Executando Localmente

### Pré-requisitos
- PHP 8.2+ com extensões `pdo_sqlite`, `bcmath`, `intl`, `zip`.
- Composer 2+

### Passos de Instalação

```bash
# 1. Instalar dependências
composer install

# 2. Configurar variáveis de ambiente
cp .env.example .env
php artisan key:generate

# 3. Executar migrações
php artisan migrate

# 4. Criar administrador inicial
php artisan db:seed
# Ou via comando interativo:
php artisan admin:criar --email=admin@empresa.com.br --nome="Administrador" --senha=admin123

# 5. Iniciar servidor local
php artisan serve --port=8080
```

Acesse o painel em: **`http://localhost:8080/admin`**
- **E-mail:** `admin@empresa.com.br`
- **Senha:** `admin123`

---

## 🧪 Testes Automatizados

A suíte de testes cobre regras de negócio críticas e compatibilidade de segurança:

```bash
# Executar todos os testes
php artisan test
# ou
vendor/bin/phpunit
```

Testes inclusos:
- `tests/Feature/AdminAuthTest.php`: Validação de login com coluna customizada `senha_hash` e autorização do Filament.
- `tests/Feature/ChaveGeracaoTest.php`: Geração segura do token `rj_live_`, cálculo do SHA-256 e revogação.
- `tests/Feature/ClienteResourceTest.php`: Cadastro e relacionamentos de clientes.
- `tests/Unit/ConsumoCalculoTest.php`: Cálculo exato do consumo mensal no fuso horário de Brasília e consistência de saldo.
