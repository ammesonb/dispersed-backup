package mydb

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/file"

	// Just included for the driver, for now
	_ "github.com/mattn/go-sqlite3"
)

var openDB = sql.Open
var withInstance = sqlite3.WithInstance
var migrateInstance = migrate.NewWithInstance

// getConnStr returns a DSN for a given database path
func getConnStr(dbPath string) string {
	return fmt.Sprintf("file:%s?_foreign_keys=true", dbPath)
}

// OpenDB opens and returns a new connection to the DB
func OpenDB(dbPath string) *sql.DB {
	connStr := getConnStr(dbPath)
	db, err := openDB("sqlite3", connStr)
	if err != nil {
		panic(err)
	}

	err = pingDB(db)
	if err != nil {
		panic(err)
	}

	checkMigration(db, dbPath)

	return db
}

// Pings a DB for liveliness - variable for mocking in tests
var pingDB = func(db *sql.DB) error {
	return db.Ping()
}

// Check database for any needed migrations
var checkMigration = func(db *sql.DB, dbPath string) {
	driver, err := withInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	migrationSource, err := getMigrations()
	if err != nil {
		panic(err)
	}

	migration, err := migrateInstance("file", migrationSource, "db", driver)
	if err != nil {
		panic(err)
	}
	if err = migration.Up(); err != nil {
		panic(err)
	}
}

var getMigrations = func() (source.Driver, error) {
	return (&file.File{}).Open("file://mydb/migrations")
}
