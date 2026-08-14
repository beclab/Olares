package sidecar

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"
)

const (
	NestedIfaceBypassInitContainerName = "nested-iface-bypass-iptables"
	AnnotMeshBypassInterfaces          = "gateway.olares.io/mesh-bypass-interfaces"

	nestedIfaceBypassImage = "beclab/init:v1.2.3"
	nestedIfaceBypassEnv   = "MESH_BYPASS_IFACES"
	defaultNestedIface     = "docker"
)

// ResolveNestedIfaceBypass decides whether to inject the nested-iface bypass init.
// Annotation CSV wins; none|off|-|false disables; else auto on dockurr image or kvm∧tun.
func ResolveNestedIfaceBypass(pod *corev1.Pod) (ifaces []string, ok bool) {
	if pod == nil || !podHasLinkerdMesh(pod) {
		return nil, false
	}
	if pod.Annotations != nil {
		if raw, exists := pod.Annotations[AnnotMeshBypassInterfaces]; exists {
			v := strings.TrimSpace(raw)
			switch strings.ToLower(v) {
			case "", "none", "off", "-", "false", "disable", "disabled":
				return nil, false
			}
			if parsed := parseIfaceList(v); len(parsed) > 0 {
				return parsed, true
			}
			return nil, false
		}
	}
	if podHasDockurrImage(pod) || podHasKvmAndTun(pod) {
		return []string{defaultNestedIface}, true
	}
	return nil, false
}

// ApplyNestedIfaceBypass injects/refreshes the bypass init when Resolve says so.
// Reports whether the pod spec changed (for probe-only admission).
func ApplyNestedIfaceBypass(pod *corev1.Pod) bool {
	ifaces, ok := ResolveNestedIfaceBypass(pod)
	if !ok {
		return false
	}
	want := strings.Join(ifaces, ",")
	changed := nestedBypassEnv(pod) != want || !nestedBypassIsLast(pod)
	EnsureNestedIfaceBypassLast(pod, ifaces)
	return changed
}

func parseIfaceList(raw string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, p := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\t'
	}) {
		if p == "" || len(p) > 15 || !ifaceNameOK(p) {
			continue
		}
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func ifaceNameOK(name string) bool {
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			continue
		}
		return false
	}
	return name != ""
}

func podHasLinkerdMesh(pod *corev1.Pod) bool {
	if pod.Annotations != nil &&
		strings.EqualFold(strings.TrimSpace(pod.Annotations["linkerd.io/inject"]), "enabled") {
		return true
	}
	for _, c := range pod.Spec.Containers {
		if c.Name == "linkerd-proxy" {
			return true
		}
	}
	for _, c := range pod.Spec.InitContainers {
		if c.Name == "linkerd-init" {
			return true
		}
	}
	return false
}

func podHasDockurrImage(pod *corev1.Pod) bool {
	for _, c := range pod.Spec.Containers {
		if c.Name == "linkerd-proxy" || c.Name == "olares-mesh-in-agent" {
			continue
		}
		base := imageBase(c.Image)
		if base == "dockurr-windows" || strings.HasPrefix(base, "dockurr-windows-") ||
			base == "dockurr-macos" || strings.HasPrefix(base, "dockurr-macos-") {
			return true
		}
	}
	return false
}

func imageBase(image string) string {
	s := strings.ToLower(strings.TrimSpace(image))
	if i := strings.Index(s, "@"); i >= 0 {
		s = s[:i]
	}
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.Index(s, ":"); i >= 0 {
		s = s[:i]
	}
	return s
}

func podHasKvmAndTun(pod *corev1.Pod) bool {
	var kvm, tun bool
	for _, v := range pod.Spec.Volumes {
		if v.HostPath == nil {
			continue
		}
		switch strings.TrimSpace(v.HostPath.Path) {
		case "/dev/kvm":
			kvm = true
		case "/dev/net/tun":
			tun = true
		}
	}
	return kvm && tun
}

