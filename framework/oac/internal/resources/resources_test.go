package resources

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliresource "k8s.io/cli-runtime/pkg/resource"
)

// mustQuantity panics on parse failure. Tests should only feed it strings
// that are known-good.
func mustQuantity(s string) apiresource.Quantity {
	q, err := apiresource.ParseQuantity(s)
	if err != nil {
		panic(err)
	}
	return q
}

func newDeployment(name string, containers ...corev1.Container) *cliresource.Info {
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDeployment,
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: dep}
}

func newStatefulSet(name string, containers ...corev1.Container) *cliresource.Info {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindStatefulSet,
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: sts}
}

func newDaemonSet(name string, containers ...corev1.Container) *cliresource.Info {
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDaemonSet,
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Containers: containers},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: ds}
}

// goodContainer is a container that satisfies CheckResourceLimits — every
// dimension has a request and limit.
func goodContainer(name, cpuReq, cpuLim, memReq, memLim string, mounts ...corev1.VolumeMount) corev1.Container {
	return corev1.Container{
		Name:  name,
		Image: "registry/example:" + name,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    mustQuantity(cpuReq),
				corev1.ResourceMemory: mustQuantity(memReq),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    mustQuantity(cpuLim),
				corev1.ResourceMemory: mustQuantity(memLim),
			},
		},
		VolumeMounts: mounts,
	}
}

// ---- ExtractWorkloadImages ----

// TestExtractWorkloadImages_DeploymentAndStatefulSet documents that only
// Deployment and StatefulSet images feed the output, matching the workload
// kinds walkPodContainers actually inspects. DaemonSet is intentionally
// skipped; see TestExtractWorkloadImages_SkipsDaemonSet.
func TestExtractWorkloadImages_DeploymentAndStatefulSet(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", corev1.Container{Image: "img/a:1"}),
		newStatefulSet("b", corev1.Container{Image: "img/b:1"}),
	}
	got := ExtractWorkloadImages(list)
	want := []string{"img/a:1", "img/b:1"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// TestExtractWorkloadImages_SkipsDaemonSet pins the behaviour that keeps
// image scanning aligned with the resource-limits / upload-mount checks in
// walkPodContainers: a DaemonSet would slip past those checks, so its
// images must not appear in the pull list either.
func TestExtractWorkloadImages_SkipsDaemonSet(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", corev1.Container{Image: "img/a:1"}),
		newDaemonSet("d", corev1.Container{Image: "img/d:1"}),
	}
	got := ExtractWorkloadImages(list)
	if len(got) != 1 || got[0] != "img/a:1" {
		t.Fatalf("DaemonSet image leaked into output: %v", got)
	}
}

func TestExtractWorkloadImages_DedupAndSort(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a",
			corev1.Container{Image: "img/x:2"},
			corev1.Container{Image: "img/x:2"},
			corev1.Container{Image: "img/a:1"},
		),
	}
	got := ExtractWorkloadImages(list)
	want := []string{"img/a:1", "img/x:2"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractWorkloadImages_IgnoresUnknownKinds(t *testing.T) {
	// A Service has no containers; it must be skipped silently.
	svc := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "svc"},
	}
	list := kube.ResourceList{
		{Name: "svc", Object: svc},
		newDeployment("a", corev1.Container{Image: "img/a:1"}),
	}
	got := ExtractWorkloadImages(list)
	if len(got) != 1 || got[0] != "img/a:1" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestExtractWorkloadImages_SkipsInitContainers(t *testing.T) {
	dep := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: KindDeployment, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "a"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{Image: "img/init:1"}},
					Containers:     []corev1.Container{{Image: "img/main:1"}},
				},
			},
		},
	}
	list := kube.ResourceList{{Name: "a", Object: dep}}
	got := ExtractWorkloadImages(list)
	if len(got) != 1 || got[0] != "img/main:1" {
		t.Fatalf("init container leaked into output: %v", got)
	}
}

// ---- MergeImages ----

