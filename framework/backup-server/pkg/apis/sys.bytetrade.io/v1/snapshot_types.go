/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SnapshotSpec defines the desired state of Snapshot.
type SnapshotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	BackupId      string            `json:"backupId"`
	Location      string            `json:"location"`
	SnapshotType  *int              `json:"snapshotType"`
	SnapshotId    *string           `json:"snapshotId,omitempty"`
	Size          *uint64           `json:"size,omitempty"`
	CreateAt      *metav1.Time      `json:"createAt"`
	StartAt       *metav1.Time      `json:"startAt"`
	EndAt         *metav1.Time      `json:"endAt,omitempty"`
	Progress      int               `json:"progress,omitempty"`
	Phase         *string           `json:"phase"`
	Message       *string           `json:"message,omitempty"`
	ResticPhase   *string           `json:"resticPhase,omitempty"`
	ResticMessage *string           `json:"resticMessage,omitempty"`
	Extra         map[string]string `json:"extra,omitempty"`
}

// SnapshotStatus defines the observed state of Snapshot.
type SnapshotStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced, categories={all}
//+kubebuilder:printcolumn:JSONPath=.spec.location, name=location, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.snapshotType, name=snapshotType, type=string
//+kubebuilder:printcolumn:JSONPath=.spec.phase, name=phase, type=string
//+kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=creation, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Snapshot is the Schema for the snapshots API.
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SnapshotList contains a list of Snapshot.
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Snapshot{}, &SnapshotList{})
}
