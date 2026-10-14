package calculator

import (
	"time"

	"github.com/shopspring/decimal"
)

// TipoParcela define a natureza de uma parcela na liquidação
type TipoParcela string

const (
	ParcelaDebito           TipoParcela = "DEBITO"
	ParcelaCusta            TipoParcela = "CUSTA_PROCESSUAL"
	ParcelaDespesa          TipoParcela = "DESPESA_PROCESSUAL"
	ParcelaPagamentoParcial TipoParcela = "PAGAMENTO_PARCIAL"
)

// TipoJuros define a regra legal ou convencional aplicável
type TipoJuros string

const (
	JurosConvencionalMensal TipoJuros = "CONVENCIONAL_MENSAL" // Ex: 1.0% a.m.
	JurosTaxaLegal          TipoJuros = "TAXA_LEGAL"          // Lei 14.905/2024: SELIC menos IPCA
	JurosSelic              TipoJuros = "SELIC"               // Selic pura (juros + correção)
	JurosNenhum             TipoJuros = "SEM_JUROS"           // Apenas correção
)

// ModoCapitalizacao define a forma de acumulação dos juros
type ModoCapitalizacao string

const (
	CapitalizacaoSimples  ModoCapitalizacao = "SIMPLES"
	CapitalizacaoComposta ModoCapitalizacao = "COMPOSTO"
)

// IndiceInfo representa um indexador econômico cadastrado
type IndiceInfo struct {
	Codigo             string `json:"codigo"`
	Nome               string `json:"nome"`
	Fonte              string `json:"fonte"`
	Periodicidade      string `json:"periodicidade"`
	DataInicioDisponivel string `json:"data_inicio_disponivel"`
	DataFimDisponivel    string `json:"data_fim_disponivel"`
	Descricao          string `json:"descricao"`
}

// PontoSerie representa um ponto na série histórica do índice
type PontoSerie struct {
	DataReferencia      time.Time       `json:"data_referencia"`
	DataReferenciaTexto string          `json:"data_referencia_texto"` // Ex: "2024-01"
	ValorPercentual     decimal.Decimal `json:"valor_percentual"`      // Variação percentual do mês
	FatorMultiplicador  decimal.Decimal `json:"fator_multiplicador"`   // 1 + (taxa/100)
	NumeroIndice        decimal.Decimal `json:"numero_indice"`         // Número-índice acumulado
}

// ParcelaItem representa uma entrada individual de valor a ser liquidada
type ParcelaItem struct {
	NumeroItem      int             `json:"numero_item"`
	Descricao       string          `json:"descricao"`
	Tipo            TipoParcela     `json:"tipo"`
	DataVencimento  string          `json:"data_vencimento"` // Formato "AAAA-MM-DD" ou "DD/MM/AAAA"
	DataInicioJuros string          `json:"data_inicio_juros,omitempty"` // Se vazia, usa DataVencimento
	ValorOriginal   decimal.Decimal `json:"valor_original"`
}

// JurosMoraConfig define os parâmetros de juros moratórios
type JurosMoraConfig struct {
	Tipo              TipoJuros         `json:"tipo"`               // CONVENCIONAL_MENSAL, TAXA_LEGAL, SELIC, SEM_JUROS
	PercentualMensal  decimal.Decimal   `json:"percentual_mensal"`  // Ex: 1.0 para 1% a.m.
	Modo              ModoCapitalizacao `json:"modo"`               // SIMPLES ou COMPOSTO
	DataInicioGeral   string            `json:"data_inicio_geral,omitempty"` // Se preenchido, sobrepõe a data de cada parcela (ex: data da citação)
	AplicarProRataDie bool              `json:"aplicar_prorata_die"`
}

// JurosCompensatoriosConfig define juros compensatórios / remuneratórios (opcional)
type JurosCompensatoriosConfig struct {
	Ativo            bool              `json:"ativo"`
	PercentualMensal decimal.Decimal   `json:"percentual_mensal"`
	Modo             ModoCapitalizacao `json:"modo"`
	DataInicio       string            `json:"data_inicio,omitempty"`
	DataFim          string            `json:"data_fim,omitempty"`
}

