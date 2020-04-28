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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
			if plan.Status.Secret != nil {
				if err := r.Delete(ctx, &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: plan.Status.Secret.Namespace,
						Name:      plan.Status.Secret.Name,
					},
				}); client.IgnoreNotFound(err) != nil {
					log.Error(err, "failed to remove owned Secret")
					r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", "Failed to remove owned Secret")
					return ctrl.Result{}, err
				} else {
					plan.Status.Secret = nil
				}
			}
			if plan.Status.CronJob != nil {
				if err := r.Delete(ctx, &batchv1beta1.CronJob{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: plan.Status.CronJob.Namespace,
						Name:      plan.Status.CronJob.Name,
					},
				}); client.IgnoreNotFound(err) != nil {
					log.Error(err, "failed to remove owned CronJob")
					r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", "Failed to remove owned CronJob")
					return ctrl.Result{}, err
				} else {
					plan.Status.CronJob = nil
				}
			}
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
	// TODO: if default destination is used, check if additional resources (e.g. secret) should be created

	// First we create or update the Secret before checking the related CronJob
	var secret corev1.Secret
	// If Secret does not exist, let's create a new one
	if plan.Status.Secret != nil {
		err := r.Get(ctx, types.NamespacedName{
			Namespace: plan.Status.Secret.Namespace,
			Name:      plan.Status.Secret.Name,
		}, &secret)
		if client.IgnoreNotFound(err) != nil { // Unexpected error
			r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Checking owned Secret failed with: %v", err))
			return ctrl.Result{}, err
		} else if err != nil {
			// Not found so let's reset the reference and let's re-create it
			plan.Status.Secret = nil
		}
	}
	if plan.Status.Secret == nil { // Checking here as above control flow can reset secret
		secret.ObjectMeta.Name = req.Name // TODO: maybe introduce a hash of content?
		secret.ObjectMeta.Namespace = req.Namespace
		err := controllerutil.SetControllerReference(&plan, &secret, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// Let's compute the content of the Secret
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
	// Finally create or update the Secret
	if plan.Status.Secret != nil {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Updating Secret")
		err = r.Update(ctx, &secret)
	} else {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Creating Secret")
		err = r.Create(ctx, &secret)
	}
	if err != nil {
		log.Error(err, "failed to create or update Secret")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Update or creation of Secret failed with: %v", err))
		return ctrl.Result{}, err
	}
	// Let's make sure to store the reference
	secretRef, err := ref.GetReference(r.Scheme, &secret)
	if err != nil {
		log.Error(err, "failed to get Secret reference")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Failed to get Secret reference: %v", err))
		return ctrl.Result{}, err

	}
	plan.Status.Secret = secretRef

	// Finally create or update the CronJob
	var cronJob batchv1beta1.CronJob
	// If CronJob does not exist, let's create a new one
	if plan.Status.CronJob != nil {
		err := r.Get(ctx, types.NamespacedName{
			Namespace: plan.Status.CronJob.Namespace,
			Name:      plan.Status.CronJob.Name,
		}, &cronJob)
		if client.IgnoreNotFound(err) != nil { // Unexpected error
			r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Checking owned CronJob failed with: %v", err))
			return ctrl.Result{}, err
		} else if err != nil {
			// Not found so let's reset the reference and let's re-create it
			plan.Status.CronJob = nil
		}
	}
	if plan.Status.CronJob == nil { // Checking here as above control flow can reset CronJob
		cronJob.ObjectMeta.Name = req.Name
		cronJob.ObjectMeta.Namespace = req.Namespace
		err := controllerutil.SetControllerReference(&plan, &cronJob, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Properly construct the spec
	err = UpdateCronJobSpec(&cronJob, secretRef,
		plan.Spec.Schedule,
		plan.Spec.ActiveDeadlineSeconds,
		r.WorkerImage,
		plan.Spec.Env,
		"mongodb") // TODO: const?
	if err != nil {
		return ctrl.Result{}, err
	}

	// Finally create or update the cronjob
	if plan.Status.CronJob != nil {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Updating CronJob")
		err = r.Update(ctx, &cronJob)
	} else {
		r.Recorder.Event(&plan, corev1.EventTypeNormal, "Info", "Creating CronJob")
		err = r.Create(ctx, &cronJob)
	}
	if err != nil {
		log.Error(err, "failed to create or update CronJob")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Update or creation of CronJob failed with: %v", err))
		return ctrl.Result{}, err
	}
	// Let's make sure to store the reference
	cronJobRef, err := ref.GetReference(r.Scheme, &cronJob)
	if err != nil {
		log.Error(err, "failed to get CronJob reference")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Failed to get CronJob reference: %v", err))
		return ctrl.Result{}, err

	}
	plan.Status.CronJob = cronJobRef

	if err := r.Update(ctx, &plan); err != nil {
		log.Error(err, "status update failed")
		r.Recorder.Event(&plan, corev1.EventTypeWarning, "Problem", fmt.Sprintf("Failed to update MongoDBBackupPlan: %v", err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MongoDBBackupPlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("mongodbbackupplan-controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1alpha1.MongoDBBackupPlan{}).
		Owns(&corev1.Secret{}).
		Owns(&batchv1beta1.CronJob{}).
		Complete(r)
}
