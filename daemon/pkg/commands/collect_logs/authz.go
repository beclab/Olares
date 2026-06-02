package collectlogs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beclab/Olares/daemon/pkg/utils"
	"github.com/beclab/Olares/framework/app-service/pkg/security"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	roleOwner  = "owner"
	roleAdmin  = "admin"
	roleNormal = "normal"
)

// ErrNothingRequested is returned when an authorized request would collect
// nothing at all (caller asked for an empty scope).
var ErrNothingRequested = errors.New("collect-logs: request selects nothing to collect")

// ScopeDeniedError reports the parts of a request that the caller's role is
// not allowed to collect. The handler maps this to HTTP 403.
type ScopeDeniedError struct {
	Denied []string
}

func (e *ScopeDeniedError) Error() string {
	return fmt.Sprintf("requested scope exceeds caller permission, denied: %s", strings.Join(e.Denied, "; "))
}

// resolvedScope is the post-authz, post-wildcard-expansion collection plan
// that translateFlags turns into olares-cli logs flags.
type resolvedScope struct {
	username string
	role     string

	collectSystemd    bool
	systemdComponents []string // empty while collectSystemd => all known services
	since             string
	maxLines          int

	collectDmesg       bool
	collectNetwork     bool
	collectClusterInfo bool

	collectPods   bool
	allNamespaces bool     // true => no --pod-namespaces filter (every namespace)
	podNamespaces []string // concrete subset when !allNamespaces
}

// authorize resolves the caller's role, validates the requested scope against
// it, and expands wildcards into a concrete plan. Privileged roles
// (owner/admin) may collect anything; a normal user may only collect pod logs
// from namespaces they own.
func authorize(ctx context.Context, kubeClient kubernetes.Interface, dynamicClient dynamic.Interface, p *Param) (*resolvedScope, error) {
	username, role, err := resolveCaller(ctx, dynamicClient, p.CallerOlaresID)
	if err != nil {
		return nil, err
	}

	rs := &resolvedScope{username: username, role: role}
	privileged := role == roleOwner || role == roleAdmin
	var denied []string

	if len(p.Systemd.Components) > 0 {
		if !privileged {
			denied = append(denied, "systemd (requires owner/admin)")
		} else {
			rs.collectSystemd = true
			if !hasWildcard(p.Systemd.Components) {
				rs.systemdComponents = p.Systemd.Components
			}
			rs.since = p.Systemd.Since
			rs.maxLines = p.Systemd.MaxLines
		}
	}

	if p.Host.Dmesg {
		if !privileged {
			denied = append(denied, "host.dmesg (requires owner/admin)")
		} else {
			rs.collectDmesg = true
		}
	}
	if p.Host.Network {
		if !privileged {
			denied = append(denied, "host.network (requires owner/admin)")
		} else {
			rs.collectNetwork = true
		}
	}

	if p.Cluster.Info {
		if !privileged {
			denied = append(denied, "cluster.info (requires owner/admin)")
		} else {
			rs.collectClusterInfo = true
		}
	}

	if len(p.Namespaces.Names) > 0 {
		rs.collectPods = true
		if hasWildcard(p.Namespaces.Names) {
			if privileged {
				rs.allNamespaces = true
			} else {
				owned, err := listOwnedNamespaces(ctx, kubeClient, username)
				if err != nil {
					return nil, err
				}
				rs.podNamespaces = owned
			}
		} else if privileged {
			rs.podNamespaces = p.Namespaces.Names
		} else {
			owned, err := listOwnedNamespaces(ctx, kubeClient, username)
			if err != nil {
				return nil, err
			}
			ownedSet := make(map[string]struct{}, len(owned))
			for _, ns := range owned {
				ownedSet[ns] = struct{}{}
			}
			for _, ns := range p.Namespaces.Names {
				if _, ok := ownedSet[ns]; ok {
					rs.podNamespaces = append(rs.podNamespaces, ns)
				} else {
					denied = append(denied, fmt.Sprintf("namespace %q (not owned by caller)", ns))
				}
			}
		}
	}

	if len(denied) > 0 {
		return nil, &ScopeDeniedError{Denied: denied}
	}

	// Guard against privilege escalation via the CLI's "empty filter = all"
	// semantics: a normal user who owns nothing but asked for namespaces must
	// not fall through to collecting every namespace.
	if rs.collectPods && !rs.allNamespaces && len(rs.podNamespaces) == 0 {
		rs.collectPods = false
	}

	if !rs.collectSystemd && !rs.collectDmesg && !rs.collectNetwork && !rs.collectClusterInfo && !rs.collectPods {
		return nil, ErrNothingRequested
	}

	return rs, nil
}

// resolveCaller maps a verified Olares ID to its username and role.
func resolveCaller(ctx context.Context, dynamicClient dynamic.Interface, olaresID string) (username, role string, err error) {
	return utils.GetUserRoleByOlaresID(ctx, dynamicClient, olaresID)
}

func listOwnedNamespaces(ctx context.Context, kubeClient kubernetes.Interface, username string) ([]string, error) {
	nss, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", security.NamespaceOwnerLabel, username),
	})
	if err != nil {
		return nil, fmt.Errorf("list owned namespaces error: %w", err)
	}
	out := make([]string, 0, len(nss.Items))
	for _, ns := range nss.Items {
		out = append(out, ns.Name)
	}
	return out, nil
}
