package record

// Record contains metadata for a backed-up file or folder
type Record struct {
	RecordID int
	// The path on the local file system
	FilePath string
	// The path to the backed-up record (only applies to files)
	BackupPath string
	// Whether this is a file or folder
	IsFile bool
	// For folders, this is just the path + timestamp
	Checksum string
	// N/A for folders, otherwise md5sum, sha256sum, etc
	Algorithm string
	// ex 0755
	Permissions string
	// The text name of the owner and group, NOT UID
	// since UIDs may not match between installations
	PathOwner string
	PathGroup string
}
