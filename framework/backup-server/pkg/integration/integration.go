package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"olares.com/backup-server/pkg/client"
	"olares.com/backup-server/pkg/constant"
	"olares.com/backup-server/pkg/util"
	"olares.com/backup-server/pkg/util/log"
	"olares.com/backup-server/pkg/util/repo"
)

var IntegrationService *Integration

type Integration struct {
	Factory      client.Factory
	Location     string
	Name         string
	rest         *resty.Client
	OlaresTokens map[string]*SpaceToken
	authToken    map[string]*authToken
}

type authToken struct {
	token  string
	expire time.Time
}

func NewIntegrationManager(factory client.Factory) {
	var debug = false
	d := os.Getenv(constant.EnvIntegrationDebug)
	if d == "1" {
		debug = true
	}

	IntegrationService = &Integration{
		Factory:      factory,
		rest:         resty.New().SetTimeout(20 * time.Second).SetDebug(debug),
		OlaresTokens: make(map[string]*SpaceToken),
		authToken:    make(map[string]*authToken),
	}
}

func IntegrationManager() *Integration {
	return IntegrationService
}

func (i *Integration) GetIntegrationSpaceToken(ctx context.Context, owner string, integrationName string) (*SpaceToken, error) {
	token := i.OlaresTokens[integrationName]
	if token != nil {
		if !token.Expired() {
			return token, nil
		}
	}

	data, err := i.query(ctx, owner, constant.BackupLocationSpace.String(), integrationName)
	if err != nil {
		return nil, err
	}

	token = i.withSpaceToken(data)
	i.OlaresTokens[integrationName] = token

	if token.Expired() {
		log.Errorf("olares space token expired, expiresAt: %d, format: %s", token.ExpiresAt, util.ParseUnixMilliToDate(token.ExpiresAt))
		return nil, errors.New("Access token expired. Please re-connect to your Olares Space in LarePass.")
	}

	return token, nil
}

func (i *Integration) GetIntegrationCloudToken(ctx context.Context, owner, location, integrationName string) (*IntegrationToken, error) {
	data, err := i.query(ctx, owner, location, integrationName)
	if err != nil {
		return nil, err
	}
	return i.withCloudToken(data), nil
}

// todo
func (i *Integration) GetIntegrationCloudAccount(ctx context.Context, owner, location, endpoint string) (*IntegrationToken, error) {
	accounts, err := i.queryIntegrationAccounts(ctx, owner)
	if err != nil {
		return nil, err
	}

	var token *IntegrationToken
	for _, account := range accounts {
		if account.Type == "space" {
			continue
		}
		if strings.Contains(location, account.Type) {
			token, err = i.GetIntegrationCloudToken(ctx, owner, location, account.Name)
			if err != nil {
				return nil, err
			}
			if token != nil {
				if location == constant.BackupLocationAwsS3.String() {
					tokenRepoInfo, err := repo.FormatS3(token.Endpoint)
					if err != nil {
						return nil, err
					}
					urlRepoInfo, err := repo.FormatS3(endpoint)
					if err != nil {
						return nil, err
					}
					if tokenRepoInfo.Endpoint == urlRepoInfo.Endpoint {
						break
					}
				} else if location == constant.BackupLocationTencentCloud.String() {
					tokenRepoInfo, err := repo.FormatCosByRawUrl(token.Endpoint)
					if err != nil {
						return nil, err
					}
					urlRepoInfo, err := repo.FormatCosByRawUrl(endpoint)
					if err != nil {
						return nil, err
					}
					if tokenRepoInfo.Endpoint == urlRepoInfo.Endpoint {
						break
					}
				}
			}
		}
	}

	if token == nil {
		return nil, fmt.Errorf("integration token not found")
	}

	return token, nil
}

func (i *Integration) GetIntegrationAccountsByLocation(ctx context.Context, owner, location string) ([]string, error) {

	accounts, err := i.queryIntegrationAccounts(ctx, owner)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, account := range accounts {
		if account.Type == "space" {
			continue
		}
		if strings.Contains(location, account.Type) {
			result = append(result, account.Name)
		}
	}

	return result, nil
}

