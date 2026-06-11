package terminus

import (
	"context"
	"testing"

	agwconfig "github.com/beclab/Olares/framework/app-gateway/pkg/config"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAppGatewayStackEnabledDefaultOn(t *testing.T) {
	t.Setenv("APP_GATEWAY_STACK_ENABLED", "")
	if !appGatewayStackEnabled() {
		t.Fatal("expected app-gateway stack enabled by default")
	}
}

func TestInstallAppGatewaySystem_SkipsWhenDisabled(t *testing.T) {
	t.Setenv("APP_GATEWAY_STACK_ENABLED", "0")

	oldValidate := validateAppGatewaySystemInstallerArtifactsFunc
	oldUpgrade := upgradeChartsSkipCRDsWaitFunc
	validateAppGatewaySystemInstallerArtifactsFunc = func(string) error {
		t.Fatal("validate should not be called when stack disabled")
		return nil
	}
	upgradeChartsSkipCRDsWaitFunc = func(context.Context, *action.Configuration, *cli.EnvSettings, string, string, string, string, map[string]interface{}, bool) error {
		t.Fatal("upgrade should not be called when stack disabled")
		return nil
	}
	defer func() {
		validateAppGatewaySystemInstallerArtifactsFunc = oldValidate
		upgradeChartsSkipCRDsWaitFunc = oldUpgrade
	}()

	if err := (&InstallAppGatewaySystem{}).Execute(nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestPrepareLinkerdPKI_PreservesExistingSecret(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      linkerdPKISecretName,
			Namespace: agwconfig.LinkerdNamespace(),
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			linkerdPKICACrt:       []byte("ca"),
			linkerdPKICAKey:       []byte("cakey"),
			linkerdPKIIssuerCrt:   []byte("issuer"),
			linkerdPKIIssuerKey:   []byte("issuerkey"),
			linkerdPKIMetadataKey: []byte("{}"),
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: agwconfig.LinkerdNamespace()}}, secret).
		Build()

	installerDir := t.TempDir()
	if err := prepareLinkerdPKIWithClient(context.Background(), c, installerDir); err != nil {
		t.Fatalf("prepare pki failed: %v", err)
	}

	var got corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: secret.Namespace, Name: secret.Name}, &got); err != nil {
		t.Fatalf("get secret failed: %v", err)
	}
	if string(got.Data[linkerdPKICACrt]) != "ca" {
		t.Fatalf("expected existing secret data preserved, got %q", string(got.Data[linkerdPKICACrt]))
	}
}

func TestPrepareLinkerdPKI_CreatesSecretAndOwnership(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	installerDir := t.TempDir()
	oldLoadOrCreate := loadOrCreateLinkerdPKIFunc
	var gotPKIVendorPath string
	loadOrCreateLinkerdPKIFunc = func(ctx context.Context, cl client.Client, ns, certsDir string) (*linkerdPKIMaterial, error) {
		gotPKIVendorPath = certsDir
		mat := &linkerdPKIMaterial{
			CACrt:     []byte("ca"),
			CAKey:     []byte("cakey"),
			IssuerCrt: []byte("issuer"),
			IssuerKey: []byte("issuerkey"),
		}
		err := cl.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      linkerdPKISecretName,
				Namespace: ns,
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{
				linkerdPKICACrt:       mat.CACrt,
				linkerdPKICAKey:       mat.CAKey,
				linkerdPKIIssuerCrt:   mat.IssuerCrt,
				linkerdPKIIssuerKey:   mat.IssuerKey,
				linkerdPKIMetadataKey: []byte("{}"),
			},
		})
		return mat, err
	}
	defer func() { loadOrCreateLinkerdPKIFunc = oldLoadOrCreate }()

	if err := prepareLinkerdPKIWithClient(context.Background(), c, installerDir); err != nil {
		t.Fatalf("prepare pki failed: %v", err)
	}

	var gotSecret corev1.Secret
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: agwconfig.LinkerdNamespace(), Name: linkerdPKISecretName}, &gotSecret); err != nil {
		t.Fatalf("get secret failed: %v", err)
	}
	if len(gotSecret.Data[linkerdPKIIssuerCrt]) == 0 || len(gotSecret.Data[linkerdPKIIssuerKey]) == 0 {
		t.Fatal("expected issuer cert/key to be created")
	}

	var gotNS corev1.Namespace
	if err := c.Get(context.Background(), types.NamespacedName{Name: agwconfig.LinkerdNamespace()}, &gotNS); err != nil {
		t.Fatalf("get namespace failed: %v", err)
	}
	if gotNS.Annotations["meta.helm.sh/release-name"] != appGatewaySystemReleaseName {
		t.Fatalf("unexpected release-name annotation: %q", gotNS.Annotations["meta.helm.sh/release-name"])
	}
	if gotNS.Annotations["meta.helm.sh/release-namespace"] != resolveAppGatewayNamespace() {
		t.Fatalf("unexpected release-namespace annotation: %q", gotNS.Annotations["meta.helm.sh/release-namespace"])
	}
	wantVendorPath := appGatewayVendorPath(installerDir)
	if gotPKIVendorPath != wantVendorPath {
		t.Fatalf("expected vendor path %q, got %q", wantVendorPath, gotPKIVendorPath)
	}
}

