# Cluster Exec Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `olares-cli cluster {pod,container} exec` — run a command inside a container (one-shot, AI-friendly with structured JSON + real exit code) or attach an interactive `-it` shell — over the native Kubernetes exec WebSocket.

**Architecture:** A new pure-Go protocol package (`cli/pkg/clusterexec`) implements the `v4.channel.k8s.io` framing and exit-code parsing; a gorilla/websocket client dials `wss://control-hub.<terminus>/api/v1/.../exec` with the three required auth headers. Thin cobra verbs under `pod` and `container` mirror the existing `logs` verbs. One edge nginx location is extended to allow the exec WebSocket upgrade. The skill doc teaches the AI the contract + ephemeral-fix caveat.

**Tech Stack:** Go, cobra, `github.com/gorilla/websocket` (already vendored), `golang.org/x/term`, `k8s.io` metav1 status shape. Spec: `docs/superpowers/specs/2026-06-18-cluster-exec-design.md`.

---

## Task 1: Protocol package (pure framer) — TDD

**Files:**
- Create: `cli/pkg/clusterexec/protocol.go`
- Test: `cli/pkg/clusterexec/protocol_test.go`

- [ ] **Step 1: Write the failing test**

```go
// cli/pkg/clusterexec/protocol_test.go
package clusterexec

import "testing"

func TestParseExitStatus(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    int
	}{
		{"success", `{"status":"Success"}`, 0},
		{"nonzero", `{"status":"Failure","reason":"NonZeroExitCode","details":{"causes":[{"reason":"ExitCode","message":"127"}]}}`, 127},
		{"failure no cause", `{"status":"Failure","message":"boom"}`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseExitStatus([]byte(tc.payload))
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSinkCaps(t *testing.T) {
	s := NewSink(4)
	s.Write(ChannelStdout, []byte("ab"))
	s.Write(ChannelStdout, []byte("cdef"))
	s.Write(ChannelStderr, []byte("xy"))
	if string(s.Stdout) != "abcd" {
		t.Fatalf("stdout = %q", s.Stdout)
	}
	if string(s.Stderr) != "xy" {
		t.Fatalf("stderr = %q", s.Stderr)
	}
	if !s.Truncated {
		t.Fatalf("expected Truncated=true")
	}
}

func TestResizeFrame(t *testing.T) {
	f, err := ResizeFrame(80, 24)
	if err != nil {
		t.Fatal(err)
	}
	if f[0] != ChannelResize {
		t.Fatalf("channel byte = %d", f[0])
	}
	if string(f[1:]) != `{"Width":80,"Height":24}` {
		t.Fatalf("payload = %q", f[1:])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd cli && go test ./pkg/clusterexec/...`
Expected: FAIL — package/functions not defined.

- [ ] **Step 3: Write minimal implementation**

