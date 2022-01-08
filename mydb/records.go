package mydb

import (
	"database/sql"

	"github.com/ammesonb/dispersed-backup/record"
)

// GetRecords returns all records in the database
func GetRecords(db *sql.DB) []*record.Record {
	rows, err := db.Query("SELECT * FROM records")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var records []*record.Record

	for rows.Next() {
		var (
			recordID    int
			filePath    string
			backupPath  string
			isFile      bool
			checksum    string
			algorithm   string
			permissions string
			pathOwner   string
			pathGroup   string
		)
		err := rows.Scan(&recordID, &filePath, &backupPath, &isFile, &checksum, &algorithm, &permissions, &pathOwner, &pathGroup)
		if err != nil {
			panic(err)
		}

		rec := record.Record{
			RecordID:    recordID,
			FilePath:    filePath,
			BackupPath:  backupPath,
			IsFile:      isFile,
			Checksum:    checksum,
			Algorithm:   algorithm,
			Permissions: permissions,
			PathOwner:   pathOwner,
			PathGroup:   pathGroup,
		}

		records = append(records, &rec)

		if !rows.NextResultSet() {
			break
		}

	}

	return records
}

// AddDevice adds a new device to the given database instance
/*func AddDevice(db *sql.DB, newDevice device.Device) (device.Device, error) {
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
}*/
