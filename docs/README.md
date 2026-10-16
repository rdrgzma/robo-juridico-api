# Índice Geral da Documentação Técnica e de Produto: Robô Jurídico

Este diretório reúne o pacote completo de especificações técnicas, documentos de requisitos de produto (PRD), manuais de integração de API e guias de governança operacional desenvolvidos a partir da análise dos documentos originais e do código-fonte do projeto **Robô Jurídico (`consulta-processos`)**.

---

## Estrutura da Documentação

| Arquivo | Título | Objetivo | Público-Alvo |
|---|---|---|---|
| [`01-ANALISE-DO-PROJETO.md`](01-ANALISE-DO-PROJETO.md) | **Análise Estratégica, Arquitetural e Técnica** | Diagnóstico executivo da solução, análise SWOT, riscos e viabilidade de engenharia e negócios. | CTOs, Tech Leads, Product Managers |
| [`02-PRD-DOCUMENTO-DE-REQUISITOS-DE-PRODUTO.md`](02-PRD-DOCUMENTO-DE-REQUISITOS-DE-PRODUTO.md) | **Product Requirements Document (PRD)** | Definição de personas, jornadas, requisitos funcionais (RF), não-funcionais (RNF), precificação e roadmap. | Product Owners, Empreendedores, Desenvolvedores |
| [`03-DOCUMENTACAO-TECNICA-E-ARQUITETURA.md`](03-DOCUMENTACAO-TECNICA-E-ARQUITETURA.md) | **Especificação Técnica e Arquitetura de Software** | Diagramas C4/Mermaid, fluxo concorrente, cache anti-thundering herd, banco de dados SQLite WAL e segurança. | Arquitetos de Software, Engenheiros de Backend |
| [`04-MANUAL-DE-INTEGRACAO-E-API-REFERENCE.md`](04-MANUAL-DE-INTEGRACAO-E-API-REFERENCE.md) | **Manual de Integração & API Reference** | Catálogo dos endpoints REST, schemas de payload, dicionário de campos, tratamento de erros e snippets (Python, Node). | Desenvolvedores Clientes, Integradores de Sistemas |
| [`05-GUIA-DE-OPERACAO-SEGURANCA-E-LGPD.md`](05-GUIA-DE-OPERACAO-SEGURANCA-E-LGPD.md) | **Guia de Operações, Infraestrutura, Segurança e LGPD** | Procedimentos de deploy (Docker, Nginx, Systemd), backups a quente, proxies residenciais, governança LGPD e troubleshooting. | Engenheiros DevOps, Sysadmins, DPOs/Compliance |
| [`06-ANALISE-REFATORACAO-PHP-GO-RUST.md`](06-ANALISE-REFATORACAO-PHP-GO-RUST.md) | **Estudo de Refatoração: PHP vs Go vs Rust** | Avaliação comparativa de prós e contras, desafios de concorrência, evasão de WAF e viabilidade técnica. | Arquitetos de Software, CTOs, Engenheiros de Backend |
| [`07-GUIA-DE-DEPLOY-DA-API.md`](07-GUIA-DE-DEPLOY-DA-API.md) | **Guia de Deploy e Operação em Produção** | Manual passo a passo para deploy via Docker, Systemd Bare-Metal, Proxy Reverso (Caddy/Nginx SSL), Backups e CI/CD. | Engenheiros DevOps, Sysadmins, SREs |
| [`08-GUIA-SISTEMA-JURIDICO-DESKTOP-ON-PREMISE.md`](08-GUIA-SISTEMA-JURIDICO-DESKTOP-ON-PREMISE.md) | **Guia do Sistema Desktop On-Premise (Win/Linux)** | Especificação de arquitetura, licenciamento criptográfico por Hardware ID, modo offline e manual de operação local. | Desenvolvedores Desktop, Peritos, Advogados |
| [`openapi.yaml`](openapi.yaml) / [`openapi.json`](openapi.json) | **Especificação OpenAPI 3.1.0 (Swagger)** | Definição formal de todos os endpoints, esquemas de entrada/saída, headers de cota e códigos HTTP. | Desenvolvedores, Integradores, Ferramentas Swagger/Postman |

---

## Resumo dos Componentes do Sistema

```
Robo-Juridico-API/
├── 1-Comece-Aqui.pdf                  # Onboarding em 3 passos
├── 2-Workbook-API-Robo-Juridico.pdf   # Manual didático original (24 páginas)
├── docs/                              # Documentação corporativa gerada
│   ├── 01-ANALISE-DO-PROJETO.md
│   ├── 02-PRD-DOCUMENTO-DE-REQUISITOS-DE-PRODUTO.md
│   ├── 03-DOCUMENTACAO-TECNICA-E-ARQUITETURA.md
│   ├── 04-MANUAL-DE-INTEGRACAO-E-API-REFERENCE.md
│   ├── 05-GUIA-DE-OPERACAO-SEGURANCA-E-LGPD.md
│   └── README.md
└── 3-api-robo-juridico/
    └── api-robo-juridico/             # Código-fonte Python
        ├── consulta_processos/        # Pacote principal
        │   ├── fontes/                # Drivers DataJud e DJEN
        │   ├── painel/                # Admin SPA, SQLite e Segurança
        │   ├── api.py                 # Orquestrador concorrente
        │   ├── cnj.py                 # Validador de DV Módulo 97 e roteador
        │   ├── cache.py               # Cache TTL com single-flight lock
        │   ├── http.py                # Cliente HTTP com curl_cffi
        │   ├── proxies.py             # Pool residencial em paralelo
        │   ├── rest.py                # Servidor FastAPI e middleware
        │   └── web.py                 # Web UI local zero-dependency
        └── tests/                     # Bateria de testes automatizados
```
