package v1alpha1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	appv1 "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/internal/log"
	"bytetrade.io/web3os/bfl/pkg/api"
	"bytetrade.io/web3os/bfl/pkg/api/response"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apiserver/runtime"
	"bytetrade.io/web3os/bfl/pkg/app_service/v1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"bytetrade.io/web3os/bfl/pkg/utils/certmanager"

	iamV1alpha2 "github.com/beclab/api/iam/v1alpha2"
	"github.com/emicklei/go-restful/v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func (h *Handler) setupAppPolicy(req *restful.Request, resp *restful.Response) {
	appname := req.PathParameter(ParamAppName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	var policy app_service.ApplicationSettingsPolicy
	err := req.ReadEntity(&policy)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	settings := app_service.ApplicationsSettings{
		app_service.ApplicationSettingsPolicyKey: policy,
	}

	appServiceClient := app_service.NewAppServiceClient()

	ret, err := appServiceClient.SetupAppPolicy(appname, token, settings)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, ret)
}

func (h *Handler) getAppPolicy(req *restful.Request, resp *restful.Response) {
	appname := req.PathParameter(ParamAppName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	appServiceClient := app_service.NewAppServiceClient()

	settings, err := appServiceClient.GetAppPolicy(appname, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	policyStr, ok := settings[app_service.ApplicationSettingsPolicyKey]

	if !ok {
		response.HandleNotFound(resp, errors.New("no policy"))
		return
	}

	var policy map[string]interface{}
	err = json.Unmarshal([]byte(policyStr.(string)), &policy)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, policy)
}

func (h *Handler) handleCertConfig(ctx context.Context, customDomain map[string]interface{}) error {
	cert, key, thirdPartyDomain := customDomain[appv1.AppEntranceCertConfigMapCertKey], customDomain[appv1.AppEntranceCertConfigMapKeyKey], customDomain[constants.ApplicationThirdPartyDomain]
	if cert == nil || key == nil || thirdPartyDomain == nil {
		log.Infof("skip storing empty cert config, cert: %s, key: %s, thirdPartyDomain: %s", cert, key, thirdPartyDomain)
		return nil
	}
	var certData, keyData, zoneData string
	certData, ok := cert.(string)
	if !ok {
		return nil
	}
	keyData, ok = key.(string)
	if !ok {
		return nil
	}
	zoneData, ok = thirdPartyDomain.(string)
	if !ok {
		return nil
	}
	if len(certData) == 0 || len(keyData) == 0 || len(zoneData) == 0 {
		log.Infof("skip storing empty cert config, cert: %s, key: %s, thirdPartyDomain: %s", cert, key, thirdPartyDomain)
		return nil
	}
	configMapName := fmt.Sprintf(appv1.AppEntranceCertConfigMapNameTpl, zoneData)

	client, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return err
	}
	err = client.Kubernetes().CoreV1().ConfigMaps(constants.Namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: constants.Namespace,
			Labels: map[string]string{
				appv1.AppEntranceCertConfigMapLabel: "true",
			},
		},
		Data: map[string]string{
			appv1.AppEntranceCertConfigMapKeyKey:  keyData,
			appv1.AppEntranceCertConfigMapCertKey: certData,
			appv1.AppEntranceCertConfigMapZoneKey: zoneData,
		},
	}
	_, err = client.Kubernetes().CoreV1().ConfigMaps(constants.Namespace).Create(ctx, configMap, metav1.CreateOptions{})
	return err
}

func (h *Handler) checkCertExists(ctx context.Context, domainName string) error {
	configMapName := fmt.Sprintf(appv1.AppEntranceCertConfigMapNameTpl, domainName)
	client, err := runtime.NewKubeClientInCluster()
	if err != nil {
		return err
	}
	_, err = client.Kubernetes().CoreV1().ConfigMaps(constants.Namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return errors.New("current reverse proxy mode is FRP, the HTTPS certificate and key for custom domain must be uploaded")
		}
		return fmt.Errorf("failed to check existence of HTTPS cert config map: %v", err)
	}
	return nil
}

