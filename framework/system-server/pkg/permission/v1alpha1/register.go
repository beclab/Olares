package permission

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	sysv1alpha1 "bytetrade.io/web3os/system-server/pkg/apis/sys/v1alpha1"
	"bytetrade.io/web3os/system-server/pkg/apiserver/v1alpha1/api"
	"bytetrade.io/web3os/system-server/pkg/constants"
	sysclientset "bytetrade.io/web3os/system-server/pkg/generated/clientset/versioned"
	prodiverregistry "bytetrade.io/web3os/system-server/pkg/providerregistry/v1alpha1"
	serviceproxy "bytetrade.io/web3os/system-server/pkg/serviceproxy/v1alpha1"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	MODULE_TAGS  = []string{"permission-control"}
	MODULE_ROUTE = "/permission/v1alpha1"
)

func AddPermissionControlToContainer(c *restful.Container,
	ctrlSet *PermissionControlSet,
	kubeconfig *rest.Config,
) error {
	handler := newHandler(ctrlSet, kubeconfig)

	ws := newWebService()
	ws.Route(ws.GET("/nonce").
		To(handler.nonce).
		Doc("get backend request call nonce").
		Metadata(restfulspec.KeyOpenAPITags, MODULE_TAGS).
		Returns(http.StatusOK, "Success to get nonce", ""))

	c.Add(ws)

	return nil
}

func newWebService() *restful.WebService {
	webservice := restful.WebService{}

	webservice.Path(MODULE_ROUTE).
		Produces(restful.MIME_JSON)

	return &webservice
}

func ValidateAccessTokenWithRequest(token string, op string, req *restful.Request, ctrlSet *PermissionControlSet) (string, error) {
	datatype := req.PathParameter(api.ParamDataType)
	version := req.PathParameter(api.ParamVersion)
	group := req.PathParameter(api.ParamGroup)

	return ValidateAccessToken(token, op, datatype, version, group, ctrlSet)
}

func ValidateAccessToken(token string, op, datatype, version, group string, ctrlSet *PermissionControlSet) (string, error) {
	permReq, err := ctrlSet.Mgr.getPermWithToken(token)
	if err != nil {
		return "", err
	}

	accReq := sysv1alpha1.PermissionRequire{
		Group:    group,
		DataType: datatype,
		Version:  version,
		Ops: []string{
			op,
		},
	}

	if permReq.Include(&accReq, true) {
		return permReq.AppKey, nil
	}

	return "", errors.New("data access denied")
}

func ValidateAppKeyWithRequest(appKey string, req *restful.Request) error {
	datatype := req.PathParameter(api.ParamDataType)
	version := req.PathParameter(api.ParamVersion)
	group := req.PathParameter(api.ParamGroup)
	subPath := req.PathParameter(serviceproxy.ParamSubPath)

	signature := req.HeaderParameter("X-Auth-Signature")
	err := ValidateAppKey(appKey, subPath, datatype, version, group, signature)
	if err != nil {
		klog.Infof("ValidateAppKeyWithRequest err=%v", err)
	}
	return err
}

func ValidateAppKey(appKey, subPath, dataType, version, group, signature string) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	sysClient, err := sysclientset.NewForConfig(config)
	if err != nil {
		return err
	}
	aps, err := sysClient.SysV1alpha1().ApplicationPermissions(constants.MyNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	var appPerm *sysv1alpha1.ApplicationPermission
	for _, ap := range aps.Items {
		if ap.Spec.Key == appKey {
			appPerm = &ap
			break
		}
	}
	if appPerm == nil {
		return errors.New("cannot find application permission by appKey")
	}

	appSecret := appPerm.Spec.Secret
	now := time.Now()
	mLevelTime := now.Truncate(time.Minute)
	timestamp := strconv.Itoa(int(mLevelTime.Unix()))

	sha := sha256.New()
	sha.Write([]byte(appKey))
	sha.Write([]byte(appSecret))
	sha.Write([]byte(timestamp))
	hash := hex.EncodeToString(sha.Sum(nil))
	klog.Infof("hash = %s", hash)
	if hash != signature {
		return errors.New("invalid signature ")
	}

	if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	accReq := sysv1alpha1.PermissionRequire{
		Group:    group,
		DataType: dataType,
		Version:  version,
		Ops:      []string{subPath},
	}
	klog.Infof("accReq: %#v", accReq)
	providerRegistries, err := sysClient.SysV1alpha1().
		ProviderRegistries(constants.MyNamespace).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	uris := make([]string, 0)
	var providerReg *sysv1alpha1.ProviderRegistry
	if len(providerRegistries.Items) > 0 {
		for _, pr := range providerRegistries.Items {
			if pr.Status.State == sysv1alpha1.Active {
				if pr.Spec.DataType == dataType &&
					pr.Spec.Group == group &&
					pr.Spec.Version == version &&
					pr.Spec.Kind == sysv1alpha1.Provider {

					providerReg = &pr
					break
				}
			}
		}
	}
	if providerReg == nil {
		return prodiverregistry.ErrProviderNotFound
	}
	requiredOps := sets.String{}
	for _, opReq := range appPerm.Spec.Permission {
		if providerReg.Spec.DataType == opReq.DataType && providerReg.Spec.Group == opReq.Group &&
			providerReg.Spec.Version == opReq.Version {
			for _, op := range opReq.Ops {
				requiredOps.Insert(op)
			}
		}
	}

	for _, op := range providerReg.Spec.OpApis {
		if requiredOps.Has(op.Name) {
			uris = append(uris, op.URI)
		}
	}
	klog.Infof("uris: %v", uris)
	for _, reqURI := range accReq.Ops {
		for _, uri := range uris {
			if strings.HasPrefix(reqURI, uri) {
				return nil
			}
		}
	}

	return errors.New("permission denied")
}
