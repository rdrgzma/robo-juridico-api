package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
	"github.com/shopspring/decimal"
)

type respostaIBGESIDRA []struct {
	Resultados []struct {
		Series []struct {
			Serie map[string]string `json:"serie"` // Chaves: "AAAAMM" (ex: "202609"), Valor: "0.82"
		} `json:"series"`
	} `json:"resultados"`
}

// ColetorIBGESIDRA consome diretamente a API oficial do IBGE para índices de preços ao consumidor
type ColetorIBGESIDRA struct {
	httpClient *http.Client
}

func NovoColetorIBGESIDRA(timeout time.Duration) *ColetorIBGESIDRA {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &ColetorIBGESIDRA{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *ColetorIBGESIDRA) Nome() string {
	return "IBGE (SIDRA)"
}

// Coletar busca IPCA, INPC e IPCA-15 dos últimos 12 períodos divulgados
func (c *ColetorIBGESIDRA) Coletar(ctx context.Context, repo *calculator.RepositorioIndices) (int, error) {
	configuracoes := []struct {
		codigo   string
		agregado int
		variavel int
	}{
		{"IPCA", 1737, 63},
		{"INPC", 1736, 44},
		{"IPCA_15", 3065, 355},
	}

	totalPontos := 0
	var ultimoErro error

	for _, cfg := range configuracoes {
		select {
		case <-ctx.Done():
			return totalPontos, ctx.Err()
		default:
		}

		pontos, err := c.buscarSerieSIDRA(ctx, cfg.agregado, cfg.variavel, 12)
		if err != nil {
			ultimoErro = fmt.Errorf("erro no IBGE para %s: %w", cfg.codigo, err)
			continue
		}

		for anoMesStr, valStr := range pontos {
			if len(anoMesStr) != 6 {
				continue
			}
			ano, errAno := strconv.Atoi(anoMesStr[:4])
			mes, errMes := strconv.Atoi(anoMesStr[4:])
			if errAno != nil || errMes != nil || mes < 1 || mes > 12 {
				continue
			}

			val, errVal := decimal.NewFromString(valStr)
			if errVal != nil {
				continue
			}

			dt := time.Date(ano, time.Month(mes), 1, 0, 0, 0, 0, time.UTC)
			repo.AdicionarPonto(cfg.codigo, dt, val, decimal.Zero)
			totalPontos++
		}
	}

	return totalPontos, ultimoErro
}

func (c *ColetorIBGESIDRA) buscarSerieSIDRA(ctx context.Context, agregado, variavel, periodos int) (map[string]string, error) {
	url := fmt.Sprintf("https://servicodados.ibge.gov.br/api/v3/agregados/%d/periodos/-%d/variaveis/%d?localidades=N1[all]",
		agregado, periodos, variavel)

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
		return nil, fmt.Errorf("status HTTP inesperado do IBGE SIDRA: %d", resp.StatusCode)
	}

	var data respostaIBGESIDRA
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if len(data) == 0 || len(data[0].Resultados) == 0 || len(data[0].Resultados[0].Series) == 0 {
		return nil, fmt.Errorf("dados nao encontrados na resposta do IBGE SIDRA")
	}

	return data[0].Resultados[0].Series[0].Serie, nil
}
