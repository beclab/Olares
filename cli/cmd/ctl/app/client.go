package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	marketServiceName      = "appstore-service"
	chartRepoServiceName   = "chart-repo-service"
	marketServiceNamespace = "os-framework"
	apiPrefix              = "/app-store/api/v2"
	headerBflUser          = "X-Bfl-User"
	defaultRequestTimeout  = 5 * time.Minute
)

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type MarketClient struct {
	baseURL    string
	httpClient *http.Client
	user       string
	source     string
}

func NewMarketClient(host, user, source string) *MarketClient {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	host = strings.TrimRight(host, "/")
	return &MarketClient{
		baseURL:    host + apiPrefix,
		httpClient: &http.Client{Timeout: defaultRequestTimeout},
		user:       user,
		source:     source,
	}
}

func (c *MarketClient) doRequest(ctx context.Context, method, path string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.user != "" {
		req.Header.Set(headerBflUser, c.user)
	}

	return c.executeRequest(req)
}

func (c *MarketClient) doMultipart(ctx context.Context, path, filename string, data io.Reader, source string) (*APIResponse, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("chart", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, data); err != nil {
		return nil, fmt.Errorf("failed to copy chart data: %w", err)
	}
	if err := writer.WriteField("source", source); err != nil {
		return nil, fmt.Errorf("failed to write source field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	if c.user != "" {
		req.Header.Set(headerBflUser, c.user)
	}

	return c.executeRequest(req)
}

func (c *MarketClient) executeRequest(req *http.Request) (*APIResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || !apiResp.Success {
		message := strings.TrimSpace(apiResp.Message)
		if message == "" {
			message = strings.TrimSpace(string(respBody))
		}
		if message == "" {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &apiResp, fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, message)
	}

	return &apiResp, nil
}

func (c *MarketClient) GetMarketData(ctx context.Context) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/market/data", nil)
}

func (c *MarketClient) GetMarketState(ctx context.Context) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodGet, "/market/state", nil)
}

func (c *MarketClient) GetAppsInfo(ctx context.Context, apps []AppQueryInfo) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps", map[string]interface{}{
		"apps": apps,
	})
}

func (c *MarketClient) UploadChart(ctx context.Context, filePath, source string) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return c.doMultipart(ctx, "/apps/upload", file.Name(), file, source)
}

func (c *MarketClient) UploadChartFromReader(ctx context.Context, filename string, data io.Reader, source string) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doMultipart(ctx, "/apps/upload", filename, data, source)
}

func (c *MarketClient) DeleteLocalApp(ctx context.Context, appName, appVersion, sourceID string) (*APIResponse, error) {
	if sourceID == "" {
		sourceID = c.source
	}
	return c.doRequest(ctx, http.MethodDelete, "/local-apps/delete", map[string]string{
		"app_name":    appName,
		"app_version": appVersion,
		"source":      sourceID,
	})
}

func (c *MarketClient) InstallApp(ctx context.Context, appName, version, source string, envs []AppEnvVar) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/install", InstallRequest{
		Source:  source,
		AppName: appName,
		Version: version,
		Sync:    true,
		Envs:    envs,
	})
}

func (c *MarketClient) CloneApp(ctx context.Context, appName, source, title string, envs []AppEnvVar, entrances []AppEntrance) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPost, "/apps/"+appName+"/clone", CloneRequest{
		Source:    source,
		AppName:   appName,
		Title:     title,
		Sync:      true,
		Envs:      envs,
		Entrances: entrances,
	})
}

func (c *MarketClient) UninstallApp(ctx context.Context, appName string, all, deleteData bool) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, "/apps/"+appName, UninstallRequest{
		Sync:       true,
		All:        all,
		DeleteData: deleteData,
	})
}

func (c *MarketClient) UpgradeApp(ctx context.Context, appName, version, source string, envs []AppEnvVar) (*APIResponse, error) {
	if source == "" {
		source = c.source
	}
	return c.doRequest(ctx, http.MethodPut, "/apps/"+appName+"/upgrade", InstallRequest{
		Source:  source,
		AppName: appName,
		Version: version,
		Sync:    true,
		Envs:    envs,
	})
}

func (c *MarketClient) CancelOperation(ctx context.Context, appName string) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodDelete, "/apps/"+appName+"/install", map[string]interface{}{
		"sync": true,
	})
}

func (c *MarketClient) ResumeApp(ctx context.Context, appName string) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps/resume", map[string]string{
		"appName": appName,
	})
}

func (c *MarketClient) StopApp(ctx context.Context, appName string, all bool) (*APIResponse, error) {
	return c.doRequest(ctx, http.MethodPost, "/apps/stop", map[string]interface{}{
		"appName": appName,
		"all":     all,
	})
}

func newKubeClient(kubeconfig string) (client.Client, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := iamv1alpha2.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add user scheme: %w", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add core scheme: %w", err)
	}

	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}
	return kubeClient, nil
}

func discoverMarketEndpoint(kubeClient client.Client) (string, error) {
	var svc corev1.Service
	key := types.NamespacedName{
		Name:      marketServiceName,
		Namespace: marketServiceNamespace,
	}
	if err := kubeClient.Get(context.Background(), key, &svc); err != nil {
		return "", fmt.Errorf("failed to get service %s/%s: %w", marketServiceNamespace, marketServiceName, err)
	}

	clusterIP := svc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		return "", fmt.Errorf("service %s/%s has no ClusterIP", marketServiceNamespace, marketServiceName)
	}

	port := 81
	for _, p := range svc.Spec.Ports {
		if p.Name == "appstore-backend" || p.Port == 81 {
			port = int(p.Port)
			break
		}
	}

	return fmt.Sprintf("%s:%d", clusterIP, port), nil
}

func discoverChartRepoEndpoint(kubeClient client.Client) (string, error) {
	var svc corev1.Service
	key := types.NamespacedName{
		Name:      chartRepoServiceName,
		Namespace: marketServiceNamespace,
	}
	if err := kubeClient.Get(context.Background(), key, &svc); err != nil {
		return "", fmt.Errorf("failed to get service %s/%s: %w", marketServiceNamespace, chartRepoServiceName, err)
	}

	clusterIP := svc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		return "", fmt.Errorf("service %s/%s has no ClusterIP", marketServiceNamespace, chartRepoServiceName)
	}

	port := 82
	for _, p := range svc.Spec.Ports {
		if p.Port != 0 {
			port = int(p.Port)
			break
		}
	}

	return fmt.Sprintf("%s:%d", clusterIP, port), nil
}

func resolveUser(kubeClient client.Client) (string, error) {
	var userList iamv1alpha2.UserList
	if err := kubeClient.List(context.Background(), &userList); err != nil {
		return "", fmt.Errorf("failed to list users: %w", err)
	}
	users := userList.Items
	if len(users) == 0 {
		return "", fmt.Errorf("no users found in cluster")
	}
	if len(users) == 1 {
		return users[0].Name, nil
	}
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}
	return "", fmt.Errorf("multiple users found (%s), use --user to specify one", strings.Join(names, ", "))
}
