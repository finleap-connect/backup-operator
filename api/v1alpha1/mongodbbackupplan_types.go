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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MongoDBBackupPlanKind = "MongoDBBackupPlan"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongoDBBackupPlanSpec defines the desired state of MongoDBBackupPlan
type MongoDBBackupPlanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Schedule in cron format
	Schedule string `json:"schedule"`

	// +kubebuilder:validation:Minimum=1
	//
	ActiveDeadlineSeconds int64 `json:"activeDeadlineSeconds"`

	// +kubebuilder:validation:Minimum=1
	// Number of backups to keep
	Retention int64 `json:"retention"`

	// +optional
	// Environments for the CronJob
	Env []corev1.EnvVar `json:"env,omitempty"`

	// Fully qualifying MongoDB URI connection string. Environment variables
	// will be evaluated before usage.
	URI string `json:"uri"`

	// +optional
	// Setup for metrics
	Pushgateway *Pushgateway `json:"pushgateway,omitempty"`

	// +optional
	// Destination for the backup. If none is provided the default destination
	// will be tried.
	Destination *Destination `json:"destination,omitempty"`
}

// MongoDBBackupPlanStatus defines the observed state of MongoDBBackupPlan
type MongoDBBackupPlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CronJob *corev1.ObjectReference `json:"cronJob,omitempty"`
	Secret  *corev1.ObjectReference `json:"secret,omitempty"`
}

// +kubebuilder:object:root=true

// MongoDBBackupPlan is the Schema for the mongodbbackupplans API
type MongoDBBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBBackupPlanSpec   `json:"spec,omitempty"`
	Status MongoDBBackupPlanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MongoDBBackupPlanList contains a list of MongoDBBackupPlan
type MongoDBBackupPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDBBackupPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDBBackupPlan{}, &MongoDBBackupPlanList{})
}
