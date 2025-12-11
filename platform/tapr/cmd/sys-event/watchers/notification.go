package watchers

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	appKey, appSecret, err := n.getUserAppKey(ctx, user)
	if err != nil {
		return err
	}

	client := NewEventClient(appKey, appSecret, "system-server.user-system-"+user)

	return client.CreateEvent("notification", msg, data)
}

func (n *Notification) getUserAppKey(ctx context.Context, user string) (appKey, appSecret string, err error) {
	namespace := "user-system-" + user
	data, err := n.DynamicClient.Resource(AppPermGVR).Namespace(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		klog.Error("get bfl application permission error, ", err, ", ", namespace)
		return
	}

	var appPerm ApplicationPermission

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &appPerm)
	if err != nil {
		klog.Error("convert bfl application permission error, ", err, ", ", namespace)
		return
	}

	appKey = appPerm.Spec.Key
	appSecret = appPerm.Spec.Secret

	return
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
