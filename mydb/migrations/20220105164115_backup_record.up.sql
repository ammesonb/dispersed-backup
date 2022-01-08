CREATE TABLE record (
        recordID INTEGER PRIMARY KEY AUTOINCREMENT,
        fsPath TEXT NOT NULL UNIQUE,
        backupPath TEXT NOT NULL UNIQUE,
        isFile BOOLEAN NOT NULL,
        checksum TEXT NOT NULL UNIQUE,
        algorithm TEXT NOT NULL,
        permissions TEXT,
        pathOwner TEXT,
        pathGroup TEXT
)
