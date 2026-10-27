package apps

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sort"
	"strings"
	"time"

	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	clientset "bytetrade.io/web3os/bfl/pkg/client/clientset/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/client/dynamic_client"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils/certmanager"
	"bytetrade.io/web3os/bfl/pkg/watchers"
	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

var GVR = schema.GroupVersionResource{
	Group: "app.bytetrade.io", Version: "v1alpha1", Resource: "applications",
}

type Subscriber struct {
	*watchers.Subscriber
	client        clientset.ClientInterface
	username      string
	dynamicClient *dynamic_client.ResourceClient[appv1.Application]
}

func (s *Subscriber) WithKubeConfig(config *rest.Config) *Subscriber {
	s.dynamicClient = dynamic_client.NewResourceClientOrDie[appv1.Application](GVR)
	s.client, _ = clientset.NewKubeClient(config)
	return s
}

func (s *Subscriber) HandleEvent() cache.ResourceEventHandler {
	return cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			app, ok := obj.(*appv1.Application)
			if !ok {
				klog.Error("not application resource, invalid obj")
				return false
			}

			if strings.HasPrefix(app.Namespace, "user-space-") {
				return false
			}

			return true
		},

		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       obj,
					Action:    watchers.ADD,
				}
				s.Watchers.Enqueue(eobj)
			},
			DeleteFunc: func(obj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       obj,
					Action:    watchers.DELETE,
				}
				s.Watchers.Enqueue(eobj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				eobj := watchers.EnqueueObj{
					Subscribe: s,
					Obj:       newObj,
					Action:    watchers.UPDATE,
				}
				s.Watchers.Enqueue(eobj)
			},
		},
	}
}

func (s *Subscriber) Do(ctx context.Context, obj interface{}, action watchers.Action) error {
	app, ok := obj.(*appv1.Application)
	if !ok {
		return errors.New("invalid application object")
	}
	var usernames []string
	if appv1.IsShared(app) {
		for username, settings := range app.Spec.UserSettings {
			if username != "" && settings[constants.ApplicationCustomDomain] != "" {
				usernames = append(usernames, username)
			}
		}
		sort.Strings(usernames)
	} else if app.Spec.Owner != "" {
		usernames = append(usernames, app.Spec.Owner)
	}
	var errs []error
	for _, username := range usernames {
		// The informer and clients are shared; user state belongs to this task.
		scoped := *s
		scoped.username = username
		var err error
		if action == watchers.DELETE {
			err = scoped.removeCustomDomainRetry(ctx, app)
		} else {
			current, getErr := scoped.getObj(app.Name)
			if kerrors.IsNotFound(getErr) {
				continue
			}
			if getErr != nil {
				err = getErr
			} else {
				err = scoped.checkCustomDomainStatus(ctx, current)
			}
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("user %s, app %s: %w", username, app.Name, err))
		}
	}
	return utilerrors.NewAggregate(errs)
}

func (s *Subscriber) removeCustomDomainRetry(ctx context.Context, request *appv1.Application) error {
	backoff := wait.Backoff{
		Duration: 2 * time.Second,
		Factor:   2,
		Jitter:   0.1,
		Steps:    5,
	}

	var err = retry.OnError(backoff, func(err error) bool {
		return ctx.Err() == nil
	}, func() error {
		return s.removeCustomDomainCnameData(ctx, request)
	})

	if err != nil {
		klog.Errorf("remove custom domain error: %v", err)
	}

	return err
}

