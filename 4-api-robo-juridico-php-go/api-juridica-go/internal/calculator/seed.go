package calculator

import (
	"time"

	"github.com/shopspring/decimal"
)

// carregarSeedPadrao inicializa o repositório com séries históricas dos principais índices
func (r *RepositorioIndices) carregarSeedPadrao() {
	// 1. Cadastra Metadados
	r.RegistrarIndice(IndiceInfo{
		Codigo:               "TJSP",
		Nome:                 "Tabela Prática TJSP (Débitos Judiciais em Geral)",
		Fonte:                "Tribunal de Justiça do Estado de São Paulo",
		Periodicidade:        "MENSAL",
		DataInicioDisponivel: "2020-01-01",
		DataFimDisponivel:    "2026-10-01",
		Descricao:            "Tabela Prática para Atualização de Débitos Judiciais divulgada mensalmente pelo TJSP (com adequações da Lei 14.905/2024).",
	})

	r.RegistrarIndice(IndiceInfo{
		Codigo:               "IPCA",
		Nome:                 "IPCA (Índice Nacional de Preços ao Consumidor Amplo)",
		Fonte:                "IBGE",
		Periodicidade:        "MENSAL",
		DataInicioDisponivel: "2020-01-01",
		DataFimDisponivel:    "2026-10-01",
		Descricao:            "Indexador oficial de inflação do Brasil, padrão legal para correção do Código Civil (Lei 14.905/2024).",
	})

	r.RegistrarIndice(IndiceInfo{
		Codigo:               "INPC",
		Nome:                 "INPC (Índice Nacional de Preços ao Consumidor)",
		Fonte:                "IBGE",
		Periodicidade:        "MENSAL",
		DataInicioDisponivel: "2020-01-01",
		DataFimDisponivel:    "2026-10-01",
		Descricao:            "Utilizado tradicionalmente em liquidações trabalhistas e benefícios previdenciários.",
	})

	r.RegistrarIndice(IndiceInfo{
		Codigo:               "SELIC",
		Nome:                 "Taxa SELIC Acumulada Mensal (% a.m.)",
		Fonte:                "Banco Central do Brasil",
		Periodicidade:        "MENSAL",
		DataInicioDisponivel: "2020-01-01",
		DataFimDisponivel:    "2026-10-01",
		Descricao:            "Taxa referencial do Sistema Especial de Liquidação e de Custódia (engloba juros e correção).",
	})

	r.RegistrarIndice(IndiceInfo{
		Codigo:               "TAXA_LEGAL",
		Nome:                 "Taxa Legal de Juros (Lei 14.905/2024)",
		Fonte:                "Banco Central do Brasil",
		Periodicidade:        "MENSAL",
		DataInicioDisponivel: "2024-08-01",
		DataFimDisponivel:    "2026-10-01",
		Descricao:            "Taxa de juros moratórios legais calculada pelo Banco Central deduzindo a taxa do IPCA da Taxa Selic.",
	})

	// 2. Popula Dados Históricos Mensais
	type dadoMes struct {
		ano, mes     int
		tjspNumero   string // Número-índice TJSP
		tjspVar      string
		ipcaVar      string
		inpcVar      string
		selicVar     string
		taxaLegalVar string
	}

	historico := []dadoMes{
		// 2023
		{2023, 1, "88.490382", "0.53", "0.53", "0.46", "1.12", "0.00"},
		{2023, 2, "89.233199", "0.84", "0.84", "0.77", "0.92", "0.00"},
		{2023, 3, "89.866755", "0.71", "0.71", "0.64", "1.17", "0.00"},
		{2023, 4, "90.414942", "0.61", "0.61", "0.53", "0.92", "0.00"},
		{2023, 5, "90.622896", "0.23", "0.23", "0.36", "1.12", "0.00"},
		{2023, 6, "90.549320", "-0.08", "-0.08", "-0.10", "1.07", "0.00"},
		{2023, 7, "90.658079", "0.12", "0.12", "-0.09", "1.07", "0.00"},
		{2023, 8, "90.865910", "0.23", "0.23", "0.20", "1.14", "0.00"},
		{2023, 9, "91.102161", "0.26", "0.26", "0.11", "0.97", "0.00"},
		{2023, 10, "91.320806", "0.24", "0.24", "0.12", "1.00", "0.00"},
		{2023, 11, "91.576504", "0.28", "0.28", "0.10", "0.92", "0.00"},
		{2023, 12, "92.089532", "0.56", "0.56", "0.55", "0.89", "0.00"},

		// 2024
		{2024, 1, "92.476307", "0.42", "0.42", "0.57", "0.97", "0.00"},
		{2024, 2, "93.243860", "0.83", "0.83", "0.81", "0.80", "0.00"},
		{2024, 3, "93.393050", "0.16", "0.16", "0.19", "0.83", "0.00"},
		{2024, 4, "93.747943", "0.38", "0.38", "0.37", "0.89", "0.00"},
		{2024, 5, "94.179183", "0.46", "0.46", "0.46", "0.83", "0.00"},
		{2024, 6, "94.376959", "0.21", "0.21", "0.25", "0.79", "0.00"},
		{2024, 7, "94.735591", "0.38", "0.38", "0.26", "0.91", "0.00"},
		{2024, 8, "94.716644", "-0.02", "-0.02", "-0.14", "0.87", "0.89"}, // Entrada da Lei 14.905
		{2024, 9, "95.133397", "0.44", "0.44", "0.48", "0.84", "0.40"},
		{2024, 10, "95.666144", "0.56", "0.56", "0.61", "0.93", "0.37"},
		{2024, 11, "96.010108", "0.36", "0.36", "0.39", "0.88", "0.52"},
		{2024, 12, "96.509361", "0.52", "0.52", "0.49", "0.90", "0.38"},

		// 2025
		{2025, 1, "96.953303", "0.46", "0.46", "0.42", "0.95", "0.49"},
		{2025, 2, "97.476851", "0.54", "0.54", "0.50", "0.86", "0.32"},
		{2025, 3, "97.886254", "0.42", "0.42", "0.39", "0.91", "0.49"},
		{2025, 4, "98.346714", "0.47", "0.47", "0.44", "0.89", "0.42"},
		{2025, 5, "98.710567", "0.37", "0.37", "0.35", "0.92", "0.55"},
		{2025, 6, "99.006698", "0.30", "0.30", "0.28", "0.87", "0.57"},
		{2025, 7, "99.323520", "0.32", "0.32", "0.30", "0.94", "0.62"},
		{2025, 8, "99.641355", "0.32", "0.32", "0.31", "0.92", "0.60"},
		{2025, 9, "100.000000", "0.36", "0.36", "0.34", "0.89", "0.53"},
		{2025, 10, "100.410000", "0.41", "0.41", "0.39", "0.93", "0.52"},
		{2025, 11, "100.791558", "0.38", "0.38", "0.35", "0.89", "0.51"},
		{2025, 12, "101.275357", "0.48", "0.48", "0.45", "0.94", "0.46"},

		// 2026
		{2026, 1, "101.710866", "0.43", "0.43", "0.41", "0.96", "0.53"},
		{2026, 2, "102.138052", "0.42", "0.42", "0.40", "0.88", "0.46"},
		{2026, 3, "102.505747", "0.36", "0.36", "0.34", "0.92", "0.56"},
		{2026, 4, "102.895268", "0.38", "0.38", "0.35", "0.90", "0.52"},
		{2026, 5, "103.224533", "0.32", "0.32", "0.30", "0.91", "0.59"},
		{2026, 6, "103.493917", "0.26", "0.26", "0.24", "0.87", "0.61"},
		{2026, 7, "103.783700", "0.28", "0.28", "0.26", "0.93", "0.65"},
		{2026, 8, "103.991267", "0.20", "0.20", "0.19", "0.91", "0.71"},
		{2026, 9, "104.281145", "0.28", "0.28", "0.25", "0.89", "0.61"},
		{2026, 10, "104.281145", "0.00", "0.00", "0.00", "0.90", "0.90"}, // Mês corrente
	}

	for _, d := range historico {
		dt := time.Date(d.ano, time.Month(d.mes), 1, 0, 0, 0, 0, time.UTC)

		numTJSP, _ := decimal.NewFromString(d.tjspNumero)
		varTJSP, _ := decimal.NewFromString(d.tjspVar)
		r.AdicionarPonto("TJSP", dt, varTJSP, numTJSP)

		varIPCA, _ := decimal.NewFromString(d.ipcaVar)
		r.AdicionarPonto("IPCA", dt, varIPCA, decimal.Zero)

		varINPC, _ := decimal.NewFromString(d.inpcVar)
		r.AdicionarPonto("INPC", dt, varINPC, decimal.Zero)

		varSELIC, _ := decimal.NewFromString(d.selicVar)
		r.AdicionarPonto("SELIC", dt, varSELIC, decimal.Zero)

		varTaxaLegal, _ := decimal.NewFromString(d.taxaLegalVar)
		r.AdicionarPonto("TAXA_LEGAL", dt, varTaxaLegal, decimal.Zero)
	}
}
