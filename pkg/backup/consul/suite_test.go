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

package consul

import (
	"fmt"
	"testing"

	"github.com/finleap-connect/backup-operator/pkg/logger"
	"github.com/finleap-connect/backup-operator/pkg/testutil"
	"github.com/ory/dockertest/v3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	pool        *dockertest.Pool
	srcResource *dockertest.Resource
	dstResource *dockertest.Resource
	srcURI      string
	dstURI      string
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("consulsetup")

	By("bootstrapping both consuls")
	pool, err = dockertest.NewPool("")
	Expect(err).ToNot(HaveOccurred())

	log.Info("spawn src consul container")
	srcResource, err = pool.Run("consul", "1.7", nil)
	Expect(err).ToNot(HaveOccurred())
	srcURI = fmt.Sprintf("localhost:%s", srcResource.GetPort("8500/tcp"))

	log.Info("spawn dst consul container")
	dstResource, err = pool.Run("consul", "1.7", nil)
	Expect(err).ToNot(HaveOccurred())
	dstURI = fmt.Sprintf("localhost:%s", dstResource.GetPort("8500/tcp"))

	err = testutil.WaitForConsul(pool, srcURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("src consul ready")

	err = testutil.WaitForConsul(pool, dstURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("dst consul ready")

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	var err error
	err = pool.Purge(srcResource)
	Expect(err).ToNot(HaveOccurred())
	err = pool.Purge(dstResource)
	Expect(err).ToNot(HaveOccurred())
})
