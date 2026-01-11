package server

import (
	"context"
	"net/http"
	"strings"
	"encoding/json"
)

type PaymentRequest struct {
	ClientID string
	Amount int64
	Currency string
}

type Response struct {
	ClientID string
	Balance int64
	Currency string
}


type ClientStore interface {
	GetBalance(ctx context.Context, clientId string) (int64, string, error) 
	CreatePayment(ctx context.Context, clientId string, amount int64,) (int64, error) 
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

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}


func (h *Handler) postPayments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var paymentReq PaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&paymentReq); err != nil {
		http.Error(w, "failed to load request", http.StatusBadRequest)
		return
	}

	client_id := paymentReq.ClientID
	amount := paymentReq.Amount
	currency := paymentReq.Currency

	if client_id == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	if amount <= 0 {
		http.Error(w, "amount must be greater than 0", http.StatusBadRequest)
		return
	}

	if currency == "" {
		http.Error(w, "currency is required", http.StatusBadRequest)
		return
	}

	newBalance, err := h.store.CreatePayment(r.Context(), client_id, amount)
	if err != nil {
		http.Error(w, "failed to initiate payment", http.StatusBadRequest)	
		return
	}

	encodeToJSON(w, client_id, newBalance, currency)


}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	// We want to call GetBalance
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	client_id := strings.TrimPrefix(r.URL.Path, "/clients/")
	client_id = strings.TrimSuffix(client_id, "/balance")

	balance, currency, err := h.store.GetBalance(r.Context(), client_id)

	if err != nil {
		http.Error(w, "failed to get balance", http.StatusNotFound)
		return
	}

	encodeToJSON(w, client_id, balance, currency)
}


func encodeToJSON(w http.ResponseWriter, client_id string, balance int64, currency string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{client_id, balance, currency})
}
