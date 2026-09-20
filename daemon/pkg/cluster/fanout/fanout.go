// Package fanout is a generic mechanism for the master olaresd to dispatch the
// same node-local action to a caller-chosen set of node olaresds and aggregate
// the results per node. It is intentionally decoupled from any specific command:
// a node being unreachable is a first-class result, not a silently dropped item.
//
// Target selection and per-call timeout belong to the consumer. Inventory is the
// node directory; collect-logs filters Ready nodes and sets a 15-minute timeout;
// power ops pass the full inventory with their own short timeout.
package fanout

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// OlaresdPort is the fixed port every node's olaresd listens on.
	OlaresdPort = 18088

	// fallbackTimeout applies only when a caller forgets Timeout. Consumers must
	// set their own bound; log collection must not rely on this value.
	fallbackTimeout = 30 * time.Second

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
	// Headers are forwarded verbatim, for a node-local endpoint that needs
	// more than the access token to accept the call.
	Headers map[string]string
	// Port overrides the olaresd port. Zero means OlaresdPort; anything else
	// is a test pointing the fan-out at a server it started.
	Port int
	// Timeout bounds each per-node call. Callers must set this; zero falls
	// back to fallbackTimeout only as a safety net.
	Timeout time.Duration
	// Parallel bounds concurrent calls. Defaults to defaultParallel.
	Parallel int
}

// Run dispatches to every target concurrently and returns one result per
// target. A failing node never aborts the others.
func (d *Dispatcher) Run(ctx context.Context, targets []NodeTarget, payloadFor func(NodeTarget) any) []NodeResult {
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = fallbackTimeout
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
	port := d.Port
	if port == 0 {
		port = OlaresdPort
	}
	url := nodeURL(host, port, d.PeerPath)

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
	for k, v := range d.Headers {
		req.Header.Set(k, v)
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

func nodeURL(host string, port int, path string) string {
	return "http://" + net.JoinHostPort(host, strconv.Itoa(port)) + path
}
