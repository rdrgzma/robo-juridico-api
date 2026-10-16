package calculator

import (
	"time"

	"github.com/shopspring/decimal"
)

// DataVigenciaLei14905 marca a transição da taxa de juros legais no Código Civil (28 de agosto de 2024)
var DataVigenciaLei14905 = time.Date(2024, 8, 28, 0, 0, 0, 0, time.UTC)

// CalcularJurosMoratorios processa a taxa e o valor dos juros de mora aplicáveis a um valor corrigido
func CalcularJurosMoratorios(
	repo *RepositorioIndices,
	cfg JurosMoraConfig,
	dataInicio, dataFim time.Time,
	valorBase decimal.Decimal,
) (taxaAcumuladaPct, valorJuros decimal.Decimal, obs string, err error) {

	if cfg.Tipo == JurosNenhum || dataInicio.After(dataFim) || dataInicio.Equal(dataFim) {
		return decimal.Zero, decimal.Zero, "Sem incidência de juros", nil
	}

	cento := decimal.NewFromInt(100)
	um := decimal.NewFromInt(1)

	switch cfg.Tipo {
	case JurosConvencionalMensal:
		taxaMensal := cfg.PercentualMensal
		if taxaMensal.IsZero() {
			taxaMensal = decimal.NewFromInt(1) // Padrão: 1% a.m.
		}

		mesesCheios, diasRestantes := calcularMesesEDias(dataInicio, dataFim)
		fatorMeses := decimal.NewFromInt(int64(mesesCheios))
		if cfg.AplicarProRataDie && diasRestantes > 0 {
			fatorDias := decimal.NewFromInt(int64(diasRestantes)).Div(decimal.NewFromInt(30))
			fatorMeses = fatorMeses.Add(fatorDias)
		}

		if cfg.Modo == CapitalizacaoComposta {
			// (1 + i)^n - 1
			fatorTaxa := um.Add(taxaMensal.Div(cento))
			fatorAcumulado := decimalPow(fatorTaxa, fatorMeses)
			taxaAcumuladaPct = fatorAcumulado.Sub(um).Mul(cento).Round(4)
		} else {
			// Juros Simples: i * n
			taxaAcumuladaPct = taxaMensal.Mul(fatorMeses).Round(4)
		}

		valorJuros = valorBase.Mul(taxaAcumuladaPct.Div(cento)).Round(2)
		return taxaAcumuladaPct, valorJuros, "Juros moratórios convencionais", nil

	case JurosTaxaLegal, JurosLei14905:
		// Aplicação híbrida da Lei 14.905/2024:
		// Se dataInicio < 28/08/2024: período 1 com 1% a.m. simples até 28/08/2024
		// A partir de 28/08/2024: soma da Taxa Legal mensal divulgada pelo Bacen
		taxaTotal := decimal.Zero

		// Período 1: Pré-Lei 14.905
		if dataInicio.Before(DataVigenciaLei14905) {
			limiteP1 := dataFim
			if limiteP1.After(DataVigenciaLei14905) {
				limiteP1 = DataVigenciaLei14905
			}
			m1, d1 := calcularMesesEDias(dataInicio, limiteP1)
			mesesP1 := decimal.NewFromInt(int64(m1)).Add(decimal.NewFromInt(int64(d1)).Div(decimal.NewFromInt(30)))
			taxaP1 := cfg.PercentualMensal
			if taxaP1.IsZero() {
				taxaP1 = decimal.NewFromInt(1) // 1% ao mês no regime anterior
			}
			taxaTotal = taxaTotal.Add(taxaP1.Mul(mesesP1))
		}

		// Período 2: Pós-Lei 14.905 (mês a mês pela TAXA_LEGAL)
		if dataFim.After(DataVigenciaLei14905) {
			inicioP2 := dataInicio
			if inicioP2.Before(DataVigenciaLei14905) {
				inicioP2 = DataVigenciaLei14905
			}

			cursor := time.Date(inicioP2.Year(), inicioP2.Month(), 1, 0, 0, 0, 0, time.UTC)
			fim := time.Date(dataFim.Year(), dataFim.Month(), 1, 0, 0, 0, 0, time.UTC)

			for !cursor.After(fim) {
				ponto, err := repo.ObterPonto("TAXA_LEGAL", cursor)
				if err == nil {
					taxaTotal = taxaTotal.Add(ponto.ValorPercentual)
				}
				cursor = cursor.AddDate(0, 1, 0)
			}
		}

		taxaAcumuladaPct = taxaTotal.Round(4)
		valorJuros = valorBase.Mul(taxaAcumuladaPct.Div(cento)).Round(2)
		return taxaAcumuladaPct, valorJuros, "Juros pela Taxa Legal (Art. 406 CC / Lei 14.905)", nil

	case JurosSelic:
		// Acumulação simples das taxas mensais da Selic
		taxaTotal := decimal.Zero
		cursor := time.Date(dataInicio.Year(), dataInicio.Month(), 1, 0, 0, 0, 0, time.UTC)
		fim := time.Date(dataFim.Year(), dataFim.Month(), 1, 0, 0, 0, 0, time.UTC)

		for !cursor.After(fim) {
			ponto, err := repo.ObterPonto("SELIC", cursor)
			if err == nil {
				taxaTotal = taxaTotal.Add(ponto.ValorPercentual)
			}
			cursor = cursor.AddDate(0, 1, 0)
		}

		taxaAcumuladaPct = taxaTotal.Round(4)
		valorJuros = valorBase.Mul(taxaAcumuladaPct.Div(cento)).Round(2)
		return taxaAcumuladaPct, valorJuros, "Juros moratórios pela Taxa SELIC", nil
	}

	return decimal.Zero, decimal.Zero, "", nil
}

