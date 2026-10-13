# Product Requirements Document (PRD): Robô Jurídico API

> **Identificador:** PRD-RJ-001  
> **Produto:** Robô Jurídico — API & Plataforma de Consulta Processual Unificada  
> **Versão:** 1.0.0  
> **Status:** Aprovado / Baseline  
> **Autor:** Engenharia de Produto & Arquitetura de Soluções  
> **Data de Emissão:** Outubro de 2026  

---

## 1. Visão do Produto & Proposta de Valor

### 1.1. Propósito
O **Robô Jurídico API** é uma solução de infraestrutura de dados e inteligência jurídica (*Data-as-a-Service / API-as-a-Product*) que permite a desenvolvedores, escritórios de advocacia, *legaltechs* e departamentos jurídicos consultar, monitorar e agregar dados de processos judiciais brasileiros diretamente das bases oficiais do Conselho Nacional de Justiça (CNJ), sem fricção, sem quebra de captchas e com alta previsibilidade de formato.

### 1.2. Proposta de Valor Única (UVP)
- **Eliminação de Scrapers Frágeis:** Não depende da leitura direta dos portais de dezenas de tribunais locais (PJe, Projudi, ESAJ, Eproc) sujeitos a captchas, bloqueios e alterações repentinas de interface.
- **Fusão Inteligente de Fontes Oficiais:** Resolve a omissão de partes/advogados na API pública do DataJud combinando-a em tempo real com o Diário de Justiça Eletrônico Nacional (DJEN).
- **SaaS "Out-of-the-Box":** O operador da API não precisa desenvolver infraestrutura de cobrança, rate limit ou autenticação: o sistema já vem com gestão de clientes B2B, geração de API keys, controle de cotas mensais atômicas e painel administrativo com cálculo de MRR.

---

## 2. Personas e Casos de Uso

### 2.1. Personas Primárias

| Persona | Perfil | Dores Principais | Como o Robô Jurídico Resolve |
|---|---|---|---|
| **P1: Desenvolvedor / Legaltech SaaS** | Engenheiro de software criando ERPs jurídicos ou automações. | APIs comerciais de birôs custam caro (R$ 0,50 a R$ 2,00 por consulta); scrapers próprios quebram semanalmente. | API REST padronizada, rápida, documentada com OpenAPI/Swagger e formato JSON estruturado por um custo marginal mínimo. |
| **P2: Advogado Individual / Gestor de Escritório** | Profissional de direito focado em contencioso civil, trabalhista ou tributário. | Medo constante de perda de prazos processuais; dificuldade de listar publicações diárias por OAB em múltiplos tribunais. | Endpoint específico para monitoramento de OAB no DJEN (`/v1/oab/{uf}/{numero}/publicacoes`) e interface web local sem necessidade de codificar. |
| **P3: Empreendedor de Tecnologia Jurídica** | Dono de produto vendendo acesso a consultas processuais para terceiros. | Alto custo de desenvolvimento de controle de cotas, faturamento, autenticação e dashboards. | O sistema já inclui o módulo `/admin`, permitindo cadastrar clientes, emitir chaves, travar cotas mensais e acompanhar a receita recorrente (MRR). |

### 2.2. Jornadas de Usuário Chave

```
[Cliente B2B / Desenvolvedor]
   │
   ├─► 1. Recebe a chave da API (rj_live_...) gerada pelo operador no painel.
   ├─► 2. Valida o número do processo em memória sem custo (/v1/cnj/{numero}).
   ├─► 3. Dispara consulta unificada (/v1/processos/{numero}).
   ├─► 4. Recebe JSON com metadados, capa, partes, advogados, movimentos e publicações.
   └─► 5. Acompanha consumo em tempo real pelos headers X-Cota-Restante e X-Cota-Usada.

[Advogado via Interface Web Local]
   │
   ├─► 1. Executa "consulta-processos web" no terminal.
   ├─► 2. Abre a tela em http://localhost:8765.
   ├─► 3. Informa o número do processo ou OAB + UF.
   └─► 4. Visualiza na tela o teor integral da publicação e a linha do tempo dos autos.
```

---

## 3. Requisitos Funcionais (RF)

