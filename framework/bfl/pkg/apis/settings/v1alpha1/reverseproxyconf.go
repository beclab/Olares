package v1alpha1

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"time"

	appv1alpha1 "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"bytetrade.io/web3os/bfl/pkg/utils/certmanager"
	"bytetrade.io/web3os/bfl/pkg/utils/k8sutil"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	applyAppsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	applyCorev1 "k8s.io/client-go/applyconfigurations/core/v1"
	applyMetav1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ReverseProxyConfigurator struct {
	kubeClient   kubernetes.Interface
	ctrlClient   ctrlclient.Client
	userOp       *operator.UserOperator
	cm           certmanager.Interface
	user         *iamV1alpha2.User
	terminusName string
}

type ReverseProxyConfig struct {
	FRPConfig
	EnableCloudFlareTunnel bool `json:"enable_cloudflare_tunnel"`
	EnableFRP              bool `json:"enable_frp"`
}

type FRPConfig struct {
	FRPServer     string `json:"frp_server"`
	FRPPort       int    `json:"frp_port"`
	FRPAuthMethod string `json:"frp_auth_method"`
	FRPAuthToken  string `json:"frp_auth_token"`
}

var (
	FRPOptionServer     string = "server"
	FRPOptionPort       string = "port"
	FRPOptionAuthMethod string = "auth-method"
	FRPOptionAuthToken  string = "auth-token"
	FRPOptionUserName   string = "username"
	FRPAuthMethodJWS    string = "jws"
	FRPAuthMethodToken  string = "token"

	ReverseProxyConfigKeyPublicIP         = "public_ip"
	ReverseProxyConfigKeyCloudFlareEnable = "cloudflare.enable"
	ReverseProxyConfigKeyFRPEnable        = "frp.enable"
	ReverseProxyConfigValueEnabled        = "1"
	ReverseProxyConfigKeyFRPServer        = "frp.server"
	ReverseProxyConfigKeyFRPPort          = "frp.port"
	ReverseProxyConfigKeyFRPAuthMethod    = "frp.auth_method"
	ReverseProxyConfigKeyFRPAuthToken     = "frp.auth_token"
)

func NewReverseProxyConfigurator() (*ReverseProxyConfigurator, error) {
	userOp, err := operator.NewUserOperator()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user operator")
	}
	user, err := userOp.GetUser("")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	terminusName := userOp.GetTerminusName(user)
	if terminusName == "" {
		return nil, errors.New("olares name of user is empty")
	}
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get in-cluster config")
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get kubernetes client")
	}
	controllerClient, err := ctrlclient.New(restConfig, ctrlclient.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get controller client")
	}
	if err = appv1alpha1.AddToScheme(controllerClient.Scheme()); err != nil {
		return nil, errors.Wrap(err, "failed to add apps to scheme")
	}
	cm := certmanager.NewCertManager(constants.TerminusName(terminusName))
	return &ReverseProxyConfigurator{
		kubeClient:   kubeClient,
		ctrlClient:   controllerClient,
		userOp:       userOp,
		cm:           cm,
		user:         user,
		terminusName: terminusName,
	}, nil
}

func (conf *ReverseProxyConfig) Check() error {
	if conf == nil {
		return errors.New("nil ReverseProxyConfig")
	}
	if conf.EnableCloudFlareTunnel {
		if conf.EnableFRP {
			return errors.New("only one of public IP, FRP, or CloudFlare tunnel should be selected")
		}
		return nil
	}
	if conf.EnableFRP {
		if conf.FRPServer == "" {
			return errors.New("FRP server is not provided")
		}
		if conf.FRPAuthMethod == FRPAuthMethodToken && conf.FRPAuthToken == "" {
			return errors.New("FRP auth method is selected as token but no token is provided")
		}
		return nil
	}
	// as long as both FRP and cloudflare tunnel are not enabled,
	// consider it as a public IP
	// no matter whether the specific IP address is provided (it shouldn't have been provided by the frontend before in the first place)
	// if conf.IP == "" {
	// 	return errors.New("one of public IP, FRP, or CloudFlare tunnel should be selected")
	// }
	return nil
}