func TestInstallAppGatewaySystem_UsesExpectedHelmParameters(t *testing.T) {
	t.Setenv("APP_GATEWAY_STACK_ENABLED", "1")
	t.Setenv("OLARES_INSTALLER_DIR", t.TempDir())

	oldValidate := validateAppGatewaySystemInstallerArtifactsFunc
	oldGetConfig := getConfigForSystemInstall
	oldNewClient := newClientForSystemInstall
	oldInit := initConfigForSystemInstall
	oldUpgrade := upgradeChartsSkipCRDsWaitFunc
	oldLoadDefaults := loadAppGatewayDefaultsFunc
	defer func() {
		validateAppGatewaySystemInstallerArtifactsFunc = oldValidate
		getConfigForSystemInstall = oldGetConfig
		newClientForSystemInstall = oldNewClient
		initConfigForSystemInstall = oldInit
		upgradeChartsSkipCRDsWaitFunc = oldUpgrade
		loadAppGatewayDefaultsFunc = oldLoadDefaults
	}()

	validateAppGatewaySystemInstallerArtifactsFunc = func(string) error { return nil }
	getConfigForSystemInstall = func() (*rest.Config, error) { return &rest.Config{}, nil }
	newClientForSystemInstall = func(*rest.Config) (client.Client, error) {
		scheme := runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		return fake.NewClientBuilder().WithScheme(scheme).WithObjects(
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      linkerdPKISecretName,
					Namespace: agwconfig.LinkerdNamespace(),
				},
				Data: map[string][]byte{
					linkerdPKICACrt:     []byte("ca-pem"),
					linkerdPKICAKey:     []byte("ca-key"),
					linkerdPKIIssuerCrt: []byte("issuer-pem"),
					linkerdPKIIssuerKey: []byte("issuer-key"),
				},
			},
		).Build(), nil
	}
	initConfigForSystemInstall = func(*rest.Config, string) (*action.Configuration, *cli.EnvSettings, error) {
		return &action.Configuration{}, &cli.EnvSettings{}, nil
	}
	loadAppGatewayDefaultsFunc = func() (agwconfig.Defaults, error) { return agwconfig.Defaults{}, nil }

	var called bool
	var gotRelease, gotNamespace string
	var gotValues map[string]interface{}
	upgradeChartsSkipCRDsWaitFunc = func(_ context.Context, _ *action.Configuration, _ *cli.EnvSettings,
		appName, _ string, _ string, namespace string, vals map[string]interface{}, _ bool) error {
		called = true
		gotRelease = appName
		gotNamespace = namespace
		gotValues = vals
		return nil
	}

	if err := (&InstallAppGatewaySystem{}).Execute(nil); err != nil {
		t.Fatalf("install app-gateway-system failed: %v", err)
	}
	if !called {
		t.Fatal("expected UpgradeChartsSkipCRDsWait to be called")
	}
	if gotRelease != appGatewaySystemReleaseName {
		t.Fatalf("got release %q, want %q", gotRelease, appGatewaySystemReleaseName)
	}
	if gotNamespace != resolveAppGatewayNamespace() {
		t.Fatalf("got namespace %q, want %q", gotNamespace, resolveAppGatewayNamespace())
	}
	linkerdVals, ok := gotValues["linkerd"].(map[string]interface{})
	if !ok || linkerdVals == nil {
		t.Fatal("expected linkerd subchart values to be present")
	}
	if got := linkerdVals["identityTrustAnchorsPEM"]; got != "ca-pem" {
		t.Fatalf("expected linkerd.identityTrustAnchorsPEM to be injected from secret, got %#v", got)
	}
	identity, _ := linkerdVals["identity"].(map[string]interface{})
	issuer, _ := identity["issuer"].(map[string]interface{})
	tls, _ := issuer["tls"].(map[string]interface{})
	if got := tls["crtPEM"]; got != "issuer-pem" {
		t.Fatalf("expected linkerd identity issuer cert to be injected from secret, got %#v", got)
	}
	if got := tls["keyPEM"]; got != "issuer-key" {
		t.Fatalf("expected linkerd identity issuer key to be injected from secret, got %#v", got)
	}
}
