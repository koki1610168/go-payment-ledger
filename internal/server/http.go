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

type User struct {
	ID int64
	Email string
	Balance int64
	CretedAt time.Time
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

func UserByEmail(email string) (User, error) {
	var usr User

	row := db.QueryRow(context.Background(), "SELECT * FROM users WHERE email = $1", email)
	if err := row.Scan(&usr.ID, &usr.Email, &usr.Balance, &usr.CretedAt); err != nil {
		return usr, err
	}
	return usr, nil

}

func AddUser(usr User) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database not connected")
	}
	result, err := db.Exec(context.Background(), "INSERT INTO users (email, balance) VALUES ($1, $2)", usr.Email, usr.Balance)
	if err != nil {
		return 0, err
	}
	fmt.Println(result)
	return 0, nil
}



