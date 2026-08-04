package watchers

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"bytetrade.io/web3os/tapr/pkg/app/application"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestAppIDFromApplication(t *testing.T) {
	t.Run("user app nginx", func(t *testing.T) {
		app := &application.Application{Spec: application.ApplicationSpec{Name: "nginx"}}
		got := appIDFromApplication(app)
		sum := md5.Sum([]byte("nginx"))
		want := hex.EncodeToString(sum[:])[:8]
		if got != want {
			t.Fatalf("appIDFromApplication(nginx)=%q want %q", got, want)
		}
		if got != appIDFromApplication(app) {
			t.Fatal("appIDFromApplication is not deterministic")
		}
	})

	t.Run("sys app uses name", func(t *testing.T) {
		app := &application.Application{Spec: application.ApplicationSpec{Name: "market", IsSysApp: true}}
		if got := appIDFromApplication(app); got != "market" {
			t.Fatalf("sys app id=%q want market", got)
		}
	})
}

func TestLegacySharedEntranceID(t *testing.T) {
	prefix := legacySharedEntrancePrefix("demo1234")
	if got := legacySharedEntranceID("demo1234", 0); got != prefix+"0" {
		t.Fatalf("single entrance id=%q want %s0", got, prefix)
	}
	if got := legacySharedEntranceID("demo1234", 1); got != prefix+"1" {
		t.Fatalf("second entrance id=%q want %s1", got, prefix)
	}
}

func TestIsV3SharedApp(t *testing.T) {
	v2 := &application.Application{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
	}
	if isV3SharedApp(v2) {
		t.Fatal("app without app-shared label must not be v3 shared")
	}
	v3 := &application.Application{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{labelAppShared: labelAppSharedTrue}},
	}
	if !isV3SharedApp(v3) {
		t.Fatal("app-shared=true must be v3 shared")
	}
}

func TestBuildV2DirectSharedTemplate(t *testing.T) {
	const (
		appid      = "ee434023"
		sharedZone = "shared.olares.com"
	)
	prefix := legacySharedEntrancePrefix(appid)
	plugin := buildV2DirectSharedTemplate(prefix, 0, sharedZone, "10.233.32.202")
	if plugin == nil {
		t.Fatal("expected template plugin")
	}
	body := plugin.ToString()
	wantFQDN := legacySharedEntranceID(appid, 0) + "." + sharedZone
	if !strings.Contains(body, "IN A "+wantFQDN) {
		t.Fatalf("missing fqdn %q in %q", wantFQDN, body)
	}
	wantMatch := fmt.Sprintf(`"%s0.?(%s\.)$"`, prefix, sharedZone)
	if !strings.Contains(body, wantMatch) {
		t.Fatalf("missing match %q in %q", wantMatch, body)
	}
	if !strings.Contains(body, `"{{ .Name }} 60 IN A 10.233.32.202"`) {
		t.Fatalf("missing TTL 60 answer in %q", body)
	}
	if !strings.Contains(body, "fallthrough") {
		t.Fatalf("missing fallthrough in %q", body)
	}
}

func TestV2DirectSharedTemplatesFromCluster_TC_V2_01_singleEntrance(t *testing.T) {
	const (
		appName    = "nginx"
		appOwner   = "brucedai"
		serviceIP  = "10.233.32.202"
		userZone   = "brucedai.olares.com"
		entranceID = "ee434023"
	)

	kubeClient, dynamicClient := buildV2DirectHarness(t, buildV2HarnessOpts{
		userZone: userZone,
		apps: []application.Application{
			{
				ObjectMeta: metav1.ObjectMeta{Name: appName},
				Spec: application.ApplicationSpec{
					Name:   appName,
					Owner:  appOwner,
					Appid:  entranceID,
					Namespace: "nginx-ns",
					SharedEntrances: []application.Entrance{
						{Name: "web", Host: "nginx-svc"},
					},
				},
			},
		},
		namespaces: []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx-shared-ns",
					Labels: map[string]string{
						labelNSAppName:     appName,
						labelNSShared:      "true",
						labelNSInstallUser: appOwner,
					},
				},
			},
		},
		services: map[string]corev1.Service{
			"nginx-shared-ns/nginx-svc": {
				ObjectMeta: metav1.ObjectMeta{Name: "nginx-svc", Namespace: "nginx-shared-ns"},
				Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
			},
		},
	})

	users := []*unstructured.Unstructured{userWithZone("brucedai", userZone)}
	plugins, err := v2DirectSharedTemplatesFromCluster(context.Background(), kubeClient, dynamicClient, toUnstructuredUsers(users))
	if err != nil {
		t.Fatalf("v2DirectSharedTemplatesFromCluster failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 v2 template, got %d", len(plugins))
	}
	wantLegacyID := legacySharedEntranceID(entranceID, 0)
	wantMatch := fmt.Sprintf(`"%s0.?(shared.olares.com\.)$"`, legacySharedEntrancePrefix(entranceID))
	if !templateMatchesRegex(plugins[0], wantMatch, serviceIP) {
		t.Fatalf("template mismatch for %s: %s", wantLegacyID, plugins[0].ToString())
	}
	if !strings.Contains(plugins[0].ToString(), "IN A "+wantLegacyID+".shared.olares.com") {
		t.Fatalf("expected fqdn %s.shared.olares.com in %s", wantLegacyID, plugins[0].ToString())
	}
}

