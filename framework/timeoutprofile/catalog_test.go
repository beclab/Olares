package timeoutprofile

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestDefaultIsValid(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("default catalog is invalid:\n%v", err)
	}
}

// TestDefaultValuesPinned makes every default change an explicit, reviewed
// edit to this table.
func TestDefaultValuesPinned(t *testing.T) {
	c := Default()
	cases := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"edge route timeout", c.Edge.RouteTimeout, 0},
		{"edge request headers", c.Edge.RequestHeaders, 10 * time.Second},
		{"edge stream idle fallback", c.Edge.StreamIdleFallback, 300 * time.Second},
		{"edge max stream duration", c.Edge.MaxStreamDuration, 24 * time.Hour},
		{"edge connect", c.Edge.Connect, 5 * time.Second},
		{"edge files idle", c.Edge.FilesIdle, 1800 * time.Second},
		{"edge auth timeout", c.Edge.AuthTimeout, 15 * time.Second},
		{"interactive edge idle", c.Classes[Interactive].EdgeIdle, 600 * time.Second},
		{"streaming edge idle", c.Classes[Streaming].EdgeIdle, 600 * time.Second},
		{"websocket edge idle", c.Classes[WebSocket].EdgeIdle, 3600 * time.Second},
		{"transfer edge idle", c.Classes[Transfer].EdgeIdle, 300 * time.Second},
		{"go server read header", c.GoServer.ReadHeader, 10 * time.Second},
		{"go server idle", c.GoServer.Idle, 120 * time.Second},
		{"fasthttp server read", c.FastHTTPServer.Read, 30 * time.Second},
		{"fasthttp server idle", c.FastHTTPServer.Idle, 120 * time.Second},
		{"intranet proxy read header", c.IntranetProxy.ReadHeader, 10 * time.Second},
		{"intranet proxy idle", c.IntranetProxy.Idle, 120 * time.Second},
		{"intranet proxy upstream dial", c.IntranetProxy.UpstreamDial, 5 * time.Second},
		{"inner proxy connect", c.InnerProxy.Connect, 5 * time.Second},
		{"bounds idle min", c.Bounds.IdleMin, 30 * time.Second},
		{"bounds idle max", c.Bounds.IdleMax, 3600 * time.Second},
		{"bounds total min", c.Bounds.TotalMin, 30 * time.Second},
		{"bounds total max", c.Bounds.TotalMax, 24 * time.Hour},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %s, want %s", tc.name, tc.got, tc.want)
		}
	}
	if c.Edge.AuthFailureStatus != 503 {
		t.Errorf("edge auth failure status = %d, want 503", c.Edge.AuthFailureStatus)
	}
	if c.Edge.ConnectionBufferBytes != 32*1024 {
		t.Errorf("edge connection buffer = %d, want %d", c.Edge.ConnectionBufferBytes, 32*1024)
	}
	if c.Version != CatalogVersion {
		t.Errorf("version = %q, want %q", c.Version, CatalogVersion)
	}
}

func TestDefaultReturnsIndependentCopies(t *testing.T) {
	a := Default()
	a.Classes[Interactive] = ClassProfile{EdgeIdle: time.Second, InnerIdleMax: time.Second}
	if got := Default().Classes[Interactive].EdgeIdle; got != 600*time.Second {
		t.Fatalf("mutating one copy changed another: interactive edge idle = %s", got)
	}
}

