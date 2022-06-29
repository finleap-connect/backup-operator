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
	"path/filepath"

	"github.com/finleap-connect/backup-operator/pkg/backup"
)

func NewDirDestination(dir string) (backup.Destination, error) {
	return &dirDestination{
		dir: dir,
	}, nil
}

type dirDestination struct {
	dir string
}

func (f *dirDestination) Store(obj backup.Object) (int64, error) {
	fp := filepath.Join(f.dir, obj.ID)
	file, err := os.Create(fp)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	return io.Copy(file, obj.Data)
}
