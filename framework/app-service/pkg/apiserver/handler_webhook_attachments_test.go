package apiserver

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	appv1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func attachmentWebhookHandler(t *testing.T) *Handler {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	bfl := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{
		Name: "bfl", Namespace: "user-space-alice", Annotations: map[string]string{constants.UserSpaceDirKey: "/olares/userspace/alice"},
	}}
	return &Handler{ctrlClient: fake.NewClientBuilder().WithScheme(scheme).WithObjects(bfl).Build()}
}

func TestInjectAttachmentsTargetsDeclaredContainer(t *testing.T) {
	handler := attachmentWebhookHandler(t)
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: "sidecar"}}}}
	cfg := &appcfg.ApplicationConfig{AppName: "demo", OwnerName: "alice", WorkloadOptions: appcfg.WorkloadOptions{
		"server": {Allow: []appcfg.WorkloadCapability{
			{Type: "folder", Containers: []string{"app"}},
			{Type: "device.video", Containers: []string{"app"}},
		}},
	}}
	application := &appv1alpha1.Application{Spec: appv1alpha1.ApplicationSpec{Attachments: []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "drive/Home/photos", Name: "photos"},
		{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-v-1", Node: "node-a"},
	}}}
	if err := handler.injectAttachments(context.Background(), tpl, "server", cfg, application); err != nil {
		t.Fatal(err)
	}
	if got := tpl.Spec.NodeSelector[corev1.LabelHostname]; got != "node-a" {
		t.Fatalf("hostname selector=%q", got)
	}
	if len(tpl.Spec.Containers[0].VolumeMounts) != 1 || len(tpl.Spec.Containers[1].VolumeMounts) != 0 {
		t.Fatalf("folder leaked to non-target container: %#v", tpl.Spec.Containers)
	}
	deviceLimit := tpl.Spec.Containers[0].Resources.Limits["devices.bytetrade.io/video-v-1"]
	if deviceLimit.IsZero() {
		t.Fatal("device resource was not injected")
	}
	if len(tpl.Spec.SecurityContext.SupplementalGroups) != 1 || tpl.Spec.SecurityContext.SupplementalGroups[0] != deviceSupplementalGroups["device.video"] {
		t.Fatalf("supplemental groups=%v", tpl.Spec.SecurityContext.SupplementalGroups)
	}
	cleanupManagedAttachments(tpl, true)
	if err := handler.injectAttachments(context.Background(), tpl, "server", cfg, application); err != nil {
		t.Fatal(err)
	}
	if len(tpl.Spec.Volumes) != 1 || len(tpl.Spec.Containers[0].VolumeMounts) != 1 || len(tpl.Spec.SecurityContext.SupplementalGroups) != 1 {
		t.Fatalf("repeated injection was not idempotent: %#v", tpl.Spec)
	}
}

func TestInjectExternalFolderPinsNodeWithoutLocalPathCheck(t *testing.T) {
	t.Setenv("SHARED_LIB_PATH", "/olares/share")
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}}}}
	cfg := &appcfg.ApplicationConfig{WorkloadOptions: appcfg.WorkloadOptions{
		"server": {Allow: []appcfg.WorkloadCapability{{Type: "folder", Containers: []string{"app"}}}},
	}}
	application := &appv1alpha1.Application{Spec: appv1alpha1.ApplicationSpec{Attachments: []appv1alpha1.Attachment{
		{Workload: "server", Type: "folder", Ref: "external/node-a/usb/photos", Name: "photos", Node: "node-a"},
	}}}

	if err := (&Handler{}).injectAttachments(context.Background(), tpl, "server", cfg, application); err != nil {
		t.Fatal(err)
	}
	if got := tpl.Spec.NodeSelector[corev1.LabelHostname]; got != "node-a" {
		t.Fatalf("hostname selector=%q", got)
	}
	if len(tpl.Spec.Volumes) != 1 || tpl.Spec.Volumes[0].HostPath == nil || tpl.Spec.Volumes[0].HostPath.Path != "/olares/share/usb/photos" {
		t.Fatalf("external folder hostPath=%#v", tpl.Spec.Volumes)
	}
}

