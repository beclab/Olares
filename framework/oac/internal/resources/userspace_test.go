package resources

import (
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/kube"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cliresource "k8s.io/cli-runtime/pkg/resource"
)

// ---- fixtures for the workload kinds walkPodSpecs learned to decode ----

func jobWithPodSpec(name string, spec corev1.PodSpec) *cliresource.Info {
	job := &batchv1.Job{
		TypeMeta:   metav1.TypeMeta{Kind: KindJob, APIVersion: "batch/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
	}
	return &cliresource.Info{Name: name, Object: job}
}

func cronJobWithPodSpec(name string, spec corev1.PodSpec) *cliresource.Info {
	cj := &batchv1.CronJob{
		TypeMeta:   metav1.TypeMeta{Kind: KindCronJob, APIVersion: "batch/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
			},
		},
	}
	return &cliresource.Info{Name: name, Object: cj}
}

func podWithPodSpec(name string, spec corev1.PodSpec) *cliresource.Info {
	pod := &corev1.Pod{
		TypeMeta:   metav1.TypeMeta{Kind: KindPod, APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       spec,
	}
	return &cliresource.Info{Name: name, Object: pod}
}

// hostPathPodSpec builds a pod spec with one hostPath volume mounted into
// a single container running the given command.
func hostPathPodSpec(mountPath string, command ...string) corev1.PodSpec {
	return corev1.PodSpec{
		Volumes: []corev1.Volume{{
			Name:         "data",
			VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/olares/userdata/Data/app"}},
		}},
		Containers: []corev1.Container{{
			Name:         "main",
			Image:        "third-party/foo:1.0",
			Command:      command,
			VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: mountPath}},
		}},
	}
}

// ---- D1: walkPodSpecs coverage ----

func TestWalkPodSpecs_CoversJobCronJobPod(t *testing.T) {
	spec := corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: "foo:1"}}}
	list := kube.ResourceList{
		deploymentWithPodSpec("dep", spec),
		daemonSetWithPodSpec("ds", spec),
		jobWithPodSpec("job", spec),
		cronJobWithPodSpec("cron", spec),
		podWithPodSpec("bare", spec),
	}

	seen := map[string]string{}
	walkPodSpecs(list, func(kind, name string, _ corev1.PodSpec) {
		seen[kind] = name
	})

	for kind, want := range map[string]string{
		KindDeployment: "dep",
		KindDaemonSet:  "ds",
		KindJob:        "job",
		KindCronJob:    "cron",
		KindPod:        "bare",
	} {
		if seen[kind] != want {
			t.Errorf("walkPodSpecs missed %s: got name %q, want %q", kind, seen[kind], want)
		}
	}
}

// The widened walker means the pre-existing securityContext rule now sees
// batch workloads too. Pin that explicitly so the behaviour change is
// deliberate rather than incidental.
func TestCheckSecurityContext_JobRootContainer(t *testing.T) {
	list := kube.ResourceList{
		jobWithPodSpec("migrate", corev1.PodSpec{
			Containers: []corev1.Container{
				containerWithSC("run", "third-party/foo:1.0", &corev1.SecurityContext{RunAsUser: int64Ptr(0)}),
			},
		}),
	}
	err := CheckSecurityContextForNonBeclabImage(list)
	if err == nil {
		t.Fatal("runAsUser=0 on a non-beclab Job container must fail")
	}
	if !strings.Contains(err.Error(), "Job migrate") {
		t.Fatalf("error should pinpoint the Job: %v", err)
	}
}