func TestMergeImages(t *testing.T) {
	got := MergeImages(
		[]string{"a", "b", "", "c"},
		[]string{"b", "d", ""},
	)
	want := []string{"a", "b", "c", "d"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestMergeImages_EmptyInputs(t *testing.T) {
	if got := MergeImages(nil, nil); len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

// ---- CheckResourceLimits ----

func TestCheckResourceLimits_Happy(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", goodContainer("c1", "100m", "200m", "100Mi", "200Mi")),
	}
	limits := ResourceLimits{
		RequiredCPU: "100m", RequiredMemory: "100Mi",
		LimitedCPU: "200m", LimitedMemory: "200Mi",
	}
	if err := CheckResourceLimits(list, limits); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckResourceLimits_AppRequiredAboveLimited(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", goodContainer("c1", "100m", "200m", "100Mi", "200Mi")),
	}
	limits := ResourceLimits{
		RequiredCPU: "500m", RequiredMemory: "500Mi",
		LimitedCPU: "100m", LimitedMemory: "100Mi",
	}
	err := CheckResourceLimits(list, limits)
	if err == nil {
		t.Fatal("expected error: required > limited")
	}
	if !strings.Contains(err.Error(), "spec.requiredCpu") || !strings.Contains(err.Error(), "spec.requiredMemory") {
		t.Fatalf("error should report both dimensions: %v", err)
	}
}

func TestCheckResourceLimits_MissingContainerRequest(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", corev1.Container{
			Name: "no-resources",
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    mustQuantity("100m"),
					corev1.ResourceMemory: mustQuantity("100Mi"),
				},
			},
		}),
	}
	limits := ResourceLimits{
		RequiredCPU: "100m", RequiredMemory: "100Mi",
		LimitedCPU: "200m", LimitedMemory: "200Mi",
	}
	err := CheckResourceLimits(list, limits)
	if err == nil {
		t.Fatal("expected error: container missing requests")
	}
	if !strings.Contains(err.Error(), "must set cpu request") || !strings.Contains(err.Error(), "must set memory request") {
		t.Fatalf("error should mention missing requests: %v", err)
	}
}

func TestCheckResourceLimits_ContainerLimitOverApp(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", goodContainer("c1", "100m", "1", "100Mi", "1Gi")),
	}
	limits := ResourceLimits{
		RequiredCPU: "100m", RequiredMemory: "100Mi",
		LimitedCPU: "200m", LimitedMemory: "200Mi",
	}
	err := CheckResourceLimits(list, limits)
	if err == nil {
		t.Fatal("expected aggregate cap error")
	}
	if !strings.Contains(err.Error(), "spec.limitedCpu") || !strings.Contains(err.Error(), "spec.limitedMemory") {
		t.Fatalf("error should mention spec caps: %v", err)
	}
}

// ---- CheckUploadConfig ----

func TestCheckUploadConfig_DisabledWhenEmpty(t *testing.T) {
	if err := CheckUploadConfig(nil, ""); err != nil {
		t.Fatalf("empty dest must be a no-op: %v", err)
	}
}

func TestCheckUploadConfig_FoundMatchingMount(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a",
			corev1.Container{
				Name:         "c1",
				VolumeMounts: []corev1.VolumeMount{{MountPath: "/data/uploads"}},
			},
		),
	}
	if err := CheckUploadConfig(list, "/data/uploads"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	// Path normalisation: trailing slash should still match.
	if err := CheckUploadConfig(list, "/data/uploads/"); err != nil {
		t.Fatalf("path should be cleaned: %v", err)
	}
}

func TestCheckUploadConfig_NotFound(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("a", corev1.Container{
			Name:         "c1",
			VolumeMounts: []corev1.VolumeMount{{MountPath: "/different"}},
		}),
	}
	err := CheckUploadConfig(list, "/data/uploads")
	if err == nil {
		t.Fatal("expected error: no matching volume mount")
	}
	if !strings.Contains(err.Error(), "options.upload.dest") {
		t.Fatalf("error should mention options.upload.dest: %v", err)
	}
}

// ---- CheckDeploymentName ----

func TestCheckDeploymentName_NonAppSkipped(t *testing.T) {
	if err := CheckDeploymentName(nil, "middleware", "foo"); err != nil {
		t.Fatalf("non-app config must skip: %v", err)
	}
}

