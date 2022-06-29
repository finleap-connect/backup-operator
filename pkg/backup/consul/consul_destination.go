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
	"github.com/finleap-connect/backup-operator/pkg/backup"
	"github.com/finleap-connect/backup-operator/pkg/logger"

	consulApi "github.com/hashicorp/consul/api"
)

type consulDestination struct {
	Client *consulApi.Client
	log    logger.Logger
}

func NewConsulDestination(uri, username, password string) (backup.Destination, error) {
	consulConf := consulApi.DefaultConfig()
	consulConf.Address = uri
	if username != "" && password != "" {
		consulConf.HttpAuth = &consulApi.HttpBasicAuth{
			Username: username,
			Password: password,
		}
	}
	client, err := consulApi.NewClient(consulConf)
	if err != nil {
		return nil, err
	}

	return &consulDestination{
		Client: client,
		log:    logger.WithName("consuldst"),
	}, nil
}

func (s *consulDestination) Store(obj backup.Object) (int64, error) {
	log := s.log

	log.Info("restore starting")
	err := s.Client.Snapshot().Restore(&consulApi.WriteOptions{}, obj.Data)
	if err != nil {
		log.Error(err, "Failed to write snapshot to consul")
		return 0, err
	}
	log.Info("restore finished")
	return 0, nil // NOTE: written bytes not supported
}
