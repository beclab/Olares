package app

import (
	"context"
	"strings"
	"testing"

	sysv1alpha1 "github.com/beclab/api/api/sys.bytetrade.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
)

const (
	testOwner     = "alice"
	testAppName   = "mailer"
	testNamespace = "mailer-alice"
)

// TestSecretsBlockParses pins the manifest schema: a top-level secrets[] block
// must survive the same YAML decoding the oac loader performs. Without the
// struct tags on manifest.AppConfiguration the block is silently dropped rather
// than rejected, so this guards against a regression that would fail closed and
// leave apps with no secret at all.
func TestSecretsBlockParses(t *testing.T) {
	const manifestYAML = `
olaresManifest.version: 0.12.0
apiVersion: v3
metadata:
  name: mailer
  title: Mailer
  version: 1.0.0
spec:
  namespace: mailer
secrets:
  - name: smtp-password
    valueFrom:
      envName: OLARES_USER_SMTP_PASSWORD
  - name: api-token
    valueFrom:
      envName: OLARES_API_TOKEN
`

	var cfg appcfg.AppConfiguration
	if err := yaml.Unmarshal([]byte(manifestYAML), &cfg); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	if len(cfg.Secrets) != 2 {
		t.Fatalf("expected 2 secrets, got %d (%+v)", len(cfg.Secrets), cfg.Secrets)
	}
	if cfg.Secrets[0].Name != "smtp-password" {
		t.Errorf("secret[0].name = %q, want smtp-password", cfg.Secrets[0].Name)
	}
	if cfg.Secrets[0].ValueFrom == nil || cfg.Secrets[0].ValueFrom.EnvName != "OLARES_USER_SMTP_PASSWORD" {
		t.Errorf("secret[0].valueFrom = %+v, want envName OLARES_USER_SMTP_PASSWORD", cfg.Secrets[0].ValueFrom)
	}
	if cfg.Secrets[1].Name != "api-token" {
		t.Errorf("secret[1].name = %q, want api-token", cfg.Secrets[1].Name)
	}
}

// newSecretsTestScheme builds the scheme inline rather than using pkg/testutil,
// which would introduce an import cycle back into this package via appinstaller.
func newSecretsTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(sysv1alpha1.AddToScheme(s))
	return s
}

func newSecretsTestClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(newSecretsTestScheme()).
		WithObjects(objs...).
		Build()
}

func systemEnv(name, value string) *sysv1alpha1.SystemEnv {
	return &sysv1alpha1.SystemEnv{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: name, Value: value},
	}
}

func userEnv(owner, name, value string) *sysv1alpha1.UserEnv {
	return &sysv1alpha1.UserEnv{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: utils.UserspaceName(owner)},
		EnvVarSpec: sysv1alpha1.EnvVarSpec{EnvName: name, Value: value},
	}
}

func secretsAppConfig(secrets ...appcfg.AppSecretVar) *appcfg.ApplicationConfig {
	return &appcfg.ApplicationConfig{
		AppName:   testAppName,
		OwnerName: testOwner,
		Namespace: testNamespace,
		Secrets:   secrets,
	}
}

func appSecret(name, envName string) appcfg.AppSecretVar {
	return appcfg.AppSecretVar{
		Name:      name,
		ValueFrom: &sysv1alpha1.ValueFrom{EnvName: envName},
	}
}

// TestApplySecretsCreatesOnePerEntry covers the core contract: each entry
// becomes its own Secret, named verbatim, holding the resolved value under the
// constant key a chart's secretKeyRef hardcodes.
func TestApplySecretsCreatesOnePerEntry(t *testing.T) {
	c := newSecretsTestClient(
		systemEnv("OLARES_API_TOKEN", "tok-123"),
		userEnv(testOwner, "OLARES_USER_SMTP_PASSWORD", "hunter2"),
	)

	cfg := secretsAppConfig(
		appSecret("smtp-password", "OLARES_USER_SMTP_PASSWORD"),
		appSecret("api-token", "OLARES_API_TOKEN"),
	)

	if err := ApplySecrets(context.Background(), c, cfg); err != nil {
		t.Fatalf("ApplySecrets: %v", err)
	}

	for name, want := range map[string]string{
		"smtp-password": "hunter2",
		"api-token":     "tok-123",
	} {
		got := new(corev1.Secret)
		if err := c.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: name}, got); err != nil {
			t.Fatalf("get secret %s: %v", name, err)
		}
		if v := string(got.Data[appcfg.AppSecretValueKey]); v != want {
			t.Errorf("secret %s[%s] = %q, want %q", name, appcfg.AppSecretValueKey, v, want)
		}
		if got.Type != corev1.SecretTypeOpaque {
			t.Errorf("secret %s type = %q, want Opaque", name, got.Type)
		}
	}
}

