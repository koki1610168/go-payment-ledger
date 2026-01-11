package main

import (
	"context"
	"log"
	"net/http"

	"github.com/koki1610168/go-payment-ledger/internal/server"
)

func main() {
	ctx := context.Background()

	db, err := server.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := server.NewStore(db.Pool)
	handler := server.NewHandler(store)

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))

}
