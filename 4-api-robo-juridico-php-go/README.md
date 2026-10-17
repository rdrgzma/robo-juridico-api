# Stack Integrada: API Robô Jurídico (Go) + Painel Administrativo, Portal do Cliente & Módulo de Cálculos Judiciais (PHP/Filament)

Solução corporativa completa para consulta automatizada de processos jurídicos e **liquidação de débitos judiciais com paridade DrCalc**, combinando a altíssima performance de um microsserviço nativo em **Go** com a produtividade e elegância do **Laravel 11 + Filament v3** para gestão administrativa, autoatendimento e captação de leads.

---

## 🏛️ Visão Geral da Arquitetura

```
                     ┌─────────────────────────────────────────────────────────────┐
                     │                  Clientes / Sistemas Externos /             │
                     │          Público Geral (Calculadora & Lead Magnet)          │
                     └──────────────────────────────┬──────────────────────────────┘
                                                    │
                 ┌──────────────────────────────────┴──────────────────────────────────┐
                 │ Requisições HTTP (X-API-Key / REST)                                 │ Navegação Web / AJAX
                 ▼                                                                     ▼
┌──────────────────────────────────────────────┐     ┌───────────────────────────────────────────────────┐
│         api-juridica-go (Porta 8000)         │     │         painel (Porta 8080 - Multi-Panel)         │
│ - Go 1.22+ nativo (Sem CGO, ultra-rápido)    │     │ - PHP 8.2+ com Laravel 11 e Filament v3           │
│ - Validação DV CNJ (Módulo 97 com big.Int)   │     ├───────────────────────────────────────────────────┤
│ - Consulta concorrente DataJud e DJEN        │     │ A) PÁGINA PÚBLICA & LEAD MAGNET (GET /)           │
│ - Autenticação SHA-256 e cotas mensais       │     │    - Calculadora Judicial online gratuita         │
│ - Motor de Cálculos Judiciais (Paridade      │     │    - Importador de planilhas (.xlsx, .csv, colar) │
│   DrCalc, Lei 14.905/2024, juros e multas)   │     │    - Modal de captura de Lead (Nome, WhatsApp, OAB)│
│ - Importador e Exportador nativo .xlsx/.csv  │     │    - 1-Click Auto-Fill por número CNJ             │
│ - Revisão de Contratos (Price, SAC e Gauss)  │     │    - Emissão de Laudo Certificado com QR Code SVG │
│ - Calculadora de Prazos Fatais (CPC / CLT)   │     │    - Selo de Autenticidade Digital (/validar/{id})│
│ - Monitoramento Push Webhooks (HMAC-SHA256)  │     ├───────────────────────────────────────────────────┤
│ - Worker coletor Bacen (SGS) e IBGE (SIDRA)  │     │ B) PAINEL ADMINISTRATIVO (/admin)                 │
│ - Sincronizador de 27 tabelas de tribunais   │     │    - Gestão de clientes, planos e MRR             │
│ - Documentação Swagger interativa (/docs)    │     │    - CRM de Leads & Laudos (WhatsApp + 1-Click)   │
└──────────────────────┬───────────────────────┘     │    - Calculadora Judicial completa                │
                       │                             │    - Auditoria global do sistema                  │
                       │                             ├───────────────────────────────────────────────────┤
                       │                             │ C) PORTAL DO CLIENTE (/portal)                    │
                       │                             │    - Autoatendimento do desenvolvedor             │
                       │                             │    - Emissão/revogação de chaves rj_live_...      │
                       │                             │    - Calculadora Judicial + Histórico Salvo       │
                       │                             │    - "Atualizar para Hoje (1-Click)"              │
                       │                             │    - Gráficos de cota e histórico de requisições  │
                       │                             └─────────────────┬─────────────────────────────────┘
                       │                                               │
                       ▼                                               ▼
         ┌───────────────────────────────────────────────────────────────────────────┐
         │                    /app/dados/api.db (SQLite WAL Mode)                    │
         │ - Chaves criptografadas (SHA-256)        - Registro de Consumo e Cotas    │
         │ - Clientes, Planos e Assinaturas         - Laudos Salvos e Funil CRM      │
         └───────────────────────────────────────────────────────────────────────────┘
```

---

## ⚡ Funcionalidades do Ecossistema: Motor de Cálculos, Prazos, Revisão Bancária, Webhooks & CRM

