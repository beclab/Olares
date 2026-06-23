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
	"sync"
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

	// Set the terminal title to the target so it stays visible in the tab even
	// after the screen scrolls; restore it on exit. Terminals that don't grok
	// OSC sequences silently ignore these bytes.
	fmt.Fprintf(stdout, "\033]0;%s/%s/%s\a", o.Namespace, o.Pod, o.Container)
	defer fmt.Fprint(stdout, "\033]0;\a")

	fd := int(stdin.Fd())
	if term.IsTerminal(fd) {
		if old, merr := term.MakeRaw(fd); merr == nil {
			defer func() { _ = term.Restore(fd, old) }()
		}
	}

	// gorilla *websocket.Conn is not safe for concurrent writers; the resize
	// goroutine, the stdin pump, and the initial sendResize all write, so
	// serialize every outbound write through this helper.
	var writeMu sync.Mutex
	writeMsg := func(data []byte) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteMessage(websocket.BinaryMessage, data)
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
			_ = writeMsg(frame)
		}
	}
	sendResize()
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-winch:
				sendResize()
			case <-done:
				return
			}
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, rerr := stdin.Read(buf)
			if n > 0 {
				_ = writeMsg(Frame(ChannelStdin, buf[:n]))
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
