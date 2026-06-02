package watchers

import (
	"context"
	"sort"
	"strings"
	"testing"

	"bytetrade.io/web3os/tapr/pkg/app/application"
	"github.com/coredns/corefile-migration/migration/corefile"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestBuildSharedInclusterTemplates_empty(t *testing.T) {
	if got := buildSharedInclusterTemplates(nil, "10.0.0.5"); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
	if got := buildSharedInclusterTemplates([]SharedInclusterEntrance{}, "10.0.0.5"); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
	if got := buildSharedInclusterTemplates([]SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "ollama", EntranceID: "a5be2268", Viewer: "alice", PlatformDomain: "olares.com"},
	}, "not-an-ip"); got != nil {
		t.Fatalf("expected nil for invalid gateway IP, got %#v", got)
	}
}

func TestBuildSharedInclusterTemplates_orderingAndDedup(t *testing.T) {
	entrances := []SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "shared", EntranceID: "a5be2268", Viewer: "bob", PlatformDomain: "olares.com"},
		{AppID: "a5be2268", EntranceName: "shared", EntranceID: "a5be2268", Viewer: "alice", PlatformDomain: "olares.com"},
		{AppID: "a5be2268", EntranceName: "shared", EntranceID: "a5be2268", Viewer: "bob", PlatformDomain: "olares.com"},
	}
	plugins := buildSharedInclusterTemplates(entrances, "10.0.0.8")
	if len(plugins) != 2 {
		t.Fatalf("expected 2 template plugins, got %d", len(plugins))
	}
	wantAlice := "a5be2268.alice.olares.com"
	wantBob := "a5be2268.bob.olares.com"
	if !templateMatchesFQDN(plugins[0], wantAlice, "10.0.0.8") {
		t.Fatalf("first template does not target %s: %s", wantAlice, plugins[0].ToString())
	}
	if !templateMatchesFQDN(plugins[1], wantBob, "10.0.0.8") {
		t.Fatalf("second template does not target %s: %s", wantBob, plugins[1].ToString())
	}
}

func TestBuildSharedInclusterTemplates_perUserEntranceExcluded(t *testing.T) {
	// Per-user single-entrance host uses appid.owner.zone — must never be
	// matched by the shared-template regex (which is anchored on the exact
	// hash8-prefixed FQDN derived from SRR entries). The match regex escapes
	// '.' so the literal per-user host can be searched verbatim.
	perUserHost := `a5be2268\.bob\.olares\.com`
	entrances := []SharedInclusterEntrance{
		{AppID: "a5be2268", EntranceName: "ollama", EntranceID: "a5be2268", Viewer: "alice", PlatformDomain: "olares.com"},
	}
	plugins := buildSharedInclusterTemplates(entrances, "172.16.0.4")
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	host := "a5be2268.alice.olares.com"
	wantMatch := `"^` + strings.ReplaceAll(host, ".", `\.`) + `\.$"`
	body := plugins[0].ToString()
	if !strings.Contains(body, wantMatch) {
		t.Fatalf("expected anchored match %q in %q", wantMatch, body)
	}
	if strings.Contains(body, perUserHost) {
		t.Fatalf("per-user host %q must not appear in match regex: %q", perUserHost, body)
	}
}

func TestInClusterGatewayEnabled_defaultsTrue(t *testing.T) {
	if !inClusterGatewayEnabled(context.Background(), nil) {
		t.Fatal("nil client should default to enabled")
	}
}

func TestBuildSharedInclusterTemplates_clusterIPRotation(t *testing.T) {
	ent := SharedInclusterEntrance{
		AppID: "a5be2268", EntranceName: "api", EntranceID: "a5be2268", Viewer: "alice", PlatformDomain: "olares.com",
	}
	p1 := buildSharedInclusterTemplates([]SharedInclusterEntrance{ent}, "10.0.0.1")
	p2 := buildSharedInclusterTemplates([]SharedInclusterEntrance{ent}, "10.0.0.2")
	if len(p1) != 1 || len(p2) != 1 {
		t.Fatalf("expected single plugin per call, got %d / %d", len(p1), len(p2))
	}
	if !strings.Contains(p1[0].ToString(), "10.0.0.1") {
		t.Fatalf("p1 missing IP: %s", p1[0].ToString())
	}
	if !strings.Contains(p2[0].ToString(), "10.0.0.2") {
		t.Fatalf("p2 missing IP: %s", p2[0].ToString())
	}
}

