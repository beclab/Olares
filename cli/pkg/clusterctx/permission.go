package clusterctx

import "strings"

// GlobalRoleAdmin is the KubeSphere global-role wire constant for the
// cluster/platform administrator (the "main account"). It mirrors the SPA
// constant GLOBAL_ROLE in
// apps/packages/app/src/apps/controlPanelCommon/constant/user.ts and the
// isAdmin getter in apps/.../controlHub/stores/AppDetail.ts.
const GlobalRoleAdmin = "platform-admin"

// These mirror the SPA's AppDetail store literals (NAMESPACE_SHARED /
// OS_PROTECTED) used by hasPermission to widen an admin's exec scope
// beyond their own namespaces.
const (
	namespaceSharedSuffix = "-shared"
	namespaceOSProtected  = "os-protected"
)

// CanExecNamespace reports whether the identity (username / globalRole /
// systemNamespaces from /capi/app/detail) is allowed to open a terminal
// (exec) into a pod living in `namespace`.
//
// It is a faithful port of the ControlHub SPA's exec gate — the
// hasPermission(namespace) action in
// apps/packages/app/src/apps/controlHub/stores/AppDetail.ts, which decides
// whether the per-container Terminal button is rendered
// (v-permission="route.params.namespace" in ContainersItem.vue):
//
//   - admin (globalrole == platform-admin): allowed when the namespace
//     contains the admin's own username, OR is one of the server-provided
//     systemNamespaces, OR ends with "-shared", OR equals "os-protected".
//   - everyone else: allowed only when the namespace contains their username.
//
// The username test is a substring match (JS `value.includes(username)`),
// which lines up with Olares' per-user namespace naming
// (user-space-<username> / user-system-<username>): the main account
// (e.g. "admin") therefore may NOT exec into a sub-account's namespace such
// as "user-space-alice", matching the SPA hiding the Terminal button there.
//
// This governs exec only. Listing/viewing pods, containers, images and logs
// is intentionally NOT gated here — the SPA leaves those visible to admins
// too, and the CLI keeps that parity.
//
// Callers must ensure username is non-empty (a resolved identity always
// carries one); an empty username makes the substring test degenerate.
func CanExecNamespace(namespace, username, globalRole string, systemNamespaces []string) bool {
	ns := strings.TrimSpace(namespace)
	if ns == "" || username == "" {
		return false
	}
	if globalRole == GlobalRoleAdmin {
		if strings.Contains(ns, username) {
			return true
		}
		for _, sys := range systemNamespaces {
			if sys == ns {
				return true
			}
		}
		return strings.HasSuffix(ns, namespaceSharedSuffix) || ns == namespaceOSProtected
	}
	return strings.Contains(ns, username)
}

// CanExec is the Info-bound convenience wrapper over CanExecNamespace, so
// callers holding a freshly-fetched /capi/app/detail snapshot can gate exec
// with i.CanExec(namespace).
func (i Info) CanExec(namespace string) bool {
	return CanExecNamespace(namespace, i.Username, i.GlobalRole, i.SystemNamespaces)
}