```go
// cli/pkg/clusterexec/protocol.go

// Package clusterexec implements the client side of the Kubernetes
// remote-command (exec) WebSocket subprotocol "v4.channel.k8s.io" used by
// `olares-cli cluster {pod,container} exec`. The protocol package is pure
// (no I/O) so the framing + exit-code logic is unit-testable; the dialing
// and stream pumping live in client.go.
package clusterexec

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Subprotocol is the WebSocket subprotocol negotiated for exec. v4 is the
// stable, widely-deployed channel protocol (one leading channel byte per
// binary frame).
const Subprotocol = "v4.channel.k8s.io"

// Channel identifiers — the first byte of every binary frame.
const (
	ChannelStdin  byte = 0
	ChannelStdout byte = 1
	ChannelStderr byte = 2
	ChannelError  byte = 3
	ChannelResize byte = 4
)

// execStatus mirrors the subset of metav1.Status the error channel emits
// when the remote process exits.
type execStatus struct {
	Status  string `json:"status"`
	Details *struct {
		Causes []struct {
			Reason  string `json:"reason"`
			Message string `json:"message"`
		} `json:"causes"`
	} `json:"details"`
}

// ParseExitStatus decodes a channel-3 payload into an exit code.
// status=="Success" -> 0; otherwise the details.causes entry with
// reason=="ExitCode" supplies the code; if absent, 1.
func ParseExitStatus(payload []byte) (int, error) {
	var s execStatus
	if err := json.Unmarshal(payload, &s); err != nil {
		return 0, fmt.Errorf("decode exec status: %w (body=%q)", err, string(payload))
	}
	if s.Status == "Success" {
		return 0, nil
	}
	if s.Details != nil {
		for _, c := range s.Details.Causes {
			if c.Reason == "ExitCode" {
				if code, err := strconv.Atoi(c.Message); err == nil {
					return code, nil
				}
				return 1, nil
			}
		}
	}
	return 1, nil
}

// Frame prepends the channel byte to payload, producing one binary message.
func Frame(channel byte, payload []byte) []byte {
	out := make([]byte, 0, len(payload)+1)
	out = append(out, channel)
	return append(out, payload...)
}

// ResizeFrame builds a channel-4 terminal resize message. The wire shape is
// remotecommand.TerminalSize JSON: {"Width":w,"Height":h}.
func ResizeFrame(cols, rows uint16) ([]byte, error) {
	b, err := json.Marshal(struct {
		Width  uint16 `json:"Width"`
		Height uint16 `json:"Height"`
	}{Width: cols, Height: rows})
	if err != nil {
		return nil, err
	}
	return Frame(ChannelResize, b), nil
}

// Sink accumulates stdout/stderr up to maxBytes per stream (0 = unlimited),
// setting Truncated when a cap is hit.
type Sink struct {
	maxBytes  int
	Stdout    []byte
	Stderr    []byte
	Truncated bool
}

// NewSink builds a Sink with a per-stream cap (0 = unlimited).
func NewSink(maxBytes int) *Sink { return &Sink{maxBytes: maxBytes} }

// Write routes a demultiplexed frame into the right buffer. Non stdout/stderr
// channels are ignored (the caller handles the error channel separately).
func (s *Sink) Write(channel byte, payload []byte) {
	switch channel {
	case ChannelStdout:
		s.Stdout = s.appendCapped(s.Stdout, payload)
	case ChannelStderr:
		s.Stderr = s.appendCapped(s.Stderr, payload)
	}
}

func (s *Sink) appendCapped(buf, p []byte) []byte {
	if s.maxBytes <= 0 {
		return append(buf, p...)
	}
	room := s.maxBytes - len(buf)
	if room <= 0 {
		s.Truncated = true
		return buf
	}
	if len(p) > room {
		s.Truncated = true
		return append(buf, p[:room]...)
	}
	return append(buf, p...)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd cli && go test ./pkg/clusterexec/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cli/pkg/clusterexec/protocol.go cli/pkg/clusterexec/protocol_test.go
git commit -m "feat(cli): add clusterexec v4 channel protocol framer"
```

---

## Task 2: Factory token helper for non-transport callers

The exec WebSocket handshake bypasses `HTTPClient`'s refreshing transport, so
it needs a token guaranteed fresh. Add a small `Factory.ValidAccessToken` that
reuses the existing expiry + refresh machinery.

**Files:**
- Modify: `cli/pkg/cmdutil/factory.go` (add one method near `ValidAccessToken` neighbors, after `Refresher`)

- [ ] **Step 1: Add the method**

```go
// ValidAccessToken returns an access token fresh enough for an immediate
// request, refreshing via /api/refresh if the cached token is within
// preflightSkew of expiry. Used by callers that bypass HTTPClient's transport
// (e.g. the exec WebSocket handshake) and therefore can't rely on the reactive
// 401-refresh path. On a missing/non-JWT exp claim the token is returned as-is
// (a 401 at use time will surface the login CTA).
func (f *Factory) ValidAccessToken(ctx context.Context) (string, error) {
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return "", err
	}
	cell := f.sharedTokenCell(rp.AccessToken)
	tok := cell.snapshot()
	if tok == "" {
		tok = rp.AccessToken
	}
	expired, err := auth.IsExpired(tok, time.Now(), preflightSkew)
	if err != nil || !expired {
		return tok, nil
	}
	newAT, rerr := f.Refresher().Refresh(ctx, rp.OlaresID, rp.AuthURL, tok, rp.InsecureSkipVerify)
	if rerr != nil {
		return "", rerr
	}
	cell.update(newAT)
	return newAT, nil
}
```

(`auth`, `context`, `time` are already imported in factory.go; `preflightSkew`,
`sharedTokenCell`, `tokenCell.snapshot/update` already exist.)

- [ ] **Step 2: Verify it compiles**

Run: `cd cli && go build ./pkg/cmdutil/...`
Expected: builds clean.