func (s *Subscriber) checkCustomDomainStatus(ctx context.Context, app *appv1.Application) error {
	var err error
	// For shared apps the customDomain blob lives in
	// Spec.UserSettings[s.username]; for v1/v2 it's still
	// Spec.Settings. EffectiveSettings hides the difference.
	effective := app.EffectiveSettings(s.username)
	customDomains, ok := effective[constants.ApplicationCustomDomain]
	if !ok || customDomains == "" {
		return nil
	}

	var customDomainsObj = make(map[string]map[string]string)
	if err = json.Unmarshal([]byte(customDomains), &customDomainsObj); err != nil {
		klog.Errorf("customdomain-status-check queue unmarshal custom domains error %+v", err)
		return nil
	}

	if len(customDomainsObj) == 0 {
		return nil
	}

	var existsPending bool
	var existsCustomDomain bool
	for _, customDomainObj := range customDomainsObj {
		var domainStatus string
		customDomainName := customDomainObj[constants.ApplicationThirdPartyDomain]
		customDomainCnameTargetStatus := customDomainObj[constants.ApplicationCustomDomainCnameTargetStatus]
		customDomainCnameStatus := customDomainObj[constants.ApplicationCustomDomainCnameStatus]

		if customDomainName == "" || customDomainCnameTargetStatus == constants.CustomDomainCnameStatusNotset ||
			(customDomainCnameTargetStatus == constants.CustomDomainCnameStatusSet && customDomainCnameStatus == constants.CustomDomainCnameStatusActive) {
			continue
		}

		existsCustomDomain = true
		domainStatus, err = s.checkStatus(ctx, customDomainName)
		if err != nil {
			break
		}

		if domainStatus == constants.CustomDomainCnameStatusEmpty ||
			domainStatus == constants.CustomDomainCnameStatusNotset ||
			domainStatus == constants.CustomDomainCnameStatusError {
			err = fmt.Errorf("app custom domain status check invalid: %s", domainStatus)
			break
		}
		customDomainObj[constants.ApplicationCustomDomainCnameStatus] = domainStatus
		if domainStatus == constants.CustomDomainCnameStatusPending ||
			domainStatus == constants.CustomDomainCnameStatusCertNotFound ||
			domainStatus == constants.CustomDomainCnameStatusCertInvalid {
			existsPending = true
		}
	}

	if !existsCustomDomain {
		return nil
	}

	if err != nil {
		return err
	}

	cdos, err := json.Marshal(customDomainsObj)
	if err != nil {
		return err
	}
	// Write back into the same slot we read from: shared → UserSettings[u],
	// v1/v2 → Spec.Settings.
	if appv1.IsShared(app) {
		if app.Spec.UserSettings == nil {
			app.Spec.UserSettings = make(map[string]map[string]string)
		}
		if app.Spec.UserSettings[s.username] == nil {
			app.Spec.UserSettings[s.username] = make(map[string]string)
		}
		app.Spec.UserSettings[s.username][constants.ApplicationCustomDomain] = string(cdos)
	} else {
		app.Spec.Settings[constants.ApplicationCustomDomain] = string(cdos)
	}
	if err = s.updateApp(app); err != nil {
		return err
	}
	if existsPending {
		return errors.New("app custom domain status check is pending")
	}
	return nil
}

func (s *Subscriber) updateApp(app *appv1.Application) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := s.dynamicClient.Update(ctx, app, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *Subscriber) checkStatus(ctx context.Context, domainName string) (string, error) {
	cm, err := certmanager.NewCertManagerForUser(ctx, s.username)
	if err != nil {
		return constants.CustomDomainCnameStatusEmpty, err
	}

	domainCnameStatus, err := cm.GetCustomDomainCnameStatus(domainName)
	if err != nil {
		return constants.CustomDomainCnameStatusError, nil
	}
	if !domainCnameStatus.Success {
		return constants.CustomDomainCnameStatusNotset, nil
	}

	reverseProxyType, err := s.getReverseProxyType()
	if err != nil {
		return "", errors.Wrap(err, "get reverse proxy type error")
	}

	if reverseProxyType == constants.ReverseProxyTypeFRP {
		return s.getStatusByCustomDomainCert(domainName)
	}

	cnameStatus, err := cm.GetCustomDomainOnCloudflare(domainName)
	if err != nil {
		errmsg := cm.GetCustomDomainErrorStatus(err)
		if errmsg != constants.CustomDomainCnameStatusNone {
			return constants.CustomDomainCnameStatusEmpty, err
		}

		_, err = cm.AddCustomDomainOnCloudflare(domainName)
		if err != nil {
			return constants.CustomDomainCnameStatusEmpty, err
		}
	}
	var sslStatus, hostnameStatus string = constants.CustomDomainCnameStatusPending, constants.CustomDomainCnameStatusPending
	if cnameStatus != nil {
		sslStatus = cnameStatus.SSLStatus
		hostnameStatus = cnameStatus.HostnameStatus
	}

	return s.mergeCnameStatus(sslStatus, hostnameStatus), nil
}

func (s *Subscriber) getStatusByCustomDomainCert(domainName string) (string, error) {
	certConfigMapName := fmt.Sprintf(constants.AppEntranceCertConfigMapNameTpl, domainName)
	certConfigMap, err := s.client.Kubernetes().CoreV1().ConfigMaps(fmt.Sprintf(constants.UserspaceNameFormat, s.username)).Get(context.Background(), certConfigMapName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return constants.CustomDomainCnameStatusCertNotFound, nil
		}
		return "", errors.Wrapf(err, "get cert configmap for custom domain %s error", domainName)
	}
	if certConfigMap.Data == nil {
		return constants.CustomDomainCnameStatusCertNotFound, nil
	}
	certData := certConfigMap.Data[constants.AppEntranceCertConfigMapCertKey]
	keyData := certConfigMap.Data[constants.AppEntranceCertConfigMapKeyKey]
	zone := certConfigMap.Data[constants.AppEntranceCertConfigMapZoneKey]
	if certData == "" || keyData == "" || zone == "" {
		return constants.CustomDomainCnameStatusCertNotFound, nil
	}
	sslErr := CheckSSLCertificate([]byte(certData), []byte(keyData), zone)
	if sslErr != nil {
		return constants.CustomDomainCnameStatusCertInvalid, nil
	}
	return constants.CustomDomainCnameStatusActive, nil
}

