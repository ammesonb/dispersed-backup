package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChecksumFailsLoad(t *testing.T) {
	realOpen := osOpen

	osOpen = func(path string) (*os.File, error) {
		return nil, fmt.Errorf("No such file")
	}

	sum, err := checksum("/whatever", &WorkerContext{})
	assert.EqualErrorf(t, err, "No such file", "OS open fails with message")
	assert.Empty(t, sum, "No checksum returned")

	osOpen = realOpen

	f, err := os.CreateTemp("", "checksum_test")

	defer os.Remove(f.Name())

	assert.Nil(t, err, "No error creating temporary file")
	_, err = f.Write([]byte("abcdefghijklmnopqrstuvwxyz0123456789\n"))
	assert.Nil(t, err, "No error writing data to file")

	f.Close()

	progress := make(chan WorkProgress, 10)
	ctx := WorkerContext{
		Progress: progress,
	}
	sum, err = checksum(f.Name(), &ctx)
	assert.Nil(t, err, "No error checksumming file")
	assert.Equal(t, "c74579aeba50c05bc0cd36bce93919343ebfc1ddf74ae96417e7aba274351c5e", sum, "Correct checksum returned")

	msg := <-progress
	assert.Equal(t, "Calculating hash", msg.Status, "Checksum status correct")
	close(progress)
}
