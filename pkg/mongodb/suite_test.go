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

package mongodb

import (
	"fmt"
	"testing"

	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/testutil"
	"github.com/onsi/ginkgo/reporters"
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

func TestMongoDB(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/mongodb-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "MongoDB", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("mongosetup")
	By("bootstrapping both mongodbs")
	pool, err = dockertest.NewPool("")
	Expect(err).ToNot(HaveOccurred())
	log.Info("spawn src mongo container")
	srcResource, err = pool.Run("mongo", "4.2", nil)
	Expect(err).ToNot(HaveOccurred())
	srcURI = fmt.Sprintf("mongodb://localhost:%s", srcResource.GetPort("27017/tcp"))
	log.Info("spawn dst mongo container")
	dstResource, err = pool.Run("mongo", "4.2", nil)
	Expect(err).ToNot(HaveOccurred())
	dstURI = fmt.Sprintf("mongodb://localhost:%s", dstResource.GetPort("27017/tcp"))
	log.Info("check src mongo connection", "uri", srcURI)
	err = testutil.WaitForMongoDB(pool, srcURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("insert test data", "uri", srcURI)
	err = testutil.InsertTestData(srcURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("check dst mongo connection", "uri", dstURI)
	err = testutil.WaitForMongoDB(pool, dstURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("mongo databases ready")
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err1 := pool.Purge(srcResource)
	err2 := pool.Purge(dstResource)
	Expect(err1).ToNot(HaveOccurred())
	Expect(err2).ToNot(HaveOccurred())
})
