package mydb

import (
	"os"
	"testing"

	"github.com/ammesonb/dispersed-backup/device"
	"github.com/stretchr/testify/assert"
)

func TestAddAndGetDevice(t *testing.T) {
	realMake := makeDevice
	makeDevice = func(devID int, mountPoint string, serial string) (device.Device, error) {
		return device.Device{
			DeviceID:     devID,
			MountPoint:   mountPoint,
			DeviceSerial: serial,
		}, nil
	}

	f, err := os.CreateTemp("", "test.db")
	if err != nil {
		panic(err)
	}

	defer func() {
		makeDevice = realMake
		DeleteDB(f.Name())
	}()

	db := OpenDB(f.Name())

	newDev, err := makeDevice(0, "/mnt/foo", "abc123")
	if err != nil {
		panic(err)
	}

	dev, err := AddDevice(db, newDev)
	if err != nil {
		panic(err)
	}

	assert.Greater(t, dev.DeviceID, newDev.DeviceID, "Device ID is set")
	assert.Equal(t, dev.MountPoint, newDev.MountPoint, "Mount point persisted")
	assert.Equal(t, dev.DeviceSerial, newDev.DeviceSerial, "Device serial persisted")

	devices := GetDevices(db)

	assert.Len(t, devices, 1, "One device returned")
	assert.Equal(t, devices[0].DeviceID, dev.DeviceID, "Correct device ID returned")
	assert.Equal(t, devices[0].MountPoint, dev.MountPoint, "Correct device mount returned")
	assert.Equal(t, devices[0].DeviceSerial, dev.DeviceSerial, "Correct device serial returned")
}
