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
