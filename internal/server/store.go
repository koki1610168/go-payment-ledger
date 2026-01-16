package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClientNotFound = errors.New("client not found")
var ErrInsufficientBalance = errors.New("insufficient balance")

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
	idempotencyKey string,
) (int64, error) {
	// Start Transaction. THhis ensures that sequence of SQL executions run atomically
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	// This to make sure no table update is half-done
	defer tx.Rollback(ctx)

	if idempotencyKey != "" {
		var existing_balance int64
		err := tx.QueryRow(ctx,
			`SELECT c.balance
			FROM ledger_entries le
			JOIN clients c ON c.client_id = le.client_id
			WHERE le.idempotency_key = $1
			`, idempotencyKey).Scan(&existing_balance)

		if err == nil {
			return existing_balance, nil
		}

		if err != pgx.ErrNoRows {
			return 0, err
		}
	}

	var balance int64
	err = tx.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1 FOR UPDATE`,
		clientID).Scan(&balance)

	if err == pgx.ErrNoRows {
		return 0, ErrClientNotFound
	}
	if err != nil {
		return 0, err
	}

	newBalance := balance + amount

	if newBalance < 0 {
		return 0, ErrInsufficientBalance
	}

	_, err = tx.Exec(ctx,
		`UPDATE clients SET balance = $1 WHERE client_id = $2`, newBalance, clientID,
	)

	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO ledger_entries (entry_id, client_id, amount, idempotency_key) VALUES (gen_random_uuid(), $1, $2, $3)`, clientID, amount, idempotencyKey)

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

func (s *Store) Transfer(
	ctx context.Context,
	fromClientId string,
	toClientId string,
	amount int64,
	idempotencyKey string,
) (int64, int64, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, 0, err
	}

	// This to make sure no table update is half-done
	defer tx.Rollback(ctx)

	if idempotencyKey != "" {
		var count int
		err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM ledger_entries WHERE idempotency_key = $1`, 
			idempotencyKey).Scan(&count)
		
		if err != nil {
			return 0, 0, err
		}

		if count > 0 {
			var fromBalance, toBalance int64
			err = tx.QueryRow(ctx,
			`SELECT balance FROM clients WHERE client_id = $1`, fromClientId).Scan(&fromBalance)
			if err != nil {
				return 0, 0, err
			}

			err = tx.QueryRow(ctx,
			`SELECT balance FROM clients WHERE client_id = $1`, toClientId).Scan(&toBalance)
			if err != nil {
				return 0, 0, err
			}
			return fromBalance, toBalance, nil
		}
	}

	var oldFromBalance, oldToBalance int64
	err = tx.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1`, 
		fromClientId).Scan(&oldFromBalance)

	if err == pgx.ErrNoRows {
		return 0, 0, ErrClientNotFound
	}
	if err != nil {
		return 0, 0, err
	}

	err = tx.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1`, 
		toClientId).Scan(&oldToBalance)

	if err == pgx.ErrNoRows {
		return 0, 0, ErrClientNotFound
	}
	if err != nil {
		return 0, 0, err
	}

	newFromBalance := oldFromBalance - amount
	newToBalance := oldToBalance + amount

	_, err = tx.Exec(ctx,
		`UPDATE clients SET balance = $1 WHERE client_id = $2`, newFromBalance, fromClientId)
	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec(ctx,
		`UPDATE clients SET balance = $1 WHERE client_id = $2`, newToBalance, toClientId)
	if err != nil {
		return 0, 0, err
	}


	_, err = tx.Exec(ctx,
		`INSERT INTO ledger_entries (entry_id, client_id, amount, idempotency_key) VALUES (gen_random_uuid(), $1, $2, $3)`, 
		fromClientId, -amount, idempotencyKey)

	if err != nil {
		return 0, 0, err
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO ledger_entries (entry_id, client_id, amount) VALUES (gen_random_uuid(), $1, $2)`, 
		toClientId, amount)

	if err != nil {
		return 0, 0, err
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, 0, err
	}
	return newFromBalance, newToBalance, nil
}
