package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() error {
	// Connect to a database file (creates it if it doesn't exist)
	db, err := sql.Open("sqlite3", "legalaid.sqlite3")
	if err != nil {
		panic(err)
	}

	// Check if the connection is OK
	if err := db.Ping(); err != nil {
		panic(err)
	}
	DB = db
	return err
}
