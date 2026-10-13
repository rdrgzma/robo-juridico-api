# Análise de Viabilidade Técnica e Arquitetural: Refatoração em PHP, Go e Rust

> **Identificador:** ESTUDO-REFATORACAO-001  
> **Escopo:** Avaliação da reescrita da API e componentes do Robô Jurídico em PHP (projeto completo) e em Go / Rust (camada de API de alto desempenho).  
> **Data:** Outubro de 2026  
> **Status:** Estudo Técnico Concluído  

---

## 1. Visão Geral e Matriz Comparativa

O projeto atual (**Python 3.10+ com FastAPI, Uvicorn, SQLite WAL e `curl_cffi`**) atende perfeitamente ao estágio inicial e intermediário (MVP até escala média). No entanto, conforme a demanda por consultas simultâneas cresce ou a infraestrutura exige integração com ecossistemas legados, a migração para outras linguagens surge como uma decisão de engenharia estratégica.

### Matriz Comparativa Direta

| Critério Técnico | Python (Atual) | PHP (8.2+ Moderno) | Go (Golang 1.22+) | Rust (2024 Edition) |
|---|---|---|---|---|
| **Paradigma de Concorrência** | Assíncrono + Threads (`ThreadPoolExecutor`) | Síncrono (FPM) ou Fibras/Event-loop (Swoole/RoadRunner) | Nativo e Leve (Goroutines + Channels) | Async com Tokio (Futures + Tasks) |
| **Evasão TLS / WAF gov.br** | Excelente (`curl_cffi` nativo Chrome 124) | Difícil (Requer libcurl-impersonate compilado ou FFI) | Excelente (`bogdanfinn/tls-client` / `utls`) | Médio/Avançado (`rquest` / BoringSSL) |
| **Cache em Memória Local** | Nativo (`CacheTTL` com locks) | Requer Redis/APCu (ou RoadRunner/Swoole) | Trivial e Ultrarrápido (`sync.Map`, `singleflight`) | Trivial e Seguro (`dashmap`, `moka`) |
| **Consumo de Memória (RAM)** | ~80 MB a 150 MB | ~30 MB/worker (FPM) ou ~60 MB (Swoole) | **~15 MB a 30 MB (Binário único)** | **~8 MB a 20 MB (Binário único)** |
| **Velocidade de Desenvolvimento** | Altíssima | Altíssima (Laravel/Filament) | Alta | Média/Baixa (Borrow Checker) |
| **Deploy & Operação** | Python venv / Docker (~250 MB) | PHP-FPM + Nginx ou Docker (~180 MB) | **Binário estático (~20 MB em Scratch/Alpine)** | **Binário estático (~15 MB em Scratch/Alpine)** |

---

## 2. Cenário 1: Refatoração Completa em PHP (Novo Projeto)

### 2.1. Arquitetura Proposta
* **Abordagem Tradicional (Laravel / Symfony + Filament Admin):**
  * Uso do ecossistema Laravel com o pacote **Filament** para gerar o painel administrativo em poucas horas (clientes, cotas, chaves, gráficos).
* **Abordagem de Alta Performance (RoadRunner ou Swoole):**
  * Para manter a aplicação residente em memória (evitando o ciclo de vida stateless clássico do PHP-FPM, que destruiria o cache local e rate limits a cada request).

### 2.2. Prós (Vantagens)
1. **Ecossistema de Painel Administrativo Incomparável:** O ecossistema PHP (especialmente com **Laravel + Filament**) permite criar painéis administrativos, relatórios, gestão de permissões e faturamento muito mais rapidamente e com visual superior a qualquer solução em Python ou Go.
2. **Manipulação de Texto e HTML:** PHP nasceu na web. A limpeza do teor de publicações (`divBody`), decodificação de entidades HTML e expressões regulares com PCRE2 nativo são extremamente rápidas em C.
3. **Maturidade de Integração com Bancos Relacionais:** Eloquent / Doctrine / PDO SQLite oferecem abstrações maduras e amplamente conhecidas.
4. **Hospedagem Econômica:** Pode rodar em qualquer VPS simples ou hospedagem compartilhada corporativa existente.

### 2.3. Contras (Desvantagens e Gargalos Críticos)
1. **O Grande Gargalo: Evasão de TLS (Fingerprinting JA3/JA4):**
   * O WAF do portal gov.br (DJEN) bloqueia saídas que não apresentem assinatura de navegadores comuns quando em volume. O PHP padrão usa `ext-curl` compilado com OpenSSL comum. Não existe um wrapper oficial maduro de `curl-impersonate` para PHP sem compilação de extensões C customizadas ou FFI (`Foreign Function Interface`), o que fragiliza o ambiente de produção.
