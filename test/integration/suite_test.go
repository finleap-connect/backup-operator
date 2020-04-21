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

package integration

import (
	"os"
	"testing"

	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/testutil"
	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	kind *testutil.KindEnv
	helm *testutil.HelmEnv
)

// TODO:
// * create helm test env (make sure to override XDG_CONFIG_HOME etc
// * install mongodb
// * install minio
// * creater operator helm chart
// * make prebuild containers available
// * install backup-operator
// * kubectl wrapper?
// * use ingress to check content of minio?

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/integration-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Integration", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("integrationsetup")
	By("bootstrapping kind test cluster")
	kind, err = testutil.NewKindEnv(&testutil.KindEnvConfig{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	Expect(err).ToNot(HaveOccurred())
	err = kind.Start("int-test")
	Expect(err).ToNot(HaveOccurred())
	helm, err = testutil.NewHelmEnv(&testutil.HelmEnvConfig{
		Kubeconfig: kind.Kubeconfig,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	})
	Expect(err).ToNot(HaveOccurred())
	err = helm.RepoAdd("stable", "https://kubernetes-charts.storage.googleapis.com/")
	Expect(err).ToNot(HaveOccurred())
	err = helm.RepoAdd("bitnami", "https://charts.bitnami.com/bitnami")
	Expect(err).ToNot(HaveOccurred())
	err = helm.RepoUpdate()
	Expect(err).ToNot(HaveOccurred())
	err = helm.Install("a", "bitnami/mongodb")
	Expect(err).ToNot(HaveOccurred())
	err = helm.Install("b", "bitnami/mongodb")
	Expect(err).ToNot(HaveOccurred())
	err = helm.Install("c", "stable/minio")
	Expect(err).ToNot(HaveOccurred())
	log.Info("setup done")
	close(done)
}, 1200)

var _ = AfterSuite(func() {
	By("tearing down test cluster")
	if kind != nil {
		err := kind.Close()
		Expect(err).ToNot(HaveOccurred())
	}
})
