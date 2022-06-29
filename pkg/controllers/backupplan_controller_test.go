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

package controllers

import (
	"context"
	"encoding/json"

	backupv1alpha1 "github.com/finleap-connect/backup-operator/api/v1alpha1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

const (
	accessKeyID     = "TESTACCESSKEY"
	secretAccessKey = "TESTSECRETKEY"
)

// Add api types to test here
var planTypes = [2]backupv1alpha1.BackupPlan{
	&backupv1alpha1.ConsulBackupPlan{},
	&backupv1alpha1.MongoDBBackupPlan{},
}

type CreateNewBackupPlanFunc = func(namespace string) backupv1alpha1.BackupPlan

// Add function to create api types to test here
var createTypeFuncs = map[string]CreateNewBackupPlanFunc{
	backupv1alpha1.ConsulBackupPlanKind: func(namespace string) backupv1alpha1.BackupPlan {
		return newConsulBackupPlan(namespace)
	},
	backupv1alpha1.MongoDBBackupPlanKind: func(namespace string) backupv1alpha1.BackupPlan {
		return newMongoDBBackupPlan(namespace)
	},
}

type UpdateMongoDBBackupPlanFunc = func(spec *backupv1alpha1.MongoDBBackupPlan)
type UpdateConsulBackupPlanFunc = func(spec *backupv1alpha1.ConsulBackupPlan)

func newObjectMeta(namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Namespace: namespace,
		Name:      newTestName(),
	}
}

func newBackupPlanSpec(namespace string) backupv1alpha1.BackupPlanSpec {
	return backupv1alpha1.BackupPlanSpec{
		Schedule:              "* * * * *",
		ActiveDeadlineSeconds: 3600,
		Retention:             2,
		Destination: &backupv1alpha1.Destination{
			S3: &backupv1alpha1.S3{
				Endpoint:        "localhost:8000",
				Bucket:          "test",
				UseSSL:          false,
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				PartSize:        5242880,
			},
		},
		Pushgateway: &backupv1alpha1.Pushgateway{},
	}
}

func newConsulBackupPlan(namespace string, updates ...UpdateConsulBackupPlanFunc) backupv1alpha1.BackupPlan {
	plan := &backupv1alpha1.ConsulBackupPlan{
		ObjectMeta: newObjectMeta(namespace),
		Spec: backupv1alpha1.ConsulBackupPlanSpec{
			BackupPlanSpec: newBackupPlanSpec(namespace),
			Address:        "localhost:27017",
		},
	}
	for _, f := range updates {
		f(plan)
	}
	return plan
}

func newMongoDBBackupPlan(namespace string, updates ...UpdateMongoDBBackupPlanFunc) backupv1alpha1.BackupPlan {
	plan := &backupv1alpha1.MongoDBBackupPlan{
		ObjectMeta: newObjectMeta(namespace),
		Spec: backupv1alpha1.MongoDBBackupPlanSpec{
			BackupPlanSpec: newBackupPlanSpec(namespace),
			URI:            "mongodb://localhost:27017",
		},
	}
	for _, f := range updates {
		f(plan)
	}
	return plan
}

func mustCreateNewBackupPlan(planType backupv1alpha1.BackupPlan, namespace string) backupv1alpha1.BackupPlan {
	f := createTypeFuncs[planType.GetKind()]
	plan := f(namespace)
	Expect(k8sClient.Create(context.Background(), plan)).Should(Succeed())
	return plan
}