func (i *Integration) ValidIntegrationNameByLocationName(ctx context.Context, owner string, location string, locationConfigName string) (string, error) {
	accounts, err := i.queryIntegrationAccounts(ctx, owner)
	if err != nil {
		return "", err
	}

	var name string

	for _, account := range accounts {
		if util.ListContains([]string{constant.BackupLocationSpace.String(), constant.BackupLocationFileSystem.String()}, location) {
			if account.Type == "space" && (account.Name == locationConfigName || strings.Contains(account.Name, locationConfigName)) {
				name = account.Name
				break
			}
		}
		if location == constant.BackupLocationAwsS3.String() {
			if account.Type == "awss3" && account.Name == locationConfigName {
				name = account.Name
				break
			}
		} else if location == constant.BackupLocationTencentCloud.String() {
			if account.Type == "tencent" && account.Name == locationConfigName {
				name = account.Name
				break
			}
		}

	}

	if name == "" {
		return "", fmt.Errorf("integration account not found, owner: %s, location: %s", owner, location)
	}

	return name, nil
}

func (i *Integration) GetIntegrationNameByLocation(ctx context.Context, owner, location, bucket, region, prefix string) (string, error) {
	if location == constant.BackupLocationFileSystem.String() {
		return "", nil
	}

	accounts, err := i.queryIntegrationAccounts(ctx, owner)
	if err != nil {
		return "", err
	}

	var name string

	if location == constant.BackupLocationFileSystem.String() {
		for _, account := range accounts {
			if account.Type == "space" {
				name = account.Name
				break
			}
		}
		return name, nil
	}

	for _, account := range accounts {
		// account.Type includes: space, awss3, tencent
		// location includes: space, awss3, tencentcloud, filesystem
		if strings.Contains(location, account.Type) {
			if account.Type == "space" {
				name = account.Name
				break
			} else if account.Type == "awss3" {
				token, err := i.GetIntegrationCloudToken(ctx, owner, location, account.Name)
				if err != nil {
					return "", err
				}
				tokenInfo, err := repo.FormatS3(token.Endpoint)
				if err != nil {
					return "", err
				}
				if tokenInfo.Bucket == bucket && tokenInfo.Region == region {
					name = account.Name
					break
				}
			} else if account.Type == "tencent" {
				token, err := i.GetIntegrationCloudToken(ctx, owner, location, account.Name)
				if err != nil {
					return "", err
				}
				tokenInfo, err := repo.FormatCosByRawUrl(token.Endpoint)
				if err != nil {
					return "", err
				}
				if tokenInfo.Bucket == bucket && tokenInfo.Region == region {
					name = account.Name
					break
				}
			}
		}

	}

	if name == "" {
		return "", fmt.Errorf("integration account not found, owner: %s, location: %s", owner, location)
	}

	return name, nil
}

func (i *Integration) withSpaceToken(data *accountResponseData) *SpaceToken {
	return &SpaceToken{
		Name:        data.Name,
		Type:        data.Type,
		OlaresDid:   data.RawData.UserId,
		AccessToken: data.RawData.AccessToken,
		ExpiresAt:   data.RawData.ExpiresAt,
		Available:   data.RawData.Available,
	}
}

func (i *Integration) withCloudToken(data *accountResponseData) *IntegrationToken {
	return &IntegrationToken{
		Name:      data.Name,
		Type:      data.Type,
		AccessKey: data.Name,
		SecretKey: data.RawData.AccessToken,
		Endpoint:  data.RawData.Endpoint,
		Bucket:    data.RawData.Bucket,
		Available: data.RawData.Available,
	}
}

