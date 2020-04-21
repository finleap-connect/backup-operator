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

	"github.com/kubism/backup-operator/pkg/logger"
)

const kindConfig = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
        authorization-mode: "AlwaysAllow"
  extraPortMappings:
  - containerPort: 80
    hostPort: 9080
    protocol: TCP
  - containerPort: 443
    hostPort: 9443
    protocol: TCP`

type KindEnvConfig struct {
	Stdout io.Writer
	Stderr io.Writer
}

type KindEnv struct {
	Config     *KindEnvConfig
	Dir        string
	Bin        string
	Name       string
	ConfigFile string // Kind specific config.yaml
	Kubeconfig string
	log        logger.Logger
}

func NewKindEnv(config *KindEnvConfig) (*KindEnv, error) {
	bin := "kind" // fallback
	if value, ok := os.LookupEnv("KIND"); ok {
		bin = value
	}
	dir, err := ioutil.TempDir("", "kindenv")
	if err != nil {
		return nil, err
	}
	configFile := filepath.Join(dir, "kind.yaml")
	err = ioutil.WriteFile(configFile, []byte(kindConfig), 0644)
	if err != nil {
		return nil, err
	}
	return &KindEnv{
		Config:     config,
		Dir:        dir,
		Bin:        bin,
		ConfigFile: configFile,
		log:        logger.WithName("kindenv"),
	}, nil
}

func (e *KindEnv) Start(name string) error {
	cmd := exec.Command(e.Bin, "create", "cluster",
		"--image", "kindest/node:v1.16.4", "--name", name,
		"--config", e.ConfigFile, "--wait", "5m")
	e.setupCmd(cmd)
	err := cmd.Run()
	if err != nil {
		return err
	}
	e.Name = name // let's remember the cluster name for cleanup
	cmd = exec.Command(e.Bin, "get", "kubeconfig", "--name", name)
	out, err := cmd.Output() // do not use setupCmd here
	if err != nil {
		return err
	}
	e.Kubeconfig = filepath.Join(e.Dir, "kubeconfig")
	err = ioutil.WriteFile(e.Kubeconfig, out, 0644)
	if err != nil {
		return err
	}
	e.log.Info("cluster created", "name", name)
	return nil
}

func (e *KindEnv) Stop() error {
	cmd := exec.Command(e.Bin, "delete", "cluster", "--name", e.Name)
	e.setupCmd(cmd)
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

func (e *KindEnv) setupCmd(cmd *exec.Cmd) {
	cmd.Stdout = e.Config.Stdout
	cmd.Stderr = e.Config.Stderr
}
