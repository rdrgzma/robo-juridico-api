# ROBÔ JURÍDICO · ARQUITETURA GO + PHP FILAMENT
# Comece por aqui

> **Três passos para sair do zero e ter a API de consulta de processos em Go e os Painéis em PHP/Filament rodando do jeito que você precisa — tanto localmente para testes quanto em uma VPS de produção, com ou sem Docker, com a ajuda da sua IA.**

---

## 📦 Neste Pacote: O que você recebeu

Esta versão (`4-api-robo-juridico-php-go`) representa a evolução de engenharia do Robô Jurídico: unimos a **velocidade brutal e concorrência nativa do Go** para o motor de consulta pública às fontes do CNJ com a **produtividade e elegância do Laravel 13 + Filament v5** para os painéis de gestão e autoatendimento.

| Arquivo / Diretório | Para que serve |
| :--- | :--- |
| **`4-api-robo-juridico-php-go/1-COMECE-AQUI.md`** | **Este guia.** Visão executiva em 3 passos para iniciar os testes e deploys imediatamente. |
| **`4-api-robo-juridico-php-go/2-WORKBOOK-API-ROBO-JURIDICO.md`** | **O manual definitivo.** Guia completo com 12 capítulos: pré-requisitos, passo a passo local (Docker / Sem Docker), deploy em VPS (Docker / Bare-metal), referência completa de endpoints, erros, exemplos em código, proxies e exercícios práticos. |
| **`4-api-robo-juridico-php-go/api-juridica-go/`** | **Motor da API REST em Go (Porta 8000).** Alta performance, compilação estática (`CGO_ENABLED=0`), validação Módulo 97 com `big.Int`, concorrência simultânea DataJud + DJEN, cache anti-thundering herd e controle atômico de cotas. |
| **`4-api-robo-juridico-php-go/painel/`** | **Painéis Web em PHP 8.4 + Filament v5 (Porta 8080).** Arquitetura Multi-Panel contendo o Painel Administrativo (`/admin`) e o Portal de Autoatendimento do Cliente/Desenvolvedor (`/portal`). |
| **`4-api-robo-juridico-php-go/docker-compose.yml`** | **Orquestração completa pronta.** Sobe a API Go e os Painéis PHP integrados em segundos, compartilhando o volume do banco SQLite em modo WAL (`api.db`). |
| **`4-api-robo-juridico-php-go/api.http`** | **Coleção de testes prontos.** Arquivo para executar requisições diretamente pelo VS Code (extensão REST Client). |

---

## 🧭 Os 3 Passos para Sair do Zero

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                                   FLUXO DE EXECUÇÃO                                    │
│                                                                                        │
│   01. ENTENDA A ARQUITETURA   ──►   02. ESCOLHA SEU DEPLOY   ──►   03. LEVE PARA A IA  │
│   (Go + PHP Multi-Panel)            (Local vs VPS / Docker)        (Integre ao seu CRM)│
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

### Passo 1: Entenda a Arquitetura Dual

Em vez de sobrecarregar um único framework, separamos o trabalho cirurgicamente:

1. **A API em Go (`:8000`)** atende as requisições que exigem altíssima performance. Ela valida números de processo em nanossegundos, consulta o DataJud e o DJEN concorrentemente em goroutines paralelas e devolve o JSON unificado em sub-segundos.
2. **O Painel em PHP/Filament (`:8080`)** oferece interface gráfica corporativa:
   - **`/admin`**: Gestão da sua empresa, faturamento (MRR), clientes, cotas e auditoria global.
   - **`/portal`**: Portal white-label para o seu cliente/desenvolvedor emitir e revogar chaves `rj_live_...`, consultar gráficos de consumo e ver logs próprios.
3. **O Elo de Ligação**: Um único arquivo SQLite (`dados/api.db`) configurado em **modo WAL (Write-Ahead Logging)** com concorrência segura. O PHP escreve os clientes e planos; o Go lê as chaves criptografadas em SHA-256 e audita os consumos.

---

### Passo 2: Escolha seu Caminho de Deploy

Consulte a matriz abaixo para decidir como rodar agora:

```
                                  SUA FINALIDADE HOJE?
                                            │
                    ┌───────────────────────┴───────────────────────┐
                    ▼                                               ▼
          [ Testes Locais ]                                 [ VPS em Produção ]
          (No seu computador)                               (Servidor na Nuvem)
                    │                                               │
          ┌─────────┴─────────┐                           ┌─────────┴─────────┐
          ▼                   ▼                           ▼                   ▼
    COM DOCKER          SEM DOCKER                  COM DOCKER          SEM DOCKER
(1 comando no terminal) (Go + PHP nativos)      (Isolado + Caddy SSL) (Systemd + PHP-FPM)
   [Opção 2.1]         [Opção 2.2]                 [Opção 2.3]         [Opção 2.4]
```

