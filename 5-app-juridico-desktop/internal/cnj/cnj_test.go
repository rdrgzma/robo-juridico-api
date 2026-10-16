package cnj_test

import (
	"testing"

	"github.com/rdrgzma/robo-juridico-desktop/internal/cnj"
)

func TestCNJ_Valido(t *testing.T) {
	testes := []struct {
		nome         string
		entrada      string
		esperaValido bool
		tribunal     string
		ano          int
	}{
		{
			nome:         "TJMG com pontuação",
			entrada:      "5010754-07.2024.8.13.0625",
			esperaValido: true,
			tribunal:     "TJMG",
			ano:          2024,
		},
		{
			nome:         "TJMG sem pontuação (20 dígitos)",
			entrada:      "50107540720248130625",
			esperaValido: true,
			tribunal:     "TJMG",
			ano:          2024,
		},
		{
			nome:         "TRT3 trabalhista válido",
			entrada:      "0010734-59.2026.5.03.0076",
			esperaValido: true,
			tribunal:     "TRT3",
			ano:          2026,
		},
		{
			nome:         "TJMG processo 2025 válido",
			entrada:      "1000254-02.2025.8.13.0079",
			esperaValido: true,
			tribunal:     "TJMG",
			ano:          2025,
		},
		{
			nome:         "Dígito verificador errado (último dígito alterado)",
			entrada:      "5010754-08.2024.8.13.0625",
			esperaValido: false,
			tribunal:     "TJMG",
			ano:          2024,
		},
		{
			nome:         "Número incompleto",
			entrada:      "12345",
			esperaValido: false,
		},
		{
			nome:         "String vazia",
			entrada:      "",
			esperaValido: false,
		},
	}

	for _, tt := range testes {
		t.Run(tt.nome, func(t *testing.T) {
			p := cnj.Parse(tt.entrada)
			if p.Valido != tt.esperaValido {
				t.Errorf("esperava Valido=%v, obteve %v para %s", tt.esperaValido, p.Valido, tt.entrada)
			}
			if tt.esperaValido {
				if p.Tribunal != tt.tribunal {
					t.Errorf("esperava Tribunal=%s, obteve %s", tt.tribunal, p.Tribunal)
				}
				if p.Ano != tt.ano {
					t.Errorf("esperava Ano=%d, obteve %d", tt.ano, p.Ano)
				}
			}
		})
	}
}
