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

package consul

import (
	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/stream"
)

type consulSource struct {
	URI      string
	Username string
	Password string
	log      logger.Logger
}

func NewConsulSource(uri, username, password string) (stream.Source, error) {
	return &consulSource{
		URI:      uri,
		Username: username,
		Password: password,
		log:      logger.WithName("consulsrc"),
	}, nil
}

func (s *consulSource) Stream(dst stream.Destination) error {
	panic("not implemented")
}
