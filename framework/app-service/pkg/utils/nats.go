package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/beclab/Olares/framework/app-service/api/app.bytetrade.io/v1alpha1"

	"github.com/nats-io/nats.go"
	"k8s.io/klog/v2"
)

type Event struct {
	EventID          string                    `json:"eventID"`
	CreateTime       time.Time                 `json:"createTime"`
	Name             string                    `json:"name"`
	RawAppName       string                    `json:"rawAppName"`
	Type             string                    `json:"type"`
	OpType           string                    `json:"opType,omitempty"`
	OpID             string                    `json:"opID,omitempty"`
	State            string                    `json:"state"`
	Progress         string                    `json:"progress,omitempty"`
	User             string                    `json:"user"`
	EntranceStatuses []v1alpha1.EntranceStatus `json:"entranceStatuses,omitempty"`
	Title            string                    `json:"title,omitempty"`
	Reason           string                    `json:"reason,omitempty"`
	Message          string                    `json:"message,omitempty"`
	SharedEntrances  []v1alpha1.Entrance       `json:"sharedEntrances,omitempty"`
}

// EventParams defines parameters to publish an app-related event
type EventParams struct {
	Owner            string
	Name             string
	OpType           string
	OpID             string
	State            string
	Progress         string
	EntranceStatuses []v1alpha1.EntranceStatus
	RawAppName       string
	Type             string // "app" (default) or "middleware"
	Title            string
	Reason           string
	Message          string
	SharedEntrances  []v1alpha1.Entrance
}

func PublishEvent(nc *nats.Conn, subject string, data interface{}) error {
	return publish(nc, subject, data)
}

func publish(nc *nats.Conn, subject string, data interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = nc.Publish(subject, d)
	if err != nil {
		klog.Infof("publish err=%v", err)
		return err
	}
	return nil
}

func NewNatsConn() (*nats.Conn, error) {
	natsHost := os.Getenv("NATS_HOST")
	natsPort := os.Getenv("NATS_PORT")
	username := os.Getenv("NATS_USERNAME")
	password := os.Getenv("NATS_PASSWORD")

	natsURL := fmt.Sprintf("nats://%s:%s", natsHost, natsPort)

	opts := []nats.Option{
		nats.UserInfo(username, password),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				klog.Warningf("NATS disconnected: %v, will attempt to reconnect", err)
			} else {
				klog.Infof("NATS disconnected, will attempt to reconnect")
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			klog.Infof("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			klog.Errorf("NATS connection closed permanently: %v", nc.LastError())
		}),
	}

	nc, err := nats.Connect(natsURL, opts...)
	if err != nil {
		klog.Errorf("failed to connect to NATS: %v", err)
		return nil, err
	}
	klog.Infof("connected to NATS at %s", natsURL)
	return nc, nil
}
