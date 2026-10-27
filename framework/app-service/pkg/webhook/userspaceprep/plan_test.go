package userspaceprep

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

const (
	userspaceHost = "/olares/userdata"
	appCacheHost  = "/olares/appcache"
)

// testRoots mirrors what resolveUserspaceRoots derives for an app named
// "lares" that was granted all four permissions.
func testRoots() Roots {
	return Roots{
		AppData:   userspaceHost + "/Data/lares",
		AppCache:  appCacheHost + "/lares",
		UserData:  userspaceHost + "/Home",
		AppCommon: "/olares/rootfs/Common",
	}
}

func hostPathVolume(name, path string) corev1.Volume {
	return corev1.Volume{
		Name:         name,
		VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: path}},
	}
}

func podWith(volumes []corev1.Volume, containers ...corev1.Container) *corev1.Pod {
	return &corev1.Pod{Spec: corev1.PodSpec{Volumes: volumes, Containers: containers}}
}

// targetPaths flattens a plan to "hostPath[subPath,subPath]" strings so
// assertions stay readable.
func targetPaths(p Plan) []string {
	out := make([]string, 0, len(p.Targets))
	for _, t := range p.Targets {
		s := t.HostPath
		if len(t.SubPaths) > 0 {
			s += "["
			for i, sp := range t.SubPaths {
				if i > 0 {
					s += ","
				}
				s += sp
			}
			s += "]"
		}
		out = append(out, s)
	}
	return out
}

func TestBuildPlan(t *testing.T) {
	cases := []struct {
		name             string
		pod              *corev1.Pod
		wantOK           bool
		wantTargets      []string
		wantExtraVolumes []string
	}{
		{
			name: "assembled leaf under appData also owns the conventional root",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares/config")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/config"}}},
			),
			wantOK: true,
			wantTargets: []string{
				userspaceHost + "/Data/lares",
				userspaceHost + "/Data/lares/config",
			},
			wantExtraVolumes: []string{userspaceHost + "/Data/lares"},
		},
		{
			name: "mounting the conventional root directly needs no extra volume",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}},
			),
			wantOK:      true,
			wantTargets: []string{userspaceHost + "/Data/lares"},
		},
		{
			// Only the conventional root and the leaf are owned. The
			// directories in between are created 0755 by kubelet, so uid
			// 1000 can traverse them; chaining chown down the path would
			// only widen the blast radius.
			name: "intermediate segments are left alone",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("cache", appCacheHost+"/lares/models/c0")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "cache", MountPath: "/cache"}}},
			),
			wantOK: true,
			wantTargets: []string{
				appCacheHost + "/lares",
				appCacheHost + "/lares/models/c0",
			},
			wantExtraVolumes: []string{appCacheHost + "/lares"},
		},
		{
			name: "static subPaths are collected, deduped and sorted",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
					{Name: "data", MountPath: "/b", SubPath: "state"},
					{Name: "data", MountPath: "/a", SubPath: "config"},
					{Name: "data", MountPath: "/c", SubPath: "config"},
				}},
			),
			wantOK:      true,
			wantTargets: []string{userspaceHost + "/Data/lares[config,state]"},
		},
		{
			name: "read-only mounts still need the directory to exist and be owned",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
					{Name: "data", MountPath: "/data", SubPath: "config", ReadOnly: true},
				}},
			),
			wantOK:      true,
			wantTargets: []string{userspaceHost + "/Data/lares[config]"},
		},
		{
			// The webhook cannot expand subPathExpr, so it must not
			// pretend to prepare the directory it would resolve to.
			name: "subPathExpr is ignored",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
					{Name: "data", MountPath: "/data", SubPathExpr: "$(POD_NAME)"},
				}},
			),
			wantOK:      true,
			wantTargets: []string{userspaceHost + "/Data/lares"},
		},
		{
			name: "non-userspace hostPaths are out of scope",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("sock", "/var/run/docker.sock")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "sock", MountPath: "/sock"}}},
			),
			wantOK: false,
		},
		{
			name:   "no volumes at all",
			pod:    podWith(nil, corev1.Container{Name: "main"}),
			wantOK: false,
		},
		{
			name: "Home is a conventional root like any other",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("home", userspaceHost+"/Home")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "home", MountPath: "/home"}}},
			),
			wantOK:      true,
			wantTargets: []string{userspaceHost + "/Home"},
		},
		{
			name: "appCommon is cluster-wide but behaves the same",
			pod: podWith(
				[]corev1.Volume{hostPathVolume("common", "/olares/rootfs/Common/models")},
				corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "common", MountPath: "/models"}}},
			),
			wantOK: true,
			wantTargets: []string{
				"/olares/rootfs/Common",
				"/olares/rootfs/Common/models",
			},
			wantExtraVolumes: []string{"/olares/rootfs/Common"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, ok := BuildPlan(tc.pod, testRoots())
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got := targetPaths(plan); !reflect.DeepEqual(got, tc.wantTargets) {
				t.Errorf("targets = %v, want %v", got, tc.wantTargets)
			}
			var extras []string
			for _, v := range plan.ExtraVolumes {
				extras = append(extras, v.HostPath.Path)
			}
			if len(extras) != len(tc.wantExtraVolumes) {
				t.Fatalf("extra volumes = %v, want %v", extras, tc.wantExtraVolumes)
			}
			for i := range extras {
				if extras[i] != tc.wantExtraVolumes[i] {
					t.Errorf("extra volume[%d] = %q, want %q", i, extras[i], tc.wantExtraVolumes[i])
				}
			}
			for _, v := range plan.ExtraVolumes {
				if v.HostPath.Type == nil || *v.HostPath.Type != corev1.HostPathDirectoryOrCreate {
					t.Errorf("synthesized volume %s must be DirectoryOrCreate", v.Name)
				}
			}
		})
	}
}

