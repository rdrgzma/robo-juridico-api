package domain

import (
	"fmt"
	"strings"
	"time"
)

type Fonte string

const (
	FonteDataJud  Fonte = "datajud"
	FonteComunica Fonte = "comunica"
)

type TipoMovimentacao string

const (
	TipoDespacho      TipoMovimentacao = "despacho"
	TipoDecisao       TipoMovimentacao = "decisao"
	TipoSentenca      TipoMovimentacao = "sentenca"
	TipoAudiencia     TipoMovimentacao = "audiencia"
	TipoIntimacao     TipoMovimentacao = "intimacao"
	TipoPublicacaoDJE TipoMovimentacao = "publicacao_dje"
	TipoOutro         TipoMovimentacao = "outro"
)

type Polo string

const (
	PoloAtivo    Polo = "ativo"
	PoloPassivo  Polo = "passivo"
	PoloTerceiro Polo = "terceiro"
)

type Advogado struct {
	Nome      string `json:"nome"`
	OABNumero string `json:"oab_numero"`
	OABUF     string `json:"oab_uf"`
	OAB       string `json:"oab"`
}

func (a Advogado) Formatada() string {
	if a.OABNumero != "" && a.OABUF != "" {
		return fmt.Sprintf("OAB/%s %s", strings.ToUpper(a.OABUF), a.OABNumero)
	}
	return ""
}

type Parte struct {
	Nome      string     `json:"nome"`
	Polo      Polo       `json:"polo"`
	Advogados []Advogado `json:"advogados"`
}

type Movimentacao struct {
	Data          time.Time `json:"data"`
	DataFormatada string    `json:"data_formatada"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	Complemento   string    `json:"complemento"`
	Orgao         string    `json:"orgao"`
	Fonte         string    `json:"fonte"`
	Link          string    `json:"link"`
	IDExterno     string    `json:"-"` // identificador único interno para deduplicação
}

type Publicacao struct {
	ID                   string     `json:"id"`
	DataDisponibilizacao string     `json:"data_disponibilizacao"`
	DataFormatada        string     `json:"data_formatada"`
	TipoComunicacao      string     `json:"tipo_comunicacao"`
	TipoDocumento        string     `json:"tipo_documento"`
	Orgao                string     `json:"orgao"`
	NumeroProcesso       string     `json:"numero_processo"`
	Classe               string     `json:"classe"`
	Texto                string     `json:"texto"`
	Link                 string     `json:"link"`
	Partes               []Parte    `json:"partes"`
	Advogados            []Advogado `json:"advogados"`
}

type Polos struct {
	Ativo     []Parte `json:"ativo"`
	Passivo   []Parte `json:"passivo"`
	Terceiros []Parte `json:"terceiros"`
}

type Totais struct {
	Partes        int `json:"partes"`
	Advogados     int `json:"advogados"`
	Movimentacoes int `json:"movimentacoes"`
	Publicacoes   int `json:"publicacoes"`
}

type Processo struct {
	Numero                         string         `json:"numero"`
	NumeroDigitos                  string         `json:"numero_digitos"`
	Titulo                         string         `json:"titulo"`
	Tribunal                       string         `json:"tribunal"`
	Segmento                       string         `json:"segmento"`
	Grau                           string         `json:"grau"`
	Classe                         string         `json:"classe"`
	Assunto                        string         `json:"assunto"`
	OrgaoJulgador                  string         `json:"orgao_julgador"`
	Comarca                        string         `json:"comarca"`
	Link                           string         `json:"link"`
	ValorCausa                     *float64       `json:"valor_causa"`
	ValorCausaFormatado            string         `json:"valor_causa_formatado"`
	DataAjuizamento                *time.Time     `json:"data_ajuizamento"`
	DataAjuizamentoFormatada       string         `json:"data_ajuizamento_formatada"`
	DataUltimaMovimentacao         string         `json:"data_ultima_movimentacao"`
	DataUltimaMovimentacaoFormatada string         `json:"data_ultima_movimentacao_formatada"`
	Partes                         []Parte        `json:"-"`
	Polos                          Polos          `json:"polos"`
	Advogados                      []Advogado     `json:"advogados"`
	Movimentacoes                  []Movimentacao `json:"movimentacoes"`
	Publicacoes                    []Publicacao   `json:"publicacoes"`
	Totais                         Totais         `json:"totais"`
	Fontes                         []string       `json:"fontes"`
}

// Respostas da API REST

type RespostaProcesso struct {
	OK           bool     `json:"ok"`
	ConsultadoEm string   `json:"consultado_em"`
	TempoMs      int64    `json:"tempo_ms"`
	Avisos       []string `json:"avisos"`
	Processo     Processo `json:"processo"`
}

type RespostaExiste struct {
	OK     bool   `json:"ok"`
	Numero string `json:"numero"`
	Existe bool   `json:"existe"`
	Fonte  string `json:"fonte"`
}

type RespostaCNJ struct {
	OK            bool   `json:"ok"`
	Numero        string `json:"numero"`
	NumeroDigitos string `json:"numero_digitos"`
	Valido        bool   `json:"valido"`
	Tribunal      string `json:"tribunal,omitempty"`
	Segmento      string `json:"segmento,omitempty"`
	Ano           int    `json:"ano,omitempty"`
	OrigemCodigo  int    `json:"origem_codigo,omitempty"`
}

type ResumoProcesso struct {
	Numero                         string `json:"numero"`
	Titulo                         string `json:"titulo"`
	Tribunal                       string `json:"tribunal"`
	Classe                         string `json:"classe"`
	OrgaoJulgador                  string `json:"orgao_julgador"`
	DataUltimaMovimentacao         string `json:"data_ultima_movimentacao"`
	DataUltimaMovimentacaoFormatada string `json:"data_ultima_movimentacao_formatada"`
	Publicacoes                    int    `json:"publicacoes"`
}

type RespostaOABProcessos struct {
	OK          bool             `json:"ok"`
	OAB         string           `json:"oab"`
	PeriodoDias int              `json:"periodo_dias"`
	Total       int              `json:"total"`
	Processos   []ResumoProcesso `json:"processos"`
}

type RespostaOABPublicacoes struct {
	OK          bool         `json:"ok"`
	OAB         string       `json:"oab"`
	Inicio      string       `json:"inicio"`
	Fim         string       `json:"fim"`
	Total       int          `json:"total"`
	Publicacoes []Publicacao `json:"publicacoes"`
}

type RespostaConta struct {
	OK              bool   `json:"ok"`
	Cliente         string `json:"cliente"`
	Plano           string `json:"plano"`
	LimiteMensal    *int   `json:"limite_mensal"`
	UsadoNoMes      int    `json:"usado_no_mes"`
	Restante        *int   `json:"restante"`
	ReiniciaEm      string `json:"reinicia_em"`
	LimitePorMinuto int    `json:"limite_por_minuto"`
}

type ErroDetalhe struct {
	Campo    string `json:"campo,omitempty"`
	Mensagem string `json:"mensagem"`
}

type ErroCorpo struct {
	Codigo   string        `json:"codigo"`
	Mensagem string        `json:"mensagem"`
	Detalhes []ErroDetalhe `json:"detalhes,omitempty"`
}

type RespostaErro struct {
	OK           bool      `json:"ok"`
	Erro         ErroCorpo `json:"erro"`
	IDRequisicao string    `json:"id_requisicao,omitempty"`
}
