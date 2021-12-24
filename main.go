package main

import (
	"flag"
	"github.com/ammesonb/dispersed-backup/db"
)

func main() {
	dbPath := flag.String("db", "/var/lib/dispersed-backup/metadata.db", "Path to database file")
	_ = db.OpenDB(*dbPath)
}
