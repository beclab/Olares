package network

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"text/template"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Controller) ensureProxy(ctx context.Context, binding DomainBinding, spec ProxySpec, jws string) (ProxyStatus, error) {
	status := ProxyStatus{ObservedRevision: spec.Revision}
	conf := spec.Config
	switch conf.Mode {
	case ModeNone, ModePublicIP:
		if err := c.removeAgent(ctx); err != nil {
			return status, err
		}
		if conf.Mode == ModeNone {
			status.Phase, status.Message = "not_configured", "Configure a reverse proxy or public IP to continue"
		} else {
			entrance, err := c.Store.Entrance(ctx)
			if err != nil {
				return status, err
			}
			if entrance.PublicIP == "" || entrance.PublicIPError != "" {
				status.Phase, status.Message = "waiting_public_ip", "Waiting for public IP discovery"
				if entrance.PublicIPError != "" {
					status.Message = entrance.PublicIPError
				}
			} else {
				status.Phase, status.DNSTarget = "ready", entrance.PublicIP
			}
		}
		return status, nil
	case ModeFRP:
		ip, err := c.l4IP(ctx)
		if err != nil {
			return status, err
		}
		if ip == "" {
			status.Phase, status.Message = "waiting_l4", "Waiting for the L4 proxy"
			return status, nil
		}
		if conf.FRPAuthMethod == "token" {
			conf.FRPAuthToken, err = c.Store.AuthToken(ctx, "token")
			if err != nil {
				return status, err
			}
			if conf.FRPAuthToken == "" {
				return status, fmt.Errorf("FRP authentication token is missing")
			}
		}
		domains, err := c.customDomains(ctx)
		if err != nil {
			return status, err
		}
		config, err := RenderFRPC(binding.Domain, conf, jws, ip, utils.EnvOrDefault("L4_PROXY_LISTEN_PROXY_PROTOCOL", constants.L4ListenSSLProxyProtocolPort), domains)
		if err != nil {
			return status, err
		}
		if err := c.Store.PutSecret(ctx, FRPCConfigName, map[string][]byte{"frpc.toml": config}); err != nil {
			return status, err
		}
		dep, err := c.applyAgent(ctx, config)
		if err != nil {
			return status, err
		}
		if dep.Status.ObservedGeneration < dep.Generation || dep.Status.AvailableReplicas < 1 {
			status.Phase, status.Message = "configuring", "Waiting for FRP proxy registration"
			return status, nil
		}
		status.Phase, status.DNSTarget = "ready", conf.DNSTarget
		return status, nil
	default:
		return status, fmt.Errorf("unsupported reverse proxy mode %q", conf.Mode)
	}
}

func (c *Controller) l4IP(ctx context.Context) (string, error) {
	var pods corev1.PodList
	if err := c.Store.Client.List(ctx, &pods, client.InNamespace(Namespace), client.MatchingLabels{"app": "l4-bfl-proxy"}); err != nil {
		return "", err
	}
	sort.Slice(pods.Items, func(i, j int) bool { return pods.Items[i].Name < pods.Items[j].Name })
	for _, pod := range pods.Items {
		if pod.DeletionTimestamp != nil || pod.Status.HostIP == "" || pod.Status.Phase != corev1.PodRunning {
			continue
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				ip := pod.Status.HostIP
				if pod.Spec.NodeName != "" {
					var node corev1.Node
					if err := c.Store.Client.Get(ctx, client.ObjectKey{Name: pod.Spec.NodeName}, &node); apierrors.IsNotFound(err) {
						continue
					} else if err != nil {
						return "", err
					}
					if node.DeletionTimestamp != nil {
						continue
					}
					for _, address := range node.Status.Addresses {
						candidate := net.ParseIP(address.Address)
						if address.Type == corev1.NodeInternalIP && candidate != nil && (candidate.To4() != nil) == (net.ParseIP(ip).To4() != nil) {
							ip = candidate.String()
							break
						}
					}
				}
				if address := net.ParseIP(ip); address != nil && !address.IsLoopback() && !address.IsUnspecified() {
					return address.String(), nil
				}
			}
		}
	}
	return "", nil
}

func (c *Controller) customDomains(ctx context.Context) ([]string, error) {
	var apps appv1.ApplicationList
	if err := c.Store.Client.List(ctx, &apps); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, app := range apps.Items {
		settings := []map[string]string{app.Spec.Settings}
		if appv1.IsShared(&app) {
			settings = nil
			for username := range app.Spec.UserSettings {
				settings = append(settings, app.EffectiveSettings(username))
			}
		}
		for _, entry := range settings {
			var entrances map[string]map[string]string
			if raw := entry[constants.ApplicationCustomDomain]; raw == "" {
				continue
			} else if err := json.Unmarshal([]byte(raw), &entrances); err != nil {
				return nil, fmt.Errorf("application %s has invalid custom domain settings: %w", app.Name, err)
			}
			for _, entrance := range entrances {
				if domain := entrance[constants.ApplicationThirdPartyDomain]; domain != "" {
					seen[domain] = true
				}
			}
		}
	}
	domains := make([]string, 0, len(seen))
	for domain := range seen {
		domains = append(domains, domain)
	}
	sort.Strings(domains)
	return domains, nil
}

func (c *Controller) removeAgent(ctx context.Context) error {
	err := c.Store.Client.Delete(ctx, &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: AgentName, Namespace: Namespace}})
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	return client.IgnoreNotFound(c.Store.Client.Delete(ctx, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: FRPCConfigName, Namespace: Namespace}}))
}

