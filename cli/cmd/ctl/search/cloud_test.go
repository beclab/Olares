package search

import (
	"strings"
	"testing"
)

func TestCloudProviderMapping(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		p           cloudProvider
		wantApp     string
		wantAccount string
		wantUse     string
	}{
		{"google drive", googleDriveProvider, appGoogleDrive, "google", "gdrive"},
		{"dropbox", dropboxProvider, appDropbox, "dropbox", "dropbox"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.p.app != tc.wantApp {
				t.Fatalf("app = %q, want %q", tc.p.app, tc.wantApp)
			}
			if tc.p.accountType != tc.wantAccount {
				t.Fatalf("accountType = %q, want %q", tc.p.accountType, tc.wantAccount)
			}
			if tc.p.use != tc.wantUse {
				t.Fatalf("use = %q, want %q", tc.p.use, tc.wantUse)
			}
		})
	}
}

func TestCloudEmptyMessage(t *testing.T) {
	t.Parallel()

	p := dropboxProvider

	t.Run("probe failed falls back to no results", func(t *testing.T) {
		t.Parallel()
		got := cloudEmptyMessage(p, nil, true)
		if got != "no results" {
			t.Fatalf("got %q, want %q", got, "no results")
		}
	})

	t.Run("available account stays no results", func(t *testing.T) {
		t.Parallel()
		got := cloudEmptyMessage(p, []accountMini{{Name: "me", Available: true}}, false)
		if got != "no results" {
			t.Fatalf("got %q, want %q", got, "no results")
		}
	})

	t.Run("no available account prints LarePass binding hint", func(t *testing.T) {
		t.Parallel()
		got := cloudEmptyMessage(p, []accountMini{{Name: "stale", Available: false}}, false)
		if !strings.Contains(got, "LarePass → Settings → Integration") {
			t.Fatalf("expected LarePass binding hint, got %q", got)
		}
		if !strings.Contains(got, "list-by-type dropbox") {
			t.Fatalf("expected list-by-type hint, got %q", got)
		}
		if !strings.Contains(got, "Dropbox") {
			t.Fatalf("expected provider title, got %q", got)
		}
	})

	t.Run("empty account list prints binding hint", func(t *testing.T) {
		t.Parallel()
		got := cloudEmptyMessage(p, nil, false)
		if !strings.Contains(got, "LarePass → Settings → Integration") {
			t.Fatalf("expected LarePass binding hint, got %q", got)
		}
	})
}

func TestHasAvailableAccount(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   []accountMini
		want bool
	}{
		{"nil", nil, false},
		{"empty", []accountMini{}, false},
		{"all unavailable", []accountMini{{Available: false}, {Available: false}}, false},
		{"one available", []accountMini{{Available: false}, {Available: true}}, true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := hasAvailableAccount(tc.in); got != tc.want {
				t.Fatalf("hasAvailableAccount = %v, want %v", got, tc.want)
			}
		})
	}
}
