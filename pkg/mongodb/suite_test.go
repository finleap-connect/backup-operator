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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

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
	By("bootstrapping mongodb")
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
	log.Info("retry src mongo connection", "uri", srcURI)
	err = waitForMongoDB(pool, srcURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("insert test data", "uri", srcURI)
	err = insertTestData(srcURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("retry dst mongo connection", "uri", dstURI)
	err = waitForMongoDB(pool, dstURI)
	Expect(err).ToNot(HaveOccurred())
	log.Info("mongo databases ready")
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := pool.Purge(srcResource)
	Expect(err).ToNot(HaveOccurred())
	err = pool.Purge(dstResource)
	Expect(err).ToNot(HaveOccurred())
})

func waitForMongoDB(pool *dockertest.Pool, uri string) error {
	return pool.Retry(func() error {
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
}

func insertTestData(uri string) error {
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
	collection := client.Database("testing").Collection("numbers")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = collection.InsertOne(ctx, bson.M{"name": "pi", "value": 3.14159})
	if err != nil {
		return err
	}
	return nil
}

func findTestData(uri string) error {
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
	collection := client.Database("testing").Collection("numbers")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return err
	}
	if n < 1 {
		return fmt.Errorf("fount insufficent documents: %d", n)
	}
	return nil
}
