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

// TestExtractWorkloadImages_DedupesAcrossInitAndMain reuses the same
// image across an init container and a main container to assert the
// dedup step in ExtractWorkloadImages collapses them to a single entry.
func TestExtractWorkloadImages_DedupesAcrossInitAndMain(t *testing.T) {
	dep := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: KindDeployment, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "d"},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{Name: "shared", Image: "img/x:2"},
					},
					Containers: []corev1.Container{
						{Name: "main", Image: "img/x:2"},
					},
				},
			},
		},
	}
	list := kube.ResourceList{{Name: "d", Object: dep}}

	got := ExtractWorkloadImages(list)
	if len(got) != 1 || got[0] != "img/x:2" {
		t.Fatalf("init+main duplicate must dedup to one entry, got %v", got)
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

// TestExtractWorkloadImages_IncludesInitContainersOnControllers pins
// the contract that init containers contribute to the pull list. They
// are pulled before the main containers on every node that schedules
// the workload, so any tool that primes a registry must see them too.
func TestExtractWorkloadImages_IncludesInitContainersOnControllers(t *testing.T) {
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
	want := []string{"img/init:1", "img/main:1"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
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

// ---- CheckHostPath ----

// hostPathVolume builds a corev1.Volume backed by hostPath.
func hostPathVolume(name, path string) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: path}},
	}
}

// emptyDirVolume builds a corev1.Volume backed by emptyDir (no hostPath).
func emptyDirVolume(name string) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
}

func deploymentWithStrategy(name string, t appsv1.DeploymentStrategyType, volumes ...corev1.Volume) *cliresource.Info {
	dep := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: KindDeployment, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{Type: t},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Volumes: volumes},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: dep}
}

func statefulSetWithStrategy(name string, t appsv1.StatefulSetUpdateStrategyType, volumes ...corev1.Volume) *cliresource.Info {
	sts := &appsv1.StatefulSet{
		TypeMeta:   metav1.TypeMeta{Kind: KindStatefulSet, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appsv1.StatefulSetSpec{
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: t},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{Volumes: volumes},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: sts}
}

func TestCheckHostPath_EmptyListIsOK(t *testing.T) {
	if err := CheckHostPath(nil); err != nil {
		t.Fatalf("empty list must be accepted: %v", err)
	}
}

func TestCheckHostPath_DeploymentRecreateWithHostPathOK(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithStrategy("ok", appsv1.RecreateDeploymentStrategyType,
			hostPathVolume("data", "/var/lib/data")),
	}
	if err := CheckHostPath(list); err != nil {
		t.Fatalf("Recreate + hostPath must pass: %v", err)
	}
}

func TestCheckHostPath_DeploymentRollingWithoutHostPathOK(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithStrategy("ok", appsv1.RollingUpdateDeploymentStrategyType,
			emptyDirVolume("cache")),
	}
	if err := CheckHostPath(list); err != nil {
		t.Fatalf("RollingUpdate without hostPath must pass: %v", err)
	}
}

func TestCheckHostPath_DeploymentRollingWithHostPathFails(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithStrategy("bad", appsv1.RollingUpdateDeploymentStrategyType,
			hostPathVolume("data", "/var/lib/data")),
	}
	err := CheckHostPath(list)
	if err == nil {
		t.Fatal("RollingUpdate + hostPath must fail")
	}
	if !strings.Contains(err.Error(), "deployment bad") || !strings.Contains(err.Error(), "/var/lib/data") {
		t.Fatalf("error should mention workload and path: %v", err)
	}
}

func TestCheckHostPath_DeploymentEmptyStrategyTreatedAsRolling(t *testing.T) {
	// Empty Strategy.Type defaults to RollingUpdate on Kubernetes' side,
	// so the check must surface the conflict.
	list := kube.ResourceList{
		deploymentWithStrategy("bad", "", hostPathVolume("data", "/var/lib/data")),
	}
	if err := CheckHostPath(list); err == nil {
		t.Fatal("empty Strategy.Type must be treated as rolling update")
	}
}

func TestCheckHostPath_StatefulSetOnDeleteWithHostPathOK(t *testing.T) {
	list := kube.ResourceList{
		statefulSetWithStrategy("ok", appsv1.OnDeleteStatefulSetStrategyType,
			hostPathVolume("data", "/var/lib/data")),
	}
	if err := CheckHostPath(list); err != nil {
		t.Fatalf("OnDelete + hostPath must pass: %v", err)
	}
}

