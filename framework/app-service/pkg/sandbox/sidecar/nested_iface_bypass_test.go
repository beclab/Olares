package sidecar

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func linkerdPod(ann map[string]string, spec corev1.PodSpec) *corev1.Pod {
	if ann == nil {
		ann = map[string]string{}
	}
	ann["linkerd.io/inject"] = "enabled"
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: ann}, Spec: spec}
}

func hostPathVol(name, path string) corev1.Volume {
	return corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{
		HostPath: &corev1.HostPathVolumeSource{Path: path},
	}}
}

func TestResolveNestedIfaceBypass(t *testing.T) {
	cases := []struct {
		name    string
		pod     *corev1.Pod
		wantOK  bool
		wantIF  []string
	}{
		{
			name: "annot CSV",
			pod: linkerdPod(map[string]string{AnnotMeshBypassInterfaces: "docker, br0"}, corev1.PodSpec{}),
			wantOK: true, wantIF: []string{"docker", "br0"},
		},
		{
			name: "annot none disables",
			pod: linkerdPod(map[string]string{AnnotMeshBypassInterfaces: "none"}, corev1.PodSpec{
				Volumes:    []corev1.Volume{hostPathVol("kvm", "/dev/kvm"), hostPathVol("tun", "/dev/net/tun")},
				Containers: []corev1.Container{{Name: "windows", Image: "beclab/dockurr-windows:igpu1"}},
			}),
		},
		{
			name: "dockurr-macos primary",
			pod: linkerdPod(nil, corev1.PodSpec{
				Containers: []corev1.Container{{Name: "windows", Image: "beclab/dockurr-macos:5.14"}},
			}),
			wantOK: true, wantIF: []string{"docker"},
		},
		{
			name: "dockurr-windows primary",
			pod: linkerdPod(nil, corev1.PodSpec{
				Containers: []corev1.Container{{Name: "windows", Image: "ghcr.io/beclab/dockurr-windows:igpu1"}},
			}),
			wantOK: true, wantIF: []string{defaultNestedIface},
		},
		{
			name: "kvm and tun secondary",
			pod: linkerdPod(nil, corev1.PodSpec{
				Volumes:    []corev1.Volume{hostPathVol("kvm", "/dev/kvm"), hostPathVol("tun", "/dev/net/tun")},
				Containers: []corev1.Container{{Name: "windows", Image: "myorg/custom-win-vm:1"}},
			}),
			wantOK: true, wantIF: []string{defaultNestedIface},
		},
		{
			name: "kvm alone",
			pod: linkerdPod(nil, corev1.PodSpec{
				Volumes:    []corev1.Volume{hostPathVol("kvm", "/dev/kvm")},
				Containers: []corev1.Container{{Name: "windows", Image: "busybox"}},
			}),
		},
		{
			name: "tun alone",
			pod: linkerdPod(nil, corev1.PodSpec{
				Volumes:    []corev1.Volume{hostPathVol("tun", "/dev/net/tun")},
				Containers: []corev1.Container{{Name: "vpn", Image: "busybox"}},
			}),
		},
		{
			name: "dri vfio",
			pod: linkerdPod(nil, corev1.PodSpec{
				Volumes:    []corev1.Volume{hostPathVol("dri", "/dev/dri"), hostPathVol("vfio", "/dev/vfio")},
				Containers: []corev1.Container{{Name: "app", Image: "busybox"}},
			}),
		},
		{
			name: "sidecar image ignored",
			pod: linkerdPod(nil, corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "linkerd-proxy", Image: "beclab/dockurr-windows:fake"},
					{Name: "app", Image: "nginx:1.25"},
				},
			}),
		},
		{
			name: "no linkerd",
			pod: &corev1.Pod{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "windows", Image: "beclab/dockurr-windows:igpu1"}},
			}},
		},
		{
			name: "ordinary app",
			pod: linkerdPod(nil, corev1.PodSpec{
				Containers: []corev1.Container{{Name: "app", Image: "nginx:1.25"}},
			}),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ResolveNestedIfaceBypass(tc.pod)
			if ok != tc.wantOK {
				t.Fatalf("ok=%v want %v ifaces=%v", ok, tc.wantOK, got)
			}
			if !tc.wantOK {
				return
			}
			if len(got) != len(tc.wantIF) {
				t.Fatalf("ifaces=%v want %v", got, tc.wantIF)
			}
			for i := range tc.wantIF {
				if got[i] != tc.wantIF[i] {
					t.Fatalf("ifaces=%v want %v", got, tc.wantIF)
				}
			}
		})
	}
}

func TestImageBase(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"beclab/dockurr-windows:igpu1", "dockurr-windows"},
		{"ghcr.io/beclab/dockurr-macos:5.14", "dockurr-macos"},
		{"dockurr-windows-custom@sha256:abc", "dockurr-windows-custom"},
		{"docker.io/library/nginx:1.25", "nginx"},
		{"", ""},
	} {
		if got := imageBase(tc.in); got != tc.want {
			t.Fatalf("base(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestEnsureNestedIfaceBypassLast(t *testing.T) {
	pod := &corev1.Pod{Spec: corev1.PodSpec{InitContainers: []corev1.Container{
		{Name: "linkerd-init"},
		{Name: "olares-mesh-in-agent-iptables"},
		{Name: NestedIfaceBypassInitContainerName},
	}}}
	EnsureNestedIfaceBypassLast(pod, []string{"docker"})
	if n := len(pod.Spec.InitContainers); n != 3 ||
		pod.Spec.InitContainers[2].Name != NestedIfaceBypassInitContainerName ||
		nestedBypassEnv(pod) != "docker" {
		t.Fatalf("inits=%v env=%q", pod.Spec.InitContainers, nestedBypassEnv(pod))
	}
	EnsureNestedIfaceBypassLast(nil, []string{"docker"})
	EnsureNestedIfaceBypassLast(&corev1.Pod{}, nil)
}

func TestApplyNestedIfaceBypass(t *testing.T) {
	pod := linkerdPod(nil, corev1.PodSpec{
		Containers: []corev1.Container{{Name: "windows", Image: "beclab/dockurr-windows:1"}},
	})
	if !ApplyNestedIfaceBypass(pod) {
		t.Fatal("first apply must report change")
	}
	if ApplyNestedIfaceBypass(pod) {
		t.Fatal("idempotent apply must report no change")
	}
	pod.Annotations[AnnotMeshBypassInterfaces] = "docker,br0"
	if !ApplyNestedIfaceBypass(pod) || nestedBypassEnv(pod) != "docker,br0" {
		t.Fatalf("env update: %q", nestedBypassEnv(pod))
	}
}

func TestGetNestedIfaceBypassInitContainerSpec(t *testing.T) {
	c := GetNestedIfaceBypassInitContainer([]string{"docker"})
	if c.Name != NestedIfaceBypassInitContainerName || c.Image != nestedIfaceBypassImage {
		t.Fatalf("name=%s image=%s", c.Name, c.Image)
	}
	if c.SecurityContext == nil || c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 0 {
		t.Fatal("must run as root")
	}
	if !strings.Contains(NestedIfaceBypassScript(), "iptables-nft") {
		t.Fatal("script must use iptables-nft")
	}
}