func (h *Handler) setupAppCustomDomain(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	entranceName := req.PathParameter(ParamEntranceName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	customDomain := make(map[string]interface{})
	data, err := io.ReadAll(req.Request.Body)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	err = json.Unmarshal(data, &customDomain)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	var settings app_service.ApplicationsSettings
	appServiceClient := app_service.NewAppServiceClient()

	var terminusName, zone string
	terminusName, zone, err = h.getUserInfo()
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	formatSettings := func(customDomainStore map[string]interface{}, needTargetCname, thirdLevelDomainName, thirdPartyDomainName string) {
		customDomainStore[constants.ApplicationCustomDomainCnameTarget] = needTargetCname
		customDomainStore[constants.ApplicationCustomDomainCnameTargetStatus] = ""
		customDomainStore[constants.ApplicationCustomDomainCnameStatus] = ""
		if thirdLevelDomainName != "" || thirdPartyDomainName != "" {
			customDomainStore[constants.ApplicationThirdLevelDomain] = thirdLevelDomainName
			customDomainStore[constants.ApplicationThirdPartyDomain] = thirdPartyDomainName
		}
	}

	var reqCustomDomain = h.getCustomDomainValue(customDomain, constants.ApplicationThirdPartyDomain)
	if reqCustomDomain != "" {
		domainErrs := validation.IsFullyQualifiedDomainName(field.NewPath("domain"), reqCustomDomain)
		if len(domainErrs) > 0 {
			response.HandleError(resp, domainErrs.ToAggregate())
			return
		}
		authLevel, err := h.getEntranceAuthLevel(appServiceClient, appName, entranceName, token)
		if err != nil {
			response.HandleError(resp, err)
			return
		}
		if authLevel == "private" {
			response.HandleError(resp, errors.New("custom domain can not be set when auth level is private"))
			return
		}
		if err := h.handleCertConfig(req.Request.Context(), customDomain); err != nil {
			response.HandleError(resp, err)
			return
		}
	}

	existsAppCustomDomain, err := h.getExistsCustomDomain(appServiceClient, appName, entranceName, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	cm := certmanager.NewCertManager(constants.TerminusName(terminusName))

	var operate = h.getCustomDomainOperation(reqCustomDomain, existsAppCustomDomain)
	log.Infof("setAppCustomDomain: app: %s-%s, reqDomain: %s, existsDomain: %s, operate: %d, req: %s",
		appName, entranceName, reqCustomDomain, existsAppCustomDomain, operate, utils.ToJSON(customDomain))

	switch operate {
	case constants.CustomDomainIgnore, constants.CustomDomainCheck:
		ret, _ := appServiceClient.GetAppCustomDomain(appName, token)
		entranceCustomDomainStr, ok := ret[constants.ApplicationCustomDomain]
		if !ok {
			formatSettings(customDomain, zone, "", "")
			break
		}
		var entrancesCustomDomainM = make(map[string]interface{})
		err := json.Unmarshal([]byte(entranceCustomDomainStr.(string)), &entrancesCustomDomainM)
		if err != nil {
			response.HandleError(resp, err)
			return
		}
		entranceCustomDomainM, ok := entrancesCustomDomainM[entranceName]
		if !ok {
			formatSettings(customDomain, zone, "", "")
		} else {
			entranceCustomDomainMap := entranceCustomDomainM.(map[string]interface{})
			cnameTarget := h.getCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTarget)
			reqThirdLevel := h.getCustomDomainValue(customDomain, constants.ApplicationThirdLevelDomain)
			existsThirdLevel := h.getCustomDomainValue(entranceCustomDomainMap, constants.ApplicationThirdLevelDomain)
			if cnameTarget == "" {
				h.setCustomDomainValue(entranceCustomDomainMap, constants.ApplicationCustomDomainCnameTarget, zone)
			}
			if existsThirdLevel != reqThirdLevel {
				h.setCustomDomainValue(entranceCustomDomainMap, constants.ApplicationThirdLevelDomain, reqThirdLevel)
			}
			customDomain = entranceCustomDomainMap
		}
	case constants.CustomDomainDelete, constants.CustomDomainUpdate:
		_, err := cm.DeleteCustomDomainOnCloudflare(existsAppCustomDomain)
		if err != nil {
			log.Errorf("setAppCustomDomain: app: %s-%s, delete custom domain error %v", appName, entranceName, err)
			response.HandleError(resp, err)
			return
		}
		fallthrough
	case constants.CustomDomainAdd:
		formatSettings(customDomain, zone, "", "")
		if operate == constants.CustomDomainUpdate || operate == constants.CustomDomainAdd {
			op, err := operator.NewUserOperator()
			if err != nil {
				response.HandleError(resp, fmt.Errorf("create user operator failed: %v", err))
			}
			reverseProxyType, err := op.GetReverseProxyType()
			if err != nil {
				response.HandleError(resp, fmt.Errorf("get reverse proxy type failed: %v", err))
			}
			if reverseProxyType == constants.ReverseProxyTypeFRP {
				err := h.checkCertExists(req.Request.Context(), reqCustomDomain)
				if err != nil {
					response.HandleError(resp, err)
					return
				}
			} else {
				log.Infof("reverse proxy type: %s, skip custom domain cert check", reverseProxyType)
			}
			h.setCustomDomainValue(customDomain, constants.ApplicationReverseProxyType, reverseProxyType)
			h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTargetStatus, constants.CustomDomainCnameStatusNotset)
			h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameStatus, constants.CustomDomainCnameStatusNotset)
		}
	}

	settings = app_service.ApplicationsSettings{
		app_service.ApplicationSettingsDomainKey: customDomain,
	}
	ret, err := appServiceClient.SetupAppCustomDomain(appName, entranceName, token, settings)
	log.Infof("setAppCustomDomain: app: %s-%s, ret: %s", appName, entranceName, utils.ToJSON(ret))
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, ret)
}

