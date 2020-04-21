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
	"fmt"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	testNamespace = "test"
)

type reconciler interface {
	Reconcile(req ctrl.Request) (ctrl.Result, error)
}

var (
	testConfig *rest.Config
	testClient client.Client
	testEnv    *envtest.Environment

	testReconcilers map[string]reconciler
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
	}

	var err error
	testConfig, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(testConfig).ToNot(BeNil())

	err = backupv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	testClient, err = client.New(testConfig, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(testClient).ToNot(BeNil())

	testReconcilers = map[string]reconciler{
		backupv1alpha1.MongoDBBackupPlanKind: &MongoDBBackupPlanReconciler{
			Client:             testClient,
			Log:                logf.Log.WithName("controllers").WithName("MongoDBBackupPlan"),
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

func mustReconcile(obj runtime.Object) ctrl.Result {
	var reconciler reconciler
	if _, ok := obj.(*backupv1alpha1.MongoDBBackupPlan); ok {
		reconciler = testReconcilers[backupv1alpha1.MongoDBBackupPlanKind]
	} else {
		panic("deadcode, otherwise reconciler was not properly registered for test")
	}
	req := newRequestFor(obj)
	res, err := reconciler.Reconcile(req)
	Expect(err).ToNot(HaveOccurred())
	return res
}