func TestBuildPlan_NoGrantedRoots(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
		corev1.Container{Name: "main"},
	)
	if _, ok := BuildPlan(pod, Roots{}); ok {
		t.Fatal("an app with no userspace permission must produce no plan")
	}
}

func TestBuildPlan_InitContainerSubPathsCount(t *testing.T) {
	pod := podWith([]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")})
	pod.Spec.InitContainers = []corev1.Container{{
		Name:         "fix-permissions",
		VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data", SubPath: "seed"}},
	}}
	plan, ok := BuildPlan(pod, testRoots())
	if !ok {
		t.Fatal("expected a plan")
	}
	want := []string{userspaceHost + "/Data/lares[seed]"}
	if got := targetPaths(plan); !reflect.DeepEqual(got, want) {
		t.Fatalf("targets = %v, want %v", got, want)
	}
}

// The generated command ends up inside the pod spec, so it has to be a
// pure function of the spec's content. If volume or mount ordering
// leaked into it, every re-admission of the same pod would produce a
// different patch.
func TestBuildPlan_Deterministic(t *testing.T) {
	volumes := []corev1.Volume{
		hostPathVolume("cache", appCacheHost+"/lares"),
		hostPathVolume("data", userspaceHost+"/Data/lares/config"),
		hostPathVolume("home", userspaceHost+"/Home"),
	}
	mounts := []corev1.VolumeMount{
		{Name: "data", MountPath: "/z", SubPath: "z"},
		{Name: "cache", MountPath: "/a", SubPath: "a"},
		{Name: "data", MountPath: "/m", SubPath: "m"},
	}

	planA, _ := BuildPlan(podWith(volumes, corev1.Container{Name: "main", VolumeMounts: mounts}), testRoots())

	reversedVolumes := []corev1.Volume{volumes[2], volumes[1], volumes[0]}
	reversedMounts := []corev1.VolumeMount{mounts[2], mounts[1], mounts[0]}
	planB, _ := BuildPlan(podWith(reversedVolumes, corev1.Container{Name: "main", VolumeMounts: reversedMounts}), testRoots())

	if Script(planA) != Script(planB) {
		t.Fatalf("script is order-dependent:\n--- A ---\n%s\n--- B ---\n%s", Script(planA), Script(planB))
	}
}