O sistema conta com um motor completo de inteligência processual, cálculos financeiros e judiciais com **100% de paridade com o DrCalc / PJe-Calc**:

1. **Motor de Cálculo Judicial de Alta Precisão (Go):**
   - Cálculos em precisão monetária arbitrária (`shopspring/decimal`), sem perdas por ponto flutuante.
   - Totalmente adaptado à **Lei Federal nº 14.905/2024** (nova Taxa Legal de Juros do Código Civil baseada em $\max(0, \text{Selic} - \text{IPCA})$ pós-agosto/2024).
   - **Juros Moratórios e Compensatórios/Remuneratórios** com capitalização simples ou composta e cálculo diário *pro rata die*.
   - **Correção Monetária *Pro Rata Die* Linear** opcional (`aplicar_prorata_correcao`) e deflação negativa configurável.
   - **Imputação de Pagamentos Parciais pelo Art. 354 do Código Civil** (`imputar_art_354_cc`), abatendo primeiro dos juros vencidos na data do pagamento e o excedente sobre o principal.
   - Encargos da fase de cumprimento de sentença: **Art. 523, §1º do CPC** (+10% de multa e +10% de honorários advocatícios) e honorários sucumbenciais (Art. 85 CPC).
   - **27 indexadores oficiais integrados com carga histórica offline (2023–2026) e coleta automática de 60 meses:**
     - **Estaduais:** TJSP, TJRJ, TJMG (Geral e Expurgada), TJRS, TJPR (Média IGP-DI/INPC e Precatórios IPCA-E), TJSC, TJES, TJCE, TJDF.
     - **Justiça Federal & Fazenda:** CJF Res. 267/2013 (Ações Condenatórias Gerais, Previdenciário INPC/IPCA-E, Fazenda Pública TR/Lei 11.960 e Desapropriação).
     - **Trabalhista:** TST Trabalhista (ADC 58 STF - IPCA-E / Selic).
     - **Índices Econômicos:** IPCA (IBGE), IPCA-15 (IBGE), INPC (IBGE), SELIC Acumulada, CDI, Taxa Legal Bacen, IGP-M (FGV), IGP-DI (FGV), INCC (FGV), TR e Poupança.

2. **Autopreenchimento Judicial "1-Click Liquidation" & Consulta Unificada Inteligente (DataJud + DJEN):**
   - Ao informar o número do processo CNJ na Calculadora Pública ou nos Painéis (`/admin` e `/portal`), o sistema decodifica o tribunal, seleciona automaticamente o indexador correspondente (ex.: `TJMG`, `TJSP`, `TRF_GERAL`, `TST_IPCAE`) e consulta o DataJud/DJEN para preencher automaticamente o **Valor da Causa**, **Data de Ajuizamento** e **Partes (Autor × Réu)**.
   - **Fast-Fallback & Cache Background:** Consultas em modo `completo` disparam goroutines paralelas no DataJud e no DJEN. Se o DJEN responder rapidamente e o DataJud do tribunal estiver lento (> 4,5s), a API retorna imediatamente os dados consolidados com aviso transparente e conclui a sincronização profunda do DataJud em background no cache.
   - **Links Oficiais Separados (`link` vs `link_peca`):**
     - `link`: URL oficial de consulta pública dos autos no respectivo Tribunal (PJe, eproc, e-SAJ).
     - `link_peca`: Link direto para o documento/peça processual ou **Certidão Oficial em PDF do CNJ** (`https://comunicaapi.pje.jus.br/api/v1/comunicacao/{hash}/certidao`), preservando IDs longos sem perda de precisão em ponto flutuante.
   - **Correlação Inteligente DataJud ↔ DJEN:** Movimentações publicáveis do DataJud (`sentenca`, `decisao`, `despacho`, `intimacao`, `publicacao_dje`) recebem automaticamente o `link_peca` da certidão do DJEN quando publicadas na mesma janela temporal (até 72h), além de manterem o `link` de consulta ao tribunal.
   - **Sanitização Avançada de Partes e Textos:** Filtragem automática de cabeçalhos de secretaria, assinaturas de magistrados, varas e sufixos societários órfãos (`LTDA`, `S/A`) nas publicações do DJEN, e decodificação de *mojibake* e complementos tabelados no DataJud.

