package mydb

import (
	"database/sql"

	"github.com/ammesonb/dispersed-backup/device"
)

var makeDevice = device.MakeDevice

// GetDevices returns the cached devices from the provided database connection
func GetDevices(db *sql.DB) []device.Device {
	rows, err := db.Query("SELECT * FROM devices")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var devs []device.Device

	for rows.Next() {
		var (
			deviceID     int
			mountPoint   string
			serialNumber string
		)
		err := rows.Scan(&deviceID, &mountPoint, &serialNumber)
		if err != nil {
			panic(err)
		}

		newDev, err := makeDevice(
			deviceID,
			mountPoint,
			serialNumber,
		)
		if err != nil {
			panic(err)
		}

		devs = append(devs, newDev)

		if !rows.NextResultSet() {
			break
		}

	}

	return devs
}

// AddDevice adds a new device to the given database instance
func AddDevice(db *sql.DB, newDevice device.Device) (device.Device, error) {
	var id int
	err := db.QueryRow(`
    INSERT INTO devices (
      mountPoint,
      serialNumber
    )
    VALUES (
      $1,
      $2
    )
    RETURNING deviceID
  `, newDevice.MountPoint, newDevice.DeviceSerial).Scan(&id)
	if err != nil {
		return device.Device{}, err
	}

	return makeDevice(id, newDevice.MountPoint, newDevice.DeviceSerial)
}