func TestCheckDeploymentName_FindsDeployment(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("firefox", corev1.Container{Image: "x"}),
	}
	if err := CheckDeploymentName(list, "app", "firefox"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckDeploymentName_FindsStatefulSet(t *testing.T) {
	list := kube.ResourceList{
		newStatefulSet("firefox", corev1.Container{Image: "x"}),
	}
	if err := CheckDeploymentName(list, "app", "firefox"); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestCheckDeploymentName_NotFound(t *testing.T) {
	list := kube.ResourceList{
		newDeployment("other", corev1.Container{Image: "x"}),
	}
	err := CheckDeploymentName(list, "app", "firefox")
	if err == nil {
		t.Fatal("expected missing-named-workload error")
	}
	if !strings.Contains(err.Error(), `"firefox"`) {
		t.Fatalf("error should mention app name: %v", err)
	}
}

// ---- LoadForbiddenRules / CheckServiceAccountRules ----

func TestLoadForbiddenRules_Default(t *testing.T) {
	rules, err := LoadForbiddenRules("")
	if err != nil {
		t.Fatalf("LoadForbiddenRules: %v", err)
	}
	if len(rules) == 0 {
		t.Fatal("default rules must not be empty")
	}
	// Sanity: the default policy targets nodes & networkpolicies.
	found := false
	for _, r := range rules {
		for _, res := range r.Resources {
			if res == "nodes" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("default rules should mention 'nodes'")
	}
}

// roleWith builds a Role with the given rules.
func roleWith(name string, rules ...rbacv1.PolicyRule) *cliresource.Info {
	r := &rbacv1.Role{
		TypeMeta:   metav1.TypeMeta{Kind: KindRole, APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Rules:      rules,
	}
	return &cliresource.Info{Name: name, Object: r}
}

func roleBindingFor(name, roleName string) *cliresource.Info {
	rb := &rbacv1.RoleBinding{
		TypeMeta:   metav1.TypeMeta{Kind: KindRoleBinding, APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Subjects:   []rbacv1.Subject{{Kind: KindServiceAccount, Name: "sa"}},
		RoleRef:    rbacv1.RoleRef{Kind: KindRole, Name: roleName},
	}
	return &cliresource.Info{Name: name, Object: rb}
}

func TestCheckServiceAccountRules_NoBindingIsOK(t *testing.T) {
	list := kube.ResourceList{
		roleWith("read-pods", rbacv1.PolicyRule{
			APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"},
		}),
	}
	forbidden, _ := LoadForbiddenRules("")
	if err := CheckServiceAccountRules(list, forbidden); err != nil {
		t.Fatalf("unbound role must not trip the check: %v", err)
	}
}

func TestCheckServiceAccountRules_AllowsBenignBoundRole(t *testing.T) {
	list := kube.ResourceList{
		roleWith("read-pods", rbacv1.PolicyRule{
			APIGroups: []string{""}, Resources: []string{"pods"}, Verbs: []string{"get"},
		}),
		roleBindingFor("rb", "read-pods"),
	}
	forbidden, _ := LoadForbiddenRules("")
	if err := CheckServiceAccountRules(list, forbidden); err != nil {
		t.Fatalf("benign bound role must pass: %v", err)
	}
}

// TestCheckServiceAccountRules_RejectsForbidden covers the positive-side hole
// that let B5 (flipped rulesAllow boolean) slip through review: a Role bound
// to a ServiceAccount that asks for "delete nodes" — which the default
// policy forbids — MUST return an error.
func TestCheckServiceAccountRules_RejectsForbidden(t *testing.T) {
	list := kube.ResourceList{
		roleWith("bad", rbacv1.PolicyRule{
			APIGroups: []string{""}, Resources: []string{"nodes"}, Verbs: []string{"delete"},
		}),
		roleBindingFor("rb", "bad"),
	}
	forbidden, _ := LoadForbiddenRules("")
	if err := CheckServiceAccountRules(list, forbidden); err == nil {
		t.Fatal("delete-nodes request should be flagged against the default forbidden policy")
	}
}

// TestCheckServiceAccountRules_RejectsForbiddenWithMultipleRules is the
// regression test for B5: when the forbidden policy contains more than one
// rule, a request that matches ONE of them must still be rejected. Before
// the fix, rulesAllow returned "allowed" as soon as any forbidden rule
// failed to cover the request, so this test would have passed silently.
func TestCheckServiceAccountRules_RejectsForbiddenWithMultipleRules(t *testing.T) {
	list := kube.ResourceList{
		roleWith("bad", rbacv1.PolicyRule{
			APIGroups: []string{""}, Resources: []string{"nodes"}, Verbs: []string{"delete"},
		}),
		roleBindingFor("rb", "bad"),
	}
	forbidden := []rbacv1.PolicyRule{
		{APIGroups: []string{"*"}, Resources: []string{"nodes"}, Verbs: []string{"delete"}},
		// A second, unrelated forbidden rule that does NOT cover the
		// request. The old implementation would short-circuit here and
		// wrongly report "allowed".
		{APIGroups: []string{"*"}, Resources: []string{"networkpolicies"}, Verbs: []string{"create"}},
	}
	if err := CheckServiceAccountRules(list, forbidden); err == nil {
		t.Fatal("delete-nodes request should still be flagged when a second, non-matching forbidden rule is present")
	}
}

// TestCheckServiceAccountRules_AllowsRequestCoveredByOneNonResourceURL is a
// B6 regression test: a forbidden rule that grants /metrics and /healthz/*
// must not reject a request for /healthz/live just because /metrics does
// not cover it. Under the old loop ordering, the check short-circuited on
// the first non-matching ruleURL.
func TestCheckServiceAccountRules_AllowsRequestCoveredByOneNonResourceURL(t *testing.T) {
	list := kube.ResourceList{
		roleWith("probe", rbacv1.PolicyRule{
			Verbs:           []string{"get"},
			NonResourceURLs: []string{"/healthz/live"},
		}),
		roleBindingFor("rb", "probe"),
	}
	// A forbidden rule that DOES cover /healthz/live via the wildcard.
	// The request should be rejected.
	forbidden := []rbacv1.PolicyRule{{
		Verbs:           []string{"get"},
		NonResourceURLs: []string{"/metrics", "/healthz/*"},
	}}
	if err := CheckServiceAccountRules(list, forbidden); err == nil {
		t.Fatal("request on /healthz/live should be covered by the /healthz/* wildcard and therefore flagged")
	}
}

