package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/ammesonb/dispersed-backup/device"
	"github.com/stretchr/testify/assert"
)

// Check device adding works as expected
func TestAddDevice(t *testing.T) {
	realMake := makeDevice

	makeDevice = func(devID int, mountPoint string, serial string) (device.Device, error) {
		return device.Device{}, fmt.Errorf("No mount")
	}

	defer func() {
		makeDevice = realMake
	}()

	results := make(chan DeviceResult, 1)

	addDevice(DeviceCommand{mountPoint: "", serial: ""}, &sql.DB{}, results)

	select {
	case res := <-results:
		assert.False(t, res.success, "Device addition should fail")
		assert.EqualErrorf(t, res.err, "No mount", "Device addition error correct")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Did not receive result")
	}

	close(results)
}

func TestReserveSpace(t *testing.T) {
	devices := make([]*device.Device, 0)
	err := reserveSpace(DeviceCommand{}, devices)
	assert.EqualErrorf(t, err, "No devices available -- add one first", "Expected no device available")

	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 2, MountPoint: "/mnt/2", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 3, MountPoint: "/mnt/3", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 50})
	err = reserveSpace(DeviceCommand{mountPoint: "/mnt/1"}, devices)
	assert.EqualErrorf(t, err, "Insufficient space on requested device", "Expected insufficient space")

	err = reserveSpace(DeviceCommand{space: 200}, devices)
	assert.EqualErrorf(t, err, "No device with sufficient space -- add another or make space", "Expected no device with space")

	err = reserveSpace(DeviceCommand{mountPoint: "/mnt/3", space: 25}, devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, uint64(125), devices[2].RemainingSpace(), "25 bytes was reserved")

	err = reserveSpace(DeviceCommand{space: 50}, devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, uint64(150), devices[1].AllocatedSpace, "50 bytes was reserved")

	err = reserveSpace(DeviceCommand{space: 125}, devices)
	assert.Errorf(t, err, "No device with sufficient space -- add another or make space", "Cannot fully max out a drive")

	err = reserveSpace(DeviceCommand{space: 75}, devices)
	assert.Nil(t, err, "No error requesting allocatable space")
	assert.Equal(t, uint64(150), devices[2].AllocatedSpace, "75 bytes was reserved")

	assert.Equal(t, uint64(0), devices[0].RemainingSpace(), "No space remaining on device 1")
	assert.Equal(t, uint64(50), devices[1].RemainingSpace(), "50 remaining on device 2")
	assert.Equal(t, uint64(50), devices[2].RemainingSpace(), "25 remaining on device 3")
}

func TestFreeSpace(t *testing.T) {
	devices := make([]*device.Device, 0)
	err := freeSpace(DeviceCommand{}, devices)
	assert.EqualErrorf(t, err, "Mountpoint required", "Requires mountpoint error returned")

	err = freeSpace(DeviceCommand{mountPoint: "/mnt2"}, devices)
	assert.EqualErrorf(t, err, "No such mountpoint", "Mountpoint not found error returned")

	devices = append(devices, &device.Device{DeviceID: 1, MountPoint: "/mnt/1", DeviceSerial: "ABC123", AvailableSpace: 100, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 2, MountPoint: "/mnt/2", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 100})
	devices = append(devices, &device.Device{DeviceID: 3, MountPoint: "/mnt/3", DeviceSerial: "ABC223", AvailableSpace: 200, AllocatedSpace: 50})

	err = freeSpace(DeviceCommand{mountPoint: "/mnt/2", space: 50}, devices)
	assert.Nil(t, err, "No error decreasing space")

	assert.Equal(t, uint64(0), devices[0].RemainingSpace(), "No space remaining on device 1")
	assert.Equal(t, uint64(50), devices[1].RemainingSpace(), "50 remaining on device 2")
	assert.Equal(t, uint64(150), devices[2].RemainingSpace(), "150 remaining on device 3")
}
