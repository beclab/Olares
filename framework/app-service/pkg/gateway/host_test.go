package gateway

import "testing"

func TestNormalizeHostPattern(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{in: "App.Example.COM", want: "app.example.com"},
		{in: "host.example.com:443", want: "host.example.com"},
		{in: " a.b.c ", want: "a.b.c"},
		{in: "https://x.y", wantErr: true},
		{in: "a/b", wantErr: true},
		{in: "*.example.com", wantErr: true},
		{in: "", wantErr: true},
	}
	for _, c := range cases {
		got, err := NormalizeHostPattern(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("NormalizeHostPattern(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("NormalizeHostPattern(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("NormalizeHostPattern(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeHostOrLogicalPattern(t *testing.T) {
	if got, err := NormalizeHostOrLogicalPattern("ab12cd34.*.olares.com"); err != nil || got != "ab12cd34.*.olares.com" {
		t.Errorf("logical pattern got %q err %v", got, err)
	}
	if got, err := NormalizeHostOrLogicalPattern("ab12cd34.shared.olares.com"); err != nil || got != "ab12cd34.shared.olares.com" {
		t.Errorf("shared exact host got %q err %v", got, err)
	}
	if !IsLogicalHostPattern("ab12cd34.*.olares.com") {
		t.Error("IsLogicalHostPattern should be true for wildcard pattern")
	}
	if IsLogicalHostPattern("a.example.com") {
		t.Error("IsLogicalHostPattern should be false for exact host")
	}
	for _, bad := range []string{"*.*.olares.com", "ab.cd.*", "*.olares.com"} {
		if _, err := NormalizeHostOrLogicalPattern(bad); err == nil {
			t.Errorf("NormalizeHostOrLogicalPattern(%q) expected error", bad)
		}
	}
}
