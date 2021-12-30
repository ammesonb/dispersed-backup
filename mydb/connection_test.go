package mydb

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test connection string is correct
func TestGetConnStr(t *testing.T) {
	conn := getConnStr("/test/foo")
	assert.Equal(t, "file:/test/foo?_foreign_keys=true", conn, "Expected connection string returned")
}

// Test opening database fails
func TestOpenDBFails(t *testing.T) {
	realOpen := openDB

	openDB = func(driver string, source string) (*sql.DB, error) {
		return nil, fmt.Errorf("RIP")
	}
	defer func() { openDB = realOpen }()

	assert.PanicsWithErrorf(t, "RIP", func() { OpenDB("/tmp/fail") }, "Should panic if DB is not openable")
}

// Test ping database fails
func TestPingDBFails(t *testing.T) {
	realOpen := openDB
	realPing := pingDB

	openDB = func(driver string, source string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}

	pingDB = func(db *sql.DB) error {
		return fmt.Errorf("Ping failed")
	}

	defer func() {
		openDB = realOpen
		pingDB = realPing
	}()

	assert.PanicsWithErrorf(t, "Ping failed", func() { OpenDB("/tmp/fail") }, "Should panic if DB is not pingable")
}

// Check success case, with migration checked
func TestOpenDBSuccess(t *testing.T) {
	realOpen := openDB
	realPing := pingDB
	realMigrate := checkMigration

	stubDB := sql.DB{}
	openDB = func(driver string, source string) (*sql.DB, error) {
		return &stubDB, nil
	}

	pingDB = func(db *sql.DB) error {
		return nil
	}

	var migrated bool = false
	checkMigration = func(_ *sql.DB, _ string) {
		migrated = true
	}

	defer func() {
		openDB = realOpen
		pingDB = realPing
		checkMigration = realMigrate
	}()

	db := OpenDB("/tmp/success")

	assert.Equal(t, &stubDB, db, "Correct DB handle returned")
	assert.True(t, migrated, "Check migration called")
}