// CalcularJurosCompensatorios apura os juros compensatórios/remuneratórios sobre o valor corrigido da parcela
func CalcularJurosCompensatorios(
	cfg JurosCompensatoriosConfig,
	dataVencimento, dataAtualizacao time.Time,
	valorBase decimal.Decimal,
) (taxaAcumuladaPct, valorJuros decimal.Decimal, obs string) {
	taxaMensal := cfg.PercentualMensal
	if taxaMensal.LessThanOrEqual(decimal.Zero) {
		taxaMensal = cfg.TaxaMensal
	}
	if !cfg.Ativo || taxaMensal.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, ""
	}

	inicio := dataVencimento
	if cfg.DataInicio != "" {
		if dt, errParse := ParseDataFlexivel(cfg.DataInicio); errParse == nil && !dt.IsZero() {
			inicio = dt
		}
	}

	fim := dataAtualizacao
	if cfg.DataFim != "" {
		if dt, errParse := ParseDataFlexivel(cfg.DataFim); errParse == nil && !dt.IsZero() {
			fim = dt
		}
	}

	if !inicio.Before(fim) {
		return decimal.Zero, decimal.Zero, ""
	}

	cento := decimal.NewFromInt(100)
	um := decimal.NewFromInt(1)

	mesesCheios, diasRestantes := calcularMesesEDias(inicio, fim)
	fatorMeses := decimal.NewFromInt(int64(mesesCheios)).Add(
		decimal.NewFromInt(int64(diasRestantes)).Div(decimal.NewFromInt(30)),
	)

	if cfg.Modo == CapitalizacaoComposta {
		fatorTaxa := um.Add(taxaMensal.Div(cento))
		fatorAcumulado := decimalPow(fatorTaxa, fatorMeses)
		taxaAcumuladaPct = fatorAcumulado.Sub(um).Mul(cento).Round(4)
	} else {
		taxaAcumuladaPct = taxaMensal.Mul(fatorMeses).Round(4)
	}

	valorJuros = valorBase.Mul(taxaAcumuladaPct.Div(cento)).Round(2)
	return taxaAcumuladaPct, valorJuros, "Juros compensatórios"
}

// calcularMesesEDias calcula a distância em meses comerciais e dias entre duas datas
func calcularMesesEDias(inicio, fim time.Time) (meses int, dias int) {
	if inicio.After(fim) {
		return 0, 0
	}

	anos := fim.Year() - inicio.Year()
	meses = int(fim.Month()) - int(inicio.Month()) + (anos * 12)
	dias = fim.Day() - inicio.Day()

	if dias < 0 {
		meses--
		dias += 30
	}

	if meses < 0 {
		meses = 0
	}
	return meses, dias
}

// decimalPow calcula base^expoente com precisão para expoentes inteiros e decimais aproximados
func decimalPow(base, exp decimal.Decimal) decimal.Decimal {
	fBase, _ := base.Float64()
	if fBase <= 0 {
		return decimal.Zero
	}
	intExp := exp.IntPart()
	fracExp := exp.Sub(decimal.NewFromInt(intExp))

	pot := base.Pow(decimal.NewFromInt(intExp))
	if !fracExp.IsZero() {
		fFrac, _ := fracExp.Float64()
		pot = pot.Mul(decimal.NewFromFloat(1.0 + (fBase-1.0)*fFrac))
	}
	return pot
}