### 3.1. Consulta e Normalização de Processos
- **RF-01 (Validação de Número CNJ):** O sistema deve validar a estrutura e o dígito verificador (DV) do número CNJ conforme Resolução CNJ 65/2008 utilizando o algoritmo Módulo 97. Números inválidos devem retornar HTTP 422 sem realizar chamadas externas e sem debitar cota.
- **RF-02 (Roteamento de Tribunais):** O sistema deve extrair o segmento de justiça e o código do tribunal a partir do número do processo e direcionar a consulta diretamente ao índice Elasticsearch correspondente do DataJud (ex.: `api_publica_tjmg`, `api_publica_trt3`).
- **RF-03 (Fusão de Dados DataJud + DJEN):** Na consulta de processo completa (`modo=completo`), o sistema deve consultar simultaneamente o DataJud e o DJEN, mesclar os resultados, extrair partes e advogados do diário e deduplicar movimentações internas.
- **RF-04 (Consulta Rápida pelo Diário):** O sistema deve oferecer o parâmetro `modo=rapido` que ignora a consulta ao DataJud e pesquisa unicamente no DJEN, retornando a resposta em menos de 2 segundos.
- **RF-05 (Consulta de Existência Prévia):** Disponibilizar endpoint (`/v1/processos/{numero}/existe`) de custo zero para checagem rápida se um processo recém-distribuído já consta em alguma das bases do CNJ.
- **RF-06 (Decodificação Sintática):** Disponibilizar endpoint (`/v1/cnj/{numero}`) sem custo e sem chamadas externas para extração sintática do tribunal, ano, segmento e comarca a partir do número digitado.

### 3.2. Pesquisa e Monitoramento por Advogado (OAB)
- **RF-07 (Processos Agrupados por OAB):** O sistema deve permitir listar processos nos quais a OAB especificada teve publicações nos últimos *N* dias (1 a 365), agrupando as publicações por número de processo e suportando filtro opcional por sigla de tribunal.
- **RF-08 (Publicações da OAB no Período):** O sistema deve disponibilizar endpoint para retorno sequencial de todas as publicações da OAB no DJEN dentro de uma janela temporal (1 a 90 dias), com paginação de 100 itens por página, retornando o teor do ato sanitizado (sem HTML).

### 3.3. Gestão de Contas, Cotas e Faturamento
- **RF-09 (Autenticação por API Key):** Toda requisição nas rotas de dados deve exigir o cabeçalho `X-API-Key`. A chave deve ser validada contra os hashes SHA-256 armazenados no banco de dados ou contra as chaves internas no arquivo `.env`.
- **RF-10 (Controle Atômico de Cotas):** O sistema deve manter cota mensal por cliente. A cota deve ser reservada atomicamente no início da chamada e estornada caso a requisição resulte em erro de cliente ou de servidor (4xx/5xx).
- **RF-11 (Renovação Mensal Automática):** As cotas de consultas devem reiniciar automaticamente no primeiro segundo do dia 1º de cada mês, considerando o fuso horário oficial de Brasília (`America/Sao_Paulo`).
- **RF-12 (Consulta de Saldo pelo Cliente):** O endpoint `/v1/conta` e os cabeçalhos de resposta `X-Cota-Limite`, `X-Cota-Usada` e `X-Cota-Restante` devem informar a situação do plano do cliente em tempo real.
- **RF-13 (Rate Limiting por Cliente):** Cada cliente deve ter um limite de requisições por minuto (RPM) definido no seu plano. Ao atingir o teto, a API deve responder HTTP 429 com cabeçalho `Retry-After`.

### 3.4. Painel de Administração e Auditoria (/admin)
- **RF-14 (Gestão de Clientes e Planos):** O administrador deve conseguir cadastrar, editar, suspender e reativar clientes, ajustando limites mensais de consultas, RPM e mensalidade contratada.
- **RF-15 (Emissão e Revogação de Chaves):** O painel deve permitir gerar novas chaves (com prefixo `rj_live_`), exibindo o segredo completo apenas uma vez no momento da criação, e revogar chaves existentes com efeito imediato.
- **RF-16 (Trilha de Auditoria e Métricas):** O painel deve apresentar gráficos de consumo dos últimos 30 dias, MRR total consolidado, clientes com maior consumo e log imutável de eventos com carimbo de data/hora, administrador responsável e IP de origem.

---

## 4. Requisitos Não-Funcionais (RNF)

