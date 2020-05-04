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
	"bytes"

	"github.com/kubism/backup-operator/pkg/backup"
)

func NewBufferSource(name string, data []byte) (*BufferSource, error) {
	return &BufferSource{
		Name: name,
		Data: data,
	}, nil
}

type BufferSource struct {
	Name string
	Data []byte
}

func (b *BufferSource) Stream(dst backup.Destination) error {
	return dst.Store(backup.Object{
		ID:   b.Name,
		Data: bytes.NewReader(b.Data),
	})
}
