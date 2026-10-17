# Manual de Integração e API Reference: Robô Jurídico

> **Identificador:** API-REF-RJ-001  
> **Versão da API:** 0.1.0  
> **Base URL (Padrão Local):** `http://127.0.0.1:8000`  
> **Protocolo:** HTTP/1.1 ou HTTP/2 sobre JSON (UTF-8)  
> **Data:** Outubro de 2026  

---

## 1. Guia de Início Rápido (Quickstart)

Toda chamada aos endpoints de dados da API requer a inclusão do cabeçalho de autorização `X-API-Key` contendo a chave de acesso atribuída ao cliente.

### 1.1. Exemplo de Requisição cURL
```bash
curl.exe -H "X-API-Key: rj_live_SUA_CHAVE_AQUI" \
  "http://127.0.0.1:8000/v1/processos/5010754-07.2024.8.13.0625"
```

> [!NOTE]
> No Windows PowerShell, utilize explicitamente `curl.exe` em vez de `curl`, pois `curl` é um alias padrão para o cmdlet `Invoke-WebRequest`.

---

## 2. Cabeçalhos HTTP e Controle de Cotas

### 2.1. Cabeçalhos de Entrada
| Cabeçalho | Obrigatório? | Descrição |
|---|---|---|
| `X-API-Key` | Sim (rotas de dados) | Chave de autenticação gerada no painel administrativo (`rj_live_...`) ou chave interna configurada no `.env`. |

### 2.2. Cabeçalhos de Resposta (Consumo e Rastreamento)
Todas as respostas bem-sucedidas ou associadas a uma chave de cliente retornam metadados atualizados sobre o consumo contratado:

```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Cota-Limite: 2000
X-Cota-Usada: 42
X-Cota-Restante: 1958
X-Request-ID: 0090ea8b83174305
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
```

| Cabeçalho | Significado |
|---|---|
| `X-Cota-Limite` | Volume total de consultas contratadas para o mês corrente. |
| `X-Cota-Usada` | Quantidade de consultas já faturadas no mês atual. |
| `X-Cota-Restante` | Saldo disponível de consultas antes do esgotamento da cota. |
| `X-Request-ID` | Identificador único da requisição para fins de log e suporte técnico. |
| `Retry-After` | Presente em respostas HTTP 429 ou 503 indicando os segundos que o cliente deve aguardar antes de tentar novamente. |

---

## 3. Catálogo Completo de Endpoints

### 3.1. `GET /v1/processos/{numero}`
Retorna a visão integral do processo: capa, órgãos julgadores, partes polarizadas, advogados, histórico unificado de movimentações e publicações no diário oficial.

- **Desconta Cota:** **Sim (1 consulta)**
- **Parâmetros de Caminho:**
  - `numero` (string, obrigatório): Número do processo no formato CNJ com ou sem pontuação (ex.: `5010754-07.2024.8.13.0625` ou `50107540720248130625`).
- **Parâmetros de Query:**
  - `modo` (string, opcional, padrão: `completo`):
    - `completo`: Consulta simultaneamente DataJud e DJEN. Retorna histórico profundo e diário unificados.
    - `rapido`: Consulta exclusivamente o DJEN. Resposta em menos de 2 segundos, sem movimentos internos do DataJud.
  - `publicacoes` (boolean, opcional, padrão: `true`): Permite ignorar as publicações do diário em `modo=completo`.

