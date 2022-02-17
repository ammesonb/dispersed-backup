package main

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/ammesonb/dispersed-backup/device"
	"github.com/ammesonb/dispersed-backup/mydb"
)

var getDevices = mydb.GetDevices
var makeDevice = device.MakeDevice
var addDBDevice = mydb.AddDevice

// DevCommandAddDevice instructs the manager to add a new device by mount
const DevCommandAddDevice int = 1

// DevCommandReserveSpace instructs the manager to reserve an amount of space, optionally on a specific mount
const DevCommandReserveSpace int = 2

// DevCommandFreeSpace instructs the manager to free an amount of space on a given mount
const DevCommandFreeSpace int = 3

// DeviceCommand contains information needed to execute a command
type DeviceCommand struct {
	// Command integer, see variables above
	command int
	// Mountpoint and serial, for adding a new device
	// Mountpoint may also be used to request storing a file on a specific mountpoint
	mountPoint string
	serial     string
	// Space to allocate or free
	space int64
}

// DeviceResult contains details about the executed action
type DeviceResult struct {
	success bool
	message string
	err     error
}

// DevMan contains the necessary components for interacting with the device manager goroutine
type DevMan struct {
	commands chan DeviceCommand
	results  chan DeviceResult
	lock     sync.Mutex
}

// RunManager should be used in a goroutine, and is responsible for managing available device space for file backups
// A MutEx should be used to maintain one-to-one command -> result behavior
func RunManager(db *sql.DB, commands <-chan DeviceCommand, results chan<- DeviceResult) {
	devices := getDevices(db)

	go func() {
		process(&devices, db, commands, results)
	}()
}

func process(devices *[]*device.Device, db *sql.DB, commands <-chan DeviceCommand, results chan<- DeviceResult) {
	for command := range commands {
		handle(command, devices, db, commands, results)
	}
}

var handle = func(command DeviceCommand, devices *[]*device.Device, db *sql.DB, commands <-chan DeviceCommand, results chan<- DeviceResult) {
	defer func() {
		if r := recover(); r != nil {
			// Ignore errors, since need to keep processing requests
			fmt.Println("Recovered. Error:\n", r)
			// Since only called when command received, ensure we inform the caller there was an error
			results <- DeviceResult{false, "", fmt.Errorf("Panic during execution")}
		}
	}()

	switch command.command {
	case DevCommandAddDevice:
		if len(command.mountPoint) == 0 {
			results <- DeviceResult{false, "", fmt.Errorf("Mountpoint required")}
			break
		}

		device, err := addDevice(command, db)
		if err == nil {
			*devices = append(*devices, &device)
			results <- DeviceResult{true, "Device added successfully", nil}
		} else {
			results <- DeviceResult{false, "", err}
		}
	case DevCommandReserveSpace:
		mount, err := reserveSpace(command, devices)
		if err != nil {
			results <- DeviceResult{false, "", err}
		} else {
			results <- DeviceResult{true, mount, nil}
		}
	case DevCommandFreeSpace:
		err := freeSpace(command, devices)
		if err != nil {
			results <- DeviceResult{false, "", err}
		} else {
			results <- DeviceResult{true, "Space freed", nil}
		}

	default:
		results <- DeviceResult{false, "", fmt.Errorf("%d at path %s is not a recognized command", command.command, command.mountPoint)}
	}
}

// addDevice wraps functionality to add a new device, returning the result
var addDevice = func(command DeviceCommand, db *sql.DB) (device.Device, error) {
	toAdd, err := makeDevice(0, command.mountPoint, command.serial)
	if err != nil {
		return device.Device{}, err
	}

	addedDev, err := addDBDevice(db, toAdd)
	if err != nil {
		return device.Device{}, err
	}

	return addedDev, nil
}

// reserveSpace attempts to allocate space on
var reserveSpace = func(command DeviceCommand, devices *[]*device.Device) (string, error) {
	if len(*devices) == 0 {
		return "", fmt.Errorf("No devices available -- add one first")
	}

	for _, dev := range *devices {
		// If requested size is negative, then would be less than an int64 representation of remaining space anyways
		if len(command.mountPoint) > 0 && command.mountPoint == dev.MountPoint {
			if dev.RemainingSpace() > uint64(command.space) {
				dev.ReserveSpace(command.space)
				return dev.MountPoint, nil
			}

			return "", fmt.Errorf("Insufficient space on requested device")
			// Check device space
		} else if len(command.mountPoint) == 0 && dev.RemainingSpace() > uint64(command.space) {
			dev.ReserveSpace(command.space)
			return dev.MountPoint, nil
		}
	}

	return "", fmt.Errorf("No device with sufficient space -- add another or make space")
}

var freeSpace = func(command DeviceCommand, devices *[]*device.Device) error {
	if len(command.mountPoint) == 0 {
		return fmt.Errorf("Mountpoint required")
	}

	var selected *device.Device = &device.Device{}

	for _, dev := range *devices {
		if dev.MountPoint == command.mountPoint {
			selected = dev
			break
		}
	}
	if selected.DeviceID == 0 {
		return fmt.Errorf("No such mountpoint")
	}

	selected.ReserveSpace(-1 * command.space)
	return nil
}
