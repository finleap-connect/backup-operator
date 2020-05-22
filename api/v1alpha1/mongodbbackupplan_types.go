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

const MongoDBBackupPlanKind = "MongoDBBackupPlan"
const MongoDBBackupPlanWorkerCommand = "mongodb"

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MongoDBBackupPlanSpec defines the desired state of MongoDBBackupPlan
type MongoDBBackupPlanSpec struct {
	// Fully qualifying MongoDB URI connection string. Environment variables
	// will be evaluated before usage.
	URI string `json:"uri"`
}

// +kubebuilder:object:root=true

// MongoDBBackupPlan is the Schema for the mongodbbackupplans API
type MongoDBBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupPlanSpec   `json:"spec,omitempty"`
	Status BackupPlanStatus `json:"status,omitempty"`

	MongoDbSpec MongoDBBackupPlanSpec `json:"specMongoDb,omitempty"`
}

func (p *MongoDBBackupPlan) GetTypeMeta() *metav1.TypeMeta {
	return &p.TypeMeta
}

func (p *MongoDBBackupPlan) GetObjectMeta() *metav1.ObjectMeta {
	return &p.ObjectMeta
}

func (p *MongoDBBackupPlan) GetSpec() *BackupPlanSpec {
	return &p.Spec
}

func (p *MongoDBBackupPlan) GetStatus() *BackupPlanStatus {
	return &p.Status
}

func (p *MongoDBBackupPlan) GetKind() string {
	return MongoDBBackupPlanKind
}

func (p *MongoDBBackupPlan) GetCmd() string {
	return MongoDBBackupPlanWorkerCommand
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
