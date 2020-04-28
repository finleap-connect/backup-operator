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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	"github.com/kubism/backup-operator/pkg/mongodb"
	"github.com/kubism/backup-operator/pkg/s3"
	"github.com/spf13/cobra"
)

var mongodbCmd = &cobra.Command{
	Use:   "mongodb [flags] config",
	Short: "Backups mongodb using specified config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("config path expected as one and only argument")
		}
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()
		raw, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		var plan backupv1alpha1.MongoDBBackupPlan
		err = json.Unmarshal([]byte(os.ExpandEnv(string(raw))), &plan)
		if err != nil {
			return err
		}
		accessKeyID := ""
		if v := os.Getenv("S3_SECRET_ACCESS_KEY"); v != "" {
			accessKeyID = v
		}
		secretAccessKey := ""
		if v := os.Getenv("S3_SECRET_ACCESS_KEY"); v != "" {
			secretAccessKey = v
		}
		name := fmt.Sprintf("backup-%s.tgz", time.Now().Format("20060102150405"))
		src, err := mongodb.NewMongoDBSource(plan.Spec.URI, "", name)
		if err != nil {
			return err
		}
		prefix := fmt.Sprintf("%s/%s", plan.ObjectMeta.Namespace, plan.ObjectMeta.Name)
		s3c := plan.Spec.Destination.S3
		dst, err := s3.NewS3Destination(s3c.Endpoint, accessKeyID, secretAccessKey, s3c.UseSSL, s3c.Bucket, prefix)
		if err != nil {
			return err
		}
		err = src.Stream(dst)
		if err != nil {
			return err
		}
		err = dst.EnsureRetention(int(plan.Spec.Retention))
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mongodbCmd)
}
