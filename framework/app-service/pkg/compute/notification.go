package compute

import (
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	"k8s.io/klog/v2"
)

const (
	notificationSubject                  = "os.notification"
	notificationTopicComputeInsufficient = "ApplicationComputeResourceInsufficient"
)

type resourceNotification struct {
	Topic   string                      `json:"topic"`
	Payload resourceNotificationPayload `json:"payload"`
}

type resourceNotificationPayload struct {
	User      string    `json:"user"`
	AppName   string    `json:"appName"`
	Namespace string    `json:"namespace"`
	Resource  string    `json:"resource"`
	Mode      string    `json:"mode"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func PublishComputeInsufficientNotification(appConfig *appcfg.ApplicationConfig, reason error) {
	if appConfig == nil {
		return
	}
	message := fmt.Sprintf("No available compute resource for %s mode.", appConfig.SelectedGpuType)
	if reason != nil {
		message = reason.Error()
	}
	data := resourceNotification{
		Topic: notificationTopicComputeInsufficient,
		Payload: resourceNotificationPayload{
			User:      appConfig.OwnerName,
			AppName:   appConfig.AppName,
			Namespace: appConfig.Namespace,
			Resource:  "compute",
			Mode:      appConfig.SelectedGpuType,
			Reason:    "insufficient-resources",
			Message:   message,
			Timestamp: time.Now(),
		},
	}
	nc, err := utils.NewNatsConn()
	if err != nil {
		klog.Warningf("failed to connect NATS for compute resource notification: %v", err)
		return
	}
	defer nc.Close()
	if err := utils.PublishEvent(nc, notificationSubject, data); err != nil {
		klog.Warningf("failed to publish compute resource notification: %v", err)
	}
}
