package backup

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// BackupSpec defines the desired state of Backup
type BackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Owner *string `json:"owner"`

	TerminusVersion *string `json:"terminusVersion"`

	Size *int64 `json:"size,omitempty"`

	Phase *string `json:"phase"`

	FailedMessage *string `json:"failedMessage,omitempty"`

	MiddleWarePhase *string `json:"middleWarePhase,omitempty"`

	MiddleWareFailedMessage *string `json:"middleWareFailedMessage,omitempty"`

	ResticPhase *string `json:"resticPhase,omitempty"`

	ResticFailedMessage *string `json:"resticFailedMessage,omitempty"`

	Extra map[string]string `json:"extra,omitempty"`
}

// BackupStatus defines the observed state of Backup
type BackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// Backup is the Schema for the backups API
type Backup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupSpec   `json:"spec,omitempty"`
	Status BackupStatus `json:"status,omitempty"`
}

var GroupVersion = schema.GroupVersion{Group: "sys.bytetrade.io", Version: "v1"}
var BackupGVR = schema.GroupVersionResource{
	Group:    GroupVersion.Group,
	Version:  GroupVersion.Version,
	Resource: "backups",
}

var (
	// backup phase
	BackupStart   string = "Started"
	BackupNew     string = "Pending"
	BackupRunning string = "Running"
	BackupSucceed string = "Succeed"
	BackupFailed  string = "Failed"
	BackupCancel  string = "Canceled"

	FinalizingPartiallyFailed string = "FinalizingPartiallyFailed"
	PartiallyFailed           string = "PartiallyFailed"
	FailedValidation          string = "FailedValidation"
	VeleroBackupCompleted     string = "Completed"
)