- [ ] **Step 3: Commit**

```bash
git add cli/pkg/cmdutil/factory.go
git commit -m "feat(cli): add Factory.ValidAccessToken for non-transport callers"
```

---

## Task 3: WebSocket exec client (gorilla)

**Files:**
- Create: `cli/pkg/clusterexec/client.go`

- [ ] **Step 1: Write the client**

```go
// cli/pkg/clusterexec/client.go
package clusterexec

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// Options describes one exec invocation.
type Options struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
	Stdin     bool
	TTY       bool
}

// Result is the outcome of a one-shot exec.
type Result struct {
	Stdout    []byte
	Stderr    []byte
	ExitCode  *int // nil = unknown (timeout / no status frame)
	Truncated bool
}

// buildWSURL turns the ControlHub base URL + options into the exec WS URL.
func buildWSURL(controlHubURL string, o Options) (string, error) {
	base := strings.TrimRight(controlHubURL, "/")
	base = strings.Replace(base, "https://", "wss://", 1)
	base = strings.Replace(base, "http://", "ws://", 1)
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/exec",
		base, url.PathEscape(o.Namespace), url.PathEscape(o.Pod)))
	if err != nil {
		return "", err
	}
	q := url.Values{}
	if o.Container != "" {
		q.Set("container", o.Container)
	}
	for _, c := range o.Command {
		q.Add("command", c)
	}
	if o.Stdin {
		q.Set("stdin", "true")
	}
	q.Set("stdout", "true")
	if o.TTY {
		q.Set("tty", "true") // stderr MUST be omitted in TTY mode (PTY merges it)
	} else {
		q.Set("stderr", "true")
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// dial opens the exec WebSocket with the three auth headers the
// control-hub -> K8s path requires (X-Authorization + auth_token cookie +
// X-Unauth-Error; without the cookie the proxy returns 403 system:anonymous —
// see cli/pkg/cmdutil/factory.go refreshingTransport.send).
func dial(ctx context.Context, controlHubURL, token string, insecure bool, o Options) (*websocket.Conn, *http.Response, error) {
	wsURL, err := buildWSURL(controlHubURL, o)
	if err != nil {
		return nil, nil, err
	}
	d := websocket.Dialer{
		Subprotocols:     []string{Subprotocol},
		HandshakeTimeout: 30 * time.Second,
	}
	if insecure {
		d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicit profile opt-in
	}
	h := http.Header{}
	h.Set("X-Authorization", token)
	h.Set("X-Unauth-Error", "Non-Redirect")
	h.Set("Cookie", "auth_token="+token)
	return d.DialContext(ctx, wsURL, h)
}

// handshakeError maps a failed WS handshake to a user-actionable message.
func handshakeError(err error, resp *http.Response) error {
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden, 459:
			return fmt.Errorf("server rejected exec (HTTP %d); please run: olares-cli profile login", resp.StatusCode)
		case http.StatusNotFound:
			return fmt.Errorf("exec endpoint not found (HTTP 404): the pod or container may not exist")
		default:
			if resp.StatusCode != http.StatusSwitchingProtocols {
				return fmt.Errorf("exec WebSocket not established (HTTP %d): this Olares version may not support `cluster exec`; please upgrade Olares", resp.StatusCode)
			}
		}
	}
	return fmt.Errorf("exec handshake failed: %w", err)
}

// RunOneShot runs the command with no TTY and stdin closed, accumulating
// stdout/stderr (capped per stream at maxBytes) and returning the exit code.
// A ctx deadline bounds the run; on expiry the partial output is returned with
// ExitCode == nil and ctx.Err() as the error.
func RunOneShot(ctx context.Context, rp *credential.ResolvedProfile, token string, o Options, maxBytes int) (Result, error) {
	o.Stdin = false
	o.TTY = false
	conn, resp, err := dial(ctx, rp.ControlHubURL, token, rp.InsecureSkipVerify, o)
	if err != nil {
		return Result{}, handshakeError(err, resp)
	}
	defer conn.Close()

	sink := NewSink(maxBytes)
	var exit *int

	// Close the connection when ctx fires so a blocked ReadMessage returns.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()

	for {
		_, msg, rerr := conn.ReadMessage()
		if rerr != nil {
			if ctx.Err() != nil {
				return Result{Stdout: sink.Stdout, Stderr: sink.Stderr, Truncated: sink.Truncated}, ctx.Err()
			}
			break // normal/abnormal close after the process finished
		}
		if len(msg) == 0 {
			continue
		}
		ch, payload := msg[0], msg[1:]
		if ch == ChannelError {
			if code, perr := ParseExitStatus(payload); perr == nil {
				exit = &code
			}
			continue
		}
		sink.Write(ch, payload)
	}
	return Result{Stdout: sink.Stdout, Stderr: sink.Stderr, ExitCode: exit, Truncated: sink.Truncated}, nil
}

// RunInteractive attaches a TTY: local stdin -> channel 0, server stdout/stderr
// -> stdout, SIGWINCH -> channel 4 resize. Returns the exit code (nil if none
// was reported).
func RunInteractive(ctx context.Context, rp *credential.ResolvedProfile, token string, o Options, stdin *os.File, stdout io.Writer) (*int, error) {
	o.Stdin = true
	o.TTY = true
	conn, resp, err := dial(ctx, rp.ControlHubURL, token, rp.InsecureSkipVerify, o)
	if err != nil {
		return nil, handshakeError(err, resp)
	}
	defer conn.Close()

	fd := int(stdin.Fd())
	if term.IsTerminal(fd) {
		if old, merr := term.MakeRaw(fd); merr == nil {
			defer func() { _ = term.Restore(fd, old) }()
		}
	}

	sendResize := func() {
		if !term.IsTerminal(fd) {
			return
		}
		w, h, gerr := term.GetSize(fd)
		if gerr != nil {
			return
		}
		if frame, ferr := ResizeFrame(uint16(w), uint16(h)); ferr == nil {
			_ = conn.WriteMessage(websocket.BinaryMessage, frame)
		}
	}
	sendResize()
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)
	go func() {
		for range winch {
			sendResize()
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := stdin.Read(buf)
			if n > 0 {
				_ = conn.WriteMessage(websocket.BinaryMessage, Frame(ChannelStdin, buf[:n]))
			}
			if rerr != nil {
				return
			}
		}
	}()

	var exit *int
	for {
		_, msg, rerr := conn.ReadMessage()
		if rerr != nil {
			break
		}
		if len(msg) == 0 {
			continue
		}
		ch, payload := msg[0], msg[1:]
		switch ch {
		case ChannelStdout, ChannelStderr:
			_, _ = stdout.Write(payload)
		case ChannelError:
			if code, perr := ParseExitStatus(payload); perr == nil {
				exit = &code
			}
		}
	}
	return exit, nil
}
```

