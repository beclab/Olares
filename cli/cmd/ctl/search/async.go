package search

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

const (
	asyncSearchMinOlaresVersion = "1.12.7"
	asyncSearchLoginTimeout     = 15 * time.Second
	asyncSearchHeartbeat        = 25 * time.Second
	// user-service publishes its own timeout job_finished at ten minutes.
	// Leave enough time for that terminal frame to cross the WebSocket.
	asyncSearchJobTimeout = 10*time.Minute + 30*time.Second
)

var federatedFileSources = []string{appFilesV2, appGoogleDrive, appDropbox, appSeafile}

type asyncSearchMessage struct {
	MessageKind   string                        `json:"message_kind"`
	ReqID         string                        `json:"reqid"`
	Source        string                        `json:"source"`
	Hits          []json.RawMessage             `json:"hits"`
	HitsRange     []int                         `json:"hits_range"`
	IsFinalForJob bool                          `json:"is_final_for_job"`
	JobStatus     string                        `json:"job_status"`
	Error         string                        `json:"error"`
	SourceSummary map[string]asyncSourceSummary `json:"sources_summary"`
}

type asyncSourceSummary struct {
	Status   string `json:"status"`
	HitCount int    `json:"hit_count"`
}

type asyncMoreResponse struct {
	Entries []struct {
		Index int             `json:"index"`
		Hit   json.RawMessage `json:"hit"`
	} `json:"entries"`
}

type asyncHitCollector struct {
	bySource map[string]map[int]json.RawMessage
}

type asyncIndexedHit struct {
	Source string
	Index  int
	Hit    json.RawMessage
}

func newAsyncHitCollector() *asyncHitCollector {
	return &asyncHitCollector{bySource: make(map[string]map[int]json.RawMessage)}
}

func (c *asyncHitCollector) addBatch(source string, start int, hits []json.RawMessage) []asyncIndexedHit {
	if source == "" || len(hits) == 0 {
		return nil
	}
	if c.bySource[source] == nil {
		c.bySource[source] = make(map[int]json.RawMessage)
	}
	added := make([]asyncIndexedHit, 0, len(hits))
	for i, hit := range hits {
		index := start + i
		if _, exists := c.bySource[source][index]; !exists {
			c.bySource[source][index] = hit
			added = append(added, asyncIndexedHit{Source: source, Index: index, Hit: hit})
		}
	}
	return added
}

func (c *asyncHitCollector) missingRanges(source string, total int) [][]int {
	if total <= 0 {
		return nil
	}
	received := c.bySource[source]
	var ranges [][]int
	start := -1
	for i := 0; i < total; i++ {
		_, ok := received[i]
		if !ok && start < 0 {
			start = i
		}
		if ok && start >= 0 {
			ranges = append(ranges, []int{start, i - 1})
			start = -1
		}
	}
	if start >= 0 {
		ranges = append(ranges, []int{start, total - 1})
	}
	return ranges
}

func (c *asyncHitCollector) rows(sources []string) []json.RawMessage {
	var rows []json.RawMessage
	for _, source := range sources {
		indexed := c.bySource[source]
		indices := make([]int, 0, len(indexed))
		for index := range indexed {
			indices = append(indices, index)
		}
		sort.Ints(indices)
		for _, index := range indices {
			rows = append(rows, indexed[index])
		}
	}
	return rows
}

func runVersionedFileSearch(ctx context.Context, f *cmdutil.Factory, keyword, searchType string, o *pagingOptions, onHit func(asyncIndexedHit) error) ([]resultItem, bool, error) {
	useAsync, err := f.OlaresBackendAtLeast(ctx, asyncSearchMinOlaresVersion)
	if err != nil {
		return nil, false, err
	}
	if !useAsync {
		items, err := runSessionSearch(ctx, f, keyword, appFilesV2, searchType, o)
		return items, false, err
	}
	items, err := runAsyncSearch(ctx, f, keyword, federatedFileSources, searchType, o, onHit)
	return items, true, err
}

func runAsyncSearch(ctx context.Context, f *cmdutil.Factory, keyword string, sources []string, searchType string, o *pagingOptions, onHit func(asyncIndexedHit) error) ([]resultItem, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if f == nil {
		return nil, fmt.Errorf("internal error: search not wired with cmdutil.Factory")
	}

	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return nil, err
	}
	token, err := f.ValidAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	hc, err := f.HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	doer := whoami.NewHTTPClient(hc, rp.DesktopURL, rp.OlaresID)

	conn, err := dialSearchWebSocket(ctx, rp, token)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := loginSearchWebSocket(conn); err != nil {
		return nil, err
	}

	reqid := uuid.NewString()
	defer cancelAsyncSearch(doer, reqid)

	initBody := map[string]interface{}{
		"reqid":   reqid,
		"keyword": keyword,
		"type":    searchType,
		"sources": sources,
	}
	if err := doEnvelope(ctx, doer, "POST", "/api/search/nats/init", initBody, nil); err != nil {
		return nil, err
	}

	collector := newAsyncHitCollector()
	finished, err := collectAsyncMessages(ctx, conn, reqid, collector, onHit)
	if err != nil {
		return nil, err
	}
	if err := asyncSearchTerminalError(finished); err != nil {
		return nil, err
	}
	if err := recoverMissingAsyncHits(ctx, doer, reqid, sources, finished.SourceSummary, collector, onHit); err != nil {
		return nil, err
	}

	window := paginateRaw(collector.rows(sources), o.offset, o.limit)
	return decodeResultRows(window)
}

func asyncSearchTerminalError(message asyncSearchMessage) error {
	switch message.JobStatus {
	case "failed", "timeout", "cancelled":
		detail := strings.TrimSpace(message.Error)
		if detail == "" {
			detail = "job finished with status " + message.JobStatus
		}
		return fmt.Errorf("asynchronous search %s: %s", message.JobStatus, detail)
	default:
		return nil
	}
}

