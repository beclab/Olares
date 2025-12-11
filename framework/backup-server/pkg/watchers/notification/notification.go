package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
)

var UserSchemeGroupVersionResource = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}

var DataSender *dataSender

type EventPayload struct {
	Type string      `json:"eventType"`
	Data interface{} `json:"eventData,omitempty"`
}

type config struct {
	Host     string
	Port     string
	Username string
	Password string
	Subject  string
}

type dataSender struct {
	conn    *nats.Conn
	subject string
	enabled bool
}

type Notification struct {
	Factory client.Factory
}

func NewSender() {
	DataSender = new(dataSender)

	var config = &config{
		Host:     util.GetEnvOrDefault("NATS_HOST", "localhost"),
		Port:     util.GetEnvOrDefault("NATS_PORT", "4222"),
		Username: util.GetEnvOrDefault("NATS_USERNAME", ""),
		Password: util.GetEnvOrDefault("NATS_PASSWORD", ""),
		Subject:  util.GetEnvOrDefault("NATS_SUBJECT_SYSTEM_BACKUP_STATE", "os.backup"),
	}

	natsURL := fmt.Sprintf("nats://%s:%s@%s:%s",
		config.Username, config.Password, config.Host, config.Port)

	conn, err := nats.Connect(natsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1), // -1 means unlimited reconnection attempts
		nats.ReconnectJitter(100*time.Millisecond, 1*time.Second),
		nats.Timeout(10*time.Second),
		nats.PingInterval(30*time.Second),
		nats.MaxPingsOutstanding(5),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Errorf("NATS disconnected: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Errorf("NATS reconnected to %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Errorf("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Errorf("NATS error: %v", err)
		}),
	)

	if err != nil {
		log.Errorf("failed to connect to NATS: %v, url: %s", err, natsURL)
		panic(err)
	}

	log.Infof("Connected to NATS server at %s:%s", config.Host, config.Port)

	DataSender.conn = conn
	DataSender.subject = config.Subject
	DataSender.enabled = true
}

func (d *dataSender) Send(user string, data interface{}) {
	if d == nil || d.conn == nil {
		log.Errorf("NATS data sender is disabled, skipping message send")
		return
	}

	var subject = fmt.Sprintf("%s.%s", d.subject, user)
	var msg, err = json.Marshal(data)
	if err != nil {
		log.Errorf("NATS data marshal error: %v", err)
		return
	}

	log.Infof("NATS send data: %s", string(msg))

	if err = d.conn.Publish(subject, msg); err != nil {
		log.Errorf("NATS publish msg error: %v", err)
	} else {
		log.Infof("NATS publish succeed, subject: %s", subject)
	}

}

func (n *Notification) Send(ctx context.Context, eventType, user, msg string, data interface{}) error {
	appKey, appSecret, err := n.getUserAppKey(ctx, user)
	if err != nil {
		return err
	}

	client := NewEventClient(appKey, appSecret, "system-server.user-system-"+user)

	return client.CreateEvent(eventType, msg, data)
}

func (n *Notification) getUserAppKey(ctx context.Context, user string) (appKey, appSecret string, err error) {
	dynamicClient, err := n.Factory.DynamicClient()
	if err != nil {
		return
	}
	namespace := "user-system-" + user
	data, err := dynamicClient.Resource(AppPermGVR).Namespace(namespace).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		log.Errorf("get bfl application permission error: %v", err)
		return
	}

	var appPerm ApplicationPermission

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(data.Object, &appPerm)
	if err != nil {
		log.Errorf("convert bfl application permission error: %v ", err)
		return
	}

	appKey = appPerm.Spec.Key
	appSecret = appPerm.Spec.Secret

	return
}