- [ ] **Step 2: Tidy + build**

Run: `cd cli && go mod tidy && go build ./pkg/clusterexec/...`
Expected: builds clean; `gorilla/websocket` moves from `// indirect` to a direct require in `cli/go.mod`.

- [ ] **Step 3: Commit**

```bash
git add cli/pkg/clusterexec/client.go cli/go.mod cli/go.sum
git commit -m "feat(cli): add clusterexec websocket client (one-shot + interactive)"
```

---

## Task 4: `cluster pod exec` verb

**Files:**
- Create: `cli/cmd/ctl/cluster/pod/exec.go`
- Modify: `cli/cmd/ctl/cluster/pod/root.go` (register the verb)

- [ ] **Step 1: Write the command**

```go
// cli/cmd/ctl/cluster/pod/exec.go
package pod

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/pkg/clusterexec"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// ExecParams is the shared input for `cluster pod exec` and the
// `cluster container exec` alias.
type ExecParams struct {
	Namespace string
	Pod       string
	Container string
	Command   []string
	Stdin     bool
	TTY       bool
	AssumeYes bool
	Timeout   time.Duration
	MaxBytes  int
}

// execJSON is the one-shot `-o json` result shape (see spec "AI-friendly
// contract"). ExitCode is a pointer so a timeout serializes as null.
type execJSON struct {
	Namespace  string   `json:"namespace"`
	Pod        string   `json:"pod"`
	Container  string   `json:"container"`
	Command    []string `json:"command"`
	Stdout     string   `json:"stdout"`
	Stderr     string   `json:"stderr"`
	ExitCode   *int     `json:"exitCode"`
	Truncated  bool     `json:"truncated"`
	DurationMs int64    `json:"durationMs"`
}

// NewExecCommand: `olares-cli cluster pod exec <ns/pod | pod> [-c C] [-it] -- CMD`.
//
// One-shot by default (no TTY, stdin closed): runs CMD, captures separated
// stdout/stderr, propagates the container's exit code as the CLI exit code.
// -it allocates an interactive TTY (requires a local terminal; prompts y/N).
func NewExecCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		stdinFlag bool
		ttyFlag   bool
		assumeYes bool
		timeout   time.Duration
		maxBytes  int
	)
	cmd := &cobra.Command{
		Use:   "exec <ns/pod | pod> [-c CONTAINER] [-it] -- CMD [args...]",
		Short: "run a command inside a container (one-shot; -it for an interactive shell)",
		Long: `Run a command inside a container.