func dialSearchWebSocket(ctx context.Context, rp *credential.ResolvedProfile, token string) (*websocket.Conn, error) {
	wsURL, err := searchWebSocketURL(rp.DesktopURL)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{HandshakeTimeout: asyncSearchLoginTimeout}
	if rp.InsecureSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	headers := http.Header{}
	headers.Set("X-Authorization", token)
	headers.Set("X-Unauth-Error", "Non-Redirect")
	headers.Set("Cookie", "auth_token="+token)
	conn, resp, err := dialer.DialContext(ctx, wsURL, headers)
	if err == nil {
		return conn, nil
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == 459 {
			return nil, fmt.Errorf("search WebSocket authentication failed (HTTP %d); please run: olares-cli profile login", resp.StatusCode)
		}
		return nil, fmt.Errorf("search WebSocket handshake failed (HTTP %d)", resp.StatusCode)
	}
	return nil, fmt.Errorf("search WebSocket handshake failed: %w", err)
}

func searchWebSocketURL(desktopURL string) (string, error) {
	u, err := url.Parse(strings.TrimRight(desktopURL, "/"))
	if err != nil {
		return "", fmt.Errorf("parse Desktop URL: %w", err)
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	default:
		return "", fmt.Errorf("unsupported Desktop URL scheme %q", u.Scheme)
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/ws"
	return u.String(), nil
}

func loginSearchWebSocket(conn *websocket.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(asyncSearchLoginTimeout)); err != nil {
		return err
	}
	login := map[string]interface{}{
		"event": "login",
		"data": map[string]interface{}{
			"application": "desktop",
			// user-service indexes live application connections by (application,
			// token). Reusing the access token here would replace a real Desktop
			// connection. Authentication already happened in the WS handshake, so
			// use an isolated routing token for this short-lived CLI connection.
			"token": "olares-cli-" + uuid.NewString(),
			"id":    "olares-cli-" + uuid.NewString(),
		},
	}
	if err := conn.WriteJSON(login); err != nil {
		return fmt.Errorf("send search WebSocket login: %w", err)
	}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("wait for search WebSocket login: %w", err)
		}
		var ack struct {
			Event string `json:"event"`
			Topic string `json:"topic"`
		}
		if json.Unmarshal(raw, &ack) == nil && (ack.Event == "pong" || ack.Topic == "pong") {
			return conn.SetReadDeadline(time.Time{})
		}
	}
}

func collectAsyncMessages(ctx context.Context, conn *websocket.Conn, reqid string, collector *asyncHitCollector, onHit func(asyncIndexedHit) error) (asyncSearchMessage, error) {
	deadline := time.Now().Add(asyncSearchJobTimeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetReadDeadline(deadline); err != nil {
		return asyncSearchMessage{}, err
	}

	done := make(chan struct{})
	heartbeatErr := make(chan error, 1)
	defer close(done)
	go func() {
		ticker := time.NewTicker(asyncSearchHeartbeat)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				_ = conn.Close()
				return
			case <-ticker.C:
				if err := conn.WriteJSON(map[string]string{"event": "ping"}); err != nil {
					select {
					case heartbeatErr <- err:
					default:
					}
					_ = conn.Close()
					return
				}
			case <-done:
				return
			}
		}
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return asyncSearchMessage{}, ctx.Err()
			}
			select {
			case heartbeatCause := <-heartbeatErr:
				return asyncSearchMessage{}, fmt.Errorf("send asynchronous search heartbeat: %w", heartbeatCause)
			default:
			}
			var netErr interface{ Timeout() bool }
			if errors.As(err, &netErr) && netErr.Timeout() {
				return asyncSearchMessage{}, fmt.Errorf("asynchronous search timed out after %s", asyncSearchJobTimeout)
			}
			return asyncSearchMessage{}, fmt.Errorf("read asynchronous search result: %w", err)
		}
		var message asyncSearchMessage
		if err := json.Unmarshal(raw, &message); err != nil || message.ReqID != reqid {
			continue
		}
		if message.MessageKind == "hits_batch" {
			start := 0
			if len(message.HitsRange) > 0 {
				start = message.HitsRange[0]
			}
			for _, hit := range collector.addBatch(message.Source, start, message.Hits) {
				if onHit != nil {
					if err := onHit(hit); err != nil {
						return asyncSearchMessage{}, err
					}
				}
			}
		}
		if message.MessageKind == "job_finished" || message.IsFinalForJob {
			return message, nil
		}
	}
}

func recoverMissingAsyncHits(ctx context.Context, doer *whoami.HTTPClient, reqid string, sources []string, summaries map[string]asyncSourceSummary, collector *asyncHitCollector, onHit func(asyncIndexedHit) error) error {
	for _, source := range sources {
		missing := collector.missingRanges(source, summaries[source].HitCount)
		if len(missing) == 0 {
			continue
		}
		body := map[string]interface{}{
			"reqid":          reqid,
			"source":         source,
			"missing_ranges": missing,
		}
		var more asyncMoreResponse
		if err := doEnvelope(ctx, doer, "POST", "/api/search/nats/more", body, &more); err != nil {
			return err
		}
		for _, entry := range more.Entries {
			for _, hit := range collector.addBatch(source, entry.Index, []json.RawMessage{entry.Hit}) {
				if onHit != nil {
					if err := onHit(hit); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func cancelAsyncSearch(doer *whoami.HTTPClient, reqid string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = doEnvelope(ctx, doer, "POST", "/api/search/nats/cancel", map[string]interface{}{"reqid": reqid}, nil)
}
