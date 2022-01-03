package main

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ammesonb/dispersed-backup/device"
	"github.com/stretchr/testify/assert"
)

// Checks expected functions are called
func TestRunManager(t *testing.T) {
	realGet := getDevices
	realProcess := process

	getCalled := false

	getDevices = func(_ *sql.DB) []*device.Device {
		getCalled = true
		return make([]*device.Device, 0)
	}

	// Can't use simple bool since this runs in separate goroutine
	process = func(_ *sync.Mutex, _ *bool, _ *[]*device.Device, _ *sql.DB, _ <-chan DeviceCommand, result chan<- DeviceResult) {
		result <- DeviceResult{true, "Called", nil}
	}

	results := make(chan DeviceResult, 1)
	defer func() {
		close(results)
		getDevices = realGet
		process = realProcess
	}()

	RunManager(&sql.DB{}, make(chan DeviceCommand, 1), results)
	assert.True(t, getCalled, "Get devies called")

	select {
	case msg := <-results:
		assert.Equal(t, "Called", msg.message, "Processing message received")
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "Processing message not received")
	}
}

// Check process calls handle, and restarts it with state on panic
func TestProcesRestartsAndPersists(t *testing.T) {
	realHandle := handle

	handleCalled := 0
	handle = func(_ *sync.Mutex, _ *bool, devices *[]*device.Device, _ *sql.DB, _ <-chan DeviceCommand, _ chan<- DeviceResult) {
		handleCalled++

		(*devices)[0].ReserveSpace(int64(10 * handleCalled))
		if handleCalled < 2 {
			panic("Call handle again")
		}
	}

	lock := sync.Mutex{}
	processing := false
	devices := make([]*device.Device, 0)
	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 10})
	commands := make(chan DeviceCommand, 1)
	results := make(chan DeviceResult, 1)

	defer func() {
		close(commands)
		close(results)
		handle = realHandle
	}()

	process(&lock, &processing, &devices, &sql.DB{}, commands, results)

	assert.Equal(t, 2, handleCalled, "Handle restarted once before clean exit")
	assert.Equal(t, uint64(60), devices[0].RemainingSpace(), "Reserved space persisted across panic")

	select {
	case <-results:
		assert.Fail(t, "Should not have received message on results")
	case <-time.After(100 * time.Millisecond):
		break

	}
}

// Check process unlocks the processing flag on panic
func TestProcessUnlocks(t *testing.T) {
	realHandle := handle

	handleCalled := 0
	handle = func(_ *sync.Mutex, _ *bool, devices *[]*device.Device, _ *sql.DB, _ <-chan DeviceCommand, _ chan<- DeviceResult) {
		handleCalled++

		(*devices)[0].ReserveSpace(int64(10 * handleCalled))
		if handleCalled < 2 {
			panic("Call handle again")
		}
	}

	lock := sync.Mutex{}
	lock.Lock()
	processing := true

	devices := make([]*device.Device, 0)
	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 10})
	commands := make(chan DeviceCommand, 1)
	results := make(chan DeviceResult, 1)

	defer func() {
		close(commands)
		close(results)
		handle = realHandle
	}()

	process(&lock, &processing, &devices, &sql.DB{}, commands, results)

	assert.False(t, processing, "Processing cleared")
}

// Check handle range locks
func TestHandleLocks(t *testing.T) {
	realAdd := addDevice

	addDevice = func(_ DeviceCommand, _ *sql.DB) (device.Device, error) {
		panic("nope")
	}

	lock := sync.Mutex{}
	processing := false
	devices := make([]*device.Device, 0)
	commands := make(chan DeviceCommand, 10)
	results := make(chan DeviceResult, 10)

	defer func() {
		r := recover()
		if r == nil {
			assert.Fail(t, "Panic should be recoverable")
		} else {
			assert.True(t, processing, "Processing should be set")
		}

		addDevice = realAdd
		close(results)
	}()

	commands <- DeviceCommand{command: DevCommandAddDevice, mountPoint: "anything"}
	close(commands)

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)

}

