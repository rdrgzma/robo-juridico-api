package comunica

import (
	"testing"

	"github.com/rdrgzma/robo-juridico-desktop/internal/domain"
)

func TestSemHTML(t *testing.T) {
	entrada := `<p>Fica intimado o <strong>autor</strong> para manifestação &amp; prosseguimento.</p><style>body { color: red; }</style>`
	esperado := "Fica intimado o autor para manifestação & prosseguimento."
	obtido := SemHTML(entrada)

	if obtido != esperado {
		t.Errorf("SemHTML falhou. Esperado '%s', obtido '%s'", esperado, obtido)
	}
}

func TestExtrairTeorComDivBody(t *testing.T) {
	entrada := `<html><body><div id="cabecalho">Topo</div><div id="divBody">Conteúdo relevante do ato judicial</div></body></html>`
	esperado := "Conteúdo relevante do ato judicial"
	obtido := ExtrairTeor(entrada)

	if obtido != esperado {
		t.Errorf("ExtrairTeor falhou. Esperado '%s', obtido '%s'", esperado, obtido)
	}
}

func TestClassificarTipo(t *testing.T) {
	testes := []struct {
		comunicacao string
		esperado    domain.TipoMovimentacao
	}{
		{"Intimação Eletrônica", domain.TipoIntimacao},
		{"Despacho do Juiz", domain.TipoDespacho},
		{"Decisão Interlocutória", domain.TipoDecisao},
		{"Sentença com Resolução de Mérito", domain.TipoSentenca},
		{"Publicação em Edital", domain.TipoPublicacaoDJE},
		{"Outro ato ordinatório", domain.TipoOutro},
	}

	for _, tt := range testes {
		obtido := ClassificarTipo(tt.comunicacao)
		if obtido != tt.esperado {
			t.Errorf("ClassificarTipo(%q) = %v; esperado %v", tt.comunicacao, obtido, tt.esperado)
		}
	}
}
