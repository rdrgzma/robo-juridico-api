package calculator

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrIndiceNaoEncontrado = errors.New("indice economico nao encontrado")
	ErrDataForaDaSerie     = errors.New("data solicitada esta fora do periodo coberto pela serie historica")
)

// RepositorioIndices gerencia as séries históricas de índices econômicos em memória com thread-safety
type RepositorioIndices struct {
	mu      sync.RWMutex
	indices map[string]IndiceInfo
	// series: map[codigo_indice]map[ano_mes_string]PontoSerie, ex: "2024-01"
	series  map[string]map[string]PontoSerie
}

func NovoRepositorioIndices() *RepositorioIndices {
	repo := &RepositorioIndices{
		indices: make(map[string]IndiceInfo),
		series:  make(map[string]map[string]PontoSerie),
	}
	repo.carregarSeedPadrao()
	return repo
}

// RegistrarIndice cadastra os metadados de um novo índice
func (r *RepositorioIndices) RegistrarIndice(info IndiceInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.indices[info.Codigo] = info
	if _, ok := r.series[info.Codigo]; !ok {
		r.series[info.Codigo] = make(map[string]PontoSerie)
	}
}

// AdicionarPonto insere ou atualiza um ponto na série histórica
func (r *RepositorioIndices) AdicionarPonto(codigoIndice string, dataRef time.Time, valorPct, numeroIndice decimal.Decimal) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.series[codigoIndice]; !ok {
		r.series[codigoIndice] = make(map[string]PontoSerie)
	}

	chaveAnoMes := dataRef.Format("2006-01")
	fator := decimal.NewFromInt(1).Add(valorPct.Div(decimal.NewFromInt(100)))

	ponto := PontoSerie{
		DataReferencia:      dataRef,
		DataReferenciaTexto: chaveAnoMes,
		ValorPercentual:     valorPct,
		FatorMultiplicador:  fator,
		NumeroIndice:        numeroIndice,
	}

	r.series[codigoIndice][chaveAnoMes] = ponto
}

// ObterIndice retorna as informações cadastrais do índice
func (r *RepositorioIndices) ObterIndice(codigo string) (IndiceInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.indices[codigo]
	if !ok {
		return IndiceInfo{}, fmt.Errorf("%w: %s", ErrIndiceNaoEncontrado, codigo)
	}
	return info, nil
}

// ListarIndices retorna a lista de todos os índices disponíveis
func (r *RepositorioIndices) ListarIndices() []IndiceInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lista := make([]IndiceInfo, 0, len(r.indices))
	for _, info := range r.indices {
		lista = append(lista, info)
	}
	return lista
}

// ObterPonto busca o registro específico de um mês
func (r *RepositorioIndices) ObterPonto(codigo string, data time.Time) (PontoSerie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serie, ok := r.series[codigo]
	if !ok {
		return PontoSerie{}, fmt.Errorf("%w: %s", ErrIndiceNaoEncontrado, codigo)
	}

	chaveAnoMes := data.Format("2006-01")
	ponto, ok := serie[chaveAnoMes]
	if !ok {
		return PontoSerie{}, fmt.Errorf("%w: %s para o mes %s", ErrDataForaDaSerie, codigo, chaveAnoMes)
	}

	return ponto, nil
}

// ObterFatorAcumulado calcula o fator de correção entre duas datas
// Fator = NumeroIndice(DataFim) / NumeroIndice(DataInicio)
// Caso deflação seja desativada, calcula via produtório das variações mensais não-negativas
func (r *RepositorioIndices) ObterFatorAcumulado(codigo string, dataInicio, dataFim time.Time, permitirDeflacao bool) (decimal.Decimal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serie, ok := r.series[codigo]
	if !ok {
		return decimal.NewFromInt(1), fmt.Errorf("%w: %s", ErrIndiceNaoEncontrado, codigo)
	}

	// Se dataInicio >= dataFim, fator é 1.0 (sem correção)
	if !dataInicio.Before(dataFim) {
		return decimal.NewFromInt(1), nil
	}

	pontoInicio, okInicio := serie[dataInicio.Format("2006-01")]
	pontoFim, okFim := serie[dataFim.Format("2006-01")]

	if !okInicio || !okFim {
		return decimal.NewFromInt(1), fmt.Errorf("%w: periodo %s ate %s para indice %s",
			ErrDataForaDaSerie, dataInicio.Format("2006-01"), dataFim.Format("2006-01"), codigo)
	}

	// 1. Caso com deflação permitida e número-índice presente: quociente direto
	if permitirDeflacao && !pontoInicio.NumeroIndice.IsZero() && !pontoFim.NumeroIndice.IsZero() {
		return pontoFim.NumeroIndice.DivRound(pontoInicio.NumeroIndice, 10), nil
	}

	// 2. Caso sem deflação (ou sem número-índice): produtório mês a mês
	fatorTotal := decimal.NewFromInt(1)
	cursor := time.Date(dataInicio.Year(), dataInicio.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(dataFim.Year(), dataFim.Month(), 1, 0, 0, 0, 0, time.UTC)

	for !cursor.After(fim) {
		chave := cursor.Format("2006-01")
		ponto, ok := serie[chave]
		if ok {
			var variacao decimal.Decimal
			if !permitirDeflacao && ponto.ValorPercentual.IsNegative() {
				variacao = decimal.Zero // Trava deflação em 0%
			} else {
				variacao = ponto.ValorPercentual
			}

			fatorMes := decimal.NewFromInt(1).Add(variacao.Div(decimal.NewFromInt(100)))
			fatorTotal = fatorTotal.Mul(fatorMes)
		}
		cursor = cursor.AddDate(0, 1, 0)
	}

	return fatorTotal.Round(10), nil
}

// ObterSerieCompleta retorna todos os pontos ordenados para exibição
func (r *RepositorioIndices) ObterSerieCompleta(codigo string) ([]PontoSerie, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serie, ok := r.series[codigo]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrIndiceNaoEncontrado, codigo)
	}

	lista := make([]PontoSerie, 0, len(serie))
	for _, p := range serie {
		lista = append(lista, p)
	}
	return lista, nil
}