// Check device is added only for non-errors, and persisted outside scope
func TestDeviceAdding(t *testing.T) {
	realAdd := addDevice

	addCount := 0
	addDevice = func(_ DeviceCommand, _ *sql.DB) (device.Device, error) {
		addCount++

		if addCount == 1 {
			return device.Device{}, fmt.Errorf("Invalid")
		}

		return device.Device{DeviceID: 2}, nil
	}

	lock := sync.Mutex{}
	processing := false
	devices := make([]*device.Device, 0)
	commands := make(chan DeviceCommand, 10)
	results := make(chan DeviceResult, 10)

	defer func() {
		addDevice = realAdd
		close(results)
	}()

	commands <- DeviceCommand{command: DevCommandAddDevice}
	commands <- DeviceCommand{
		command:    DevCommandAddDevice,
		mountPoint: "/mnt/1",
	}
	commands <- DeviceCommand{
		command:    DevCommandAddDevice,
		mountPoint: "/mnt/2",
	}
	close(commands)

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)

	result := <-results
	assert.False(t, result.success, "Should fail device add without mount")
	assert.Errorf(t, result.err, "Mountpoint required", "Mount required error returned")

	result = <-results
	assert.False(t, result.success, "Should fail device add with invalid mount")
	assert.Errorf(t, result.err, "Invalid", "Device add failure returned")

	result = <-results
	assert.True(t, result.success, "Device adding succeeds")
	assert.Equal(t, "Device added successfully", result.message, "Added message correct")
	assert.Nil(t, result.err, "No errors if device added")
	assert.Len(t, devices, 1, "Device was added to array")
}

func TestReserving(t *testing.T) {
	realRes := reserveSpace

	count := 0
	reserveSpace = func(_ DeviceCommand, _ *[]*device.Device) (string, error) {
		count++

		if count == 1 {
			return "", fmt.Errorf("Invalid")
		}

		return "/mnt/1", nil
	}

	lock := sync.Mutex{}
	processing := false
	devices := make([]*device.Device, 0)
	commands := make(chan DeviceCommand, 10)
	results := make(chan DeviceResult, 10)

	defer func() {
		reserveSpace = realRes
		close(results)
	}()

	commands <- DeviceCommand{command: DevCommandReserveSpace}
	commands <- DeviceCommand{command: DevCommandReserveSpace}
	close(commands)

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)
	result := <-results
	assert.False(t, result.success, "Should fail reserve space")
	assert.Errorf(t, result.err, "Invalid", "Error message returned")
	assert.Equal(t, "", result.message, "No mount returned")

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)
	result = <-results
	assert.True(t, result.success, "Should succeed")
	assert.Nil(t, result.err, "No error returned")
	assert.Equal(t, "/mnt/1", result.message, "Mount path returned")
}

// Check device adding works as expected
func TestFreeing(t *testing.T) {
	realRes := freeSpace

	count := 0
	freeSpace = func(_ DeviceCommand, _ *[]*device.Device) error {
		count++

		if count == 1 {
			return fmt.Errorf("Invalid")
		}

		return nil
	}

	lock := sync.Mutex{}
	processing := false
	devices := make([]*device.Device, 0)
	commands := make(chan DeviceCommand, 10)
	results := make(chan DeviceResult, 10)

	defer func() {
		freeSpace = realRes
		close(results)
	}()

	commands <- DeviceCommand{command: DevCommandFreeSpace}
	commands <- DeviceCommand{command: DevCommandFreeSpace}
	close(commands)

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)
	result := <-results
	assert.False(t, result.success, "Should fail free space")
	assert.Errorf(t, result.err, "Invalid", "Error message returned")

	handle(&lock, &processing, &devices, &sql.DB{}, commands, results)
	result = <-results
	assert.True(t, result.success, "Should succeed")
	assert.Nil(t, result.err, "No error returned")
}

