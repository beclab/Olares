// Package fanout is a generic mechanism for the master olaresd to dispatch the
// same node-local action to every node's olaresd and aggregate the results per
// node. It is intentionally decoupled from any specific command (collect-logs
// is the first consumer): a node being unreachable is a first-class result, not
// a silently dropped item.
package fanout

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/nets"
	"github.com/beclab/Olares/daemon/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	// OlaresdPort is the fixed port every node's olaresd listens on.
	OlaresdPort = 18088

	// DefaultTimeout bounds each per-node dispatch call. Exported so node-local
	// executors can cap their own work to finish before the master gives up.
	DefaultTimeout = 15 * time.Minute

	defaultParallel = 4
	authHeader      = "X-Authorization"
)

// NodeStatus classifies a per-node dispatch outcome.
type NodeStatus string

const (
	StatusOK          NodeStatus = "ok"
	StatusUnreachable NodeStatus = "unreachable"
	StatusTimeout     NodeStatus = "timeout"
	StatusError       NodeStatus = "error"
)

// NodeTarget identifies one node and how to reach its olaresd.
type NodeTarget struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	IsSelf   bool   `json:"isSelf"`
	IsMaster bool   `json:"isMaster"`
}

// NodeResult is the generic per-node outcome. Data carries the consumer's
// node-local payload verbatim and is opaque to the fan-out layer.
type NodeResult struct {
	Node   NodeTarget      `json:"node"`
	Status NodeStatus      `json:"status"`
	Err    string          `json:"err,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

// Dispatcher fans a request out to a set of node-local olaresd endpoints.
type Dispatcher struct {
	// PeerPath is the node-local endpoint path, e.g. "/command/collect-logs-node".
	PeerPath string
	// AuthToken is forwarded as X-Authorization so each node authenticates the
	// same caller.
	AuthToken string
	// Timeout bounds each per-node call. Defaults to defaultTimeout.
	Timeout time.Duration
	// Parallel bounds concurrent calls. Defaults to defaultParallel.
	Parallel int
}

// ListReadyNodes enumerates Ready nodes with an internal IP, tagging the local
// node and master nodes.
func ListReadyNodes(ctx context.Context) ([]NodeTarget, error) {
	client, err := utils.GetKubeClient()
	if err != nil {
		return nil, err
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	selfIPs := map[string]struct{}{}
	if ips, err := nets.LookupHostIps(); err != nil {
		klog.Warningf("fanout: lookup host ips error: %v", err)
	} else {
		for _, ip := range ips {
			selfIPs[ip] = struct{}{}
		}
	}

	var targets []NodeTarget
	for i := range nodes.Items {
		n := &nodes.Items[i]
		if !utils.IsNodeReady(n) {
			continue
		}
		ip := internalIP(n)
		if ip == "" {
			continue
		}
		_, self := selfIPs[ip]
		targets = append(targets, NodeTarget{
			Name:     n.Name,
			IP:       ip,
			IsSelf:   self,
			IsMaster: utils.IsMasterNode(n),
		})
	}
	return targets, nil
}

// Run dispatches to every target concurrently and returns one result per
// target. A failing node never aborts the others.
func (d *Dispatcher) Run(ctx context.Context, targets []NodeTarget, payloadFor func(NodeTarget) any) []NodeResult {
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	parallel := d.Parallel
	if parallel <= 0 {
		parallel = defaultParallel
	}

	results := make([]NodeResult, len(targets))
	sem := make(chan struct{}, parallel)
	var wg sync.WaitGroup
	for idx := range targets {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results[i] = d.call(ctx, targets[i], payloadFor(targets[i]), timeout)
		}(idx)
	}
	wg.Wait()
	return results
}

func (d *Dispatcher) call(ctx context.Context, t NodeTarget, payload any, timeout time.Duration) NodeResult {
	res := NodeResult{Node: t}

	body, err := json.Marshal(payload)
	if err != nil {
		res.Status = StatusError
		res.Err = fmt.Sprintf("marshal payload: %v", err)
		return res
	}

	host := t.IP
	if t.IsSelf {
		host = "127.0.0.1"
	}
	url := fmt.Sprintf("http://%s:%d%s", host, OlaresdPort, d.PeerPath)

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		res.Status = StatusError
		res.Err = err.Error()
		return res
	}
	req.Header.Set("Content-Type", "application/json")
	if d.AuthToken != "" {
		req.Header.Set(authHeader, d.AuthToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			res.Status = StatusTimeout
		} else {
			res.Status = StatusUnreachable
		}
		res.Err = err.Error()
		return res
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		res.Status = StatusError
		res.Err = fmt.Sprintf("node returned %d: %s", resp.StatusCode, string(respBody))
		return res
	}

	res.Status = StatusOK
	res.Data = respBody
	return res
}

func internalIP(n *corev1.Node) string {
	for _, addr := range n.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
