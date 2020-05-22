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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ConsulBackupPlanKind = "ConsulBackupPlan"
const ConsulBackupPlanWorkerCommand = "consul"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulBackupPlanSpec defines the desired state of ConsulBackupPlan
type ConsulBackupPlanSpec struct {
	// Address of Consul. Environment variables
	// will be evaluated before usage.
	Address string `json:"address"`

	// +optional
	// Username to authenticate with consul
	Username string `json:"username,omitempty"`

	// +optional
	// Password to authenticate with consul
	Password string `json:"password,omitempty"`
}

// +kubebuilder:object:root=true

// ConsulBackupPlan is the Schema for the consulbackupplans API
type ConsulBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupPlanSpec   `json:"spec,omitempty"`
	Status BackupPlanStatus `json:"status,omitempty"`

	ConsulSpec ConsulBackupPlanSpec `json:"specConsul,omitempty"`
}

func (p *ConsulBackupPlan) GetTypeMeta() *metav1.TypeMeta {
	return &p.TypeMeta
}

func (p *ConsulBackupPlan) GetObjectMeta() *metav1.ObjectMeta {
	return &p.ObjectMeta
}

func (p *ConsulBackupPlan) GetSpec() *BackupPlanSpec {
	return &p.Spec
}

func (p *ConsulBackupPlan) GetStatus() *BackupPlanStatus {
	return &p.Status
}

func (p *ConsulBackupPlan) GetKind() string {
	return ConsulBackupPlanKind
}

func (p *ConsulBackupPlan) GetCmd() string {
	return ConsulBackupPlanWorkerCommand
}

func (p *ConsulBackupPlan) New() BackupPlan {
	return &ConsulBackupPlan{}
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
