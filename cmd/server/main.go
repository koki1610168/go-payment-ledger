package main

import (
	"fmt"
	//"log"
	"github.com/koki1610168/go-payment-ledger/internal/server"
)

func main() {
	server.ConnectDatabase()
	defer server.CloseDatabase()
	fmt.Println("Ok!")

}
