package server

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"context"
	"encoding/json"
	"reflect"
	"errors"
	"bytes"
)

var ErrorNotFound = errors.New("error not found")

type StubStore struct {
	balances map[string]int64
	currencies map[string]string
	idempotencyKeys map[string]int64
}

func NewStubClient() *StubStore {
	return &StubStore{
		balances: make(map[string]int64),
		currencies: make(map[string]string),
		idempotencyKeys: make(map[string]int64),
	}
}

func (s *StubStore) SeedClient(clientId string, balance int64, currency string) {
	s.balances[clientId] = balance
	s.currencies[clientId] = currency
}

func (s *StubStore) GetBalance(ctx context.Context, clientId string) (int64, string, error) {
	b, ok := s.balances[clientId]
	if !ok {
		return 0, "", ErrorNotFound
	}
	return b, s.currencies[clientId], nil

}

func (s *StubStore) CreatePayment(ctx context.Context, clientId string, amount int64, idempotencyKey string) (int64, error) {
	_, ok := s.balances[clientId]
	if !ok {
		return 0, ErrorNotFound
	}

	if idempotencyKey != "" {
		_, ok := s.idempotencyKeys[idempotencyKey]
		if ok {
			return s.idempotencyKeys[idempotencyKey], nil
		}
	}
	s.balances[clientId] += amount

	if idempotencyKey != "" {
		s.idempotencyKeys[idempotencyKey] = s.balances[clientId]
	}
	return s.balances[clientId], nil

}

func TestHandler(t *testing.T) {
	// The request should be json and the resonse is also json
	// I want to make a fake dateabase
	t.Run("getting balance from an existing client ", func(t *testing.T) {
		store := NewStubClient()
		store.SeedClient("client_001", 10000, "JPY")

		handler := NewHandler(store)

		request, _ := http.NewRequest(http.MethodGet, "/clients/client_001/balance", nil)
		response := httptest.NewRecorder()

		//handler.postPayments(response, request)
		handler.mux.ServeHTTP(response, request)
		want := Response{
			ClientID: "client_001",
			Balance: 10000,
			Currency: "JPY",
		}

		balanceClient := decodeJSON(t, response)
		assertEqualResponse(t, balanceClient, want)
	})

	t.Run("payment request from an exisiting client successful", func(t *testing.T) {
		store := NewStubClient()
		store.SeedClient("client_001", 10000, "JPY")

		want := int64(10000 + 1400)

 		handler := NewHandler(store)

		idempotencKey := "payment-12345"

		var buf bytes.Buffer
		reqJSON := map[string]any{
			"clientID": "client_001",
			"amount": 1400,
			"currency": "JPY",
			"idempotencyKey": idempotencKey,
		}

		err := json.NewEncoder(&buf).Encode(reqJSON)
		if err != nil {
			t.Fatalf("Failed to encode request JSON, %v", err)
		}

		request, _ := http.NewRequest(http.MethodPost, "/payments", &buf)
		response := httptest.NewRecorder()

		handler.mux.ServeHTTP(response, request)

		balanceClient := decodeJSON(t, response)
		afterBalance := balanceClient.Balance

		assertEqualBalance(t, afterBalance, want)
	})
}

func TestDoulbeCharge(t *testing.T) {
	store := NewStubClient()
	initialBalance := int64(10000)
	store.SeedClient("client_001", initialBalance, "JPY")

	handler := NewHandler(store)
	paymentAmount := int64(1000)

	idempotencyKey := "pay-1234"

	makePayment := func() int64 {
        var buf bytes.Buffer
        json.NewEncoder(&buf).Encode(map[string]any{
            "clientID": "client_001",
            "amount":   paymentAmount,
            "currency": "JPY",
            "idempotencyKey": idempotencyKey,
        })
        req, _ := http.NewRequest(http.MethodPost, "/payments", &buf)
        res := httptest.NewRecorder()
        handler.mux.ServeHTTP(res, req)
        return decodeJSON(t, res).Balance
    }

	balance1 := makePayment()
	balance2 := makePayment()

	expectedBalance := initialBalance + paymentAmount

	assertEqualBalance(t, balance1, expectedBalance)
	assertEqualBalance(t, balance2, expectedBalance)

}

func decodeJSON(t testing.TB, response *httptest.ResponseRecorder) Response {
	t.Helper()
	var balanceClient Response
	err := json.NewDecoder(response.Body).Decode(&balanceClient)
	if err != nil {
		t.Errorf("error occured")
	}
	return balanceClient
}

func assertEqualBalance(t testing.TB, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("got %d want %d", got, want)

	}
}

func assertEqualResponse(t testing.TB, got, want Response) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