// FaseExecucaoConfig reúne encargos do Art. 523 CPC e honorários sucumbenciais
type FaseExecucaoConfig struct {
	AplicarArt523Multa                  bool            `json:"aplicar_art_523_multa"`                  // Multa de 10%
	AplicarArt523Honorarios             bool            `json:"aplicar_art_523_honorarios"`             // Honorários executivos de 10%
	HonorariosSucumbenciaisPercentual   decimal.Decimal `json:"honorarios_sucumbenciais_percentual"`   // Ex: 10% a 20%
	HonorariosSucumbenciaisValorFixo    decimal.Decimal `json:"honorarios_sucumbenciais_valor_fixo"`    // Valor fixo em R$
	MultaMoratoriaPercentual            decimal.Decimal `json:"multa_moratoria_percentual"`            // Ex: 2% a 10% contratual/moratória
}

// CalculoJudicialRequest é o payload completo de requisição de cálculo
type CalculoJudicialRequest struct {
	NumeroProcesso      string                     `json:"numero_processo,omitempty"`
	DataAtualizacao     string                     `json:"data_atualizacao"` // Data ad quem ("AAAA-MM-DD" ou "DD/MM/AAAA")
	IndexadorCodigo     string                     `json:"indexador_codigo"` // Ex: "TJSP", "IPCA", "INPC", "IGPM"
	AplicarDeflacao     bool                       `json:"aplicar_deflacao"` // Se false, variações negativas viram 0
	JurosMora           JurosMoraConfig            `json:"juros_mora"`
	JurosCompensatorios JurosCompensatoriosConfig  `json:"juros_compensatorios,omitempty"`
	FaseExecucao        FaseExecucaoConfig         `json:"fase_execucao"`
	Parcelas            []ParcelaItem              `json:"parcelas"`
}

// ItemLiquidado representa a memória analítica de cálculo de uma parcela
type ItemLiquidado struct {
	NumeroItem              int             `json:"numero_item"`
	Descricao               string          `json:"descricao"`
	Tipo                    TipoParcela     `json:"tipo"`
	DataVencimento          string          `json:"data_vencimento"`
	DataInicioJuros         string          `json:"data_inicio_juros"`
	ValorOriginal           decimal.Decimal `json:"valor_original"`
	FatorCorrecao           decimal.Decimal `json:"fator_correcao"`
	ValorCorrigido          decimal.Decimal `json:"valor_corrigido"`
	TaxaJurosAcumulada      decimal.Decimal `json:"taxa_juros_acumulada_percentual"` // Ex: 45.32%
	ValorJurosMora          decimal.Decimal `json:"valor_juros_mora"`
	ValorJurosCompensatorios decimal.Decimal `json:"valor_juros_compensatorios"`
	ValorMultaMoratoria     decimal.Decimal `json:"valor_multa_moratoria"`
	TotalItem               decimal.Decimal `json:"total_item"`
	Observacoes             string          `json:"observacoes,omitempty"`
}

// ResumoCalculo consolida todos os totais e encargos finais da liquidação
type ResumoCalculo struct {
	DataAtualizacao              string          `json:"data_atualizacao"`
	IndexadorNome                string          `json:"indexador_nome"`
	TotalPrincipalOriginal       decimal.Decimal `json:"total_principal_original"`
	TotalPrincipalCorrigido      decimal.Decimal `json:"total_principal_corrigido"`
	TotalCorrecaoMonetaria       decimal.Decimal `json:"total_correcao_monetaria"`
	TotalJurosMora               decimal.Decimal `json:"total_juros_mora"`
	TotalJurosCompensatorios     decimal.Decimal `json:"total_juros_compensatorios"`
	TotalMultaMoratoria          decimal.Decimal `json:"total_multa_moratoria"`
	TotalCustasEDespesas         decimal.Decimal `json:"total_custas_e_despesas"`
	TotalAbatimentos             decimal.Decimal `json:"total_abatimentos"`
	SubtotalCondenacao           decimal.Decimal `json:"subtotal_condenacao"`
	MultaArt523CPC               decimal.Decimal `json:"multa_art_523_cpc"`
	HonorariosArt523CPC          decimal.Decimal `json:"honorarios_art_523_cpc"`
	HonorariosSucumbenciais      decimal.Decimal `json:"honorarios_sucumbenciais"`
	TotalGeralDevido             decimal.Decimal `json:"total_geral_devido"`
}

// CalculoJudicialResponse é o resultado estruturado retornado pela API
type CalculoJudicialResponse struct {
	Sucesso          bool            `json:"sucesso"`
	Mensagem         string          `json:"mensagem,omitempty"`
	Resumo           ResumoCalculo   `json:"resumo"`
	ItensLiquidados  []ItemLiquidado `json:"itens_liquidados"`
	NotasMetodologicas []string      `json:"notas_metodologicas"`
}