3. **Módulo de Revisão de Contratos Bancários (`PRICE`, `SAC` e `GAUSS`):**
   - Endpoint `POST /v1/calculos/financiamento` que compara o sistema contratado (Tabela Price, SAC ou Gauss) com a taxa média de mercado ou método linear de Gauss (juros simples sem anatocismo), apurando diferença mensal, saldo devedor recalculado e **repetição do indébito simples e em dobro (Art. 42, parágrafo único do CDC)**.

4. **Calculadora de Prazos Processuais Fatais (CPC/2015 & CLT):**
   - Endpoints `POST /v1/prazos/calcular` e `GET /v1/prazos/feriados` que calculam **Data de Disponibilização no DJe**, **Data de Publicação** (Art. 224, §2º CPC), **Início da Contagem** (Art. 224, §3º CPC) e **Vencimento Fatal** em dias úteis ou corridos, considerando feriados nacionais, móveis (Páscoa, Carnaval, Corpus Christi), **recesso forense de 20/dez a 20/jan (Art. 220 CPC)** e **prazo em dobro** (Art. 180/183/186 CPC).

5. **Monitoramento Push via Webhooks Assinados (`HMAC-SHA256`):**
   - Endpoints `GET/POST/DELETE /v1/webhooks` e `POST /v1/webhooks/testar` para cadastrar assinaturas de monitoramento contínuo de processos CNJ ou inscrições OAB, entregando notificações HTTP assinadas com `X-Webhook-Signature: sha256=<hex>`.

6. **Importador e Exportador Universal de Planilhas (Excel `.xlsx` e `.csv`):**
   - Importação nativa (`/v1/calculos/importar-planilha`) e exportação nativa (`/v1/calculos/exportar-planilha`) de memórias de cálculo em `.xlsx` (OpenXML) e `.csv` (UTF-8 com BOM) sem dependências C.

7. **Laudo Pericial Certificado com QR Code SVG Nativo & CRM de Leads:**
   - Emissão de Laudo A4 (`/laudo/{uuid}/imprimir`) com Hash **SHA-256** e **QR Code SVG matricial nativo** (`App\Services\QrCodeGenerator`) apontando para `/validar/{uuid}`.
   - **Funil CRM no Painel Admin (`/admin/calculos-salvos`):** Gestão de leads captados com status (*Novo*, *Contatado*, *Qualificado*, *Convertido*, *Perdido*), link direto para WhatsApp e botão **"Converter em Cliente (1-Click)"**.
   - **Histórico de Cálculos no Portal do Assinante (`/portal/meus-calculos`):** Armazenamento dos cálculos do escritório com botão **"Atualizar para Hoje (1-Click)"**.

---

## ⚡ Como Rodar com Docker Compose

Suba toda a infraestrutura com apenas um comando:

```bash
docker compose up -d --build
```

