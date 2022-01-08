package main

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

var osOpen = os.Open

// WorkRequest takes a workerID and will get another Job
type WorkRequest struct {
	WorkerID int
}

// Job will either be a backup request or system command, such as "Exit"
// This channel is unique to each worker, to ensure system commands are received by all
type Job struct {
	// Which WorkerCommand to execute
	Command int
	// File path in question
	Path string
	// Specify a device, if applicable
	// e.g. for backing up to a specific place, or moving from one device to another
	_ string
}

// WorkProgress reports command progress - primarily for backups and restorations
type WorkProgress struct {
	// Which path is in progress
	Path string
	// What is currently happening - e.g. checksum, copying, etc
	Status string
	// Size completed
	Completed int64
	// total size
	Total int64
}

// WorkComplete reports job completion
type WorkComplete struct {
	// The path that is completed
	Path string
	// Whether it was successful
	Succeeded bool
}

// WorkerContext contains needed operational data to process commands
type WorkerContext struct {
	DevCtx
	DB *sql.DB

	WorkerID int
	// If interrupted, stop after the command completes
	Interrupted *bool
	// On hard interrupt, rollback any changes made so far and exit
	HardInterrupted *bool

	Requests chan<- WorkRequest
	Jobs     <-chan Job
	Progress chan<- WorkProgress
	Complete chan<- WorkComplete
}

// WorkerCommandBackup backs up a provided path
var WorkerCommandBackup int = 1

// WorkerCommandRestore restores a provided path
var WorkerCommandRestore int = 2

// WorkerCommandRemove removes a provided path from the backup
var WorkerCommandRemove int = 3

// WorkerCommandVerify verifies a file on the local FS has valid checksum
var WorkerCommandVerify int = 4

// WorkerCommandConfirm confirms a file on the backup has valid checksum
var WorkerCommandConfirm int = 5

// WorkerCommandMove moves a file from one backup device to another
var WorkerCommandMove int = 6

// WorkerCommandUpdate checks a path for changes and backs up new files
var WorkerCommandUpdate int = 7

// StartWorker kicks off a new worker
func StartWorker(ctx *WorkerContext) {
	var processing bool = false

	go func() {
		listen(ctx, &processing)
	}()
}

var listen = func(ctx *WorkerContext, processing *bool) {

	for {
		select {
		case job := <-ctx.Jobs:
			runJob(job, ctx, processing)
		case <-time.After(100 * time.Millisecond):
			break
		}
	}
}

var runJob = func(job Job, ctx *WorkerContext, processing *bool) {
	defer func() {
		// On panic, make sure to roll back the job to avoid dangling files/references etc
		if r := recover(); r != nil {
			fmt.Println("Worker recovered. Error:\n", r)
			if *processing {
				rollBack(job, ctx)
				*processing = false
			}
		}

		ctx.Requests <- WorkRequest{ctx.WorkerID}
	}()

	*processing = true
	switch job.Command {
	// TODO: other commands
	case WorkerCommandBackup:
		break
	case WorkerCommandRestore:
		break
	case WorkerCommandUpdate:
		break
	case WorkerCommandVerify:
		err := verify(job, ctx)
		ctx.Complete <- WorkComplete{
			job.Path,
			err == nil,
		}
	case WorkerCommandConfirm:
		break
	case WorkerCommandMove:
		break
	case WorkerCommandRemove:
		break
	}

	*processing = false
}

var verify = func(job Job, ctx *WorkerContext) error {
	sum, err := checksum(job.Path, ctx)
	if err != nil {
		return err
	}

	// TODO: actual checksum here
	if sum != "" {
		return fmt.Errorf("Checksum mismatch")
	}

	return nil
}

var checksum = func(path string, ctx *WorkerContext) (string, error) {
	f, err := osOpen(path)
	if err != nil {
		return "", err
	}
	stats, err := f.Stat()
	if err != nil {
		return "", err
	}

	// TODO: buffer size (default 65 MB) should be a setting - custom thresholds, also?
	// seems to be a speed vs memory tradeoff, with diminishing returns
	// 65 MB on a 14 GB file was ~20% faster than the builtin sha256sum
	// For files under 500 MB, use 1/16th of their file size (at 500 MB is ~32 MB, so should still be very efficient)
	var chunkSize int
	if stats.Size() < (500 * 1024 * 1024 * 1024) {
		chunkSize = int(stats.Size() / 16)
		if chunkSize < 16 {
			chunkSize = 16
		}
	} else {
		chunkSize = 65536
	}
	reader := bufio.NewReaderSize(f, chunkSize)

	sum := sha256.New()

	var i int64
	for i = 0; i < stats.Size(); i += int64(chunkSize) {
		data, err := reader.Peek(chunkSize)
		ctx.Progress <- WorkProgress{path, "Calculating hash", i, stats.Size()}
		if err != nil && err != io.EOF {
			return "", err
		}

		// Since reader is peeking, won't advance the file pointer
		// Seeking isn't a thing, just discard an amount of data
		_, err = reader.Discard(chunkSize)
		if err != nil && err != io.EOF {
			return "", err
		}
		sum.Write(data)
	}

	return hex.EncodeToString(sum.Sum(make([]byte, 0))), nil
}

var rollBack = func(job Job, ctx *WorkerContext) {
	// For checksum verifications, no need to do anything since nothing will have been changed
	if (job.Command == WorkerCommandVerify) || (job.Command == WorkerCommandConfirm) {
		return
	}
}