2. **Natureza Stateless do PHP-FPM:**
   * No modelo PHP-FPM padrão, cada requisição morre ao final. **O cache em memória e o rate limit por sliding window não funcionariam**, obrigando a introdução imediata de um serviço externo como **Redis**.
3. **Concorrência HTTP In-Process:**
   * Disparar simultaneamente chamadas para o DataJud e DJEN e correr 4 proxies residenciais em paralelo no PHP clássico exige `curl_multi_exec` manual ou Guzzle Assíncrono com Promises. A sintaxe é muito mais prolixa e propensa a vazamento de conexões se comparada ao Python/Go.

---

## 3. Cenário 2: Reescrita da API em Go (Golang)

### 3.1. Arquitetura Proposta
* **Framework Web:** `Chi` ou `Gin` (leve, idiomático e com middleware simples).
* **Camada HTTP & Evasão:** Biblioteca `bogdanfinn/tls-client` ou `uTLS` (imita com precisão cirúrgica os perfis TLS de Chrome, Firefox e Safari).
* **Concorrência:** Goroutines, Channels com `select` e `sync.singleflight` para o cache anti-thundering herd.
* **Persistência:** `mattn/go-sqlite3` ou `modernc.org/sqlite` (pure Go, sem CGO).

### 3.2. Prós (Vantagens) — *O Casamento Ideal para APIs de Consulta*
1. **Concorrência Nativa Perfeita (Goroutines & Channels):**
   * Disparar DataJud + DJEN em paralelo e fazer a corrida de 4 proxies residenciais é o caso de uso canônico para o qual Go foi concebido. O cancelamento por `context.Context` e seleção do primeiro proxy vencedor via `select` leva poucas linhas de código.
2. **Anti-Thundering Herd Nativo (`golang.org/x/sync/singleflight`):**
   * Go possui no seu ecossistema oficial o pacote `singleflight`, que resolve de forma elegante e nativa o problema de 100 requisições simultâneas para um mesmo processo aguardarem uma única chamada externa ao CNJ.
3. **Excelente Ecossistema de Evasão TLS:**
   * A comunidade Go possui as melhores e mais mantidas bibliotecas de cliente TLS customizado do mundo (`uTLS`, `tls-client`), tornando a evasão de WAF trivial sem dependências de C.
4. **Binário Único e Distribuição Imbatível:**
   * Compila em um executável estático único de ~20 MB. O contêiner Docker pode usar imagens `scratch` ou `alpine` de menos de 30 MB, sem runtime Python ou dependências dinâmicas.
5. **Consumo de Memória Ridículo:**
   * Menos de 25 MB de RAM para suportar milhares de conexões simultâneas.

### 3.3. Contras (Desvantagens)
1. **Construção do Painel Administrativo:**
   * Não existe um "Filament" ou "Django Admin" em Go. Construir o painel `/admin` exigirá codificar os templates HTML/HTMX ou criar um frontend separado em Vue/React.
2. **Verbosidade em Parsing JSON:**
   * Go exige structs rígidas com tags JSON para cada campo. Como o CNJ altera campos ocasionalmente ou omite tipos, lidar com JSONs heterogêneos exige cuidado extra com `json.RawMessage` ou unmarshaling defensivo.

---

## 4. Cenário 3: Reescrita da API em Rust

### 4.1. Arquitetura Proposta
* **Framework Web:** `Axum` sobre o runtime assíncrono `Tokio` e `Tower`.
* **Cliente HTTP:** `rquest` (baseado em reqwest com suporte a impersonate TLS via BoringSSL) ou `reqwest` com proxy socks5.
* **Cache & Estado Compartilhado:** `moka` (cache concorrente de altíssimo desempenho) ou `dashmap`.
* **Banco de Dados:** `sqlx` (assíncrono e compile-time checked) ou `rusqlite`.

### 4.2. Prós (Vantagens)
1. **Desempenho Bruto e Eficiência Absoluta:**
   * Latência de processamento de CPU próxima de zero (microssegundos).
   * Uso de memória irrisório (~8 a 15 MB de RAM).
2. **Segurança de Tipos e Corretude em Tempo de Compilação:**
   * Se compilar, praticamente não quebra em produção (*No null pointers, no data races*).
3. **Modelagem de Domínio Imutável:**
   * Enums com dados acoplados (`Result`, `Option`) eliminam erros de estado inválido ao fundir árvores de processos.