func TestCheckHostPath_StatefulSetRollingWithHostPathFails(t *testing.T) {
	list := kube.ResourceList{
		statefulSetWithStrategy("bad", appsv1.RollingUpdateStatefulSetStrategyType,
			hostPathVolume("data", "/var/lib/data")),
	}
	err := CheckHostPath(list)
	if err == nil {
		t.Fatal("StatefulSet RollingUpdate + hostPath must fail")
	}
	if !strings.Contains(err.Error(), "statefulset bad") || !strings.Contains(err.Error(), "/var/lib/data") {
		t.Fatalf("error should mention workload and path: %v", err)
	}
}

func TestCheckHostPath_StatefulSetEmptyStrategyTreatedAsRolling(t *testing.T) {
	list := kube.ResourceList{
		statefulSetWithStrategy("bad", "", hostPathVolume("data", "/var/lib/data")),
	}
	if err := CheckHostPath(list); err == nil {
		t.Fatal("empty UpdateStrategy.Type must be treated as rolling update")
	}
}

func TestCheckHostPath_AggregatesMultipleViolations(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithStrategy("bad-dep", appsv1.RollingUpdateDeploymentStrategyType,
			hostPathVolume("data", "/var/lib/data"),
			hostPathVolume("cache", "/var/lib/cache"),
		),
		statefulSetWithStrategy("bad-sts", appsv1.RollingUpdateStatefulSetStrategyType,
			hostPathVolume("logs", "/var/log/app")),
	}
	err := CheckHostPath(list)
	if err == nil {
		t.Fatal("expected aggregated error")
	}
	for _, want := range []string{"/var/lib/data", "/var/lib/cache", "/var/log/app"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error should mention path %s: %v", want, err)
		}
	}
}

// ---- CheckResourceNamespace ----

// withNamespace mutates an existing *cliresource.Info to set Namespace
// and aligns the embedded object's metadata.namespace so tests don't have
// to know whether the check reads from Info.Namespace or the object.
func withNamespace(info *cliresource.Info, ns string) *cliresource.Info {
	info.Namespace = ns
	if accessor, ok := info.Object.(metav1.Object); ok {
		accessor.SetNamespace(ns)
	}
	return info
}

// newConfigMap returns a namespaced object used to exercise the
// non-workload code path of CheckResourceNamespace.
func newConfigMap(name string) *cliresource.Info {
	cm := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	return &cliresource.Info{Name: name, Object: cm}
}

