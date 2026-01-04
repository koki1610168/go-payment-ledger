package main

import (
	"fmt"
	"context"
	"log"
	"github.com/koki1610168/go-payment-ledger/internal/server"

)

func main() {
	ctx := context.Background()

	db, err := server.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("DB Connected")

}
