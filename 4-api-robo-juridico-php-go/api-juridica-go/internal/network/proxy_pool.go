package network

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

type ProxyPool struct {
	proxies       []string
	paralelo      int
	maxTentativas int
	mu            sync.RWMutex
}

func NovoProxyPool(fixos []string, gwHost, gwLogin, gwSenha string, gwPortaInicial, gwPortas int, paralelo, maxTentativas int) *ProxyPool {
	var pool []string

	// Expande gateway residencial em 1 IP por porta
	if gwHost != "" && gwLogin != "" && gwSenha != "" && gwPortas > 0 {
		for p := gwPortaInicial; p < gwPortaInicial+gwPortas; p++ {
			pool = append(pool, fmt.Sprintf("http://%s:%s@%s:%d", gwLogin, gwSenha, gwHost, p))
		}
	}

	// Adiciona proxies fixos
	for _, p := range fixos {
		if p != "" {
			pool = append(pool, p)
		}
	}

	if paralelo <= 0 {
		paralelo = 4
	}
	if maxTentativas <= 0 {
		maxTentativas = 16
	}

	return &ProxyPool{
		proxies:       pool,
		paralelo:      paralelo,
		maxTentativas: maxTentativas,
	}
}

func (p *ProxyPool) TemProxies() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.proxies) > 0
}

func (p *ProxyPool) ObterLista() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cp := make([]string, len(p.proxies))
	copy(cp, p.proxies)
	return cp
}

// ExecutarComCorrida executa chamadas em paralelo através dos proxies e retorna a primeira resposta definitiva.
func (p *ProxyPool) ExecutarComCorrida(ctx context.Context, cliente *ClienteHTTP, endpoint string, params url.Values, cabecalhos map[string][]string) (*RespostaHTTP, error) {
	if !p.TemProxies() {
		// Sem proxies: chamada direta
		return cliente.Get(ctx, endpoint, params, cabecalhos, "")
	}

	lista := p.ObterLista()
	// Embaralha para distribuir carga
	rand.Shuffle(len(lista), func(i, j int) {
		lista[i], lista[j] = lista[j], lista[i]
	})

	teto := len(lista)
	if teto > p.maxTentativas {
		teto = p.maxTentativas
	}

	type resultado struct {
		resp *RespostaHTTP
		err  error
	}

	for inicio := 0; inicio < teto; inicio += p.paralelo {
		fim := inicio + p.paralelo
		if fim > teto {
			fim = teto
		}
		lote := lista[inicio:fim]

		ctxLote, cancel := context.WithCancel(ctx)
		ch := make(chan resultado, len(lote))

		for _, proxyURL := range lote {
			go func(pURL string) {
				r, err := cliente.Get(ctxLote, endpoint, params, cabecalhos, pURL)
				ch <- resultado{resp: r, err: err}
			}(proxyURL)
		}

		var ultimaResposta *RespostaHTTP
		var recebidos int

		for recebidos < len(lote) {
			select {
			case <-ctx.Done():
				cancel()
				return nil, ctx.Err()
			case res := <-ch:
				recebidos++
				if res.err == nil && res.resp != nil {
					// Resposta definitiva encontrada: cancela as outras requisições do lote
					if res.resp.StatusCode == 200 || res.resp.StatusCode == 400 || res.resp.StatusCode == 404 {
						cancel()
						return res.resp, nil
					}
					ultimaResposta = res.resp
				}
			case <-time.After(15 * time.Second):
				cancel()
				recebidos = len(lote)
			}
		}
		cancel()

		if ultimaResposta != nil && ultimaResposta.StatusCode == 200 {
			return ultimaResposta, nil
		}
	}

	return nil, fmt.Errorf("nenhum proxy do pool obteve resposta com sucesso")
}
