package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=".spec.clusterName",description="Cluster name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.state",description="Job status"
// +kubebuilder:printcolumn:name="Completed",type=date,JSONPath=".status.completed",description="Completed time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={pgc-backup}, categories={all}
// PGClusterBackup is the Schema for the PostgreSQL Cluster Backup job
type PGClusterBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PGClusterBackupSpec   `json:"spec,omitempty"`
	Status PGClusterBackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PGClusterBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PGClusterBackup `json:"items"`
}

type PGClusterBackupSpec struct {
	ClusterName string `json:"clusterName"`

	// the volume provided backup file storage
	VolumeSpec *corev1.Volume `json:"volumeSpec"`
}

// PGClusterBackupStatus defines the observed state of PGClusterBackup
type PGClusterBackupStatus struct {
	State       BackupState  `json:"state,omitempty"`
	StartAt     *metav1.Time `json:"start,omitempty"`
	CompletedAt *metav1.Time `json:"completed,omitempty"`
	Error       string       `json:"error,omitempty"`
	BackupPath  string       `json:"backupPath,omitempty"`
}

type BackupState string

const (
	BackupStateNew       BackupState = ""
	BackupStateWaiting   BackupState = "waiting"
	BackupStateRequested BackupState = "requested"
	BackupStateRejected  BackupState = "rejected"
	BackupStateRunning   BackupState = "running"
	BackupStateError     BackupState = "error"
	BackupStateReady     BackupState = "ready"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=".spec.clusterName",description="Cluster name"
// +kubebuilder:printcolumn:name="Backup",type=string,JSONPath=".spec.backupName",description="Backup name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.state",description="Job status"
// +kubebuilder:printcolumn:name="Completed",type=date,JSONPath=".status.completed",description="Completed time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={pgc-restore}, categories={all}
// PGClusterRestore is the Schema for the PostgreSQL Cluster Restore job
type PGClusterRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PGClusterRestoreSpec   `json:"spec,omitempty"`
	Status PGClusterRestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PGClusterRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PGClusterRestore `json:"items"`
}

type PGClusterRestoreSpec struct {
	ClusterName string `json:"clusterName"`
	BackupName  string `json:"backupName"`
}

// PGClusterRstoreStatus defines the observed state of PGClusterRestore
type PGClusterRestoreStatus struct {
	State       RestoreState `json:"state,omitempty"`
	StartAt     *metav1.Time `json:"start,omitempty"`
	CompletedAt *metav1.Time `json:"completed,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// RestoreState is for restore status states
type RestoreState string

const (
	RestoreStateNew       RestoreState = ""
	RestoreStateWaiting   RestoreState = "waiting"
	RestoreStateRequested RestoreState = "requested"
	RestoreStateRejected  RestoreState = "rejected"
	RestoreStateRunning   RestoreState = "running"
	RestoreStateError     RestoreState = "error"
	RestoreStateReady     RestoreState = "ready"
)