func TestBuildSharedInclusterTemplates_roundTripCorefile(t *testing.T) {
	plugins := buildSharedInclusterTemplates([]SharedInclusterEntrance{
		{AppID: "bc2bd381", EntranceName: "litellm", EntranceID: "bc2bd381", Viewer: "alice", PlatformDomain: "olares.com"},
	}, "192.168.1.10")
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	// Pre-parse rendering must contain the quoted, anchored match string —
	// CoreDNS expects the match arg to be a quoted regex literal.
	host := "bc2bd381.alice.olares.com"
	wantMatch := `"^` + strings.ReplaceAll(host, ".", `\.`) + `\.$"`
	if pre := plugins[0].ToString(); !strings.Contains(pre, wantMatch) {
		t.Fatalf("pre-parse body missing %q in %q", wantMatch, pre)
	}

	server := &corefile.Server{
		DomPorts: []string{".:53"},
		Plugins:  plugins,
	}
	parsed, err := corefile.New(server.ToString())
	if err != nil {
		t.Fatalf("parse generated corefile: %v", err)
	}
	if len(parsed.Servers) != 1 || len(parsed.Servers[0].Plugins) != 1 {
		t.Fatalf("unexpected parsed structure: %+v", parsed)
	}
	got := parsed.Servers[0].Plugins[0]
	if got.Name != "template" {
		t.Fatalf("plugin name=%s", got.Name)
	}
	// After parsing the parser strips wrapping quotes from quoted args, so
	// the anchored regex appears unquoted in the round-tripped body. The
	// escaped FQDN itself must still be present, along with the gateway IP
	// and fallthrough directive.
	escapedFQDN := strings.ReplaceAll(host, ".", `\.`)
	body := got.ToString()
	if !strings.Contains(body, escapedFQDN) {
		t.Fatalf("missing escaped FQDN %q in %q", escapedFQDN, body)
	}
	if !strings.Contains(body, "192.168.1.10") {
		t.Fatalf("missing gateway IP in %q", body)
	}
	if !strings.Contains(body, "fallthrough") {
		t.Fatalf("shared template must fall through to wildcard: %q", body)
	}
}

func templateMatchesFQDN(plugin *corefile.Plugin, fqdn, ip string) bool {
	if plugin == nil || plugin.Name != "template" {
		return false
	}
	body := plugin.ToString()
	// Pre-parse render keeps the quoted, anchored match regex; CoreDNS
	// requires the surrounding double quotes for the regex literal.
	wantMatch := `"^` + strings.ReplaceAll(fqdn, ".", `\.`) + `\.$"`
	return strings.Contains(body, wantMatch) && strings.Contains(body, ip)
}

func TestSharedInclusterEntrancesFromSRRItems(t *testing.T) {
	prefix := "a5be2268"
	srr := unstructuredSRR("ollama-shared", "shared-a5be2268-ollamav2", map[string]string{
		labelSRRAppID:    "a5be2268",
		labelSRREntrance: "ollamav2",
	}, "gateway", []string{prefix + ".*.olares.com"})

	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srr},
		nil,
		[]string{"alice", "bob"},
	)
	if len(got) != 2 {
		t.Fatalf("expected 2 entrances, got %d", len(got))
	}
	hosts := make([]string, 0, len(got))
	for _, e := range got {
		hosts = append(hosts, e.fqdn())
	}
	sort.Strings(hosts)
	want0 := prefix + ".alice.olares.com"
	want1 := prefix + ".bob.olares.com"
	if hosts[0] != want0 || hosts[1] != want1 {
		t.Fatalf("hosts=%v want %q and %q", hosts, want0, want1)
	}

	perUser := "a5be2268.bob.olares.com"
	plugins := buildSharedInclusterTemplates(got, "10.0.0.5")
	for _, p := range plugins {
		if strings.Contains(p.ToString(), perUser) {
			t.Fatalf("per-user host must not appear: %s", p.ToString())
		}
	}
}