// TestApplySecretsUpdatesExistingValue ensures a changed SystemEnv propagates on
// re-apply (upgrade), rather than the first value sticking forever.
func TestApplySecretsUpdatesExistingValue(t *testing.T) {
	c := newSecretsTestClient(systemEnv("OLARES_API_TOKEN", "old"))
	cfg := secretsAppConfig(appSecret("api-token", "OLARES_API_TOKEN"))

	ctx := context.Background()
	if err := ApplySecrets(ctx, c, cfg); err != nil {
		t.Fatalf("first ApplySecrets: %v", err)
	}

	env := new(sysv1alpha1.SystemEnv)
	if err := c.Get(ctx, client.ObjectKey{Name: "OLARES_API_TOKEN"}, env); err != nil {
		t.Fatalf("get systemenv: %v", err)
	}
	env.Value = "new"
	if err := c.Update(ctx, env); err != nil {
		t.Fatalf("update systemenv: %v", err)
	}

	if err := ApplySecrets(ctx, c, cfg); err != nil {
		t.Fatalf("second ApplySecrets: %v", err)
	}

	got := new(corev1.Secret)
	if err := c.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "api-token"}, got); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if v := string(got.Data[appcfg.AppSecretValueKey]); v != "new" {
		t.Errorf("value = %q, want %q (re-apply must refresh)", v, "new")
	}
}

// TestApplySecretsUserEnvOverridesSystemEnv documents the precedence inherited
// from the envs[] resolution: the owner's UserEnv wins over a same-named
// cluster SystemEnv.
func TestApplySecretsUserEnvOverridesSystemEnv(t *testing.T) {
	c := newSecretsTestClient(
		systemEnv("OLARES_SHARED", "system-value"),
		userEnv(testOwner, "OLARES_SHARED", "user-value"),
	)
	cfg := secretsAppConfig(appSecret("shared", "OLARES_SHARED"))

	if err := ApplySecrets(context.Background(), c, cfg); err != nil {
		t.Fatalf("ApplySecrets: %v", err)
	}

	got := new(corev1.Secret)
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: testNamespace, Name: "shared"}, got); err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if v := string(got.Data[appcfg.AppSecretValueKey]); v != "user-value" {
		t.Errorf("value = %q, want user-value", v)
	}
}