| Identificador | Categoria | Descrição do Requisito |
|---|---|---|
| **RNF-01** | **Segurança** | Chaves de API de clientes nunca devem ser persistidas em texto plano. Deve-se armazenar apenas o hash SHA-256. |
| **RNF-02** | **Segurança** | Senhas de administradores devem ser criptografadas utilizando `scrypt` com parâmetros de alto custo computacional ($N=2^{14}, r=8, p=1$). |
| **RNF-03** | **Segurança** | O painel web deve empregar cookies com flags `HttpOnly`, `SameSite=Strict`, `Secure` e validação de tokens anti-CSRF em métodos mutáveis (`POST`, `PATCH`, `DELETE`). |
| **RNF-04** | **Segurança** | Acesso ao painel `/admin` deve poder ser restrito por lista de blocos CIDR (`PAINEL_IPS_PERMITIDOS`). Requisições fora do escopo devem receber HTTP 404. |
| **RNF-05** | **Performance & Concorrência** | O cache em memória (`CacheTTL`) deve impedir *thundering herd* através de travas por chave individualizada (`single-flight locking`). |
| **RNF-06** | **Performance & Latência** | Consultas em `modo=rapido` (DJEN) devem responder com P95 abaixo de 2,5 segundos em condições normais de rede. |
| **RNF-07** | **Resiliência de Rede** | O cliente HTTP deve suportar impersonate TLS de navegadores reais via `curl_cffi` para evitar bloqueios de WAF. |
| **RNF-08** | **Rotatividade de Proxies** | Em caso de uso de pool de proxies residenciais, o sistema deve despachar tentativas em paralelo por lotes, cancelando as demais assim que uma resposta válida for recebida. |
| **RNF-09** | **Integridade de Dados** | O banco de dados SQLite deve operar com WAL (`Write-Ahead Logging`) e isolamento de permissão de arquivo `0600` no sistema operacional. |
| **RNF-10** | **Compatibilidade** | A aplicação deve rodar em Python 3.10 ou superior, funcionando nativamente em Linux, macOS e Windows. |
| **RNF-11** | **Conformidade LGPD** | A API não deve capturar ou indexar processos que tramitam em segredo de justiça e não deve persistir dados sensíveis além dos registros operacionais de auditoria. |
| **RNF-12** | **Contrato de API** | Todos os erros devem retornar estrutura JSON uniforme contendo `ok: false`, `erro.codigo`, `erro.mensagem` e `id_requisicao`. |

---

## 5. Modelo de Negócio, Planos e Monetização

O Robô Jurídico foi projetado para permitir que o adquirente monetize o acesso à API vendendo planos de assinatura mensal B2B para escritórios ou desenvolvedores:

### 5.1. Matriz de Planos Pré-Configurados

| Plano | Consultas / Mês | Limite de Requisições | Mensalidade Sugerida | Custo Médio / Consulta |
|---|---|---|---|---|
| **Start** | 2.000 consultas | 30 req / minuto | **R$ 197,00** | R$ 0,098 |
| **Pro** | 10.000 consultas | 60 req / minuto | **R$ 497,00** | R$ 0,049 |
| **Revenda / Enterprise** | 30.000 consultas | 120 req / minuto | **R$ 997,00** | R$ 0,033 |
| **Interno / Ilimitado** | Ilimitado | Customizado (60 pad.) | Gratuito | N/A |

### 5.2. Regras de Cobrança e Consumo
1. **Cobram 1 consulta da cota:**
   - `GET /v1/processos/{numero}` (tanto em `modo=completo` quanto em `modo=rapido`).
   - `GET /v1/oab/{uf}/{numero}/processos`.
   - `GET /v1/oab/{uf}/{numero}/publicacoes` (independente do número de páginas buscadas).
2. **Consultas Gratuitas (Custo Zero):**
   - `GET /v1/processos/{numero}/existe`.
   - `GET /v1/cnj/{numero}`.
   - `GET /v1/conta`.
   - `GET /saude`.
3. **Política de Isenção em Erros:**
   - Requisições que resultam em códigos de erro (401, 402, 403, 404, 422, 429, 500, 502, 503) **não debitam saldo da cota mensal**.
4. **Cache:**
   - Consultas repetidas dentro do prazo de validade do cache (6h para DataJud, 30m para diário) respondem em milissegundos, mas **são contabilizadas normalmente na cota**.

---

## 6. Métricas de Sucesso do Produto (KPIs)

- **MRR (Monthly Recurring Revenue):** Acompanhamento da soma das mensalidades de clientes ativos no painel `/admin`.
- **Taxa de Disponibilidade (SLA):** Superior a 99,5% para o servidor REST local/produção.
- **Taxa de Cache Hit:** Proporção de consultas repetidas resolvidas em milissegundos pelo cache em memória (alvo > 30% em escritórios com múltiplos operadores).
- **Taxa de Estorno por Falha Externa:** Menor que 5% de erros 502/503 decorrentes de instabilidades momentâneas no CNJ.
- **P95 Latência Modo Rápido:** < 2,0 segundos.

---

## 7. Roadmap Evolutivo do Produto

```
v0.1.0 (MVP Atual)
  ├── Dual sourcing (DataJud + DJEN)
  ├── Painel Admin com SQLite WAL
  ├── CLI, Web UI local e API REST
  └── Autenticação X-API-Key e Cotas Atômicas

v0.2.0 (Planejado - Automações & Webhooks)
  ├── Webhooks para notificação assíncrona de novas publicações
  ├── Disparo automático de alertas via WhatsApp/Email (ex: integração Evolution API / Baileys)
  ├── Exportação de relatórios gerenciais em CSV/PDF diretamente pelo painel
  └── Suporte opcional a backend de Cache e Rate Limit em Redis

v1.0.0 (Enterprise & IA)
  ├── Módulo de extração de prazos fatais via IA Generativa sobre o texto do ato
  ├── Multi-tenant escalável com cluster Kubernetes e PostgreSQL
  └── Integração de busca semântica por termos e jurisprudência vinculada
```