func (h *Handler) listEntrancesWithCustomDomain(req *restful.Request, resp *restful.Response) {
	config, err := ctrl.GetConfig()
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	client, err := ctrlclient.New(config, ctrlclient.Options{Scheme: k8sruntime.NewScheme()})
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	err = appv1.AddToScheme(client.Scheme())
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	appList := &appv1.ApplicationList{}
	err = client.List(req.Request.Context(), appList)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	var entrances app_service.EntrancesWithCustomDomain
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
		for _, entrance := range app.Spec.Entrances {
			entranceSetting := customDomainSettings[entrance.Name]
			if entranceSetting == nil {
				continue
			}
			thirdPartyDomain, ok := entranceSetting[constants.ApplicationThirdPartyDomain]
			if !ok || thirdPartyDomain == "" {
				continue
			}
			entrances = append(entrances, app_service.EntranceWithCustomDomain{
				Entrance:     entrance,
				AppName:      app.Name,
				CustomDomain: thirdPartyDomain,
			})
		}
	}
	response.Success(resp, entrances)
}

func (h *Handler) finishAppCustomDomainCnameTarget(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	entranceName := req.PathParameter(ParamEntranceName)
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)
	var err error

	var settings app_service.ApplicationsSettings
	appServiceClient := app_service.NewAppServiceClient()
	ret, _ := appServiceClient.GetAppCustomDomain(appName, token)
	if ret == nil {
		response.HandleError(resp, errors.New("app custom domain not found"))
		return
	}

	entranceCustomDomainStr, ok := ret[constants.ApplicationCustomDomain]
	if !ok {
		response.HandleError(resp, errors.New("app custom domain not found"))
		return
	}

	var entrancesCustomDomainM = make(map[string]interface{})
	if err = json.Unmarshal([]byte(entranceCustomDomainStr.(string)), &entrancesCustomDomainM); err != nil {
		response.HandleError(resp, err)
		return
	}

	entranceCustomDomainM, ok := entrancesCustomDomainM[entranceName]
	if !ok {
		response.HandleError(resp, errors.New("app custom domain not found"))
		return
	}

	entranceCustomDomainMap := entranceCustomDomainM.(map[string]interface{})
	thirdPartyValue := entranceCustomDomainMap[constants.ApplicationThirdPartyDomain]
	if thirdPartyValue == nil {
		response.HandleError(resp, errors.New("app not set custom domain"))
		return
	}

	if thirdPartyValue.(string) == "" {
		response.HandleError(resp, errors.New("app not set custom domain"))
		return
	}

	entranceCustomDomainMap[constants.ApplicationCustomDomainCnameTargetStatus] = constants.CustomDomainCnameStatusSet
	entranceCustomDomainMap[constants.ApplicationCustomDomainCnameStatus] = constants.CustomDomainCnameStatusPending

	settings = app_service.ApplicationsSettings{
		app_service.ApplicationSettingsDomainKey: entranceCustomDomainMap,
	}
	r, err := appServiceClient.SetupAppCustomDomain(appName, entranceName, token, settings)
	log.Infof("finishAppCustomDomainCnameTarget: app: %s-%s, ret: %s", appName, entranceName, utils.ToJSON(r))
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, nil)
}

