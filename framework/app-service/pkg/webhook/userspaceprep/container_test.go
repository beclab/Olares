package userspaceprep

import (
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
)

func TestScript(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
		corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data", SubPath: "config"},
		}},
	)
	plan, ok := BuildPlan(pod, testRoots())
	if !ok {
		t.Fatal("expected a plan")
	}

	want := `set -e
[ "$(stat -c %u:%g "/olares-prepare/0")" = "1000:1000" ] || chown 1000:1000 "/olares-prepare/0"
mkdir -p "/olares-prepare/0/config"
chown 1000:1000 "/olares-prepare/0/config"
`
	if got := Script(plan); got != want {
		t.Fatalf("script mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// The whole design rests on never walking the data tree, so a recursive
// flag must never appear in the generated script.
func TestScript_NeverRecursive(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares/config")},
		corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data", SubPath: "state"},
		}},
	)
	plan, _ := BuildPlan(pod, testRoots())
	script := Script(plan)
	for _, forbidden := range []string{"chown -R", "chown -r", "chmod -R"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("script must never recurse, found %q in:\n%s", forbidden, script)
		}
	}
}

// Script addresses directories through the prepare container's own mount
// points. Leaking a host path would break the moment the two differ.
func TestScript_UsesMountPathsNotHostPaths(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares")},
		corev1.Container{Name: "main"},
	)
	plan, _ := BuildPlan(pod, testRoots())
	if strings.Contains(Script(plan), userspaceHost) {
		t.Fatalf("script leaked a host path:\n%s", Script(plan))
	}
}

func TestInitContainer(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares/config")},
		corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data", ReadOnly: true},
		}},
	)
	plan, _ := BuildPlan(pod, testRoots())
	c := InitContainer(plan)

	if c.Name != constants.UserspacePrepareInitContainerName {
		t.Errorf("name = %q, want %q", c.Name, constants.UserspacePrepareInitContainerName)
	}
	if c.Image != userspacePrepareImage {
		t.Errorf("image = %q, want %q", c.Image, userspacePrepareImage)
	}
	// The image ships with the platform, so pulling it on every pod
	// start would only add a registry round-trip to the critical path.
	if c.ImagePullPolicy != corev1.PullIfNotPresent {
		t.Errorf("imagePullPolicy = %q, want %q", c.ImagePullPolicy, corev1.PullIfNotPresent)
	}

	sc := c.SecurityContext
	if sc == nil {
		t.Fatal("prepare needs an explicit securityContext")
	}
	// Container-level runAsUser overrides the pod-level 1000 injected
	// alongside it, which is what lets prepare chown while the app
	// containers still drop to 1000. Running as uid 0 is also what
	// grants CAP_CHOWN, so no explicit capability set is needed.
	if sc.RunAsUser == nil || *sc.RunAsUser != 0 {
		t.Errorf("runAsUser = %v, want 0", sc.RunAsUser)
	}

	wantCmd := []string{"sh", "-c", Script(plan)}
	if !reflect.DeepEqual(c.Command, wantCmd) {
		t.Errorf("command = %q, want %q", c.Command, wantCmd)
	}

	// Every target must be reachable, or prepare would chown a path
	// that is not the one the app ends up using.
	if len(c.VolumeMounts) != len(plan.Targets) {
		t.Fatalf("got %d mounts for %d targets", len(c.VolumeMounts), len(plan.Targets))
	}
	for i, m := range c.VolumeMounts {
		target := plan.Targets[i]
		if m.Name != target.VolumeName {
			t.Errorf("mount[%d].name = %q, want %q", i, m.Name, target.VolumeName)
		}
		if m.MountPath != target.MountPath {
			t.Errorf("mount[%d].mountPath = %q, want %q", i, m.MountPath, target.MountPath)
		}
		// The app mounts this volume read-only; prepare must not.
		if m.ReadOnly {
			t.Errorf("mount %s must be writable for prepare to chown it", m.Name)
		}
	}
}

// The conventional root above an assembled leaf has no volume of its
// own in the chart, so BuildPlan synthesizes one. The container has to
// mount it under the same name or the pod would fail to schedule.
func TestInitContainer_MountsSynthesizedVolumes(t *testing.T) {
	pod := podWith(
		[]corev1.Volume{hostPathVolume("data", userspaceHost+"/Data/lares/config")},
		corev1.Container{Name: "main", VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}},
	)
	plan, _ := BuildPlan(pod, testRoots())
	if len(plan.ExtraVolumes) == 0 {
		t.Fatal("this fixture is meant to produce a synthesized volume")
	}

	declared := map[string]struct{}{}
	for _, v := range pod.Spec.Volumes {
		declared[v.Name] = struct{}{}
	}
	for _, v := range plan.ExtraVolumes {
		declared[v.Name] = struct{}{}
	}

	for _, m := range InitContainer(plan).VolumeMounts {
		if _, ok := declared[m.Name]; !ok {
			t.Errorf("mount %q refers to a volume no one declares", m.Name)
		}
	}
}
