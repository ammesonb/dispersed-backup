CREATE TABLE devices (
  deviceID INTEGER PRIMARY KEY AUTOINCREMENT,
  mountPoint TEXT NOT NULL UNIQUE,
  serialNumber TEXT NOT NULL UNIQUE
);