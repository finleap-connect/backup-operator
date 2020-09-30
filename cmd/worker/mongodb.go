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
	"github.com/kubism/backup-operator/pkg/backup/mongodb"
	"github.com/kubism/backup-operator/pkg/backup/s3"
	"github.com/kubism/backup-operator/pkg/logger"
	"github.com/kubism/backup-operator/pkg/metrics"
	"github.com/kubism/backup-operator/pkg/util"
	"github.com/spf13/cobra"
)

var mongodbCmd = &cobra.Command{
	Use:   "mongodb [flags] config",
	Short: "Backups mongodb using specified config",
	RunE: func(cmd *cobra.Command, args []string) error {
		log := logger.WithName("worker")
		// Load configuration
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
		// Setup metrics publisher
		mps := plan.Spec.Pushgateway
		mpc := metrics.DefaultConfig().
			WithApp("mongodb").
			WithURL(util.FallbackToEnv(mps.URL, "PUSHGATEWAY_URL")).
			WithUsername(util.FallbackToEnv(mps.Username, "PUSHGATEWAY_USERNAME")).
			WithPassword(util.FallbackToEnv(mps.Password, "PUSHGATEWAY_PASSWORD"))
		var mp metrics.MetricsPublisher
		if err := mpc.Validate(); err != nil {
			log.Error(err, "invalid metrics configuration falling back to NewNopMetricsPublisher")
			mp = metrics.NewNopMetricsPublisher()
		} else {
			log.Info("using pushgateway for metrics", "url", mpc.URL)
			mp = metrics.NewMetricsPublisher(mpc)
		}
		defer func() {
			mp.StopTimer()
			mp.PublishMetrics()
		}()
		// Backup
		mp.StartTimer()
		name := fmt.Sprintf("backup-%s.tgz", time.Now().Format("20060102150405"))
		src, err := mongodb.NewMongoDBSource(plan.Spec.URI, "", name)
		if err != nil {
			return err
		}
		prefix := fmt.Sprintf("%s/%s", plan.ObjectMeta.Namespace, plan.ObjectMeta.Name)
		s3c := plan.Spec.Destination.S3
		conf := &s3.S3DestinationConf{
			Endpoint:            s3c.Endpoint,
			AccessKey:           util.FallbackToEnv(s3c.AccessKeyID, "S3_ACCESS_KEY_ID"),
			SecretKey:           util.FallbackToEnv(s3c.SecretAccessKey, "S3_SECRET_ACCESS_KEY"),
			EncryptionKey:       util.NilIfEmpty(util.FallbackToEnv(s3c.EncryptionKey, "S3_ENCRYPTION_KEY")),
			EncryptionAlgorithm: util.FallbackToEnv(s3c.EncryptionAlgorithm, "S3_ENCRYPTION_ALGORITHM"),
			DisableSSL:          !s3c.UseSSL,
			Bucket:              s3c.Bucket,
			Prefix:              prefix,
		}
		dst, err := s3.NewS3Destination(conf)
		if err != nil {
			return err
		}
		written, err := src.Stream(dst)
		if err != nil {
			return err
		}
		mp.SetBackupSizeInBytes(written)
		err = dst.EnsureRetention(int(plan.Spec.Retention))
		if err != nil {
			return err
		}
		mp.SetSuccessfulRun()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mongodbCmd)
}
