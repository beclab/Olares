package collectlogs

// Wildcard means "all" within a group. For namespaces it is role-relative
// (owner/admin => every namespace; normal => every namespace they own); for
// systemd components it means every known Olares service.
const Wildcard = "*"

// SystemdGroup selects systemd service logs. Components nil/empty => skip the
// group; ["*"] => all known services; a concrete list => that subset. Since
// and MaxLines only affect this group (journalctl).
type SystemdGroup struct {
	Components []string `json:"components,omitempty"`
	Since      string   `json:"since,omitempty"`
	MaxLines   int      `json:"maxLines,omitempty"`
}

// HostGroup selects node-level, non-namespaced host data.
type HostGroup struct {
	Dmesg   bool `json:"dmesg,omitempty"`
	Network bool `json:"network,omitempty"`
}

// ClusterGroup selects cluster-level info (kubectl describe / pods-list /
// envoy config) plus olares-cli's own logs.
type ClusterGroup struct {
	Info bool `json:"info,omitempty"`
}

// NamespacesGroup selects pod logs by namespace. Names nil/empty => skip;
// ["*"] => all (role-relative); a concrete list => those namespaces.
type NamespacesGroup struct {
	Names []string `json:"names,omitempty"`
}

// Param is the collect-logs request contract.
type Param struct {
	Systemd    SystemdGroup    `json:"systemd"`
	Host       HostGroup       `json:"host"`
	Cluster    ClusterGroup    `json:"cluster"`
	Namespaces NamespacesGroup `json:"namespaces"`

	// CallerOlaresID is injected by the handler from the signature-verified
	// client; it is intentionally not decoded from the request body.
	CallerOlaresID string `json:"-"`
	// CallerSignature is the verified X-Signature, injected by the master
	// handler so the orchestrator can forward it to each node's olaresd.
	CallerSignature string `json:"-"`
}

// NodeRequest is what the master orchestrator sends to each node's node-local
// endpoint. It carries the (un-expanded) user scope plus the run id used to
// compute the shared staging directory. Identity fields stay json:"-" so each
// node re-derives them from its own signature check.
type NodeRequest struct {
	Param
	RunID string `json:"runID"`
}

func hasWildcard(items []string) bool {
	for _, it := range items {
		if it == Wildcard {
			return true
		}
	}
	return false
}
