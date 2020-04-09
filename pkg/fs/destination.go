package fs

import (
	"io"
	"os"

	"github.com/kubism-io/backup-operator/pkg/backup"
)

func NewFileDestination(filepath string) (backup.Destination, error) {
	return &fileDestination{
		filepath: filepath,
	}, nil
}

type fileDestination struct {
	filepath string
}

func (f *fileDestination) Store(data io.Reader) error {
	file, err := os.Create(f.filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	return err
}