// newClusterRole returns a cluster-scoped object so we can assert the
// check skips it regardless of Namespace contents.
func newClusterRole(name string) *cliresource.Info {
	cr := &rbacv1.ClusterRole{
		TypeMeta:   metav1.TypeMeta{Kind: KindClusterRole, APIVersion: "rbac.authorization.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}
	return &cliresource.Info{Name: name, Object: cr}
}

func TestCheckResourceNamespace_EmptyListIsOK(t *testing.T) {
	if err := CheckResourceNamespace(nil); err != nil {
		t.Fatalf("empty list must be accepted: %v", err)
	}
}

func TestCheckResourceNamespace_WorkloadInAppNamespaceOK(t *testing.T) {
	list := kube.ResourceList{
		withNamespace(newDeployment("d"), AppNamespace),
		withNamespace(newStatefulSet("s"), AppNamespace),
		withNamespace(newDaemonSet("ds"), AppNamespace),
	}
	if err := CheckResourceNamespace(list); err != nil {
		t.Fatalf("workloads in %q must pass: %v", AppNamespace, err)
	}
}

func TestCheckResourceNamespace_WorkloadInUserSystemIsRejected(t *testing.T) {
	list := kube.ResourceList{
		withNamespace(newDeployment("bad"), "user-system-alice"),
	}
	err := CheckResourceNamespace(list)
	if err == nil {
		t.Fatal("workloads must NOT be allowed in user-system-* namespaces")
	}
	if !strings.Contains(err.Error(), "illegal namespace: user-system-alice") ||
		!strings.Contains(err.Error(), "for Deployment") ||
		!strings.Contains(err.Error(), "name bad") {
		t.Fatalf("error should pinpoint the workload, got: %v", err)
	}
}

func TestCheckResourceNamespace_WorkloadInOtherNamespaceIsRejected(t *testing.T) {
	for _, ns := range []string{"default", "kube-system", "my-app"} {
		list := kube.ResourceList{
			withNamespace(newDeployment("bad"), ns),
		}
		err := CheckResourceNamespace(list)
		if err == nil {
			t.Fatalf("workloads in %q must fail", ns)
		}
		if !strings.Contains(err.Error(), "illegal namespace: "+ns) {
			t.Fatalf("error should mention the namespace, got: %v", err)
		}
	}
}

func TestCheckResourceNamespace_OtherResourceInAppOrUserSystemOK(t *testing.T) {
	list := kube.ResourceList{
		withNamespace(newConfigMap("cfg"), AppNamespace),
		withNamespace(newConfigMap("user-cfg"), "user-system-alice"),
	}
	if err := CheckResourceNamespace(list); err != nil {
		t.Fatalf("non-workload resources in app-namespace / user-system-* must pass: %v", err)
	}
}

func TestCheckResourceNamespace_OtherResourceInForeignNamespaceIsRejected(t *testing.T) {
	list := kube.ResourceList{
		withNamespace(newConfigMap("cfg"), "default"),
	}
	err := CheckResourceNamespace(list)
	if err == nil {
		t.Fatal("non-workload resources outside the allowed set must fail")
	}
	if !strings.Contains(err.Error(), "illegal namespace: default") ||
		!strings.Contains(err.Error(), "for ConfigMap") {
		t.Fatalf("error should pinpoint the resource, got: %v", err)
	}
}

func TestCheckResourceNamespace_ClusterScopedSkipped(t *testing.T) {
	// ClusterRole and an explicitly empty-namespace Deployment both have
	// Namespace == "" and must not be flagged by the check.
	list := kube.ResourceList{
		newClusterRole("cluster-r"),
		newDeployment("workload-without-ns"),
	}
	if err := CheckResourceNamespace(list); err != nil {
		t.Fatalf("cluster-scoped / empty-namespace resources must be skipped: %v", err)
	}
}

func TestCheckResourceNamespace_AggregatesMultipleViolations(t *testing.T) {
	list := kube.ResourceList{
		withNamespace(newDeployment("bad-dep"), "default"),
		withNamespace(newConfigMap("bad-cm"), "kube-system"),
		withNamespace(newStatefulSet("bad-sts"), "user-system-alice"),
	}
	err := CheckResourceNamespace(list)
	if err == nil {
		t.Fatal("expected aggregated error for multiple violations")
	}
	for _, want := range []string{"bad-dep", "bad-cm", "bad-sts"} {
		if !strings.Contains(err.Error(), "name "+want) {
			t.Fatalf("error should mention %q: %v", want, err)
		}
	}
}

// ---- CheckSecurityContextForNonBeclabImage ----

func boolPtr(b bool) *bool { return &b }
func int64Ptr(i int64) *int64 { return &i }

// containerWithSC returns a container with the given image and an
// explicit SecurityContext so tests can craft the exact pointer shape
// the check inspects.
func containerWithSC(name, image string, sc *corev1.SecurityContext) corev1.Container {
	return corev1.Container{Name: name, Image: image, SecurityContext: sc}
}

// deploymentWithPodSpec wraps a corev1.PodSpec inside a Deployment so we
// can feed walkPodSpecs the exact pod-level securityContext / container
// list a test needs.
func deploymentWithPodSpec(name string, spec corev1.PodSpec) *cliresource.Info {
	dep := &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: KindDeployment, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
	}
	return &cliresource.Info{Name: name, Object: dep}
}

func daemonSetWithPodSpec(name string, spec corev1.PodSpec) *cliresource.Info {
	ds := &appsv1.DaemonSet{
		TypeMeta:   metav1.TypeMeta{Kind: KindDaemonSet, APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
	}
	return &cliresource.Info{Name: name, Object: ds}
}

func TestIsBeclabImage(t *testing.T) {
	cases := []struct {
		image string
		want  bool
	}{
		{"beclab/foo", true},
		{"beclab/foo:1.0", true},
		{"docker.io/beclab/foo", true},
		{"docker.io/beclab/foo:1.0", true},
		{"registry.example.com/beclab/foo:1.0", true},
		{"my/beclab/foo", true},

		{"not-beclab/foo", false},
		{"beclab-bar/foo", false},
		{"foo", false},
		{"foo:1.0", false},
		{"alpine:3.20", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isBeclabImage(tc.image); got != tc.want {
			t.Errorf("isBeclabImage(%q) = %v, want %v", tc.image, got, tc.want)
		}
	}
}

func TestCheckSecurityContext_EmptyListIsOK(t *testing.T) {
	if err := CheckSecurityContextForNonBeclabImage(nil); err != nil {
		t.Fatalf("empty list must be accepted: %v", err)
	}
}

func TestCheckSecurityContext_NilSecurityContextOK(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main", Image: "third-party/foo:1.0"}, // no SC
			},
		}),
	}
	if err := CheckSecurityContextForNonBeclabImage(list); err != nil {
		t.Fatalf("no SecurityContext must pass: %v", err)
	}
}

func TestCheckSecurityContext_BeclabImageBypassed(t *testing.T) {
	// Even with the worst possible SC, beclab/* must be exempt.
	bad := &corev1.SecurityContext{
		Privileged:   boolPtr(true),
		RunAsUser:    int64Ptr(0),
		RunAsNonRoot: boolPtr(false),
	}
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			Containers: []corev1.Container{
				containerWithSC("c", "beclab/foo:1.0", bad),
				containerWithSC("c2", "docker.io/beclab/bar:1.0", bad),
			},
		}),
	}
	if err := CheckSecurityContextForNonBeclabImage(list); err != nil {
		t.Fatalf("beclab/* images must be exempt regardless of SC: %v", err)
	}
}