func TestSharedInclusterEntrancesFromSRRItems_passthroughAndDirect(t *testing.T) {
	prefix := "a5be2268"
	srrGateway := unstructuredSRR("ollama-shared", "shared-a5be2268-api", map[string]string{
		labelSRRAppID: "a5be2268", labelSRREntrance: "api",
	}, "gateway", []string{prefix + ".*.olares.com"})
	srrDirect := unstructuredSRR("other-shared", "shared-bc2bd381-litellm", map[string]string{
		labelSRRAppID: "bc2bd381", labelSRREntrance: "litellm",
	}, "direct", []string{"bc2bd381.*.olares.com"})
	passthrough := map[string]struct{}{"litellm-ns": {}}
	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srrGateway, *srrDirect},
		passthrough,
		[]string{"alice"},
	)
	if len(got) != 1 {
		t.Fatalf("expected 1 entrance, got %d (%+v)", len(got), got)
	}
	if got[0].EntranceName != "api" {
		t.Fatalf("got %+v", got[0])
	}
}

func TestSharedInclusterEntrancesFromSRRItems_logicalPatternNotFirst(t *testing.T) {
	prefix := "a5be2268"
	srr := unstructuredSRR("ollama-shared", "shared-a5be2268-api", map[string]string{
		labelSRRAppID: "a5be2268", labelSRREntrance: "api",
	}, "gateway", []string{
		"api.shared.olares.com",
		prefix + ".*.olares.com",
	})
	got := sharedInclusterEntrancesFromSRRItems(
		[]unstructured.Unstructured{*srr},
		nil,
		[]string{"alice"},
	)
	if len(got) != 1 {
		t.Fatalf("expected 1 entrance, got %d (%+v)", len(got), got)
	}
	if got[0].fqdn() != prefix+".alice.olares.com" {
		t.Fatalf("unexpected fqdn %q", got[0].fqdn())
	}
}

