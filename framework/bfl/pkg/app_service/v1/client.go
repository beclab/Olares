package app_service

import (
	"fmt"
	"net/http"
	"os"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

type Client struct {
	httpClient *http.Client
	username   string
}

const (
	AppServiceUserAppListURLTempl       = "http://%s:%s/app-service/v1/user-apps/%s"
	AppServiceAppSetupURLTempl          = "http://%s:%s/app-service/v1/applications/%s/setup"
	AppServiceAppEntrancesURLTempl      = "http://%s:%s/app-service/v1/applications/%s/entrances"
	AppServiceAppEntranceSetupURLTempl  = "http://%s:%s/app-service/v1/applications/%s/%s/setup"
	AppServiceAppEntranceAuthURLTempl   = "http://%s:%s/app-service/v1/applications/%s/%s/auth-level"
	AppServiceAppEntrancePolicyURLTempl = "http://%s:%s/app-service/v1/applications/%s/%s/policy"
	AppServiceUserMetricsURLTempl       = "http://%s:%s/app-service/v1/users/%s/metrics"
	AppServiceAppSubjectListTempl       = "http://%s:%s/app-service/v1/apps/%s/subject"

	AppServiceAppPermissionTempl = "http://%s:%s/app-service/v1/perms/%s"

	AppServiceHostEnv = "APP_SERVICE_SERVICE_HOST"
	AppServicePortEnv = "APP_SERVICE_SERVICE_PORT"
)

var appServiceClient *Client

func init() {
	transport := &http.Transport{
		MaxIdleConnsPerHost: 50,
	}

	appServiceClient = &Client{
		httpClient: &http.Client{Timeout: time.Second * 5,
			Transport: transport},
	}
}

func NewAppServiceClient() *Client {
	return appServiceClient
}

func (c *Client) ForUser(username string) *Client {
	client := *c
	client.username = username
	return &client
}

func (c *Client) url(templ string, a ...any) string {
	appServiceHost := os.Getenv(AppServiceHostEnv)
	appServicePort := os.Getenv(AppServicePortEnv)
	if appServiceHost == "" {
		appServiceHost = "app-service.os-framework"
	}
	if appServicePort == "" {
		appServicePort = "6755"
	}
	return fmt.Sprintf(templ, append([]any{appServiceHost, appServicePort}, a...)...)
}

func (c *Client) ListAppInfosByUser(user string) ([]*AppInfo, error) {
	c = c.ForUser(user)
	app, err := c.FetchUserAppList(user)
	if err != nil {
		return nil, err
	}

	return c.getAppListFromData(app)
}

func (c *Client) FetchUserAppList(user string) ([]appv1.Application, error) {
	return c.doHttpGetApplicationList(c.url(AppServiceUserAppListURLTempl, user), "")
}

func (c *Client) SetupAppPolicy(app, token string, settings ApplicationsSettings) (map[string]interface{}, error) {
	return c.doHttpPost(c.url(AppServiceAppSetupURLTempl, app), token, settings)
}

func (c *Client) SetupAppEntrancePolicy(app, entranceName, token string, settings ApplicationsSettings) (map[string]interface{}, error) {
	return c.doHttpPost(c.url(AppServiceAppEntrancePolicyURLTempl, app, entranceName), token, settings)
}

func (c *Client) GetAppPolicy(app, token string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceAppSetupURLTempl, app), token)
}

func (c *Client) SetupAppCustomDomain(app, entranceName, token string, settings ApplicationsSettings) (map[string]interface{}, error) {
	return c.doHttpPost(c.url(AppServiceAppEntranceSetupURLTempl, app, entranceName), token, settings)
}

func (c *Client) GetAppCustomDomain(app, token string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceAppSetupURLTempl, app), token)
}

func (c *Client) GetAppEntrances(app, token string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceAppEntrancesURLTempl, app), token)
}

func (c *Client) SetupAppAuthorizationLevel(app, entranceName, token string, settings ApplicationsSettings) (map[string]interface{}, error) {
	return c.doHttpPost(c.url(AppServiceAppEntranceAuthURLTempl, app, entranceName), token, settings)
}

func (c *Client) GetUserMetrics(user, token string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceUserMetricsURLTempl, user), token)
}

func (c *Client) GetApplicationPermission(token, app string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceAppPermissionTempl, app), token)
}

func (c *Client) GetApplicationSubjectList(appName, token string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceAppSubjectListTempl, appName), token)
}
