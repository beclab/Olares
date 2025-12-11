package controllers

import (
	"bytes"
	"bytetrade.io/web3os/bfl/internal/frpc/command"
	v1alpha1App "bytetrade.io/web3os/bfl/internal/ingress/api/app.bytetrade.io/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/apis/iam/v1alpha1/operator"
	"bytetrade.io/web3os/bfl/pkg/apis/settings/v1alpha1"
	"bytetrade.io/web3os/bfl/pkg/constants"
	"bytetrade.io/web3os/bfl/pkg/utils"
	"bytetrade.io/web3os/bfl/pkg/utils/k8sutil"
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"os"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"text/template"
	"time"
)

type FRPCConfig = struct {
	Server         string
	Port           int
	AuthMethod     string
	UserName       string
	TerminusName   string
	JWSToken       string
	AuthToken      string
	L4ProxyIP      string
	L4ProxySSLPort string
	UserZone       string
	CustomDomains  []string
}

var QueueSize = 50

var retryInterval = time.Second * 5

var frpConfigTmpl = `
serverAddr = "{{ .Server }}"
serverPort = {{ .Port }}
webServer.port = 7400
user = "{{ .TerminusName }}"
{{- if eq .AuthMethod "jws" }}
auth.method = "jws"
auth.jws = "{{ .JWSToken }}"
{{- end }}
{{- if eq .AuthMethod "token" }}
auth.method = "token"
auth.token = "{{ .AuthToken }}"
{{ end }}
loginFailExit = false

[[proxies]]
name = "web"
type = "http"
localPort = 80
localIp = "{{ .L4ProxyIP }}"
customDomains = ["{{ .UserZone }}"]

[[proxies]]
name = "web_wildcard"
type = "http"
localPort = 80
localIp = "{{ .L4ProxyIP }}"
customDomains = ["*.{{ .UserZone }}"]

[[proxies]]
name = "web_ssl"
type = "https"
localPort = {{ .L4ProxySSLPort }}
localIp = "{{ .L4ProxyIP }}"
customDomains = ["{{ .UserZone }}"]
transport.proxyProtocolVersion = "v1"

[[proxies]]
name = "web_ssl_wildcard"
type = "https"
localPort = {{ .L4ProxySSLPort }}
localIp = "{{ .L4ProxyIP }}"
customDomains = ["*.{{ .UserZone }}"]
transport.proxyProtocolVersion = "v1"

{{- if gt (len .CustomDomains) 0 }}
{{- $port := .L4ProxySSLPort }}
{{- $ip := .L4ProxyIP -}}
{{- range $server := .CustomDomains }}
{{ if ne $server "" }}
[[proxies]]
name = "web_ssl_{{ $server }}"
type = "https"
localPort = {{ $port }}
localIp = "{{ $ip }}"
customDomains = ["{{ $server }}"]
transport.proxyProtocolVersion = "v1"
{{- end }}
{{- end }}
{{- end }}
`

type FrpcController struct {
	client.Client
	Config         *FRPCConfig
	Log            logr.Logger
	command        *command.FrpcCommand
	ReconcileQueue chan string
}

func (f *FrpcController) RunFrpc(ctx context.Context) error {
	f.command = command.NewFrpcCommand()
	if err := f.InitializeConfig(); err != nil {
		return errors.Wrap(err, "init frpc config failed")
	}
	cmd := f.command.StartCmd()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start frpc failed")
	}

	go f.runQueue(ctx)

	return nil
}

func (f *FrpcController) InitializeConfig() error {
	if err := f.populateConfigWithUserData(); err != nil {
		return errors.Wrap(err, "populate user data failed")
	}
	if err := f.populateConfigWithL4ProxyData(); err != nil {
		return errors.Wrap(err, "populateConfigWithL4ProxyData failed")
	}
	// the custom domains of the apps will be populated to the config
	// as soon as the manager starts
	if err := f.writeConfig(); err != nil {
		return errors.Wrap(err, "writeConfig failed")
	}
	f.Log.Info("initialized frp config", "config", f.Config)
	return nil
}

func (f *FrpcController) populateConfigWithUserData() error {
	userOp, err := operator.NewUserOperator()
	if err != nil {
		return errors.Wrap(err, "failed to create user operator")
	}

	user, err := userOp.GetUser(f.Config.UserName)
	if err != nil {
		return errors.Wrap(err, "failed to get user")
	}
	terminusName := constants.TerminusName(userOp.GetTerminusName(user))
	f.Config.TerminusName = string(terminusName)
	f.Config.UserZone = terminusName.UserZone()
	if f.Config.AuthMethod == v1alpha1.FRPAuthMethodJWS {
		jwsToken := userOp.GetUserAnnotation(user, constants.UserCertManagerJWSToken)
		if jwsToken == "" {
			return errors.Wrap(err, "auth method is selected as jws but no jws token is provided")
		}
		f.Config.JWSToken = jwsToken
	}
	return nil
}

func (f *FrpcController) populateConfigWithL4ProxyData() error {
	ip, err := k8sutil.GetL4ProxyNodeIP(context.Background(), 10*time.Second)
	if err != nil {
		return errors.Wrap(err, "failed to get l4 proxy ip")
	}
	sslPort := utils.EnvOrDefault("L4_PROXY_LISTEN_PROXY_PROTOCOL", constants.L4ListenSSLProxyProtocolPort)
	f.Config.L4ProxyIP = *ip
	f.Config.L4ProxySSLPort = sslPort
	return nil
}

