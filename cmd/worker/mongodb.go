package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	backupv1alpha1 "github.com/kubism-io/backup-operator/api/v1alpha1"
	"github.com/kubism-io/backup-operator/pkg/mongodb"
	"github.com/kubism-io/backup-operator/pkg/s3"
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
		name := fmt.Sprintf("%s-%s-%s", plan.ObjectMeta.Name, plan.ObjectMeta.Namespace, time.Now().Format("20060102150405"))
		src, err := mongodb.NewMongoDBSource(plan.Spec.URI, "", name)
		if err != nil {
			return err
		}
		c := plan.Spec.Destination.S3
		dst, err := s3.NewS3Destination(c.Endpoint, c.AccessKeyID, c.SecretAccessKey, c.UseSSL, c.Bucket)
		err = src.Stream(dst)
		if err != nil {
			return err
		}
		// TODO: retention
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mongodbCmd)
}
