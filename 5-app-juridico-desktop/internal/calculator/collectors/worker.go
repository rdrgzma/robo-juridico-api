package collectors

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/rdrgzma/robo-juridico-desktop/internal/calculator"
)

// WorkerColetor gerencia a execução periódica e manual de todos os coletores de dados oficiais
type WorkerColetor struct {
	repo             *calculator.RepositorioIndices
	coletores        []Coletor
	intervalo        time.Duration
	mu               sync.RWMutex
	emExecucao       bool
	ultimaExecucao   time.Time
	ultimoRelatorio  RelatorioColeta
}

func NovoWorker(repo *calculator.RepositorioIndices, intervalo time.Duration) *WorkerColetor {
	if intervalo <= 0 {
		intervalo = 24 * time.Hour // Padrão: 1 vez ao dia
	}

	return &WorkerColetor{
		repo:      repo,
		intervalo: intervalo,
		coletores: []Coletor{
			NovoColetorBacenSGS(20 * time.Second),
			NovoColetorIBGESIDRA(20 * time.Second),
			NovoColetorTribunais(),
		},
	}
}

// Iniciar dispara o agendador em background em uma goroutine não-bloqueante
func (w *WorkerColetor) Iniciar(ctx context.Context) {
	go func() {
		log.Printf("[WORKER_INDICES] Worker de coleta diária iniciado (intervalo: %v). Executando coleta inicial...", w.intervalo)

		// Coleta inicial logo após subir a API (com timeout de 2 minutos)
		initCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		rel := w.Executar(initCtx)
		cancel()
		log.Printf("[WORKER_INDICES] Coleta inicial concluída. %d pontos atualizados em %d fontes.",
			rel.TotalPontosAtualizados, len(rel.FontesProcessadas))

		ticker := time.NewTicker(w.intervalo)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[WORKER_INDICES] Worker de coleta finalizado.")
				return
			case <-ticker.C:
				log.Println("[WORKER_INDICES] Disparando rotina agendada de atualização de índices...")
				runCtx, runCancel := context.WithTimeout(ctx, 3*time.Minute)
				r := w.Executar(runCtx)
				runCancel()
				log.Printf("[WORKER_INDICES] Ciclo concluído: %d pontos atualizados.", r.TotalPontosAtualizados)
			}
		}
	}()
}

// Executar roda todos os coletores em sequência de forma atômica
func (w *WorkerColetor) Executar(ctx context.Context) RelatorioColeta {
	w.mu.Lock()
	if w.emExecucao {
		w.mu.Unlock()
		return RelatorioColeta{
			Erros: map[string]string{"aviso": "Uma coleta ja esta em andamento no momento."},
		}
	}
	w.emExecucao = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.emExecucao = false
		w.mu.Unlock()
	}()

	relatorio := RelatorioColeta{
		Erros:             make(map[string]string),
		FontesProcessadas: make([]string, 0, len(w.coletores)),
	}

	total := 0
	for _, c := range w.coletores {
		select {
		case <-ctx.Done():
			relatorio.Erros["cancelamento"] = "Operação cancelada antes de concluir todos os coletores."
			return relatorio
		default:
		}

		pontos, err := c.Coletar(ctx, w.repo)
		relatorio.FontesProcessadas = append(relatorio.FontesProcessadas, c.Nome())
		if err != nil {
			relatorio.Erros[c.Nome()] = err.Error()
			log.Printf("[WORKER_INDICES] Aviso no coletor '%s': %v", c.Nome(), err)
		} else {
			total += pontos
		}
	}

	relatorio.TotalPontosAtualizados = total

	w.mu.Lock()
	w.ultimaExecucao = time.Now().UTC()
	w.ultimoRelatorio = relatorio
	w.mu.Unlock()

	return relatorio
}

// UltimoRelatorio retorna o status da última coleta realizada
func (w *WorkerColetor) UltimoRelatorio() (RelatorioColeta, time.Time, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.ultimoRelatorio, w.ultimaExecucao, w.emExecucao
}
