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

package mongodb

import (
	"fmt"

	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/stream"
	"github.com/mongodb/mongo-tools/mongorestore"
)

func NewMongoDBDestination(uri string) (stream.Destination, error) {
	return &mongoDBDestination{
		URI: uri,
		log: logger.WithName("mongodst"),
	}, nil
}

type mongoDBDestination struct {
	URI     string
	restore *mongorestore.MongoRestore
	log     logger.Logger
}

func (m *mongoDBDestination) Store(obj stream.Object) error {
	log := m.log
	args := []string{
		fmt.Sprintf("--uri=\"%s\"", m.URI),
		"--archive",
		"--gzip",
	}
	opts, err := mongorestore.ParseOptions(args, "custom", "custom")
	if err != nil {
		return err
	}
	m.restore, err = mongorestore.New(opts)
	if err != nil {
		return err
	}
	defer m.restore.Close()
	m.restore.InputReader = obj.Data
	// start the restoral
	result := m.restore.Restore()
	if result.Err != nil {
		return result.Err
	}
	if m.restore.ToolOptions.WriteConcern.Acknowledged() {
		log.Info(fmt.Sprintf("%v document(s) restored successfully. %v document(s) failed to restore.", result.Successes, result.Failures))
	} else {
		log.Info("done")
	}
	return nil
}

func (m *mongoDBDestination) Close() error {
	if m.restore != nil {
		m.restore.HandleInterrupt()
	}
	return nil
}
