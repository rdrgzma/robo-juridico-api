package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type BancoSQLite struct {
	db *sql.DB
}

func NovoBancoSQLite(caminho string) (*BancoSQLite, error) {
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(15000)&_pragma=foreign_keys(ON)", caminho)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &BancoSQLite{db: db}, nil
}

func (s *BancoSQLite) Fechar() error {
	return s.db.Close()
}

func (s *BancoSQLite) ValidarChave(ctx context.Context, hashChave string) (*ChaveCliente, error) {
	query := `
		SELECT k.id, c.id, c.nome, c.ativo, c.plano, c.limite_mensal, c.limite_por_minuto
		FROM chaves k
		JOIN clientes c ON c.id = k.cliente_id
		WHERE k.hash = ? AND k.revogada_em IS NULL
	`

	var chave ChaveCliente
	var ativoInt int
	err := s.db.QueryRowContext(ctx, query, hashChave).Scan(
		&chave.ChaveID,
		&chave.ClienteID,
		&chave.NomeCliente,
		&ativoInt,
		&chave.Plano,
		&chave.LimiteMensal,
		&chave.LimitePorMinuto,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Chave não encontrada
		}
		return nil, err
	}

	chave.Ativo = ativoInt == 1
	return &chave, nil
}

func (s *BancoSQLite) ReservarCota(ctx context.Context, clienteID, chaveID, limiteMensal int, metodo, caminho, ip string, cobravel bool) (bool, int64, int, error) {
	inicioMesStr := InicioDoMesBrasil().Format(time.RFC3339)
	agoraStr := time.Now().UTC().Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, 0, 0, err
	}
	defer tx.Rollback()

	// Contagem atômica de consultas cobradas no mês
	var usado int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM uso WHERE cliente_id = ? AND cobrado = 1 AND momento >= ?
	`, clienteID, inicioMesStr).Scan(&usado)
	if err != nil {
		return false, 0, 0, err
	}

	cobravelInt := 0
	if cobravel {
		cobravelInt = 1
	}

	if cobravel && usado >= limiteMensal {
		_, _ = tx.ExecContext(ctx, `
			INSERT INTO uso (cliente_id, chave_id, momento, metodo, caminho, status, ms, ip, cobravel, cobrado)
			VALUES (?, ?, ?, ?, ?, 402, 0, ?, 1, 0)
		`, clienteID, chaveID, agoraStr, metodo, caminho, ip)
		_ = tx.Commit()
		return false, 0, usado, nil
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO uso (cliente_id, chave_id, momento, metodo, caminho, ip, cobravel, cobrado)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, clienteID, chaveID, agoraStr, metodo, caminho, ip, cobravelInt, cobravelInt)
	if err != nil {
		return false, 0, 0, err
	}

	usoID, err := res.LastInsertId()
	if err != nil {
		return false, 0, 0, err
	}

	_, _ = tx.ExecContext(ctx, `UPDATE chaves SET ultimo_uso = ? WHERE id = ?`, agoraStr, chaveID)

	if err := tx.Commit(); err != nil {
		return false, 0, 0, err
	}

	totalUsado := usado
	if cobravel {
		totalUsado++
	}

	return true, usoID, totalUsado, nil
}

func (s *BancoSQLite) FinalizarUso(ctx context.Context, usoID int64, status, ms int, rota string) error {
	query := `
		UPDATE uso SET status = ?, ms = ?, rota = ?,
		cobrado = CASE WHEN cobravel = 1 AND ? < 400 THEN 1 ELSE 0 END
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, status, ms, rota, status, usoID)
	return err
}

func (s *BancoSQLite) ObterConsumoMes(ctx context.Context, clienteID int) (int, error) {
	inicioMesStr := InicioDoMesBrasil().Format(time.RFC3339)
	var usado int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM uso WHERE cliente_id = ? AND cobrado = 1 AND momento >= ?
	`, clienteID, inicioMesStr).Scan(&usado)
	return usado, err
}