func (h *Handler) getAppCustomDomain(req *restful.Request, resp *restful.Response) {
	appname := req.PathParameter(ParamAppName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	appServiceClient := app_service.NewAppServiceClient()
	entrances, err := appServiceClient.GetAppEntrances(appname, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	if len(entrances) == 0 {
		response.HandleError(resp, fmt.Errorf("app %s entrances not found", appname))
		return
	}

	var zone string
	_, zone, err = h.getUserInfo()
	if err != nil {
		klog.Errorf("getAppCustomDomain: get user info error %v", err)
		response.HandleError(resp, err)
		return
	}

	settings, err := appServiceClient.GetAppCustomDomain(appname, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	getEntrancesDefaultResp := func() map[string]interface{} {
		var res = make(map[string]interface{})
		for _, entrancemap := range entrances {
			var r = make(map[string]interface{})
			var ename = entrancemap["name"].(string)
			r[constants.ApplicationThirdLevelDomain] = ""
			r[constants.ApplicationThirdPartyDomain] = ""
			r[constants.ApplicationCustomDomainCnameTarget] = zone
			r[constants.ApplicationCustomDomainCnameStatus] = ""
			r[constants.ApplicationCustomDomainCnameTargetStatus] = ""
			res[ename] = r
		}
		return res
	}

	appDomain, ok := settings[app_service.ApplicationSettingsDomainKey]
	if !ok {
		response.Success(resp, getEntrancesDefaultResp())
		return
	}

	var appDomainMap = make(map[string]interface{})
	if err = json.Unmarshal([]byte(appDomain.(string)), &appDomainMap); err != nil {
		log.Errorf("getAppCustomDomain: unmarshal customDomain error %v, %s", err, appDomain.(string))
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, appDomainMap)
}

func (h *Handler) getAppEntrances(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	appServiceClient := app_service.NewAppServiceClient()

	entrances, err := appServiceClient.GetAppEntrances(appName, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, api.NewListResult(entrances))
}

func (h *Handler) setupAppAuthorizationLevel(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	entranceName := req.PathParameter(ParamEntranceName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	authLevel := make(map[string]interface{})
	data, err := io.ReadAll(req.Request.Body)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	defer req.Request.Body.Close()
	err = json.Unmarshal(data, &authLevel)
	if err != nil {
		klog.Errorf("setupAppAuthorizationLevel: unmarshal authLevel error %v, %s", err, data)
		response.HandleError(resp, err)
		return
	}

	settings := app_service.ApplicationsSettings{
		app_service.ApplicationAuthorizationLevelKey: authLevel,
	}

	appServiceClient := app_service.NewAppServiceClient()

	ret, err := appServiceClient.SetupAppAuthorizationLevel(appName, entranceName, token, settings)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, ret)
}

func (h *Handler) setupAppEntrancePolicy(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)
	entranceName := req.PathParameter(ParamEntranceName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	var policy app_service.ApplicationSettingsPolicy
	err := req.ReadEntity(&policy)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	settings := app_service.ApplicationsSettings{
		app_service.ApplicationSettingsPolicyKey: policy,
	}

	appServiceClient := app_service.NewAppServiceClient()

	ret, err := appServiceClient.SetupAppEntrancePolicy(appName, entranceName, token, settings)
	if err != nil {
		response.HandleError(resp, err)
		return
	}

	response.Success(resp, ret)
}

func (h *Handler) getUserInfo() (string, string, error) {
	var terminusName string
	var zone string
	var err error
	var op *operator.UserOperator
	var user *iamV1alpha2.User

	op, err = operator.NewUserOperator()
	if err != nil {
		return "", "", fmt.Errorf("new user operator: %v", err)
	}

	user, err = op.GetUser(constants.Username)
	if err != nil {
		return "", "", fmt.Errorf("new user operator, and get user err: %v", err)
	}

	terminusName = op.GetTerminusName(user)
	if terminusName == "" {
		return "", "", fmt.Errorf("no olares naame has binding")
	}

	zone = op.GetUserZone(user)
	if zone == "" {
		return "", "", fmt.Errorf("no zone has binding")
	}

	return terminusName, zone, nil
}

func (h *Handler) getCustomDomainOperation(_reqCustomDomain, _existsAppCustomDomain string) constants.CustomDomain {
	switch {
	case _reqCustomDomain != "" && _existsAppCustomDomain == "":
		return constants.CustomDomainAdd
	case _reqCustomDomain != "" && _existsAppCustomDomain != "" && _reqCustomDomain != _existsAppCustomDomain:
		return constants.CustomDomainUpdate
	case _reqCustomDomain == "" && _existsAppCustomDomain != "":
		return constants.CustomDomainDelete
	case _reqCustomDomain != "" && _reqCustomDomain == _existsAppCustomDomain:
		return constants.CustomDomainCheck
	case _reqCustomDomain == "" && _reqCustomDomain == _existsAppCustomDomain:
		fallthrough
	default:
		return constants.CustomDomainIgnore
	}
}

func (h *Handler) getEntranceAuthLevel(appServiceClient *app_service.Client, appName, entranceName, token string) (string, error) {
	entrances, err := appServiceClient.GetAppEntrances(appName, token)
	if err != nil {
		return "", err
	}
	var ret string
	for _, entrance := range entrances {
		if name, ok := entrance["name"].(string); ok && name == entranceName {
			if authLevel, ok := entrance["authLevel"].(string); ok {
				ret = authLevel
			}
		}
	}
	return ret, nil
}

func (h *Handler) getExistsCustomDomain(appServiceClient *app_service.Client, appName, entranceName, token string) (string, error) {
	appCustomDomainExists, err := appServiceClient.GetAppCustomDomain(appName, token)
	if err != nil {
		log.Errorf("app found error: %+v", err)
		return "", err
	}

	if appCustomDomainExists == nil {
		return "", fmt.Errorf("app %s not found", appName)
	}

	existsAppCustomDomain, ok := appCustomDomainExists[constants.ApplicationCustomDomain]
	if !ok {
		return "", nil
	}
	var custdomDomainEntrances = make(map[string]map[string]string)
	if err = json.Unmarshal([]byte(existsAppCustomDomain.(string)), &custdomDomainEntrances); err != nil {
		return "", nil
	}

	custdomDomainEntrance, ok := custdomDomainEntrances[entranceName]
	if !ok {
		return "", nil
	}
	return custdomDomainEntrance[constants.ApplicationThirdPartyDomain], nil
}

func (h *Handler) updateAppCustomDomain(cm certmanager.Interface, entranceName string, customDomain map[string]interface{}) map[string]interface{} {
	var thirdParty = h.getCustomDomainValue(customDomain, constants.ApplicationThirdPartyDomain)
	if thirdParty == "" {
		return nil
	}

	cnameStatus, err := cm.GetCustomDomainCnameStatus(thirdParty)
	if err != nil {
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTargetStatus, constants.CustomDomainCnameStatusNotset)
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameStatus, constants.CustomDomainCnameStatusNotset)
		return customDomain
	}
	log.Infof("reloadAppCustomDomain: get cname status %v, entranceName: %s", cnameStatus.Success, entranceName)
	if !cnameStatus.Success {
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTargetStatus, constants.CustomDomainCnameStatusNotset)
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameStatus, constants.CustomDomainCnameStatusNotset)
		return customDomain
	}

	getStatus, _ := cm.GetCustomDomainOnCloudflare(thirdParty)
	if getStatus != nil {
		log.Infof("reloadAppCustomDomain: get ssl status ssl: %s  hostname: %s, entranceName: %s", getStatus.SSLStatus, getStatus.HostnameStatus, entranceName)
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTargetStatus, constants.CustomDomainCnameStatusSet)
		h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameStatus, h.setCustomDomainCnameStatus(getStatus.SSLStatus, getStatus.HostnameStatus))
		return customDomain
	}

	_, err = cm.AddCustomDomainOnCloudflare(thirdParty)
	if err != nil {
		log.Errorf("reloadAppCustomDomain: add custom domain error %v", err)
	}

	h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameTargetStatus, constants.CustomDomainCnameStatusSet)
	h.setCustomDomainValue(customDomain, constants.ApplicationCustomDomainCnameStatus, constants.CustomDomainCnameStatusPending)

	return customDomain
}