func TestInjectAttachmentsUsesDeviceClassSupplementalGroups(t *testing.T) {
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{
		Containers:      []corev1.Container{{Name: "app"}},
		SecurityContext: &corev1.PodSecurityContext{SupplementalGroups: []int64{1234}},
	}}
	cfg := &appcfg.ApplicationConfig{WorkloadOptions: appcfg.WorkloadOptions{
		"server": {Allow: []appcfg.WorkloadCapability{
			{Type: "device.serial", Containers: []string{"app"}},
			{Type: "device.audio", Containers: []string{"app"}},
			{Type: "device.video", Containers: []string{"app"}},
			{Type: "device.hid", Containers: []string{"app"}},
		}},
	}}
	application := &appv1alpha1.Application{Spec: appv1alpha1.ApplicationSpec{Attachments: []appv1alpha1.Attachment{
		{Workload: "server", Type: "device.serial", Ref: "devices.bytetrade.io/serial-s-1", Node: "node-a"},
		{Workload: "server", Type: "device.audio", Ref: "devices.bytetrade.io/audio-s-1", Node: "node-a"},
		{Workload: "server", Type: "device.video", Ref: "devices.bytetrade.io/video-s-1", Node: "node-a"},
		{Workload: "server", Type: "device.hid", Ref: "devices.bytetrade.io/hid-s-1", Node: "node-a"},
	}}}

	cleanupManagedAttachments(tpl, true)
	if err := (&Handler{}).injectAttachments(context.Background(), tpl, "server", cfg, application); err != nil {
		t.Fatal(err)
	}
	want := []int64{1234, 20, 29, 44}
	if len(tpl.Spec.SecurityContext.SupplementalGroups) != len(want) {
		t.Fatalf("supplemental groups=%v, want %v", tpl.Spec.SecurityContext.SupplementalGroups, want)
	}
	for i := range want {
		if tpl.Spec.SecurityContext.SupplementalGroups[i] != want[i] {
			t.Fatalf("supplemental groups=%v, want %v", tpl.Spec.SecurityContext.SupplementalGroups, want)
		}
	}
	if got := tpl.Annotations[attachmentGroupsAnnotation]; got != "20,29,44" {
		t.Fatalf("injected groups annotation=%q", got)
	}
	cleanupManagedAttachments(tpl, true)
	if len(tpl.Spec.SecurityContext.SupplementalGroups) != 1 || tpl.Spec.SecurityContext.SupplementalGroups[0] != 1234 {
		t.Fatalf("cleanup removed an unmanaged group or retained managed groups: %v", tpl.Spec.SecurityContext.SupplementalGroups)
	}
}

func TestInjectBluetoothHidesHostBusFromApplication(t *testing.T) {
	t.Setenv("BLUEZ_DBUS_PROXY_IMAGE", "example/bluez-proxy:test")
	tpl := &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app"}, {Name: "other"}}}}
	cfg := &appcfg.ApplicationConfig{WorkloadOptions: appcfg.WorkloadOptions{
		"server": {Allow: []appcfg.WorkloadCapability{{Type: "bluetooth", Containers: []string{"app"}}}},
	}}
	application := &appv1alpha1.Application{Spec: appv1alpha1.ApplicationSpec{Settings: map[string]string{"bluetooth": "true"}}}
	if err := (&Handler{}).injectAttachments(context.Background(), tpl, "server", cfg, application); err != nil {
		t.Fatal(err)
	}
	if len(tpl.Spec.Containers) != 3 || tpl.Spec.Containers[2].Name != bluezProxyContainerName {
		t.Fatalf("bluez proxy not injected: %#v", tpl.Spec.Containers)
	}
	if len(tpl.Spec.InitContainers) != 0 {
		t.Fatalf("bluetooth injection added unexpected init containers: %#v", tpl.Spec.InitContainers)
	}
	if len(tpl.Spec.Containers[0].VolumeMounts) != 1 || tpl.Spec.Containers[0].VolumeMounts[0].Name != bluezVolumeName {
		t.Fatal("target container did not receive filtered socket")
	}
	if len(tpl.Spec.Containers[1].VolumeMounts) != 0 {
		t.Fatal("bluetooth socket leaked to non-target container")
	}
	proxy := tpl.Spec.Containers[2]
	if proxy.SecurityContext == nil || proxy.SecurityContext.RunAsUser == nil || *proxy.SecurityContext.RunAsUser != 0 {
		t.Fatalf("bluez proxy must run as uid 0 for system D-Bus EXTERNAL authentication: %#v", proxy.SecurityContext)
	}
	if proxy.SecurityContext.RunAsNonRoot == nil || *proxy.SecurityContext.RunAsNonRoot {
		t.Fatalf("bluez proxy must allow uid 0: %#v", proxy.SecurityContext)
	}
	if proxy.SecurityContext.AppArmorProfile == nil || proxy.SecurityContext.AppArmorProfile.Type != corev1.AppArmorProfileTypeUnconfined {
		t.Fatalf("bluez proxy must be unconfined so the host system bus accepts D-Bus calls: %#v", proxy.SecurityContext)
	}
	if proxy.StartupProbe == nil || proxy.ReadinessProbe == nil {
		t.Fatal("bluez proxy must have startup and readiness probes")
	}
	if len(proxy.VolumeMounts) != 2 || proxy.VolumeMounts[0].Name != attachmentVolumePrefix+"bluez-host" {
		t.Fatal("raw system bus must only be mounted into the bluez proxy")
	}
}

func TestRequiredAffinityAllowsHostname(t *testing.T) {
	affinity := &corev1.Affinity{NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
		{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: corev1.LabelHostname, Operator: corev1.NodeSelectorOpIn, Values: []string{"node-a"}}}},
	}}}}
	if !requiredAffinityAllowsHostname(affinity, "node-a") || requiredAffinityAllowsHostname(affinity, "node-b") {
		t.Fatal("hostname affinity compatibility was evaluated incorrectly")
	}
}
