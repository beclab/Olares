package sidecar

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	// MacvlanBypassInitContainerName is the init container that keeps traffic on
	// the macvlan NIC out of any in-pod transparent proxy interception.
	MacvlanBypassInitContainerName = "macvlan-bypass-iptables"

	// macvlanBypassImage must ship an iptables binary; it is the same image the
	// envoy sidecar iptables init uses.
	macvlanBypassImage = "beclab/init:v1.2.3"

	// macvlanIfaceEnv carries the macvlan NIC name into the bypass script.
	macvlanIfaceEnv = "MACVLAN_IFACE"
)

// MacvlanBypassScript renders the bypass shell (unit-tested).
//
// Rules match on the interface only, so every protocol and port on the macvlan
// NIC is covered without tracking business port lists. Each rule is re-inserted
// at position 1 of a built-in chain, which short-circuits the chain before any
// jump a data-plane proxy installed (linkerd-init appends its jump to
// PREROUTING/OUTPUT, mesh-in inserts its own REDIRECT at the head of OUTPUT).
//
// The container must therefore run as the last init that touches iptables; it
// writes once and exits instead of polling to reclaim the chain head.
func MacvlanBypassScript() string {
	return `set -u
IFACE="${MACVLAN_IFACE:-net1}"
echo "macvlan-bypass: interface=$IFACE"
ip link show "$IFACE" >/dev/null 2>&1 \
  || echo "macvlan-bypass: $IFACE not attached yet; interface-matched rules apply once it appears" >&2

reinsert() {
  t="$1"
  c="$2"
  d="$3"
  while iptables -t "$t" -C "$c" "$d" "$IFACE" -j ACCEPT 2>/dev/null; do
    iptables -t "$t" -D "$c" "$d" "$IFACE" -j ACCEPT || {
      echo "macvlan-bypass: failed to drop stale rule -t $t $c $d $IFACE" >&2
      exit 1
    }
  done
  iptables -t "$t" -I "$c" 1 "$d" "$IFACE" -j ACCEPT || {
    echo "macvlan-bypass: failed to install -t $t $c $d $IFACE" >&2
    exit 1
  }
  echo "macvlan-bypass: installed -t $t -I $c 1 $d $IFACE -j ACCEPT"
}

reinsert nat PREROUTING -i
reinsert nat OUTPUT -o
reinsert filter INPUT -i
reinsert filter OUTPUT -o

echo "macvlan-bypass: nat/filter head after install:"
iptables -t nat -S PREROUTING 2>/dev/null || true
iptables -t nat -S OUTPUT 2>/dev/null || true
iptables -t filter -S INPUT 2>/dev/null || true
iptables -t filter -S OUTPUT 2>/dev/null || true
`
}

// GetMacvlanBypassInitContainer returns the macvlan bypass init container spec.
func GetMacvlanBypassInitContainer() corev1.Container {
	return corev1.Container{
		Name:            MacvlanBypassInitContainerName,
		Image:           macvlanBypassImage,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser:                pointer.Int64Ptr(0),
			RunAsNonRoot:             pointer.BoolPtr(false),
			AllowPrivilegeEscalation: pointer.BoolPtr(false),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
				Add:  []corev1.Capability{"NET_ADMIN", "NET_RAW"},
			},
		},
		Command: []string{"/bin/sh", "-c", MacvlanBypassScript()},
		Env: []corev1.EnvVar{
			{Name: macvlanIfaceEnv, Value: defaultMacvlanIface},
		},
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

// EnsureMacvlanBypassLast moves the bypass init container to the end of the
// init sequence, appending it when absent.
//
// Admission runs in several passes (macvlan webhook, sidecar webhook, then the
// Linkerd injector), and each pass may append its own iptables init. Re-placing
// the bypass on every pass is what keeps it behind mesh-in and linkerd-init
// without depending on webhook ordering.
func EnsureMacvlanBypassLast(pod *corev1.Pod) {
	if pod == nil {
		return
	}
	kept := make([]corev1.Container, 0, len(pod.Spec.InitContainers)+1)
	for _, c := range pod.Spec.InitContainers {
		if c.Name == MacvlanBypassInitContainerName {
			continue
		}
		kept = append(kept, c)
	}
	pod.Spec.InitContainers = append(kept, GetMacvlanBypassInitContainer())
}
