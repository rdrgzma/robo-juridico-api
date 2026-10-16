package auth_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/rdrgzma/robo-juridico-desktop/internal/auth"
	"github.com/rdrgzma/robo-juridico-desktop/internal/domain"
)

const (
	chaveValida = "minha-chave-secreta-muito-segura-com-mais-de-32-chars"
)

func TestHashSHA256(t *testing.T) {
	h1 := auth.HashSHA256("teste")
	h2 := auth.HashSHA256("teste")
	if h1 != h2 {
		t.Errorf("Esperava hashes iguais para mesmo input")
	}
	if len(h1) != 64 {
		t.Errorf("Esperava 64 caracteres hexadecimais, obtidos: %d", len(h1))
	}
}

func TestServicoAuth_ChaveInterna(t *testing.T) {
	chavesInternas := map[string][32]byte{
		"admin-interno": sha256.Sum256([]byte(chaveValida)),
	}

	servico := auth.NovoServicoAuth(nil, chavesInternas, 60)
	ctx := context.Background()

	// 1. Chave válida
	id, err := servico.Autenticar(ctx, chaveValida, "127.0.0.1")
	if err != nil {
		t.Fatalf("Esperava sucesso na autenticação, erro: %v", err)
	}
	if id.NomeCliente != "admin-interno" || !id.EhInterna || id.LimiteMensal != nil {
		t.Errorf("Identidade incorreta: %+v", id)
	}

	// 2. Chave inválida
	_, errInv := servico.Autenticar(ctx, "chave-totalmente-errada-de-tamanho-razoavel", "127.0.0.1")
	if errInv != domain.ErrNaoAutorizado {
		t.Errorf("Esperava ErrNaoAutorizado, obtido: %v", errInv)
	}

	// 3. Chave vazia
	_, errVazia := servico.Autenticar(ctx, "", "127.0.0.1")
	if errVazia != domain.ErrNaoAutorizado {
		t.Errorf("Esperava ErrNaoAutorizado para chave vazia, obtido: %v", errVazia)
	}
}

func TestServicoAuth_RateLimit(t *testing.T) {
	chavesInternas := map[string][32]byte{
		"cliente-limitado": sha256.Sum256([]byte(chaveValida)),
	}

	servico := auth.NovoServicoAuth(nil, chavesInternas, 5) // 5 requisições por minuto
	ctx := context.Background()

	id, err := servico.Autenticar(ctx, chaveValida, "127.0.0.1")
	if err != nil {
		t.Fatalf("Falha na autenticação: %v", err)
	}

	// Consome o burst permitido
	for i := 0; i < 5; i++ {
		if !servico.VerificarLimiteTaxa(id) {
			t.Fatalf("Requisição %d dentro da cota deveria ter sido aceita", i+1)
		}
	}

	// 6ª requisição deve ser bloqueada pelo token bucket
	if servico.VerificarLimiteTaxa(id) {
		t.Errorf("6ª requisição deveria ter sido rejeitada pelo rate limiter")
	}
}
