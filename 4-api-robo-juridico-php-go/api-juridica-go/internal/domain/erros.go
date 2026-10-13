package domain

import (
	"fmt"
	"net/http"
)

type ErroAPI struct {
	Status   int           `json:"-"`
	Codigo   string        `json:"codigo"`
	Mensagem string        `json:"mensagem"`
	TentarEm int           `json:"tentar_em,omitempty"`
	Detalhes []ErroDetalhe `json:"detalhes,omitempty"`
}

func (e *ErroAPI) Error() string {
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Codigo, e.Mensagem, e.Status)
}

func NovoErroAPI(status int, codigo, mensagem string) *ErroAPI {
	return &ErroAPI{
		Status:   status,
		Codigo:   codigo,
		Mensagem: mensagem,
	}
}

func NovoErroComRetry(status int, codigo, mensagem string, tentarEm int) *ErroAPI {
	return &ErroAPI{
		Status:   status,
		Codigo:   codigo,
		Mensagem: mensagem,
		TentarEm: tentarEm,
	}
}

var (
	ErrNaoAutorizado = &ErroAPI{
		Status:   http.StatusUnauthorized,
		Codigo:   "nao_autorizado",
		Mensagem: "Envie uma chave válida no cabeçalho X-API-Key.",
	}

	ErrCotaEsgotada = &ErroAPI{
		Status:   http.StatusPaymentRequired,
		Codigo:   "cota_esgotada",
		Mensagem: "A cota de consultas do mês acabou. Ela renova no dia 1º; para ampliar, fale com o suporte.",
	}

	ErrContaSuspensa = &ErroAPI{
		Status:   http.StatusForbidden,
		Codigo:   "conta_suspensa",
		Mensagem: "Conta suspensa. Fale com o suporte.",
	}

	ErrProcessoNaoEncontrado = &ErroAPI{
		Status:   http.StatusNotFound,
		Codigo:   "processo_nao_encontrado",
		Mensagem: "O processo não aparece no DataJud nem no DJEN. Se foi distribuído agora, tente mais tarde.",
	}

	ErrNumeroInvalido = &ErroAPI{
		Status:   http.StatusUnprocessableEntity,
		Codigo:   "numero_invalido",
		Mensagem: "Número fora do padrão CNJ (NNNNNNN-DD.AAAA.J.TR.OOOO) ou com dígito verificador errado.",
	}

	ErrUFInvalida = &ErroAPI{
		Status:   http.StatusUnprocessableEntity,
		Codigo:   "uf_invalida",
		Mensagem: "UF desconhecida.",
	}

	ErrFonteIndisponivel = &ErroAPI{
		Status:   http.StatusServiceUnavailable,
		Codigo:   "fonte_indisponivel",
		Mensagem: "A fonte do CNJ não respondeu a tempo. Tente de novo em instantes.",
		TentarEm: 30,
	}

	ErrErroInterno = &ErroAPI{
		Status:   http.StatusInternalServerError,
		Codigo:   "erro_interno",
		Mensagem: "Erro interno do servidor. Informe o id_requisicao ao suporte.",
	}
)
