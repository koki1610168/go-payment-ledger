package server

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"context"
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
	t.Run("Initial setup ", func(t *testing.T) {
		s := StubStore{}
		handler := NewHandler(&s)

		request, _ := http.NewRequest(http.MethodPost, "/payments", nil)
		response := httptest.NewRecorder()

		handler.postPayments(response, request)
		handler.getBalance(response, request)
	})
}