### Serviços e Rotas Disponíveis:
- **Página Pública / Calculadora Lead Magnet:** [http://localhost:8080](http://localhost:8080) ou [http://localhost:8080/calculadora](http://localhost:8080/calculadora)
- **API Jurídica e Motor de Cálculos (Go):** [http://localhost:8000](http://localhost:8000)
- **Documentação Swagger UI:** [http://localhost:8000/docs](http://localhost:8000/docs)
- **Especificação OpenAPI JSON:** [http://localhost:8000/openapi.json](http://localhost:8000/openapi.json)
- **Painel do Administrador:** [http://localhost:8080/admin](http://localhost:8080/admin)
  - **Login:** `admin@empresa.com.br` | **Senha:** `admin123`
  - **Calculadora Judicial Interna:** [http://localhost:8080/admin/calculadora-judicial](http://localhost:8080/admin/calculadora-judicial)
  - **CRM de Laudos & Leads Captados:** [http://localhost:8080/admin/calculos-salvos](http://localhost:8080/admin/calculos-salvos)
- **Portal do Desenvolvedor / Cliente:** [http://localhost:8080/portal](http://localhost:8080/portal)
  - **Login:** `cliente@empresa.com.br` | **Senha:** `cliente123`
  - **Calculadora Judicial para Clientes:** [http://localhost:8080/portal/calculadora-judicial](http://localhost:8080/portal/calculadora-judicial)
  - **Meus Cálculos Salvos:** [http://localhost:8080/portal/meus-calculos](http://localhost:8080/portal/meus-calculos)

---

## 📡 Catálogo de Endpoints da API Go

| Método | Rota | Autenticação | Custo de Cota | Descrição |
|---|---|---|---|---|
| `GET` | `/saude` | Não | Não | Healthcheck e versão da API |
| `GET` | `/v1/cnj/{numero}` | Sim (`X-API-Key`) | Não | Valida e decodifica o número CNJ (Módulo 97 + Tribunal) |
| `GET` | `/v1/processos/{numero}/existe` | Sim (`X-API-Key`) | Não | Checagem rápida de existência no DataJud ou DJEN |
| `GET` | `/v1/processos/{numero}` | Sim (`X-API-Key`) | **Sim (1)** | Consulta completa consolidada (DataJud + DJEN) |
| `POST` | `/v1/processos/consultar` | Sim (`X-API-Key`) | **Sim (1)** | Consulta de processo via payload JSON |
| `GET` | `/v1/oab/{uf}/{numero}/processos` | Sim (`X-API-Key`) | **Sim (1)** | Processos com publicações da OAB no período |
| `GET` | `/v1/oab/{uf}/{numero}/publicacoes` | Sim (`X-API-Key`) | **Sim (1)** | Publicações do DJEN na íntegra por OAB |
| `GET` | `/v1/conta` | Sim (`X-API-Key`) | Não | Extrato de cota mensal usada, restante e renovação |
| `GET` | `/v1/calculos/indices` | Sim (`X-API-Key`) | Não | Lista todos os 27 indexadores econômicos e tabelas de tribunais |
| `GET` | `/v1/calculos/indices/{codigo}/serie` | Sim (`X-API-Key`) | Não | Retorna a série histórica mensal completa do índice |
| `POST` | `/v1/calculos/judicial` | Sim (`X-API-Key`) | Não | Executa liquidação judicial completa (Lei 14.905, Art. 354 CC, Art. 523 CPC) |
| `POST` | `/v1/calculos/financiamento` | Sim (`X-API-Key`) | Não | Revisão de contratos bancários (`PRICE`, `SAC`, `GAUSS` e Art. 42 CDC) |
| `POST` | `/v1/calculos/importar-planilha` | Sim (`X-API-Key`) | Não | Importa arquivo Excel (`.xlsx`), CSV ou texto colado |
| `POST` | `/v1/calculos/exportar-planilha` | Sim (`X-API-Key`) | Não | Exporta memória de cálculo liquidada em `.xlsx` ou `.csv` |
| `POST` | `/v1/calculos/coletar` | Sim (`X-API-Key`) | Não | Dispara imediatamente a coleta no Bacen (SGS) e IBGE (SIDRA) |
| `GET` | `/v1/calculos/coletar/status` | Sim (`X-API-Key`) | Não | Consulta o status e relatório da última coleta automática |
| `POST` | `/v1/prazos/calcular` | Sim (`X-API-Key`) | Não | Calcula prazos processuais fatais (CPC/2015 Art. 224 e CLT Art. 775) |
| `GET` | `/v1/prazos/feriados` | Sim (`X-API-Key`) | Não | Lista feriados nacionais, móveis e forenses do ano |
| `GET` | `/v1/webhooks` | Sim (`X-API-Key`) | Não | Lista webhooks de monitoramento push cadastrados |
| `POST` | `/v1/webhooks` | Sim (`X-API-Key`) | Não | Registra novo webhook de monitoramento de processo ou OAB |
| `DELETE` | `/v1/webhooks/{id}` | Sim (`X-API-Key`) | Não | Remove um webhook cadastrado |
| `POST` | `/v1/webhooks/testar` | Sim (`X-API-Key`) | Não | Simula disparo de evento push com assinatura `HMAC-SHA256` |

---

## 🧪 Testes Automatizados

O ecossistema conta com baterias completas de testes unitários e de integração:

```bash
# 1. Testes do Motor em Go (incluindo engine de cálculo, financiamento, prazos, webhooks, importador, exportador e coletores)
cd api-juridica-go
go test -v ./...

# 2. Testes da Aplicação Web e Painel PHP (incluindo geração de laudos, QR Code SVG nativo, auto-fill CNJ e validação de hash)
cd ../painel
php vendor/bin/phpunit
```

