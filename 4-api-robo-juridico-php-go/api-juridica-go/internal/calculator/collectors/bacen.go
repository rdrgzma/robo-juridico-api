package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/shopspring/decimal"
)

// PontoBacenSGS representa um registro retornado pela API do Banco Central
type PontoBacenSGS struct {
	Data  string `json:"data"`  // DD/MM/AAAA
	Valor string `json:"valor"` // Taxa em string (ex: "0.82" ou "-0.32")
}

// ColetorBacenSGS consome as séries econômicas oficiais do Banco Central do Brasil
type ColetorBacenSGS struct {
	httpClient *http.Client
	seriesMap  map[string]int // Mapeamento: Código Nosso -> Código SGS Bacen
}

func NovoColetorBacenSGS(timeout time.Duration) *ColetorBacenSGS {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &ColetorBacenSGS{
		httpClient: &http.Client{Timeout: timeout},
		seriesMap: map[string]int{
			"IPCA":     433,  // IPCA variação mensal
			"INPC":     188,  // INPC variação mensal
			"IGPM":     189,  // IGP-M variação mensal
			"IGP_DI":   190,  // IGP-DI variação mensal
			"INCC":     192,  // INCC-DI variação mensal
			"SELIC":    4390, // Taxa Selic mensal acumulada
			"CDI":      4389, // CDI mensal
			"TR":       226,  // TR mensal
			"POUPANCA": 195,  // Poupança Nova
		},
	}
}

func (c *ColetorBacenSGS) Nome() string {
	return "Banco Central do Brasil (SGS)"
}

// Coletar busca os dados mais recentes de todas as séries cadastradas no Bacen
func (c *ColetorBacenSGS) Coletar(ctx context.Context, repo *calculator.RepositorioIndices) (int, error) {
	totalPontos := 0
	var ultimoErro error

	for codigoIndice, codigoSGS := range c.seriesMap {
		select {
		case <-ctx.Done():
			return totalPontos, ctx.Err()
		default:
		}

		pontos, err := c.buscarSerieSGS(ctx, codigoSGS, 12) // Últimos 12 meses
		if err != nil {
			ultimoErro = fmt.Errorf("erro na serie %s (SGS %d): %w", codigoIndice, codigoSGS, err)
			continue
		}

		for _, p := range pontos {
			dt, err := time.Parse("02/01/2006", p.Data)
			if err != nil {
				continue
			}
			// Padroniza sempre para o dia 1 do mês em UTC
			dtMes := time.Date(dt.Year(), dt.Month(), 1, 0, 0, 0, 0, time.UTC)

			val, err := decimal.NewFromString(p.Valor)
			if err != nil {
				continue
			}

			repo.AdicionarPonto(codigoIndice, dtMes, val, decimal.Zero)
			totalPontos++
		}
	}

	// Calcula automaticamente a TAXA_LEGAL (Lei 14.905/2024 = max(0, Selic - IPCA))
	totalTaxaLegal := c.calcularTaxaLegal(repo)
	totalPontos += totalTaxaLegal

	return totalPontos, ultimoErro
}

// buscarSerieSGS faz a requisição HTTP à API REST do Bacen
func (c *ColetorBacenSGS) buscarSerieSGS(ctx context.Context, codigoSGS, ultimos int) ([]PontoBacenSGS, error) {
	url := fmt.Sprintf("https://api.bcb.gov.br/dados/serie/bcdata.sgs.%d/dados/ultimos/%d?formato=json", codigoSGS, ultimos)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Robo-Juridico-API/1.0 (Calculadora Judicial)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status HTTP inesperado do Bacen: %d", resp.StatusCode)
	}

	var pontos []PontoBacenSGS
	if err := json.NewDecoder(resp.Body).Decode(&pontos); err != nil {
		return nil, err
	}

	return pontos, nil
}

// calcularTaxaLegal apura os juros da Lei 14.905/2024 para cada mês após ago/2024
func (c *ColetorBacenSGS) calcularTaxaLegal(repo *calculator.RepositorioIndices) int {
	pontosAtualizados := 0
	inicioLei := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	agora := time.Now().UTC()
	cursor := inicioLei

	for !cursor.After(agora) {
		pontoSelic, errSelic := repo.ObterPonto("SELIC", cursor)
		pontoIPCA, errIPCA := repo.ObterPonto("IPCA", cursor)

		if errSelic == nil && errIPCA == nil {
			// Taxa Legal = Selic - IPCA
			diferenca := pontoSelic.ValorPercentual.Sub(pontoIPCA.ValorPercentual)
			taxaLegal := diferenca
			if taxaLegal.IsNegative() {
				taxaLegal = decimal.Zero // Art. 406 CC: se negativa, taxa de juros é zero
			}

			repo.AdicionarPonto("TAXA_LEGAL", cursor, taxaLegal, decimal.Zero)
			pontosAtualizados++
		}
		cursor = cursor.AddDate(0, 1, 0)
	}

	return pontosAtualizados
}
