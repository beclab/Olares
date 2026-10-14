package capi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apis"
	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	terminusdHostEnv = "TERMINUSD_HOST"
	defaultOlaresd   = "master-node:18088"
)

var systemNamespaces = []string{
	"os-platform",
	"os-network",
	"os-framework",
	"os-gpu",
	"os-mesh",
	"os-gateway",
	"kubesphere-monitoring-federated",
	"kubesphere-controls-system",
	"kubesphere-system",
	"kubesphere-monitoring-system",
	"kubekey-system",
	"kube-system",
	"kube-public",
	"kube-node-lease",
	"default",
}

type handler struct {
	base       apis.Base
	appService *app_service.Client
	kubeClient v1alpha1client.ClientInterface
	httpClient *http.Client
	ksBaseURL  string
	olaresdURL string
}

func newHandler() (*handler, error) {
	kubeClient, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, fmt.Errorf("create capi kube client: %w", err)
	}
	return &handler{
		appService: app_service.NewAppServiceClient(),
		kubeClient: kubeClient,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		ksBaseURL:  normalizeHTTPBase(constants.KubeSphereAPIHost),
		olaresdURL: normalizeHTTPBase(envOrDefault(terminusdHostEnv, defaultOlaresd)),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func normalizeHTTPBase(host string) string {
	host = strings.TrimRight(host, "/")
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "http://" + host
}

func (h *handler) myApps(req *restful.Request, resp *restful.Response) {
	req.SetAttribute(constants.MyAppsIsLocalKey, false)
	apps, err := h.base.GetAppListAndServicePort(req, h.appService,
		func() (string, []*app_service.AppInfo, error) {
			return h.base.GetAppViaOwner(h.appService)
		})
	if err != nil {
		writeUpstreamError(resp, fmt.Errorf("list apps: %w", err))
		return
	}
	writeJSON(resp, http.StatusOK, apps)
}

func (h *handler) systemIFS(req *restful.Request, resp *restful.Response) {
	target := h.olaresdURL + "/system/ifs"
	if req.Request.URL.RawQuery != "" {
		target += "?" + req.Request.URL.RawQuery
	}
	upstreamReq, err := http.NewRequestWithContext(req.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		writeUpstreamError(resp, err)
		return
	}
	upstreamReq.Header.Set("X-Signature", "did jws")
	upstreamResp, err := h.httpClient.Do(upstreamReq)
	if err != nil {
		writeUpstreamError(resp, fmt.Errorf("request olaresd: %w", err))
		return
	}
	defer upstreamResp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(upstreamResp.Body, 10<<20))
	if err != nil {
		writeUpstreamError(resp, fmt.Errorf("read olaresd response: %w", err))
		return
	}
	if upstreamResp.StatusCode < http.StatusOK || upstreamResp.StatusCode >= http.StatusMultipleChoices {
		writeJSON(resp, upstreamResp.StatusCode, json.RawMessage(body))
		return
	}

	data, err := unwrapData(body)
	if err != nil {
		writeUpstreamError(resp, fmt.Errorf("decode olaresd response: %w", err))
		return
	}
	resp.AddHeader("Content-Type", restful.MIME_JSON)
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write(data)
}

func unwrapData(body []byte) ([]byte, error) {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return body, nil
	}
	return envelope.Data, nil
}

type appDetailResponse struct {
	LocaleManifest   map[string]any `json:"localeManifest"`
	User             userDetail     `json:"user"`
	Runtime          string         `json:"runtime"`
	ClusterRole      string         `json:"clusterRole"`
	GitOpsEngine     string         `json:"gitopsEngine"`
	SystemNamespaces []string       `json:"systemNamespaces"`
}

type userDetail struct {
	Email           string              `json:"email"`
	Lang            string              `json:"lang,omitempty"`
	Username        string              `json:"username"`
	GlobalRole      string              `json:"globalrole"`
	GrantedClusters []string            `json:"grantedClusters"`
	LastLoginTime   *metav1.Time        `json:"lastLoginTime,omitempty"`
	GlobalRules     map[string][]string `json:"globalRules"`
}

