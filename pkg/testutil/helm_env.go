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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kubism-io/backup-operator/pkg/logger"
)

type HelmEnv struct {
	Dir string
	Bin string
	Env []string
	log logger.Logger
}

func NewHelmEnv(kubeconfig string) (*HelmEnv, error) {
	bin := "helm" // fallback
	if value, ok := os.LookupEnv("HELM3"); ok {
		bin = value
	}
	dir, err := ioutil.TempDir("", "helmenv")
	if err != nil {
		return nil, err
	}
	return &HelmEnv{
		Dir: dir,
		Bin: bin,
		Env: append(os.Environ(),
			fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
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
	cmd.Env = e.Env
	return cmd.Run()
}

func (e *HelmEnv) Close() error {
	return os.RemoveAll(e.Dir)
}
