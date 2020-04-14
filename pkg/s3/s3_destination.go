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
	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/kubism-io/backup-operator/pkg/stream"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func NewS3Destination(endpoint, accessKeyID, secretAccessKey string, useSSL bool, bucket string) (*S3Destination, error) {
	newSession, err := session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(!useSSL),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	client := s3.New(newSession)
	// Create bucket, if not exists
	_, err = client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil { // If bucket already exists ignore error
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() != s3.ErrCodeBucketAlreadyExists || aerr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &S3Destination{
		Session:  newSession,
		Client:   client,
		Uploader: s3manager.NewUploader(newSession),
		Bucket:   bucket,
		log:      logger.WithName("s3dst"),
	}, nil
}

type S3Destination struct {
	Session  *session.Session
	Client   *s3.S3
	Uploader *s3manager.Uploader
	Bucket   string
	log      logger.Logger
}

func (s *S3Destination) Store(obj stream.Object) error {
	params := &s3manager.UploadInput{
		Bucket: &s.Bucket,
		Key:    &obj.ID,
		Body:   obj.Data,
	}
	s.log.Info("upload starting", "bucket", s.Bucket, "key", obj.ID)
	result, err := s.Uploader.Upload(params)
	if err != nil {
		return err
	}
	s.log.Info("upload successful", "result", result)
	return nil
}
