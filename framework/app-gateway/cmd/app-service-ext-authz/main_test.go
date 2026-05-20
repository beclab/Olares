package main

import (
	"context"
	"log/slog"
	"os"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	authv3attr "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/grpc/codes"
)

func newCheckReq(host, path, method string, headers map[string]string) *authv3.CheckRequest {
	return &authv3attr.CheckRequest{
		Attributes: &authv3.AttributeContext{
			Request: &authv3.AttributeContext_Request{
				Http: &authv3.AttributeContext_HttpRequest{
					Host:    host,
					Path:    path,
					Method:  method,
					Headers: headers,
				},
			},
		},
	}
}

func TestAuthzServer_AllowMode(t *testing.T) {
	s := &authzServer{allow: true, logger: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
	resp, err := s.Check(context.Background(), newCheckReq("a.shared.example.com", "/api/tags", "GET", map[string]string{"x-request-id": "rid-1"}))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if resp.GetStatus().GetCode() != int32(codes.OK) {
		t.Fatalf("status code = %d, want OK", resp.GetStatus().GetCode())
	}
	if resp.GetOkResponse() == nil {
		t.Fatalf("expected OkResponse, got %T", resp.GetHttpResponse())
	}
}

func TestAuthzServer_DenyMode(t *testing.T) {
	s := &authzServer{allow: false, logger: slog.New(slog.NewJSONHandler(os.Stderr, nil))}
	resp, err := s.Check(context.Background(), newCheckReq("a.shared.example.com", "/", "GET", nil))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if resp.GetStatus().GetCode() != int32(codes.PermissionDenied) {
		t.Fatalf("status code = %d, want PermissionDenied", resp.GetStatus().GetCode())
	}
	if resp.GetDeniedResponse() == nil {
		t.Fatalf("expected DeniedResponse, got %T", resp.GetHttpResponse())
	}
}

func TestExtractRequestID(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]string
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "empty", in: map[string]string{}, want: ""},
		{name: "x-request-id", in: map[string]string{"x-request-id": "rid"}, want: "rid"},
		{name: "fallback traceparent", in: map[string]string{"traceparent": "00-...-01"}, want: "00-...-01"},
		{name: "ignore unrelated", in: map[string]string{"user-agent": "curl"}, want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractRequestID(tc.in); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
