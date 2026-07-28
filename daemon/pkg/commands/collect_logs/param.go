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

	// CallerUsername and CallerRole are injected by the handler from the
	// access-token-verified caller (utils.ValidToken); they are intentionally
	// not decoded from the request body.
	CallerUsername string `json:"-"`
	CallerRole     string `json:"-"`
	// CallerToken is the verified X-Authorization access token, injected by the
	// master handler so the orchestrator can forward it to each node's olaresd.
	CallerToken string `json:"-"`
}

// CollectResult is returned synchronously when a collection is accepted. The
// archive is produced asynchronously, but its location is deterministic, so the
// caller can poll state and then fetch File from their Home via Files.
type CollectResult struct {
	RunID string `json:"runID"`
	File  string `json:"file"`
	// Path is the archive path relative to the caller's Files root, e.g.
	// "Home/pod_logs/olares-logs-<runID>.tar.gz".
	Path string `json:"path"`
}

// NodeRequest is what the master orchestrator sends to each node's node-local
// endpoint. It carries the (un-expanded) user scope plus the run id used to
// compute the shared staging directory. Identity fields stay json:"-" so each
// node re-derives them from its own access-token check.
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
