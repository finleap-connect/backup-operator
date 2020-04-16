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

package testutil

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kubism-io/backup-operator/pkg/logger"
)

type KindEnv struct {
	Dir        string
	Bin        string
	Name       string
	Kubeconfig string
	log        logger.Logger
}

func NewKindEnv() (*KindEnv, error) {
	bin := "kind" // fallback
	if value, ok := os.LookupEnv("KIND"); ok {
		bin = value
	}
	dir, err := ioutil.TempDir("", "kindenv")
	if err != nil {
		return nil, err
	}
	return &KindEnv{
		Dir: dir,
		Bin: bin,
		log: logger.WithName("kindenv"),
	}, nil
}

func (e *KindEnv) Start(name string) error {
	cmd := exec.Command(e.Bin, "create", "cluster", "--image", "kindest/node:v1.16.4", "--name", name, "--wait", "5m")
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command(e.Bin, "get", "kubeconfig", "--name", name)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	e.Kubeconfig = filepath.Join(e.Dir, "kubeconfig")
	err = ioutil.WriteFile(e.Kubeconfig, out, 0644)
	if err != nil {
		return err
	}
	e.log.Info("cluster created", "name", name)
	e.Name = name // let's remember the cluster name for cleanup
	return nil
}

func (e *KindEnv) Stop() error {
	cmd := exec.Command(e.Bin, "delete", "cluster", "--name", e.Name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	e.Name = ""
	return nil
}

func (e *KindEnv) Close() error {
	if e.Name != "" {
		return e.Stop()
	}
	return os.RemoveAll(e.Dir)
}
