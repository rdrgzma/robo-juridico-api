package datajud

import (
	"testing"

	"github.com/rdrgzma/robo-juridico-desktop/internal/domain"
)

func TestClassificarTipo(t *testing.T) {
	testes := []struct {
		desc     string
		esperado domain.TipoMovimentacao
	}{
		{"Prolatada sentença com mérito", domain.TipoSentenca},
		{"Decisão liminar concedida", domain.TipoDecisao},
		{"Mero despacho ordinatório", domain.TipoDespacho},
		{"Designada audiência de conciliação", domain.TipoAudiencia},
		{"Expedida intimação ao réu", domain.TipoIntimacao},
		{"Disponibilizado no diário da justiça", domain.TipoPublicacaoDJE},
		{"Juntada de petição", domain.TipoOutro},
	}

	for _, tt := range testes {
		obtido := ClassificarTipo(tt.desc)
		if obtido != tt.esperado {
			t.Errorf("ClassificarTipo(%q) = %v; esperado %v", tt.desc, obtido, tt.esperado)
		}
	}
}

func TestFormatarDinheiroBR(t *testing.T) {
	v1 := 1250.50
	if obtido := FormatarDinheiroBR(&v1); obtido != "R$ 1.250,50" {
		t.Errorf("Esperado 'R$ 1.250,50', obtido '%s'", obtido)
	}

	v2 := 1000000.00
	if obtido := FormatarDinheiroBR(&v2); obtido != "R$ 1.000.000,00" {
		t.Errorf("Esperado 'R$ 1.000.000,00', obtido '%s'", obtido)
	}

	if obtido := FormatarDinheiroBR(nil); obtido != "" {
		t.Errorf("Esperado '', obtido '%s'", obtido)
	}
}
