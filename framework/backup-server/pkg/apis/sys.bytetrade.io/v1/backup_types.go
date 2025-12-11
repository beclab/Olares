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

type BackupPolicy struct {
	Enabled           bool   `json:"enabled"`
	SnapshotFrequency string `json:"snapshotFrequency"`
	TimesOfDay        string `json:"timesOfDay"`
	TimespanOfDay     string `json:"timespanOfDay"`
	DayOfWeek         int    `json:"dayOfWeek"`
	DateOfMonth       int    `json:"dateOfMonth"`
}

// BackupSpec defines the desired state of Backup.
type BackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name         string            `json:"name"`
	Owner        string            `json:"owner"`
	Location     map[string]string `json:"location"`
	BackupPolicy *BackupPolicy     `json:"backupPolicy,omitempty"`
	BackupType   map[string]string `json:"backupType"`
	Size         *uint64           `json:"size,omitempty"`
	RestoreSize  *uint64           `json:"restoreSize,omitempty"`
	CreateAt     *metav1.Time      `json:"createAt"`
	Notified     bool              `json:"notified"`
	Deleted      bool              `json:"deleted"`
	Extra        map[string]string `json:"extra,omitempty"`
}

// BackupStatus defines the observed state of Backup.
type BackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	State      string      `json:"state"`
	UpdateTime metav1.Time `json:"updateTime"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={bc}, categories={all}
// +kubebuilder:printcolumn:JSONPath=.spec.name, name=backup name, type=string
// +kubebuilder:printcolumn:JSONPath=.spec.owner, name=owner, type=string
// +kubebuilder:printcolumn:JSONPath=.spec.deleted, name=deleted, type=boolean
// +kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp, name=creation, type=date
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Backup is the Schema for the backups API.
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupList contains a list of Backup.
type BackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Backup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Backup{}, &BackupList{})
}
