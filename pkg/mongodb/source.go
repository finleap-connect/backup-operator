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

	"github.com/kubism-io/backup-operator/pkg/backup"

	"github.com/mongodb/mongo-tools-common/options"
	"github.com/mongodb/mongo-tools/mongodump"
)

func NewMongoDBSource(uri, database string) (backup.Source, error) {
	return &mongoDBSource{}, nil
}

type mongoDBSource struct {
	URI      string
	Database string
	dump     *mongodump.MongoDump
}

func (m *mongoDBSource) Backup(dst backup.Destination) error {

	opts := options.New("mongodump", "custom", "custom", mongodump.Usage, options.EnabledOptions{Auth: true, Connection: true, Namespace: true, URI: true})
	inputOpts := &mongodump.InputOptions{}
	opts.AddOptions(inputOpts)
	outputOpts := &mongodump.OutputOptions{}
	opts.AddOptions(outputOpts)
	opts.URI.AddKnownURIParameters(options.KnownURIOptionsReadPreference)
	_, err := opts.ParseArgs([]string{
		fmt.Sprintf("--uri=\"%s\"", m.URI),
		"--archive",
		"--gzip",
	})
	if err != nil {
		return err
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
		return err
	}
	pr, pw := io.Pipe()
	m.dump.OutputWriter = pw
	// start the backup in a separate routine
	errc := make(chan error)
	defer close(errc)
	go func() {
		defer pw.Close()
		if err = m.dump.Dump(); err != nil {
			errc <- err
		} else {
			errc <- nil
		}
	}()
	// process output with destination implementation
	err = dst.Store(pr)
	// TODO: ...
	err = <-errc
	// TODO: ...

	return nil
}

func (m *mongoDBSource) Close() error {
	if m.dump != nil {
		m.dump.HandleInterrupt()
	}
	return nil
}
