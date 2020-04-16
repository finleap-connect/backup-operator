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
	"testing"

	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/kubism-io/backup-operator/pkg/testutil"
	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	env *testutil.KindEnv
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/integration-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Integration", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("integrationsetup")
	By("bootstrapping kind test cluster")
	env, err = testutil.NewKindEnv()
	Expect(err).ToNot(HaveOccurred())
	err = env.Start("int-test")
	Expect(err).ToNot(HaveOccurred())
	log.Info("setup done")
	close(done)
}, 300)

var _ = AfterSuite(func() {
	By("tearing down test cluster")
	if env != nil {
		err := env.Close()
		Expect(err).ToNot(HaveOccurred())
	}
})