func TestAllContainers_IncludesInitAndEphemeral(t *testing.T) {
	spec := corev1.PodSpec{
		InitContainers:      []corev1.Container{{Name: "init"}},
		Containers:          []corev1.Container{{Name: "main"}},
		EphemeralContainers: []corev1.EphemeralContainer{{EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug"}}},
	}
	var names []string
	for _, c := range allContainers(spec) {
		names = append(names, c.Name)
	}
	got := strings.Join(names, ",")
	if got != "init,main,debug" {
		t.Fatalf("allContainers order/contents = %q, want %q", got, "init,main,debug")
	}
}

// ---- D2: subPathExpr ----

func TestCheckSubPathExpr_StaticSubPathPasses(t *testing.T) {
	list := kube.ResourceList{
		deploymentWithPodSpec("d", corev1.PodSpec{
			Containers: []corev1.Container{{
				Name:         "main",
				VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data", SubPath: "config"}},
			}},
		}),
	}
	if err := CheckSubPathExpr(list); err != nil {
		t.Fatalf("a static subPath must pass: %v", err)
	}
}

func TestCheckSubPathExpr_RejectedEverywhere(t *testing.T) {
	mount := []corev1.VolumeMount{{Name: "data", MountPath: "/data", SubPathExpr: "$(POD_NAME)"}}

	cases := []struct {
		name       string
		info       *cliresource.Info
		wantWorkld string
		wantCtr    string
	}{
		{
			name: "deployment main container",
			info: deploymentWithPodSpec("web", corev1.PodSpec{
				Containers: []corev1.Container{{Name: "main", VolumeMounts: mount}},
			}),
			wantWorkld: "Deployment web",
			wantCtr:    "container main",
		},
		{
			name: "init container",
			info: deploymentWithPodSpec("web", corev1.PodSpec{
				InitContainers: []corev1.Container{{Name: "setup", VolumeMounts: mount}},
			}),
			wantWorkld: "Deployment web",
			wantCtr:    "container setup",
		},
		{
			name: "job",
			info: jobWithPodSpec("migrate", corev1.PodSpec{
				Containers: []corev1.Container{{Name: "run", VolumeMounts: mount}},
			}),
			wantWorkld: "Job migrate",
			wantCtr:    "container run",
		},
		{
			name: "ephemeral container",
			info: deploymentWithPodSpec("web", corev1.PodSpec{
				EphemeralContainers: []corev1.EphemeralContainer{{
					EphemeralContainerCommon: corev1.EphemeralContainerCommon{Name: "debug", VolumeMounts: mount},
				}},
			}),
			wantWorkld: "Deployment web",
			wantCtr:    "container debug",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckSubPathExpr(kube.ResourceList{tc.info})
			if err == nil {
				t.Fatal("subPathExpr must be rejected")
			}
			for _, want := range []string{tc.wantWorkld, tc.wantCtr, "mount data", "$(POD_NAME)"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error missing %q: %v", want, err)
				}
			}
		})
	}
}

// ---- D3: blind recursive chown ----

func TestCheckBlindRecursiveChown(t *testing.T) {
	cases := []struct {
		name    string
		spec    corev1.PodSpec
		wantErr bool
	}{
		{
			name:    "recursive chown on the hostPath mount",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -R 1000:1000 /data && exec app"),
			wantErr: true,
		},
		{
			name:    "recursive chown below the hostPath mount",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -R 1000:1000 /data/config"),
			wantErr: true,
		},
		{
			name:    "bundled flags",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -Rf 1000:1000 /data"),
			wantErr: true,
		},
		{
			name:    "recursive flag after another flag",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -h -R 1000:1000 /data"),
			wantErr: true,
		},
		{
			name:    "tolerated failure is still recursive",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -R 1000:1000 /data || true"),
			wantErr: true,
		},
		{
			name:    "non-recursive chown is what the platform itself does",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown 1000:1000 /data"),
			wantErr: false,
		},
		{
			name:    "recursive chown outside the hostPath mount",
			spec:    hostPathPodSpec("/data", "sh", "-c", "chown -R 1000:1000 /opt/app"),
			wantErr: false,
		},
		{
			name: "no hostPath volume at all",
			spec: corev1.PodSpec{
				Volumes: []corev1.Volume{{
					Name:         "scratch",
					VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
				}},
				Containers: []corev1.Container{{
					Name:         "main",
					Command:      []string{"sh", "-c", "chown -R 1000:1000 /scratch"},
					VolumeMounts: []corev1.VolumeMount{{Name: "scratch", MountPath: "/scratch"}},
				}},
			},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckBlindRecursiveChown(kube.ResourceList{deploymentWithPodSpec("d", tc.spec)})
			if tc.wantErr && err == nil {
				t.Fatal("expected a blind chown -R violation")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected violation: %v", err)
			}
		})
	}
}

func TestCheckBlindRecursiveChown_ReportsMountAndContainer(t *testing.T) {
	spec := hostPathPodSpec("/data", "sh", "-c", "chown -R 1000:1000 /data")
	spec.InitContainers = []corev1.Container{{
		Name:         "fix-permissions",
		Command:      []string{"sh", "-c", "chown -R 1000:1000 /data"},
		VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}},
	}}

	err := CheckBlindRecursiveChown(kube.ResourceList{deploymentWithPodSpec("lares", spec)})
	if err == nil {
		t.Fatal("expected violations")
	}
	for _, want := range []string{"Deployment lares", "container fix-permissions", "container main", "/data"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
}
