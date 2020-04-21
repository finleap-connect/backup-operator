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
	"fmt"

	backupv1alpha1 "github.com/kubism/backup-operator/api/v1alpha1"
	"github.com/kubism/backup-operator/pkg/util"

	"github.com/go-logr/logr"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	finalizerName   = "backup.kubism.io"
	secretFieldName = "plan.json"
)

// MongoDBBackupPlanReconciler reconciles a MongoDBBackupPlan object
type MongoDBBackupPlanReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	Recorder           record.EventRecorder
	DefaultDestination *backupv1alpha1.Destination // TODO: to implement
	WorkerImage        string
}

// +kubebuilder:rbac:groups=backup.kubism.io,resources=mongodbbackupplans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=backup.kubism.io,resources=mongodbbackupplans/status,verbs=get;update;patch

func (r *MongoDBBackupPlanReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("mongodbbackupplan", req.NamespacedName)

	var plan backupv1alpha1.MongoDBBackupPlan
	if err := r.Get(ctx, req.NamespacedName, &plan); err != nil {
		log.Error(err, "unable to fetch MongoDBBackupPlan")
		// We'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check whether object is being deleted
	if plan.ObjectMeta.DeletionTimestamp.IsZero() {
		// Object is not being deleted, but let's make sure is has our finalizer
		if !util.ContainsString(plan.ObjectMeta.Finalizers, finalizerName) {
			plan.ObjectMeta.Finalizers = append(plan.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &plan); err != nil {
				return ctrl.Result{}, err
			}
			log.Info("added finalizer")
			r.Recorder.Event(&plan, corev1.EventTypeNormal, "Updated", "Added finalizer to object")
		}
	} else { // Object is being deleted
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Deletion in progress")
		if util.ContainsString(plan.ObjectMeta.Finalizers, finalizerName) {
			// Finalizer is present, so let's cleanup our owned resources
			// TODO: delete cronjob
			// TODO: if default destination is used, check if additional resources (e.g. secret) should be freed
			// Finally remove the finalizer
			plan.ObjectMeta.Finalizers = util.RemoveString(plan.ObjectMeta.Finalizers, finalizerName)
			if err := r.Update(ctx, &plan); err != nil {
				log.Error(err, "failed to remove finalizer")
				r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", "Failed to remove finalizer")
				return ctrl.Result{}, err
			}
		}
		// Cleanup was successful or not required, so let's return
		return ctrl.Result{}, nil
	}

	// TODO: validate plan

	// First we create or update the secret before checking the related CronJob
	var secret corev1.Secret
	// If secret does not exist, let's create a new name
	if plan.Status.Secret == nil {
		secret.ObjectMeta.Name = req.Name // TODO: maybe introduce a hash of content?
		secret.ObjectMeta.Namespace = req.Namespace
	} else {
		err := r.Get(ctx, types.NamespacedName{
			Namespace: plan.Status.Secret.Namespace,
			Name:      plan.Status.Secret.Name,
		}, &secret)
		if client.IgnoreNotFound(err) != nil { // Unexpected error
			r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Checking owned secret failed with: %v", err))
			return ctrl.Result{}, err
		} else if err != nil {
			// Not found so let's reset the reference and let's re-create it
			secret.ObjectMeta.Name = plan.Status.Secret.Name
			secret.ObjectMeta.Namespace = plan.Status.Secret.Namespace
			plan.Status.Secret = nil
		}
	}
	// Let's compute the content of the secret
	if secret.Data == nil {
		secret.Data = map[string][]byte{}
	}
	raw, err := json.Marshal(&plan)
	if err != nil {
		// TODO: the follow can potentially be used to extract information from the outputted json \o/
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Unable to marshal plan: %v", err))
		return ctrl.Result{}, err
	}
	secret.Data[secretFieldName] = raw
	// Finally create or update the secret
	if plan.Status.Secret != nil {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Updating secret")
		err = r.Update(ctx, &secret)
	} else {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Creating secret")
		err = r.Create(ctx, &secret)
	}
	if err != nil {
		log.Error(err, "failed to create or update secret")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Update or creation of secret failed with: %v", err))
		return ctrl.Result{}, err
	}
	// Let's make sure to store the reference
	secretRef, err := ref.GetReference(r.Scheme, &secret)
	if err != nil {
		log.Error(err, "failed to get secret reference")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Failed to get secret reference: %v", err))
		return ctrl.Result{}, err

	}
	plan.Status.Secret = secretRef

	// TODO: create or update cronjob
	// TODO: if default destination is used, check if additional resources (e.g. secret) should be created

	return ctrl.Result{}, nil
}

func (r *MongoDBBackupPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("mongodbbackupplan-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1alpha1.MongoDBBackupPlan{}).
		Owns(&batchv1beta1.CronJob{}).
		Complete(r)
}
