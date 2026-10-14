# WORKBOOK · API ROBÔ JURÍDICO (GO + LARAVEL FILAMENT)
# Manual Definitivo de Instalação, Operação e Integração

> **Versão da API:** 1.0.0 (Core Engine em Go 1.24+ & Painéis em Laravel 13 / Filament v5)  
> **Fontes Oficiais:** DataJud (Histórico Processual CNJ) + DJEN (Diário de Justiça Eletrônico Nacional)  
> **Ambientes Contemplados:** Local (Desenvolvimento/Testes) e Servidor VPS (Produção), com Docker ou Bare-Metal (Sem Docker)  
> **Data:** Outubro de 2026  

---

## 🧭 Como Usar este Workbook

Siga os capítulos na ordem. Em cerca de **15 a 20 minutos**, você terá toda a infraestrutura rodando no seu computador ou na sua VPS, com um administrador ativo, um cliente cadastrado, uma chave de API emitida e suas primeiras consultas integradas.

### Convenções Utilizadas
* **Comandos de Terminal:** Blocos de código escuros com sintaxe para **Windows (PowerShell)** e **Linux / macOS (Bash)**.
* **Variáveis Customizáveis:** Textos envolvidos por `<sinais-de-menor-maior>` (ex: `<sua-chave>`, `<seu-email>`) devem ser substituídos pelos seus valores reais.
* **Validação por Exercícios:** O **Capítulo 11** contém 5 exercícios práticos com critérios claros de sucesso para auditar sua instalação.

---

## 📑 Índice Geral

