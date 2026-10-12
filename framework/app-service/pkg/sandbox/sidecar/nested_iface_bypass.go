package sidecar

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/beclab/Olares/framework/app-service/pkg/gateway/meshinagent"
)

const (
	NestedIfaceBypassInitContainerName = "nested-iface-bypass-iptables"
	AnnotMeshBypassInterfaces          = "gateway.olares.io/mesh-bypass-interfaces"

	nestedIfaceBypassImage = "beclab/init:v1.2.3"
	nestedIfaceBypassEnv   = "MESH_BYPASS_IFACES"
	meshInRedirectPortsEnv = "MESH_IN_REDIRECT_PORTS"
	defaultNestedIface     = "docker"
)

// meshInRedirectPorts lists the mesh-in listeners whose OUTPUT redirects are
// mirrored onto nested ifaces, so guest traffic to the gateway gets a caller JWT.
func meshInRedirectPorts() string {
	return fmt.Sprintf("%d,%d", meshinagent.HTTPListenPort, meshinagent.HTTPSListenPort)
}

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
	changed := nestedBypassEnv(pod) != want ||
		nestedBypassEnvValue(pod, meshInRedirectPortsEnv) != meshInRedirectPorts() ||
		!nestedBypassIsLast(pod)
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
# Guest packets are forwarded, so they miss the mesh-in OUTPUT redirects.
# Copy each gateway /32 REDIRECT onto the nested iface in PREROUTING, before
# that iface's ACCEPT (ACCEPT ends the nat chain). rules is "ip dport toport".
ports=",$(printf '%s' "${MESH_IN_REDIRECT_PORTS:-}" | tr -d ' '),"
rules=""
if [ "$ports" != ",," ]; then
  rules=$("$IPTABLES_BIN" -t nat -S OUTPUT 2>/dev/null | awk -v ports="$ports" '
    $1 == "-A" && $2 == "OUTPUT" {
      d = ""; dp = ""; tp = ""; j = ""
      for (i = 3; i < NF; i++) {
        if ($i == "-d") d = $(i + 1)
        else if ($i == "--dport") dp = $(i + 1)
        else if ($i == "-j") j = $(i + 1)
        else if ($i == "--to-ports") tp = $(i + 1)
      }
      if (j == "REDIRECT" && d ~ /^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\/32$/ && dp ~ /^[0-9]+$/ &&
          tp ~ /^[0-9]+$/ && index(ports, "," tp ",") > 0 && !seen[d " " dp " " tp]++)
        print d, dp, tp
    }')
fi
guest_spec() {
  printf -- '-d %s -i %s -p tcp -m tcp --dport %s -j REDIRECT --to-ports %s' "$1" "$2" "$3" "$4"
}
expect=""
redirect_guest() {
  [ -n "$rules" ] || return 0
  while read -r ip dp tp; do
    spec=$(guest_spec "$ip" "$1" "$dp" "$tp")
    expect="$expect-A PREROUTING $spec
"
    while "$IPTABLES_BIN" -t nat -C PREROUTING $spec 2>/dev/null; do
      "$IPTABLES_BIN" -t nat -D PREROUTING $spec || exit 1
    done
    "$IPTABLES_BIN" -t nat -I PREROUTING 1 $spec || exit 1
  done <<EOF
$rules
EOF
}
verify_guest() {
  [ -n "$rules" ] || return 0
  gdump=$("$IPTABLES_BIN" -t nat -S PREROUTING 2>/dev/null || true)
  while read -r ip dp tp; do
    spec=$(guest_spec "$ip" "$1" "$dp" "$tp")
    printf '%s\n' "$gdump" | awk -v want="-A PREROUTING $spec" -v acc="-A PREROUTING -i $1 -j ACCEPT" '
      $0 == want { ok = 1; exit }
      $0 == acc || /PROXY_INIT_REDIRECT/ { exit }
      END { exit !ok }
    ' || { echo "nested-iface-bypass: guest redirect not before ACCEPT/PROXY_INIT: $spec" >&2; exit 1; }
  done <<EOF
$rules
EOF
}
verify() {
  t=$1; c=$2; d=$3; iface=$4; last=$5
  dump=$("$IPTABLES_BIN" -t "$t" -S "$c" 2>/dev/null || true)
  echo "$dump" | grep -Fq -- "-A $c $d $iface -j ACCEPT" || {
    echo "nested-iface-bypass: missing ACCEPT $iface on $t/$c" >&2; exit 1
  }
  first=$(printf '%s\n' "$dump" | awk -v t="$t" -v c="$c" -v expect="$expect" '
    BEGIN { n = split(expect, a, "\n"); for (i = 1; i <= n; i++) if (a[i] != "") skip[a[i]] = 1 }
    /^-A / { if (t == "nat" && c == "PREROUTING" && ($0 in skip)) next; print; exit }
  ')
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
  redirect_guest "$iface"
  reinsert nat OUTPUT -o "$iface"
  reinsert filter INPUT -i "$iface"
  reinsert filter OUTPUT -o "$iface"
  last=$iface; n=$((n+1))
done
[ "$n" -gt 0 ] || { echo "nested-iface-bypass: empty MESH_BYPASS_IFACES" >&2; exit 1; }
for iface in $list; do
  [ -n "$iface" ] || continue
  verify_guest "$iface"
  verify nat PREROUTING -i "$iface" "$last"
  verify nat OUTPUT -o "$iface" "$last"
  verify filter INPUT -i "$iface" "$last"
  verify filter OUTPUT -o "$iface" "$last"
done
g=0
[ -n "$rules" ] && g=$(printf '%s\n' "$rules" | grep -c .)
echo "nested-iface-bypass: mirrored $g mesh-in redirect(s) onto guest ifaces"
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
		Command: []string{"/bin/sh", "-c", NestedIfaceBypassScript()},
		Env: []corev1.EnvVar{
			{Name: nestedIfaceBypassEnv, Value: list},
			{Name: meshInRedirectPortsEnv, Value: meshInRedirectPorts()},
		},
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
	return nestedBypassEnvValue(pod, nestedIfaceBypassEnv)
}

func nestedBypassEnvValue(pod *corev1.Pod, name string) string {
	for _, c := range pod.Spec.InitContainers {
		if c.Name != NestedIfaceBypassInitContainerName {
			continue
		}
		for _, e := range c.Env {
			if e.Name == name {
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
