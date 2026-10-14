package cache_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/cache"
)

func TestCacheTTL_BasicoEExpiracao(t *testing.T) {
	c := cache.NovoCacheTTL()

	// 1. Chave inexistente
	if _, ok := c.Obter("inexistente"); ok {
		t.Errorf("Esperava false para chave inexistente")
	}

	// 2. Definir e Obter com TTL válido
	c.Definir("chave1", "valor1", 500*time.Millisecond)
	val, ok := c.Obter("chave1")
	if !ok || val != "valor1" {
		t.Errorf("Esperava 'valor1', obtido: %v (ok=%v)", val, ok)
	}

	// 3. Expirar após TTL
	time.Sleep(550 * time.Millisecond)
	if _, ok := c.Obter("chave1"); ok {
		t.Errorf("Esperava false após expiração do TTL")
	}
}

func TestCacheTTL_AntiThunderingHerdConcorrente(t *testing.T) {
	c := cache.NovoCacheTTL()

	const totalGoroutines = 50
	var chamadasCalculo int64
	var wg sync.WaitGroup

	chave := "processo:5053283-30.2024.8.13.0079"

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := c.ObterOuBuscar(chave, 2*time.Second, func() (any, error) {
				// Simula operação externa cara (ex: chamada ao tribunal)
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt64(&chamadasCalculo, 1)
				return "dados-processo-cnj", nil
			})
			if err != nil {
				t.Errorf("Erro inesperado: %v", err)
			}
			if res != "dados-processo-cnj" {
				t.Errorf("Resultado inesperado: %v", res)
			}
		}()
	}

	wg.Wait()

	// Com o singleflight, mesmo com 50 goroutines concorrentes disparadas ao mesmo tempo,
	// a função de busca externa cara deve ser chamada apenas UMA vez!
	if chamadas := atomic.LoadInt64(&chamadasCalculo); chamadas != 1 {
		t.Errorf("Esperava exatamente 1 chamada ao fetch externo, obtidas: %d", chamadas)
	}
}

func TestCacheTTL_Limpar(t *testing.T) {
	c := cache.NovoCacheTTL()
	c.Definir("k1", "v1", 1*time.Minute)
	c.Definir("k2", "v2", 1*time.Minute)

	c.Limpar()

	if _, ok := c.Obter("k1"); ok {
		t.Errorf("Esperava false para k1 após Limpar()")
	}
	if _, ok := c.Obter("k2"); ok {
		t.Errorf("Esperava false para k2 após Limpar()")
	}
}
