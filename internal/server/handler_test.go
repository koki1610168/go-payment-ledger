package server

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"context"
	"encoding/json"
	"reflect"
)


type StubStore struct {


}

func (s *StubStore) GetBalance(ctx context.Context, clientId string) (int64, string, error) {
	return 0, "", nil
}

func (s *StubStore) CreatePayment(ctx context.Context, clientID string, amount int64,) (int64, error) {
	return 0, nil
}

func TestHandler(t *testing.T) {
	// The request should be json and the resonse is also json
	// I want to make a fake dateabase
	t.Run("getting balance from an existing client ", func(t *testing.T) {
		_, _, store  := newTestStore(t)

		handler := NewHandler(store)

		request, _ := http.NewRequest(http.MethodGet, "/clients/client_001/balance", nil)
		response := httptest.NewRecorder()

		//handler.postPayments(response, request)
		handler.getBalance(response, request)
		want := BalanceResponse{
			ClientID: "client_001",
			Balance: 5150,
			Currency: "JPY",
		}

		var balanceClient BalanceResponse
		err := json.NewDecoder(response.Body).Decode(&balanceClient)
		if err != nil {
			t.Errorf("error occured")
		}
		assertEqualBalanceResponse(t, balanceClient, want)
	})
}

func assertEqualBalanceResponse(t testing.TB, got, want BalanceResponse) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

