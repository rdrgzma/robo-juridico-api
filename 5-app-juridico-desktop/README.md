# ⚖️ Robô Jurídico Desktop On-Premise (Windows & Linux)

> **Aplicativo Desktop Jurídico Autocontido:** Executa 100% dentro da máquina do cliente final (escritório de advocacia ou perito judicial), com zero dependências externas (sem Docker, sem PHP, sem Node.js e sem necessidade de servidores externos).

---

## 📦 O que contém neste projeto (`5-app-juridico-desktop`)

Este projeto foi construído em uma pasta completamente isolada para preservar a stack original da API Go e do Painel PHP Filament.

| Arquivo / Diretório | Função e Propósito |
| :--- | :--- |
| **`robo-juridico-desktop.exe`** | **Executável nativo compilado para Windows (x64).** Tamanho de apenas ~18.8 MB com frontend embutido. |
| **`robo-juridico-desktop-linux`** | **Executável nativo compilado para Linux (x64 ELF).** Pronto para Ubuntu, Debian, Fedora, Mint, etc. |
| **`iniciar-desktop.bat`** | **Script de 1 clique para inicialização no Windows.** Abre o executável e o navegador padrão automaticamente. |
| **`iniciar-desktop.sh`** | **Script de 1 clique para inicialização no Linux.** Dá permissão de execução e sobe o app nativamente. |
| **`cmd/app/main.go`** | Ponto de entrada do executável: inicializa o SQLite local, a rede loopback e dispara o navegador. |
| **`ui/`** | Interface do usuário (HTML5, CSS profissional escuro e JavaScript reativo) embutida via `//go:embed`. |
| **`internal/bridge/`** | Camada de serviços desacoplada de HTTP que conecta a UI diretamente ao motor Go. |
| **`internal/licensing/`** | Módulo de identificação de máquina (`Hardware Fingerprint`) e validação criptográfica ECDSA P-256. |
| **`internal/calculator/`** | Motor de liquidação judicial (DrCalc/PJe-Calc, Lei 14.905/2024, Art. 354 CC, revisional Price/SAC/Gauss). |
| **`internal/prazos/`** | Calculadora forense de prazos processuais (CPC e CLT em dias úteis, Art. 224 e Art. 220). |
| **`dados/desktop.db`** | Banco de dados SQLite local no modo WAL para histórico de cálculos e consultas do cliente. |

---

## 🚀 Como Executar

### No Windows:
Basta dar duplo clique em:
```text
iniciar-desktop.bat
```
Ou rodar diretamente no terminal:
```powershell
.\robo-juridico-desktop.exe
```
O aplicativo abrirá automaticamente a interface em `http://127.0.0.1:7890`.

### No Linux:
Abra o terminal na pasta e execute:
```bash
chmod +x iniciar-desktop.sh
./iniciar-desktop.sh
```

---

## 🛠️ Recursos e Telas Disponíveis

1. **Dashboard do Sistema:**
   * Exibição do Hardware Fingerprint único da máquina.
   * Status de ativação da licença e data de expiração.
   * Total de índices carregados e atalhos rápidos.

