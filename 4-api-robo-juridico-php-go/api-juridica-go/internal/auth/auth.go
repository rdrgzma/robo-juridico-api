package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/domain"
	"github.com/rdrgzma/robo-juridico-api/api-juridica-go/internal/storage"
	"golang.org/x/time/rate"
)

type Identidade struct {
	ClienteID       int
	NomeCliente     string
	Plano           string
	LimiteMensal    *int
	LimitePorMinuto int
	ChaveID         int
	EhInterna       bool
}

type ServicoAuth struct {
	banco            storage.BancoDados
	chavesInternas   map[string][32]byte
	limitersCliente  sync.Map
	limitersIPFalhas sync.Map
	limitePadraoRPM  int
}

func NovoServicoAuth(banco storage.BancoDados, chavesInternas map[string][32]byte, limitePadraoRPM int) *ServicoAuth {
	if limitePadraoRPM <= 0 {
		limitePadraoRPM = 60
	}
	return &ServicoAuth{
		banco:           banco,
		chavesInternas:  chavesInternas,
		limitePadraoRPM: limitePadraoRPM,
	}
}

func HashSHA256(chave string) string {
	h := sha256.Sum256([]byte(chave))
	return hex.EncodeToString(h[:])
}

func (s *ServicoAuth) Autenticar(ctx context.Context, chave string, ip string) (*Identidade, error) {
	if s.ipBloqueado(ip) {
		return nil, domain.NovoErroComRetry(429, "muitas_tentativas", "Muitas tentativas com chave inválida. Aguarde.", 60)
	}

	if chave == "" || len(chave) > 200 {
		s.registrarFalhaIP(ip)
		return nil, domain.ErrNaoAutorizado
	}

	hashRecebido := sha256.Sum256([]byte(chave))
	hashHex := hex.EncodeToString(hashRecebido[:])

	// 1. Checa chaves internas sem cota (.env)
	for nome, hashEsperado := range s.chavesInternas {
		if hmac.Equal(hashRecebido[:], hashEsperado[:]) {
			return &Identidade{
				NomeCliente:     nome,
				Plano:           "interno",
				LimiteMensal:    nil,
				LimitePorMinuto: s.limitePadraoRPM,
				EhInterna:       true,
			}, nil
		}
	}

	// 2. Checa chaves no banco SQLite
	if s.banco != nil {
		chaveCli, err := s.banco.ValidarChave(ctx, hashHex)
		if err != nil {
			return nil, domain.ErrErroInterno
		}
		if chaveCli != nil {
			if !chaveCli.Ativo {
				return nil, domain.ErrContaSuspensa
			}

			lim := chaveCli.LimiteMensal
			return &Identidade{
				ClienteID:       chaveCli.ClienteID,
				NomeCliente:     chaveCli.NomeCliente,
				Plano:           chaveCli.Plano,
				LimiteMensal:    &lim,
				LimitePorMinuto: chaveCli.LimitePorMinuto,
				ChaveID:         chaveCli.ChaveID,
				EhInterna:       false,
			}, nil
		}
	}

	s.registrarFalhaIP(ip)
	return nil, domain.ErrNaoAutorizado
}

func (s *ServicoAuth) VerificarLimiteTaxa(identidade *Identidade) bool {
	rpm := identidade.LimitePorMinuto
	if rpm <= 0 {
		rpm = s.limitePadraoRPM
	}

	key := identidade.NomeCliente
	val, _ := s.limitersCliente.LoadOrStore(key, rate.NewLimiter(rate.Every(time.Minute/time.Duration(rpm)), rpm))
	limiter := val.(*rate.Limiter)

	return limiter.Allow()
}

func (s *ServicoAuth) ipBloqueado(ip string) bool {
	val, ok := s.limitersIPFalhas.Load(ip)
	if !ok {
		return false
	}
	limiter := val.(*rate.Limiter)
	return !limiter.Allow()
}

func (s *ServicoAuth) registrarFalhaIP(ip string) {
	// Permite 10 falhas a cada 15 minutos antes de travar
	val, _ := s.limitersIPFalhas.LoadOrStore(ip, rate.NewLimiter(rate.Every(15*time.Minute/10), 10))
	limiter := val.(*rate.Limiter)
	_ = limiter.Allow()
}
