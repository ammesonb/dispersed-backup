package device

import (
	"fmt"

	"github.com/shirou/gopsutil/disk"
)

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
	parts, err := disk.Partitions(false)
	if err != nil {
		panic(err)
	}

	for _, part := range parts {
		if part.Mountpoint == path {
			usage, err := disk.Usage(path)
			if err != nil {
				panic(err)
			}

			if serial == "" {
				serial = disk.GetDiskSerialNumber(part.Device)
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
