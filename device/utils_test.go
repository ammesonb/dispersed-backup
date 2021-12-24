package device

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
