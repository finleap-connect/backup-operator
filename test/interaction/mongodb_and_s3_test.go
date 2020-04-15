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

package interaction

import (
	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kubism-io/backup-operator/pkg/mongodb"
	"github.com/kubism-io/backup-operator/pkg/s3"
	"github.com/kubism-io/backup-operator/pkg/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MongoAndS3", func() {
	It("should stream from MongoDBSource to S3Destination and back", func() {
		name := "backup.tgz"
		src, err := mongodb.NewMongoDBSource(srcURI, "", name)
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		bucket := "bucketc"
		dst, err := s3.NewS3Destination(endpoint, accessKeyID, secretAccessKey, nil, false, bucket)
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		err = src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
		input := awss3.GetObjectInput{
			Bucket: &bucket,
			Key:    &name,
		}
		buf := aws.NewWriteAtBuffer([]byte{})
		downloader := s3manager.NewDownloader(dst.Session)
		_, err = downloader.Download(buf, &input)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(buf.Bytes())).Should(BeNumerically(">", 100))
		src, err = s3.NewS3Source(endpoint, accessKeyID, secretAccessKey, nil, false, bucket, []string{name})
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		mdst, err := mongodb.NewMongoDBDestination(dstURI)
		Expect(err).ToNot(HaveOccurred())
		Expect(mdst).ToNot(BeNil())
		err = src.Stream(mdst)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.FindTestData(dstURI)
		Expect(err).ToNot(HaveOccurred())
	})
})