// configureDNS configures DNS records
// and also update the corresponding annotations on the user resource
func (configurator *ReverseProxyConfigurator) configureDNS(publicIP, localIP, publicCName string) error {
	var userPatches []func(*iamV1alpha2.User)
	if publicIP != "" {
		if err := configurator.cm.AddDNSRecord(&publicIP, nil); err != nil {
			return errors.Wrap(err, "failed to configure DNS record for public IP")
		}
		userPatches = append(userPatches, func(user *iamV1alpha2.User) {
			user.Annotations[constants.UserAnnotationPublicDomainIp] = publicIP
		})
	}
	natGatewayIP := configurator.userOp.GetUserAnnotation(configurator.user, constants.UserAnnotationNatGatewayIp)
	if natGatewayIP != "" {
		localIP = natGatewayIP
	}
	if localIP != "" {
		userPatches = append(userPatches, func(user *iamV1alpha2.User) {
			user.Annotations[constants.UserAnnotationLocalDomainIp] = localIP
			user.Annotations[constants.UserAnnotationLocalDomainDNSRecord] = localIP
		})
	}
	if publicCName != "" {
		if err := configurator.cm.AddDNSRecord(nil, &publicCName); err != nil {
			return errors.Wrap(err, "failed to configure DNS record for public CName")
		}
		userPatches = append(userPatches, func(user *iamV1alpha2.User) {
			user.Annotations[constants.UserAnnotationPublicDomainIp] = publicCName
		})
	}
	// switched from public IP or FRP to cloudflare tunnel
	if publicIP == "" && publicCName == "" {
		userPatches = append(userPatches, func(user *iamV1alpha2.User) {
			user.Annotations[constants.UserAnnotationPublicDomainIp] = ""
		})
	}
	return configurator.userOp.UpdateUser(configurator.user, userPatches)
}

func (configurator *ReverseProxyConfigurator) markApplying(ctx context.Context, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[constants.ReverseProxyStatusKey] = constants.ReverseProxyStatusApplying
	cmApply := &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:        pointer.String(constants.ReverseProxyConfigMapName),
			Namespace:   pointer.String(constants.Namespace),
			Annotations: cm.Annotations,
			Labels:      cm.Labels,
		},
		Data: cm.Data,
	}
	return configurator.kubeClient.CoreV1().ConfigMaps(constants.Namespace).Apply(
		ctx,
		cmApply,
		metav1.ApplyOptions{FieldManager: constants.ApplyPatchFieldManager},
	)
}

func (configurator *ReverseProxyConfigurator) markApplied(ctx context.Context, cm *corev1.ConfigMap, conf *ReverseProxyConfig) error {
	confStr, err := json.Marshal(conf)
	if err != nil {
		return errors.Wrap(err, "failed to marshal applied reverse proxy config")
	}
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[constants.ReverseProxyStatusKey] = constants.ReverseProxyStatusApplied
	cm.Annotations[constants.ReverseProxyLastAppliedConfigKey] = string(confStr)
	cmApply := &applyCorev1.ConfigMapApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("ConfigMap"),
			APIVersion: pointer.String(corev1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:        pointer.String(constants.ReverseProxyConfigMapName),
			Namespace:   pointer.String(constants.Namespace),
			Annotations: cm.Annotations,
			Labels:      cm.Labels,
		},
		Data: cm.Data,
	}
	_, err = configurator.kubeClient.CoreV1().ConfigMaps(constants.Namespace).Apply(
		ctx,
		cmApply,
		metav1.ApplyOptions{FieldManager: constants.ApplyPatchFieldManager},
	)
	return err
}

