package cache

import (
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

type itemCache struct {
	valor    any
	expiraEm time.Time
}

type CacheTTL struct {
	dados sync.Map
	sf    singleflight.Group
}

func NovoCacheTTL() *CacheTTL {
	c := &CacheTTL{}
	go c.limpezaPeriodica()
	return c
}

func (c *CacheTTL) Obter(chave string) (any, bool) {
	val, ok := c.dados.Load(chave)
	if !ok {
		return nil, false
	}
	item := val.(itemCache)
	if time.Now().After(item.expiraEm) {
		c.dados.Delete(chave)
		return nil, false
	}
	return item.valor, true
}

func (c *CacheTTL) Definir(chave string, valor any, ttl time.Duration) {
	c.dados.Store(chave, itemCache{
		valor:    valor,
		expiraEm: time.Now().Add(ttl),
	})
}

// ObterOuBuscar busca no cache ou executa fn garantindo que chamadas concorrentes
// para a mesma chave executem fn apenas uma única vez (Anti-Thundering Herd).
func (c *CacheTTL) ObterOuBuscar(chave string, ttl time.Duration, fn func() (any, error)) (any, error) {
	if val, ok := c.Obter(chave); ok {
		return val, nil
	}

	res, err, _ := c.sf.Do(chave, func() (any, error) {
		// Dupla checagem sob o singleflight
		if val, ok := c.Obter(chave); ok {
			return val, nil
		}
		v, err := fn()
		if err != nil {
			return nil, err // Erros nunca entram no cache
		}
		c.Definir(chave, v, ttl)
		return v, nil
	})

	return res, err
}

func (c *CacheTTL) Limpar() {
	c.dados.Range(func(key, value any) bool {
		c.dados.Delete(key)
		return true
	})
}

func (c *CacheTTL) limpezaPeriodica() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		agora := time.Now()
		c.dados.Range(func(key, value any) bool {
			item := value.(itemCache)
			if agora.After(item.expiraEm) {
				c.dados.Delete(key)
			}
			return true
		})
	}
}