func TestCheckSecurityContext_ContainerPrivilegedFails(t *testing.T) {
	sc := &corev1.SecurityContext{Privileged: boolPtr(true)}
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			Containers: []corev1.Container{
				containerWithSC("bad", "third-party/foo:1.0", sc),
			},
		}),
	}
	err := CheckSecurityContextForNonBeclabImage(list)
	if err == nil {
		t.Fatal("privileged=true on non-beclab image must fail")
	}
	if !strings.Contains(err.Error(), "third-party/foo:1.0") ||
		!strings.Contains(err.Error(), "container bad") ||
		!strings.Contains(err.Error(), "Deployment d") {
		t.Fatalf("error should pinpoint image+container+workload: %v", err)
	}
}

func TestCheckSecurityContext_ContainerRunAsRootFails(t *testing.T) {
	cases := []*corev1.SecurityContext{
		{RunAsUser: int64Ptr(0)},
		{RunAsNonRoot: boolPtr(false)},
	}
	for _, sc := range cases {
		list := kube.ResourceList{
			deploymentWithPodSpec("d", corev1.PodSpec{
				Containers: []corev1.Container{
					containerWithSC("bad", "third-party/foo:1.0", sc),
				},
			}),
		}
		if err := CheckSecurityContextForNonBeclabImage(list); err == nil {
			t.Fatalf("non-beclab container with SC %#v must fail", sc)
		}
	}
}

func TestCheckSecurityContext_ContainerRunAsNonZeroAccepted(t *testing.T) {
	sc := &corev1.SecurityContext{
		RunAsUser:    int64Ptr(1000),
		RunAsNonRoot: boolPtr(true),
	}
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			Containers: []corev1.Container{
				containerWithSC("ok", "third-party/foo:1.0", sc),
			},
		}),
	}
	if err := CheckSecurityContextForNonBeclabImage(list); err != nil {
		t.Fatalf("non-root SC must pass: %v", err)
	}
}

func TestCheckSecurityContext_PodLevelRootFailsAllNonBeclabContainers(t *testing.T) {
	// Pod-level securityContext sets runAsUser=0. Every non-beclab
	// container (init and main) inherits and must be flagged once.
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{RunAsUser: int64Ptr(0)},
			InitContainers: []corev1.Container{
				{Name: "init-third", Image: "third-party/init:1.0"},
				{Name: "init-beclab", Image: "beclab/init:1.0"},
			},
			Containers: []corev1.Container{
				{Name: "main-third", Image: "third-party/main:1.0"},
				{Name: "main-beclab", Image: "beclab/main:1.0"},
			},
		}),
	}
	err := CheckSecurityContextForNonBeclabImage(list)
	if err == nil {
		t.Fatal("pod-level runAsUser=0 must trip every non-beclab container")
	}
	if !strings.Contains(err.Error(), "main-third") ||
		!strings.Contains(err.Error(), "init-third") {
		t.Fatalf("error should mention every non-beclab container, got: %v", err)
	}
	if strings.Contains(err.Error(), "main-beclab") ||
		strings.Contains(err.Error(), "init-beclab") {
		t.Fatalf("beclab containers must not be reported, got: %v", err)
	}
}

func TestCheckSecurityContext_PodLevelSafeFallsThroughToContainerCheck(t *testing.T) {
	// Pod-level securityContext is safe (runAsNonRoot=true). The
	// container check still runs, so an explicit container-level
	// runAsUser=0 must still surface.
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{RunAsNonRoot: boolPtr(true)},
			Containers: []corev1.Container{
				containerWithSC("bad", "third-party/foo:1.0",
					&corev1.SecurityContext{RunAsUser: int64Ptr(0)}),
			},
		}),
	}
	if err := CheckSecurityContextForNonBeclabImage(list); err == nil {
		t.Fatal("safe pod SC must not mask unsafe container SC")
	}
}

func TestCheckSecurityContext_DaemonSetCovered(t *testing.T) {
	sc := &corev1.SecurityContext{Privileged: boolPtr(true)}
	list := kube.ResourceList{
		daemonSetWithPodSpec("ds", corev1.PodSpec{
			Containers: []corev1.Container{
				containerWithSC("bad", "third-party/foo:1.0", sc),
			},
		}),
	}
	err := CheckSecurityContextForNonBeclabImage(list)
	if err == nil {
		t.Fatal("DaemonSet workloads must be covered by the check")
	}
	if !strings.Contains(err.Error(), "DaemonSet ds") {
		t.Fatalf("error should mention DaemonSet kind: %v", err)
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