func (configurator *ReverseProxyConfigurator) setProxyTypeForRelevantComponents(ctx context.Context, proxyType string) error {
	if configurator.userOp.GetUserAnnotation(configurator.user, constants.UserAnnotationReverseProxyType) != proxyType {
		err := configurator.userOp.UpdateAnnotation(configurator.user, constants.UserAnnotationReverseProxyType, proxyType)
		if err != nil {
			return errors.Wrap(err, "failed to set reverse proxy type annotation to user")
		}
	}
	appList := &appv1alpha1.ApplicationList{}
	err := configurator.ctrlClient.List(ctx, appList)
	if err != nil {
		return errors.Wrap(err, "failed to list applications")
	}
	for _, app := range appList.Items {
		if app.Spec.Owner != constants.Username {
			continue
		}
		customDomainSettingsStr, ok := app.Spec.Settings[constants.ApplicationCustomDomain]
		if !ok || customDomainSettingsStr == "" {
			continue
		}
		customDomainSettings := make(map[string]map[string]string)
		err = json.Unmarshal([]byte(customDomainSettingsStr), &customDomainSettings)
		if err != nil {
			log.Errorf("failed to unmarshal custom domain settings of app %s: %w", app.Name, err)
			continue
		}
		for _, entry := range customDomainSettings {
			// if the app does not have a custom domain, a reverse proxy type switch wouldn't matter
			if thirdPartyDomain, ok := entry[constants.ApplicationThirdPartyDomain]; !ok || thirdPartyDomain == "" {
				continue
			}
			// if the app has already been synced with the same type of reverse proxy, e.g., a change of the FRP server
			// no need to sync again
			if entry[constants.ApplicationReverseProxyType] == proxyType {
				continue
			}
			entry[constants.ApplicationReverseProxyType] = proxyType
			entry[constants.ApplicationCustomDomainCnameStatus] = constants.CustomDomainCnameStatusPending
		}
		customDomainSettingsByte, err := json.Marshal(customDomainSettings)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal custom domain settings of app %s", app.Name)
		}
		app.Spec.Settings[constants.ApplicationCustomDomain] = string(customDomainSettingsByte)
		if err := configurator.ctrlClient.Update(ctx, &app); err != nil {
			return errors.Wrapf(err, "failed to update reverse proxy type to application %s", app.Name)
		}
	}
	return nil
}

