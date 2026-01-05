package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClientNotFound = errors.New("client not found")

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) CreatePayment(
	ctx context.Context,
	clientID string,
	amount int64,
) (int64, error) {
	// Start Transaction. THhis ensures that sequence of SQL executions run atomically
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	// This to make sure no table update is half-done
	defer tx.Rollback(ctx)

	var balance int64
	err = tx.QueryRow(ctx, 
		`SELECT balance FROM clients WHERE client_id = $1 FOR UPDATE`, 
		clientID,).Scan(&balance)

	if err == pgx.ErrNoRows {
		return 0, ErrClientNotFound
	}
	if err != nil {
		return 0, err 
	}

	newBalance := balance + amount

	_, err = tx.Exec(ctx,
		`UPDATE clients SET balance = $1 WHERE client_id = $2`, newBalance, clientID,
	)

	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO ledger_entries (entry_id, client_id, amount) VALUES (gen_random_uuid(), $1, $2)`, clientID, amount)

	if err != nil {
		return 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return newBalance, nil
}

 func (s *Store) GetBalance(
	ctx context.Context,
	clientId string,
) (int64, string, error) {

	var balance int64
	var currency string
	err := s.db.QueryRow(ctx,
		`SELECT balance, currency FROM clients WHERE client_id = $1`, clientId).Scan(&balance, &currency)

	if err == pgx.ErrNoRows {
		return 0, "", ErrClientNotFound
	}
	if err != nil {
		return 0, "", err 
	}

	return balance, currency, nil
}
