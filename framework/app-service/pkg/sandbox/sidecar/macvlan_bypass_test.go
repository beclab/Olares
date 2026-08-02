package sidecar

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestMacvlanBypassScriptInstallsFourHeadAcceptRules(t *testing.T) {
	script := MacvlanBypassScript()

	for _, want := range []string{
		"reinsert nat PREROUTING -i",
		"reinsert nat OUTPUT -o",
		"reinsert filter INPUT -i",
		"reinsert filter OUTPUT -o",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("bypass script missing %q in:\n%s", want, script)
		}
	}
	if !strings.Contains(script, `iptables -t "$t" -I "$c" 1 "$d" "$IFACE" -j ACCEPT`) {
		t.Fatalf("bypass script must insert ACCEPT at chain head in:\n%s", script)
	}
	if !strings.Contains(script, `IFACE="${MACVLAN_IFACE:-net1}"`) {
		t.Fatalf("bypass script must read the iface from env with net1 default in:\n%s", script)
	}
}

func TestMacvlanBypassScriptStaysSingleShotAndNonDestructive(t *testing.T) {
	script := MacvlanBypassScript()

	for _, forbidden := range []string{
		"NOTRACK",
		"iptables -F",
		"iptables -t nat -F",
		"-P INPUT",
		"-P OUTPUT",
		"sleep",
		"while true",
	} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("bypass script must not contain %q (no reclaim loop, no flush, no policy change) in:\n%s", forbidden, script)
		}
	}
	if strings.Count(script, "exit 1") == 0 {
		t.Fatalf("bypass script must fail closed when a rule cannot be installed in:\n%s", script)
	}
}

func TestGetMacvlanBypassInitContainerSpec(t *testing.T) {
	c := GetMacvlanBypassInitContainer()

	if c.Name != MacvlanBypassInitContainerName {
		t.Fatalf("unexpected container name %q", c.Name)
	}
	if strings.HasSuffix(c.Image, ":latest") || !strings.Contains(c.Image, ":") {
		t.Fatalf("bypass image must be pinned to a tag other than latest, got %q", c.Image)
	}
	if c.SecurityContext == nil || c.SecurityContext.Capabilities == nil {
		t.Fatalf("bypass container needs a security context with capabilities")
	}
	hasNetAdmin := false
	for _, add := range c.SecurityContext.Capabilities.Add {
		if add == "NET_ADMIN" {
			hasNetAdmin = true
		}
	}
	if !hasNetAdmin {
		t.Fatalf("bypass container needs NET_ADMIN to write iptables, got %v", c.SecurityContext.Capabilities.Add)
	}
	if c.SecurityContext.RunAsUser == nil || *c.SecurityContext.RunAsUser != 0 {
		t.Fatalf("bypass container must run as root to write iptables")
	}
	iface := ""
	for _, e := range c.Env {
		if e.Name == macvlanIfaceEnv {
			iface = e.Value
		}
	}
	if iface != defaultMacvlanIface {
		t.Fatalf("expected %s=%s, got %q", macvlanIfaceEnv, defaultMacvlanIface, iface)
	}
}

func TestEnsureMacvlanBypassLast(t *testing.T) {
	tests := []struct {
		name  string
		inits []string
		want  []string
	}{
		{
			name:  "append to empty init list",
			inits: nil,
			want:  []string{MacvlanBypassInitContainerName},
		},
		{
			name:  "append after linkerd and mesh-in iptables writers",
			inits: []string{"linkerd-init", "macvlan-reply-via-eth0", "olares-mesh-in-agent-iptables"},
			want:  []string{"linkerd-init", "macvlan-reply-via-eth0", "olares-mesh-in-agent-iptables", MacvlanBypassInitContainerName},
		},
		{
			name:  "move an already injected bypass back to the tail",
			inits: []string{"linkerd-init", MacvlanBypassInitContainerName, "olares-mesh-in-agent-iptables"},
			want:  []string{"linkerd-init", "olares-mesh-in-agent-iptables", MacvlanBypassInitContainerName},
		},
		{
			name:  "collapse duplicates",
			inits: []string{MacvlanBypassInitContainerName, "linkerd-init", MacvlanBypassInitContainerName},
			want:  []string{"linkerd-init", MacvlanBypassInitContainerName},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &corev1.Pod{}
			for _, n := range tt.inits {
				pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{Name: n})
			}

			EnsureMacvlanBypassLast(pod)

			got := make([]string, 0, len(pod.Spec.InitContainers))
			for _, c := range pod.Spec.InitContainers {
				got = append(got, c.Name)
			}
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("init order = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureMacvlanBypassLastIsIdempotent(t *testing.T) {
	pod := &corev1.Pod{}
	pod.Spec.InitContainers = []corev1.Container{{Name: "linkerd-init"}}

	EnsureMacvlanBypassLast(pod)
	first := len(pod.Spec.InitContainers)
	EnsureMacvlanBypassLast(pod)

	if len(pod.Spec.InitContainers) != first {
		t.Fatalf("repeated calls must not grow the init list: %d -> %d", first, len(pod.Spec.InitContainers))
	}
	last := pod.Spec.InitContainers[len(pod.Spec.InitContainers)-1]
	if last.Name != MacvlanBypassInitContainerName {
		t.Fatalf("bypass must stay last, got %q", last.Name)
	}
}

func TestEnsureMacvlanBypassLastNilPod(t *testing.T) {
	EnsureMacvlanBypassLast(nil)
}
