/*
Copyright 2020 Backup Operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fs

import (
	"io"
	"os"

	"github.com/kubism-io/backup-operator/pkg/stream"
)

func NewFileDestination(filepath string) (stream.Destination, error) {
	return &fileDestination{
		filepath: filepath,
	}, nil
}

type fileDestination struct {
	filepath string
}

func (f *fileDestination) Store(obj stream.Object) error {
	file, err := os.Create(f.filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, obj.Data)
	return err
}
