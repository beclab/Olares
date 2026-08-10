package clusterop

import (
	"context"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/fanout"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
)

const (
	// PeerPath is the node-local endpoint the master hands a power command to.
	PeerPath = "/command/power-node"

	// dispatchTimeout bounds one power command. The node answers as soon as it
	// has accepted the command, before it acts on it, so this is short.
	dispatchTimeout = 30 * time.Second

	// inspectTimeout bounds one node-status read, used by the precheck and by
	// every restart probe.
	inspectTimeout = 10 * time.Second
)

// NewDeps assembles the orchestrator's side effects against the real cluster.
// Everything it builds is a seam the tests replace; this is the only place the
// production versions are wired together.
// The store is left out: it is opened from the state directory by whoever
// installs these, so this function has nothing to decide about the disk.
func NewDeps() Deps {
	return Deps{
		Inventory:  inventory.List,
		Inspect:    inspectNode,
		Dispatch:   dispatchPower,
		Observe:    inventory.Observe,
		HostBootID: HostBootID,
		PowerSelf:  PowerHost,

		// The same function PowerHost consults, so the precheck's answer
		// about this machine is the execution point's answer.
		LocalPowerSupport: LocalPowerSupport,
	}
}

// dispatchPower hands the power command to each node's own endpoint through
// the existing fan-out, which reports an unreachable node as a result rather
// than dropping it.
//
// The only credential that crosses to a worker is the operation-bound
// signature. The caller's access token is deliberately left behind: it opens
// every route that user can reach, for as long as it lives, so forwarding it
// would hand each node in the cluster a general-purpose credential in order to
// ask one question. The signature names the operation it authorizes and
// nothing else, and the node checks it against the request it received.
func dispatchPower(ctx context.Context, nodes []inventory.Node, req PeerRequest, creds Credentials) []DispatchOutcome {
	d := &fanout.Dispatcher{
		PeerPath: PeerPath,
		Headers:  signatureHeader(creds),
		Timeout:  dispatchTimeout,
	}
	results := d.Run(ctx, peerTargets(nodes), func(fanout.NodeTarget) any { return req })
	return peerOutcomes(results)
}

func signatureHeader(creds Credentials) map[string]string {
	if creds.Signature == "" {
		return nil
	}
	return map[string]string{signatureHeaderName: creds.Signature}
}

// signatureHeaderName matches the header the power endpoints read. The second
// hop presents the same owner signature as the first, so a node never accepts
// on weaker evidence than the master did.
const signatureHeaderName = "X-Signature"
const authorizationHeaderName = "X-Authorization"

func nodeStatusHeaders(creds Credentials) map[string]string {
	if creds.Token != "" {
		return map[string]string{authorizationHeaderName: creds.Token}
	}
	return signatureHeader(creds)
}

// peerTargets describes the nodes to the fan-out. A node's address comes from
// the master's directory, never from a request.
func peerTargets(nodes []inventory.Node) []fanout.NodeTarget {
	targets := make([]fanout.NodeTarget, 0, len(nodes))
	for _, n := range nodes {
		targets = append(targets, fanout.NodeTarget{
			Name:     n.NodeName,
			IP:       n.IP,
			IsSelf:   n.IsSelf,
			IsMaster: n.Role == inventory.RoleMaster,
		})
	}
	return targets
}

// peerOutcomes turns fan-out results into the operation's own vocabulary. A
// node that never answered and a node that refused are different failures, and
// the record keeps them apart.
func peerOutcomes(results []fanout.NodeResult) []DispatchOutcome {
	outcomes := make([]DispatchOutcome, 0, len(results))
	for _, r := range results {
		o := DispatchOutcome{NodeName: r.Node.Name, Err: r.Err}
		switch r.Status {
		case fanout.StatusOK:
			o.Err = ""
		case fanout.StatusUnreachable, fanout.StatusTimeout:
			o.Code = CodeNodeUnreachable
		default:
			o.Code = CodeDispatchFailed
		}
		if o.Code != "" && o.Err == "" {
			o.Err = string(r.Status)
		}
		outcomes = append(outcomes, o)
	}
	return outcomes
}

// inspectNode reads one node's own report of itself, which is where its power
// capabilities are declared. This is a precheck-only call, made once per node
// while the cluster is still whole: nothing during the reboot itself depends
// on a node answering an endpoint that needs a user credential.
func inspectNode(ctx context.Context, node inventory.Node, creds Credentials) (nodestatus.Status, error) {
	if node.IP == "" {
		return nodestatus.Status{}, errors.New("node has no internal address")
	}
	host := node.IP
	if node.IsSelf {
		host = "127.0.0.1"
	}

	reqCtx, cancel := context.WithTimeout(ctx, inspectTimeout)
	defer cancel()
	baseURL := "http://" + net.JoinHostPort(host, strconv.Itoa(fanout.OlaresdPort))
	return nodestatus.FetchWithHeaders(reqCtx, baseURL, nodeStatusHeaders(creds))
}

// bootIDPath is where the kernel publishes an identifier regenerated on every
// boot. Reading it is how this daemon tells its own restart apart from the
// machine's.
const bootIDPath = "/proc/sys/kernel/random/boot_id"

// HostBootID reports the boot the machine running this daemon is on. An empty
// value is not an error to the caller that matters: a reboot that cannot be
// proved is left at command_issued rather than promoted.
func HostBootID() (string, error) {
	raw, err := os.ReadFile(bootIDPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(raw)), nil
}
