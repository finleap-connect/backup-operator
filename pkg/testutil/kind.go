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
	"os"
	"os/exec"
	"strings"

	"github.com/kubism-io/backup-operator/pkg/logger"
)

type KindEnv struct {
	Bin  string
	Name string
	log  logger.Logger
}

func NewKindEnv() (*KindEnv, error) {

	bin := "kind"
	if value, ok := os.LookupEnv("KIND"); ok {
		bin = value
	}
	return &KindEnv{
		Bin: bin,
		log: logger.WithName("kindenv"),
	}, nil
}

func (e *KindEnv) Start(name string) error {
	cmd := exec.Command(e.Bin, "create", "cluster", "--image", "kindest/node:v1.16.4", "--name", name, "--wait", "4m")
	err := cmd.Run()
	if err != nil {
		return err
	}
	e.Name = name
	cmd = exec.Command(e.Bin, "get", "kubeconfig", "--name", name)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	kubeconfig := strings.TrimSpace(string(out))
	e.log.Info("cluster created", "name", name)
	err = os.Setenv("KUBECONFIG", kubeconfig)
	if err != nil {
		return err
	}
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
	return nil
}
