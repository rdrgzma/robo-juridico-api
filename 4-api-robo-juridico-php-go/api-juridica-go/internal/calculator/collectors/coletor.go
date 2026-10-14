package collectors

import (
	"context"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/calculator"
)

// Coletor define o contrato para ingestão de séries históricas de indicadores econômicos e judiciais
type Coletor interface {
	Nome() string
	Coletar(ctx context.Context, repo *calculator.RepositorioIndices) (int, error)
}

// RelatorioColeta consolida o resultado da execução dos coletores
type RelatorioColeta struct {
	TotalPontosAtualizados int               `json:"total_pontos_atualizados"`
	Erros                  map[string]string `json:"erros,omitempty"`
	FontesProcessadas      []string          `json:"fontes_processadas"`
}
