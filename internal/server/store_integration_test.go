package server

import (
	"testing"
	"fmt"
	"net/http"
	"net/http/httptest"
	"bytes"
	"encoding/json"
	"crypto/rand"
	"encoding/hex"

)

func TestTransferCorrectly(t *testing.T) {
	ctx, db, store := newTestStore(t)
	from_client_id := "client_001"
	to_client_id := "client_002"

	oldFromBal := getBalance(t, ctx, db, from_client_id) 
	oldToBal := getBalance(t, ctx, db, to_client_id) 

	handler := NewHandler(store)

	t.Run("getting client_001 balance correctly", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s/balance", from_client_id), nil)
		res := httptest.NewRecorder()
		handler.mux.ServeHTTP(res, req)

		if got := decodePaymentResponseJSON(t, res).Balance; got != oldFromBal {
			t.Errorf("got %d, want %d", got, oldFromBal)
		}

	})

	t.Run("getting client_002 balance correctly", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s/balance", to_client_id), nil)
		res := httptest.NewRecorder()
		handler.mux.ServeHTTP(res, req)

		if got := decodePaymentResponseJSON(t, res).Balance; got != oldToBal {
			t.Errorf("got %d, want %d", got, oldToBal)
		}

	})

	t.Run("transfer from client_001 to client_002 correctly", func(t *testing.T) {
		paymentAmount := int64(300)
		expectedFromBalance := oldFromBal - paymentAmount
		expectedToBalance := oldToBal + paymentAmount
		beforeFromEntriesCount := countLedgerEntries(t, ctx, db, from_client_id)
		beforeToEntriesCount := countLedgerEntries(t, ctx, db, to_client_id)
		

		idemPotencyKey, _ := NewIdempotencyKey(t)
        var buf bytes.Buffer
        json.NewEncoder(&buf).Encode(map[string]any{
			"from_client_id": from_client_id,
			"to_client_id": to_client_id,
			"amount": paymentAmount,
			"idempotencyKey": idemPotencyKey,
        })

        req, _ := http.NewRequest(http.MethodPost, "/transfer", &buf)
        res := httptest.NewRecorder()
        handler.mux.ServeHTTP(res, req)

		transferResponse := decodeTransferResponseJSON(t, res)
		afterFromEntriesCount := countLedgerEntries(t, ctx, db, from_client_id)
		afterToEntriesCount := countLedgerEntries(t, ctx, db, to_client_id)

		if transferResponse.FromNewBalance != expectedFromBalance || transferResponse.ToNewBalance != expectedToBalance {
			t.Errorf("from new balance: got %d, want %d \n to new balance: got %d, want %d", transferResponse.FromNewBalance,
				expectedFromBalance, transferResponse.ToNewBalance, expectedToBalance)
		}

		if beforeFromEntriesCount + 1 != afterFromEntriesCount || beforeToEntriesCount + 1 != afterToEntriesCount {
			t.Errorf("ledger entries were not correctly updated \n from: got %d, want %d \n to: got %d, want %d", 
				afterFromEntriesCount, beforeFromEntriesCount+1, afterToEntriesCount, beforeToEntriesCount+1)
		}
	})

}


func TestGetLedger(t *testing.T) {
	_, _, store := newTestStore(t)
	handler := NewHandler(store)
	client := "client_001"
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/clients/%s/ledger", client), nil)
	res := httptest.NewRecorder()
	handler.mux.ServeHTTP(res, req)

	fmt.Println(res.Body)

}


func NewIdempotencyKey(t testing.TB) (string, error) {
	b := make([]byte, 32) // 256-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
