package device

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
)

var getParts = disk.Partitions
var getUsage = disk.Usage
var getSerial = disk.GetDiskSerialNumber

// Device represents a mount point on the system that can be used for backing up files
type Device struct {
	DeviceID       int
	MountPoint     string
	DeviceSerial   string
	AvailableSpace uint64
	AllocatedSpace uint64
}

// Reserves the requested space on the device, returning a new device
func reserveSpace(dev Device, needed uint64) Device { //nolint
	return Device{
		dev.DeviceID,
		dev.MountPoint,
		dev.DeviceSerial,
		dev.AvailableSpace,
		dev.AllocatedSpace + needed,
	}
}

// MakeDevice creates a device based on the provided path and optional serial
func MakeDevice(devID int, path string, serial string) (Device, error) {
	parts, err := getParts(false)
	if err != nil {
		return Device{}, fmt.Errorf("Failed to get partitions: %v", err)
	}

	for _, part := range parts {
		if part.Mountpoint == path {
			usage, err := getUsage(path)
			if err != nil {
				return Device{}, fmt.Errorf("Failed to get disk usage: %v", err)
			}

			if serial == "" {
				serial = getSerial(part.Device)
			}

			return Device{
				devID,
				path,
				serial,
				usage.Free,
				0,
			}, nil
		}
	}

	return Device{}, fmt.Errorf("No device mounted on %s", path)
}
