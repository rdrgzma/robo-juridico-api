package calculator

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestEngineCalculo_CenarioTradicionalDrCalc(t *testing.T) {
	repo := NovoRepositorioIndices()
	engine := NovoEngineCalculo(repo)

	// Cenário equivalente ao submetido no DrCalc:
	// Parcela: R$ 10.000,00
	// Vencimento: 15/01/2023
	// Data de atualização: 14/10/2026
	// Indexador: TJSP
	// Juros: 1% a.m. compostos
	// Multa: 10%
	// Art. 523 CPC: 10% multa + 10% honorários
	req := CalculoJudicialRequest{
		DataAtualizacao: "2026-10-14",
		IndexadorCodigo: "TJSP",
		AplicarDeflacao: true,
		JurosMora: JurosMoraConfig{
			Tipo:              JurosConvencionalMensal,
			PercentualMensal:  decimal.NewFromInt(1),
			Modo:              CapitalizacaoComposta,
			AplicarProRataDie: true,
		},
		FaseExecucao: FaseExecucaoConfig{
			AplicarArt523Multa:                  true,
			AplicarArt523Honorarios:             true,
			MultaMoratoriaPercentual:            decimal.NewFromInt(10),
			HonorariosSucumbenciaisValorFixo:    decimal.NewFromInt(10), // R$ 10 de honorários adicionais
		},
		Parcelas: []ParcelaItem{
			{
				NumeroItem:     1,
				Descricao:      "Parcela 01",
				Tipo:           ParcelaDebito,
				DataVencimento: "2023-01-15",
				ValorOriginal:  decimal.NewFromInt(10000),
			},
		},
	}

	resp, err := engine.Executar(req)
	if err != nil {
		t.Fatalf("Erro ao executar calculo: %v", err)
	}

	if !resp.Sucesso {
		t.Fatalf("Calculo retornou sucesso = false")
	}

	if len(resp.ItensLiquidados) != 1 {
		t.Fatalf("Esperado 1 item liquidado, obtido %d", len(resp.ItensLiquidados))
	}

	item := resp.ItensLiquidados[0]

	// Validação de Correção Monetária:
	// R$ 10.000 * (104.281145 / 88.490382) ~ 11.784,46
	fCorrigido, _ := item.ValorCorrigido.Float64()
	if fCorrigido < 11780.0 || fCorrigido > 11790.0 {
		t.Errorf("Valor corrigido esperado proximo de 11784.46, obtido: %f", fCorrigido)
	}

	// Validação de Totais
	resumo := resp.Resumo
	fTotalGeral, _ := resumo.TotalGeralDevido.Float64()
	if fTotalGeral <= 0 {
		t.Errorf("Total geral deve ser positivo, obtido %f", fTotalGeral)
	}

	t.Logf("Resultado do Teste:")
	t.Logf("  Principal Corrigido: R$ %s", resumo.TotalPrincipalCorrigido.String())
	t.Logf("  Juros Moratórios: R$ %s", resumo.TotalJurosMora.String())
	t.Logf("  Multa Moratória: R$ %s", resumo.TotalMultaMoratoria.String())
	t.Logf("  Subtotal Condenação: R$ %s", resumo.SubtotalCondenacao.String())
	t.Logf("  Multa Art. 523 (10%%): R$ %s", resumo.MultaArt523CPC.String())
	t.Logf("  Honorários Art. 523 (10%%): R$ %s", resumo.HonorariosArt523CPC.String())
	t.Logf("  TOTAL GERAL DEVIDO: R$ %s", resumo.TotalGeralDevido.String())
}

func TestEngineCalculo_Lei14905(t *testing.T) {
	repo := NovoRepositorioIndices()
	engine := NovoEngineCalculo(repo)

	// Cenário com a Lei 14.905/2024
	req := CalculoJudicialRequest{
		DataAtualizacao: "2026-10-14",
		IndexadorCodigo: "IPCA",
		AplicarDeflacao: true,
		JurosMora: JurosMoraConfig{
			Tipo:              JurosTaxaLegal, // Lei 14.905
			PercentualMensal:  decimal.NewFromInt(1),
			Modo:              CapitalizacaoSimples,
			AplicarProRataDie: true,
		},
		FaseExecucao: FaseExecucaoConfig{
			AplicarArt523Multa:      true,
			AplicarArt523Honorarios: true,
		},
		Parcelas: []ParcelaItem{
			{
				NumeroItem:     1,
				Descricao:      "Indenização Cível",
				Tipo:           ParcelaDebito,
				DataVencimento: "2024-01-10",
				ValorOriginal:  decimal.NewFromInt(50000),
			},
			{
				NumeroItem:     2,
				Descricao:      "Custas Iniciais",
				Tipo:           ParcelaCusta,
				DataVencimento: "2024-01-15",
				ValorOriginal:  decimal.NewFromInt(500),
			},
			{
				NumeroItem:     3,
				Descricao:      "Pagamento Parcial do Devedor",
				Tipo:           ParcelaPagamentoParcial,
				DataVencimento: "2025-06-01",
				ValorOriginal:  decimal.NewFromInt(-10000),
			},
		},
	}

	resp, err := engine.Executar(req)
	if err != nil {
		t.Fatalf("Erro ao executar calculo Lei 14.905: %v", err)
	}

	if len(resp.ItensLiquidados) != 3 {
		t.Fatalf("Esperado 3 itens liquidados, obtido %d", len(resp.ItensLiquidados))
	}

	// O item de custas não deve ter juros
	custas := resp.ItensLiquidados[1]
	if !custas.ValorJurosMora.IsZero() {
		t.Errorf("Custas nao podem sofrer incidencia de juros, obtido: %s", custas.ValorJurosMora.String())
	}

	resumo := resp.Resumo
	t.Logf("Resultado Lei 14.905:")
	t.Logf("  Principal Original: R$ %s", resumo.TotalPrincipalOriginal.String())
	t.Logf("  Principal Corrigido: R$ %s", resumo.TotalPrincipalCorrigido.String())
	t.Logf("  Juros Lei 14.905: R$ %s", resumo.TotalJurosMora.String())
	t.Logf("  Custas Corrigidas: R$ %s", resumo.TotalCustasEDespesas.String())
	t.Logf("  Abatimentos: R$ %s", resumo.TotalAbatimentos.String())
	t.Logf("  Subtotal: R$ %s", resumo.SubtotalCondenacao.String())
	t.Logf("  Total Geral: R$ %s", resumo.TotalGeralDevido.String())
}
