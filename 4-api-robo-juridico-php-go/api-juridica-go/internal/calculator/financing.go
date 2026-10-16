package calculator

import (
	"fmt"
	"math"
	"strings"

	"github.com/shopspring/decimal"
)

// CalcularRevisaoFinanciamento realiza o cálculo pericial revisional de contratos bancários
// comparando o sistema contratado (ex: Tabela Price capitalizada) com o sistema ou taxa de mercado recalculada
// nos sistemas PRICE, SAC ou GAUSS (juros simples).
func CalcularRevisaoFinanciamento(req CalculoFinanciamentoRequest) (*CalculoFinanciamentoResponse, error) {
	if req.NumeroParcelas == 0 && req.QuantidadeParcelas > 0 {
		req.NumeroParcelas = req.QuantidadeParcelas
	}
	if req.TaxaContratadaMensal.IsZero() && !req.TaxaMensalContratada.IsZero() {
		req.TaxaContratadaMensal = req.TaxaMensalContratada
	}
	if req.TaxaMercadoMensal.IsZero() && !req.TaxaMensalMercado.IsZero() {
		req.TaxaMercadoMensal = req.TaxaMensalMercado
	}
	if !req.RestituicaoEmDobro && req.RepeticaoEmDobro {
		req.RestituicaoEmDobro = true
	}

	if req.ValorFinanciado.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("valor financiado deve ser maior que zero")
	}
	if req.NumeroParcelas <= 0 || req.NumeroParcelas > 600 {
		return nil, fmt.Errorf("numero de parcelas deve estar entre 1 e 600")
	}
	if req.TaxaContratadaMensal.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("taxa contratada mensal invalida")
	}

	sistema := SistemaAmortizacao(strings.ToUpper(strings.TrimSpace(string(req.SistemaAmortizacao))))
	if sistema == "" {
		sistema = SistemaPrice
	}
	if sistema != SistemaPrice && sistema != SistemaSAC && sistema != SistemaGauss {
		return nil, fmt.Errorf("sistema de amortizacao invalido: use PRICE, SAC ou GAUSS")
	}

	taxaRevisada := req.TaxaMercadoMensal
	if taxaRevisada.LessThanOrEqual(decimal.Zero) {
		taxaRevisada = req.TaxaContratadaMensal
	}

	// Prestações no contrato original (sempre Tabela Price com taxa contratada para comparação bancária padrão)
	prestacaoContratada := calcularPrestacaoPrice(req.ValorFinanciado, req.TaxaContratadaMensal, req.NumeroParcelas)

	var tabela []ParcelaFinanciamentoItem
	var prestacaoRevisadaInicial decimal.Decimal

	switch sistema {
	case SistemaPrice:
		tabela, prestacaoRevisadaInicial = gerarTabelaPrice(req.ValorFinanciado, taxaRevisada, req.NumeroParcelas, prestacaoContratada)
	case SistemaSAC:
		tabela, prestacaoRevisadaInicial = gerarTabelaSAC(req.ValorFinanciado, taxaRevisada, req.NumeroParcelas, prestacaoContratada)
	case SistemaGauss:
		tabela, prestacaoRevisadaInicial = gerarTabelaGauss(req.ValorFinanciado, taxaRevisada, req.NumeroParcelas, prestacaoContratada)
	}

	totalContratado := prestacaoContratada.Mul(decimal.NewFromInt(int64(req.NumeroParcelas))).Round(2)
	totalRevisado := decimal.Zero
	diferencaTotal := decimal.Zero

	for _, item := range tabela {
		totalRevisado = totalRevisado.Add(item.PrestacaoRevisada)
		diferencaTotal = diferencaTotal.Add(item.DiferencaParcela)
	}

	indebitoCobrado := diferencaTotal.Round(2)
	if req.RestituicaoEmDobro && indebitoCobrado.IsPositive() {
		indebitoCobrado = indebitoCobrado.Mul(decimal.NewFromInt(2)).Round(2)
	}

	notas := []string{
		fmt.Sprintf("Revisão contratual calculada pelo método %s sobre capital financiado de R$ %s em %d prestações.", sistema, req.ValorFinanciado.StringFixed(2), req.NumeroParcelas),
		fmt.Sprintf("Taxa contratada: %s%% a.m. | Taxa revisada/mercado: %s%% a.m.", req.TaxaContratadaMensal.StringFixed(4), taxaRevisada.StringFixed(4)),
	}
	if sistema == SistemaGauss {
		notas = append(notas, "Método de Gauss aplicado com distribuição linear de juros simples e amortização ponderada sem anatocismo (Súmula 121 STF).")
	}
	if req.RestituicaoEmDobro {
		notas = append(notas, "Indébito apurado em dobro nos termos do Art. 42, parágrafo único, do Código de Defesa do Consumidor (CDC).")
	}

	return &CalculoFinanciamentoResponse{
		Sucesso:                  true,
		SistemaAmortizacao:       sistema,
		ValorFinanciado:          req.ValorFinanciado.Round(2),
		NumeroParcelas:           req.NumeroParcelas,
		TaxaContratadaMensal:     req.TaxaContratadaMensal,
		TaxaMercadoMensal:        taxaRevisada,
		PrestacaoContratada:      prestacaoContratada,
		PrestacaoRevisadaInicial: prestacaoRevisadaInicial,
		TotalPagoContratado:      totalContratado,
		TotalDevidoRevisado:      totalRevisado.Round(2),
		DiferencaTotalSimples:    diferencaTotal.Round(2),
		IndebitoCobrado:          indebitoCobrado,
		TabelaEvolucao:           tabela,
		NotasMetodologicas:       notas,
	}, nil
}

