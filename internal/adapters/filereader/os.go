package filereader

import (
	"os"

	zenitherrors "github.com/maneeshaindrachapa/zenith/internal/errors"
)

// OS reads files from the local filesystem.
type OS struct{}

// ReadFile reads path from disk.
func (OS) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &zenitherrors.ConfigError{Kind: zenitherrors.ErrReadFile, Operation: "read file", Path: path, CauseErr: err}
	}
	return data, nil
}
