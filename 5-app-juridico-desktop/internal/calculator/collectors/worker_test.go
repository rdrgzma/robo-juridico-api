package collectors

import (
	"context"
	"testing"
	"time"

	"github.com/rdrgzma/robo-juridico-desktop/internal/calculator"
	"github.com/shopspring/decimal"
)

func TestColetorTribunais_RegistraETestSincronizacao(t *testing.T) {
	repo := calculator.NovoRepositorioIndices()
	coletor := NovoColetorTribunais()

	// Insere pontos base de IPCA, INPC e IGP-DI para testar o cálculo dos tribunais
	dt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	repo.AdicionarPonto("IPCA", dt, decimal.NewFromFloat(0.82), decimal.Zero)
	repo.AdicionarPonto("INPC", dt, decimal.NewFromFloat(0.82), decimal.Zero)
	repo.AdicionarPonto("IGP_DI", dt, decimal.NewFromFloat(1.50), decimal.Zero)
	repo.AdicionarPonto("IGPM", dt, decimal.NewFromFloat(1.57), decimal.Zero)

	pontos, err := coletor.Coletar(context.Background(), repo)
	if err != nil {
		t.Fatalf("Erro ao sincronizar tabelas dos tribunais: %v", err)
	}

	if pontos == 0 {
		t.Errorf("Esperado atualizar pontos dos tribunais, obtido 0")
	}

	// Verifica se os tribunais solicitados pelo usuário estão todos presentes no catálogo
	tribunaisEsperados := []string{
		"TJSP", "TJRJ", "TJMG", "TJMG_EXPURGADA", "TJRS", "TJPR", "TJPR_PRECAT",
		"TJSC", "TJES", "TJCE", "TJDF", "JF_GERAL", "JF_PREVIDENCIARIO", "JF_FAZENDA",
		"JF_DESAPROPRIACAO", "TST_TRABALHISTA", "IGPM", "IGP_DI", "INCC", "CDI", "TR", "POUPANCA",
	}

	indicesCadastrados := repo.ListarIndices()
	mapaIndices := make(map[string]bool)
	for _, idx := range indicesCadastrados {
		mapaIndices[idx.Codigo] = true
	}

	for _, cod := range tribunaisEsperados {
		if !mapaIndices[cod] {
			t.Errorf("Tribunal/Indexador '%s' esperado não foi registrado no repositório!", cod)
		}
	}

	// Verifica o critério específico do TJPR: média entre IGP-DI e INPC: (1.50 + 0.82) / 2 = 1.16
	pontoTJPR, err := repo.ObterPonto("TJPR", dt)
	if err != nil {
		t.Fatalf("Erro ao obter ponto do TJPR: %v", err)
	}
	fMedia, _ := pontoTJPR.ValorPercentual.Float64()
	if fMedia < 1.15 || fMedia > 1.17 {
		t.Errorf("Média do TJPR esperada ~1.16, obtido %f", fMedia)
	}

	t.Logf("Total de índices cadastrados no catálogo: %d", len(indicesCadastrados))
	t.Logf("Ponto TJPR calculado: %s%%", pontoTJPR.ValorPercentual.String())
}