func (h *Handler) getCustomDomainValue(data map[string]interface{}, key string) string {
	v, ok := data[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (h *Handler) setCustomDomainValue(data map[string]interface{}, key string, val string) {
	if data == nil {
		return
	}
	_, ok := data[key]
	if ok {
		data[key] = val
	}
}

func (h *Handler) setCustomDomainCnameStatus(sslStatus, hostnameStatus string) string {
	switch {
	case hostnameStatus == sslStatus && sslStatus == "active":
		return constants.CustomDomainCnameStatusActive
	default:
		return constants.CustomDomainCnameStatusPending
	}
}

func (h *Handler) applicationPermissionList(req *restful.Request, resp *restful.Response) {
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	appServiceClient := app_service.NewAppServiceClient()

	aps, err := appServiceClient.GetApplicationPermissionList(token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, api.NewListResult(aps))
}

func (h *Handler) getApplicationProviderList(req *restful.Request, resp *restful.Response) {
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)
	appName := req.PathParameter(ParamAppName)

	appServiceClient := app_service.NewAppServiceClient()

	aps, err := appServiceClient.GetApplicationProviderList(appName, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, api.NewListResult(aps))
}

func (h *Handler) getApplicationSubjectList(req *restful.Request, resp *restful.Response) {
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)
	appName := req.PathParameter(ParamAppName)

	appServiceClient := app_service.NewAppServiceClient()

	aps, err := appServiceClient.GetApplicationSubjectList(appName, token)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, api.NewListResult(aps))
}

func (h *Handler) applicationPermission(req *restful.Request, resp *restful.Response) {
	appName := req.PathParameter(ParamAppName)

	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	appServiceClient := app_service.NewAppServiceClient()

	ap, err := appServiceClient.GetApplicationPermission(token, appName)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, ap)
}

func (h *Handler) getProviderRegistry(req *restful.Request, resp *restful.Response) {
	// fetch token from request
	token := req.Request.Header.Get(constants.UserAuthorizationTokenKey)

	dataType := req.PathParameter(ParamDataType)
	group := req.PathParameter(ParamGroup)
	version := req.PathParameter(ParamVersion)

	appServiceClient := app_service.NewAppServiceClient()

	pr, err := appServiceClient.GetProviderRegistry(token, dataType, group, version)
	if err != nil {
		response.HandleError(resp, err)
		return
	}
	response.Success(resp, pr)
}