func TestApplySecretsErrors(t *testing.T) {
	cases := []struct {
		name    string
		secrets []appcfg.AppSecretVar
	}{
		{
			name:    "unknown env reference",
			secrets: []appcfg.AppSecretVar{appSecret("api-token", "NOPE")},
		},
		{
			name:    "missing valueFrom",
			secrets: []appcfg.AppSecretVar{{Name: "api-token"}},
		},
		{
			name:    "empty name",
			secrets: []appcfg.AppSecretVar{appSecret("", "OLARES_API_TOKEN")},
		},
		{
			// Two entries targeting one Secret would clobber each other; the
			// single-value-per-Secret model has no way to merge them.
			name: "duplicate secret name",
			secrets: []appcfg.AppSecretVar{
				appSecret("api-token", "OLARES_API_TOKEN"),
				appSecret("api-token", "OLARES_API_TOKEN"),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newSecretsTestClient(systemEnv("OLARES_API_TOKEN", "tok"))
			if err := ApplySecrets(context.Background(), c, secretsAppConfig(tc.secrets...)); err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}

// TestApplySecretsForReportsOnlyRealChanges pins the change-detection contract
// the AppEnv controller relies on to decide whether a rotation warrants a
// redeploy. Re-applying identical data must report nothing, otherwise every
// no-op reconcile would trigger a spurious helm upgrade.
func TestApplySecretsForReportsOnlyRealChanges(t *testing.T) {
	ctx := context.Background()
	c := newSecretsTestClient(
		systemEnv("OLARES_API_TOKEN", "tok"),
		systemEnv("OLARES_OTHER", "other"),
	)
	secrets := []appcfg.AppSecretVar{
		appSecret("api-token", "OLARES_API_TOKEN"),
		appSecret("other", "OLARES_OTHER"),
	}

	changed, err := ApplySecretsFor(ctx, c, testNamespace, testAppName, testOwner, secrets)
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if len(changed) != 2 {
		t.Fatalf("first apply changed = %v, want both secrets created", changed)
	}

	changed, err = ApplySecretsFor(ctx, c, testNamespace, testAppName, testOwner, secrets)
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("re-applying identical data changed = %v, want none", changed)
	}

	env := new(sysv1alpha1.SystemEnv)
	if err := c.Get(ctx, client.ObjectKey{Name: "OLARES_API_TOKEN"}, env); err != nil {
		t.Fatalf("get systemenv: %v", err)
	}
	env.Value = "rotated"
	if err := c.Update(ctx, env); err != nil {
		t.Fatalf("update systemenv: %v", err)
	}

	changed, err = ApplySecretsFor(ctx, c, testNamespace, testAppName, testOwner, secrets)
	if err != nil {
		t.Fatalf("third apply: %v", err)
	}
	if len(changed) != 1 || changed[0] != "api-token" {
		t.Fatalf("after rotating one env, changed = %v, want [api-token]", changed)
	}
}

// TestApplyAppEnvRecordsSecretDeclarations covers the link that makes live
// propagation possible at all: the declarations must land on the AppEnv CR,
// because that is the only thing the SystemEnv/UserEnv controllers can scan to
// discover which apps reference a rotated variable.
func TestApplyAppEnvRecordsSecretDeclarations(t *testing.T) {
	ctx := context.Background()
	c := newSecretsTestClient(systemEnv("OLARES_API_TOKEN", "tok"))

	cfg := secretsAppConfig(appSecret("api-token", "OLARES_API_TOKEN"))
	appEnv, err := ApplyAppEnv(ctx, c, cfg)
	if err != nil {
		t.Fatalf("ApplyAppEnv: %v", err)
	}
	if appEnv == nil {
		t.Fatal("ApplyAppEnv returned nil for an app declaring only secrets")
	}

	stored := new(sysv1alpha1.AppEnv)
	key := client.ObjectKey{Namespace: testNamespace, Name: FormatAppEnvName(testAppName, testOwner)}
	if err := c.Get(ctx, key, stored); err != nil {
		t.Fatalf("get appenv: %v", err)
	}
	if len(stored.Secrets) != 1 {
		t.Fatalf("appEnv.Secrets = %+v, want one declaration", stored.Secrets)
	}
	if stored.Secrets[0].Name != "api-token" {
		t.Errorf("recorded name = %q, want api-token", stored.Secrets[0].Name)
	}
	if stored.Secrets[0].ValueFrom == nil || stored.Secrets[0].ValueFrom.EnvName != "OLARES_API_TOKEN" {
		t.Errorf("recorded valueFrom = %+v, want OLARES_API_TOKEN", stored.Secrets[0].ValueFrom)
	}
}

// TestApplyAppEnvDoesNotRecordSecretValues is the guard on the core privacy
// property: the AppEnv CR must never carry a resolved secret value. AppSecretVar
// has no value field today; this fails loudly if one is ever added and populated.
func TestApplyAppEnvDoesNotRecordSecretValues(t *testing.T) {
	ctx := context.Background()
	const secretValue = "super-secret-value"
	c := newSecretsTestClient(systemEnv("OLARES_API_TOKEN", secretValue))

	cfg := secretsAppConfig(appSecret("api-token", "OLARES_API_TOKEN"))
	if _, err := ApplyAppEnv(ctx, c, cfg); err != nil {
		t.Fatalf("ApplyAppEnv: %v", err)
	}
	if err := ApplySecrets(ctx, c, cfg); err != nil {
		t.Fatalf("ApplySecrets: %v", err)
	}

	stored := new(sysv1alpha1.AppEnv)
	key := client.ObjectKey{Namespace: testNamespace, Name: FormatAppEnvName(testAppName, testOwner)}
	if err := c.Get(ctx, key, stored); err != nil {
		t.Fatalf("get appenv: %v", err)
	}

	serialized, err := yaml.Marshal(stored)
	if err != nil {
		t.Fatalf("marshal appenv: %v", err)
	}
	if strings.Contains(string(serialized), secretValue) {
		t.Fatalf("resolved secret value leaked into the AppEnv CR:\n%s", serialized)
	}
}

// TestApplyAppEnvSecretsDoNotMutateConfig guards against writing the pending
// status through the shared ValueFrom pointer, which would mutate the caller's
// ApplicationConfig.
func TestApplyAppEnvSecretsDoNotMutateConfig(t *testing.T) {
	ctx := context.Background()
	c := newSecretsTestClient(systemEnv("OLARES_API_TOKEN", "tok"))

	cfg := secretsAppConfig(appSecret("api-token", "OLARES_API_TOKEN"))
	if _, err := ApplyAppEnv(ctx, c, cfg); err != nil {
		t.Fatalf("ApplyAppEnv: %v", err)
	}
	if s := cfg.Secrets[0].ValueFrom.Status; s != "" {
		t.Errorf("caller's config was mutated: ValueFrom.Status = %q, want empty", s)
	}
}

// TestApplySecretsNoOpWithoutDeclaration keeps apps that declare no secrets[]
// completely untouched.
func TestApplySecretsNoOpWithoutDeclaration(t *testing.T) {
	c := newSecretsTestClient()
	if err := ApplySecrets(context.Background(), c, secretsAppConfig()); err != nil {
		t.Fatalf("ApplySecrets: %v", err)
	}

	list := new(corev1.SecretList)
	if err := c.List(context.Background(), list); err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected no secrets created, got %d", len(list.Items))
	}
}