func (f *FrpcController) populateConfigWithCustomDomains() error {
	apps, err := f.getApps(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to get apps")
	}

	if apps == nil || len(apps.Items) == 0 {
		f.Log.Info("no apps found, will not update custom domains")
		return nil
	}

	domains, err := f.getCustomDomains(apps)
	if err != nil {
		return errors.Wrap(err, "failed to get custom domains from apps")
	}

	f.Config.CustomDomains = domains
	return nil
}

// writeConfig renders the frp config template with the current FrpcController.Config
// a temporary file with the new configuration is created for verification
// the actual config file is written to only after the verification succeeds
func (f *FrpcController) writeConfig() error {
	t, err := template.New("frpconfig").Parse(frpConfigTmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse frp config template")
	}

	var bf bytes.Buffer
	if err = t.Execute(&bf, f.Config); err != nil {
		return errors.Wrap(err, "failed to render frp config template")
	}

	content := bf.Bytes()
	tmpConfFile, err := os.CreateTemp("", "frp-tmp-conf-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp frp config file to verify")
	}
	defer func() {
		if err := os.Remove(tmpConfFile.Name()); err != nil {
			f.Log.Error(err, "failed to remove temp frp config file", "file", tmpConfFile.Name())
		}
	}()
	if _, err := tmpConfFile.Write(content); err != nil {
		return errors.Wrap(err, "failed to write temp frp config file to verify")
	}
	if verifyOutput, err := f.command.VerifyConfig(tmpConfFile.Name()); err != nil {
		f.Log.Error(err,
			"failed to verify config, refuse to write config",
			"output", string(verifyOutput),
			"content", string(content),
		)
		return errors.Wrap(err, "failed to verify config")
	}

	return errors.Wrap(f.command.WriteConfig(content), "failed to write frp config")
}

func (f *FrpcController) runQueue(ctx context.Context) {
	// do reload when there's change of the apps
	// or a previous reload attempt failed
	var retry bool
	for {
		select {
		case app := <-f.ReconcileQueue:
			// if there are multiple buffered app resources
			// merge them to trigger only a single update action
			apps := []string{app}
			for len(f.ReconcileQueue) > 0 {
				apps = append(apps, <-f.ReconcileQueue)
			}
			if err := f.reloadFRPCWithLatestDomains(); err != nil {
				f.Log.Error(err, "failed to reload latest domains")
				retry = true
			}
			retry = false
		case <-time.After(retryInterval):
			if !retry {
				continue
			}
			if err := f.reloadFRPCWithLatestDomains(); err != nil {
				f.Log.Error(err, "failed to reload latest domains")
			} else {
				retry = false
			}
		case <-ctx.Done():
			return
		}
	}
}

// reloadFRPCWithLatestDomains updates the frpc config file
// with the latest custom domains derived from the changed app resources
// and reloads the frpc process,
// other entries within the config like ip, port or authentication are not updated
func (f *FrpcController) reloadFRPCWithLatestDomains() error {
	if err := f.populateConfigWithCustomDomains(); err != nil {
		return errors.Wrap(err, "populate custom domains failed")
	}
	if err := f.writeConfig(); err != nil {
		return errors.Wrap(err, "writeConfig failed")
	}

	if reloadOutput, err := f.command.Reload(); err != nil {
		f.Log.Error(err, "failed to reload frp config", "output", string(reloadOutput))
		return errors.Wrap(err, "failed to reload frp config")
	}

	f.Log.Info("updated frp config", "config", f.Config)
	return nil
}

// +kubebuilder:rbac:groups=app.bytetrade.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=app.bytetrade.io,resources=applications/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=app.bytetrade.io,resources=applications/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (f *FrpcController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	f.ReconcileQueue <- req.Name

	return ctrl.Result{}, nil
}

func (f *FrpcController) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).For(&v1alpha1App.Application{}, builder.WithPredicates(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		CreateFunc: func(ce event.CreateEvent) bool {
			return isOwnerApp(ce.Object)
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			return isOwnerApp(de.Object)
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			if !isOwnerApp(ue.ObjectOld, ue.ObjectNew) {
				return false
			}
			old, ok1 := ue.ObjectOld.(*v1alpha1App.Application)
			_new, ok2 := ue.ObjectNew.(*v1alpha1App.Application)
			if !(ok1 && ok2) || reflect.DeepEqual(old.Spec, _new.Spec) {
				return false
			}
			return true
		},
	})).Build(f)

	if err != nil {
		return err
	}

	return nil
}

func (f *FrpcController) getApps(parentCtx context.Context) (*v1alpha1App.ApplicationList, error) {
	ctx, cancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer cancel()

	var err error
	var appList v1alpha1App.ApplicationList

	if err = f.List(ctx, &appList, client.InNamespace("")); err != nil {
		return nil, err
	}

	return &appList, nil
}

func (f *FrpcController) getCustomDomains(apps *v1alpha1App.ApplicationList) (customDomains []string, err error) {
	for _, app := range apps.Items {
		var settings = app.Spec.Settings
		if settings == nil || len(settings) == 0 {
			continue
		}

		var customDomainSettings = settings[constants.ApplicationCustomDomain]
		if customDomainSettings == "" {
			continue
		}

		var r = map[string]map[string]string{}
		if err = json.Unmarshal([]byte(customDomainSettings), &r); err != nil {
			continue
		}

		for _, v := range r {
			if domain := v[constants.ApplicationThirdPartyDomain]; domain != "" {
				customDomains = append(customDomains, domain)
			}
		}
	}
	return
}

func isOwnerApp(objs ...client.Object) bool {
	var isTrue = len(objs) != 0
	for _, obj := range objs {
		app, ok := obj.(*v1alpha1App.Application)
		isTrue = ok && app.Spec.Owner == constants.Username && isTrue
	}
	return isTrue
}
