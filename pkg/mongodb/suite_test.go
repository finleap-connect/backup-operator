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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/onsi/ginkgo/reporters"
	"github.com/ory/dockertest/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	pool     *dockertest.Pool
	resource *dockertest.Resource
	uri      string
)

func TestMongoDB(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter("../../reports/mongodb-junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "MongoDB", []Reporter{junitReporter})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	log := logger.WithName("mongosetup")
	By("bootstrapping mongodb")
	pool, err = dockertest.NewPool("")
	Expect(err).ToNot(HaveOccurred())
	log.Info("spawn mongo container")
	resource, err = pool.Run("mongo", "4.2", nil)
	Expect(err).ToNot(HaveOccurred())
	uri = fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp"))
	log.Info("retry mongo connection", "uri", uri)
	err = pool.Retry(func() error {
		var err error
		client, err := mongo.NewClient(options.Client().ApplyURI(uri))
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = client.Connect(ctx)
		if err != nil {
			return err
		}
		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			return err
		}
		return nil
	})
	Expect(err).ToNot(HaveOccurred())
	log.Info("mongo ready")
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := pool.Purge(resource)
	Expect(err).ToNot(HaveOccurred())
})