func calcularPrestacaoPrice(pv decimal.Decimal, taxaMensalPct decimal.Decimal, n int) decimal.Decimal {
	if taxaMensalPct.IsZero() {
		return pv.Div(decimal.NewFromInt(int64(n))).Round(2)
	}
	i, _ := taxaMensalPct.Div(decimal.NewFromInt(100)).Float64()
	pvFloat, _ := pv.Float64()
	fator := math.Pow(1+i, float64(n))
	pmt := pvFloat * (i * fator) / (fator - 1)
	return decimal.NewFromFloat(pmt).Round(2)
}

func gerarTabelaPrice(pv decimal.Decimal, taxaPct decimal.Decimal, n int, pmtContratada decimal.Decimal) ([]ParcelaFinanciamentoItem, decimal.Decimal) {
	pmtRevisada := calcularPrestacaoPrice(pv, taxaPct, n)
	taxa := taxaPct.Div(decimal.NewFromInt(100))
	saldo := pv
	itens := make([]ParcelaFinanciamentoItem, 0, n)

	for k := 1; k <= n; k++ {
		juros := saldo.Mul(taxa).Round(2)
		amort := pmtRevisada.Sub(juros).Round(2)
		saldo = saldo.Sub(amort).Round(2)
		if k == n && !saldo.IsZero() {
			amort = amort.Add(saldo)
			saldo = decimal.Zero
		}
		dif := pmtContratada.Sub(pmtRevisada).Round(2)
		itens = append(itens, ParcelaFinanciamentoItem{
			NumeroParcela:       k,
			PrestacaoContratada: pmtContratada,
			PrestacaoRevisada:   pmtRevisada,
			JurosRevisado:       juros,
			AmortizacaoRevisada: amort,
			SaldoDevedor:        decimal.Max(saldo, decimal.Zero),
			DiferencaParcela:    dif,
		})
	}
	return itens, pmtRevisada
}

func gerarTabelaSAC(pv decimal.Decimal, taxaPct decimal.Decimal, n int, pmtContratada decimal.Decimal) ([]ParcelaFinanciamentoItem, decimal.Decimal) {
	amortConstante := pv.Div(decimal.NewFromInt(int64(n))).Round(2)
	taxa := taxaPct.Div(decimal.NewFromInt(100))
	saldo := pv
	itens := make([]ParcelaFinanciamentoItem, 0, n)
	primeiraPrestacao := decimal.Zero

	for k := 1; k <= n; k++ {
		juros := saldo.Mul(taxa).Round(2)
		amort := amortConstante
		if k == n {
			amort = saldo
		}
		pmtRevisada := amort.Add(juros).Round(2)
		if k == 1 {
			primeiraPrestacao = pmtRevisada
		}
		saldo = saldo.Sub(amort).Round(2)
		dif := pmtContratada.Sub(pmtRevisada).Round(2)
		itens = append(itens, ParcelaFinanciamentoItem{
			NumeroParcela:       k,
			PrestacaoContratada: pmtContratada,
			PrestacaoRevisada:   pmtRevisada,
			JurosRevisado:       juros,
			AmortizacaoRevisada: amort,
			SaldoDevedor:        decimal.Max(saldo, decimal.Zero),
			DiferencaParcela:    dif,
		})
	}
	return itens, primeiraPrestacao
}

func gerarTabelaGauss(pv decimal.Decimal, taxaPct decimal.Decimal, n int, pmtContratada decimal.Decimal) ([]ParcelaFinanciamentoItem, decimal.Decimal) {
	// Fórmula de Gauss: PMT = [2 * PV * (1 + n*i)] / [n * (2 + i*(n - 1))]
	i := taxaPct.Div(decimal.NewFromInt(100))
	nDec := decimal.NewFromInt(int64(n))
	dois := decimal.NewFromInt(2)
	um := decimal.NewFromInt(1)

	numerador := dois.Mul(pv).Mul(um.Add(nDec.Mul(i)))
	denominador := nDec.Mul(dois.Add(i.Mul(nDec.Sub(um))))
	pmtGauss := numerador.Div(denominador).Round(2)

	// Juros total no método de Gauss: J_total = (n * PMT) - PV
	jurosTotal := pmtGauss.Mul(nDec).Sub(pv)
	// Soma dos pesos (1 + 2 + ... + n) = n*(n+1)/2
	somaPesos := decimal.NewFromInt(int64(n * (n + 1) / 2))
	indicePonderadoJuros := jurosTotal.Div(somaPesos)

	saldo := pv
	itens := make([]ParcelaFinanciamentoItem, 0, n)

	for k := 1; k <= n; k++ {
		peso := decimal.NewFromInt(int64(n - k + 1))
		juros := indicePonderadoJuros.Mul(peso).Round(2)
		amort := pmtGauss.Sub(juros).Round(2)
		saldo = saldo.Sub(amort).Round(2)
		if k == n && !saldo.IsZero() {
			amort = amort.Add(saldo)
			saldo = decimal.Zero
		}
		dif := pmtContratada.Sub(pmtGauss).Round(2)
		itens = append(itens, ParcelaFinanciamentoItem{
			NumeroParcela:       k,
			PrestacaoContratada: pmtContratada,
			PrestacaoRevisada:   pmtGauss,
			JurosRevisado:       juros,
			AmortizacaoRevisada: amort,
			SaldoDevedor:        decimal.Max(saldo, decimal.Zero),
			DiferencaParcela:    dif,
		})
	}
	return itens, pmtGauss
}
