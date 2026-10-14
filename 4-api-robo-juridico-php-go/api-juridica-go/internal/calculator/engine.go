package calculator

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// EngineCalculo orquestra a liquidação judicial completa de débitos e créditos
type EngineCalculo struct {
	repo *RepositorioIndices
}

func NovoEngineCalculo(repo *RepositorioIndices) *EngineCalculo {
	if repo == nil {
		repo = NovoRepositorioIndices()
	}
	return &EngineCalculo{repo: repo}
}

func (e *EngineCalculo) Repositorio() *RepositorioIndices {
	return e.repo
}

// Executar realiza a liquidação analítica e sintética das parcelas informadas
func (e *EngineCalculo) Executar(req CalculoJudicialRequest) (*CalculoJudicialResponse, error) {
	if len(req.Parcelas) == 0 {
		return nil, fmt.Errorf("nenhuma parcela informada para liquidacao")
	}

	dataAtualizacao, err := ParseDataFlexivel(req.DataAtualizacao)
	if err != nil {
		return nil, fmt.Errorf("data de atualizacao invalida ('%s'): %w", req.DataAtualizacao, err)
	}

	indiceInfo, err := e.repo.ObterIndice(req.IndexadorCodigo)
	if err != nil {
		return nil, err
	}

	// Data inicial de juros global se especificada
	var dataInicioJurosGlobal *time.Time
	if req.JurosMora.DataInicioGeral != "" {
		dt, err := ParseDataFlexivel(req.JurosMora.DataInicioGeral)
		if err == nil {
			dataInicioJurosGlobal = &dt
		}
	}

	itensLiquidados := make([]ItemLiquidado, 0, len(req.Parcelas))

	totalPrincipalOriginal := decimal.Zero
	totalPrincipalCorrigido := decimal.Zero
	totalJurosMora := decimal.Zero
	totalJurosCompensatorios := decimal.Zero
	totalMultaMoratoria := decimal.Zero
	totalCustasEDespesas := decimal.Zero
	totalAbatimentos := decimal.Zero

	for idx, p := range req.Parcelas {
		dtVencto, err := ParseDataFlexivel(p.DataVencimento)
		if err != nil {
			return nil, fmt.Errorf("item %d: data de vencimento invalida ('%s'): %w", idx+1, p.DataVencimento, err)
		}

		dtJuros := dtVencto
		if dataInicioJurosGlobal != nil {
			dtJuros = *dataInicioJurosGlobal
		} else if p.DataInicioJuros != "" {
			dtJ, err := ParseDataFlexivel(p.DataInicioJuros)
			if err == nil {
				dtJuros = dtJ
			}
		}

		// 1. Correção Monetária
		fatorCorrecao, err := e.repo.ObterFatorAcumulado(req.IndexadorCodigo, dtVencto, dataAtualizacao, req.AplicarDeflacao)
		if err != nil {
			fatorCorrecao = decimal.NewFromInt(1)
		}

		valorOriginal := p.ValorOriginal
		valorCorrigido := valorOriginal.Mul(fatorCorrecao).Round(2)

		taxaJurosAcumulada := decimal.Zero
		valorJuros := decimal.Zero
		valorMultaItem := decimal.Zero
		obs := ""

		switch p.Tipo {
		case ParcelaDebito:
			// Débitos sofrem juros e multa
			taxaJuros, vJuros, obsJuros, _ := CalcularJurosMoratorios(
				e.repo,
				req.JurosMora,
				dtJuros,
				dataAtualizacao,
				valorCorrigido,
			)
			taxaJurosAcumulada = taxaJuros
			valorJuros = vJuros
			obs = obsJuros

			if !req.FaseExecucao.MultaMoratoriaPercentual.IsZero() {
				valorMultaItem = valorCorrigido.Mul(req.FaseExecucao.MultaMoratoriaPercentual).Div(decimal.NewFromInt(100)).Round(2)
			}

			totalPrincipalOriginal = totalPrincipalOriginal.Add(valorOriginal)
			totalPrincipalCorrigido = totalPrincipalCorrigido.Add(valorCorrigido)
			totalJurosMora = totalJurosMora.Add(valorJuros)
			totalMultaMoratoria = totalMultaMoratoria.Add(valorMultaItem)

		case ParcelaCusta, ParcelaDespesa:
			// Custas sofrem apenas correção monetária
			obs = "Atualização monetária sem incidência de juros ou multa"
			totalCustasEDespesas = totalCustasEDespesas.Add(valorCorrigido)

		case ParcelaPagamentoParcial:
			// Abatimento/Pagamento Parcial (subtrai do saldo)
			if valorOriginal.IsPositive() {
				valorOriginal = valorOriginal.Neg()
				valorCorrigido = valorCorrigido.Neg()
			}
			obs = "Pagamento parcial deduzido da conta geral"
			totalAbatimentos = totalAbatimentos.Add(valorCorrigido.Abs())
		}

		totalItem := valorCorrigido.Add(valorJuros).Add(valorMultaItem)

		numItem := p.NumeroItem
		if numItem == 0 {
			numItem = idx + 1
		}

		itensLiquidados = append(itensLiquidados, ItemLiquidado{
			NumeroItem:              numItem,
			Descricao:               p.Descricao,
			Tipo:                    p.Tipo,
			DataVencimento:          dtVencto.Format("02/01/2006"),
			DataInicioJuros:         dtJuros.Format("02/01/2006"),
			ValorOriginal:           valorOriginal,
			FatorCorrecao:           fatorCorrecao,
			ValorCorrigido:          valorCorrigido,
			TaxaJurosAcumulada:      taxaJurosAcumulada,
			ValorJurosMora:          valorJuros,
			ValorJurosCompensatorios: decimal.Zero,
			ValorMultaMoratoria:     valorMultaItem,
			TotalItem:               totalItem,
			Observacoes:             obs,
		})
	}

	// 2. Consolidação de Subtotais
	totalCorrecao := totalPrincipalCorrigido.Sub(totalPrincipalOriginal)
	subtotalCondenacao := totalPrincipalCorrigido.
		Add(totalJurosMora).
		Add(totalMultaMoratoria).
		Add(totalCustasEDespesas).
		Sub(totalAbatimentos)

	// 3. Encargos da Fase de Execução (Art. 523 CPC e Honorários Sucumbenciais)
	multaArt523, honorariosArt523, honorariosSucumb := CalcularEncargosExecucao(
		req.FaseExecucao,
		subtotalCondenacao,
	)

	totalGeralDevido := subtotalCondenacao.
		Add(multaArt523).
		Add(honorariosArt523).
		Add(honorariosSucumb)

	// 4. Notas Metodológicas e Fundamentação
	notas := []string{
		fmt.Sprintf("Valores atualizados monetariamente até %s com base no indexador %s (%s).", dataAtualizacao.Format("02/01/2006"), indiceInfo.Nome, indiceInfo.Fonte),
	}
	if req.JurosMora.Tipo == JurosTaxaLegal {
		notas = append(notas, "Juros moratórios apurados em regime híbrido pela Lei Federal nº 14.905/2024 (Art. 406 do Código Civil).")
	} else if req.JurosMora.Tipo == JurosConvencionalMensal {
		modo := "simples"
		if req.JurosMora.Modo == CapitalizacaoComposta {
			modo = "compostos"
		}
		notas = append(notas, fmt.Sprintf("Juros moratórios %s fixados à razão de %s%% ao mês.", modo, req.JurosMora.PercentualMensal.String()))
	}

	if req.FaseExecucao.AplicarArt523Multa || req.FaseExecucao.AplicarArt523Honorarios {
		notas = append(notas, "Incidência dos encargos previstos no Art. 523, § 1º do Código de Processo Civil (Cumprimento de Sentença).")
	}

	return &CalculoJudicialResponse{
		Sucesso:  true,
		Mensagem: "Cálculo judicial executado com sucesso.",
		Resumo: ResumoCalculo{
			DataAtualizacao:          dataAtualizacao.Format("02/01/2006"),
			IndexadorNome:            indiceInfo.Nome,
			TotalPrincipalOriginal:   totalPrincipalOriginal.Round(2),
			TotalPrincipalCorrigido:  totalPrincipalCorrigido.Round(2),
			TotalCorrecaoMonetaria:   totalCorrecao.Round(2),
			TotalJurosMora:           totalJurosMora.Round(2),
			TotalJurosCompensatorios: totalJurosCompensatorios.Round(2),
			TotalMultaMoratoria:      totalMultaMoratoria.Round(2),
			TotalCustasEDespesas:     totalCustasEDespesas.Round(2),
			TotalAbatimentos:         totalAbatimentos.Round(2),
			SubtotalCondenacao:       subtotalCondenacao.Round(2),
			MultaArt523CPC:           multaArt523.Round(2),
			HonorariosArt523CPC:      honorariosArt523.Round(2),
			HonorariosSucumbenciais:  honorariosSucumb.Round(2),
			TotalGeralDevido:         totalGeralDevido.Round(2),
		},
		ItensLiquidados:    itensLiquidados,
		NotasMetodologicas: notas,
	}, nil
}