func (configurator *ReverseProxyConfigurator) Configure(ctx context.Context) (err error) {
	configurator.user, err = configurator.userOp.GetUser("")
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}
	cm, err := configurator.kubeClient.CoreV1().ConfigMaps(constants.Namespace).Get(ctx, constants.ReverseProxyConfigMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("reverse proxy configmap not found, skip configuring")
			return nil
		}
		log.Error(err, "failed to get reverse proxy configmap")
		return errors.Wrap(err, "failed to get the configmap for reverse proxy config")
	}
	log.Infof("reverse proxy configmap found: %#v", cm)
	conf := &ReverseProxyConfig{}
	if err := conf.readFromReverseProxyConfigMapData(cm.Data); err != nil {
		return errors.Wrap(err, "failed to read reverse proxy configmap data")
	}
	shouldConfigure := func() bool {
		if err := conf.Check(); err != nil {
			log.Warn("invalid reverse proxy config, skip configuring", err)
			return false
		}
		if cm.Annotations == nil {
			return true
		}
		if status, ok := cm.Annotations[constants.ReverseProxyStatusKey]; !ok || status != constants.ReverseProxyStatusApplied {
			return true
		}
		lastAppliedConfigStr, ok := cm.Annotations[constants.ReverseProxyLastAppliedConfigKey]
		if !ok {
			return true
		}
		lastAppliedConfig := &ReverseProxyConfig{}
		if err := json.Unmarshal([]byte(lastAppliedConfigStr), lastAppliedConfig); err != nil {
			log.Error(err, "failed to marshal last applied config of reverse proxy, will try to configure again")
			return true
		}
		if !reflect.DeepEqual(conf, lastAppliedConfig) {
			return true
		}
		return false
	}()
	if !shouldConfigure {
		log.Info("current applied reverse proxy config is already the expected one")
		return nil
	}
	var proxyType, publicIP, localIP, publicCName string
	defer func() {
		if err != nil {
			return
		}
		err = configurator.setProxyTypeForRelevantComponents(ctx, proxyType)
		if err != nil {
			return
		}
		err = errors.Wrap(configurator.configureDNS(publicIP, localIP, publicCName), "failed to configure DNS")
		if err != nil {
			return
		}
		err = errors.Wrap(configurator.markApplied(ctx, cm, conf), "failed to update reverse proxy configmap to mark apply completed")
	}()
	cm, err = configurator.markApplying(ctx, cm)
	if err != nil {
		return errors.Wrap(err, "failed to mark reverse proxy configmap as applying")
	}
	localL4ProxyIP, err := k8sutil.GetL4ProxyNodeIP(ctx, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "failed to get local l4 proxy ip")
	}
	localIP = *localL4ProxyIP
	localPort := utils.EnvOrDefault("L4_PROXY_LISTEN", constants.L4ListenSSLPort)

	if !conf.EnableCloudFlareTunnel && !conf.EnableFRP {
		// the public IP DNS record is handled by the ip watching mechanism
		// delete the reverse proxy agent, if existing
		err := configurator.kubeClient.AppsV1().Deployments(constants.Namespace).Delete(ctx, ReverseProxyAgentDeploymentName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to delete existing reverse proxy agent")
		}
		proxyType = constants.ReverseProxyTypeNone
		return nil
	}

	reverseProxyDeployment := newDefaultReverseProxyAgentDeploymentApplyConfiguration()
	if conf.EnableFRP {
		if net.ParseIP(conf.FRPServer) != nil {
			publicIP = conf.FRPServer
		} else {
			publicCName = conf.FRPServer
		}
		setReverseProxyAgentDeploymentToFRP(reverseProxyDeployment, conf.FRPConfig)
		proxyType = constants.ReverseProxyTypeFRP
	} else if conf.EnableCloudFlareTunnel {
		// get cloudflare token
		jws := configurator.userOp.GetUserAnnotation(configurator.user, constants.UserCertManagerJWSToken)
		if jws == "" {
			return errors.New("no jws token found in user annotation")
		}
		req := TunnelRequest{
			Name:    configurator.terminusName,
			Service: fmt.Sprintf("https://%s:%s", localIP, localPort),
		}
		res, err := resty.New().SetTimeout(30 * time.Second).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).R().
			SetHeaders(map[string]string{
				restful.HEADER_ContentType: restful.MIME_JSON,
				restful.HEADER_Accept:      restful.MIME_JSON,
				"Authorization":            "Bearer " + jws,
			}).
			SetBody(req).
			SetResult(&TunnelResponse{}).
			Post(constants.APIDNSSetCloudFlareTunnel)
		if err != nil {
			return errors.Wrap(err, "failed to request cloudflare tunnel api")
		}
		if res.StatusCode() != http.StatusOK {
			return fmt.Errorf("error response from cloudflare tunnel api: %s", res.Body())
		}
		responseData := res.Result().(*TunnelResponse)
		if !responseData.Success || responseData.Data == nil || responseData.Data.Token == "" {
			return fmt.Errorf("error response from cloudflare tunnel api: %v", responseData)
		}
		setReverseProxyAgentDeploymentToCloudFlare(reverseProxyDeployment, responseData.Data.Token)
		proxyType = constants.ReverseProxyTypeCloudflare
	}
	_, err = configurator.kubeClient.AppsV1().Deployments(constants.Namespace).Apply(ctx,
		reverseProxyDeployment, metav1.ApplyOptions{Force: true, FieldManager: constants.ApplyPatchFieldManager})
	if err != nil {
		return errors.Wrap(err, "failed to apply reverse proxy agent")
	}
	return nil
}

