package main

import (
	"flag"
	"sync"

	"github.com/ammesonb/dispersed-backup/mydb"
)

func main() {
	dbPath := flag.String("db", "/var/lib/dispersed-backup/metadata.db", "Path to database file")
	db := mydb.OpenDB(*dbPath)

	devCommands := make(chan DeviceCommand, 1)
	devResults := make(chan DeviceResult, 1)
	RunManager(db, devCommands, devResults)

	// Device Manager for controlled access to device status & availability
	_ = DevMan{devCommands, devResults, sync.Mutex{}}

	// TODO: create worker pool here

	// TODO: only close dev channels after all workers complete
	close(devCommands)
	close(devResults)
}
