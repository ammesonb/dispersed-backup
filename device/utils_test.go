package device

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
)

func TestReserveSpace(t *testing.T) {
	var available, allocated, needed uint64 = 100, 5, 50

	dev := Device{
		123,
		"/mount",
		"123abc",
		available,
		allocated,
	}
	dev = reserveSpace(dev, needed)
	assert.Equal(t, 123, dev.DeviceID, "DeviceID persisted")
	assert.Equal(t, "/mount", dev.MountPoint, "MountPoint persisted")
	assert.Equal(t, "123abc", dev.DeviceSerial, "Serial persisted")
	assert.Equal(t, available, dev.AvailableSpace, "AvailableSpace persisted")
	assert.Equal(t, allocated+needed, dev.AllocatedSpace, "AllocatedSpace incremented")
}

func makeTestParts(resultCount int, err string) func(bool) ([]disk.PartitionStat, error) {
	return func(all bool) ([]disk.PartitionStat, error) {
		var parts []disk.PartitionStat = make([]disk.PartitionStat, resultCount)

		for i := 0; i < resultCount; i++ {
			parts[i] = disk.PartitionStat{
				Mountpoint: fmt.Sprintf("/mnt/%d", i),
				Fstype:     "ext4",
				Device:     fmt.Sprintf("/dev/sda%d", i),
				Opts:       "rw",
			}
		}

		if err != "" {
			return parts, fmt.Errorf(err)
		}

		return parts, nil
	}
}

func makeTestUsage(err string) func(string) (*disk.UsageStat, error) {
	return func(path string) (*disk.UsageStat, error) {
		if err != "" {
			return &disk.UsageStat{}, fmt.Errorf(err)
		}

		return &disk.UsageStat{
			Free:              uint64(123),
			Fstype:            "ext4",
			InodesFree:        uint64(1000),
			InodesTotal:       uint64(2000),
			InodesUsed:        uint64(1000),
			InodesUsedPercent: float64(50),
			Path:              path,
			Total:             uint64(246),
			Used:              uint64(123),
			UsedPercent:       float64(50),
		}, nil
	}
}

// Check not getting partitions returns error
func TestMakeDeviceNoPartitions(t *testing.T) {
	realParts := getParts
	getParts = makeTestParts(0, "access denied")
	defer func() { getParts = realParts }()

	dev, err := MakeDevice(123, "/mount", "serial")
	assert.EqualErrorf(t, err, "Failed to get partitions: access denied", "Access denied error thrown")
	assert.Equal(t, Device{}, dev, "Empty device returned on partition failure")
}

// Check failing to get disk usage returns error
func TestMakeDeviceNoUsage(t *testing.T) {
	realParts := getParts
	realUsage := getUsage

	getParts = makeTestParts(1, "")
	getUsage = makeTestUsage("access denied")
	defer func() {
		getParts = realParts
		getUsage = realUsage
	}()

	dev, err := MakeDevice(123, "/mnt/0", "serial")
	assert.EqualErrorf(t, err, "Failed to get disk usage: access denied", "Access denied error thrown")
	assert.Equal(t, Device{}, dev, "Empty device returned on usage failure")
}

// Ensure getting serial number works
func TestMakeDeviceGetsSerial(t *testing.T) {
	realParts := getParts
	realUsage := getUsage
	realSerial := getSerial

	getParts = makeTestParts(2, "")
	getUsage = makeTestUsage("")
	getSerial = func(path string) string {
		return "a-very-real-serial"
	}
	defer func() {
		getParts = realParts
		getUsage = realUsage
		getSerial = realSerial
	}()

	dev, err := MakeDevice(123, "/mnt/1", "")
	assert.Nil(t, err, "No error thrown when serial auto-detected")
	assert.Equal(t, "a-very-real-serial", dev.DeviceSerial, "Serial provided")
}

// Pass in serial number gets passed through
func TestMakeDeviceWithSerial(t *testing.T) {
	realParts := getParts
	realUsage := getUsage
	realSerial := getSerial

	getParts = makeTestParts(2, "")
	getUsage = makeTestUsage("")
	called := false
	getSerial = func(path string) string {
		called = true
		return "a-very-real-serial"
	}
	defer func() {
		getParts = realParts
		getUsage = realUsage
		getSerial = realSerial
	}()

	dev, err := MakeDevice(123, "/mnt/1", "cached-serial")
	assert.Nil(t, err, "No error thrown when serial given")
	assert.Equal(t, "cached-serial", dev.DeviceSerial, "Serial input overrides function")
	assert.False(t, called, "getSerial should not be called if one is provided")
}

// Check no partition found
func TestMakeDevicePartitionNotFound(t *testing.T) {
	realParts := getParts
	realUsage := getUsage

	getParts = makeTestParts(2, "")
	getUsage = makeTestUsage("")
	defer func() {
		getParts = realParts
		getUsage = realUsage
	}()

	dev, err := MakeDevice(123, "/mnt/2", "serial")
	assert.EqualErrorf(t, err, "No device mounted on /mnt/2", "No mounted device error returned")
	assert.Equal(t, Device{}, dev, "Device empty if mountpoint not matched")
}
