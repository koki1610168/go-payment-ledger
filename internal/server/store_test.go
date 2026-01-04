package server

import (
	"context"
	"testing"
)

func TestCreatePayment(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx)
	if err != nil {
		t.Errorf("Connection to the database failed, %v", err)
	}
	defer db.Close()

	store := NewStore(db.Pool)

	var client_ID string = "client_001"
	var payment_amount int64 = 50

	var balance_before int64
	err = db.Pool.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1`, client_ID).Scan(&balance_before)

	if err != nil {
		t.Errorf("fetching row failed: %v", err)
	}

	newBalance, err := store.CreatePayment(ctx, client_ID, payment_amount)
	if err != nil {
		t.Errorf("create payment failed: %v", err)
	}
	if balance_before + payment_amount != newBalance {
		t.Errorf("Balance mismatch before and after")
	}
}


func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	db, err := NewDB(ctx)
	if err != nil {
		t.Errorf("Connection to the database failed, %v", err)
	}
	defer db.Close()

	store := NewStore(db.Pool)

	var client_ID string = "client_001"
	newBalance, err := store.GetBalance(ctx, client_ID)
	if err != nil {
		t.Errorf("Getting Balance failed: %v", err)
	}
	var query_balance  int64
	err = db.Pool.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1`, client_ID).Scan(&query_balance)

	if err != nil {
		t.Errorf("fetching row failed: %v", err)
	}
	if query_balance != newBalance {
		t.Errorf("balance mismatch between GetBalance and queried balance")
	}
}

