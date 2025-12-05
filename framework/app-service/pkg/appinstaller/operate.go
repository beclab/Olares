package appinstaller

import (
	"bytetrade.io/web3os/app-service/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Status represents the status of an application.
type Status struct {
	Name              string                     `json:"name"`
	AppID             string                     `json:"appID"`
	Namespace         string                     `json:"namespace"`
	CreationTimestamp metav1.Time                `json:"creationTimestamp"`
	Source            string                     `json:"source"`
	AppStatus         v1alpha1.ApplicationStatus `json:"status"`
}

// Operate represents the certain operation for application/recommend/model/agent.
type Operate struct {
	AppName           string                           `json:"appName"`
	AppNamespace      string                           `json:"appNamespace,omitempty"`
	AppOwner          string                           `json:"appOwner,omitempty"`
	State             v1alpha1.ApplicationManagerState `json:"state"`
	OpType            v1alpha1.OpType                  `json:"opType"`
	OpID              string                           `json:"opID"`
	Message           string                           `json:"message"`
	ResourceType      string                           `json:"resourceType"`
	CreationTimestamp metav1.Time                      `json:"creationTimestamp"`
	Source            string                           `json:"source"`
	Progress          string                           `json:"progress,omitempty"`
}

// OperateHistory represents the certain operation history for application/recommend/model/agent.
type OperateHistory struct {
	AppName           string `json:"appName"`
	AppNamespace      string `json:"appNamespace"`
	AppOwner          string `json:"appOwner"`
	ResourceType      string `json:"resourceType"`
	v1alpha1.OpRecord `json:",inline"`
}

type ProviderOperation string

const (
	Register   ProviderOperation = "register"
	Unregister ProviderOperation = "unregister"
)
