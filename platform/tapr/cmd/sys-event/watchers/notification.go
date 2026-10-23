package watchers

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
)

var UserSchemeGroupVersionResource = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}

type EventPayload struct {
	Type string      `json:"eventType"`
	Data interface{} `json:"eventData,omitempty"`
}

type Notification struct {
	DynamicClient *dynamic.DynamicClient
}

func (n *Notification) Send(ctx context.Context, user, msg string, data interface{}) error {
	// Skip sending for now. A missing bfl ApplicationPermission made Send fail,
	// the workqueue requeued the item, and HTTP callbacks (e.g. /metrics/highload)
	// were invoked again even after the original condition had recovered.
	// TODO: remove unused notification send path (getUserAppKey and related code).
	return nil
}

func (n *Notification) AdminUser(ctx context.Context) (string, error) {
	list, err := n.DynamicClient.Resource(UserSchemeGroupVersionResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error("list user error, ", err)
		return "", err
	}

	for _, user := range list.Items {
		if role, ok := user.GetAnnotations()["bytetrade.io/owner-role"]; ok {
			if role == "owner" {
				return user.GetName(), nil
			}
		}
	}

	return "", errors.New("admin not found")
}