func newDefaultReverseProxyAgentDeploymentApplyConfiguration() *applyAppsv1.DeploymentApplyConfiguration {
	imageVersion := utils.EnvOrDefault("REVERSE_PROXY_AGENT_IMAGE_VERSION", constants.ReverseProxyAgentImageVersion)
	imageName := fmt.Sprintf("%s:%s", utils.EnvOrDefault("REVERSE_PROXY_AGENT_IMAGE_NAME", constants.ReverseProxyAgentImage), imageVersion)
	imagePullPolicy := corev1.PullIfNotPresent

	return &applyAppsv1.DeploymentApplyConfiguration{
		TypeMetaApplyConfiguration: applyMetav1.TypeMetaApplyConfiguration{
			Kind:       pointer.String("Deployment"),
			APIVersion: pointer.String(appsv1.SchemeGroupVersion.String()),
		},
		ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
			Name:      pointer.String(ReverseProxyAgentDeploymentName),
			Namespace: pointer.String(constants.Namespace),
			Labels: map[string]string{
				"app":                                  ReverseProxyAgentDeploymentName,
				"applications.app.bytetrade.io/author": constants.AnnotationGroup,
				"applications.app.bytetrade.io/owner":  constants.Username,
			},
		},
		Spec: &applyAppsv1.DeploymentSpecApplyConfiguration{
			Replicas: pointer.Int32(ReverseProxyAgentDeploymentReplicas),

			Selector: &applyMetav1.LabelSelectorApplyConfiguration{
				MatchLabels: map[string]string{
					"app": ReverseProxyAgentDeploymentName,
				},
			},
			Template: &applyCorev1.PodTemplateSpecApplyConfiguration{
				ObjectMetaApplyConfiguration: &applyMetav1.ObjectMetaApplyConfiguration{
					Labels: map[string]string{
						"app": ReverseProxyAgentDeploymentName,
					},
				},
				Spec: &applyCorev1.PodSpecApplyConfiguration{
					PriorityClassName: func() *string {
						name := "system-cluster-critical"
						return &name
					}(),
					SchedulerName:      pointer.String("default-scheduler"),
					ServiceAccountName: pointer.String("bytetrade-controller"),
					Containers: []applyCorev1.ContainerApplyConfiguration{
						{
							Name:            pointer.String("agent"),
							Image:           pointer.String(imageName),
							ImagePullPolicy: &imagePullPolicy,
						},
					},
				},
			},
		},
	}
}

func setEnvToReverseProxyAgentDeployment(deployConf *applyAppsv1.DeploymentApplyConfiguration, key, val string) {
	deployConf.Spec.Template.Spec.Containers[0].Env = append(
		deployConf.Spec.Template.Spec.Containers[0].Env,
		applyCorev1.EnvVarApplyConfiguration{Name: &key, Value: &val})
}

func setArgsToReverseProxyAgentDeployment(deployConf *applyAppsv1.DeploymentApplyConfiguration, args []string) {
	deployConf.Spec.Template.Spec.Containers[0].Args = args
}

func setReverseProxyAgentDeploymentToFRP(deployConf *applyAppsv1.DeploymentApplyConfiguration, frpConf FRPConfig) {
	setEnvToReverseProxyAgentDeployment(deployConf, ReverseProxyAgentSelectEnvKey, ReverseProxyAgentSelectFRPEnvVal)
	args := []string{utils.DashedOption(FRPOptionServer), frpConf.FRPServer, utils.DashedOption(FRPOptionUserName), constants.Username}
	if frpConf.FRPPort != 0 {
		args = append(args, utils.DashedOption(FRPOptionPort), strconv.Itoa(frpConf.FRPPort))
	}
	if frpConf.FRPAuthMethod != "" {
		args = append(args, utils.DashedOption(FRPOptionAuthMethod), frpConf.FRPAuthMethod)
		if frpConf.FRPAuthMethod == FRPAuthMethodToken {
			args = append(args, utils.DashedOption(FRPOptionAuthToken), frpConf.FRPAuthToken)
		}
	}
	setArgsToReverseProxyAgentDeployment(deployConf, args)
}