func NestedIfaceBypassScript() string {
	return `set -u
IFACES="${MESH_BYPASS_IFACES:-docker}"
IPTABLES_BIN="${IPTABLES_BIN:-iptables-nft}"
command -v "$IPTABLES_BIN" >/dev/null 2>&1 || { echo "nested-iface-bypass: missing $IPTABLES_BIN" >&2; exit 1; }
case "$("$IPTABLES_BIN" --version 2>/dev/null || true)" in *nf_tables*) ;; *)
  echo "nested-iface-bypass: need nf_tables" >&2; exit 1 ;;
esac
reinsert() {
  t=$1; c=$2; d=$3; iface=$4
  while "$IPTABLES_BIN" -t "$t" -C "$c" "$d" "$iface" -j ACCEPT 2>/dev/null; do
    "$IPTABLES_BIN" -t "$t" -D "$c" "$d" "$iface" -j ACCEPT || exit 1
  done
  "$IPTABLES_BIN" -t "$t" -I "$c" 1 "$d" "$iface" -j ACCEPT || exit 1
}
verify() {
  t=$1; c=$2; d=$3; iface=$4; last=$5
  dump=$("$IPTABLES_BIN" -t "$t" -S "$c" 2>/dev/null || true)
  echo "$dump" | grep -Fq -- "-A $c $d $iface -j ACCEPT" || {
    echo "nested-iface-bypass: missing ACCEPT $iface on $t/$c" >&2; exit 1
  }
  first=$(printf '%s\n' "$dump" | awk '/^-A /{print; exit}')
  want="-A $c $d $last -j ACCEPT"
  [ "$first" = "$want" ] || {
    echo "nested-iface-bypass: $t/$c head want $want got ${first:-empty}" >&2; exit 1
  }
  echo "$dump" | awk -v want="-A $c $d $iface -j ACCEPT" '
    /PROXY_INIT_REDIRECT|PROXY_INIT_OUTPUT/ { bad=1; exit }
    $0 == want { ok=1; exit }
    END { exit !(ok && !bad) }
  ' || {
    echo "nested-iface-bypass: $iface ACCEPT not before PROXY_INIT on $t/$c" >&2; exit 1
  }
}
list=$(printf '%s' "$IFACES" | tr ',;' ' ')
last=""; n=0
for iface in $list; do
  [ -n "$iface" ] || continue
  ip link show "$iface" >/dev/null 2>&1 || true
  reinsert nat PREROUTING -i "$iface"
  reinsert nat OUTPUT -o "$iface"
  reinsert filter INPUT -i "$iface"
  reinsert filter OUTPUT -o "$iface"
  last=$iface; n=$((n+1))
done
[ "$n" -gt 0 ] || { echo "nested-iface-bypass: empty MESH_BYPASS_IFACES" >&2; exit 1; }
for iface in $list; do
  [ -n "$iface" ] || continue
  verify nat PREROUTING -i "$iface" "$last"
  verify nat OUTPUT -o "$iface" "$last"
  verify filter INPUT -i "$iface" "$last"
  verify filter OUTPUT -o "$iface" "$last"
done
echo "nested-iface-bypass: verified ACCEPT heads for $n iface(s) on nft"
`
}

func GetNestedIfaceBypassInitContainer(ifaces []string) corev1.Container {
	list := strings.Join(ifaces, ",")
	if list == "" {
		list = defaultNestedIface
	}
	return corev1.Container{
		Name:            NestedIfaceBypassInitContainerName,
		Image:           nestedIfaceBypassImage,
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
		Command:                  []string{"/bin/sh", "-c", NestedIfaceBypassScript()},
		Env:                      []corev1.EnvVar{{Name: nestedIfaceBypassEnv, Value: list}},
		TerminationMessagePath:   "/dev/termination-log",
		TerminationMessagePolicy: corev1.TerminationMessageReadFile,
	}
}

func EnsureNestedIfaceBypassLast(pod *corev1.Pod, ifaces []string) {
	if pod == nil || len(ifaces) == 0 {
		return
	}
	kept := make([]corev1.Container, 0, len(pod.Spec.InitContainers)+1)
	for _, c := range pod.Spec.InitContainers {
		if c.Name != NestedIfaceBypassInitContainerName {
			kept = append(kept, c)
		}
	}
	pod.Spec.InitContainers = append(kept, GetNestedIfaceBypassInitContainer(ifaces))
}

func nestedBypassEnv(pod *corev1.Pod) string {
	for _, c := range pod.Spec.InitContainers {
		if c.Name != NestedIfaceBypassInitContainerName {
			continue
		}
		for _, e := range c.Env {
			if e.Name == nestedIfaceBypassEnv {
				return e.Value
			}
		}
	}
	return ""
}

func nestedBypassIsLast(pod *corev1.Pod) bool {
	n := len(pod.Spec.InitContainers)
	return n > 0 && pod.Spec.InitContainers[n-1].Name == NestedIfaceBypassInitContainerName
}
