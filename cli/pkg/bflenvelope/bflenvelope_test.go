package bflenvelope

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDecodeMessageShapes(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "plain_string",
			body: `{"code":400,"message":"bad request","data":null}`,
			want: "bad request",
		},
		{
			// user-service re-wraps an app-service failure verbatim, so
			// the real reason arrives nested inside an outer 500.
			name: "nested_envelope",
			body: `{"code":500,"message":{"code":400,"message":"value not in options"},"data":null}`,
			want: "value not in options",
		},
		{
			name: "doubly_nested",
			body: `{"code":500,"message":{"code":500,"message":{"code":400,"message":"deep"}},"data":null}`,
			want: "deep",
		},
		{
			name: "empty_string",
			body: `{"code":400,"message":"","data":null}`,
			want: "",
		},
		{
			// An unrecognized shape is kept verbatim: a reader can act on
			// the raw JSON, but not on a decode error.
			name: "unexpected_shape_kept_raw",
			body: `{"code":400,"message":[1,2],"data":null}`,
			want: "[1,2]",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var env Envelope
			if err := json.Unmarshal([]byte(tc.body), &env); err != nil {
				t.Fatalf("unmarshal must not fail on any message shape: %v", err)
			}
			if env.Message.Text != tc.want {
				t.Errorf("message = %q; want %q", env.Message.Text, tc.want)
			}
		})
	}
}

func TestDataSurfacesTheNestedReason(t *testing.T) {
	var env Envelope
	body := `{"code":500,"message":{"code":400,"message":"value not in options"},"data":null}`
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	err := Data("PUT", "/api/env/userenvs", env, nil)
	if err == nil {
		t.Fatal("Data returned nil on a non-2xx envelope code")
	}
	for _, sub := range []string{"PUT /api/env/userenvs", "500", "value not in options"} {
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("error %q does not contain %q", err, sub)
		}
	}
}

func TestDataAcceptsBothSuccessCodes(t *testing.T) {
	for _, code := range []int{0, 200} {
		var out struct {
			Name string `json:"name"`
		}
		env := Envelope{Code: code, Data: json.RawMessage(`{"name":"ok"}`)}
		if err := Data("GET", "/api/thing", env, &out); err != nil {
			t.Fatalf("code %d treated as failure: %v", code, err)
		}
		if out.Name != "ok" {
			t.Errorf("code %d: payload = %q; want %q", code, out.Name, "ok")
		}
	}
}

func TestDataIgnoresPayloadWhenCallerWantsNone(t *testing.T) {
	env := Envelope{Code: 200, Data: json.RawMessage(`{"anything":true}`)}
	if err := Data("POST", "/api/thing", env, nil); err != nil {
		t.Errorf("Data with a nil out errored: %v", err)
	}
}

func TestDataReportsAnUndecodablePayload(t *testing.T) {
	var out struct {
		N int `json:"n"`
	}
	env := Envelope{Code: 200, Data: json.RawMessage(`{"n":"not-a-number"}`)}
	err := Data("GET", "/api/thing", env, &out)
	if err == nil || !strings.Contains(err.Error(), "decode data") {
		t.Errorf("error = %v; want it to name the payload decode failure", err)
	}
}