func setReverseProxyAgentDeploymentToCloudFlare(deployConf *applyAppsv1.DeploymentApplyConfiguration, token string) {
	setEnvToReverseProxyAgentDeployment(deployConf, ReverseProxyAgentSelectEnvKey, ReverseProxyAgentSelectCloudFlareEnvVal)
	setArgsToReverseProxyAgentDeployment(deployConf, []string{"tunnel", "run", "--token", token})
}

func (conf *ReverseProxyConfig) readFromReverseProxyConfigMapData(cmData map[string]string) error {
	if cmData[ReverseProxyConfigKeyCloudFlareEnable] == ReverseProxyConfigValueEnabled {
		conf.EnableCloudFlareTunnel = true
		// don't break circuit here or at the above public ip logic
		// because the validity check will be done by the configurator
		// this method is only meant for parsing
	}
	if cmData[ReverseProxyConfigKeyFRPEnable] == ReverseProxyConfigValueEnabled {
		conf.EnableFRP = true
		conf.FRPServer = cmData[ReverseProxyConfigKeyFRPServer]
		if portStr := cmData[ReverseProxyConfigKeyFRPPort]; portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return errors.Wrapf(err, "invalid frp port %s", portStr)
			}
			conf.FRPPort = port
		}
		conf.FRPAuthMethod = cmData[ReverseProxyConfigKeyFRPAuthMethod]
		conf.FRPAuthToken = cmData[ReverseProxyConfigKeyFRPAuthToken]
	}
	return nil
}

func (conf *ReverseProxyConfig) generateReverseProxyConfigMapData() map[string]string {
	cmData := make(map[string]string)
	if conf.EnableCloudFlareTunnel {
		cmData[ReverseProxyConfigKeyCloudFlareEnable] = ReverseProxyConfigValueEnabled
	}
	if conf.EnableFRP {
		cmData[ReverseProxyConfigKeyFRPEnable] = ReverseProxyConfigValueEnabled
		cmData[ReverseProxyConfigKeyFRPServer] = conf.FRPServer
		if conf.FRPPort != 0 {
			cmData[ReverseProxyConfigKeyFRPPort] = strconv.Itoa(conf.FRPPort)
		}
		cmData[ReverseProxyConfigKeyFRPAuthMethod] = conf.FRPAuthMethod
		cmData[ReverseProxyConfigKeyFRPAuthToken] = conf.FRPAuthToken
	}
	return cmData
}

func GetReverseProxyConfig(ctx context.Context) (*ReverseProxyConfig, error) {
	return GetReverseProxyConfigFromNamespace(ctx, constants.Namespace)
}

func GetReverseProxyConfigFromNamespace(ctx context.Context, namespace string) (*ReverseProxyConfig, error) {
	configData, err := k8sutil.GetConfigMapData(ctx, namespace, constants.ReverseProxyConfigMapName)
	if err != nil {
		return nil, errors.Wrap(err, "error getting configmap")
	}
	conf := &ReverseProxyConfig{}
	if err := conf.readFromReverseProxyConfigMapData(configData); err != nil {
		return nil, errors.Wrap(err, "error parsing reverse proxy config data")
	}
	return conf, nil
}

func (conf *ReverseProxyConfig) writeToReverseProxyConfigMap(ctx context.Context) error {
	cmData := conf.generateReverseProxyConfigMapData()
	err := k8sutil.WriteConfigMapData(ctx, constants.Namespace, constants.ReverseProxyConfigMapName, cmData)
	return errors.Wrap(err, "error writing configmap")
}
