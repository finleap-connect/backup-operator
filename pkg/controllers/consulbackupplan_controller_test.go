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

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

type UpdateConsulBackupPlanFunc = func(spec *backupv1alpha1.ConsulBackupPlanSpec)

func newConsulBackupPlan(namespace string, updates ...UpdateConsulBackupPlanFunc) *backupv1alpha1.ConsulBackupPlan {
	plan := &backupv1alpha1.ConsulBackupPlan{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      newTestName(),
		},
		Spec: backupv1alpha1.ConsulBackupPlanSpec{
			Schedule:              "* * * * *",
			ActiveDeadlineSeconds: 3600,
			Retention:             2,
			Address:               "localhost:27017",
			Destination: &backupv1alpha1.Destination{
				S3: &backupv1alpha1.S3{
					Endpoint:        "localhost:8000",
					Bucket:          "test",
					UseSSL:          false,
					AccessKeyID:     "a",
					SecretAccessKey: "b",
				},
			},
		},
	}
	for _, f := range updates {
		f(&plan.Spec)
	}
	return plan
}

func mustCreateNewConsulBackupPlan(namespace string, updates ...UpdateConsulBackupPlanFunc) *backupv1alpha1.ConsulBackupPlan {
	plan := newConsulBackupPlan(namespace, updates...)
	Expect(k8sClient.Create(context.Background(), plan)).Should(Succeed())
	return plan
}

var _ = Describe("ConsulBackupPlanReconciler", func() {
	ctx := context.Background()
	namespace := ""

	BeforeEach(func() {
		namespace = mustCreateNamespace()
	})
	AfterEach(func() {
		mustDeleteNamespace(namespace)
	})

	It("can create ConsulBackupPlans", func() {
		Context("with missing data", func() {
			Expect(k8sClient.Create(ctx, &backupv1alpha1.ConsulBackupPlan{})).ShouldNot(Succeed())
		})
		Context("with valid data", func() {
			plan := mustCreateNewConsulBackupPlan(namespace)
			defer mustRemoveFinalizers(plan)
		})
	})
	It("can process ConsulBackupPlans", func() {
		Context("which are just created", func() {
			plan := mustCreateNewConsulBackupPlan(namespace)
			defer mustRemoveFinalizers(plan)
			res := mustReconcile(plan)
			Expect(res.Requeue).To(Equal(false))
		})
		Context("which were deleted", func() {
			plan := mustCreateNewConsulBackupPlan(namespace)
			defer func() {
				// If this test fails, we need to make sure the finalizers are removed
				if err := k8sClient.Get(ctx, namespacedName(plan), plan); err == nil {
					mustRemoveFinalizers(plan)
				}
			}()
			res := mustReconcile(plan)
			Expect(res.Requeue).To(Equal(false))
			Expect(k8sClient.Delete(ctx, plan)).Should(Succeed())
			Expect(k8sClient.Get(ctx, namespacedName(plan), plan)).Should(Succeed())
			res = mustReconcile(plan)
			Expect(res.Requeue).To(Equal(false))
		})
	})
	DescribeTable("can process ConsulBackupPlans multiple times",
		func(count int) {
			plan := mustCreateNewConsulBackupPlan(namespace)
			defer mustRemoveFinalizers(plan)
			for i := 0; i < count; i++ {
				res := mustReconcile(plan)
				Expect(res.Requeue).To(Equal(false))
			}
		},
		Entry("twice", 2),
		Entry("three times", 3),
		Entry("five times", 5),
	)
	It("creates relevant Secret", func() {
		plan := mustCreateNewConsulBackupPlan(namespace)
		defer mustRemoveFinalizers(plan)
		res := mustReconcile(plan)
		Expect(res.Requeue).To(Equal(false))
		Expect(k8sClient.Get(ctx, namespacedName(plan), plan)).Should(Succeed())
		var secret corev1.Secret
		Expect(k8sClient.Get(ctx, types.NamespacedName{
			Namespace: plan.Status.Secret.Namespace,
			Name:      plan.Status.Secret.Name,
		}, &secret)).Should(Succeed())
		Expect(secret.Data).NotTo(BeNil())
		raw, ok := secret.Data[secretFieldName]
		Expect(ok).To(Equal(true))
		var content backupv1alpha1.ConsulBackupPlan
		Expect(json.Unmarshal(raw, &content)).Should(Succeed())
		Expect(content.Spec).To(Equal(plan.Spec))
	})
})
