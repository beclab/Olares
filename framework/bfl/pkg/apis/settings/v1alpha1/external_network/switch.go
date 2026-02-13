package external_network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/constants"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

const (
	PhaseProcessing = "PROCESSING"
	PhaseCompleted  = "COMPLETED"
	PhaseFailed     = "FAILED"
)

type SwitchConfig struct {
	Spec   SwitchSpec   `json:"spec"`
	Status SwitchStatus `json:"status"`
}

type SwitchSpec struct {
	Disabled bool `json:"disabled"`
}

type SwitchStatus struct {
	Phase     string `json:"phase,omitempty"`
	StartedAt string `json:"startedAt,omitempty"`
	TaskID    string `json:"taskId,omitempty"`
	Message   string `json:"message,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func DefaultConfig() *SwitchConfig {
	return &SwitchConfig{
		Spec: SwitchSpec{Disabled: false},
		Status: SwitchStatus{
			Phase: PhaseCompleted,
		},
	}
}

func (c *SwitchConfig) PrettyJSON() (string, error) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ParseConfigMap(cm *corev1.ConfigMap) (*SwitchConfig, error) {
	if cm == nil || cm.Data == nil {
		return DefaultConfig(), nil
	}
	raw := cm.Data[constants.ExternalNetworkSwitchConfigKey]
	if raw == "" {
		return DefaultConfig(), nil
	}
	cfg := &SwitchConfig{}
	if err := json.Unmarshal([]byte(raw), cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal external network switch config")
	}
	// tolerate missing fields
	if cfg.Status.Phase == "" {
		cfg.Status.Phase = PhaseCompleted
	}
	return cfg, nil
}

func Load(ctx context.Context) (*SwitchConfig, *corev1.ConfigMap, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, nil, err
	}
	cm, err := kc.Kubernetes().CoreV1().ConfigMaps(constants.OSSystemNamespace).Get(ctx, constants.ExternalNetworkSwitchConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return DefaultConfig(), nil, nil
		}
		return nil, nil, err
	}
	cfg, err := ParseConfigMap(cm)
	return cfg, cm, err
}

func Upsert(ctx context.Context, mutate func(*SwitchConfig) error) (*SwitchConfig, error) {
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return nil, err
	}

	for {
		cm, err := kc.Kubernetes().CoreV1().ConfigMaps(constants.OSSystemNamespace).Get(ctx, constants.ExternalNetworkSwitchConfigMapName, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return nil, err
		}

		var cfg *SwitchConfig
		if apierrors.IsNotFound(err) {
			cfg = DefaultConfig()
			cm = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.ExternalNetworkSwitchConfigMapName,
					Namespace: constants.OSSystemNamespace,
				},
				Data: map[string]string{},
			}
		} else {
			cfg, err = ParseConfigMap(cm)
			if err != nil {
				return nil, err
			}
			if cm.Data == nil {
				cm.Data = map[string]string{}
			}
		}

		if mutate != nil {
			if err := mutate(cfg); err != nil {
				return nil, err
			}
		}
		cfg.Status.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		raw, err := cfg.PrettyJSON()
		if err != nil {
			return nil, err
		}
		cm.Data[constants.ExternalNetworkSwitchConfigKey] = raw

		if cm.ResourceVersion == "" {
			_, err = kc.Kubernetes().CoreV1().ConfigMaps(constants.OSSystemNamespace).Create(ctx, cm, metav1.CreateOptions{})
			if err != nil {
				if apierrors.IsAlreadyExists(err) {
					continue
				}
				return nil, err
			}
			return cfg, nil
		}

		_, err = kc.Kubernetes().CoreV1().ConfigMaps(constants.OSSystemNamespace).Update(ctx, cm, metav1.UpdateOptions{})
		if err != nil {
			if apierrors.IsConflict(err) {
				continue
			}
			return nil, err
		}
		return cfg, nil
	}
}

type IntegrationAccount struct {
	AccessToken string
	UserID      string
	RawData     any
}

type settingsRetrieveAccountResponse struct {
	Code    int `json:"code"`
	Message any `json:"message"`
	Data    *struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		RawData struct {
			RefreshToken string `json:"refresh_token"`
			AccessToken  string `json:"access_token"`
			ExpiresIn    int64  `json:"expires_in"`
			ExpiresAt    int64  `json:"expires_at"`
			UserID       string `json:"userid"`
			Available    bool   `json:"available"`
			CreateAt     int64  `json:"create_at"`
		} `json:"raw_data"`
	} `json:"data"`
}

func getUserBackendServiceAccountToken(ctx context.Context, ownerUsername string) (string, error) {
	if ownerUsername == "" {
		return "", errors.New("empty owner username")
	}
	ns := fmt.Sprintf("user-system-%s", ownerUsername)
	kc, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return "", err
	}
	// Use TokenRequest API to mint a token for serviceaccount "user-backend" in the user's system namespace.
	tr := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			// leave audiences default; just request a short-lived token
			ExpirationSeconds: pointer.Int64(3600),
		},
	}
	resp, err := kc.Kubernetes().CoreV1().ServiceAccounts(ns).CreateToken(ctx, "user-backend", tr, metav1.CreateOptions{})
	if err != nil {
		return "", errors.Wrap(err, "create serviceaccount token failed")
	}
	if resp == nil || resp.Status.Token == "" {
		return "", errors.New("empty serviceaccount token")
	}
	return resp.Status.Token, nil
}

// GetIntegrationAccount fetches integration account token/userid from settings backend:
// POST http://settings.user-system-{owner}:28080/api/account/retrieve
// body: {"name":"integration-account:space:{terminusName}"}
// auth: serviceaccount token of user-backend in namespace user-system-{owner}.
func GetIntegrationAccount(ctx context.Context, ownerUsername, terminusName string) (*IntegrationAccount, error) {
	if ownerUsername == "" {
		return nil, errors.New("empty owner username")
	}
	if terminusName == "" {
		return nil, errors.New("empty terminus name")
	}
	saToken, err := getUserBackendServiceAccountToken(ctx, ownerUsername)
	if err != nil {
		return nil, err
	}

	ns := fmt.Sprintf("user-system-%s", ownerUsername)
	reqURL := fmt.Sprintf("http://settings.%s:28080/api/account/retrieve", ns)
	payload := map[string]string{
		"name": fmt.Sprintf("integration-account:space:%s", terminusName),
	}
	var out settingsRetrieveAccountResponse
	r, err := resty.New().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", saToken)).
		SetBody(payload).
		SetResult(&out).
		Post(reqURL)
	if err != nil {
		return nil, err
	}
	if r.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("settings retrieve account: status=%d body=%s", r.StatusCode(), string(r.Body()))
	}
	if out.Code != 0 || out.Data == nil {
		return nil, fmt.Errorf("settings retrieve account: code=%d message=%v", out.Code, out.Message)
	}
	if out.Data.RawData.AccessToken == "" || out.Data.RawData.UserID == "" {
		return nil, fmt.Errorf("settings retrieve account: missing access_token/userid")
	}
	return &IntegrationAccount{
		AccessToken: out.Data.RawData.AccessToken,
		UserID:      out.Data.RawData.UserID,
		RawData:     out.Data.RawData,
	}, nil
}

type PublicDNSDeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		TaskID string `json:"taskId"`
		Status string `json:"status"`
	} `json:"data"`
}

type PublicDNSInfoResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		DNSType  string `json:"dnsType"`
		DNSValue string `json:"dnsValue"`
		DNSName  string `json:"dnsName"`
	} `json:"data"`
}

type PublicDNSTaskResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		TaskID     string `json:"taskId"`
		TerminusID string `json:"terminusId"`
		Tasks      []struct {
			UserID string `json:"userid"`
			Status string `json:"status"`
			Msg    string `json:"message"`
		} `json:"tasks"`
	} `json:"data"`
}

func publicDNSBase() (string, error) {
	u, err := url.JoinPath(constants.OlaresRemoteService, "/space/v1/resource/public-dns")
	if err != nil {
		return "", err
	}
	return u, nil
}

type PublicDNSRecoverResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type PublicDNSAuthPayload struct {
	Token  string `json:"token"`
	UserID string `json:"userid"`
}

func RecoverPublicDNS(ctx context.Context, auth PublicDNSAuthPayload) error {
	base, err := publicDNSBase()
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("%s/recover", base)
	var out PublicDNSRecoverResponse
	r, err := resty.New().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(auth).
		SetResult(&out).
		Post(endpoint)
	if err != nil {
		return err
	}
	if r.StatusCode() != http.StatusOK {
		return fmt.Errorf("public dns recover: status=%d body=%s", r.StatusCode(), string(r.Body()))
	}
	if out.Code != 200 {
		return fmt.Errorf("public dns recover: unexpected response: code=%d message=%s", out.Code, out.Message)
	}
	return nil
}

func CreatePublicDNSDeleteTask(ctx context.Context, auth PublicDNSAuthPayload) (string, error) {
	base, err := publicDNSBase()
	if err != nil {
		return "", err
	}
	endpoint := fmt.Sprintf("%s/delete", base)
	var out PublicDNSDeleteResponse
	r, err := resty.New().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(auth).
		SetResult(&out).
		Post(endpoint)
	if err != nil {
		return "", err
	}
	if r.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("public dns delete: status=%d body=%s", r.StatusCode(), string(r.Body()))
	}
	if out.Code != 200 || out.Data == nil || out.Data.TaskID == "" {
		return "", fmt.Errorf("public dns delete: unexpected response: code=%d message=%s", out.Code, out.Message)
	}
	return out.Data.TaskID, nil
}

func GetPublicDNSInfo(ctx context.Context, auth PublicDNSAuthPayload) (bool, error) {
	base, err := publicDNSBase()
	if err != nil {
		return false, err
	}
	endpoint := fmt.Sprintf("%s/info", base)
	var out PublicDNSInfoResponse
	r, err := resty.New().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(auth).
		SetResult(&out).
		Post(endpoint)
	if err != nil {
		return false, err
	}
	if r.StatusCode() != http.StatusOK {
		return false, fmt.Errorf("public dns info: status=%d body=%s", r.StatusCode(), string(r.Body()))
	}
	switch out.Code {
	case 200:
		if out.Data == nil {
			return false, fmt.Errorf("public dns info: missing data")
		}
		return true, nil
	case 501:
		return false, nil
	default:
		return false, fmt.Errorf("public dns info: unexpected response: code=%d message=%s", out.Code, out.Message)
	}
}

func GetPublicDNSTask(ctx context.Context, auth PublicDNSAuthPayload, taskID string) (*PublicDNSTaskResponse, error) {
	base, err := publicDNSBase()
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/task/%s", base, url.PathEscape(taskID))
	var out PublicDNSTaskResponse
	r, err := resty.New().SetTimeout(30*time.Second).R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(auth).
		SetResult(&out).
		Post(endpoint)
	if err != nil {
		return nil, err
	}
	if r.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("public dns task: status=%d body=%s", r.StatusCode(), string(r.Body()))
	}
	if out.Code != 200 || out.Data == nil {
		return nil, fmt.Errorf("public dns task: unexpected response: code=%d message=%s", out.Code, out.Message)
	}
	return &out, nil
}

func (r *PublicDNSTaskResponse) AllTasksCompleted() (bool, string) {
	if r == nil || r.Data == nil {
		return false, "empty response"
	}
	if len(r.Data.Tasks) == 0 {
		return false, "no tasks"
	}
	var failed []string
	for _, t := range r.Data.Tasks {
		switch t.Status {
		case "COMPLETED":
			continue
		case "FAILED":
			failed = append(failed, fmt.Sprintf("%s:%s", t.UserID, t.Msg))
		default:
			// PENDING/PROCESSING
			return false, ""
		}
	}
	if len(failed) > 0 {
		return false, fmt.Sprintf("failed tasks: %v", failed)
	}
	return true, ""
}