### 4.3. Contras (Desvantagens)
1. **Curva de Aprendizado e Complexidade do Borrow Checker:**
   * Lidar com referências compartilhadas entre múltiplas threads assíncronas de proxies e cache (`Arc<Mutex<...>>` ou `Arc<RwLock<...>>`) eleva drasticamente o tempo de desenvolvimento.
2. **Tempo de Compilação Lento:**
   * Compilações com dependências de rede e criptografia (BoringSSL/OpenSSL) demoram minutos.
3. **Sobrecarga de Engenharia para um Domínio I/O-Bound:**
   * A maior parte do tempo da API é gasta **esperando a resposta externa do DataJud (30 a 60 segundos)** e do DJEN (1 a 2 segundos). O ganho de CPU do Rust não reduz a latência da fonte pública do CNJ.
4. **Manutenção Complexa do Painel:**
   * Desenvolver o painel de clientes, autenticação de sessão e telas em Rust demanda esforço desproporcional.

---

## 5. Matriz Comparativa de Implementação das Features Críticas

```
┌─────────────────────────┬──────────────┬──────────────┬──────────────┬──────────────┐
│ Funcionalidade          │ Python (Hoje)│     PHP      │      Go      │     Rust     │
├─────────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ 1. Corrida de Proxies   │ Fácil        │ Difícil      │ Perfeito     │ Médio/Difícil│
│    (Racing Requests)    │ (ThreadPool) │ (curl_multi) │ (Channels)   │ (Tokio select│
├─────────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ 2. Evasão TLS Navegador │ Excelente    │ Fraco        │ Excelente    │ Bom          │
│    (JA3/Chrome)         │ (curl_cffi)  │ (Sem C/FFI)  │ (tls-client) │ (rquest)     │
├─────────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ 3. Single-Flight Cache  │ Nativo       │ Requer Redis │ Pacote Oficial│ Moka Crate  │
│    (Anti-Thundering)    │ (Locks)      │ (No FPM)     │ (singleflight│              │
├─────────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ 4. Painel Administrativo│ Bom          │ Imbatível    │ Básico       │ Complexo     │
│    (/admin com métricas)│ (Vanilla SPA)│ (Filament)   │ (Manual/HTMX)│ (Manual/Maud)│
├─────────────────────────┼──────────────┼──────────────┼──────────────┼──────────────┤
│ 5. Custo de Manutenção  │ Baixo        │ Médio        │ Muito Baixo  │ Alto         │
└─────────────────────────┴──────────────┴──────────────┴──────────────┴──────────────┘
```

---

## 6. Recomendação Estratégica do Arquiteto

Se a meta for **evoluir o produto para escalabilidade enterprise**:

### 🎯 Recomendação #1: A Melhor Escolha para a API é **Go (Golang)**
Se você for criar um novo projeto para a parte da API (`/v1/...`), **Go é a escolha mais equilibrada e tecnicamente superior**:
* Possui o padrão `singleflight` nativo no time de Go (`golang.org/x/sync/singleflight`).
* Goroutines e canais resolvem a corrida concorrente de proxies e a fusão DataJud + DJEN em menos de 100 linhas de código limpo.
* A biblioteca `bogdanfinn/tls-client` em Go é a mais adotada do mercado para scrapers corporativos contra WAFs.
* Gera um binário Docker minúsculo (< 30 MB) consumindo 20 MB de RAM.

### 🎯 Recomendação #2: Cenário Ideal para PHP
Refatorar em PHP só é recomendado se:
1. Sua equipe já possui forte especialização em PHP/Laravel.
2. O objetivo principal for criar um SaaS corporativo com um painel administrativo ultra-avançado e rico em telas (usando **Laravel 11 + Filament 3**).
3. **Ressalva obrigatória:** Deverá ser utilizado **RoadRunner** ou **Swoole** (para manter a API em memória e suportar concorrência assíncrona) e uma instância de **Redis** para centralizar cache e rate-limits.

### 🎯 Recomendação #3: Por que Rust Não é Prioritário Agora
Rust é uma obra-prima de engenharia para sistemas de processamento massivo de CPU, gateways financeiros de microssegundos e drivers. No entanto, o **Robô Jurídico é um sistema 99% I/O-Bound** (passa a maior parte do tempo aguardando a resposta de 40 segundos do Elasticsearch do CNJ). O custo de desenvolvimento, o rigor do compilador e a manutenção tornam o ROI (retorno sobre o investimento) inferior ao de Go.