func (h *handler) appDetail(req *restful.Request, resp *restful.Response) {
	identity, err := authn.Authenticate(req.Request)
	if err != nil {
		writeJSON(resp, http.StatusUnauthorized, map[string]string{"message": err.Error()})
		return
	}

	var user iamv1alpha2.User
	if err := h.kubeClient.CtrlClient().Get(req.Request.Context(), types.NamespacedName{Name: identity.Username}, &user); err != nil {
		writeUpstreamError(resp, fmt.Errorf("get user: %w", err))
		return
	}

	clusterRole := h.clusterRole(req.Request.Context())
	detail := appDetailResponse{
		LocaleManifest:   map[string]any{},
		Runtime:          h.containerRuntime(req.Request.Context()),
		ClusterRole:      clusterRole,
		GitOpsEngine:     "argocd",
		SystemNamespaces: append([]string(nil), systemNamespaces...),
		User: userDetail{
			Email:           user.Spec.Email,
			Lang:            user.Spec.Lang,
			Username:        user.Name,
			GlobalRole:      globalRole(user.Annotations[constants.UserAnnotationOwnerRole]),
			GrantedClusters: parseStringList(user.Annotations["iam.kubesphere.io/granted-clusters"]),
			LastLoginTime:   user.Status.LastLoginTime,
			GlobalRules:     map[string][]string{},
		},
	}
	if identity.Token != "" {
		detail.User.GlobalRules = h.globalRules(req.Request.Context(), identity.Token, identity.Username, clusterRole)
	}
	writeJSON(resp, http.StatusOK, detail)
}

func globalRole(ownerRole string) string {
	if ownerRole == constants.RoleOwner || ownerRole == constants.RoleAdmin {
		return "platform-admin"
	}
	return ownerRole
}

func parseStringList(value string) []string {
	if value == "" {
		return []string{}
	}
	var values []string
	if json.Unmarshal([]byte(value), &values) == nil {
		return values
	}
	return []string{value}
}

func (h *handler) containerRuntime(ctx context.Context) string {
	nodes, err := h.kubeClient.Kubernetes().CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil || len(nodes.Items) == 0 {
		return "docker"
	}
	runtimeVersion := nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion
	if name, _, ok := strings.Cut(runtimeVersion, ":"); ok && name != "" {
		return name
	}
	if runtimeVersion == "" {
		return "docker"
	}
	return runtimeVersion
}

var clusterRolePattern = regexp.MustCompile(`(?m)clusterRole:\s*(host|member)\s*$`)

func (h *handler) clusterRole(ctx context.Context) string {
	configMap, err := h.kubeClient.Kubernetes().CoreV1().ConfigMaps(constants.KubeSphereNamespace).
		Get(ctx, constants.KubeSphereConfigName, metav1.GetOptions{})
	if err != nil {
		return "host"
	}
	matches := clusterRolePattern.FindStringSubmatch(configMap.Data["kubesphere.yaml"])
	if len(matches) != 2 {
		return "host"
	}
	return matches[1]
}

func (h *handler) globalRules(ctx context.Context, token, username, clusterRole string) map[string][]string {
	var roles []struct {
		Metadata struct {
			Annotations map[string]string `json:"annotations"`
		} `json:"metadata"`
	}
	path := "/kapis/iam.kubesphere.io/v1alpha2/users/" + url.PathEscape(username) + "/globalroles"
	if err := h.getKSJSON(ctx, token, path, nil, &roles); err != nil {
		return map[string][]string{}
	}
	result := make(map[string][]string)
	for _, role := range roles {
		var rules map[string]any
		if json.Unmarshal([]byte(role.Metadata.Annotations["iam.kubesphere.io/role-template-rules"]), &rules) != nil {
			continue
		}
		for key, raw := range rules {
			for _, value := range toStrings(raw) {
				if clusterRole == "member" && (key == "users" || key == "workspaces") && value == "manage" {
					continue
				}
				if !contains(result[key], value) {
					result[key] = append(result[key], value)
				}
			}
		}
	}
	return result
}

func toStrings(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (h *handler) getKSJSON(ctx context.Context, token, path string, query url.Values, target any) error {
	requestURL := h.ksBaseURL + path
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		return fmt.Errorf("kubesphere %s returned %s: %s", path, response.Status, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(io.LimitReader(response.Body, 20<<20)).Decode(target)
}
