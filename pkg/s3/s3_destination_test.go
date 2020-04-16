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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/kubism-io/backup-operator/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3Destination", func() {
	It("should read buffer to s3", func() {
		data := []byte("temporarycontent")
		bucket := "bucketb"
		key := "keyb"
		src, _ := util.NewBufferSource(key, data)
		dst, err := NewS3Destination(endpoint, accessKeyID, secretAccessKey, false, bucket, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		err = src.Stream(dst)
		Expect(err).ToNot(HaveOccurred())
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
	It("should ensure retention", func() {
		data := []byte("testcontent")
		bucket := "bucketc"
		dst, err := NewS3Destination(endpoint, accessKeyID, secretAccessKey, false, bucket, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(dst).ToNot(BeNil())
		_, err = dst.Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &bucket,
			Key:    aws.String("a"),
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = dst.Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &bucket,
			Key:    aws.String("b"),
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = dst.Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &bucket,
			Key:    aws.String("c"),
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = dst.Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &bucket,
			Key:    aws.String("d"),
		})
		Expect(err).ToNot(HaveOccurred())
		_, err = dst.Client.PutObject(&s3.PutObjectInput{
			Body:   bytes.NewReader(data),
			Bucket: &bucket,
			Key:    aws.String("e"),
		})
		Expect(err).ToNot(HaveOccurred())
		err = dst.EnsureRetention(3)
		Expect(err).ToNot(HaveOccurred())
		input := &s3.ListObjectsInput{
			Bucket: &bucket,
		}
		found := []string{}
		expected := []string{"c", "d", "e"}
		err = dst.Client.ListObjectsPages(input,
			func(page *s3.ListObjectsOutput, lastPage bool) bool {
				for _, obj := range page.Contents {
					found = append(found, *obj.Key)
				}
				return true
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(found).To(Equal(expected))

	})
})
