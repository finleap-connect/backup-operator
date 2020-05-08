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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kubism/backup-operator/pkg/logger"
)

type HelmEnvConfig struct {
	Kubeconfig string
	Stdout     io.Writer
	Stderr     io.Writer
}

type HelmEnv struct {
	Config *HelmEnvConfig
	Dir    string
	Bin    string
	Env    []string
	log    logger.Logger
}

func NewHelmEnv(config *HelmEnvConfig) (*HelmEnv, error) {
	bin := "helm" // fallback
	if value, ok := os.LookupEnv("HELM3"); ok {
		bin = value
	}
	dir, err := ioutil.TempDir("", "helmenv")
	if err != nil {
		return nil, err
	}
	return &HelmEnv{
		Config: config,
		Dir:    dir,
		Bin:    bin,
		Env: append(os.Environ(),
			fmt.Sprintf("KUBECONFIG=%s", config.Kubeconfig),
			fmt.Sprintf("XDG_CACHE_HOME=%s", filepath.Join(dir, "cache")),
			fmt.Sprintf("XDG_CONFIG_HOME=%s", filepath.Join(dir, "config")),
			fmt.Sprintf("XDG_DATA_HOME=%s", filepath.Join(dir, "data")),
		),
		log: logger.WithName("helmenv"),
	}, nil
}

func (e *HelmEnv) RepoAdd(name, url string) error {
	e.log.Info("adding repository", "name", name, "url", url)
	cmd := exec.Command(e.Bin, "repo", "add", name, url)
	e.setupCmd(cmd)
	return cmd.Run()
}

func (e *HelmEnv) RepoUpdate() error {
	e.log.Info("updating repositories")
	cmd := exec.Command(e.Bin, "repo", "update")
	e.setupCmd(cmd)
	return cmd.Run()
}

func (e *HelmEnv) Install(namespace, release, chart string, args ...string) error {
	log := e.log.WithValues("namespace", namespace, "release", release, "chart", chart, "args", args)
	log.Info("installing chart")
	args = append([]string{"install", "--namespace", namespace, "--wait", release, chart}, args...)
	cmd := exec.Command(e.Bin, args...)
	e.setupCmd(cmd)
	return cmd.Run()
}

func (e *HelmEnv) Uninstall(namespace, release string, args ...string) error {
	log := e.log.WithValues("namespace", namespace, "release", release, "args", args)
	log.Info("uninstalling chart")
	args = append([]string{"uninstall", "--namespace", namespace, release}, args...)
	cmd := exec.Command(e.Bin, args...)
	e.setupCmd(cmd)
	return cmd.Run()
}

func (e *HelmEnv) Close() error {
	return os.RemoveAll(e.Dir)
}

func (e *HelmEnv) setupCmd(cmd *exec.Cmd) {
	cmd.Env = e.Env
	cmd.Stdout = e.Config.Stdout
	cmd.Stderr = e.Config.Stderr
}