#### Exemplo de Resposta (HTTP 200 OK):
```json
{
  "ok": true,
  "consultado_em": "2026-10-13T15:30:00Z",
  "tempo_ms": 1420,
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
    "orgao_julgador": "Turma Recursal de Belo Horizonte",
    "comarca": "Belo Horizonte",
    "link": "https://pje.tjmg.jus.br/pje/ConsultaPublica/listView.seam",
    "link_peca": "https://comunicaapi.pje.jus.br/api/v1/comunicacao/QJDEM7aXY8BIQ1aUrTVzl1L6oWe2dL/certidao",
    "valor_causa": 12500.0,
    "valor_causa_formatado": "R$ 12.500,00",
    "data_ajuizamento": "2024-05-10T14:22:00Z",
    "data_ajuizamento_formatada": "10/05/2024",
    "data_ultima_movimentacao": "2026-09-14",
    "data_ultima_movimentacao_formatada": "14/09/2026",
    "polos": {
      "ativo": [
        {
          "nome": "EMPRESA EXEMPLO S.A.",
          "polo": "ativo",
          "advogados": [
            {
              "nome": "DRA. ADVOGADA SILVA",
              "oab_numero": "123456",
              "oab_uf": "MG",
              "oab": "OAB/MG 123456"
            }
          ]
        }
      ],
      "passivo": [
        {
          "nome": "FULANO DE TAL",
          "polo": "passivo",
          "advogados": []
        }
      ],
      "terceiros": []
    },
    "advogados": [
      {
        "nome": "DRA. ADVOGADA SILVA",
        "oab_numero": "123456",
        "oab_uf": "MG",
        "oab": "OAB/MG 123456"
      }
    ],
    "movimentacoes": [
      {
        "data": "2026-09-14T12:44:48Z",
        "data_formatada": "14/09/2026",
        "tipo": "outro",
        "descricao": "Trânsito em julgado",
        "complemento": "Certificado o trânsito",
        "orgao": "Secretaria",
        "fonte": "datajud",
        "link": "https://pje.tjmg.jus.br/pje/ConsultaPublica/listView.seam",
        "id_externo": "848_2026-09-14T12:44:48_a1b2c3d4"
      },
      {
        "data": "2026-09-02T00:00:00Z",
        "data_formatada": "02/09/2026",
        "tipo": "intimacao",
        "descricao": "Intimação",
        "complemento": "Ficam as partes intimadas para cumprimento do acórdão...",
        "orgao": "Turma Recursal",
        "fonte": "comunica",
        "link": "https://comunicaapi.pje.jus.br/api/v1/comunicacao/QJDEM7aXY8BIQ1aUrTVzl1L6oWe2dL/certidao",
        "link_peca": "https://comunicaapi.pje.jus.br/api/v1/comunicacao/QJDEM7aXY8BIQ1aUrTVzl1L6oWe2dL/certidao",
        "id_externo": "comunica-715599267"
      }
    ],
    "publicacoes": [
      {
        "id": "715599267",
        "data_disponibilizacao": "2026-09-02",
        "data_formatada": "02/09/2026",
        "tipo_comunicacao": "Intimação",
        "tipo_documento": "Despacho",
        "orgao": "Turma Recursal",
        "numero_processo": "5010754-07.2024.8.13.0625",
        "classe": "Recurso Inominado Cível",
        "texto": "Ficam as partes intimadas para cumprimento do acórdão...",
        "link": "https://comunicaapi.pje.jus.br/api/v1/comunicacao/QJDEM7aXY8BIQ1aUrTVzl1L6oWe2dL/certidao",
        "link_peca": "https://comunicaapi.pje.jus.br/api/v1/comunicacao/QJDEM7aXY8BIQ1aUrTVzl1L6oWe2dL/certidao",
        "partes": [],
        "advogados": []
      }
    ],
    "totais": {
      "partes": 2,
      "advogados": 1,
      "movimentacoes": 45,
      "publicacoes": 8
    },
    "fontes": ["datajud", "comunica"]
  }
}
```

> [!TIP]
> **Diferença entre `link` e `link_peca`:**
> - `processo.link`: Aponta sempre para o portal oficial de consulta pública processual do Tribunal (PJe, eproc, e-SAJ).
> - `processo.link_peca` / `movimentacoes[].link_peca`: Aponta diretamente para o **PDF oficial da Certidão de Publicação do CNJ** (`https://comunicaapi.pje.jus.br/api/v1/comunicacao/{hash}/certidao`) ou documento público do PJe (`ConsultaDocumento/listView.seam?x=...`). Quando um movimento do DataJud corresponde a um ato publicado no DJEN (janela de até 72h), o orquestrador preenche automaticamente `link_peca` também no movimento do DataJud.

---

