package timeout

import (
	"regexp"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/timeoutprofile"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestParseSeconds(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    time.Duration
		wantErr bool
	}{
		{name: "plain seconds", in: "600", want: 10 * time.Minute},
		{name: "seconds suffix", in: "600s", want: 10 * time.Minute},
		{name: "minutes rejected", in: "10m", wantErr: true},
		{name: "millis rejected", in: "500ms", wantErr: true},
		{name: "invalid", in: "bad", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseSeconds(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseSeconds(%q) err=nil, want error", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSeconds(%q) err=%v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("ParseSeconds(%q)=%s, want %s", tc.in, got, tc.want)
			}
		})
	}
}

func TestCatalogAlignedDefaults(t *testing.T) {
	cat := timeoutprofile.Default()
	interactive := cat.Profile(timeoutprofile.Interactive)
	if DefaultSharedBackendResponseTimeout != interactive.EdgeIdle {
		t.Fatalf("DefaultSharedBackendResponseTimeout=%s, want catalog EdgeIdle %s",
			DefaultSharedBackendResponseTimeout, interactive.EdgeIdle)
	}
	if MinResponseTimeout != cat.Bounds.TotalMin {
		t.Fatalf("MinResponseTimeout=%s, want %s", MinResponseTimeout, cat.Bounds.TotalMin)
	}
	if MaxResponseTimeout != cat.Bounds.TotalMax {
		t.Fatalf("MaxResponseTimeout=%s, want %s", MaxResponseTimeout, cat.Bounds.TotalMax)
	}
}

func TestDefaultTimeout(t *testing.T) {
	t.Run("use env value when valid", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "720")
		got := DefaultTimeout()
		if got != 12*time.Minute {
			t.Fatalf("DefaultTimeout() = %s, want 12m", got)
		}
	})

	t.Run("use env value with s suffix", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "720s")
		got := DefaultTimeout()
		if got != 12*time.Minute {
			t.Fatalf("DefaultTimeout() = %s, want 12m", got)
		}
	})

	t.Run("use env above TotalMin below default", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "45")
		got := DefaultTimeout()
		if got != 45*time.Second {
			t.Fatalf("DefaultTimeout() = %s, want 45s", got)
		}
	})

	t.Run("fallback when env is empty", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "")
		got := DefaultTimeout()
		if got != DefaultSharedBackendResponseTimeout {
			t.Fatalf("DefaultTimeout() = %s, want %s", got, DefaultSharedBackendResponseTimeout)
		}
	})

	t.Run("fallback when env below TotalMin", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "10")
		got := DefaultTimeout()
		if got != DefaultSharedBackendResponseTimeout {
			t.Fatalf("DefaultTimeout() = %s, want %s", got, DefaultSharedBackendResponseTimeout)
		}
	})

	t.Run("fallback when env uses minutes", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "12m")
		got := DefaultTimeout()
		if got != DefaultSharedBackendResponseTimeout {
			t.Fatalf("DefaultTimeout() = %s, want %s", got, DefaultSharedBackendResponseTimeout)
		}
	})

	t.Run("fallback when env is invalid", func(t *testing.T) {
		t.Setenv(EnvDefaultResponseTimeout, "bad-value")
		got := DefaultTimeout()
		if got != DefaultSharedBackendResponseTimeout {
			t.Fatalf("DefaultTimeout() = %s, want %s", got, DefaultSharedBackendResponseTimeout)
		}
	})
}

func TestEffectiveResponseTimeout(t *testing.T) {
	t.Run("explicit total wins over default and floor", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(10*time.Minute, &metav1.Duration{Duration: 45 * time.Second}, 30*time.Minute)
		if err != nil {
			t.Fatalf("EffectiveResponseTimeout() err = %v", err)
		}
		if got != "45s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 45s", got)
		}
	})

	t.Run("explicit total longer than default", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(10*time.Minute, &metav1.Duration{Duration: 30 * time.Minute}, 0)
		if err != nil {
			t.Fatalf("EffectiveResponseTimeout() err = %v", err)
		}
		if got != "1800s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 1800s", got)
		}
	})

	t.Run("no manifest uses max of default and floor", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(10*time.Minute, nil, 30*time.Minute)
		if err != nil {
			t.Fatalf("EffectiveResponseTimeout() err = %v", err)
		}
		if got != "1800s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 1800s", got)
		}
	})

	t.Run("no manifest keeps platform default", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(DefaultSharedBackendResponseTimeout, nil, 0)
		if err != nil {
			t.Fatalf("EffectiveResponseTimeout() err = %v", err)
		}
		if got != "600s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 600s", got)
		}
	})

	t.Run("invalid explicit falls back to max default floor", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(10*time.Minute, &metav1.Duration{Duration: 10 * time.Second}, 20*time.Minute)
		if err != nil {
			t.Fatalf("EffectiveResponseTimeout() err = %v", err)
		}
		if got != "1200s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 1200s", got)
		}
	})

	t.Run("above max default still returns safe default", func(t *testing.T) {
		got, err := EffectiveResponseTimeout(MaxResponseTimeout+time.Second, nil, 0)
		if err == nil {
			t.Fatal("EffectiveResponseTimeout() err = nil, want error")
		}
		if got != "600s" {
			t.Fatalf("EffectiveResponseTimeout() = %q, want 600s", got)
		}
	})
}

var gatewaySecondsPattern = regexp.MustCompile(`^\d+s$`)

func TestEffectiveTimeoutSecondsFormat(t *testing.T) {
	cases := []struct {
		name       string
		def        time.Duration
		manifest   *metav1.Duration
		eg         time.Duration
		wantResult string
		wantErr    bool
	}{
		{name: "below TotalMin rejected", def: 10 * time.Second, wantResult: "600s", wantErr: true},
		{name: "TotalMin boundary", def: MinResponseTimeout, wantResult: "30s"},
		{name: "ten minutes", def: 10 * time.Minute, wantResult: "600s"},
		{name: "twelve minutes", def: 12 * time.Minute, wantResult: "720s"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EffectiveResponseTimeout(tc.def, tc.manifest, tc.eg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("err=nil, want error")
				}
			} else if err != nil {
				t.Fatalf("EffectiveResponseTimeout() err = %v", err)
			}
			if got != tc.wantResult {
				t.Fatalf("EffectiveResponseTimeout() = %q, want %q", got, tc.wantResult)
			}
			if !gatewaySecondsPattern.MatchString(got) {
				t.Fatalf("result %q does not match seconds pattern", got)
			}
		})
	}
}

func TestValidateResponseTimeout(t *testing.T) {
	cases := []struct {
		name    string
		d       time.Duration
		wantErr bool
	}{
		{name: "TotalMin", d: MinResponseTimeout, wantErr: false},
		{name: "below TotalMin", d: MinResponseTimeout - time.Second, wantErr: true},
		{name: "45s allowed", d: 45 * time.Second, wantErr: false},
		{name: "max boundary", d: MaxResponseTimeout, wantErr: false},
		{name: "above max", d: MaxResponseTimeout + time.Second, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateResponseTimeout(tc.d)
			if tc.wantErr && err == nil {
				t.Fatalf("ValidateResponseTimeout(%s) err=nil, want error", tc.d)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("ValidateResponseTimeout(%s) err=%v, want nil", tc.d, err)
			}
		})
	}
}
