package fs

import (
	"os"

	"github.com/kubism-io/backup-operator/pkg/backup"
)

func NewFileSource(filepath string) (backup.Source, error) {
	return &fileSource{
		filepath: filepath,
	}, nil
}

type fileSource struct {
	filepath string
}

func (f *fileSource) Backup(dst backup.Destination) error {
	file, err := os.Open(f.filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return dst.Store(file)
}