func TestV2DirectSharedTemplatesFromCluster_TC_V2_02_multiEntrance(t *testing.T) {
	const (
		appName   = "dual"
		appOwner  = "brucedai"
		userZone  = "brucedai.olares.com"
		appid     = "58ea3080"
		serviceIP = "10.233.44.55"
	)

	kubeClient, dynamicClient := buildV2DirectHarness(t, buildV2HarnessOpts{
		userZone: userZone,
		apps: []application.Application{
			{
				ObjectMeta: metav1.ObjectMeta{Name: appName},
				Spec: application.ApplicationSpec{
					Name:  appName,
					Owner: appOwner,
					Appid: appid,
					SharedEntrances: []application.Entrance{
						{Name: "a", Host: "dual-a"},
						{Name: "b", Host: "dual-b"},
					},
				},
			},
		},
		namespaces: []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "dual-shared-ns",
					Labels: map[string]string{
						labelNSAppName:     appName,
						labelNSShared:      "true",
						labelNSInstallUser: appOwner,
					},
				},
			},
		},
		services: map[string]corev1.Service{
			"dual-shared-ns/dual-a": {
				ObjectMeta: metav1.ObjectMeta{Name: "dual-a", Namespace: "dual-shared-ns"},
				Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
			},
			"dual-shared-ns/dual-b": {
				ObjectMeta: metav1.ObjectMeta{Name: "dual-b", Namespace: "dual-shared-ns"},
				Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
			},
		},
	})

	users := []*unstructured.Unstructured{userWithZone("brucedai", userZone)}
	plugins, err := v2DirectSharedTemplatesFromCluster(context.Background(), kubeClient, dynamicClient, toUnstructuredUsers(users))
	if err != nil {
		t.Fatalf("v2DirectSharedTemplatesFromCluster failed: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 v2 templates, got %d", len(plugins))
	}
	wantMatches := []string{
		fmt.Sprintf(`"%s0.?(shared.olares.com\.)$"`, legacySharedEntrancePrefix(appid)),
		fmt.Sprintf(`"%s1.?(shared.olares.com\.)$"`, legacySharedEntrancePrefix(appid)),
	}
	for i, want := range wantMatches {
		if !templateMatchesRegex(plugins[i], want, serviceIP) {
			t.Fatalf("template[%d] want %s got %s", i, want, plugins[i].ToString())
		}
	}
}

func TestV2DirectSharedTemplatesFromCluster_TC_V2_03_v3SharedSkipsV2(t *testing.T) {
	const (
		appName   = "ollamav3"
		appOwner  = "brucedai"
		userZone  = "brucedai.olares.com"
		serviceIP = "10.233.66.77"
	)

	kubeClient, dynamicClient := buildV2DirectHarness(t, buildV2HarnessOpts{
		userZone: userZone,
		apps: []application.Application{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: appName,
					Labels: map[string]string{
						labelAppShared: labelAppSharedTrue,
					},
				},
				Spec: application.ApplicationSpec{
					Name:  appName,
					Owner: appOwner,
					Appid: "a5be2268",
					SharedEntrances: []application.Entrance{
						{Name: "api", Host: "ollama-svc"},
					},
				},
			},
		},
		namespaces: []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ollama-shared-ns",
					Labels: map[string]string{
						labelNSAppName:     appName,
						labelNSShared:      "true",
						labelNSInstallUser: appOwner,
					},
				},
			},
		},
		services: map[string]corev1.Service{
			"ollama-shared-ns/ollama-svc": {
				ObjectMeta: metav1.ObjectMeta{Name: "ollama-svc", Namespace: "ollama-shared-ns"},
				Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
			},
		},
	})

	users := []*unstructured.Unstructured{userWithZone("brucedai", userZone)}
	plugins, err := v2DirectSharedTemplatesFromCluster(context.Background(), kubeClient, dynamicClient, toUnstructuredUsers(users))
	if err != nil {
		t.Fatalf("v2DirectSharedTemplatesFromCluster failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("v3 shared app must not emit v2 templates, got %d: %v", len(plugins), plugins[0].ToString())
	}
}

