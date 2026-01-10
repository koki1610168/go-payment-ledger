package server

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"
)
type BalanceResponse struct {
	ClientID string
	Balance int64
	Currency string
}

type ClientStore interface {
	GetBalance(ctx context.Context, clientId string) (int64, string, error) 
	CreatePayment(ctx context.Context, clientID string, amount int64,) (int64, error) 
}

type Handler struct {
	store ClientStore
	mux *http.ServeMux
}

func NewHandler(store ClientStore) *Handler{
	h := &Handler{store: store}
	mux := http.NewServeMux()

	mux.HandleFunc("/payments", h.postPayments)
	mux.HandleFunc("/clients/", h.getBalance)

	h.mux = mux
	return h
}


func (h *Handler) postPayments(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	// We want to call GetBalance
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
	client_id := strings.TrimPrefix(r.URL.Path, "/clients/")
	client_id = strings.TrimSuffix(client_id, "/balance")

	balance, currency, err := h.store.GetBalance(r.Context(), client_id)

	if err != nil {
		http.Error(w, "client not found", http.StatusNotFound)
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BalanceResponse{client_id, balance, currency})
}
