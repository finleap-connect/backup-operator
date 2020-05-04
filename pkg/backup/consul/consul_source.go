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
	"fmt"
	"io"
	"time"

	"github.com/kubism/backup-operator/pkg/backup"
	"github.com/kubism/backup-operator/pkg/logger"

	consulApi "github.com/hashicorp/consul/api"
)

type consulSource struct {
	SnapName string
	Client   *consulApi.Client
	log      logger.Logger
}

func NewConsulSource(uri, username, password, snapName string) (backup.Source, error) {
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

	return &consulSource{
		SnapName: snapName,
		Client:   client,
		log:      logger.WithName("consulsrc"),
	}, nil
}

func (s *consulSource) Stream(dst backup.Destination) (int64, error) {
	log := s.log

	reader, _, err := s.Client.Snapshot().Save(&consulApi.QueryOptions{})
	if err != nil {
		log.Error(err, "Could not get snapshot from consul")
		return 0, err
	}
	defer reader.Close()
	pr, pw := io.Pipe()

	// start the backup in a separate routine
	errc := make(chan error, 1)
	defer close(errc)
	go func() {
		defer pw.Close()
		log.Info("starting dump")
		numBytes, err := io.Copy(pw, reader)
		if err != nil {
			errc <- err
		}
		log.Info("finished dump", "numBytes", numBytes)
	}()
	written, dsterr := dst.Store(backup.Object{
		ID:   s.SnapName,
		Data: pr,
	})

	select {
	case srcerr := <-errc: // return src error if possible as well
		return written, fmt.Errorf("dst error: %v; src error: %v", dsterr, srcerr)
	case <-time.After(1 * time.Second):
		return written, dsterr
	}
}