1. [Capítulo 01: O que é a API e a Arquitetura Dual](#capítulo-01-o-que-é-a-api-e-a-arquitetura-dual)
2. [Capítulo 02: Pré-requisitos de Ambiente (Local e VPS)](#capítulo-02-pré-requisitos-de-ambiente-local-e-vps)
3. [Capítulo 03: Instalação e Execução Local para Testes](#capítulo-03-instalação-e-execução-local-para-testes)
4. [Capítulo 04: Primeiro Acesso, Painel Admin e Portal do Cliente](#capítulo-04-primeiro-acesso-painel-admin-e-portal-do-cliente)
5. [Capítulo 05: Autenticação, Cotas, Rate Limit e Headers](#capítulo-05-autenticação-cotas-rate-limit-e-headers)
6. [Capítulo 06: Referência Completa dos Endpoints da API](#capítulo-06-referência-completa-dos-endpoints-da-api)
7. [Capítulo 07: Tabela de Erros Canônicos e Estratégia de Retry](#capítulo-07-tabela-de-erros-canônicos-e-estratégia-de-retry)
8. [Capítulo 08: Consumindo a API no seu Código (Python, JS, PHP, Go)](#capítulo-08-consumindo-a-api-no-seu-código)
9. [Capítulo 09: Configuração das Variáveis de Ambiente e Proxies](#capítulo-09-configuração-das-variáveis-de-ambiente-e-proxies)
10. [Capítulo 10: Guia Completo de Deploy em VPS (Ubuntu / Debian)](#capítulo-10-guia-completo-de-deploy-em-vps)
11. [Capítulo 11: Exercícios Práticos & Checklist de Conclusão](#capítulo-11-exercícios-práticos--checklist-de-conclusão)
12. [Capítulo 12: Resolução de Problemas, Backup e Operação Segura](#capítulo-12-resolução-de-problemas-backup-e-operação-segura)

---

## Capítulo 01: O que é a API e a Arquitetura Dual

A API do Robô Jurídico permite que seu sistema envie o número de um processo no formato CNJ (`NNNNNNN-DD.AAAA.J.TR.OOOO`) ou uma inscrição de OAB e receba em troca um **JSON unificado e sanitizado** contendo: capa dos autos, tribunal, partes em seus respectivos polos, advogados com OAB, histórico completo de movimentações e o teor integral das publicações do diário oficial.

Os dados são extraídos em tempo real de duas fontes públicas do Conselho Nacional de Justiça (CNJ):

```
                        ┌──────────────────────────────────────────────┐
                        │              SEU SISTEMA / CLIENTE           │
                        │    (App, ERP Jurídico, Bot de WhatsApp)      │
                        └──────────────────────┬───────────────────────┘
                                               │ HTTP / X-API-Key
                                               ▼
                        ┌──────────────────────────────────────────────┐
                        │        API ROBÔ JURÍDICO EM GO (:8000)       │
                        │  - Valida DV CNJ (Módulo 97 em microssegundos)│
                        │  - Verifica e debita cota mensal atomicamente │
                        │  - Dispara requisições simultâneas em paralelo│
                        └──────────────┬───────────────────────────────┘
                                       │
                      ┌────────────────┴────────────────┐
                      ▼                                 ▼
         ┌────────────────────────┐        ┌────────────────────────┐
         │     DATAJUD (CNJ)      │        │   DJEN (COMUNICAPJE)   │
         │ - Histórico de fases   │        │ - Publicações recentes │
         │ - Movimentos internos  │        │ - Teor integral da peça│
         │ - Órgão julgador       │        │ - Partes e advogados   │
         └────────────────────────┘        └────────────────────────┘
```

### Comparativo das Fontes Públicas

| Característica | DJEN (ComunicaPJe) | DataJud (Elasticsearch CNJ) |
| :--- | :--- | :--- |
| **O que traz** | Publicações oficiais do Diário de Justiça com teor integral do ato | Capa do processo e histórico cronológico de movimentações internas |
| **Partes e Advogados** | Sim, com nomes e números de OAB | Não publica nomes por restrição de sigilo |
| **Busca por OAB** | Sim (Única fonte pública oficial com busca por advogado) | Não |
| **Atualização** | Imediata (no dia útil seguinte ao ato) | Varia de dias a semanas dependendo do envio pelo tribunal |
| **Tempo de Resposta** | Rápido (1 a 4 segundos) | Oscila entre 7 segundos e 60+ segundos em horário de pico |

> [!IMPORTANT]
> **Prazo se controla pelo DJEN!**  
> Os prazos processuais fatais correm com base na publicação do Diário da Justiça. Use o endpoint `/v1/oab/{uf}/{numero}/publicacoes` para automação de prazos e o DataJud como complemento para o histórico dos autos.

### Por que a Arquitetura é Dividida em Go + PHP?

Em vez de forçar um único ecossistema a resolver tarefas conflitantes:
1. **Engine em Go (`api-juridica-go` - Porta 8000):** Desenvolvido em Go moderno sem dependência de runtime externo (`CGO_ENABLED=0`). Consome menos de **30 MB de memória RAM**, executa validações matemáticas em microssegundos e lida com centenas de conexões simultâneas sem engasgar.
2. **Painéis em PHP (`painel` - Porta 8080):** Desenvolvido com **Laravel 13 + Filament v5**. Entrega dois portais corporativos prontos para uso:
   - **`/admin` (Painel Administrativo):** Visão executiva da operação, faturamento (MRR), gestão global de clientes, alteração de cotas contratuais e trilha de auditoria.
   - **`/portal` (Portal do Cliente/Desenvolvedor):** Área exclusiva do cliente final para gerar e revogar chaves de API com segurança, acompanhar o consumo da sua cota mensal em tempo real e consultar o histórico das próprias requisições.
3. **Persistência Compartilhada:** Ambos os serviços compartilham o mesmo banco de dados SQLite (`dados/api.db`) em **modo WAL (Write-Ahead Logging)** com timeout de transação de 15 segundos (`busy_timeout=15000`), permitindo escritas seguras pelo PHP e leituras concorrentes ultrarrápidas pelo Go.

---

## Capítulo 02: Pré-requisitos de Ambiente (Local e VPS)

Dependendo da estratégia de execução que você escolher, certifique-se de que as ferramentas abaixo estão instaladas no seu computador ou no servidor:

### Matriz de Pré-requisitos

| Ferramenta | Modo com Docker | Modo sem Docker (Bare-Metal) | Onde Obter / Instalação |
| :--- | :---: | :---: | :--- |
| **Docker & Docker Compose** | **Obrigatório** | Não necessário | [docker.com](https://www.docker.com/) (Windows/Mac) ou repositório oficial no Linux |
| **Go (Golang)** | Não necessário | **Obrigatório** (1.22+) | [go.dev/dl](https://go.dev/dl/) |
| **PHP** | Não necessário | **Obrigatório** (8.2+) | [php.net](https://www.php.net/) (Windows) ou `apt install php8.4` (Linux) |
| **Composer** | Não necessário | **Obrigatório** (2.x) | [getcomposer.org](https://getcomposer.org/) |
| **Extensões PHP** | Não necessário | **Obrigatório** | `pdo_sqlite`, `mbstring`, `curl`, `intl`, `bcmath`, `zip` |
| **SQLite3** | Embutido | **Obrigatório** | Windows traz nativo no PHP; no Linux: `sudo apt install sqlite3` |
| **VS Code (Opcional)** | Recomendado | Recomendado | Com a extensão **REST Client** para testes rápidos de API |

### Como Testar se seu Ambiente está Pronto (Modo sem Docker)

Abra o terminal e verifique as versões instaladas:

#### Windows (PowerShell)
```powershell
go version
# Exemplo: go version go1.24.0 windows/amd64

php -v
# Exemplo: PHP 8.4.x (cli) ...

composer --version
# Exemplo: Composer version 2.8.x ...
```

#### Linux / macOS (Bash)
```bash
go version
php -v
composer --version
```

---

## Capítulo 03: Instalação e Execução Local para Testes

### Cenário A: Com Docker Compose (Recomendado — 1 Comando)

Esta é a opção mais rápida. O Docker Compose compilará o binário estático Go e provisionará o ambiente PHP/Filament com as dependências e seeds automáticos.

#### 1. Navegue até a pasta do projeto
```bash
cd 4-api-robo-juridico-php-go
```

#### 2. Suba todos os serviços
```bash
docker compose up -d --build
```

#### 3. Verifique o status dos containers
```bash
docker compose ps
```
Você verá dois serviços com status `healthy`:
* `robo-juridico-api` rodando na porta `8000`
* `robo-juridico-painel` rodando na porta `8080`

---

### Cenário B: Sem Docker (Nativo no Windows / Linux / macOS)

Se você preferir rodar nativamente na sua máquina para ter controle direto dos processos:

#### 1. Preparar a pasta de dados compartilhada
A partir da raiz de `4-api-robo-juridico-php-go`, crie a pasta onde o SQLite residirá:

* **Windows (PowerShell):**
  ```powershell
  New-Item -ItemType Directory -Force -Path "dados"
  ```
* **Linux / macOS:**
  ```bash
  mkdir -p dados
  ```

#### 2. Configurar e Subir o Painel PHP/Laravel (Terminal 1)
O Laravel será responsável por rodar as migrations e criar as tabelas iniciais.

```bash
cd painel

# Instalar dependências do Composer
composer install

# Criar o arquivo de variáveis de ambiente
cp .env.example .env

# Gerar a chave criptográfica da aplicação
php artisan key:generate

# Configurar o caminho do banco SQLite no .env
# Certifique-se de que DB_DATABASE aponte para ../dados/api.db
```

No Windows PowerShell ou Linux, configure o `.env` do painel com:
```env
DB_CONNECTION=sqlite
DB_DATABASE=../dados/api.db
```

Execute as migrations com os dados de exemplo (seeder):
```bash
php artisan migrate --force
php artisan db:seed --force
```

Inicie o servidor de desenvolvimento do painel:
```bash
php artisan serve --port=8080
```
O painel estará disponível em: `http://localhost:8080`

#### 3. Configurar e Subir a API em Go (Terminal 2)
Abra um **segundo terminal**:

```bash
cd api-juridica-go

# Executar a suíte de testes automatizados para garantir integridade
go test -v ./...

# Configurar a variável de caminho do banco apontando para a pasta compartilhada
```

* **Windows (PowerShell):**
  ```powershell
  $env:API_BANCO = "../dados/api.db"
  $env:PORT = "8000"
  go run ./cmd/api
  ```
* **Linux / macOS:**
  ```bash
  export API_BANCO="../dados/api.db"
  export PORT="8000"
  go run ./cmd/api
  ```

Você verá no terminal:
```
[INFO] Servidor Robô Jurídico API iniciado em http://0.0.0.0:8000
[INFO] Banco SQLite conectado em ../dados/api.db (WAL mode ativado)
```

---

## Capítulo 04: Primeiro Acesso, Painel Admin e Portal do Cliente

Com os dois serviços rodando, você tem acesso a dois ambientes web com propostas e permissões totalmente distintas:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            APLICAÇÃO LARAVEL 13                             │
├──────────────────────────────────────┬──────────────────────────────────────┤
│    PAINEL ADMINISTRATIVO (/admin)    │       PORTAL DO CLIENTE (/portal)    │
│  - Gestão de clientes e contratos    │  - Autoatendimento do desenvolvedor  │
│  - Faturamento mensal (MRR)          │  - Emissão e revogação de tokens     │
│  - Auditoria global do sistema       │  - Acompanhamento de consumo e logs  │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

### 1. Acesso ao Painel do Administrador (`/admin`)
Acesse no navegador: **[http://localhost:8080/admin](http://localhost:8080/admin)**
* **E-mail:** `admin@empresa.com.br`
* **Senha:** `admin123`

#### O que você encontra no Painel Admin:
* **Dashboard:** Faturamento mensal recorrente (MRR), total de clientes ativos, volume global de requisições e bloqueios por cota (HTTP 402).
* **Clientes:** Cadastro de novas empresas, definição do plano contratado (Start, Pro, Revenda ou Customizado), ajuste manual de cotas mensais e limite de requisições por minuto (RPM).
* **Chaves de API:** Visualização de todas as chaves emitidas no sistema, status (ativa/revogada) e opção de revogação emergencial.
* **Auditoria:** Histórico em tempo real de todas as requisições recebidas pela API em Go com código de status e latência.

> [!TIP]
> **Criando administradores via linha de comando:**  
> Caso queira criar um novo administrador pelo terminal, execute:
> ```bash
> php artisan admin:criar novo_admin@empresa.com.br --nome="Diretor Técnico" --senha="SenhaSegura123!"
> ```

---

### 2. Acesso ao Portal do Cliente / Desenvolvedor (`/portal`)
Acesse no navegador: **[http://localhost:8080/portal](http://localhost:8080/portal)**
* **E-mail:** `cliente@empresa.com.br`
* **Senha:** `cliente123`

#### O que o cliente final / desenvolvedor faz no Portal:
1. **Emissão de Chave de API:**
   - Acesse o menu **Minhas Chaves de API** e clique no botão **Gerar Nova Chave de API**.
   - Defina um identificador amigável (ex: *Sistema ERP Produção*).
   - O portal exibirá o token completo gerado no padrão `rj_live_...`.
   - **Copie a chave imediatamente!** Por segurança, o sistema armazena apenas o hash criptográfico **SHA-256** do token; a chave em texto puro nunca mais será exibida.
2. **Acompanhamento de Cota:**
   - O dashboard exibe uma barra de progresso visual com o consumo do mês vigente (ex: *150 / 2.000 consultas utilizadas*) e alerta automático caso atinja 80% ou 100%.
3. **Telemetria e Histórico:**
   - O menu **Histórico de Consultas** mostra os logs estritamente vinculados às chaves daquele cliente, exibindo endpoint chamado, status code (200, 404, 422, etc.), data/hora e tempo de resposta em milissegundos.
4. **Revogação Instantânea:**
   - Se uma chave for exposta ou precisar ser desativada, o próprio cliente clica em **Revogar**, invalidando o token no mesmo microssegundo no motor em Go.

---

## Capítulo 05: Autenticação, Cotas, Rate Limit e Headers

### Como Autenticar Requisições na API em Go

A API aceita a chave emitida de duas formas idênticas:

1. **Via Cabeçalho `X-API-Key` (Padrão legado e Python):**
   ```http
   X-API-Key: rj_live_sua_chave_secreta_aqui
   ```
2. **Via Cabeçalho `Authorization: Bearer` (Padrão REST moderno):**
   ```http
   Authorization: Bearer rj_live_sua_chave_secreta_aqui
   ```

### O que Desconta e o que NÃO Desconta da Cota Mensal

| Endpoint | Método | Desconta Cota? | Justificativa |
| :--- | :---: | :---: | :--- |
| `/saude` | `GET` | **Não** | Rota pública de healthcheck para monitoramento e proxies |
| `/openapi.json` e `/docs` | `GET` | **Não** | Especificação aberta e documentação interativa |
| `/v1/cnj/{numero}` | `GET` | **Não** | Cálculo matemático de Módulo 97 em memória (zero custo de rede) |
| `/v1/processos/{numero}/existe` | `GET` | **Não** | Checagem rápida de existência para novos processos |
| `/v1/conta` | `GET` | **Não** | Consulta de saldo de cota, plano e data de renovação |
| `/v1/processos/{numero}` | `GET` | **SIM (1 consulta)** | Consulta consolidada no DataJud e DJEN (completa ou rápida) |
| `/v1/processos/consultar` | `POST` | **SIM (1 consulta)** | Consulta enviando o número via payload JSON |
| `/v1/oab/{uf}/{numero}/processos` | `GET` | **SIM (1 consulta)** | Agrupamento de processos publicados para a OAB |
| `/v1/oab/{uf}/{numero}/publicacoes` | `GET` | **SIM (1 consulta)** | Varredura do diário oficial para controle de prazos |

### Regra de Ouro: Estorno Automático em Falhas
Se uma requisição terminar com código HTTP **4xx** ou **5xx** (por exemplo: processo inexistente `404`, número de processo digitado com dígito inválido `422`, ou fonte do CNJ fora do ar `503`), **a cota reservada é estornada na hora**. Seu cliente nunca paga por consultas frustradas!

### Cabeçalhos de Resposta (Telemetria do Cliente)
Em todas as chamadas autenticadas com chave de cliente, a API em Go anexa cabeçalhos de controle:

```http
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-ID: 7b3a9c1e2f8d04a5
X-Cota-Limite: 2000
X-Cota-Usada: 42
X-Cota-Restante: 1958
```

* **`X-Request-ID`:** Identificador único hexadecimal da requisição. Guarde este ID para auditoria ou suporte técnico.
* **`X-Cota-Restante`:** Saldo disponível de consultas antes do bloqueio por HTTP 402.

---

## Capítulo 06: Referência Completa dos Endpoints da API

*Endereço base local:* `http://localhost:8000`  
*(A API aceita as rotas tanto no prefixo `/v1/...` quanto `/api/v1/...`)*

---

### 1. `GET /saude`
Checagem de vida para healthcheck do Docker, Caddy e serviços de uptime. **Não exige autenticação.**

* **Exemplo de Chamada:**
  ```bash
  curl http://localhost:8000/saude
  ```
* **Resposta 200 OK:**
  ```json
  {
    "ok": true,
    "versao": "1.0.0"
  }
  ```

---

### 2. `GET /v1/cnj/{numero}`
Valida a máscara e o dígito verificador pelo cálculo de **Módulo 97 (big.Int)** e decodifica a qual tribunal, ramo da Justiça e ano o processo pertence. Não acessa fontes externas e **não consome cota**.

* **Parâmetros:**
  * `numero` (caminho): Número do processo, com ou sem pontuação (ex: `5010754-07.2024.8.13.0625` ou `50107540720248130625`).
* **Exemplo de Chamada:**
  ```bash
  curl -H "X-API-Key: <sua-chave>" \
    http://localhost:8000/v1/cnj/5010754-07.2024.8.13.0625
  ```
* **Resposta 200 OK:**
  ```json
  {
    "ok": true,
    "numero": "5010754-07.2024.8.13.0625",
    "numero_digitos": "50107540720248130625",
    "valido": true,
    "tribunal": "TJMG",
    "segmento": "Justiça Estadual",
    "ano": 2024,
    "origem_codigo": 625
  }
  ```

---

### 3. `GET /v1/conta`
Informa os detalhes do plano contratado pelo cliente associado à chave, quantas consultas foram feitas no mês corrente, quantas restam e a data/hora exata do reinício do ciclo. **Não consome cota.**

* **Exemplo de Chamada:**
  ```bash
  curl -H "X-API-Key: <sua-chave>" \
    http://localhost:8000/v1/conta
  ```
* **Resposta 200 OK:**
  ```json
  {
    "ok": true,
    "cliente": "Empresa Advocacia Digital",
    "plano": "Pro",
    "limite_mensal": 10000,
    "usado_no_mes": 142,
    "restante": 9858,
    "reinicia_em": "2026-11-01T03:00:00Z",
    "limite_por_minuto": 60
  }
  ```

---

### 4. `GET /v1/processos/{numero}`
A rota principal do Robô Jurídico. Dispara a consulta federada concorrente às fontes oficiais e entrega o processo unificado. **Consome 1 consulta da cota.**

* **Parâmetros:**
  * `numero` (caminho): Número CNJ (pontuado ou só dígitos).
  * `modo` (query, opcional):
    * `completo` *(padrão)*: Consulta DataJud + DJEN em paralelo e consolida tudo.
    * `rapido`: Consulta exclusivamente o DJEN. Responde em sub-segundos, trazendo capa, partes, advogados e publicações, sem o histórico de movimentações internas do DataJud.
  * `publicacoes` (query, opcional): `true` *(padrão)* ou `false` para dispensar as publicações no modo completo.
* **Exemplo de Chamada:**
  ```bash
  curl -H "X-API-Key: <sua-chave>" \
    "http://localhost:8000/v1/processos/5010754-07.2024.8.13.0625?modo=completo"
  ```
* **Resposta 200 OK (Estrutura Encurtada):**
  ```json
  {
    "ok": true,
    "consultado_em": "2026-10-14T15:30:00Z",
    "tempo_ms": 780,
    "avisos": [],
    "processo": {
      "numero": "5010754-07.2024.8.13.0625",
      "numero_digitos": "50107540720248130625",
      "titulo": "EMPRESA EXEMPLO S.A. × FULANO DE TAL",
      "tribunal": "TJMG",
      "segmento": "Justiça Estadual",
      "grau": "TR",
      "classe": "Recurso Inominado Cível",
      "assunto": "Indenização por Dano Moral",
      "orgao_julgador": "Turma Recursal Exclusiva de BH",
      "data_ultima_movimentacao": "2026-10-10",
      "data_ultima_movimentacao_formatada": "10/10/2026",
      "polos": {
        "ativo": [{ "nome": "EMPRESA EXEMPLO S.A.", "polo": "ativo", "advogados": [] }],
        "passivo": [{ "nome": "FULANO DE TAL", "polo": "passivo", "advogados": [] }],
        "terceiros": []
      },
      "advogados": [
        { "nome": "DOUTOR EXEMPLO", "oab_numero": "123456", "oab_uf": "MG", "oab": "OAB/MG 123456" }
      ],
      "movimentacoes": [
        {
          "data": "2026-10-10T14:22:00Z",
          "data_formatada": "10/10/2026",
          "tipo": "decisao",
          "descricao": "Acórdão publicado",
          "fonte": "datajud"
        }
      ],
      "publicacoes": [
        {
          "id": "89123456",
          "data_disponibilizacao": "2026-10-11",
          "data_formatada": "11/10/2026",
          "tipo_comunicacao": "Intimação",
          "texto": "PODER JUDICIÁRIO DO ESTADO DE MINAS GERAIS... Fica intimado o recorrido para contra-arrazoar...",
          "advogados": [{ "nome": "DOUTOR EXEMPLO", "oab": "OAB/MG 123456" }]
        }
      ],
      "totais": {
        "partes": 2,
        "advogados": 1,
        "movimentacoes": 48,
        "publicacoes": 6
      },
      "fontes": ["datajud", "comunica"]
    }
  }
  ```

---

### 5. `POST /v1/processos/consultar`
Equivalente ao endpoint anterior, mas aceitando o payload via `POST` com `application/json`. Ideal para sistemas que preferem passar parâmetros estruturados no corpo da mensagem. **Consome 1 consulta.**

* **Exemplo de Chamada:**
  ```bash
  curl -X POST http://localhost:8000/v1/processos/consultar \
    -H "Authorization: Bearer <sua-chave>" \
    -H "Content-Type: application/json" \
    -d '{
      "numero_processo": "5010754-07.2024.8.13.0625",
      "modo": "completo"
    }'
  ```

---

### 6. `GET /v1/oab/{uf}/{numero}/publicacoes`
Busca todas as intimações e despachos publicados no DJEN onde a OAB informada figura como destinatária. Essencial para **robôs de captura diária de intimações e prazos processuais**. **Consome 1 consulta.**

* **Parâmetros:**
  * `uf` (caminho): Sigla do estado com 2 letras (ex: `MG`, `SP`, `RJ`).
  * `numero` (caminho): Número da inscrição na OAB (somente números).
  * `dias` (query, opcional): Janela retroativa em dias (1 a 90; padrão: `7`).
  * `paginas` (query, opcional): Número de páginas de 100 itens (1 a 20; padrão: `1`).
* **Exemplo de Chamada:**
  ```bash
  curl -H "X-API-Key: <sua-chave>" \
    "http://localhost:8000/v1/oab/MG/123456/publicacoes?dias=3"
  ```
* **Resposta 200 OK:**
  ```json
  {
    "ok": true,
    "oab": "OAB/MG 123456",
    "inicio": "2026-10-11",
    "fim": "2026-10-14",
    "total": 4,
    "publicacoes": [
      {
        "id": "99401221",
        "data_disponibilizacao": "2026-10-13",
        "data_formatada": "13/10/2026",
        "tipo_comunicacao": "Intimação",
        "numero_processo": "5010754-07.2024.8.13.0625",
        "texto": "Fica a parte autora intimada a manifestar sobre a petição juntada no prazo de 5 dias...",
        "link": "https://pje.tjmg.jus.br/..."
      }
    ]
  }
  ```

---

### 7. `GET /v1/oab/{uf}/{numero}/processos`
Retorna uma lista consolidada dos processos em que a OAB informada teve publicações recentes, agrupando as publicações por processo. **Consome 1 consulta.**

* **Parâmetros:**
  * `uf`, `numero` (caminho): Sigla e número da OAB.
  * `dias` (query, opcional): Janela de busca (1 a 365 dias; padrão: `30`).
  * `tribunal` (query, opcional): Filtra por tribunal específico (ex: `TJMG`, `TRT3`).
* **Exemplo de Chamada:**
  ```bash
  curl -H "X-API-Key: <sua-chave>" \
    "http://localhost:8000/v1/oab/MG/123456/processos?dias=30&tribunal=TJMG"
  ```

---

## Capítulo 07: Tabela de Erros Canônicos e Estratégia de Retry

Todos os erros retornados pela API seguem um **envelope padronizado e previsível**. Seu código cliente deve sempre validar a chave `codigo`, pois a mensagem de texto pode evoluir:

```json
{
  "ok": false,
  "erro": {
    "codigo": "numero_invalido",
    "mensagem": "Número fora do padrão CNJ (NNNNNNN-DD.AAAA.J.TR.OOOO) ou com dígito verificador errado."
  },
  "id_requisicao": "7b3a9c1e2f8d04a5"
}
```

### Catálogo de Erros

| Status HTTP | Código Canônico | Quando Ocorre | O que seu Sistema deve Fazer |
| :---: | :--- | :--- | :--- |
| **401** | `nao_autorizado` | Chave ausente, inválida ou revogada | Verifique se a chave foi informada no cabeçalho `X-API-Key` ou `Authorization: Bearer` |
| **402** | `cota_esgotada` | A cota mensal de consultas do cliente foi atingida | Alerte o usuário para aguardar o dia 1º ou fazer upgrade de plano no painel |
| **403** | `conta_suspensa` | Cliente foi desativado no Painel Administrativo | Contate a equipe financeira ou de suporte |
| **404** | `processo_nao_encontrado` | Nenhuma fonte oficial conhece esse processo | Se for recém-distribuído, aguarde a sincronização do CNJ. Segredo de justiça não é listado |
| **404** | `rota_inexistente` | Endereço ou método HTTP digitado incorretamente | Revise a URL do endpoint conforme o Capítulo 06 |
| **422** | `numero_invalido` | Máscara do CNJ violada ou falha no Módulo 97 | Corrija o número antes de reenviar (não desconta cota) |
| **422** | `uf_invalida` | Sigla da OAB inválida (ex: letras incorretas) | Envie uma das 27 siglas de estados brasileiros |
| **429** | `limite_excedido` | Ultrapassou as requisições por minuto (RPM) | Respeite os segundos indicados no cabeçalho `Retry-After` |
| **429** | `muitas_tentativas` | Bloqueio de segurança por excesso de chaves inválidas | Aguarde a janela de bloqueio do IP expirar (15 minutos) |
| **500** | `erro_interno` | Falha inesperada no servidor | Registre o `id_requisicao` e informe o suporte |
| **502** | `falha_nao_esperada` | Resposta anômala ou corrompida de uma fonte do CNJ | Tente novamente após alguns segundos |
| **503** | `fonte_indisponivel` | Instabilidade ou manutenção no DataJud ou DJEN | Aguarde o tempo do `Retry-After` (geralmente 30s) ou tente `modo=rapido` |

> [!TIP]
> **Regra Prática de Retry Automático:**  
> Somente repita a requisição de forma programática se o código retornado for **429, 502 ou 503**, aguardando o tempo indicado pelo cabeçalho `Retry-After`. Para erros 401, 402, 404 e 422, **repetir a chamada imediatamente trará o exato mesmo resultado e desperdiçará recursos**.

---

## Capítulo 08: Consumindo a API no seu Código

Copie e adapte os snippets abaixo para sua linguagem preferida. Em todas as integrações, **a chave deve ser injetada via variável de ambiente**.

### 1. Python (Com `httpx` e Retry Resiliente)

```python
import os
import time
import httpx

API_URL = os.getenv("ROBO_API_URL", "http://localhost:8000")
API_KEY = os.environ["ROBO_API_KEY"]  # Defina no ambiente!

def consultar_processo(numero: str, modo: str = "completo", tentativas: int = 3) -> dict:
    url = f"{API_URL}/v1/processos/{numero}"
    headers = {"X-API-Key": API_KEY}
    params = {"modo": modo}

    with httpx.Client(timeout=90.0) as client:
        for tentativa in range(1, tentativas + 1):
            resposta = client.get(url, headers=headers, params=params)

            # Se for erro temporário de fonte ou rate limit, repete com backoff
            if resposta.status_code in (429, 502, 503) and tentativa < tentativas:
                espera = int(resposta.headers.get("Retry-After", 5))
                print(f"[Aviso] Tentativa {tentativa} falhou ({resposta.status_code}). Aguardando {espera}s...")
                time.sleep(espera)
                continue

            dados = resposta.json()
            if not dados.get("ok"):
                erro = dados.get("erro", {})
                raise RuntimeError(f"Erro {resposta.status_code} [{erro.get('codigo')}]: {erro.get('mensagem')}")

            print(f"[Sucesso] Cota restante: {resposta.headers.get('X-Cota-Restante')}")
            return dados["processo"]

    raise TimeoutError("Fontes do CNJ indisponíveis após múltiplas tentativas.")

if __name__ == "__main__":
    p = consultar_processo("5010754-07.2024.8.13.0625", modo="rapido")
    print("Processo:", p["titulo"], "| Tribunal:", p["tribunal"])
    for mov in p["movimentacoes"][:3]:
        print(f" -> {mov['data_formatada']}: {mov['descricao']}")
```

---

### 2. Node.js / JavaScript Moderno (`fetch` Nativo)

```javascript
// consulta.mjs - Compatível com Node 18+
const API_URL = process.env.ROBO_API_URL || "http://localhost:8000";
const API_KEY = process.env.ROBO_API_KEY;

if (!API_KEY) {
  console.error("Erro: Defina a variável ROBO_API_KEY no seu ambiente.");
  process.exit(1);
}

async function consultarProcesso(numero, modo = "completo") {
  const url = `${API_URL}/v1/processos/${encodeURIComponent(numero)}?modo=${modo}`;
  
  const res = await fetch(url, {
    method: "GET",
    headers: {
      "Authorization": `Bearer ${API_KEY}`,
      "Accept": "application/json"
    }
  });

  const json = await res.json();

  if (!json.ok) {
    throw new Error(`Erro ${res.status} [${json.erro?.codigo}]: ${json.erro?.mensagem}`);
  }

  console.log(`Cota restante: ${res.headers.get("X-Cota-Restante")}`);
  return json.processo;
}

(async () => {
  try {
    const proc = await consultarProcesso("5010754-07.2024.8.13.0625", "completo");
    console.log(`Sucesso: ${proc.titulo} (${proc.tribunal})`);
    console.log(`Total de publicações: ${proc.totais.publicacoes}`);
  } catch (err) {
    console.error("Falha na consulta:", err.message);
  }
})();
```

---

### 3. PHP Moderno (cURL Nativo com Tratamento de Headers)

```php
<?php
// consulta.php
$apiUrl = getenv('ROBO_API_URL') ?: 'http://localhost:8000';
$apiKey = getenv('ROBO_API_KEY');

if (!$apiKey) {
    die("Defina a variável ROBO_API_KEY no seu ambiente.\n");
}

function consultarProcesso(string $numero, string $modo = 'completo'): array {
    global $apiUrl, $apiKey;
    
    $ch = curl_init("{$apiUrl}/v1/processos/{$numero}?modo={$modo}");
    curl_setopt_array($ch, [
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_TIMEOUT        => 90,
        CURLOPT_HTTPHEADER     => [
            "X-API-Key: {$apiKey}",
            "Accept: application/json"
        ]
    ]);

    $resposta = curl_exec($ch);
    $status   = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);

    $dados = json_decode($resposta, true);

    if ($status !== 200 || !($dados['ok'] ?? false)) {
        $codigo = $dados['erro']['codigo'] ?? 'erro_desconhecido';
        $msg    = $dados['erro']['mensagem'] ?? 'Falha na requisição';
        throw new Exception("Erro HTTP {$status} [{$codigo}]: {$msg}");
    }

    return $dados['processo'];
}

try {
    $proc = consultarProcesso('5010754-07.2024.8.13.0625');
    echo "Processo: " . $proc['titulo'] . " | Tribunal: " . $proc['tribunal'] . "\n";
} catch (Exception $e) {
    echo "Erro: " . $e->getMessage() . "\n";
}
```

---

### 4. Go Nativo (`net/http`)

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type RespostaProcesso struct {
	OK       bool `json:"ok"`
	Processo struct {
		Titulo   string `json:"titulo"`
		Tribunal string `json:"tribunal"`
	} `json:"processo"`
	Erro struct {
		Codigo   string `json:"codigo"`
		Mensagem string `json:"mensagem"`
	} `json:"erro"`
}

func main() {
	apiKey := os.Getenv("ROBO_API_KEY")
	url := "http://localhost:8000/v1/processos/5010754-07.2024.8.13.0625"

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-API-Key", apiKey)

	cliente := &http.Client{Timeout: 60 * time.Second}
	resp, err := cliente.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var resultado RespostaProcesso
	json.NewDecoder(resp.Body).Decode(&resultado)

	if !resultado.OK {
		fmt.Printf("Erro %d: %s\n", resp.StatusCode, resultado.Erro.Mensagem)
		return
	}

	fmt.Printf("Sucesso! Processo: %s (%s)\n", resultado.Processo.Titulo, resultado.Processo.Tribunal)
}
```

---

## Capítulo 09: Configuração das Variáveis de Ambiente e Proxies

### 1. Dicionário de Variáveis da API em Go (`api-juridica-go`)

| Variável | Padrão | Descrição |
| :--- | :---: | :--- |
| `HOST` | `0.0.0.0` | IP de escuta do servidor HTTP da API |
| `PORT` | `8000` | Porta onde a API responde |
| `API_BANCO` ou `DB_PATH` | `dados/api.db` | Caminho relativo ou absoluto do arquivo SQLite compartilhado |
| `DATAJUD_API_KEY` | *(embutida)* | Chave pública do CNJ para busca no DataJud |
| `DATAJUD_TIMEOUT` | `90` | Tempo limite em segundos para o retorno do DataJud |
| `DATAJUD_TENTATIVAS` | `3` | Número de retries automáticos em caso de falha no DataJud |
| `COMUNICA_TIMEOUT` | `12` | Tempo limite em segundos para o diário DJEN |
| `COMUNICA_TENTATIVAS`| `3` | Número de retries no DJEN |
| `API_LIMITE_POR_MINUTO`| `60` | Limite padrão de requisições por minuto (RPM) |
| `API_CHAVES` | `""` | Chaves internas de infraestrutura (`cliente:hash_minimo_32_chars`) |

### 2. Dicionário de Variáveis do Painel em PHP (`painel/.env`)

| Variável | Padrão | Descrição |
| :--- | :---: | :--- |
| `APP_NAME` | `Robô Jurídico Admin` | Nome apresentado nos cabeçalhos dos painéis |
| `APP_ENV` | `production` | Ambiente de execução (`local` ou `production`) |
| `APP_KEY` | `base64:...` | Chave de criptografia de sessões do Laravel |
| `APP_URL` | `http://localhost:8080` | URL pública onde o painel é acessado |
| `DB_CONNECTION` | `sqlite` | Conexão de banco de dados (sempre `sqlite`) |
| `DB_DATABASE` | `/app/dados/api.db` | Caminho absoluto ou relativo para a base SQLite |

---

### 3. Proxies Residenciais: Quando você realmente precisa?

```
                        SEU SERVIDOR ESTÁ LOCALIZADO ONDE?
                                         │
                 ┌───────────────────────┴───────────────────────┐
                 ▼                                               ▼
         [ NO BRASIL ]                                   [ NO EXTERIOR ]
  (Na sua máquina ou VPS BR)                       (AWS us-east, Hetzner, OVH)
                 │                                               │
                 ▼                                               ▼
      NÃO PRECISA DE PROXY!                           PROXY É OBRIGATÓRIO!
As fontes públicas do CNJ aceitam             O DJEN bloqueia IPs internacionais
conexões domésticas e de datacenters BR.      com HTTP 403 Forbidden por geofence.
```

Se sua VPS estiver fora do Brasil (ex: DigitalOcean Nova York, AWS Virgínia, Hetzner Alemanha, Oracle Cloud Ashburn), configure o gateway de proxies brasileiros no `.env` da API Go:

```env
# Configuração de Gateway Residencial com rotação por porta
CONSULTA_GATEWAY_HOST=gw.seuprovedor.com
CONSULTA_GATEWAY_LOGIN=seu-usuario
CONSULTA_GATEWAY_SENHA=sua-senha
CONSULTA_GATEWAY_PORTA_INICIAL=10000
CONSULTA_GATEWAY_PORTAS=50
```
O motor em Go distribuirá as requisições em paralelo entre as portas, garantindo que as chamadas ao DJEN saiam sempre com endereço IP residencial brasileiro.

---

## Capítulo 10: Guia Completo de Deploy em VPS (Ubuntu / Debian)

### Cenário 1: Deploy com Docker Compose na VPS (Recomendado)

#### 1. Conecte na VPS e instale o Docker
```bash
ssh root@sua-vps-ip

# Instalação do Docker e plugin Compose oficial
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
```

#### 2. Transfira o projeto para o servidor
```bash
# Na sua máquina local, envie a pasta da stack para a VPS:
rsync -avz --exclude 'dados/*.db' --exclude 'painel/vendor' \
  4-api-robo-juridico-php-go/ root@sua-vps-ip:/opt/robo-juridico/
```

#### 3. Ajuste permissões e suba os containers
```bash
cd /opt/robo-juridico
mkdir -p dados
chmod -R 775 dados

# Iniciar em segundo plano
docker compose up -d --build
```

#### 4. Configure o Caddy para HTTPS Automático (SSL Gratuito)
Instale o Caddy (`sudo apt install -y caddy`) e edite `/etc/caddy/Caddyfile`:

```caddy
# Subdomínio da API em Go
api.seudominio.com.br {
    reverse_proxy 127.0.0.1:8000
}

# Subdomínio dos Painéis em PHP
painel.seudominio.com.br {
    reverse_proxy 127.0.0.1:8080
}
```

Recarregue o Caddy:
```bash
sudo systemctl reload caddy
```
O Caddy cuidará automaticamente do certificado SSL (Let's Encrypt). A API estará ativa em `https://api.seudominio.com.br` e o painel em `https://painel.seudominio.com.br`.

---

### Cenário 2: Deploy Bare-Metal / Sem Docker na VPS

Para quem deseja extrair o máximo de desempenho de uma VPS modesta (ex: 1 vCPU e 1 GB RAM):

#### 1. Instalar Go, PHP 8.4 e Nginx na VPS
```bash
sudo apt update
sudo apt install -y software-properties-common
sudo add-apt-repository -y ppa:ondrej/php
sudo apt update

sudo apt install -y php8.4-fpm php8.4-cli php8.4-sqlite3 php8.4-curl \
                    php8.4-mbstring php8.4-xml php8.4-intl php8.4-zip \
                    sqlite3 nginx composer golang-go git
```

#### 2. Compilar o Binário Estático da API em Go
```bash
cd /opt/robo-juridico/api-juridica-go

CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags="-s -w -X main.version=1.0.0" \
  -o /opt/robo-juridico/bin/api ./cmd/api

chmod +x /opt/robo-juridico/bin/api
```

#### 3. Criar o Serviço Systemd para a API em Go
Crie `/etc/systemd/system/robo-juridico-api.service`:

```ini
[Unit]
Description=Robô Jurídico — API Core em Go
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/robo-juridico
Environment="PORT=8000"
Environment="API_BANCO=/opt/robo-juridico/dados/api.db"
Environment="APP_ENV=production"
ExecStart=/opt/robo-juridico/bin/api
Restart=always
RestartSec=3s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Ative e inicie o serviço da API:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now robo-juridico-api
```

#### 4. Configurar o Painel Laravel com PHP-FPM e Nginx
```bash
cd /opt/robo-juridico/painel
composer install --no-dev --optimize-autoloader
cp .env.example .env
php artisan key:generate
```

Configure no `.env`:
```env
APP_ENV=production
APP_DEBUG=false
APP_URL=https://painel.seudominio.com.br
DB_CONNECTION=sqlite
DB_DATABASE=/opt/robo-juridico/dados/api.db
```

Execute as migrações:
```bash
php artisan migrate --force
php artisan db:seed --force
```

Ajuste as permissões para o usuário do Nginx (`www-data`):
```bash
sudo chown -R www-data:www-data /opt/robo-juridico/dados
sudo chown -R www-data:www-data /opt/robo-juridico/painel/storage /opt/robo-juridico/painel/bootstrap/cache
```

Crie o arquivo do Nginx `/etc/nginx/sites-available/robo-juridico`:
```nginx
server {
    listen 80;
    server_name painel.seudominio.com.br;
    root /opt/robo-juridico/painel/public;

    index index.php;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.4-fpm.sock;
    }
}
```

Ative o site e configure o SSL gratuito via Certbot:
```bash
sudo ln -s /etc/nginx/sites-available/robo-juridico /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d painel.seudominio.com.br
```

---

## Capítulo 11: Exercícios Práticos & Checklist de Conclusão

Realize os 5 exercícios abaixo para certificar que sua infraestrutura está 100% pronta para produção:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 1. Verificação de Saúde e Portas                                   [10 MIN] │
│ 2. Emissão Autônoma de Chave no Portal                             [10 MIN] │
│ 3. Validação Matemática de CNJ e Teste de Erro 422                 [10 MIN] │
│ 4. Consulta Completa vs Consulta Rápida                            [15 MIN] │
│ 5. Monitoramento de Intimações da OAB                              [15 MIN] │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Exercício 1: Saúde da Instalação
- [ ] Chamar `http://localhost:8000/saude` (deve retornar `{"ok": true, "versao": "1.0.0"}`).
- [ ] Abrir no navegador a documentação interativa em `http://localhost:8000/docs`.
- [ ] **Critério de Conclusão:** O Swagger UI carrega perfeitamente e o endpoint `/saude` responde 200.

### Exercício 2: Gestão de Acessos e Primeira Chave
- [ ] Acessar `http://localhost:8080/admin` com `admin@empresa.com.br` / `admin123`.
- [ ] Acessar `http://localhost:8080/portal` com `cliente@empresa.com.br` / `cliente123`.
- [ ] No Portal, ir em **Minhas Chaves de API**, gerar uma nova chave e copiar o token `rj_live_...`.
- [ ] Executar chamada no terminal: `curl -H "X-API-Key: <chave>" http://localhost:8000/v1/conta`.
- [ ] **Critério de Conclusão:** A chamada retorna o nome do cliente e a cota do plano com status 200.

### Exercício 3: Validação de CNJ e Proteção de Cota
- [ ] Executar: `curl -H "X-API-Key: <chave>" http://localhost:8000/v1/cnj/5010754-07.2024.8.13.0625`.
- [ ] Trocar o último dígito do processo (ex: trocar `0625` por `0629`) e disparar novamente.
- [ ] Observar o retorno do erro `422 numero_invalido`.
- [ ] Chamar `/v1/conta` e verificar que o saldo de cota **não foi consumido**.
- [ ] **Critério de Conclusão:** A validação Módulo 97 barrou o número adulterado sem gastar cota.

### Exercício 4: Consulta de Processo e Comparação de Modos
- [ ] Disparar consulta completa: `GET /v1/processos/5010754-07.2024.8.13.0625?modo=completo`.
- [ ] Anotar o tempo de resposta (`tempo_ms`) e a presença do histórico de movimentações.
- [ ] Disparar a mesma consulta com `modo=rapido`.
- [ ] **Critério de Conclusão:** O modo rápido responde significativamente mais rápido, mantendo partes e advogados.

### Exercício 5: Varredura de Intimações da OAB
- [ ] Executar `GET /v1/oab/MG/123456/publicacoes?dias=7`.
- [ ] Conferir o campo `texto` da publicação e verificar que o HTML foi completamente sanitizado.
- [ ] **Critério de Conclusão:** O retorno contém a lista de intimações do DJEN prontas para exibição em tela.

---

## Capítulo 12: Resolução de Problemas, Backup e Operação Segura

### Diagnóstico Rápido de Falhas

| Sintoma Encontrado | Causa Mais Provável | Solução Imediata |
| :--- | :--- | :--- |
| `database is locked` no SQLite | Múltiplas conexões tentando escrever simultaneamente sem timeout | Verifique se o modo WAL está ativado e confirme que `busy_timeout=15000` está configurado |
| Status `401 nao_autorizado` | Cabeçalho incorreto ou chave revogada | No Windows PowerShell use `curl.exe` (e não o alias `curl`) e garanta o prefixo `rj_live_` |
| Status `403 Forbidden` no DJEN | O servidor está fora do Brasil e não possui proxy residencial | Configure as variáveis `CONSULTA_GATEWAY_*` com provedor de proxy brasileiro (Capítulo 09) |
| Porta 8000 ou 8080 já ocupada | Outro processo local está utilizando a porta | Altere a porta no `docker-compose.yml` (ex: `8001:8000` ou `8081:8080`) |
| Erro de permissão em `dados/api.db` no Linux | O usuário do PHP (`www-data`) não possui permissão de escrita | Execute: `sudo chown -R www-data:www-data /opt/robo-juridico/dados` e `sudo chmod -R 775 /opt/robo-juridico/dados` |

---

### Rotina de Backup a Quente do SQLite

O SQLite permite backups a quente consistentes **sem necessidade de parar os serviços**:

Crie o script `/opt/robo-juridico/backup.sh`:
```bash
#!/usr/bin/env bash
set -e

DATA=$(date +"%Y%m%d_%H%M%S")
DESTINO="/opt/robo-juridico/backups"
mkdir -p "$DESTINO"

# Gera backup atômico online
sqlite3 /opt/robo-juridico/dados/api.db ".backup /tmp/api_temp.db"
gzip -c /tmp/api_temp.db > "$DESTINO/api_$DATA.db.gz"
rm -f /tmp/api_temp.db

# Remove backups com mais de 30 dias
find "$DESTINO" -type f -name "api_*.db.gz" -mtime +30 -delete
echo "[OK] Backup gerado com sucesso: api_$DATA.db.gz"
```

Adicione ao `crontab` para rodar todas as madrugadas:
```cron
0 3 * * * /opt/robo-juridico/backup.sh >> /var/log/backup-sqlite.log 2>&1
```

---

### Operação Responsável e Conformidade com a LGPD

1. **Dados Pessoais:** Nomes de partes e advogados trafegados nas respostas são dados pessoais regulados pela LGPD. Guarde em seu banco de dados exclusivamente as informações estritamente necessárias para a finalidade da sua aplicação jurídica.
2. **Segredo de Justiça:** Processos em segredo de justiça não são expostos pelas APIs públicas do CNJ. Se um processo não for encontrado (`404`), isso pode significar que ele tramita sob sigilo.
3. **Respeito aos Servidores Públicos:** Utilize o cache da API, respeite as janelas de rate limiting e agende varreduras em lote para horários de menor movimento (entre 22:00 e 06:00).

---

## ⚡ Folha de Referência Rápida (Cheat Sheet)

```bash
# ==============================================================================
# COMANDOS ESSENCIAIS — ROBÔ JURÍDICO
# ==============================================================================

# Subir toda a stack com Docker
docker compose up -d --build

# Visualizar logs em tempo real
docker compose logs -f api-juridica-go
docker compose logs -f painel-admin

# Parar os serviços
docker compose down

# Criar novo Administrador via CLI
php artisan admin:criar admin@empresa.com.br --nome="Admin" --senha="suasenha"

# Criar novo Cliente via CLI
php artisan cliente:criar cliente@empresa.com.br --nome="Empresa" --plano=pro --senha="senha"

# Healthcheck rápido no terminal
curl http://localhost:8000/saude

# Consulta rápida de processo
curl -H "X-API-Key: <sua-chave>" "http://localhost:8000/v1/processos/<numero>?modo=rapido"
```
