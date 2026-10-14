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
│ - Importador nativo .xlsx/.csv sem deps C    │     │    - Emissão de Laudo Certificado com QR Code     │
│ - Worker coletor Bacen (SGS) e IBGE (SIDRA)  │     │    - Selo de Autenticidade Digital (/validar/{id})│
│ - Sincronizador de 26+ tabelas de tribunais  │     ├───────────────────────────────────────────────────┤
│ - Documentação Swagger interativa (/docs)    │     │ B) PAINEL ADMINISTRATIVO (/admin)                 │
└──────────────────────┬───────────────────────┘     │    - Gestão de clientes, planos e MRR             │
                       │                             │    - Calculadora Judicial completa                │
                       │                             │    - Auditoria global do sistema                  │
                       │                             ├───────────────────────────────────────────────────┤
                       │                             │ C) PORTAL DO CLIENTE (/portal)                    │
                       │                             │    - Autoatendimento do desenvolvedor             │
                       │                             │    - Emissão/revogação de chaves rj_live_...      │
                       │                             │    - Calculadora Judicial para assinantes         │
                       │                             │    - Gráficos de cota e histórico de requisições  │
                       │                             └─────────────────┬─────────────────────────────────┘
                       │                                               │
                       ▼                                               ▼
         ┌───────────────────────────────────────────────────────────────────────────┐
         │                    /app/dados/api.db (SQLite WAL Mode)                    │
         │ - Chaves criptografadas (SHA-256)        - Registro de Consumo e Cotas    │
         │ - Clientes, Planos e Assinaturas         - Laudos Salvos e Leads Captados │
         └───────────────────────────────────────────────────────────────────────────┘
```

---

## ⚡ Novas Funcionalidades: Módulo de Cálculos Judiciais & Lead Magnet

O sistema agora conta com um motor completo de cálculos financeiros e judiciais com **100% de paridade com o DrCalc / EasyCalc**:

1. **Motor de Cálculo Judicial de Alta Precisão (Go):**
   - Cálculos em precisão monetária arbitrária (`shopspring/decimal`), sem perdas por ponto flutuante.
   - Totalmente adaptado à **Lei Federal nº 14.905/2024** (nova Taxa Legal de Juros do Código Civil baseada em $\max(0, \text{Selic} - \text{IPCA})$ pós-agosto/2024).
   - Juros moratórios convencionais com capitalização simples ou composta e cálculo diário *pro rata die*.
   - Encargos da fase de cumprimento de sentença: **Art. 523, §1º do CPC** (+10% de multa e +10% de honorários advocatícios).
   - Deflação negativa configurável (preservação do principal conforme jurisprudência pacificada).
   - **Mais de 26 indexadores oficiais integrados:**
     - **Estaduais:** TJSP, TJRJ, TJMG (Geral e Expurgada), TJRS, TJPR (Média IGP-DI/INPC e Precatórios IPCA-E), TJSC, TJES, TJCE, TJDF.
     - **Justiça Federal & Fazenda:** CJF Res. 267/2013 (Ações Condenatórias Gerais, Previdenciário INPC/IPCA-E, Fazenda Pública TR/Lei 11.960 e Desapropriação).
     - **Trabalhista:** TST Trabalhista (ADC 58 STF - IPCA-E / Selic).
     - **Índices Econômicos:** IPCA (IBGE), INPC (IBGE), SELIC Acumulada, Taxa Legal Bacen, IGP-M (FGV), IGP-DI (FGV), INCC (FGV), TR e Poupança.

2. **Robô Coletor Automático Diário e Mensal:**
   - Worker em background em Go consultando a API do **Banco Central do Brasil (SGS)** e do **IBGE (SIDRA)**.
   - Atualização automática dos índices com recálculo das tabelas práticas dos tribunais.
   - Endpoints manuais de trigger (`POST /v1/calculos/coletar`) e status (`GET /v1/calculos/coletar/status`).

3. **Importador Universal de Planilhas (Excel `.xlsx` / CSV / Colar):**
   - Decodificador nativo em Go (`/v1/calculos/importar-planilha`) sem dependências pesadas de bibliotecas C ou extensões PHP.
   - Suporte para upload de arquivos `.xlsx`, `.csv`, `.tsv` ou colagem direta (`Ctrl+C / Ctrl+V`) de linhas de planilha com dezenas de parcelas.

4. **Laudo Pericial Certificado com QR Code de Autenticidade (PDF Nativo):**
   - Emissão de documento formal em formato A4 (`/laudo/{uuid}/imprimir`) pronto para juntada em petições.
   - Hash criptográfico **SHA-256** gravado no banco que garante que o laudo não foi adulterado.
   - **QR Code dinâmico** em SVG apontando para a página pública de verificação (`/validar/{uuid}`), com selo digital verde atestando a autenticidade perante magistrados, advogados e peritos.

5. **Página Pública & Captação de Leads (Lead Magnet):**
   - Interface pública responsiva em `GET /` e `GET /calculadora` construída com Tailwind CSS.
   - Cálculos em tempo real via AJAX com exibição de KPIs e tabela discriminada de parcelas.
   - Modal de conversão de Leads que solicita Nome, E-mail, WhatsApp e OAB/Empresa antes da emissão do Laudo com QR Code.

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
- **Portal do Desenvolvedor / Cliente:** [http://localhost:8080/portal](http://localhost:8080/portal)
  - **Login:** `cliente@empresa.com.br` | **Senha:** `cliente123`
  - **Calculadora Judicial para Clientes:** [http://localhost:8080/portal/calculadora-judicial](http://localhost:8080/portal/calculadora-judicial)

---

## 📡 Novos Endpoints da API Go

| Método | Rota | Autenticação | Descrição |
|---|---|---|---|
| `GET` | `/v1/calculos/indices` | Sim (`X-API-Key`) | Lista todos os indexadores econômicos e tabelas dos tribunais disponíveis |
| `GET` | `/v1/calculos/indices/{codigo}/serie` | Sim (`X-API-Key`) | Retorna a série histórica completa mensal do índice |
| `POST` | `/v1/calculos/judicial` | Sim (`X-API-Key`) | Executa a liquidação judicial completa de débitos e parcelas |
| `POST` | `/v1/calculos/importar-planilha` | Sim (`X-API-Key`) | Importa arquivo Excel (.xlsx), CSV ou texto colado e devolve as parcelas estruturadas |
| `POST` | `/v1/calculos/coletar` | Sim (`X-API-Key`) | Dispara imediatamente a rotina de coleta no Bacen (SGS) e IBGE (SIDRA) |
| `GET` | `/v1/calculos/coletar/status` | Sim (`X-API-Key`) | Consulta o status e relatório da última coleta automática de índices |

---

## 🧪 Testes Automatizados

O ecossistema conta com baterias completas de testes unitários e de integração:

```bash
# 1. Testes do Motor em Go (incluindo engine de cálculo, importador e coletores)
cd api-juridica-go
go test ./...

# 2. Testes da Aplicação Web e Painel PHP (incluindo geração de laudos e validação de hash)
cd ../painel
vendor/bin/phpunit
```
