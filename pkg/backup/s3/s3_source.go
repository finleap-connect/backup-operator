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
	"io"
	"time"

	"github.com/kubism/backup-operator/pkg/backup"
	"github.com/kubism/backup-operator/pkg/logger"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func NewS3Source(endpoint, accessKeyID, secretAccessKey string, encryptionKey *string, useSSL bool, bucket, key string) (*S3Source, error) {
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
			if aerr.Code() != s3.ErrCodeBucketAlreadyExists && aerr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &S3Source{
		Session:       newSession,
		Client:        client,
		EncryptionKey: encryptionKey,
		Downloader:    s3manager.NewDownloader(newSession),
		Bucket:        bucket,
		Key:           key,
		log:           logger.WithName("s3src"),
	}, nil
}

type S3Source struct {
	Session       *session.Session
	Client        *s3.S3
	Downloader    *s3manager.Downloader
	Bucket        string
	Key           string
	EncryptionKey *string
	log           logger.Logger
}

func (s *S3Source) Stream(dst backup.Destination) (int64, error) {
	log := s.log
	// Use sequential writes to be able tu use stub implementation
	s.Downloader.Concurrency = 1
	params := &s3.GetObjectInput{
		Bucket:         &s.Bucket,
		Key:            &s.Key,
		SSECustomerKey: s.EncryptionKey,
	}
	pr, pw := io.Pipe()
	errc := make(chan error, 1)
	defer close(errc)
	go func() {
		defer pw.Close()
		log.Info("download starting", "bucket", s.Bucket, "key", s.Key)
		numBytes, err := s.Downloader.Download(writerAtStub{pw}, params)
		if err != nil {
			errc <- err
		}
		log.Info("finished download", "numBytes", numBytes)
	}()
	written, dsterr := dst.Store(backup.Object{
		ID:   s.Key,
		Data: pr,
	})
	select {
	case srcerr := <-errc: // return src error if possible as well
		return written, fmt.Errorf("dst error: %v; src error: %v", dsterr, srcerr)
	case <-time.After(1 * time.Second):
		return written, dsterr
	}
}

type writerAtStub struct {
	w io.Writer
}

func (fw writerAtStub) WriteAt(p []byte, offset int64) (n int, err error) {
	return fw.w.Write(p) // ignore 'offset' because we forced sequential downloads
}
