package server

import (
	"context"
	"testing"
	"fmt"
	"os"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") == "" {
		fmt.Println("DATABASE_URL not set")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func newTestStore(t *testing.T) (context.Context, *DB, *Store) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	db, err := NewDB(ctx)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store := NewStore(db.Pool)
	return ctx, db, store
}

func seedClient(t *testing.T, ctx context.Context, db *DB, clientID string, 
	balance int64, currency string) {

	t.Helper()
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO clients (client_id, balance, currency)
		VALUES ($1, $2, $3)
		ON CONFLICT (client_id) DO UPDATE
		SET balance = EXCLUDED.balance, currency = EXCLUDED.currency`,
		clientID, balance, currency,
	)
	if err != nil {
		t.Fatalf("seed client: %v", err)
	}
}

func getBalance(t *testing.T, ctx context.Context, db *DB, clientID string) int64 {
	t.Helper()

	var bal int64
	if err := db.Pool.QueryRow(ctx,
		`SELECT balance FROM clients WHERE client_id = $1`,
		clientID,
	).Scan(&bal); err != nil {
		t.Fatalf("query balance: %v", err)
	}
	return bal
}

func countLedgerEntries(t *testing.T, ctx context.Context, db *DB, clientID string) int64 {
	t.Helper()

	var n int64
	if err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM ledger_entries WHERE client_id = $1`,
		clientID,
	).Scan(&n); err != nil {
		t.Fatalf("count ledger entries: %v", err)
	}
	return n
}

/*
func TestCreatePayment_AppendsLedgerAndUpdatesBalance(t *testing.T) {
	ctx, db, store := newTestStore(t) 
	clientID := fmt.Sprintf("test_client_%d", time.Now().UnixNano())
	seedClient(t, ctx, db, clientID, 100, "USD")

	beforeBal := getBalance(t, ctx, db, clientID)
	beforeEntries := countLedgerEntries(t, ctx, db, clientID)

	amount := int64(-30)
	newBal, err := store.CreatePayment(ctx, clientID, amount)
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}

	afterBal := getBalance(t, ctx, db, clientID)
	afterEntries := countLedgerEntries(t, ctx, db, clientID)

	if newBal != beforeBal + amount {
		t.Fatalf("returned balance mismatch: got=%d want=%d", newBal, beforeBal+amount)
	}
	if afterBal != beforeBal+amount {
		t.Fatalf("db balance mismatch: got=%d want=%d", afterBal, beforeBal+amount)
	}
	if afterEntries != beforeEntries+1 {
		t.Fatalf("ledger entry not appended: got=%d want=%d", afterEntries, beforeEntries+1)
	}
}
*/
