package storage

import (
	"context"
	"time"
)

type ChaveCliente struct {
	ChaveID         int
	ClienteID       int
	NomeCliente     string
	Ativo           bool
	Plano           string
	LimiteMensal    int
	LimitePorMinuto int
}

type BancoDados interface {
	ValidarChave(ctx context.Context, hashChave string) (*ChaveCliente, error)
	ReservarCota(ctx context.Context, clienteID, chaveID, limiteMensal int, metodo, caminho, ip string, cobravel bool) (permitido bool, usoID int64, usadoNoMes int, err error)
	FinalizarUso(ctx context.Context, usoID int64, status, ms int, rota string) error
	ObterConsumoMes(ctx context.Context, clienteID int) (int, error)
	Fechar() error
}

func InicioDoMesBrasil() time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*3600)
	}
	agora := time.Now().In(loc)
	return time.Date(agora.Year(), agora.Month(), 1, 0, 0, 0, 0, loc).UTC()
}

func InicioDoProximoMesBrasil() time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*3600)
	}
	agora := time.Now().In(loc)
	ano := agora.Year()
	mes := agora.Month() + 1
	if mes > 12 {
		mes = 1
		ano++
	}
	return time.Date(ano, mes, 1, 0, 0, 0, 0, loc).UTC()
}