func TestSharedInclusterEntrancesFromSRRItems_empty(t *testing.T) {
	if got := sharedInclusterEntrancesFromSRRItems(nil, nil, []string{"alice"}); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestBuildSharedInclusterTemplates_overridesUserWildcard(t *testing.T) {
	// Render incluster server block that mirrors the production layout:
	// shared-FQDN templates appear BEFORE the user-zone wildcard templates.
	// CoreDNS evaluates template handlers in declaration order; an exact-FQDN
	// match must therefore win over the wildcard and the wildcard must still
	// handle other names in the zone via fallthrough on the shared template.
	shared := buildSharedInclusterTemplates([]SharedInclusterEntrance{
		{AppID: "bc2bd381", EntranceName: "litellm", EntranceID: "bc2bd381", Viewer: "alice", PlatformDomain: "olares.com"},
	}, "10.233.38.210")
	if len(shared) != 1 {
		t.Fatalf("expected 1 shared template, got %d", len(shared))
	}
	wildcard := &corefile.Plugin{
		Name: "template",
		Args: []string{"IN", "A", "alice.olares.com"},
		Options: []*corefile.Option{
			{Name: "match", Args: []string{`"\w*\.?(alice.olares.com\.)$"`}},
			{Name: "answer", Args: []string{`"{{ .Name }} 60 IN A 192.168.128.102"`}},
			{Name: "fallthrough"},
		},
	}
	server := &corefile.Server{
		DomPorts: []string{".:53"},
		Plugins:  append(shared, wildcard),
	}
	rendered := server.ToString()
	parsed, err := corefile.New(rendered)
	if err != nil {
		t.Fatalf("CoreDNS rejected generated Corefile: %v\n%s", err, rendered)
	}
	plugins := parsed.Servers[0].Plugins
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if plugins[0].Name != "template" {
		t.Fatalf("first plugin must be the shared template, got %s", plugins[0].Name)
	}
	wantSharedFQDN := "bc2bd381.alice.olares.com"
	escapedFQDN := strings.ReplaceAll(wantSharedFQDN, ".", `\.`)
	body := plugins[0].ToString()
	if !strings.Contains(body, escapedFQDN) {
		t.Fatalf("first template missing escaped shared FQDN: %s", body)
	}
	if !strings.Contains(body, "10.233.38.210") {
		t.Fatalf("first template missing gateway IP: %s", body)
	}
	if !strings.Contains(body, "fallthrough") {
		t.Fatalf("first template must fall through to wildcard: %s", body)
	}
}

func TestRegenerateCorefileInclusterGatewayToggle(t *testing.T) {
	ctx := context.Background()

	t.Run("TC-061 ON state keeps exact shared template", func(t *testing.T) {
		kubeClient, dynamicClient := buildCorefileRegenerateHarness(t, true)
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile failed: %v", err)
		}
		corefileBody := mustReadCorefileConfigMap(t, ctx, kubeClient)
		assertContainsSharedExactTemplate(t, corefileBody)
	})

	t.Run("TC-062 OFF state removes shared templates but keeps user wildcard", func(t *testing.T) {
		kubeClient, dynamicClient := buildCorefileRegenerateHarness(t, false)
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile failed: %v", err)
		}
		corefileBody := mustReadCorefileConfigMap(t, ctx, kubeClient)
		assertNotContainsSharedExactTemplate(t, corefileBody)
		assertContainsUserWildcard(t, corefileBody)
	})

	t.Run("TC-063 OFF->ON switch restores exact shared template", func(t *testing.T) {
		kubeClient, dynamicClient := buildCorefileRegenerateHarness(t, false)
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile OFF failed: %v", err)
		}
		if err := setClusterConfigInclusterGateway(ctx, dynamicClient, true); err != nil {
			t.Fatalf("set ClusterConfig ON failed: %v", err)
		}
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile ON failed: %v", err)
		}
		corefileBody := mustReadCorefileConfigMap(t, ctx, kubeClient)
		assertContainsSharedExactTemplate(t, corefileBody)
	})

	t.Run("TC-064 ON->OFF switch removes exact shared template", func(t *testing.T) {
		kubeClient, dynamicClient := buildCorefileRegenerateHarness(t, true)
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile ON failed: %v", err)
		}
		if err := setClusterConfigInclusterGateway(ctx, dynamicClient, false); err != nil {
			t.Fatalf("set ClusterConfig OFF failed: %v", err)
		}
		if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
			t.Fatalf("RegenerateCorefile OFF failed: %v", err)
		}
		corefileBody := mustReadCorefileConfigMap(t, ctx, kubeClient)
		assertNotContainsSharedExactTemplate(t, corefileBody)
		assertContainsUserWildcard(t, corefileBody)
	})
}

func buildCorefileRegenerateHarness(t *testing.T, inClusterEnabled bool) (*kubefake.Clientset, *dynamicfake.FakeDynamicClient) {
	t.Helper()

	const (
		sharedAppID      = "a5be2268"
		sharedEntrance   = "ollamav2"
		sharedViewer     = "brucedai"
		sharedZone       = "brucedai.olares.com"
		gatewayClusterIP = "10.233.38.210"
	)
	sharedHash := "bc2bd381"

	kubeClient := kubefake.NewSimpleClientset(
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "coredns", Namespace: "kube-system"},
			Data: map[string]string{
				"Corefile": `.:53 {
    errors
    health
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
      pods insecure
      fallthrough in-addr.arpa ip6.arpa
    }
    prometheus :9153
    forward . /etc/resolv.conf
    cache 30
    loop
    reload
    loadbalance
}`,
			},
		},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-control-plane",
				Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: "192.168.128.10"}},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "l4-bfl-proxy-0",
				Namespace: "os-network",
				Labels:    map[string]string{"app": "l4-bfl-proxy"},
			},
			Status: corev1.PodStatus{PodIP: "10.233.3.99"},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "app-gateway-data", Namespace: "app-gateway"},
			Spec:       corev1.ServiceSpec{ClusterIP: gatewayClusterIP},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "litellm-ns"},
		},
	)

	clusterConfig := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.olares.io/v1alpha1",
			"kind":       "ClusterConfig",
			"metadata": map[string]interface{}{
				"name": "cluster",
			},
			"spec": map[string]interface{}{
				"inClusterGatewayEnabled": inClusterEnabled,
			},
		},
	}
	user := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "iam.kubesphere.io/v1alpha2",
			"kind":       "User",
			"metadata": map[string]interface{}{
				"name": sharedViewer,
				"annotations": map[string]interface{}{
					UserAnnotationZoneKey: sharedZone,
					UserIndexAna:          "0",
				},
			},
		},
	}
	srr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.olares.io/v1alpha1",
			"kind":       "SharedRouteRegistry",
			"metadata": map[string]interface{}{
				"name":      "shared-a5be2268-ollamav2",
				"namespace": "litellm-ns",
				"labels": map[string]interface{}{
					labelSRRAppID:    sharedAppID,
					labelSRREntrance: sharedEntrance,
				},
			},
			"spec": map[string]interface{}{
				"routeMode":    "gateway",
				"hostPatterns": []interface{}{sharedHash + ".*.olares.com"},
			},
		},
	}
	app := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "app.bytetrade.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      "shared-app",
				"namespace": "litellm-ns",
			},
			"spec": map[string]interface{}{
				"name":  "shared-app",
				"owner": sharedViewer,
				"sharedEntrances": []interface{}{
					map[string]interface{}{
						"name": "litellm",
						"host": "litellm-svc",
						"port": int64(80),
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			clusterConfigGVR:       "ClusterConfigList",
			sharedRouteRegistryGVR: "SharedRouteRegistryList",
			application.GVR:        "ApplicationList",
			{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}: "UserList",
		},
		clusterConfig, user, srr, app,
	)
	return kubeClient, dynamicClient
}

