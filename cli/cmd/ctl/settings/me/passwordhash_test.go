package me

import "testing"

func TestSaltedPasswordSchemeSwitch(t *testing.T) {
	// 1.12.0 (>= 1.12.0-0) hits the salted path.
	got := saltedPassword("hunter2", "1.12.0")
	if got == "hunter2" {
		t.Fatalf("expected salted hash for OS >= 1.12.0-0, got the raw password")
	}
	if len(got) != 32 {
		t.Fatalf("expected 32-char hex MD5, got %q", got)
	}
	// Empty version → passthrough (we couldn't determine the OS version
	// at all, so don't risk hashing against a backend that doesn't expect
	// it).
	if saltedPassword("hunter2", "") != "hunter2" {
		t.Errorf("expected passthrough for empty OS version")
	}
}

func TestLocalPart(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"alice@olares.com", "alice"},
		{"bob@example.org", "bob"},
		{"plainuser", "plainuser"},
		{"", ""},
		{"@noLocal", "@noLocal"},
		{"  trim@me.com  ", "trim"},
	}
	for _, c := range cases {
		got := localPart(c.in)
		if got != c.want {
			t.Errorf("localPart(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
