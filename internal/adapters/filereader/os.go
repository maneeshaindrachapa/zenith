package filereader

import "os"

// OS reads files from the local filesystem.
type OS struct{}

// ReadFile reads path from disk.
func (OS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