func TestValidateRejectsInversions(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*Catalog)
		wantErr string
	}{
		{
			name:    "edge total deadline re-enabled",
			mutate:  func(c *Catalog) { c.Edge.RouteTimeout = 5 * time.Minute },
			wantErr: "edge route timeout must be 0",
		},
		{
			name: "inner idle longer than edge idle",
			mutate: func(c *Catalog) {
				c.Classes[Interactive] = ClassProfile{EdgeIdle: 600 * time.Second, InnerIdleMax: 1800 * time.Second}
			},
			wantErr: `class "interactive" inner idle`,
		},
		{
			name: "class idle outside bounds",
			mutate: func(c *Catalog) {
				c.Classes[WebSocket] = ClassProfile{EdgeIdle: 2 * time.Hour, InnerIdleMax: time.Hour}
			},
			wantErr: `class "websocket" edge idle 2h0m0s outside`,
		},
		{
			name:    "class missing",
			mutate:  func(c *Catalog) { delete(c.Classes, Streaming) },
			wantErr: `class "streaming" missing`,
		},
		{
			name:    "total bound beyond hard stream limit",
			mutate:  func(c *Catalog) { c.Bounds.TotalMax = 48 * time.Hour },
			wantErr: "exceeds edge max stream duration",
		},
		{
			name:    "auth failure reported as client error",
			mutate:  func(c *Catalog) { c.Edge.AuthFailureStatus = 403 },
			wantErr: "auth failure status must be a 5xx",
		},
		{
			name:    "stream idle fallback above interactive idle",
			mutate:  func(c *Catalog) { c.Edge.StreamIdleFallback = time.Hour },
			wantErr: "stream idle fallback",
		},
		{
			name:    "files idle below transfer idle",
			mutate:  func(c *Catalog) { c.Edge.FilesIdle = time.Minute },
			wantErr: "edge files idle",
		},
		{
			name:    "intranet dial longer than edge connect",
			mutate:  func(c *Catalog) { c.IntranetProxy.UpstreamDial = 30 * time.Second },
			wantErr: "intranet proxy upstream dial",
		},
		{
			name:    "request headers timeout not shorter than idle",
			mutate:  func(c *Catalog) { c.Edge.RequestHeaders = 10 * time.Minute },
			wantErr: "must exceed request headers timeout",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Default()
			tc.mutate(&c)
			err := c.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestProfileFallsBackToInteractive(t *testing.T) {
	c := Default()
	for _, class := range []Class{"", "bulk"} {
		if got, want := c.Profile(class), c.Classes[Interactive]; got != want {
			t.Errorf("Profile(%q) = %+v, want interactive %+v", class, got, want)
		}
	}
	if got := c.Profile(WebSocket).EdgeIdle; got != 3600*time.Second {
		t.Errorf("Profile(websocket).EdgeIdle = %s, want 1h", got)
	}
}

func TestResolve(t *testing.T) {
	c := Default()
	cases := []struct {
		name     string
		class    Class
		override Override
		want     Effective
		wantErr  error
	}{
		{
			name:  "empty class uses interactive defaults",
			class: "",
			want:  Effective{Class: Interactive, Idle: 600 * time.Second, IdleSource: FromDefault, TotalSource: FromDefault},
		},
		{
			name:  "websocket default",
			class: WebSocket,
			want:  Effective{Class: WebSocket, Idle: 3600 * time.Second, IdleSource: FromDefault, TotalSource: FromDefault},
		},
		{
			name:     "idle and total overrides inside bounds",
			class:    Streaming,
			override: Override{Idle: 900 * time.Second, Total: time.Hour},
			want: Effective{
				Class: Streaming, Idle: 900 * time.Second, IdleSource: FromOverride,
				Total: time.Hour, TotalSource: FromOverride,
			},
		},
		{
			name:     "bounds are inclusive",
			class:    Interactive,
			override: Override{Idle: 30 * time.Second, Total: 24 * time.Hour},
			want: Effective{
				Class: Interactive, Idle: 30 * time.Second, IdleSource: FromOverride,
				Total: 24 * time.Hour, TotalSource: FromOverride,
			},
		},
		{name: "unknown class", class: "bulk", wantErr: ErrUnknownClass},
		{name: "idle below min", override: Override{Idle: 10 * time.Second}, wantErr: ErrOutOfBounds},
		{name: "idle above max", override: Override{Idle: 2 * time.Hour}, wantErr: ErrOutOfBounds},
		{name: "negative idle", override: Override{Idle: -time.Second}, wantErr: ErrOutOfBounds},
		{name: "total below min", override: Override{Total: time.Second}, wantErr: ErrOutOfBounds},
		{name: "total above max", override: Override{Total: 25 * time.Hour}, wantErr: ErrOutOfBounds},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.Resolve(tc.class, tc.override)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Resolve() error = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("Resolve() = %+v, want %+v", got, tc.want)
			}
		})
	}
}
