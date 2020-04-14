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

package s3

import (
	"fmt"
	"testing"

	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/kubism-io/backup-operator/pkg/testutil"
	"github.com/onsi/ginkgo/reporters"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	pool     *dockertest.Pool
	resource *dockertest.Resource
	endpoint string
)

const (
	accessKeyID     = "TESTACCESSKEY"
	secretAccessKey = "TESTSECRETKEY"
)

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/s3-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "S3", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("mongosetup")
	By("bootstrapping both mongodbs and minio")
	pool, err = dockertest.NewPool("")
	Expect(err).ToNot(HaveOccurred())
	log.Info("spawn minio container")
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
	resource, err = pool.RunWithOptions(options)
	Expect(err).ToNot(HaveOccurred())
	endpoint = fmt.Sprintf("localhost:%s", resource.GetPort("9000/tcp"))
	log.Info("check minio connection", "endpoint", endpoint)
	err = testutil.WaitForS3(pool, endpoint, accessKeyID, secretAccessKey)
	Expect(err).ToNot(HaveOccurred())
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := pool.Purge(resource)
	Expect(err).ToNot(HaveOccurred())
})