func (c *Controller) applyAgent(ctx context.Context, config []byte) (*appsv1.Deployment, error) {
	hash := sha256.Sum256(config)
	labels := map[string]string{"app": AgentName, "network.olares.io/binding": "default"}
	image := fmt.Sprintf("%s:%s", utils.EnvOrDefault("REVERSE_PROXY_AGENT_IMAGE_NAME", constants.ReverseProxyAgentImage), utils.EnvOrDefault("REVERSE_PROXY_AGENT_IMAGE_VERSION", constants.ReverseProxyAgentImageVersion))
	desired := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: AgentName, Namespace: Namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1), Strategy: appsv1.DeploymentStrategy{Type: appsv1.RecreateDeploymentStrategyType},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": AgentName}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels, Annotations: map[string]string{"network.olares.io/config-hash": hex.EncodeToString(hash[:])}},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: pointer.Bool(false), PriorityClassName: "system-cluster-critical",
					Containers: []corev1.Container{{Name: "agent", Image: image, ImagePullPolicy: corev1.PullIfNotPresent,
						Command: []string{"/frpc-agent"}, Args: []string{"--config", "/etc/frpc/frpc.toml", "--health-address", ":7401"},
						Ports:          []corev1.ContainerPort{{Name: "health", ContainerPort: 7401}},
						ReadinessProbe: &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/readyz", Port: intstr.FromInt32(7401)}}, PeriodSeconds: 3, TimeoutSeconds: 3, FailureThreshold: 3, SuccessThreshold: 1},
						LivenessProbe:  &corev1.Probe{ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromInt32(7401)}}, PeriodSeconds: 10, TimeoutSeconds: 3, FailureThreshold: 3, SuccessThreshold: 1},
						VolumeMounts:   []corev1.VolumeMount{{Name: "config", MountPath: "/etc/frpc", ReadOnly: true}},
					}},
					Volumes: []corev1.Volume{{Name: "config", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: FRPCConfigName, DefaultMode: pointer.Int32(0400)}}}},
				},
			},
		},
	}
	var result *appsv1.Deployment
	err := retry.OnError(retry.DefaultRetry, func(err error) bool {
		return apierrors.IsConflict(err) || apierrors.IsAlreadyExists(err)
	}, func() error {
		var current appsv1.Deployment
		err := c.Store.Client.Get(ctx, client.ObjectKeyFromObject(desired), &current)
		if apierrors.IsNotFound(err) {
			created := desired.DeepCopy()
			if err := c.Store.Client.Create(ctx, created); err != nil {
				return err
			}
			result = created
			return nil
		}
		if err != nil {
			return err
		}
		if equality.Semantic.DeepDerivative(desired.Spec, current.Spec) {
			result = &current
			return nil
		}
		// Keep server defaults and controller-owned annotations while updating
		// this agent's fields. The resource version also protects status races.
		patch, err := json.Marshal(map[string]interface{}{
			"metadata": map[string]interface{}{"resourceVersion": current.ResourceVersion, "labels": labels},
			"spec":     desired.Spec,
		})
		if err != nil {
			return err
		}
		if err := c.Store.Client.Patch(ctx, &current, client.RawPatch(types.StrategicMergePatchType, patch)); err != nil {
			return err
		}
		result = &current
		return nil
	})
	return result, err
}

const frpcTemplate = `serverAddr = {{quote .Config.FRPServer}}
serverPort = {{.Config.FRPPort}}
user = {{quote .CloudName}}
loginFailExit = false
webServer.addr = "127.0.0.1"
webServer.port = 7400
{{if eq .Config.FRPAuthMethod "jws"}}auth.method = "jws"
auth.jws = {{quote .JWS}}
{{else}}auth.method = "token"
auth.token = {{quote .Config.FRPAuthToken}}
{{end}}
{{range .Proxies}}
[[proxies]]
name = {{quote .Name}}
type = {{quote .Type}}
localIP = {{quote $.IP}}
localPort = {{.Port}}
customDomains = [{{quote .Domain}}]
{{if eq .Type "https"}}transport.proxyProtocolVersion = "v1"
{{end}}{{end}}`

type frpcProxy struct{ Name, Type, Domain, Port string }

func RenderFRPC(domain string, conf ProxyConfig, jws, ip, sslPort string, customDomains []string) ([]byte, error) {
	_, name, err := NormalizeDomain(domain)
	if err != nil {
		return nil, err
	}
	proxies := []frpcProxy{{"web", "http", domain, "80"}, {"web_wildcard", "http", "*." + domain, "80"}, {"web_ssl", "https", domain, sslPort}, {"web_ssl_wildcard", "https", "*." + domain, sslPort}}
	for i, custom := range customDomains {
		proxies = append(proxies, frpcProxy{fmt.Sprintf("custom_%d", i), "https", custom, sslPort})
	}
	quote := func(value interface{}) string { data, _ := json.Marshal(value); return string(data) }
	t, err := template.New("frpc").Funcs(template.FuncMap{"quote": quote}).Parse(frpcTemplate)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = t.Execute(&out, struct {
		Config             ProxyConfig
		CloudName, JWS, IP string
		Proxies            []frpcProxy
	}{conf, string(name), jws, ip, proxies})
	return out.Bytes(), err
}
