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
	"bytes"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kubism/backup-operator/pkg/backup/mem"
	"github.com/kubism/backup-operator/pkg/backup/mongodb"
	"github.com/kubism/backup-operator/pkg/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3Destination", func() {
	It("should read buffer to s3", func() {
		data := []byte("temporarycontent")
		bucket := "bucketb"
		key := "keyb"
		src, _ := mem.NewBufferSource(key, data)
		dst, err := NewS3Destination(endpoint, accessKeyID, secretAccessKey, false, bucket, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		written, err := src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
		Expect(written).To(BeNumerically(">", 0))
		input := s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &key,
		}
		buf := aws.NewWriteAtBuffer([]byte{})
		downloader := s3manager.NewDownloader(dst.Session)
		_, err = downloader.Download(buf, &input)
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.Bytes()).Should(Equal(data))
	})
	DescribeTable("ensure retention for values",
		func(retention int, count int) {
			data := []byte("testcontent")
			bucket := fmt.Sprintf("bucket%d-%d", retention, count)
			dst, err := NewS3Destination(endpoint, accessKeyID, secretAccessKey, false, bucket, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(dst).ToNot(BeNil())
			for i := 0; i < count; i++ {
				_, err := dst.Client.PutObject(&s3.PutObjectInput{
					Body:   bytes.NewReader(data),
					Bucket: &bucket,
					Key:    aws.String(fmt.Sprintf("key%d-%d-%d", retention, count, i)),
				})
				Expect(err).ToNot(HaveOccurred())
			}
			input := &s3.ListObjectsInput{
				Bucket: &bucket,
			}
			objects := sortableObjectSlice{}
			Expect(dst.Client.ListObjectsPages(input,
				func(page *s3.ListObjectsOutput, lastPage bool) bool {
					for _, obj := range page.Contents {
						objects = append(objects, obj)
					}
					return true
				})).To(Succeed())
			sort.Sort(objects)
			expected := []string{}
			for _, obj := range objects[:retention] {
				expected = append(expected, *obj.Key)
			}
			err = dst.EnsureRetention(retention)
			Expect(err).ToNot(HaveOccurred())
			found := []string{}
			err = dst.Client.ListObjectsPages(input,
				func(page *s3.ListObjectsOutput, lastPage bool) bool {
					for _, obj := range page.Contents {
						found = append(found, *obj.Key)
					}
					return true
				})
			Expect(err).ToNot(HaveOccurred())
			sort.Strings(expected)
			sort.Strings(found)
			Expect(found).To(Equal(expected))
		},
		Entry("3 out of 5", 3, 5),
		Entry("4 out of 5", 4, 5),
		Entry("5 out of 12", 5, 12),
	)
	It("should stream from MongoDBSource to S3Destination and back", func() {
		name := "backup.tgz"
		src, err := mongodb.NewMongoDBSource(srcURI, "", name)
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		bucket := "bucketc"
		dst, err := NewS3Destination(endpoint, accessKeyID, secretAccessKey, false, bucket, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		written, err := src.Stream(dst)
		Expect(written).To(BeNumerically(">", 0))
		Expect(err).ToNot(HaveOccurred())
		input := s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &name,
		}
		buf := aws.NewWriteAtBuffer([]byte{})
		downloader := s3manager.NewDownloader(dst.Session)
		_, err = downloader.Download(buf, &input)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(buf.Bytes())).Should(BeNumerically(">", 100))
		src, err = NewS3Source(endpoint, accessKeyID, secretAccessKey, false, bucket, name)
		Expect(err).ToNot(HaveOccurred())
		Expect(src).ToNot(BeNil())
		mdst, err := mongodb.NewMongoDBDestination(dstURI)
		Expect(err).ToNot(HaveOccurred())
		Expect(mdst).ToNot(BeNil())
		_, err = src.Stream(mdst)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.FindTestData(dstURI)
		Expect(err).ToNot(HaveOccurred())
	})
})
