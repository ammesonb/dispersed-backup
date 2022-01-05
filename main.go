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
	devCtx := DevCtx{devCommands, devResults, &sync.Mutex{}}

	// TODO: create worker pool here
	wg := sync.WaitGroup{}

	var threads int = 4 // TODO; CPU thread count/setting
	workerChannels := make(map[int]chan<- Job)

	var interrupt bool = false
	var hardInterrupt bool = false

	workRequests := make(chan WorkRequest, threads*2)
	// Progress needs more buffering since may see more progress than
	// there are workers
	workProgress := make(chan WorkProgress, threads*10)
	workComplete := make(chan WorkComplete, threads*2)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		jobChan := make(chan Job)
		workerChannels[i] = jobChan

		StartWorker(&WorkerContext{
			db:              db,
			workerID:        i,
			interrupted:     &interrupt,
			hardInterrupted: &hardInterrupt,
			DevCtx: DevCtx{
				devLock:     devCtx.devLock,
				devCommands: devCtx.devCommands,
				devResults:  devCtx.devResults,
			},
			requests: workRequests,
			jobs:     jobChan,
			progress: workProgress,
			complete: workComplete,
		})

	}

	// TODO: only close dev channels after all workers complete
	close(devCommands)
	close(devResults)
}
