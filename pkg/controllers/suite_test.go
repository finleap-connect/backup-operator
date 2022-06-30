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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	backupv1alpha1 "github.com/finleap-connect/backup-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

// +kubebuilder:docs-gen:collapse=Imports

var (
	k8sClient client.Client // You'll be using this client in your tests.
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc

	reconcilers map[string]reconcile.Reconciler

	workerImage   = os.Getenv("IMG")
	testNamespace = "test"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func(done Done) {
	var err error
	ctx, cancel = context.WithCancel(context.Background())
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	/*
		First, the envtest cluster is configured to read CRDs from the CRD directory Kubebuilder scaffolds for you.
	*/
	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	/*
		Then, we start the envtest cluster.
	*/
	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = backupv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	err = k8sClient.Create(ctx, &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	})
	Expect(err).ToNot(HaveOccurred())

	reconcilers = map[string]reconcile.Reconciler{
		backupv1alpha1.MongoDBBackupPlanKind: &BackupPlanReconciler{
			Client:      k8sClient,
			Log:         logf.Log.WithName("controllers").WithName("MongoDBBackupPlan"),
			Recorder:    &record.FakeRecorder{},
			Scheme:      scheme.Scheme,
			WorkerImage: workerImage,
			Type:        &backupv1alpha1.MongoDBBackupPlan{},
		},
		backupv1alpha1.ConsulBackupPlanKind: &BackupPlanReconciler{
			Client:      k8sClient,
			Log:         logf.Log.WithName("controllers").WithName("ConsulBackupPlan"),
			Recorder:    &record.FakeRecorder{},
			Scheme:      scheme.Scheme,
			WorkerImage: workerImage,
			Type:        &backupv1alpha1.ConsulBackupPlan{},
		},
	}

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// Helper

var testNameCounter = 0

func newTestName() string {
	testNameCounter += 1
	return fmt.Sprintf("test%d", testNameCounter)
}

func namespacedName(obj runtime.Object) types.NamespacedName {
	obj.GetObjectKind()
	accessor, err := meta.Accessor(obj)
	Expect(err).ToNot(HaveOccurred())
	accessor.GetResourceVersion()
	return types.NamespacedName{
		Namespace: accessor.GetNamespace(),
		Name:      accessor.GetName(),
	}
}

func newRequestFor(obj runtime.Object) ctrl.Request {
	return ctrl.Request{
		NamespacedName: namespacedName(obj),
	}
}

func mustReconcile(ctx context.Context, obj runtime.Object) ctrl.Result {
	var reconciler reconcile.Reconciler
	if _, ok := obj.(*backupv1alpha1.MongoDBBackupPlan); ok {
		reconciler = reconcilers[backupv1alpha1.MongoDBBackupPlanKind]
	} else if _, ok := obj.(*backupv1alpha1.ConsulBackupPlan); ok {
		reconciler = reconcilers[backupv1alpha1.ConsulBackupPlanKind]
	} else {
		panic("deadcode, otherwise reconciler was not properly registered for test")
	}
	req := newRequestFor(obj)
	res, err := reconciler.Reconcile(ctx, req)
	Expect(err).ToNot(HaveOccurred())
	return res
}

func mustRemoveFinalizers(ctx context.Context, obj client.Object) {
	Expect(k8sClient.Get(ctx, namespacedName(obj), obj)).To(Succeed())
	accessor, err := meta.Accessor(obj)
	Expect(err).ToNot(HaveOccurred())
	accessor.SetFinalizers([]string{})
	Expect(k8sClient.Update(ctx, obj)).To(Succeed())
}
