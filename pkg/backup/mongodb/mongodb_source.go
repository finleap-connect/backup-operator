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
	"io"
	"regexp"
	"time"

	"github.com/mongodb/mongo-tools/common/options"

	"github.com/finleap-connect/backup-operator/pkg/backup"
	"github.com/finleap-connect/backup-operator/pkg/logger"
	"github.com/mongodb/mongo-tools/mongodump"
)

var (
	filter = regexp.MustCompile("[^a-zA-Z0-9]+")
)

func NewMongoDBSource(uri, database, archiveName string) (backup.Source, error) {
	return &mongoDBSource{
		URI:         uri,
		Database:    database,
		ArchiveName: archiveName,
		log:         logger.WithName("mongosrc"),
	}, nil
}

type mongoDBSource struct {
	URI         string
	Database    string // TODO: implement
	ArchiveName string
	dump        *mongodump.MongoDump
	log         logger.Logger
}

func (m *mongoDBSource) Stream(dst backup.Destination) (int64, error) {
	log := m.log
	opts := options.New("mongodump",
		"custom",
		"custom",
		mongodump.Usage,
		true,
		options.EnabledOptions{Auth: true, Connection: true, Namespace: true, URI: true},
	)
	inputOpts := &mongodump.InputOptions{}
	opts.AddOptions(inputOpts)
	outputOpts := &mongodump.OutputOptions{}
	opts.AddOptions(outputOpts)
	args := []string{
		fmt.Sprintf("--uri=\"%s\"", m.URI),
		"--archive",
		"--gzip",
	}
	_, err := opts.ParseArgs(args)
	if err != nil {
		return 0, err
	}
	// verify uri options and log them
	opts.URI.LogUnsupportedOptions()
	// setup dump and make sure output is piped
	m.dump = &mongodump.MongoDump{
		ToolOptions:   opts,
		OutputOptions: outputOpts,
		InputOptions:  inputOpts,
	}
	if err = m.dump.Init(); err != nil {
		return 0, err
	}
	pr, pw := io.Pipe()
	m.dump.OutputWriter = pw
	// start the backup in a separate routine
	errc := make(chan error, 1)
	defer close(errc)
	go func() {
		defer pw.Close()
		log.Info("starting dump")
		if err = m.dump.Dump(); err != nil {
			errc <- err
		}
		log.Info("finished dump")
	}()
	// process output with destination implementation
	log.Info("start storing dump")
	if m.ArchiveName == "" {
		m.ArchiveName = filter.ReplaceAllString(m.URI+m.Database, "") + ".tgz"
	}
	written, dsterr := dst.Store(backup.Object{
		ID:   m.ArchiveName,
		Data: pr,
	})
	select {
	case srcerr := <-errc: // return src error if possible as well
		return written, fmt.Errorf("dst error: %v; src error: %v", dsterr, srcerr)
	case <-time.After(1 * time.Second):
		return written, dsterr
	}
}

func (m *mongoDBSource) Close() error {
	if m.dump != nil {
		m.dump.HandleInterrupt()
	}
	return nil
}
