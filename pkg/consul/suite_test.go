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

	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/testutil"
	"github.com/onsi/ginkgo/reporters"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	pool           *dockertest.Pool
	consulResource *dockertest.Resource
	consulAddr     string
	s3Resource     *dockertest.Resource
	s3Addr         string
)

const accessKeyID = "TESTACCESSKEY"
const secretAccessKey = "TESTSECRETKEY"

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/service-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "Consul", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("testsetup")
	By("bootstrapping s3 and consul")
	pool, err = dockertest.NewPool("")
	Expect(err).ToNot(HaveOccurred())
	log.Info("spawn consul container")
	consulResource, err = pool.Run("consul", "1.7", nil)
	Expect(err).ToNot(HaveOccurred())
	consulAddr = fmt.Sprintf("localhost:%s", consulResource.GetPort("8500/tcp"))
	log.Info("spawn s3 minio container")
	options := &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		PortBindings: map[dc.Port][]dc.PortBinding{
			"9000": {{HostPort: "9000"}},
		},
		Env: []string{
			fmt.Sprintf("MINIO_ACCESS_KEY=%s", accessKeyID),
			fmt.Sprintf("MINIO_SECRET_KEY=%s", secretAccessKey),
		},
	}
	s3Resource, err = pool.RunWithOptions(options)
	Expect(err).ToNot(HaveOccurred())
	s3Addr = fmt.Sprintf("localhost:%s", s3Resource.GetPort("9000/tcp"))

	err = testutil.WaitForConsul(pool, consulAddr)
	Expect(err).ToNot(HaveOccurred())
	log.Info("consul ready")

	err = testutil.WaitForS3(pool, s3Addr, accessKeyID, secretAccessKey)
	Expect(err).ToNot(HaveOccurred())
	log.Info("s3 ready")
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err1 := pool.Purge(consulResource)
	err2 := pool.Purge(s3Resource)
	Expect(err1).ToNot(HaveOccurred())
	Expect(err2).ToNot(HaveOccurred())
})
