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

package mem

import (
	"io/ioutil"

	"github.com/finleap-connect/backup-operator/pkg/backup"
)

func NewBufferDestination() (*BufferDestination, error) {
	return &BufferDestination{
		Data: map[string][]byte{},
	}, nil
}

type BufferDestination struct {
	Data map[string][]byte
}

func (b *BufferDestination) Store(obj backup.Object) (int64, error) {
	var err error
	b.Data[obj.ID], err = ioutil.ReadAll(obj.Data)
	return (int64)(len(b.Data[obj.ID])), err
}
