package device

import (
	"fmt"

	"github.com/shirou/gopsutil/disk"
)

// Device represents a mount point on the system that can be used for backing up files
type Device struct {
	MountPoint     string
	DeviceSerial   string
	AvailableSpace uint64
	AllocatedSpace uint64
}

// Reserves the requested space on the device, returning a new device
func reserveSpace(dev Device, needed uint64) Device { //nolint
	return Device{
		dev.MountPoint,
		dev.DeviceSerial,
		dev.AvailableSpace,
		dev.AllocatedSpace + needed,
	}
}

// Creates a device based on the path, failing if it is not a mountpoint
func getDevice(path string) (Device, error) { //nolint
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

			return Device{
				path,
				disk.GetDiskSerialNumber(part.Device),
				usage.Free,
				0,
			}, nil
		}
	}

	return Device{}, fmt.Errorf("No device mounted on %s", path)
}
