package v1alpha1

import (
	"fmt"
	"runtime"
	"strconv"

	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	applyAppsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
)

var (
	ReverseProxyAgentDeploymentReplicas int32 = 1

	ReverseProxyAgentDeploymentName = "reverse-proxy-agent"

	ReverseProxyAgentSelectEnvKey = "ReverseProxy"

	ReverseProxyAgentSelectCloudFlareEnvVal = "CLOUDFLARE"

	ReverseProxyAgentSelectFRPEnvVal = "FRP"

	L4ProxyDeploymentName = "l4-bfl-proxy"

	L4ProxyDeploymentReplicas int32 = 1

	CloudflaredDeploymentName = "cloudflare-tunnel"

	CloudflareDeploymentReplicas int32 = 1
)

const (
	ServiceEnabled  = "enabled"
	ServiceDisabled = "disabled"
)

type AccessLevel uint64

const (
	_ AccessLevel = iota

	WorldWide
	Public
	Protected
	Private
)

type AuthPolicy string

const (
	OneFactor = "one_factor"
	TwoFactor = "two_factor"
)

const DefaultAuthPolicy = TwoFactor

const DefaultPodsCIDR = "10.233.64.0/18"

type PostTerminusName struct {
	JWSSignature string `json:"jws_signature"`
	DID          string `json:"did"`
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	URL    string `json:"url"`
}

type ResponseServices struct {
	Services []ServiceStatus `json:"services"`
}

type LauncherAccessPolicy struct {
	AccessLevel AccessLevel `json:"access_level"`
	AuthPolicy  AuthPolicy  `json:"auth_policy"`
	AllowCIDRs  []string    `json:"allow_cidrs,omitempty"`
}

type PublicDomainAccessPolicy struct {
	DenyAll int `json:"deny_all"`
	// AllowedDomains []string `json:"allowed_domains"`
}

type ActivateRequest struct {
	Language string `json:"language"`
	Location string `json:"location"`
	Theme    string `json:"theme"`
	FRP      struct {
		Host string `json:"host"`
		JWS  string `json:"jws"`
	} `json:"frp"`
}

type TunnelRequest struct {
	Name    string `json:"name"`
	Service string `json:"service"`
}

type TunnelResponseData struct {
	Token string `json:"token"`
}

type TunnelResponse struct {
	Success bool                `json:"success"`
	Data    *TunnelResponseData `json:"data"`
}

func NewL4ProxyDeploymentApplyConfiguration(namespace, serviceAccountName string, port int) applyAppsv1.DeploymentApplyConfiguration {
	imagePullPolicy := corev1.PullIfNotPresent //corev1.PullAlways
	strategyRecreate := appsv1.RecreateDeploymentStrategyType
	dnsPolicy := corev1.DNSClusterFirstWithHostNet
	protocolTCP := corev1.ProtocolTCP
	containerPort := intstr.FromInt(port)
	nodeSelectorOperatorIn := corev1.NodeSelectorOpIn
	nodeSelectorOperatorExists := corev1.NodeSelectorOpExists

	imageVersion := utils.EnvOrDefault("L4_PROXY_IMAGE_VERSION", "v0.2.0")
	imageName := fmt.Sprintf("%s:%s", utils.EnvOrDefault("L4_PROXY_IMAGE_NAME", constants.L4ProxyImage), imageVersion)

	workerProcessesNum := strconv.Itoa(runtime.NumCPU() / 2)
	if workerProcessesNum == "" || workerProcessesNum == "0" {
		klog.Warning("get cpu num error")
		workerProcessesNum = "4"
	}

	return applyAppsv1.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("Deployment"),
			APIVersion: pointer.String(appsv1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String(L4ProxyDeploymentName),
			Namespace: pointer.String(namespace),
			Labels: map[string]string{
				"app": L4ProxyDeploymentName,
			},
			Annotations: nil,
		},
		Spec: &applyAppsv1.DeploymentSpecApplyConfiguration{
			Replicas: pointer.Int32(L4ProxyDeploymentReplicas),
			Strategy: &applyAppsv1.DeploymentStrategyApplyConfiguration{
				Type: &strategyRecreate,
			},
			Selector: &applyMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: map[string]string{
					"app": L4ProxyDeploymentName,
				},
			},
			Template: &applyCorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
					Labels: map[string]string{
						"app": L4ProxyDeploymentName,
					},
				},
				Spec: &applyCorev1.PodSpecApplyConfiguration{
					HostNetwork:        pointer.Bool(true),
					DNSPolicy:          &dnsPolicy,
					ServiceAccountName: pointer.String(serviceAccountName),
					PriorityClassName: func() *string {
						name := "system-cluster-critical"
						return &name
					}(),
					Affinity: &applyCorev1.AffinityApplyConfiguration{
						NodeAffinity: &applyCorev1.NodeAffinityApplyConfiguration{
							PreferredDuringSchedulingIgnoredDuringExecution: []applyCorev1.PreferredSchedulingTermApplyConfiguration{
								{
									Weight: pointer.Int32(10),
									Preference: &applyCorev1.NodeSelectorTermApplyConfiguration{
										MatchExpressions: []applyCorev1.NodeSelectorRequirementApplyConfiguration{
											{
												Key:      pointer.String("kubernetes.io/os"),
												Operator: &nodeSelectorOperatorIn,
												Values:   []string{"linux"},
											},
											{
												Key:      pointer.String("node-role.kubernetes.io/master"),
												Operator: &nodeSelectorOperatorExists,
											},
										},
									},
								},
							},
						},
					},
					Containers: []applyCorev1.ContainerApplyConfiguration{
						{
							Name:            pointer.String("proxy"),
							Image:           pointer.String(imageName),
							ImagePullPolicy: &imagePullPolicy,
							Command: []string{
								"/l4-bfl-proxy",
								"-w",
								workerProcessesNum,
							},
							Env: []applyCorev1.EnvVarApplyConfiguration{
								{
									Name: pointer.String("NODE_IP"),
									ValueFrom: &applyCorev1.EnvVarSourceApplyConfiguration{
										FieldRef: &applyCorev1.ObjectFieldSelectorApplyConfiguration{
											FieldPath: pointer.String("status.hostIP"),
										},
									},
								},
							},
							LivenessProbe: &applyCorev1.ProbeApplyConfiguration{
								HandlerApplyConfiguration: applyCorev1.HandlerApplyConfiguration{
									TCPSocket: &applyCorev1.TCPSocketActionApplyConfiguration{
										Port: &containerPort,
									},
								},
								FailureThreshold:    pointer.Int32(8),
								InitialDelaySeconds: pointer.Int32(3),
								PeriodSeconds:       pointer.Int32(5),
								TimeoutSeconds:      pointer.Int32(10),
							},
							ReadinessProbe: &applyCorev1.ProbeApplyConfiguration{
								HandlerApplyConfiguration: applyCorev1.HandlerApplyConfiguration{
									TCPSocket: &applyCorev1.TCPSocketActionApplyConfiguration{
										Port: &containerPort,
									},
								},
								FailureThreshold: pointer.Int32(5),
								PeriodSeconds:    pointer.Int32(3),
								TimeoutSeconds:   pointer.Int32(10),
							},
							Ports: []applyCorev1.ContainerPortApplyConfiguration{
								{
									ContainerPort: pointer.Int32(int32(port)),
									Protocol:      &protocolTCP,
								},
							},
							Resources: &applyCorev1.ResourceRequirementsApplyConfiguration{
								Limits: &corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("2"),
								},
								Requests: &corev1.ResourceList{
									corev1.ResourceCPU: resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
		},
	}
}