func TestV2DirectSharedTemplatesFromCluster_TC_V2_04_bareAppidEntranceIDExcluded(t *testing.T) {
	const (
		appName   = "nginx"
		appOwner  = "brucedai"
		userZone  = "brucedai.olares.com"
		appid     = "ee434023"
		serviceIP = "10.233.32.202"
	)

	legacySum := md5.Sum([]byte(appid + "shared"))
	legacyPrefix := hex.EncodeToString(legacySum[:])[:8]
	legacyID := legacyPrefix + "0"

	kubeClient, dynamicClient := buildV2DirectHarness(t, buildV2HarnessOpts{
		userZone: userZone,
		apps: []application.Application{
			{
				ObjectMeta: metav1.ObjectMeta{Name: appName},
				Spec: application.ApplicationSpec{
					Name:  appName,
					Owner: appOwner,
					Appid: appid,
					SharedEntrances: []application.Entrance{
						{Name: "web", Host: "nginx-svc"},
					},
				},
			},
		},
		namespaces: []corev1.Namespace{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nginx-shared-ns",
					Labels: map[string]string{
						labelNSAppName:     appName,
						labelNSShared:      "true",
						labelNSInstallUser: appOwner,
					},
				},
			},
		},
		services: map[string]corev1.Service{
			"nginx-shared-ns/nginx-svc": {
				ObjectMeta: metav1.ObjectMeta{Name: "nginx-svc", Namespace: "nginx-shared-ns"},
				Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
			},
		},
	})

	users := []*unstructured.Unstructured{userWithZone("brucedai", userZone)}
	plugins, err := v2DirectSharedTemplatesFromCluster(context.Background(), kubeClient, dynamicClient, toUnstructuredUsers(users))
	if err != nil {
		t.Fatalf("v2DirectSharedTemplatesFromCluster failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 v2 template, got %d", len(plugins))
	}
	body := plugins[0].ToString()
	if !strings.Contains(body, legacyID) {
		t.Fatalf("v2 template must use legacy id %q: %s", legacyID, body)
	}
	if strings.Contains(body, `"^`+appid+`\.shared\.`) {
		t.Fatalf("v2 template must not use bare appid anchored match: %s", body)
	}
	wantMatch := fmt.Sprintf(`"%s0.?(shared.olares.com\.)$"`, legacyPrefix)
	if !templateMatchesRegex(plugins[0], wantMatch, serviceIP) {
		t.Fatalf("expected legacy match %s in %s", wantMatch, body)
	}
}

func TestRegenerateCorefile_TC_V2_05_gatewayOffKeepsV2(t *testing.T) {
	ctx := context.Background()
	kubeClient, dynamicClient := buildV2RegenerateHarness(t, false)

	if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
		t.Fatalf("RegenerateCorefile failed: %v", err)
	}
	body := mustReadCorefileConfigMap(t, ctx, kubeClient)
	assertNotContainsSharedExactTemplate(t, body)
	nginxAppid := appIDFromApplication(&application.Application{Spec: application.ApplicationSpec{Name: "nginx"}}) // harness sets same Appid
	wantMatch := fmt.Sprintf(`"%s0.?(shared.olares.com\.)$"`, legacySharedEntrancePrefix(nginxAppid))
	if !strings.Contains(body, wantMatch) {
		t.Fatalf("expected v2 legacy template when gateway disabled, got:\n%s", body)
	}
}

func TestRegenerateCorefile_TC_V2_06_upgradeStockV2Retained(t *testing.T) {
	ctx := context.Background()
	kubeClient, dynamicClient := buildV2RegenerateHarness(t, true)

	if err := RegenerateCorefile(ctx, kubeClient, dynamicClient); err != nil {
		t.Fatalf("RegenerateCorefile failed: %v", err)
	}
	body := mustReadCorefileConfigMap(t, ctx, kubeClient)
	nginxAppid := appIDFromApplication(&application.Application{Spec: application.ApplicationSpec{Name: "nginx"}}) // harness sets same Appid
	wantMatch := fmt.Sprintf(`"%s0.?(shared.olares.com\.)$"`, legacySharedEntrancePrefix(nginxAppid))
	if !strings.Contains(body, wantMatch) {
		t.Fatalf("upgrade stock v2 app must retain v2 legacy template, got:\n%s", body)
	}
	if !strings.Contains(body, "10.233.32.202") {
		t.Fatalf("v2 template must answer Service ClusterIP, got:\n%s", body)
	}
}

