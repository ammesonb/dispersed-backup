package mydb

import (
	"database/sql"
	"fmt"

	"github.com/ammesonb/dispersed-backup/device"
)

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

		if !rows.NextResultSet() {
			panic(fmt.Errorf("Expected more devices from DB"))
		}

		newDev, err := device.MakeDevice(
			deviceID,
			mountPoint,
			serialNumber,
		)
		if err != nil {
			panic(err)
		}

		devs = append(devs, newDev)

	}

	return devs
}