### 3.2. `GET /v1/processos/{numero}/existe`
Verifica rapidamente se o número já consta como distribuído e indexado em alguma das bases públicas do CNJ.
- **Desconta Cota:** **Não (Gratuito)**
- **Resposta:**
```json
{
  "ok": true,
  "numero": "5010754-07.2024.8.13.0625",
  "existe": true,
  "fonte": "comunica"
}
```

---

### 3.3. `GET /v1/cnj/{numero}`
Decodifica sintaticamente o número processual com base na Resolução CNJ 65/2008 sem efetuar qualquer requisição externa à internet.
- **Desconta Cota:** **Não (Gratuito)**
- **Resposta:**
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

### 3.4. `GET /v1/oab/{uf}/{numero}/processos`
Lista todos os processos nos quais a OAB especificada teve atos publicados no DJEN dentro da janela de dias solicitada.
- **Desconta Cota:** **Sim (1 consulta)**
- **Parâmetros:**
  - `uf` (caminho): UF da OAB com 2 letras (ex.: `MG`, `SP`).
  - `numero` (caminho): Dígitos da inscrição (ex.: `123456`).
  - `dias` (query, opcional, padrão: `30`, min: 1, max: 365): Janela histórica.
  - `tribunal` (query, opcional): Filtra apenas processos de determinado tribunal (ex.: `TJMG`, `TRT3`).
- **Resposta:**
```json
{
  "ok": true,
  "oab": "OAB/MG 123456",
  "periodo_dias": 30,
  "total": 1,
  "processos": [
    {
      "numero": "0010734-59.2026.5.03.0076",
      "titulo": "RECLAMANTE × RECLAMADA S.A.",
      "tribunal": "TRT3",
      "classe": "Ação Trabalhista",
      "orgao_julgador": "1ª Vara do Trabalho",
      "data_ultima_movimentacao": "2026-09-25",
      "data_ultima_movimentacao_formatada": "25/09/2026",
      "publicacoes": 2
    }
  ]
}
```

---

### 3.5. `GET /v1/oab/{uf}/{numero}/publicacoes`
Retorna as publicações brutas do DJEN onde a OAB aparece no período, contendo o teor integral do ato sanitizado. Endpoint fundamental para rotinas de **controle de prazos**.
- **Desconta Cota:** **Sim (1 consulta)**
- **Parâmetros:**
  - `uf` e `numero` (caminho): Dados da OAB.
  - `dias` (query, padrão: `7`, min: 1, max: 90): Janela em dias até a data de hoje.
  - `paginas` (query, padrão: `1`, min: 1, max: 20): Quantidade de páginas (100 itens/página) a recuperar.

---

### 3.6. `GET /v1/conta`
Informa o plano contratado pelo cliente, consumo acumulado no mês e data de renovação.
- **Desconta Cota:** **Não (Gratuito)**
- **Resposta:**
```json
{
  "ok": true,
  "cliente": "Escritório Modelo Advocacia",
  "plano": "Start",
  "limite_mensal": 2000,
  "usado_no_mes": 42,
  "restante": 1958,
  "reinicia_em": "2026-11-01T03:00:00Z",
  "limite_por_minuto": 30
}
```

---

### 3.7. `GET /saude`
Healthcheck leve para balanceadores de carga, proxies reversos e monitoramento.
- **Autenticação:** Não exige cabeçalho `X-API-Key`.
- **Resposta:**
```json
{
  "ok": true,
  "versao": "0.1.0"
}
```

---

## 4. Tratamento de Erros e Respostas HTTP

Todos os erros retornam um corpo JSON padronizado com código semântico invariável:

```json
{
  "ok": false,
  "erro": {
    "codigo": "numero_invalido",
    "mensagem": "Número fora do padrão CNJ ou com dígito verificador errado."
  },
  "id_requisicao": "0090ea8b83174305"
}
```

### 4.1. Tabela de Códigos de Erro

