# Robô Jurídico — API de Consulta de Processos Judiciais

Plataforma de alta disponibilidade para busca e consulta unificada de processos judiciais brasileiros diretamente nas fontes públicas oficiais do Conselho Nacional de Justiça (CNJ): **DataJud** e **DJEN (Comunica)**.

---

## 📌 Visão Geral

O **Robô Jurídico** resolve a fragmentação dos dados jurídicos no Brasil através de um mecanismo concorrente de fusão de dados em tempo real:
* **DataJud (Elasticsearch):** Extrai capa, classe, assunto, comarca, órgão julgador e a árvore profunda de movimentações internas do processo.
* **DJEN / Comunica (Diário da Justiça Eletrônico Nacional):** Extrai o teor completo dos atos e despachos publicados, partes envolvidas com seus polos e advogados vinculados (com OAB e UF).

O projeto é disponibilizado em quatro modalidades prontas para uso:
1. **Biblioteca Python** (`from consulta_processos import consultar_processo, buscar_por_oab`)
2. **Interface Web Local** (`consulta-processos web` em `http://localhost:8765`, sem dependências externas)
3. **API REST FastAPI / Go com Painel SaaS & Calculadora Judicial** (`3-api-robo-juridico/` e `4-api-robo-juridico-php-go/`), incluindo documentação interativa Swagger (`/docs`), gestão de clientes, chaves (`rj_live_...`), motor de cálculos judiciais (DrCalc / Lei 14.905/2024 / Art. 354 CC / Art. 523 CPC), links diretos para certidões PDF do DJEN (`/api/v1/comunicacao/{hash}/certidao`), correlação inteligente DataJud ↔ DJEN e Portal do Cliente (`/admin` e `/portal`).
4. **Aplicativo Desktop On-Premise Windows & Linux** (`5-app-juridico-desktop/`): Executável nativo autocontido com SQLite local, motor de cálculo pericial completo, consulta CNJ unificada com banner em tempo real das fontes oficiais, visualizador de Certidão/Documento Oficial A4 para andamentos do DataJud e DJEN, e licenciamento criptográfico ECDSA por Hardware ID.

---

## 📚 Documentação do Projeto

Toda a documentação técnica, especificação de requisitos e guias operacionais elaborados estão disponíveis na pasta [`docs/`](docs/):

| Documento | Descrição |
|---|---|
| 📋 [**01 - Análise do Projeto**](docs/01-ANALISE-DO-PROJETO.md) | Diagnóstico arquitetural, análise SWOT, viabilidade e avaliação de riscos. |
| 📄 [**02 - PRD (Requisitos de Produto)**](docs/02-PRD-DOCUMENTO-DE-REQUISITOS-DE-PRODUTO.md) | Personas, requisitos funcionais/não-funcionais, modelo de monetização B2B e roadmap. |
| 🏗️ [**03 - Especificação Técnica e Arquitetura**](docs/03-DOCUMENTACAO-TECNICA-E-ARQUITETURA.md) | Diagramas C4/Mermaid, fluxo concorrente, cache anti-thundering herd, banco SQLite WAL e segurança. |
| 🔌 [**04 - Manual de Integração e API Reference**](docs/04-MANUAL-DE-INTEGRACAO-E-API-REFERENCE.md) | Catálogo de endpoints, dicionário de dados, cabeçalhos de saldo, tratamento de erros e snippets (Python, Node.js, cURL). |
| 🛡️ [**05 - Guia de Operações, Segurança e LGPD**](docs/05-GUIA-DE-OPERACAO-SEGURANCA-E-LGPD.md) | Deploy em produção (Docker, Nginx com TLS), proxies residenciais brasileiros, backup quente do SQLite e conformidade LGPD. |
| 💻 [**08 - Guia do Sistema Desktop On-Premise**](docs/08-GUIA-SISTEMA-JURIDICO-DESKTOP-ON-PREMISE.md) | Arquitetura do executável desktop para Windows/Linux, licenciamento criptográfico por Hardware ID e motor pericial offline. |
| 🧭 [**Índice Geral de Documentação**](docs/README.md) | Mapa de navegação unificado da documentação. |

---

## 🐳 Executando com Docker (Recomendado para Produção)

A forma mais rápida e isolada de rodar a API é através do Docker Compose:

```bash
# 1. Copie o arquivo de variáveis de ambiente
copy .env.exemplo .env   # Windows
# cp .env.exemplo .env   # Linux / macOS

# 2. Construa a imagem e suba o contêiner em segundo plano
docker compose up -d --build

# 3. Verifique a integridade e os logs
docker compose ps
docker compose logs -f
```
* **API REST:** `http://localhost:8000/v1/...`
* **Swagger Interativo:** `http://localhost:8000/docs`
* **Painel Administrativo:** `http://localhost:8000/admin`
* O banco de dados SQLite (`api.db`) é persistido automaticamente na pasta local `./dados/`.

---

## 🚀 Instalação Tradicional (Sem Docker)

### 1. Pré-requisitos
* Python 3.10 ou superior (recomendado 3.12)
* Git

### 2. Configurando o Ambiente Virtual
```bash
# Clone o repositório
git clone https://github.com/rdrgzma/robo-juridico-api.git
cd robo-juridico-api

# Acesse a pasta do código-fonte
cd 3-api-robo-juridico/api-robo-juridico

# Crie e ative o ambiente virtual
python -m venv .venv

# Windows (PowerShell)
.venv\Scripts\Activate.ps1

# Linux / macOS
source .venv/bin/activate

# Instale a API e o suporte a navegadores (evita bloqueios de WAF)
pip install -e ".[api,navegador]"

# Crie o arquivo de configuração
copy .env.exemplo .env   # Windows
# cp .env.exemplo .env   # Linux/macOS
```

---

## 💻 Como Usar

### 1. Pela Interface Web Local (Zero Dependências)
```bash
consulta-processos web
```
Abre automaticamente no navegador em `http://localhost:8765`. Permite buscar por número de processo ou consultar por OAB.

### 2. Pela Linha de Comando (CLI)
```bash
# Consultar processo
consulta-processos processo 5010754-07.2024.8.13.0625

# Processos de um advogado no período
consulta-processos oab 123456 MG --dias 30

# Publicações recentes de uma OAB no diário
consulta-processos publicacoes 123456 MG --dias 7
```

### 3. Pelo Servidor de API REST & Painel SaaS
```bash
# Crie o seu usuário administrador do painel
consulta-processos admin criar seu_email@empresa.com.br

# Inicie o servidor REST
consulta-processos api
```
* **API REST:** `http://127.0.0.1:8000/v1/...`
* **Swagger Interativo:** `http://127.0.0.1:8000/docs`
* **Painel Administrativo:** `http://127.0.0.1:8000/admin`

---

## 🔒 Segurança e Uso Ético (LGPD)

* **Fontes Oficiais:** A solução conecta-se apenas aos serviços abertos do CNJ, sem contorno de proteções ou quebra de captchas.
* **Segredo de Justiça:** Processos sigilosos não são indexados e não constam nas APIs públicas.
* **LGPD (Lei 13.709/2018):** Dados manifestamente públicos protegidos pelo Art. 7º, § 4º da LGPD. Tratamento restrito à finalidade legítima de acompanhamento processual.
* **Proteção de Segredos:** Chaves de API utilizam hashing com SHA-256; senhas de admin utilizam algoritmo `scrypt`.

---

## 📄 Licença

Distribuído sob a licença **MIT**. Veja [`3-api-robo-juridico/api-robo-juridico/LICENSE`](3-api-robo-juridico/api-robo-juridico/LICENSE) para mais informações.
