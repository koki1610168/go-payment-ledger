package server

import (
	"context"
	"net/http"
	"fmt"
)

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
	fmt.Println(http.StatusOK)

}

func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	fmt.Println(http.StatusOK)
}
