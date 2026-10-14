package calculator

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ParseDataFlexivel aceita datas nos formatos AAAA-MM-DD ou DD/MM/AAAA
func ParseDataFlexivel(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	// Se for DD/MM/AAAA
	if strings.Contains(s, "/") {
		partes := strings.Split(s, "/")
		if len(partes) == 3 {
			return time.Parse("02/01/2006", s)
		}
	}
	// Se for AAAA-MM-DD
	return time.Parse("2006-01-02", s)
}

// CalcularFatorCorrecao obtém o multiplicador de correção monetária entre a data inicial e a data final
func CalcularFatorCorrecao(repo *RepositorioIndices, codigoIndice string, dataInicio, dataFim time.Time, permitirDeflacao bool) (decimal.Decimal, error) {
	if codigoIndice == "" || codigoIndice == "SEM_CORRECAO" {
		return decimal.NewFromInt(1), nil
	}

	return repo.ObterFatorAcumulado(codigoIndice, dataInicio, dataFim, permitirDeflacao)
}