2. **Consulta Unificada CNJ (DataJud + DJEN) & Visualizador de Documentos:**
   * **Validação em tempo real** do dígito verificador CNJ (Módulo 97) e identificação automática do Tribunal/Segmento.
   * **Limpeza Automática & Card Animado de Consulta às Fontes Oficiais:** Ao digitar um novo processo ou iniciar uma nova busca, os resultados anteriores são limpos imediatamente e substituídos por um banner animado indicando as bases oficiais conectadas em tempo real (`🏛️ DataJud / CNJ` e `📰 DJEN / Comunica PJe`).
   * **Barra de Fontes Consultadas e Avisos:** Exibe de forma destacada no resultado quais bases responderam, a contagem exata de andamentos do DataJud e publicações do DJEN, além de alertas de sincronização do orquestrador.
   * **Links de Peças e Certidões PDF (`📄 Abrir Peça (PDF) ↗`):** Acesso direto em 1 clique às certidões oficiais em PDF do CNJ/DJEN (`comunicaapi.pje.jus.br/api/v1/comunicacao/{hash}/certidao`) e documentos do PJe, com correlação automática entre andamentos do DataJud e publicações do DJEN na mesma janela de 72h.
   * **Visualizador Integrado de Documento / Certidão Oficial (`📄 Ver Documento`):** Todas as movimentações (tanto do **DataJud** quanto do **DJEN**) contam com um modal em formato **Folha A4 Oficial do Poder Judiciário** exibindo os metadados completos do processo, código identificador TPU/CNJ, data/hora UTC, unidade judiciária e teor completo do andamento, com botão próprio de **`🖨️ Imprimir / Salvar PDF`**.
   * **1-Click Auto-Fill para Liquidação:** Transfere automaticamente número do processo, partes (Autor × Réu), indexador do tribunal e valor da causa para a aba de Liquidação Judicial.

3. **Liquidação Judicial de Sentença:**
   * Padrão oficial DrCalc / PJe-Calc.
   * Adequação à **Lei 14.905/2024** (SELIC deduzida de IPCA).
   * Imputação de pagamentos parciais nos juros conforme **Art. 354 do Código Civil**.
   * Multa moratória, honorários periciais e encargos do **Art. 523 do CPC** (10% de multa + 10% honorários).
   * Adição e remoção dinâmica de parcelas, custas forenses e abatimentos.
   * **Exportação para Excel (.csv)** e visualização formatada para impressão de **Laudo A4**.

4. **Revisão de Financiamento Bancário:**
   * Comparação da Tabela Price contratada vs. Taxa média de mercado divulgada pelo Bacen.
   * Recálculo pelos sistemas **PRICE**, **SAC** ou **GAUSS** (juros lineares/simples).
   * Apuração de indébito com opção de **restituição em dobro (Art. 42 do CDC)**.

5. **Calculadora de Prazos Processuais Forenses:**
   * Contagem em dias úteis no padrão **CPC/2015** e **CLT**.
   * Regra do Art. 224: o primeiro dia útil após a publicação é o início da contagem.
   * Tratamento de recesso forense (20/dez a 20/jan - Art. 220 CPC) e feriados nacionais.
   * Memória pericial detalhada dia a dia da evolução do prazo até o vencimento fatal.

6. **Índices Econômicos & Sincronização Bacen/IBGE:**
   * Tabela dos 27 índices carregados (TJSP, TJRJ, TJMG, IPCA, INPC, IGP-M, SELIC, TR, etc.).
   * Botão de **sincronização instantânea** com os webservices do Banco Central e IBGE.

7. **Histórico Local no SQLite:**
   * Todos os cálculos realizados ficam salvos na máquina do cliente para consulta futura.
   * Privacidade absoluta e conformidade total com a LGPD e sigilo profissional da OAB.

8. **Central de Licenciamento:**
   * Exibe o Machine ID para o cliente enviar ao suporte.
   * Campo para colar e ativar a licença assinada criptograficamente via ECDSA.

9. **Tema Claro e Escuro (Light & Dark Mode):**
   * Suporte nativo a alternância com 1 clique (botão na barra lateral e cabeçalho).
   * Detecção automática da preferência do sistema operacional e persistência da escolha do usuário no terminal.

---

## 🔐 Como Funciona o Licenciamento

* O sistema coleta o identificador de hardware da máquina do cliente (UUID da placa-mãe no Windows ou DMI product UUID no Linux).
* O fornecedor do software gera um arquivo/token assinado com sua **chave privada ECDSA**.
* O aplicativo desktop, usando a chave pública embutida, valida a autenticidade, validade e recursos liberados, permitindo o uso 100% offline com segurança contra clonagem de máquinas.
