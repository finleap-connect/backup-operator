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
	"io"

	"github.com/kubism-io/backup-operator/pkg/backup"
	"github.com/kubism-io/backup-operator/pkg/logger"
	"github.com/minio/minio-go/v6"
)

func NewS3Destination(endpoint, accessKeyID, secretAccessKey string, useSSL bool, bucket string, filepath string) (backup.Destination, error) {
	return &s3Destination{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		UseSSL:          useSSL,
		Bucket:          bucket,
		Filepath:        filepath,
		log:             logger.WithName("s3dst"),
	}, nil
}

type s3Destination struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Bucket          string
	Filepath        string
	log             logger.Logger
}

func (s *s3Destination) Store(data io.Reader) error {
	client, err := minio.New(s.Endpoint, s.AccessKeyID, s.SecretAccessKey, s.UseSSL)
	if err != nil {
		return err
	}
	s.log.Info("upload starting", "endpoint", s.Endpoint, "bucket", s.Bucket, "filepath", s.Filepath)
	n, err := client.PutObject(s.Bucket, s.Filepath, data, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	s.log.Info("upload successful", "n", n)
	return nil
}
