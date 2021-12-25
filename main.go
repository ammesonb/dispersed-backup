package main

import (
	"flag"

	"github.com/ammesonb/dispersed-backup/mydb"
)

func main() {
	dbPath := flag.String("db", "/var/lib/dispersed-backup/metadata.db", "Path to database file")
	_ = mydb.OpenDB(*dbPath)
}