| Código HTTP | `erro.codigo` | Quando Ocorre | Ação Recomendada do Consumidor |
|---|---|---|---|
| `401 Unauthorized` | `nao_autorizado` | Chave ausente, malformada ou revogada. | Verificar o header `X-API-Key`. Não adianta repetir. |
| `402 Payment Required`| `cota_esgotada` | Limite mensal de consultas foi atingido. | Aguardar o dia 1º ou solicitar ampliação de plano ao operador. |
| `403 Forbidden` | `conta_suspensa` | Conta do cliente foi desativada no painel. | Entrar em contato com o suporte/administrador. |
| `404 Not Found` | `processo_nao_encontrado` | Nenhuma fonte pública conhece o processo. | Se for processo novo, tentar mais tarde. Processos em segredo de justiça não aparecem. |
| `422 Unprocessable` | `numero_invalido` | Dígito verificador errado ou padrão incorreto. | Corrigir o número digitado. Não desconta cota. |
| `422 Unprocessable` | `uf_invalida` | Sigla de UF inexistente (ex.: XX). | Informar uma das 27 siglas de UF brasileiras. |
| `429 Too Many Req.` | `limite_excedido` | Ultrapassou o limite de requisições por minuto. | **Aguardar os segundos indicados no header `Retry-After`.** |
| `429 Too Many Req.` | `muitas_tentativas` | IP bloqueado por envio excessivo de chaves erradas. | Aguardar o término da janela de bloqueio (15 min). |
| `500 Server Error` | `erro_interno` | Exceção inesperada na execução do servidor. | Registrar o `id_requisicao` e reportar ao suporte. |
| `502 Bad Gateway` | `falha_na_fonte` | Resposta anômala retornada pelo CNJ. | Tentar novamente após alguns segundos. |
| `503 Service Unavail.`| `fonte_indisponivel` | DataJud ou DJEN temporariamente fora do ar. | **Repetir após o período de `Retry-After` (30s) ou usar `modo=rapido`.** |

---

## 5. Exemplos de Implementação de Clientes

### 5.1. Implementação em Python (usando `httpx`)
```python
import os
import time
import httpx

API_URL = os.environ.get("ROBO_API_URL", "http://127.0.0.1:8000")
API_KEY = os.environ["ROBO_API_KEY"]

def consultar_processo(numero: str, modo: str = "completo", max_retries: int = 3) -> dict:
    url = f"{API_URL}/v1/processos/{numero}"
    headers = {"X-API-Key": API_KEY}
    params = {"modo": modo}
    
    with httpx.Client(timeout=120.0) as client:
        for tentativa in range(1, max_retries + 1):
            response = client.get(url, params=params, headers=headers)
            
            # Repetir apenas em erros transientes que contenham Retry-After
            if response.status_code in (429, 502, 503):
                retry_after = int(response.headers.get("Retry-After", 5))
                time.sleep(retry_after)
                continue
                
            dados = response.json()
            if not dados.get("ok"):
                erro = dados.get("erro", {})
                raise RuntimeError(f"Erro {erro.get('codigo')}: {erro.get('mensagem')}")
                
            print(f"Saldo restante no mês: {response.headers.get('X-Cota-Restante')}")
            return dados["processo"]
            
    raise RuntimeError("Falha de conexão com a API após tentativas.")
```

### 5.2. Implementação em Node.js / TypeScript (Node 18+)
```typescript
import process from 'node:process';

const API_URL = process.env.ROBO_API_URL || 'http://127.0.0.1:8000';
const API_KEY = process.env.ROBO_API_KEY!;

interface ProcessoResponse {
  ok: boolean;
  processo: {
    numero: string;
    titulo: string;
    tribunal: string;
    movimentacoes: Array<{ data_formatada: string; descricao: string }>;
  };
  erro?: { codigo: string; mensagem: string };
}

async function consultarProcesso(numero: string, modo = 'completo'): Promise<ProcessoResponse['processo']> {
  const url = `${API_URL}/v1/processos/${numero}?modo=${modo}`;
  
  const res = await fetch(url, {
    method: 'GET',
    headers: { 'X-API-Key': API_KEY }
  });

  const data = (await res.json()) as ProcessoResponse;
  if (!data.ok) {
    throw new Error(`[${data.erro?.codigo}] ${data.erro?.mensagem}`);
  }

  console.log('Cota restante:', res.headers.get('X-Cota-Restante'));
  return data.processo;
}
```