func CheckSSLCertificate(cert, key []byte, hostname string) error {
	block, _ := pem.Decode(cert)
	if block == nil {
		return errors.New("error decoding certificate PEM block")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.Wrap(err, "certificate is invalid")
	}
	// verify hostname
	err = pub.VerifyHostname(hostname)
	if err != nil {
		return err
	}

	// verify certificate whether valid or expired
	currentTime := time.Now()
	if currentTime.Before(pub.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if currentTime.After(pub.NotAfter) {
		return errors.New("certificate has expired")
	}

	block, _ = pem.Decode(key)
	if block == nil {
		return errors.New("error decoding private key PEM block")
	}

	hash := sha256.Sum256([]byte("hello"))

	switch block.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing pkcs#8 private key: %v", err)
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, key.(*rsa.PrivateKey), crypto.SHA256, hash[:])
		if err != nil {
			return errors.New("failed to sign message")
		}
		rsaPub, ok := pub.PublicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("not RSA public key")
		}
		err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature)
		if err != nil {
			return errors.New("certificate and private key not match")
		}
	case "EC PRIVATE KEY":
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing ecdsa private key: %v", err)
		}

		r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])
		if err != nil {
			return fmt.Errorf("ecdsa sign err: %v", err)
		}
		verified := ecdsa.Verify(pub.PublicKey.(*ecdsa.PublicKey), hash[:], r, s)
		if !verified {
			return errors.New("certificate and private key not match")
		}
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("error parsing rsa private key: %v", err)
		}
		err = key.Validate()
		if err != nil {
			return fmt.Errorf("rsa private key failed validation: %v", err)
		}
		signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
		if err != nil {
			return errors.New("failed to sign message")
		}
		rsaPub, ok := pub.PublicKey.(*rsa.PublicKey)
		if !ok {
			return errors.New("not RSA public key")
		}
		err = rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, hash[:], signature)
		if err != nil {
			return errors.New("certificate and private key not match")
		}
	default:
		return fmt.Errorf("unknown private key type: %s", block.Type)
	}

	return nil
}

func (s *Subscriber) removeCustomDomainCnameData(ctx context.Context, app *appv1.Application) error {
	var err error

	// Reflect per-user storage when reading the customDomain blob so
	// we delete the right user's cloudflare CNAMEs on app delete.
	customDomainData := app.EffectiveSettings(s.username)[constants.ApplicationCustomDomain]
	if customDomainData == "" {
		return nil
	}

	var entrancesCustomDomain = make(map[string]map[string]interface{})
	if err = json.Unmarshal([]byte(customDomainData), &entrancesCustomDomain); err != nil {
		return nil
	}

	if len(entrancesCustomDomain) == 0 {
		return nil
	}

	cm, err := certmanager.NewCertManagerForUser(ctx, s.username)
	if err != nil {
		return err
	}

	for _, entranceCustomDomain := range entrancesCustomDomain {
		customDomainName, ok := entranceCustomDomain[constants.ApplicationThirdPartyDomain]
		if !ok || customDomainName == nil {
			continue
		}

		var domainName = customDomainName.(string)
		if domainName == "" {
			continue
		}

		_, err = cm.DeleteCustomDomainOnCloudflare(domainName)
		if err != nil {
			break
		}
	}
	return err
}

func (s *Subscriber) mergeCnameStatus(sslStatus, hostnameStatus string) string {
	switch {
	case hostnameStatus == sslStatus && sslStatus == constants.CustomDomainCnameStatusActive:
		return constants.CustomDomainCnameStatusActive
	default:
		return constants.CustomDomainCnameStatusPending
	}
}

func (s *Subscriber) getReverseProxyType() (string, error) {
	op, err := operator.NewUserOperator()
	if err != nil {
		return "", err
	}
	return op.GetReverseProxyType()
}

func (s *Subscriber) getObj(appName string) (*appv1.Application, error) {
	return s.dynamicClient.Get(context.Background(), appName, metav1.GetOptions{})
}
