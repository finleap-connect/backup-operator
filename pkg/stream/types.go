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

package stream

import (
	"io"
)

type Object struct {
	ID   string // Only used by *Many use-cases for determine filenames
	Data io.Reader
}

type Destination interface {
	Store(obj Object) error
}

type Source interface {
	Stream(dst Destination) error
}

// NOTE: *Many interface not yet used, but use case for them include backups
//       of directories in S3 etc.

type DestinationMany interface {
	Store(data chan Object) error
}

type SourceMany interface {
	Stream(dst DestinationMany) error
}
