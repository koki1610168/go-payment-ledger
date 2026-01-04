package server

import (
	// "database/sql"
	"context"
	"fmt"
	"os"
	"time"
	"github.com/jackc/pgx/v5"	
)

var db *pgx.Conn

type Client struct {
	Client_ID string
	Balacne int64
	Currency string
	CreatedAt time.Time
}

func ConnectDatabase() {
	var err error
	db, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	fmt.Println(os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database")
		os.Exit(1)
	}

	fmt.Println("Connected!")
}

func CloseDatabase() {
	if db != nil {
		db.Close(context.Background())
		fmt.Println("Database closed")
	}
}

