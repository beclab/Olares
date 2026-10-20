package controllers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	v1alpha1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestApplicationPortsFromApps(t *testing.T) {
	acl := func(proto string, dst ...string) v1alpha1.ACL {
		return v1alpha1.ACL{Proto: proto, Dst: dst}
	}
	appWith := func(owner string, tailscale []v1alpha1.ACL, legacy []v1alpha1.ACL) v1alpha1.Application {
		return v1alpha1.Application{
			Spec: v1alpha1.ApplicationSpec{
				Owner:         owner,
				TailScale:     v1alpha1.TailScale{ACLs: tailscale},
				TailScaleACLs: legacy,
			},
		}
	}

	tests := []struct {
		name string
		apps []v1alpha1.Application
		want []applicationPort
	}{
		{
			name: "star host port becomes the port number",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("tcp", "*:443")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"443"}, UDP: []string{}}},
		},
		{
			name: "closed range and udp",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("udp", "*:6000-6010")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{}, UDP: []string{"6000-6010"}}},
		},
		{
			name: "comma separated ports",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("tcp", "tag:montreal-webserver:80,443")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"443", "80"}, UDP: []string{}}},
		},
		{
			name: "wildcard port is dropped",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("tcp", "git-server:*", "*:22")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"22"}, UDP: []string{}}},
		},
		{
			name: "invalid ports are dropped",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("tcp", "*:0", "*:65536", "*:7000-6000", "*:65535")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"65535"}, UDP: []string{}}},
		},
		{
			name: "empty proto opens tcp and udp",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("", "*:47945")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"47945"}, UDP: []string{"47945"}}},
		},
		{
			name: "non tcp udp is skipped",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("icmp", "*:1"), acl("TCP", "*:22")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"22"}, UDP: []string{}}},
		},
		{
			name: "same user merges apps and legacy acls",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("tcp", "*:22")}, []v1alpha1.ACL{acl("udp", "*:5353")}),
				appWith("alice", []v1alpha1.ACL{acl("tcp", "*:22"), acl("tcp", "192.168.1.0/24:80")}, nil),
			},
			want: []applicationPort{{User: "alice", TCP: []string{"22", "80"}, UDP: []string{"5353"}}},
		},
		{
			name: "owners are trimmed and sorted",
			apps: []v1alpha1.Application{
				appWith(" bob@ ", []v1alpha1.ACL{acl("udp", "*:53")}, nil),
				appWith("alice", []v1alpha1.ACL{acl("tcp", "*:443")}, nil),
			},
			want: []applicationPort{
				{User: "alice", TCP: []string{"443"}, UDP: []string{}},
				{User: "bob", TCP: []string{}, UDP: []string{"53"}},
			},
		},
		{
			name: "user with no usable ports is omitted",
			apps: []v1alpha1.Application{
				appWith("alice", []v1alpha1.ACL{acl("icmp", "*:*")}, nil),
				appWith("  ", nil, nil),
			},
			want: []applicationPort{},
		},
		{
			name: "no apps",
			apps: nil,
			want: []applicationPort{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applicationPortsFromApps(tt.apps)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestApplicationPortsJSONEmptyArray(t *testing.T) {
	body := struct {
		ApplicationPorts []applicationPort `json:"applicationPorts"`
	}{ApplicationPorts: applicationPortsFromApps(nil)}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	require.JSONEq(t, `{"applicationPorts":[]}`, string(raw))
}

func TestApplicationPortsEqualIgnoresOrder(t *testing.T) {
	a := []applicationPort{{User: "bob", TCP: []string{"22"}, UDP: []string{"53"}}}
	b := []applicationPort{{User: " bob@ ", UDP: []string{"53", "53"}, TCP: []string{"22"}}}
	require.True(t, applicationPortsEqual(a, b))
	require.False(t, applicationPortsEqual(a, []applicationPort{
		{User: "bob", TCP: []string{"80"}, UDP: []string{}},
	}))
}

func TestHeadscalePolicyClient(t *testing.T) {
	var mu sync.Mutex
	var calls int
	var putBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		mu.Unlock()
		if r.Header.Get("Authorization") != "" {
			http.Error(w, "unexpected authorization", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			fmtResponse(w, `{"code":0,"message":"","data":{"applicationPorts":[{"user":"alice","tcp":["443"],"udp":[]}]}}`)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			mu.Lock()
			putBody = body
			mu.Unlock()
			fmtResponse(w, `{"code":0,"message":"","data":{}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	c := &headscalePolicyClient{
		url:  srv.URL + "/internal/policy/application-ports",
		http: resty.New(),
	}

	got, err := c.Get(context.Background())
	require.NoError(t, err)
	require.Equal(t, []applicationPort{{User: "alice", TCP: []string{"443"}, UDP: []string{}}}, got)

	_, err = c.Get(context.Background())
	require.NoError(t, err)

	err = c.Put(context.Background(), []applicationPort{})
	require.NoError(t, err)
	require.JSONEq(t, `{"applicationPorts":[]}`, string(putBody))

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 3, calls)
}

func TestHeadscalePolicyClientRejectsFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") == "status" {
			http.Error(w, "nope", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmtResponse(w, `{"code":1,"message":"bad ports","data":{}}`)
	}))
	t.Cleanup(srv.Close)

	c := &headscalePolicyClient{
		url:  srv.URL + "?mode=code",
		http: resty.New(),
	}
	_, err := c.Get(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "code 1")

	c.url = srv.URL + "?mode=status"
	err = c.Put(context.Background(), []applicationPort{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "400")
}

func fmtResponse(w http.ResponseWriter, body string) {
	_, _ = io.WriteString(w, body)
}
