package app_service

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

type Client struct {
	httpClient *http.Client
}

const (
	AppServiceGetURLTempl                  = "http://%s:%s/app-service/v1/apps/%s"
	AppServiceListURLTempl                 = "http://%s:%s/app-service/v1/apps"
	AppServiceAllListURLTempl              = "http://%s:%s/app-service/v1/all/apps"
	AppServiceUserAppListURLTempl          = "http://%s:%s/app-service/v1/user-apps/%s"
	AppServiceRegistryListURLTempl         = "http://%s:%s/app-service/v1/registry/applications"
	AppServiceAppDetailURLTempl            = "http://%s:%s/app-service/v1/registry/applications/%s"
	AppServiceInstallURLTempl              = "http://%s:%s/app-service/v1/apps/%s/install"
	AppServiceUpgradeAppURLTempl           = "http://%s:%s/app-service/v1/apps/%s/upgrade"
	AppServiceUninstallURLTempl            = "http://%s:%s/app-service/v1/apps/%s/uninstall"
	AppServiceInstallStatusURLTempl        = "http://%s:%s/app-service/v1/apps/%s/operate"
	AppServiceCancelInstallURLTempl        = "http://%s:%s/app-service/v1/apps/%s/cancel"
	AppServiceUserAppsInstallURLTempl      = "http://%s:%s/app-service/v1/users/apps/create/%s"
	AppServiceUserAppsUninstallURLTempl    = "http://%s:%s/app-service/v1/users/apps/delete/%s"
	AppServiceUserAppsStatusURLTempl       = "http://%s:%s/app-service/v1/users/apps/%s"
	AppServiceSystemServiceEnableURLTempl  = "http://%s:%s/app-service/v1/system/service/enable/%s"
	AppServiceSystemServiceDisableURLTempl = "http://%s:%s/app-service/v1/system/service/disable/%s"
	AppServiceAppSetupURLTempl             = "http://%s:%s/app-service/v1/applications/%s/setup"
	AppServiceAppEntrancesURLTempl         = "http://%s:%s/app-service/v1/applications/%s/entrances"
	AppServiceAppEntranceSetupURLTempl     = "http://%s:%s/app-service/v1/applications/%s/%s/setup"
	AppServiceAppEntranceAuthURLTempl      = "http://%s:%s/app-service/v1/applications/%s/%s/auth-level"
	AppServiceAppEntrancePolicyURLTempl    = "http://%s:%s/app-service/v1/applications/%s/%s/policy"
	AppServiceUpgradeNewVersionURLTempl    = "http://%s:%s/app-service/v1/upgrade/newversion"
	AppServiceUpgradeStateURLTempl         = "http://%s:%s/app-service/v1/upgrade/state"
	AppServiceUpgradeURLTempl              = "http://%s:%s/app-service/v1/upgrade"
	AppServiceUpgradeCancelURLTempl        = "http://%s:%s/app-service/v1/upgrade/cancel"
	AppServiceUserMetricsURLTempl          = "http://%s:%s/app-service/v1/users/%s/metrics"
	AppServiceAppInstallationRunningList   = "http://%s:%s/app-service/v1/apps/pending-installing/task"
	AppServiceAppPermissionListTempl       = "http://%s:%s/app-service/v1/perms"
	AppServiceAppProviderRegistryListTempl = "http://%s:%s/app-service/v1/apps/provider-registry/%s"
	AppServiceAppSubjectListTempl          = "http://%s:%s/app-service/v1/apps/%s/subject"

	AppServiceAppPermissionTempl   = "http://%s:%s/app-service/v1/perms/%s"
	AppServiceProvideRegistryTempl = "http://%s:%s/app-service/v1/perms/provider-registry/%s/%s/%s"

	AppServiceAppSuspendURLTempl = "http://%s:%s/app-service/v1/apps/%s/suspend"
	AppServiceAppResumeURLTempl  = "http://%s:%s/app-service/v1/apps/%s/resume"

	AppServiceEnableGpuManagedMemoryURLTempl  = "http://%s:%s/app-service/v1/gpu/enable/managed-memory"
	AppServiceDisableGpuManagedMemoryURLTempl = "http://%s:%s/app-service/v1/gpu/disable/managed-memory"
	AppServiceGetGpuManagedMemoryURLTempl     = "http://%s:%s/app-service/v1/gpu/managed-memory"

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

func (c *Client) url(templ string, a ...any) string {
	appServiceHost := os.Getenv(AppServiceHostEnv)
	appServicePort := os.Getenv(AppServicePortEnv)
	return fmt.Sprintf(templ, append([]any{appServiceHost, appServicePort}, a...)...)
}

func (c *Client) ListAppInfosByUser(user string) ([]*AppInfo, error) {
	app, err := c.FetchUserAppList(user)
	if err != nil {
		return nil, err
	}

	return c.getAppListFromData(app)
}

func (c *Client) FetchUserAppList(user string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceUserAppListURLTempl, user), "")
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

func (c *Client) GetAppAuthorizationLevel(app, token string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceAppSetupURLTempl, app), token)
}

func (c *Client) GetUserMetrics(user, token string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceUserMetricsURLTempl, user), token)
}

func (c *Client) GetApplicationPermissionList(token string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceAppPermissionListTempl), token)
}

func (c *Client) GetApplicationProviderList(appName, token string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceAppProviderRegistryListTempl, appName), token)
}

func (c *Client) GetApplicationPermission(token, app string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceAppPermissionTempl, app), token)
}

func (c *Client) GetProviderRegistry(token, dataType, group, version string) (map[string]interface{}, error) {
	return c.doHttpGetOne(c.url(AppServiceProvideRegistryTempl, dataType, group, version), token)
}

func (c *Client) GetApplicationSubjectList(appName, token string) ([]map[string]interface{}, error) {
	return c.doHttpGetList(c.url(AppServiceAppSubjectListTempl, appName), token)
}
