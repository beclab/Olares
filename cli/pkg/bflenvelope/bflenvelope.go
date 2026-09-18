// Package bflenvelope decodes the response envelope every `olares-cli
// settings` route replies with: {code, message, data}.
//
// It exists because `message` is not always a string. user-service
// re-wraps an upstream failure verbatim, so an app-service 400 arrives
// nested inside a 500:
//
//	{"code":500,"message":{"code":400,"message":"value not in options"},"data":null}
//
// Decoding that field as a string fails the whole unmarshal, and the
// caller gets a Go type error instead of the reason the write was
// rejected.
package bflenvelope

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Envelope is the outer response shape. Data stays raw so callers decode
// only the payload they expect.
type Envelope struct {
	Code    int             `json:"code"`
	Message Message         `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Message is an error text that may arrive as a string or as a nested
// envelope. Unmarshalling never fails: an unrecognized shape is kept as
// its raw JSON, which is more useful to a reader than a decode error.
type Message struct {
	Text string
}

func (m *Message) UnmarshalJSON(raw []byte) error {
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		m.Text = text
		return nil
	}
	var nested struct {
		Message json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(raw, &nested); err == nil && len(nested.Message) > 0 {
		var inner Message
		if inner.UnmarshalJSON(nested.Message) == nil && inner.Text != "" {
			m.Text = inner.Text
			return nil
		}
	}
	m.Text = strings.TrimSpace(string(raw))
	return nil
}

// Data applies the envelope contract: a code outside {0, 200} is an
// error, otherwise the payload is decoded into out. Pass a nil out for
// calls with no useful response body.
//
// softCodes are extra codes to accept, for routes where a non-success
// code is still informative rather than fatal (e.g. search's "no more
// results").
func Data(method, path string, env Envelope, out interface{}, softCodes ...int) error {
	if !accepts(env.Code, softCodes) {
		if msg := strings.TrimSpace(env.Message.Text); msg != "" {
			return fmt.Errorf("%s %s: upstream returned code %d: %s", method, path, env.Code, msg)
		}
		return fmt.Errorf("%s %s: upstream returned code %d", method, path, env.Code)
	}
	if out == nil || len(env.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(env.Data, out); err != nil {
		return fmt.Errorf("%s %s: decode data: %w", method, path, err)
	}
	return nil
}

func accepts(code int, softCodes []int) bool {
	if code == 0 || code == 200 {
		return true
	}
	for _, soft := range softCodes {
		if code == soft {
			return true
		}
	}
	return false
}
