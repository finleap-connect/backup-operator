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

const ConsulBackupPlanKind = "ConsulDBBackupPlan"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulBackupPlanSpec defines the desired state of ConsulBackupPlan
type ConsulBackupPlanSpec struct {
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

	// Address of Consul. Environment variables
	// will be evaluated before usage.
	Address string `json:"address"`

	// +optional
	// Username to authenticate with consul
	Username string `json:"username,omitempty"`

	// +optional
	// Password to authenticate with consul
	Password string `json:"password,omitempty"`

	// +optional
	// Setup for metrics
	Pushgateway *Pushgateway `json:"pushgateway,omitempty"`

	// +optional
	// Destination for the backup. If none is provided the default destination
	// will be tried.
	Destination *Destination `json:"destination,omitempty"`
}

// ConsulBackupPlanStatus defines the observed state of ConsulBackupPlan
type ConsulBackupPlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	CronJob *corev1.ObjectReference `json:"cronJob,omitempty"`
	Secret  *corev1.ObjectReference `json:"secret,omitempty"`
}

// +kubebuilder:object:root=true

// ConsulBackupPlan is the Schema for the consulbackupplans API
type ConsulBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulBackupPlanSpec   `json:"spec,omitempty"`
	Status ConsulBackupPlanStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConsulBackupPlanList contains a list of ConsulBackupPlan
type ConsulBackupPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsulBackupPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsulBackupPlan{}, &ConsulBackupPlanList{})
}