func setClusterConfigInclusterGateway(ctx context.Context, dynamicClient *dynamicfake.FakeDynamicClient, enabled bool) error {
	obj, err := dynamicClient.Resource(clusterConfigGVR).Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return err
	}
	if err := unstructured.SetNestedField(obj.Object, enabled, "spec", "inClusterGatewayEnabled"); err != nil {
		return err
	}
	_, err = dynamicClient.Resource(clusterConfigGVR).Update(ctx, obj, metav1.UpdateOptions{})
	return err
}

func mustReadCorefileConfigMap(t *testing.T, ctx context.Context, kubeClient *kubefake.Clientset) string {
	t.Helper()
	cm, err := kubeClient.CoreV1().ConfigMaps("kube-system").Get(ctx, "coredns", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get coredns ConfigMap failed: %v", err)
	}
	return cm.Data["Corefile"]
}

func assertContainsSharedExactTemplate(t *testing.T, corefileBody string) {
	t.Helper()
	const sharedEscaped = "bc2bd381\\.brucedai\\.olares\\.com"
	if !strings.Contains(corefileBody, sharedEscaped) {
		t.Fatalf("expected shared exact template match for bc2bd381.brucedai.olares.com, got:\n%s", corefileBody)
	}
}

func assertNotContainsSharedExactTemplate(t *testing.T, corefileBody string) {
	t.Helper()
	const sharedEscaped = "bc2bd381\\.brucedai\\.olares\\.com"
	if strings.Contains(corefileBody, sharedEscaped) {
		t.Fatalf("shared exact template must be removed when disabled, got:\n%s", corefileBody)
	}
}

func assertContainsUserWildcard(t *testing.T, corefileBody string) {
	t.Helper()
	const userTemplate = "template IN A brucedai.olares.com"
	if !strings.Contains(corefileBody, userTemplate) {
		t.Fatalf("expected user wildcard template to remain, got:\n%s", corefileBody)
	}
}

func unstructuredSRR(ns, name string, labels map[string]string, routeMode string, hostPatterns []string) *unstructured.Unstructured {
	labelObj := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		labelObj[k] = v
	}
	patterns := make([]interface{}, len(hostPatterns))
	for i, p := range hostPatterns {
		patterns[i] = p
	}
	obj := map[string]interface{}{
		"apiVersion": "gateway.olares.io/v1alpha1",
		"kind":       "SharedRouteRegistry",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": ns,
			"labels":    labelObj,
		},
		"spec": map[string]interface{}{
			"routeMode":    routeMode,
			"hostPatterns": patterns,
		},
	}
	return &unstructured.Unstructured{Object: obj}
}