type buildV2HarnessOpts struct {
	userZone   string
	apps       []application.Application
	namespaces []corev1.Namespace
	services   map[string]corev1.Service
}

func buildV2DirectHarness(t *testing.T, opts buildV2HarnessOpts) (*kubefake.Clientset, *dynamicfake.FakeDynamicClient) {
	t.Helper()

	objs := []runtime.Object{
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "coredns", Namespace: "kube-system"},
			Data:       map[string]string{"Corefile": ".:53 { errors }"},
		},
	}
	for i := range opts.namespaces {
		objs = append(objs, &opts.namespaces[i])
	}
	for key, svc := range opts.services {
		parts := strings.SplitN(key, "/", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid service key %q", key)
		}
		s := svc
		s.ObjectMeta.Namespace = parts[0]
		s.ObjectMeta.Name = parts[1]
		objs = append(objs, &s)
	}

	kubeClient := kubefake.NewSimpleClientset(objs...)

	appObjs := make([]runtime.Object, 0, len(opts.apps))
	for i := range opts.apps {
		appObjs = append(appObjs, applicationUnstructured(&opts.apps[i]))
	}

	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			application.GVR: "ApplicationList",
			calicoIPPoolGVR: "IPPoolList",
		},
		appObjs...,
	)
	return kubeClient, dynamicClient
}

func buildV2RegenerateHarness(t *testing.T, inClusterEnabled bool) (*kubefake.Clientset, *dynamicfake.FakeDynamicClient) {
	t.Helper()

	const (
		appName   = "nginx"
		appOwner  = "brucedai"
		sharedZone = "brucedai.olares.com"
		serviceIP = "10.233.32.202"
	)

	kubeClient, dynamicClient := buildCorefileRegenerateHarness(t, inClusterEnabled)
	ctx := context.Background()

	if _, err := kubeClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-shared-ns",
			Labels: map[string]string{
				labelNSAppName:     appName,
				labelNSShared:      "true",
				labelNSInstallUser: appOwner,
			},
		},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create shared namespace: %v", err)
	}
	if _, err := kubeClient.CoreV1().Services("nginx-shared-ns").Create(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx-svc", Namespace: "nginx-shared-ns"},
		Spec:       corev1.ServiceSpec{ClusterIP: serviceIP},
	}, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create shared service: %v", err)
	}
	if _, err := dynamicClient.Resource(application.GVR).Create(ctx, applicationUnstructured(&application.Application{
		ObjectMeta: metav1.ObjectMeta{Name: appName},
		Spec: application.ApplicationSpec{
			Name:  appName,
			Owner: appOwner,
			Appid: appIDFromApplication(&application.Application{Spec: application.ApplicationSpec{Name: appName}}),
			SharedEntrances: []application.Entrance{
				{Name: "web", Host: "nginx-svc"},
			},
		},
	}), metav1.CreateOptions{}); err != nil {
		t.Fatalf("create application: %v", err)
	}
	return kubeClient, dynamicClient
}

func userWithZone(name, zone string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "iam.kubesphere.io/v1alpha2",
			"kind":       "User",
			"metadata": map[string]interface{}{
				"name": name,
				"annotations": map[string]interface{}{
					UserAnnotationZoneKey: zone,
					UserIndexAna:          "0",
				},
			},
		},
	}
}

func toUnstructuredUsers(users []*unstructured.Unstructured) []unstructured.Unstructured {
	out := make([]unstructured.Unstructured, len(users))
	for i, u := range users {
		out[i] = *u
	}
	return out
}

func applicationUnstructured(app *application.Application) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name": app.Name,
	}
	if len(app.Labels) > 0 {
		labels := make(map[string]interface{}, len(app.Labels))
		for k, v := range app.Labels {
			labels[k] = v
		}
		metadata["labels"] = labels
	}

	entrances := make([]interface{}, len(app.Spec.SharedEntrances))
	for i, e := range app.Spec.SharedEntrances {
		entrances[i] = map[string]interface{}{
			"name": e.Name,
			"host": e.Host,
		}
	}

	spec := map[string]interface{}{
		"name":            app.Spec.Name,
		"owner":           app.Spec.Owner,
		"sharedEntrances": entrances,
	}
	if app.Spec.Appid != "" {
		spec["appid"] = app.Spec.Appid
	}
	if app.Spec.Namespace != "" {
		spec["namespace"] = app.Spec.Namespace
	}
	if app.Spec.IsSysApp {
		spec["isSysApp"] = true
	}

	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "app.bytetrade.io/v1alpha1",
		"kind":       "Application",
		"metadata":   metadata,
		"spec":       spec,
	}}
}
