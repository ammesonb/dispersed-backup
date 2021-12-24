package db

import (
	"database/sql"
	"fmt"
	// Just included for the driver, for now
	_ "github.com/mattn/go-sqlite3"
)

// getConnStr returns a DSN for a given database path
func getConnStr(dbPath string) string {
	return fmt.Sprintf("file:%s?_foreign_keys=true&", dbPath)
}

// OpenDB opens and returns a new connection to the DB
func OpenDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", getConnStr(dbPath))
	if err != nil {
		panic(err)
	}

	return db
}
