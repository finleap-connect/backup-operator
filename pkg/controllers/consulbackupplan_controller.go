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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
)

// ConsulBackupPlanReconciler reconciles a ConsulBackupPlan object
type ConsulBackupPlanReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	Recorder           record.EventRecorder
	DefaultDestination *backupv1alpha1.Destination // TODO: to implement
	WorkerImage        string
}

// +kubebuilder:rbac:groups=backup.kubism.io,resources=consulbackupplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=backup.kubism.io,resources=consulbackupplans/status,verbs=get;update;patch

func (r *ConsulBackupPlanReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("consulbackupplan", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ConsulBackupPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1alpha1.ConsulBackupPlan{}).
		Complete(r)
}
