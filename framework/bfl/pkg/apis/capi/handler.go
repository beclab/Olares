package capi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/apis"
	"bytetrade.io/web3os/bfl/pkg/apiserver/authn"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	v1alpha1client "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"

	iamv1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const (
	terminusdHostEnv = "TERMINUSD_HOST"
	defaultOlaresd   = "master-node:18088"

	roleTemplateRulesAnnotation = "iam.kubesphere.io/role-template-rules"
)

var (
	globalRoleBindingGVR = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "globalrolebindings"}
	globalRoleGVR        = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "globalroles"}
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
	iamClient  dynamic.Interface
	httpClient *http.Client
	ksBaseURL  string
	olaresdURL string
}

func newHandler() (*handler, error) {
	kubeClient, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, fmt.Errorf("create capi kube client: %w", err)
	}
	iamClient, err := dynamic.NewForConfig(kubeClient.Config())
	if err != nil {
		return nil, fmt.Errorf("create capi IAM client: %w", err)
	}
	return &handler{
		appService: app_service.NewAppServiceClient(),
		kubeClient: kubeClient,
		iamClient:  iamClient,
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
	User             userDetail `json:"user"`
	Runtime          string     `json:"runtime"`
	SystemNamespaces []string   `json:"systemNamespaces"`
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

	clusterRole := "host"
	detail := appDetailResponse{
		Runtime:          h.containerRuntime(req.Request.Context()),
		SystemNamespaces: append([]string(nil), systemNamespaces...),
		User: userDetail{
			Email:           user.Spec.Email,
			Lang:            user.Spec.Lang,
			Username:        user.Name,
			GlobalRole:      globalRole(user.Annotations[constants.UserAnnotationOwnerRole]),
			GrantedClusters: []string{},
			LastLoginTime:   user.Status.LastLoginTime,
			GlobalRules:     map[string][]string{},
		},
	}
	detail.User.GlobalRules = h.globalRules(req.Request.Context(), identity.Username, clusterRole)
	log.Infow("appDetail response", "username", identity.Username, "detail", detail)
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

func (h *handler) globalRules(ctx context.Context, username, clusterRole string) map[string][]string {
	rules, err := globalRulesFromClient(ctx, h.iamClient, username, clusterRole)
	if err != nil {
		log.Warnf("get global rules for %s: %v", username, err)
		return map[string][]string{}
	}
	return rules
}

func globalRulesFromClient(ctx context.Context, iamClient dynamic.Interface, username, clusterRole string) (map[string][]string, error) {
	result := make(map[string][]string)
	bindings, err := iamClient.Resource(globalRoleBindingGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return result, fmt.Errorf("list global role bindings: %w", err)
	}

	roleName := ""
	for i := range bindings.Items {
		subjects, _, _ := unstructured.NestedSlice(bindings.Items[i].Object, "subjects")
		for _, rawSubject := range subjects {
			subject, ok := rawSubject.(map[string]any)
			if !ok || subject["name"] != username {
				continue
			}
			roleName, _, _ = unstructured.NestedString(bindings.Items[i].Object, "roleRef", "name")
			break
		}
		if roleName != "" {
			break
		}
	}
	if roleName == "" {
		return result, nil
	}

	globalRole, err := iamClient.Resource(globalRoleGVR).Get(ctx, roleName, metav1.GetOptions{})
	if err != nil {
		return result, fmt.Errorf("get global role %s: %w", roleName, err)
	}
	var templateNames []string
	if raw := globalRole.GetAnnotations()[iamv1alpha2.AggregationRolesAnnotation]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &templateNames); err != nil {
			return result, nil
		}
	}

	templates := make([]*unstructured.Unstructured, 0, len(templateNames))
	for _, templateName := range templateNames {
		role, err := iamClient.Resource(globalRoleGVR).Get(ctx, templateName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return result, fmt.Errorf("get aggregated global role %s: %w", templateName, err)
		}
		templates = append(templates, role)
	}
	// KubeSphere's users/{user}/globalroles endpoint sorts aggregated roles by
	// creation time descending, then name descending when timestamps match.
	sort.SliceStable(templates, func(i, j int) bool {
		left, right := templates[i], templates[j]
		leftCreatedAt, rightCreatedAt := left.GetCreationTimestamp(), right.GetCreationTimestamp()
		if leftCreatedAt.Equal(&rightCreatedAt) {
			return left.GetName() > right.GetName()
		}
		return leftCreatedAt.After(rightCreatedAt.Time)
	})

	for _, role := range templates {
		var rules map[string]any
		if json.Unmarshal([]byte(role.GetAnnotations()[roleTemplateRulesAnnotation]), &rules) != nil {
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
	return result, nil
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
