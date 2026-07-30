package appstate

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsRetryableWebhookError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "generic", err: errors.New("something else"), want: false},
		{name: "not found", err: errors.New(`deployments.apps "foo" not found`), want: false},
		{
			name: "dial without webhook context",
			err:  errors.New(`dial tcp 10.233.24.114:8080: connect: connection refused`),
			want: false,
		},
		{
			name: "invalid argument alone is not retryable",
			err:  errors.New(`dial tcp 10.233.98.192:8433: connect: invalid argument`),
			want: false,
		},
		{
			name: "failed calling webhook",
			err:  errors.New(`Internal error occurred: failed calling webhook "gpu-limit-inject-webhook.bytetrade.io"`),
			want: true,
		},
		{
			name: "502 bad gateway",
			err:  errors.New(`proxy error from 127.0.0.1:6443 while dialing 10.233.98.192:8433, code 502: 502 Bad Gateway`),
			want: true,
		},
		{
			name: "connection refused on webhook port",
			err:  errors.New(`failed calling webhook: dial tcp 10.233.98.55:8433: connect: connection refused`),
			want: true,
		},
		{
			name: "no endpoints available",
			err:  errors.New(`failed calling webhook: no endpoints available for service "app-service"`),
			want: true,
		},
		{
			name: "no endpoints",
			err:  errors.New(`failed calling webhook: no endpoints for service "app-service"`),
			want: true,
		},
		{
			name: "wrapped",
			err:  fmt.Errorf("suspend app studio failed %w", errors.New(`failed calling webhook "gpu-limit-inject-webhook.bytetrade.io"`)),
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRetryableWebhookError(tc.err); got != tc.want {
				t.Fatalf("IsRetryableWebhookError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
