package main

import (
	"fmt"
	"log"
	"github.com/koki1610168/go-payment-ledger/internal/server"
)

func main() {
	server.ConnectDatabase()
	defer server.CloseDatabase()
	fmt.Println("Ok!")

	
	user := server.User{
		Email: "koki@test.com",
		Balance: 1000000,
	}
	_, err := server.AddUser(user)
	if err != nil {
		log.Fatal("AddUser failed %v", err)	
	}
	fmt.Println("Good")

	

}
