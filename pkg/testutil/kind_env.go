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
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/finleap-connect/backup-operator/pkg/logger"
)

type KindEnvConfig struct {
	Stdout io.Writer
	Stderr io.Writer
}

type KindEnv struct {
	Config     *KindEnvConfig
	Dir        string
	Bin        string
	Name       string
	Kubeconfig string
	log        logger.Logger
}

func NewKindEnv(config *KindEnvConfig) (*KindEnv, error) {
	log := logger.WithName("kindenv")
	bin := "kind" // fallback
	if value, ok := os.LookupEnv("KIND"); ok {
		bin = value
	}
	dir, err := ioutil.TempDir("", "kindenv")
	if err != nil {
		return nil, err
	}
	name := "test"
	if value, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		name = value
	}
	log.Info("cluster created", "name", name)
	cmd := exec.Command(bin, "get", "kubeconfig", "--name", name)
	out, err := cmd.Output() // do not use setupCmd here
	if err != nil {
		return nil, err
	}
	kubeconfig := filepath.Join(dir, "kubeconfig")
	err = ioutil.WriteFile(kubeconfig, out, 0644)
	if err != nil {
		return nil, err
	}
	return &KindEnv{
		Config:     config,
		Dir:        dir,
		Bin:        bin,
		Name:       name,
		Kubeconfig: kubeconfig,
		log:        log,
	}, nil
}

func (e *KindEnv) LoadDockerImage(image string) error {
	cmd := exec.Command(e.Bin, "load", "docker-image", "--name", e.Name, image)
	e.setupCmd(cmd)
	return cmd.Run()
}

func (e *KindEnv) Close() error {
	return os.RemoveAll(e.Dir)
}

func (e *KindEnv) setupCmd(cmd *exec.Cmd) { // nolint:deadcode,unused
	cmd.Stdout = e.Config.Stdout
	cmd.Stderr = e.Config.Stderr
}
