package controllers

import (
	"context"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type UpdateMongoDBBackupPlanFunc = func(spec *backupv1alpha1.MongoDBBackupPlanSpec)

func newMongoDBBackupPlan(updates ...UpdateMongoDBBackupPlanFunc) *backupv1alpha1.MongoDBBackupPlan {
	plan := &backupv1alpha1.MongoDBBackupPlan{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespace,
			Name:      newTestName(),
		},
		Spec: backupv1alpha1.MongoDBBackupPlanSpec{
			Schedule:              "* * * * *",
			ActiveDeadlineSeconds: 3600,
			Retention:             2,
			URI:                   "mongodb://localhost:27017",
			Destination: &backupv1alpha1.Destination{
				S3: &backupv1alpha1.S3{
					Endpoint:        "localhost:8000",
					Bucket:          "test",
					UseSSL:          false,
					AccessKeyID:     "A",
					SecretAccessKey: "B",
				},
			},
		},
	}
	for _, f := range updates {
		f(&plan.Spec)
	}
	return plan
}

func mustCreateNewMongoDBBackupPlan(updates ...UpdateMongoDBBackupPlanFunc) *backupv1alpha1.MongoDBBackupPlan {
	plan := newMongoDBBackupPlan(updates...)
	Expect(testClient.Create(context.Background(), plan)).Should(Succeed())
	return plan
}

var _ = Describe("VaultSecretReconciler", func() {
	ctx := context.Background()
	It("can create MongoDBBackupPlans", func() {
		Context("with missing data", func() {
			Expect(testClient.Create(ctx, &backupv1alpha1.MongoDBBackupPlan{})).ShouldNot(Succeed())
		})
		Context("with valid data", func() {
			mustCreateNewMongoDBBackupPlan()
		})
	})
	It("can process MongoDBBackupPlans", func() {
		Context("which are just created", func() {
			res := mustReconcile(mustCreateNewMongoDBBackupPlan())
			Expect(res.Requeue).To(Equal(false))
		})
	})
})
