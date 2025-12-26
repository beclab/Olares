package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
	"github.com/go-resty/resty/v2"
	"golang.org/x/crypto/bcrypt"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
)

const (
	GroupID           = "message-disptahcer.system-server"
	EventVersion      = "v1"
	AccessTokenHeader = "X-Access-Token"
)

type EventClient struct {
	httpClient  *resty.Client
	eventServer string
	appKey      string
	appSecret   string
}

func NewEventClient(key, secret, server string) *EventClient {
	c := resty.New()

	return &EventClient{
		httpClient:  c.SetTimeout(2 * time.Second),
		appKey:      key,
		appSecret:   secret,
		eventServer: server,
	}
}

func (c *EventClient) getAccessToken() (string, error) {
	url := fmt.Sprintf("http://%s/permission/v1alpha1/access", c.eventServer)
	now := time.Now().UnixMilli() / 1000

	password := c.appKey + strconv.Itoa(int(now)) + c.appSecret
	encode, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	perm := AccessTokenRequest{
		AppKey:    c.appKey,
		Timestamp: now,
		Token:     string(encode),
		Perm: PermissionRequire{
			Group:    GroupID,
			Version:  EventVersion,
			DataType: "event",
			Ops: []string{
				"Create",
			},
		},
	}

	postData, err := json.Marshal(perm)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.R().
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetBody(postData).
		SetResult(&AccessTokenResp{}).
		Post(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return "", errors.New(string(resp.Body()))
	}

	token := resp.Result().(*AccessTokenResp)

	if token.Code != 0 {
		return "", errors.New(token.Message)
	}

	return token.Data.AccessToken, nil
}

func (c *EventClient) CreateEvent(eventType string, msg string, data interface{}) error {
	url := fmt.Sprintf("http://%s/system-server/v1alpha1/event/message-disptahcer.system-server/v1", c.eventServer)
	token, err := c.getAccessToken()
	if err != nil {
		return err
	}

	event := Event{
		Type:    eventType,
		Version: EventVersion,
		Data: EventData{
			Message: msg,
			Payload: data,
		},
	}

	postData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.R().
		SetHeaders(map[string]string{
			restful.HEADER_ContentType: restful.MIME_JSON,
			AccessTokenHeader:          token,
		}).
		SetResult(&Response{}).
		SetBody(postData).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return errors.New(string(resp.Body()))
	}

	responseData := resp.Result().(*Response)

	if responseData.Code != 0 {
		return errors.New(responseData.Message)
	}

	log.Infof("send event success, type: %s, message: %s, result: %s", eventType, string(postData), util.ToJSON(responseData.Data))

	return nil
}