#### Opção 2.1: Local com Docker Compose (Mais Rápido para Testar)
Abra o terminal na pasta `4-api-robo-juridico-php-go/` e execute:
```bash
docker compose up -d --build
```
Em menos de 1 minuto, acesse no navegador:
* **API & Swagger UI:** [http://localhost:8000/docs](http://localhost:8000/docs)
* **Painel Admin:** [http://localhost:8080/admin](http://localhost:8080/admin) *(Login: `admin@empresa.com.br` | Senha: `admin123`)*
* **Portal do Cliente:** [http://localhost:8080/portal](http://localhost:8080/portal) *(Login: `cliente@empresa.com.br` | Senha: `cliente123`)*

#### Opção 2.2: Local Sem Docker (Bare-Metal no Windows, Mac ou Linux)
* **No terminal 1 (API Go):**
  ```bash
  cd 4-api-robo-juridico-php-go/api-juridica-go
  go run ./cmd/api
  ```
* **No terminal 2 (Painel PHP):**
  ```bash
  cd 4-api-robo-juridico-php-go/painel
  composer install
  cp .env.example .env
  php artisan key:generate
  php artisan migrate --seed
  php artisan serve --port=8080
  ```

#### Opção 2.3: VPS em Produção com Docker Compose
* Clone o repositório na VPS Ubuntu/Debian (`/opt/robo-juridico`).
* Suba a stack com `docker compose -f docker-compose.yml up -d`.
* Aponte o **Caddy Server** ou **Nginx** para as portas 8000 e 8080 com terminação SSL (HTTPS automático).

#### Opção 2.4: VPS em Produção Sem Docker (Performance Máxima)
* Compile o binário estático Go (`CGO_ENABLED=0 go build -o api ./cmd/api`) e crie um serviço no **Systemd** (`robo-juridico.service`).
* Instale o **PHP 8.4-FPM** e **Nginx** para servir o diretório `painel/public`.
* Garanta permissões de leitura/escrita no banco SQLite `dados/api.db`.

*(O passo a passo minucioso de cada opção com scripts e configurações prontas está no **Capítulo 03** e **Capítulo 10** do [4-api-robo-juridico-php-go/2-WORKBOOK-API-ROBO-JURIDICO.md](4-api-robo-juridico-php-go/2-WORKBOOK-API-ROBO-JURIDICO.md)).*

---

### Passo 3: Leve para a sua IA

Com a API rodando, você pode usar qualquer IA generativa (Claude, ChatGPT, Gemini, Copilot, Cursor) para criar aplicações sobre o Robô Jurídico (robôs de WhatsApp, integração com Zapier/n8n, painéis jurídicos ou robôs de peticionamento).

Copie e cole o prompt abaixo na sua IA de preferência:

```markdown
Anexei a documentação técnica da API do Robô Jurídico (Workbook da API Go + PHP).
Por favor, leia atentamente as rotas e contratos de retorno antes de responder.

ESTOU CRIANDO O SEGUINTE PROJETO:
[Descreva aqui o que você deseja construir: ex:
 "Um fluxo em Python/n8n que roda toda manhã às 07:00, consulta as intimações da 
  OAB/SP 123456 no endpoint /v1/oab/SP/123456/publicacoes?dias=1 e envia um 
  resumo formatado no grupo do escritório no WhatsApp/Telegram."]

REGRAS DE IMPLEMENTAÇÃO:
1. Utilize a API como ela já está (respeitando o cabeçalho X-API-Key e tratamento de erros).
2. Trate adequadamente os códigos HTTP 402 (cota esgotada) e 429 (rate limit com Retry-After).
3. Não invente novos endpoints: use estritamente os descritos no Workbook.
4. Forneça o código limpo, comentado e as instruções passo a passo para execução.
```

---

## 🔒 Regra de Ouro da Segurança

> [!CAUTION]
> **NUNCA envie suas chaves de API reais nem arquivos `.env` para nenhuma inteligência artificial.**
> 
> O repositório vem configurado exclusivamente com arquivos de modelo (`.env.example` e `.env.exemplo`). Mantenha suas chaves `rj_live_...` e credenciais salvas apenas localmente ou no cofre de segredos da sua VPS. Se a IA solicitar uma chave de acesso para os exemplos de código, utilize placeholders como `rj_live_sua_chave_aqui`.

---

## 🚀 Próximo Passo

Abra agora o arquivo **[4-api-robo-juridico-php-go/2-WORKBOOK-API-ROBO-JURIDICO.md](4-api-robo-juridico-php-go/2-WORKBOOK-API-ROBO-JURIDICO.md)** para acompanhar o manual definitivo completo, explorar os endpoints com exemplos reais e realizar os exercícios práticos passo a passo.