func (i *Integration) queryIntegrationAccounts(ctx context.Context, owner string) ([]*accountsResponseData, error) {
	var authToken, err = i.GetAuthToken(owner, constant.DefaultNamespaceOsFramework, constant.DefaultServiceAccount)
	if err != nil {
		return nil, err
	}

	var integrationUrl = fmt.Sprintf("%s/api/account/list", constant.DefaultIntegrationProviderUrl)
	var data = make(map[string]string)
	data["user"] = owner

	log.Infof("fetch integration from settings: %s", integrationUrl)
	resp, err := i.rest.R().SetContext(ctx).
		SetHeader(constant.BackendAuthorizationHeader, fmt.Sprintf("Bearer %s", authToken)).
		SetBody(data).
		SetResult(&accountsResponse{}).
		Post(integrationUrl)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.WithStack(fmt.Errorf("request account api response not ok, status: %d", resp.StatusCode()))
		return nil, err
	}

	accountsResp := resp.Result().(*accountsResponse)

	if accountsResp.Code == 1 && accountsResp.Message == "" {
		err = errors.WithStack(fmt.Errorf("integration accounts not exists"))
		return nil, err
	} else if accountsResp.Code != 0 {
		err = errors.WithStack(fmt.Errorf("get integration accounts error, status: %d, message: %s", accountsResp.Code, accountsResp.Message))
		return nil, err
	}

	if accountsResp.Data == nil || len(accountsResp.Data) == 0 {
		err = errors.WithStack(fmt.Errorf("integration accounts not exists"))
		return nil, err
	}

	return accountsResp.Data, nil
}

func (i *Integration) query(ctx context.Context, owner, integrationLocation, integrationName string) (*accountResponseData, error) {
	var authToken, err = i.GetAuthToken(owner, constant.DefaultNamespaceOsFramework, constant.DefaultServiceAccount)
	if err != nil {
		return nil, err
	}
	var integrationUrl = fmt.Sprintf("%s/api/account/retrieve", constant.DefaultIntegrationProviderUrl)

	var data = make(map[string]string)
	data["name"] = integrationName
	data["type"] = i.formatUrl(integrationLocation)
	data["user"] = owner
	log.Infof("fetch integration from settings: %s", integrationUrl)
	resp, err := i.rest.R().SetContext(ctx).
		SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
		SetHeader(constant.BackendAuthorizationHeader, fmt.Sprintf("Bearer %s", authToken)).
		SetBody(data).
		SetResult(&accountResponse{}).
		Post(integrationUrl)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		err = errors.WithStack(fmt.Errorf("request account api response not ok, status: %d", resp.StatusCode()))
		return nil, err
	}

	accountResp := resp.Result().(*accountResponse)

	if accountResp.Code == 1 && accountResp.Message == "" {
		err = errors.WithStack(fmt.Errorf("olares space is not enabled"))
		return nil, err
	} else if accountResp.Code != 0 {
		err = errors.WithStack(fmt.Errorf("request account api response error, status: %d, message: %s", accountResp.Code, accountResp.Message))
		return nil, err
	}

	if accountResp.Data == nil || accountResp.Data.RawData == nil {
		err = errors.WithStack(fmt.Errorf("request account api response data is nil, status: %d, message: %s", accountResp.Code, accountResp.Message))
		return nil, err
	}

	return accountResp.Data, nil
}

func (i *Integration) formatUrl(location string) string {
	var l string
	switch location {
	case "space":
		l = "space"
	case "awss3":
		l = "awss3"
	case "tencentcloud":
		l = "tencent"
	}
	return l
}

func (i *Integration) GetAuthToken(owner string, namespace string, sa string) (string, error) {
	var tokenKey = fmt.Sprintf("%s_%s", owner, sa)
	at, ok := i.authToken[tokenKey]
	if ok {
		if time.Now().Before(at.expire) {
			return at.token, nil
		}
	}
	var expirationSeconds int64 = 86400
	tr := &authv1.TokenRequest{
		Spec: authv1.TokenRequestSpec{
			Audiences:         []string{"https://kubernetes.default.svc.cluster.local"},
			ExpirationSeconds: &expirationSeconds, // one day
		},
	}

	kubeClient, _ := i.Factory.KubeClient()

	token, err := kubeClient.CoreV1().ServiceAccounts(namespace).
		CreateToken(context.Background(), sa, tr, metav1.CreateOptions{})
	if err != nil {
		// klog.Errorf("Failed to create token for user %s in namespace %s: %v", owner, namespace, err)
		return "", fmt.Errorf("failed to create token for user %s in namespace %s: %v", owner, namespace, err)
	}

	if !ok {
		at = &authToken{}
	}
	at.token = token.Status.Token
	at.expire = time.Now().Add(40000 * time.Second)

	i.authToken[tokenKey] = at

	return at.token, nil
}