// General backup reconciler tests
var _ = Describe("BackupPlanReconciler", func() {
	ctx := context.Background()

	It("can create BackupPlans", func() {
		Context("with missing data", func() {
			for _, planType := range planTypes {
				Expect(k8sClient.Create(ctx, planType.New())).ShouldNot(Succeed())
			}
		})
		Context("with valid data", func() {
			for _, planType := range planTypes {
				plan := mustCreateNewBackupPlan(planType, testNamespace)
				defer mustRemoveFinalizers(ctx, plan)
			}
		})
	})
	It("can process BackupPlans", func() {
		Context("which are just created", func() {
			for _, planType := range planTypes {
				plan := mustCreateNewBackupPlan(planType, testNamespace)
				defer mustRemoveFinalizers(ctx, plan)
				res := mustReconcile(ctx, plan)
				Expect(res.Requeue).To(Equal(false))
			}
		})
		Context("which were deleted", func() {
			for _, planType := range planTypes {
				plan := mustCreateNewBackupPlan(planType, testNamespace)
				defer func() {
					// If this test fails, we need to make sure the finalizers are removed
					if err := k8sClient.Get(ctx, namespacedName(plan), plan); err == nil {
						mustRemoveFinalizers(ctx, plan)
					}
				}()
				res := mustReconcile(ctx, plan)
				Expect(res.Requeue).To(Equal(false))
				Expect(k8sClient.Delete(ctx, plan)).Should(Succeed())
				Expect(k8sClient.Get(ctx, namespacedName(plan), plan)).Should(Succeed())
				res = mustReconcile(ctx, plan)
				Expect(res.Requeue).To(Equal(false))
				// Check if the owned resources were freed
				var secret corev1.Secret
				Expect(client.IgnoreNotFound(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: plan.GetStatus().Secret.Namespace,
					Name:      plan.GetStatus().Secret.Name,
				}, &secret))).Should(Succeed())
				var cronJob batchv1beta1.CronJob
				Expect(client.IgnoreNotFound(k8sClient.Get(ctx, types.NamespacedName{
					Namespace: plan.GetStatus().CronJob.Namespace,
					Name:      plan.GetStatus().CronJob.Name,
				}, &cronJob))).Should(Succeed())
			}
		})
	})
	DescribeTable("can process BackupPlans multiple times",
		func(count int) {
			for _, planType := range planTypes {
				plan := mustCreateNewBackupPlan(planType, testNamespace)
				defer mustRemoveFinalizers(ctx, plan)
				for i := 0; i < count; i++ {
					res := mustReconcile(ctx, plan)
					Expect(res.Requeue).To(Equal(false))
				}
			}
		},
		Entry("twice", 2),
		Entry("three times", 3),
		Entry("five times", 5),
	)
	It("creates relevant Secret", func() {
		for _, planType := range planTypes {
			plan := mustCreateNewBackupPlan(planType, testNamespace)
			defer mustRemoveFinalizers(ctx, plan)
			res := mustReconcile(ctx, plan)
			Expect(res.Requeue).To(Equal(false))
			Expect(k8sClient.Get(ctx, namespacedName(plan), plan)).Should(Succeed())
			var secret corev1.Secret
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: plan.GetStatus().Secret.Namespace,
				Name:      plan.GetStatus().Secret.Name,
			}, &secret)).Should(Succeed())
			Expect(secret.Data).NotTo(BeNil())
			raw, ok := secret.Data[secretFieldName]
			Expect(ok).To(Equal(true))
			content := plan.New()
			Expect(json.Unmarshal(raw, &content)).Should(Succeed())
			Expect(content.GetSpec()).To(Equal(plan.GetSpec()))
		}
	})
	It("creates relevant CronJob", func() {
		for _, planType := range planTypes {
			plan := mustCreateNewBackupPlan(planType, testNamespace)
			defer mustRemoveFinalizers(ctx, plan)
			res := mustReconcile(ctx, plan)
			Expect(res.Requeue).To(Equal(false))
			Expect(k8sClient.Get(ctx, namespacedName(plan), plan)).Should(Succeed())
			var cronJob batchv1beta1.CronJob
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: plan.GetStatus().CronJob.Namespace,
				Name:      plan.GetStatus().CronJob.Name,
			}, &cronJob)).Should(Succeed())
		}
	})
})
