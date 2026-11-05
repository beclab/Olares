package compute

import (
	"fmt"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	appevent "github.com/beclab/Olares/framework/app-service/pkg/event"
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

// BuildInsufficientPayload constructs the (subject, payload) pair
// sent when compute allocation for an app fails. Extracted from
// PublishComputeInsufficientNotification so
// pkg/applifecycle/adapters/computevalidator can produce the exact
// same NATS wire shape without depending on the AppEventQueue
// global — publishing is done by the caller (either
// PublishComputeInsufficientNotification below, or the adapter via
// its injected notifier seam).
//
// Returns the notificationSubject constant and a payload built from
// appConfig; a nil appConfig yields an empty payload so the caller
// can decide whether to publish anything at all.
func BuildInsufficientPayload(appConfig *appcfg.ApplicationConfig, reason error) (subject string, payload any) {
	if appConfig == nil {
		return notificationSubject, nil
	}
	message := fmt.Sprintf("No available compute resource for %s mode.", appConfig.SelectedGpuType)
	if reason != nil {
		message = reason.Error()
	}
	return notificationSubject, resourceNotification{
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
}

func PublishComputeInsufficientNotification(appConfig *appcfg.ApplicationConfig, reason error) {
	if appConfig == nil {
		return
	}
	subject, data := BuildInsufficientPayload(appConfig, reason)
	appevent.PublishToQueue(subject, data)
}
