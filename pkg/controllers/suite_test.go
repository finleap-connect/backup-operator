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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	"github.com/kubism/backup-operator/pkg/testutil"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

type reconciler interface {
	Reconcile(req ctrl.Request) (ctrl.Result, error)
}

var (
	config    *rest.Config
	k8sClient client.Client
	env       *envtest.Environment
	kind      *testutil.KindEnv
	helm      *testutil.HelmEnv

	reconcilers map[string]reconciler

	workerImage string = os.Getenv("DOCKER_IMG")

	shouldRunLongTests bool = os.Getenv("TEST_LONG") != ""
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	var err error
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")

	kind, err = testutil.NewKindEnv(&testutil.KindEnvConfig{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	Expect(err).ToNot(HaveOccurred())

	if shouldRunLongTests { // Only required in e2e/integration tests
		helm, err = testutil.NewHelmEnv(&testutil.HelmEnvConfig{
			Kubeconfig: kind.Kubeconfig,
			Stdout:     os.Stdout,
			Stderr:     os.Stderr,
		})
		Expect(err).ToNot(HaveOccurred())
		err = helm.RepoAdd("stable", "https://kubernetes-charts.storage.googleapis.com/")
		Expect(err).ToNot(HaveOccurred())
		err = helm.RepoAdd("bitnami", "https://charts.bitnami.com/bitnami")
		Expect(err).ToNot(HaveOccurred())
		err = helm.RepoUpdate()
		Expect(err).ToNot(HaveOccurred())

	}

	config, err = clientcmd.BuildConfigFromFlags("", kind.Kubeconfig)
	Expect(err).ToNot(HaveOccurred())
	Expect(config).ToNot(BeNil())
	useExistingCluster := true
	env = &envtest.Environment{
		Config:             config,
		UseExistingCluster: &useExistingCluster,
		CRDDirectoryPaths:  []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}
	config, err = env.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(config).ToNot(BeNil())

	err = backupv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	reconcilers = map[string]reconciler{
		backupv1alpha1.MongoDBBackupPlanKind: &MongoDBBackupPlanReconciler{
			Client:             k8sClient,
			Log:                logf.Log.WithName("controllers").WithName("MongoDBBackupPlan"),
			Recorder:           &record.FakeRecorder{},
			Scheme:             scheme.Scheme,
			DefaultDestination: nil, // TODO
			WorkerImage:        workerImage,
		},
		backupv1alpha1.ConsulBackupPlanKind: &ConsulBackupPlanReconciler{
			Client:             k8sClient,
			Log:                logf.Log.WithName("controllers").WithName("ConsulBackupPlan"),
			Recorder:           &record.FakeRecorder{},
			Scheme:             scheme.Scheme,
			DefaultDestination: nil, // TODO
			WorkerImage:        "test",
		},
	}

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := env.Stop()
	Expect(err).ToNot(HaveOccurred())
	if kind != nil {
		kind.Close()
	}
	if helm != nil {
		helm.Close()
	}
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

func mustReconcile(obj runtime.Object) ctrl.Result {
	var reconciler reconciler
	if _, ok := obj.(*backupv1alpha1.MongoDBBackupPlan); ok {
		reconciler = reconcilers[backupv1alpha1.MongoDBBackupPlanKind]
	} else if _, ok := obj.(*backupv1alpha1.ConsulBackupPlan); ok {
		reconciler = reconcilers[backupv1alpha1.ConsulBackupPlanKind]
	} else {
		panic("deadcode, otherwise reconciler was not properly registered for test")
	}
	req := newRequestFor(obj)
	res, err := reconciler.Reconcile(req)
	Expect(err).ToNot(HaveOccurred())
	return res
}

func createNamespace(name string) error {
	ctx := context.Background()
	return k8sClient.Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

func deleteNamespace(name string) error {
	ctx := context.Background()
	return k8sClient.Delete(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	})
}

func mustCreateNamespace() string {
	name := newTestName()
	Expect(createNamespace(name)).To(Succeed())
	return name
}

func mustDeleteNamespace(name string) {
	Expect(deleteNamespace(name)).To(Succeed())
}

func mustRemoveFinalizers(obj runtime.Object) {
	ctx := context.Background()
	Expect(k8sClient.Get(ctx, namespacedName(obj), obj)).To(Succeed())
	accessor, err := meta.Accessor(obj)
	Expect(err).ToNot(HaveOccurred())
	accessor.SetFinalizers([]string{})
	Expect(k8sClient.Update(ctx, obj)).To(Succeed())
}
