package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +kubebuilder:printcolumn:name="type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced, shortName={rdxc}, categories={all}
// RedixCluster is the Schema for the Redis-Compatible Cluster
type RedixCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedixClusterSpec   `json:"spec,omitempty"`
	Status RedixClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedixClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []RedixCluster `json:"items"`
}

type RedixClusterSpec struct {
	Type    RedixType    `json:"type"`
	KVRocks *KVRocksSpec `json:"kvrocks,omitempty"`
}

type RedixClusterStatus struct {
	// the state of the application: draft, submitted, passed, rejected, suspended, active
	State      string       `json:"state"`
	UpdateTime *metav1.Time `json:"updateTime,omitempty"`
	StatusTime *metav1.Time `json:"statusTime,omitempty"`
}

type KVRocksSpec struct {
	Password        PasswordVar                  `json:"password,omitempty"`
	Owner           string                       `json:"owner"`
	BackupStorage   *string                      `json:"backupStorage,omitempty"`
	Image           string                       `json:"image,omitempty"`
	ImagePullPolicy corev1.PullPolicy            `json:"imagePullPolicy,omitempty"`
	KVRocksConfig   map[string]string            `json:"kvrocksConfig,omitempty"`
	Resources       *corev1.ResourceRequirements `json:"resources,omitempty"`
}

type RedixType string

const (
	RedisCluster   RedixType = "redis-cluster"
	RedisServer    RedixType = "redis-server"
	KVRocks        RedixType = "kvrocks"
	KVRocksCluster RedixType = "kvrocks-cluster"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=".spec.clusterName",description="Cluster name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.state",description="Job status"
// +kubebuilder:printcolumn:name="Completed",type=date,JSONPath=".status.completed",description="Completed time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={kvr-backup}, categories={all}
// KVRocksBackup is the Schema for the KVRocks Backup job
type KVRocksBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KVRocksBackupSpec   `json:"spec,omitempty"`
	Status KVRocksBackupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KVRocksBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KVRocksBackup `json:"items"`
}

type KVRocksBackupSpec struct {
	ClusterName string `json:"clusterName"`
}

// KVRocksBackupStatus defines the observed state of KVRocksBackup
type KVRocksBackupStatus struct {
	State       BackupState  `json:"state,omitempty"`
	StartAt     *metav1.Time `json:"start,omitempty"`
	CompletedAt *metav1.Time `json:"completed,omitempty"`
	Error       string       `json:"error,omitempty"`
	BackupPath  string       `json:"backupPath,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Cluster",type=string,JSONPath=".spec.clusterName",description="Cluster name"
// +kubebuilder:printcolumn:name="Backup",type=string,JSONPath=".spec.backupName",description="Backup name"
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=".status.state",description="Job status"
// +kubebuilder:printcolumn:name="Completed",type=date,JSONPath=".status.completed",description="Completed time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=".metadata.creationTimestamp",description="Created time"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced, shortName={kvr-restore}, categories={all}
// KVRocksRestore is the Schema for the KVRocks Restore job
type KVRocksRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KVRocksRestoreSpec   `json:"spec,omitempty"`
	Status KVRocksRestoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KVRocksRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KVRocksRestore `json:"items"`
}

type KVRocksRestoreSpec struct {
	ClusterName   string `json:"clusterName"`
	BackupStorage string `json:"backupStorage"`
}

// KVRocksRstoreStatus defines the observed state of KVRocksRestore
type KVRocksRestoreStatus struct {
	State       RestoreState `json:"state,omitempty"`
	StartAt     *metav1.Time `json:"start,omitempty"`
	CompletedAt *metav1.Time `json:"completed,omitempty"`
	Error       string       `json:"error,omitempty"`
}