// Check device adding works as expected
func TestAddDevice(t *testing.T) {
	realMake := makeDevice
	realAdd := addDBDevice

	makeDevice = func(devID int, mountPoint string, serial string) (device.Device, error) {
		return device.Device{}, fmt.Errorf("No mount")
	}

	defer func() {
		makeDevice = realMake
		addDBDevice = realAdd
	}()

	_, err := addDevice(DeviceCommand{mountPoint: "", serial: ""}, &sql.DB{})
	assert.EqualErrorf(t, err, "No mount", "Device addition error correct")

	makeDevice = func(devID int, mountPoint string, serial string) (device.Device, error) {
		return device.Device{DeviceID: 1}, nil
	}
	addDBDevice = func(_ *sql.DB, _ device.Device) (device.Device, error) {
		return device.Device{}, fmt.Errorf("Already exists")
	}

	_, err = addDevice(DeviceCommand{mountPoint: "", serial: ""}, &sql.DB{})
	assert.EqualErrorf(t, err, "Already exists", "DB error correct")

	addDBDevice = func(_ *sql.DB, dev device.Device) (device.Device, error) {
		dev.DeviceID = 5
		return dev, nil
	}

	newDev, err := addDevice(DeviceCommand{mountPoint: "", serial: ""}, &sql.DB{})
	assert.Nil(t, err, "No error when adding")
	assert.Equal(t, 5, newDev.DeviceID, "Device ID set")
}

func TestReserveSpace(t *testing.T) {
	devices := make([]*device.Device, 0)
	mount, err := reserveSpace(DeviceCommand{}, &devices)
	assert.Equal(t, mount, "", "Mount is empty")
	assert.EqualErrorf(t, err, "No devices available -- add one first", "Expected no device available")

	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 2, MountPoint: "/mnt/2", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 3, MountPoint: "/mnt/3", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 50})
	mount, err = reserveSpace(DeviceCommand{mountPoint: "/mnt/1"}, &devices)
	assert.Equal(t, mount, "", "Mount is empty")
	assert.EqualErrorf(t, err, "Insufficient space on requested device", "Expected insufficient space")

	mount, err = reserveSpace(DeviceCommand{space: 200}, &devices)
	assert.Equal(t, mount, "", "Mount is empty")
	assert.EqualErrorf(t, err, "No device with sufficient space -- add another or make space", "Expected no device with space")

	mount, err = reserveSpace(DeviceCommand{mountPoint: "/mnt/3", space: 25}, &devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, mount, "/mnt/3", "Requested mount returned")
	assert.Equal(t, uint64(125), devices[2].RemainingSpace(), "25 bytes was reserved")

	mount, err = reserveSpace(DeviceCommand{space: 50}, &devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, mount, "/mnt/2", "Mount returned")
	assert.Equal(t, uint64(150), devices[1].AllocatedSpace, "50 bytes was reserved")

	mount, err = reserveSpace(DeviceCommand{space: 125}, &devices)
	assert.Equal(t, mount, "", "No mount returned")
	assert.Errorf(t, err, "No device with sufficient space -- add another or make space", "Cannot fully max out a drive")

	mount, err = reserveSpace(DeviceCommand{space: 75}, &devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, mount, "/mnt/3", "Third mount returned")
	assert.Equal(t, uint64(150), devices[2].AllocatedSpace, "75 bytes was reserved")

	assert.Equal(t, uint64(0), devices[0].RemainingSpace(), "No space remaining on device 1")
	assert.Equal(t, uint64(50), devices[1].RemainingSpace(), "50 remaining on device 2")
	assert.Equal(t, uint64(50), devices[2].RemainingSpace(), "25 remaining on device 3")
}

func TestFreeSpace(t *testing.T) {
	devices := make([]*device.Device, 0)
	err := freeSpace(DeviceCommand{}, &devices)
	assert.EqualErrorf(t, err, "Mountpoint required", "Requires mountpoint error returned")

	err = freeSpace(DeviceCommand{mountPoint: "/mnt2"}, &devices)
	assert.EqualErrorf(t, err, "No such mountpoint", "Mountpoint not found error returned")

	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 2, MountPoint: "/mnt/2", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 3, MountPoint: "/mnt/3", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 50})

	err = freeSpace(DeviceCommand{mountPoint: "/mnt/2", space: 50}, &devices)
	assert.Nil(t, err, "No error decreasing space")

	assert.Equal(t, uint64(0), devices[0].RemainingSpace(), "No space remaining on device 1")
	assert.Equal(t, uint64(50), devices[1].RemainingSpace(), "50 remaining on device 2")
	assert.Equal(t, uint64(150), devices[2].RemainingSpace(), "150 remaining on device 3")
}
