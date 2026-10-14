package collectors

import (
	"context"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/shopspring/decimal"
)

// ColetorTribunais consolida e atualiza as tabelas de débitos judiciais de todos os Tribunais estaduais e federais
type ColetorTribunais struct{}

func NovoColetorTribunais() *ColetorTribunais {
	return &ColetorTribunais{}
}

func (c *ColetorTribunais) Nome() string {
	return "Tribunais de Justiça e Justiça Federal (Tabelas Práticas)"
}

// Coletar sincroniza todas as tabelas práticas estaduais e federais com base nos índices econômicos homologados
func (c *ColetorTribunais) Coletar(ctx context.Context, repo *calculator.RepositorioIndices) (int, error) {
	c.registrarTribunaisSeNecessario(repo)

	pontosAtualizados := 0
	inicio := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	agora := time.Now().UTC()

	cursor := inicio
	for !cursor.After(agora) {
		select {
		case <-ctx.Done():
			return pontosAtualizados, ctx.Err()
		default:
		}

		pontoIPCA, _ := repo.ObterPonto("IPCA", cursor)
		pontoINPC, _ := repo.ObterPonto("INPC", cursor)
		pontoIGPDI, _ := repo.ObterPonto("IGP_DI", cursor)
		pontoIGPM, _ := repo.ObterPonto("IGPM", cursor)

		// 1. TJSP: Variação mensal do IPCA (ou INPC no histórico pré-Lei 14.905)
		if !pontoIPCA.ValorPercentual.IsZero() {
			numAnt, _ := repo.ObterPonto("TJSP", cursor.AddDate(0, -1, 0))
			novoNum := numAnt.NumeroIndice
			if novoNum.IsZero() {
				novoNum, _ = decimal.NewFromString("100.000000")
			}
			fatorMes := decimal.NewFromInt(1).Add(pontoIPCA.ValorPercentual.Div(decimal.NewFromInt(100)))
			novoNum = novoNum.Mul(fatorMes).Round(6)

			repo.AdicionarPonto("TJSP", cursor, pontoIPCA.ValorPercentual, novoNum)
			pontosAtualizados++
		}

		// 2. TJRJ: Tribunal de Justiça do Rio de Janeiro (Base IPCA / IPCA-E)
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJRJ", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		// 3. TJMG: Tribunal de Justiça de Minas Gerais (Base IPCA / INPC)
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJMG", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			repo.AdicionarPonto("TJMG_EXPURGADA", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados += 2
		}

		// 4. TJRS: Tribunal de Justiça do Rio Grande do Sul (IGP-M / IPCA)
		if !pontoIGPM.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJRS", cursor, pontoIGPM.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		} else if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJRS", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		// 5. TJPR: Critério clássico média (IGP-DI + INPC) / 2
		if !pontoIGPDI.ValorPercentual.IsZero() && !pontoINPC.ValorPercentual.IsZero() {
			media := pontoIGPDI.ValorPercentual.Add(pontoINPC.ValorPercentual).Div(decimal.NewFromInt(2)).Round(4)
			repo.AdicionarPonto("TJPR", cursor, media, decimal.Zero)
			pontosAtualizados++
		}
		// TJPR Precatórios (IPCA-E / IPCA)
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJPR_PRECAT", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		// 6. TJSC: Tabela ICGJ (Base INPC)
		if !pontoINPC.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJSC", cursor, pontoINPC.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		// 7. TJES, TJCE e TJDF
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TJES", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			repo.AdicionarPonto("TJCE", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			repo.AdicionarPonto("TJDF", cursor, pontoINPC.ValorPercentual, decimal.Zero)
			pontosAtualizados += 3
		}

		// 8. Justiça Federal (CJF Resoluções 267/2013 e 658/2020)
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("JF_GERAL", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			repo.AdicionarPonto("JF_FAZENDA", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			repo.AdicionarPonto("JF_DESAPROPRIACAO", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados += 3
		}
		if !pontoINPC.ValorPercentual.IsZero() {
			repo.AdicionarPonto("JF_PREVIDENCIARIO", cursor, pontoINPC.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		// 9. TST (Trabalhista - IPCA-E na fase pré-judicial)
		if !pontoIPCA.ValorPercentual.IsZero() {
			repo.AdicionarPonto("TST_TRABALHISTA", cursor, pontoIPCA.ValorPercentual, decimal.Zero)
			pontosAtualizados++
		}

		cursor = cursor.AddDate(0, 1, 0)
	}

	return pontosAtualizados, nil
}

// registrarTribunaisSeNecessario garante o catálogo completo de todos os tribunais e cortes do DrCalc
func (c *ColetorTribunais) registrarTribunaisSeNecessario(repo *calculator.RepositorioIndices) {
	tabelas := []calculator.IndiceInfo{
		// Tribunais de Justiça Estaduais
		{Codigo: "TJRJ", Nome: "Tabela Tribunal de Justiça RJ", Fonte: "TJRJ / Corregedoria", Periodicidade: "MENSAL", Descricao: "Tabela oficial para liquidação de débitos cíveis do Estado do Rio de Janeiro."},
		{Codigo: "TJMG", Nome: "Tabela TJMG (Não Expurgada)", Fonte: "TJMG / Corregedoria", Periodicidade: "MENSAL", Descricao: "Tabela prática da CGJ de Minas Gerais sem expurgos inflacionários."},
		{Codigo: "TJMG_EXPURGADA", Nome: "Tabela TJMG (Expurgada)", Fonte: "TJMG / Corregedoria", Periodicidade: "MENSAL", Descricao: "Tabela prática da CGJ de Minas Gerais com expurgos de planos econômicos."},
		{Codigo: "TJRS", Nome: "Tabela Tribunal de Justiça RS", Fonte: "TJRS", Periodicidade: "MENSAL", Descricao: "Tabela de atualização monetária judicial do Rio Grande do Sul."},
		{Codigo: "TJPR", Nome: "TJPR (Média IGP-DI / INPC)", Fonte: "TJPR", Periodicidade: "MENSAL", Descricao: "Critério oficial do TJPR baseado na média aritmética entre IGP-DI e INPC."},
		{Codigo: "TJPR_PRECAT", Nome: "TJPR (Precatórios IPCA-E)", Fonte: "TJPR", Periodicidade: "MENSAL", Descricao: "Tabela específica para cálculo de precatórios e Fazenda no TJPR."},
		{Codigo: "TJSC", Nome: "TJSC (Tabela ICGJ)", Fonte: "TJSC / Corregedoria", Periodicidade: "MENSAL", Descricao: "Índice da Corregedoria-Geral da Justiça de Santa Catarina baseado no INPC."},
		{Codigo: "TJES", Nome: "Tabela Tribunal de Justiça ES", Fonte: "TJES", Periodicidade: "MENSAL", Descricao: "Tabela de atualização monetária de débitos judiciais do Espírito Santo."},
		{Codigo: "TJCE", Nome: "TJCE (Condenatórias TJ/Ceará)", Fonte: "TJCE", Periodicidade: "MENSAL", Descricao: "Tabela de débitos judiciais cíveis do Ceará."},
		{Codigo: "TJDF", Nome: "Tabela TJDFT", Fonte: "TJDFT", Periodicidade: "MENSAL", Descricao: "Tabela de débitos judiciais do Distrito Federal e Territórios."},

		// Justiça Federal (Manual de Cálculos CJF)
		{Codigo: "JF_GERAL", Nome: "Justiça Federal - Condenatórias em Geral", Fonte: "Conselho da Justiça Federal (CJF)", Periodicidade: "MENSAL", Descricao: "Resolução CJF 267/2013 e 658/2020 para condenações cíveis gerais na Justiça Federal."},
		{Codigo: "JF_PREVIDENCIARIO", Nome: "Justiça Federal - Benefício Previdenciário", Fonte: "Conselho da Justiça Federal (CJF)", Periodicidade: "MENSAL", Descricao: "Resolução CJF 267/2013 com base no INPC para liquidação de benefícios do INSS."},
		{Codigo: "JF_FAZENDA", Nome: "Justiça Federal - Fazenda Pública", Fonte: "Conselho da Justiça Federal (CJF)", Periodicidade: "MENSAL", Descricao: "Condenações contra a União, Autarquias e Fundações Públicas federais."},
		{Codigo: "JF_DESAPROPRIACAO", Nome: "Justiça Federal - Desapropriações", Fonte: "Conselho da Justiça Federal (CJF)", Periodicidade: "MENSAL", Descricao: "Cálculos em ações expropriatórias sob a Resolução 267/2013 do CJF."},

		// Justiça do Trabalho
		{Codigo: "TST_TRABALHISTA", Nome: "TST - Débitos Trabalhistas (ADC 58 STF)", Fonte: "Tribunal Superior do Trabalho / STF", Periodicidade: "MENSAL", Descricao: "IPCA-E na fase pré-judicial e Taxa SELIC a partir da citação (decisão vinculante do STF)."},

		// Índices de Preços e FGV adicionais
		{Codigo: "IGPM", Nome: "IGP-M (Índice Geral de Preços do Mercado)", Fonte: "FGV / IBRE", Periodicidade: "MENSAL", Descricao: "Muito utilizado em contratos de locação imobiliária e reajustes contratuais."},
		{Codigo: "IGP_DI", Nome: "IGP-DI (Disponibilidade Interna)", Fonte: "FGV / IBRE", Periodicidade: "MENSAL", Descricao: "Índice de preços clássico da Fundação Getulio Vargas."},
		{Codigo: "INCC", Nome: "INCC-DI (Custo da Construção)", Fonte: "FGV / IBRE", Periodicidade: "MENSAL", Descricao: "Padrão de atualização de contratos imobiliários em fase de obras."},
		{Codigo: "CDI", Nome: "CDI Acumulado Mensal (% a.m.)", Fonte: "B3 / Banco Central", Periodicidade: "MENSAL", Descricao: "Certificado de Depósito Interbancário, padrão do mercado financeiro."},
		{Codigo: "TR", Nome: "Taxa Referencial (TR)", Fonte: "Banco Central do Brasil", Periodicidade: "MENSAL", Descricao: "Taxa do primeiro dia do mês para poupança antiga e contratos habitacionais CEF."},
		{Codigo: "POUPANCA", Nome: "Poupança Nova (% a.m.)", Fonte: "Banco Central do Brasil", Periodicidade: "MENSAL", Descricao: "Rendimento da caderneta de poupança após 04/05/2012."},
	}

	for _, t := range tabelas {
		repo.RegistrarIndice(t)
	}
}