One-shot (default): everything after ` + "`--`" + ` is the argv run in the
container (no implicit shell). stdout/stderr are captured separately and the
container's exit code becomes this command's exit code. Bounded by --timeout
and --max-output-bytes so a hung/chatty command can't stall or flood callers.
Use ` + "`-- sh -c '...'`" + ` for pipes/redirects or multi-step repairs.

Interactive (-i -t / -it): allocate a TTY and attach your terminal, like
` + "`kubectl exec -it`" + `. Requires a local terminal and prompts for
confirmation (--yes skips). Default command is ` + "`sh`" + ` when none given.

NOTE: changes made inside a running container are ephemeral — a pod restart
reverts them. Durable fixes go through the image / ConfigMap / workload spec
(see ` + "`cluster workload`" + `).
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			if dash == -1 {
				if len(args) != 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--` (e.g. exec mypod -- ls)", args[1:])
				}
				target = args[0]
			} else {
				if dash < 1 {
					return fmt.Errorf("missing <pod> before `--`")
				}
				target = args[0]
				command = args[dash:]
			}
			ns, podName, err := clusteropts.SplitNsName(namespace, target)
			if err != nil {
				return err
			}
			return RunExec(c.Context(), o, ExecParams{
				Namespace: ns, Pod: podName, Container: container,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				AssumeYes: assumeYes, Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional is a bare pod name)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required for multi-container pods)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt for interactive (-it) exec")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}

// RunExec is the shared entry point used by `cluster pod exec` and the
// `cluster container exec` alias.
func RunExec(ctx context.Context, o *clusteropts.ClusterOptions, p ExecParams) error {
	if ctx == nil {
		ctx = context.Background()
	}
	f := o.Factory()

	// Resolve the container (auto-pick / validate) — also exercises the
	// refreshing transport so an expired token rotates before the handshake.
	container := strings.TrimSpace(p.Container)
	if container == "" {
		pod, err := Get(ctx, o, p.Namespace, p.Pod)
		if err != nil {
			return err
		}
		container, err = pickContainer(pod)
		if err != nil {
			return err
		}
	}

	token, err := f.ValidAccessToken(ctx)
	if err != nil {
		return err
	}
	rp, err := f.ResolveProfile(ctx)
	if err != nil {
		return err
	}

	opts := clusterexec.Options{
		Namespace: p.Namespace, Pod: p.Pod, Container: container,
		Command: p.Command, Stdin: p.Stdin, TTY: p.TTY,
	}

	if p.TTY {
		if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("-t/--tty requires an interactive terminal; for non-interactive use run one-shot (drop -it and pass `-- CMD`)")
		}
		if len(opts.Command) == 0 {
			opts.Command = []string{"sh"}
		}
		if !p.AssumeYes {
			ok, cerr := clusteropts.ConfirmDestructive(fmt.Sprintf(
				"Open an interactive shell in %s/%s [container %s]?", p.Namespace, p.Pod, container))
			if cerr != nil {
				return cerr
			}
			if !ok {
				return fmt.Errorf("aborted by user")
			}
		}
		exit, rerr := clusterexec.RunInteractive(ctx, rp, token, opts, os.Stdin, os.Stdout)
		if rerr != nil {
			return rerr
		}
		if exit != nil && *exit != 0 {
			os.Exit(*exit)
		}
		return nil
	}

	// One-shot.
	if len(opts.Command) == 0 {
		return fmt.Errorf("no command given; pass `-- CMD [args...]` (or use -it for an interactive shell)")
	}
	if p.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.Timeout)
		defer cancel()
	}
	start := time.Now()
	res, rerr := clusterexec.RunOneShot(ctx, rp, token, opts, p.MaxBytes)
	dur := time.Since(start)
	if rerr != nil {
		if errors.Is(rerr, context.DeadlineExceeded) {
			renderOneShot(o, p, container, res, nil, dur, true)
			return fmt.Errorf("command timed out after %s", p.Timeout)
		}
		return rerr
	}
	renderOneShot(o, p, container, res, res.ExitCode, dur, false)
	if res.ExitCode != nil && *res.ExitCode != 0 {
		os.Exit(*res.ExitCode)
	}
	return nil
}

func renderOneShot(o *clusteropts.ClusterOptions, p ExecParams, container string, res clusterexec.Result, exit *int, dur time.Duration, timedOut bool) {
	if o.IsJSON() {
		_ = o.PrintJSON(execJSON{
			Namespace: p.Namespace, Pod: p.Pod, Container: container,
			Command: p.Command, Stdout: string(res.Stdout), Stderr: string(res.Stderr),
			ExitCode: exit, Truncated: res.Truncated, DurationMs: dur.Milliseconds(),
		})
		return
	}
	if o.Quiet {
		return
	}
	_, _ = os.Stdout.Write(res.Stdout)
	_, _ = os.Stderr.Write(res.Stderr)
	if res.Truncated {
		fmt.Fprintln(os.Stderr, "[output truncated: --max-output-bytes reached]")
	}
	if timedOut {
		fmt.Fprintln(os.Stderr, "[timed out]")
	}
}
```

- [ ] **Step 2: Register the verb in `pod/root.go`**

Modify `cli/cmd/ctl/cluster/pod/root.go` — add after the `NewRestartCommand` registration (line 57):

```go
	cmd.AddCommand(NewExecCommand(f))
```

- [ ] **Step 3: Build**

Run: `cd cli && go build ./...`
Expected: builds clean.

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/ctl/cluster/pod/exec.go cli/cmd/ctl/cluster/pod/root.go
git commit -m "feat(cli): add cluster pod exec verb"
```

---

## Task 5: `cluster container exec` alias

**Files:**
- Create: `cli/cmd/ctl/cluster/container/exec.go`
- Modify: `cli/cmd/ctl/cluster/container/root.go` (register the verb)

- [ ] **Step 1: Write the alias**

```go
// cli/cmd/ctl/cluster/container/exec.go
package container

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/beclab/Olares/cli/cmd/ctl/cluster/internal/clusteropts"
	"github.com/beclab/Olares/cli/cmd/ctl/cluster/pod"
	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// NewExecCommand: `olares-cli cluster container exec
// <ns/pod/container | ns/pod | pod> [-n NS] [-c NAME] [-it] -- CMD [args...]`.
//
// Thin alias over `cluster pod exec` — same wire, same semantics. The only
// difference is the positional grammar: container may be supplied as the third
// path segment. Delegates to pod.RunExec.
func NewExecCommand(f *cmdutil.Factory) *cobra.Command {
	o := clusteropts.NewClusterOptions(f)
	var (
		namespace string
		container string
		stdinFlag bool
		ttyFlag   bool
		assumeYes bool
		timeout   time.Duration
		maxBytes  int
	)
	cmd := &cobra.Command{
		Use:   "exec <ns/pod/container | ns/pod | pod> [-c NAME] [-it] -- CMD [args...]",
		Short: "run a command inside a container (one-shot; -it for an interactive shell)",
		Long: `Run a command inside a container (alias of ` + "`cluster pod exec`" + `).

Identity grammar adds a three-segment positional <ns>/<pod>/<container>; the
two-segment <ns>/<pod> + --container and bare <pod> + -n/-c forms also work.
Everything else (one-shot vs -it, --timeout, --max-output-bytes, -o json) is
identical to ` + "`cluster pod exec`" + `; this verb just delegates.
`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			dash := c.ArgsLenAtDash()
			var target string
			var command []string
			if dash == -1 {
				if len(args) != 1 {
					return fmt.Errorf("unexpected args %q; put the command after `--`", args[1:])
				}
				target = args[0]
			} else {
				if dash < 1 {
					return fmt.Errorf("missing <pod> before `--`")
				}
				target = args[0]
				command = args[dash:]
			}
			ns, podName, ctr, err := splitNsPodContainer(namespace, container, target)
			if err != nil {
				return err
			}
			return pod.RunExec(c.Context(), o, pod.ExecParams{
				Namespace: ns, Pod: podName, Container: ctr,
				Command: command, Stdin: stdinFlag, TTY: ttyFlag,
				AssumeYes: assumeYes, Timeout: timeout, MaxBytes: maxBytes,
			})
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "namespace (required when the positional doesn't include one)")
	cmd.Flags().StringVarP(&container, "container", "c", "", "container name (required when the positional doesn't include one)")
	cmd.Flags().BoolVarP(&stdinFlag, "stdin", "i", false, "keep stdin open to the container")
	cmd.Flags().BoolVarP(&ttyFlag, "tty", "t", false, "allocate a TTY (interactive); requires a local terminal")
	cmd.Flags().BoolVarP(&assumeYes, "yes", "y", false, "skip the confirmation prompt for interactive (-it) exec")
	cmd.Flags().DurationVar(&timeout, "timeout", 60*time.Second, "one-shot only: abort if the command runs longer (0 = no limit)")
	cmd.Flags().IntVar(&maxBytes, "max-output-bytes", 2<<20, "one-shot only: cap per-stream captured output in bytes (0 = unlimited)")
	o.AddDetailOutputFlags(cmd)
	return cmd
}
```

(Note: `splitNsPodContainer` already exists in `cli/cmd/ctl/cluster/container/logs.go` — reuse it; do not redefine.)

- [ ] **Step 2: Register the verb in `container/root.go`**

Modify `cli/cmd/ctl/cluster/container/root.go` — add after the `NewLogsCommand` registration (line 53):

```go
	cmd.AddCommand(NewExecCommand(f))
```

- [ ] **Step 3: Build**

Run: `cd cli && go build ./...`
Expected: builds clean.

- [ ] **Step 4: Commit**

```bash
git add cli/cmd/ctl/cluster/container/exec.go cli/cmd/ctl/cluster/container/root.go
git commit -m "feat(cli): add cluster container exec alias"
```

---

## Task 6: Edge nginx — allow the exec WebSocket upgrade

**Files:**
- Modify: `apps/docker/system-frontend/nginx/dashboard-control-hub.conf`

- [ ] **Step 1: Extend BOTH WS-upgrade location regexes**

There are two `location ~ /(kapis/terminal|api/v1/watch|apis/apps/v1/watch)`
blocks — one in the `listen 81` server (line ~116) and one in the `listen 82`
server (line ~222). In **both**, change the regex to add the exec path:

```nginx
  location ~ /(kapis/terminal|api/v1/watch|apis/apps/v1/watch|api/v1/namespaces/[^/]+/pods/[^/]+/exec) {
    proxy_pass http://SettingsServer;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Host $http_host;
  }
```

(Only the first line — the regex — changes; the body stays identical.)

- [ ] **Step 2: Sanity-check the regex**

Run: `grep -n "pods/\[\^/\]+/exec" apps/docker/system-frontend/nginx/dashboard-control-hub.conf`
Expected: two matches (one per server block).

- [ ] **Step 3: Commit**

```bash
git add apps/docker/system-frontend/nginx/dashboard-control-hub.conf
git commit -m "feat(control-hub): allow native pod exec WebSocket upgrade at the edge"
```

---

## Task 7: Skill documentation

**Files:**
- Modify: `cli/skills/olares-cluster/SKILL.md`

- [ ] **Step 1: Add `exec` to the container verb-index row**

Find the `container` row in the "Verb index" table:

```
| `container` (alias `ctr`) | `list`, `env`, `logs` | `olares-cli cluster container --help` |
```

Change it to:

```
| `container` (alias `ctr`) | `list`, `env`, `logs`, `exec` | `olares-cli cluster container --help` |
```

- [ ] **Step 2: Add `exec` to the pod verb-index row**

Find the `pod` row and append `exec`:

```
| `pod` | `list`, `get`, `yaml`, `events`, `logs`, `delete`, `restart`, `exec` | [references/olares-cluster-pod.md](references/olares-cluster-pod.md) |
```

- [ ] **Step 3: Add an exec subsection** after the "Common errors → fixes" section:

```markdown
## exec (run commands inside a container)

`cluster {pod,container} exec` runs a command inside a container over the
native K8s exec WebSocket. Two modes:

- **One-shot (default, best for agents):**
  `olares-cli cluster container exec <ns>/<pod>/<ctr> -o json -- cat /etc/hosts`
  Returns `{stdout, stderr, exitCode, truncated, durationMs}`. Judge success by
  `exitCode` (0 = ok). stdout/stderr are separated. Bounded by `--timeout`
  (default 60s) and `--max-output-bytes` (default 2MiB).
- **Interactive (`-it`, for humans):** allocates a TTY, prompts y/N, needs a
  real terminal. Agents should stay on one-shot.

Agent guidance:

- Pass argv after `--` (no implicit shell). For pipes/vars/multi-step, use
  `-- sh -c '...'` — exec is stateless, so chain steps in one call rather than
  expecting `cd`/exports to persist.
- Edit files with `-i -- sh -c 'cat > /path'` (content on stdin), or
  `-- sh -c "sed -i 's/old/new/' /path"`, or `-- sh -c 'printf "%s" ... | tee /path'`.
- **Fixes are ephemeral:** changes inside a running container revert on pod
  restart/recreate. For a durable fix, change the image / ConfigMap / Deployment
  spec via the `workload` path — do not report an in-container change as permanent.
- Common exit codes: `127` command not found, `126` not executable. A "no sh in
  container" failure means a distroless/scratch image (no shell). `EROFS`/`EACCES`
  on writes means a read-only/permission-restricted filesystem.
- exec requires `pods/exec` permission (server-side SAR; 403 if missing) and is
  audited server-side by ks-apiserver.
- Requires a recent Olares; on older versions the handshake fails with
  "this Olares version may not support `cluster exec`".
```

- [ ] **Step 4: Commit**

```bash
git add cli/skills/olares-cluster/SKILL.md
git commit -m "docs(cli): document cluster exec in olares-cluster skill"
```

---

## Task 8: Full build, vet, test, and manual e2e

**Files:** none (verification only)

- [ ] **Step 1: Build + vet + unit tests**

Run: `cd cli && go build ./... && go vet ./pkg/clusterexec/... ./cmd/ctl/cluster/... && go test ./pkg/clusterexec/...`
Expected: all pass.

- [ ] **Step 2: Manual e2e against a live Olares** (requires a logged-in profile and the rebuilt `system-frontend` image with the Task 6 nginx change deployed)

Pick a running pod (`olares-cli cluster pod list`) and verify:

```bash
# one-shot, table
olares-cli cluster container exec <ns>/<pod>/<ctr> -- sh -c 'echo hi; echo err 1>&2'
# expect: hi on stdout, err on stderr, exit 0

# one-shot, json
olares-cli cluster container exec <ns>/<pod>/<ctr> -o json -- cat /etc/hostname
# expect: {"...","stdout":"<host>\n","stderr":"","exitCode":0,...}

# non-zero exit propagation
olares-cli cluster container exec <ns>/<pod>/<ctr> -- sh -c 'exit 3'; echo "rc=$?"
# expect: rc=3

# missing command
olares-cli cluster container exec <ns>/<pod>/<ctr> -- /no/such/bin; echo "rc=$?"
# expect: rc=127

# timeout
olares-cli cluster container exec <ns>/<pod>/<ctr> --timeout 2s -- sh -c 'sleep 10'
# expect: "[timed out]" + non-zero exit

# interactive (human terminal)
olares-cli cluster container exec <ns>/<pod>/<ctr> -it
# expect: y/N prompt, then a shell; `exit` returns
```

- [ ] **Step 3: Verify old-version degradation** (optional): against an Olares
without the nginx change, one-shot should fail with the
"this Olares version may not support `cluster exec`" message rather than hang.

---

## Self-review notes

- Spec sections mapped to tasks: wire/framing → T1; token freshness → T2; WS
  client + handshake errors + ephemeral-safe one-shot → T3; CLI surface, JSON
  contract, exit-code propagation, `-it` confirm, timeout/cap → T4; container
  alias → T5; nginx WS whitelist → T6; skill (verb index, ephemeral/durable,
  recipes, audit, errors) → T7; verification → T8.
- Type consistency: `clusterexec.Options`, `clusterexec.Result`,
  `pod.ExecParams`, `pod.execJSON`, `clusterexec.{Frame,ResizeFrame,Sink,
  ParseExitStatus,RunOneShot,RunInteractive}`, `Factory.ValidAccessToken` are
  used identically across tasks.
- `splitNsPodContainer` and `pickContainer` are reused from existing
  container/pod packages — not redefined.
```
