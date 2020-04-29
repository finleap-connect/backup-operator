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

package v1alpha1

type Destination struct {
	// +optional
	// Configuration for S3 as backup target
	S3 *S3 `json:"s3,omitempty"`
}

type S3 struct {
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// +optional
	Bucket string `json:"bucket,omitempty"`
	// +optional
	UseSSL bool `json:"useSSL,omitempty"`
	// +optional
	AccessKeyID string `json:"accessKeyID,omitempty"`
	// +optional
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
}
