package utils

import "testing"

func TestUnescapeMountField(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{`/olares/share/Olares\040Research`, "/olares/share/Olares Research"},
		{`//host/Olares\040Research`, "//host/Olares Research"},
		{`/mnt/a\011b`, "/mnt/a\tb"},
		{`/mnt/a\012b`, "/mnt/a\nb"},
		{`/mnt/a\134b`, `/mnt/a\b`},
		{`/mnt/plain`, "/mnt/plain"},
		{`/mnt/bad\04`, `/mnt/bad\04`},
		{`/mnt/bad\9xx`, `/mnt/bad\9xx`},
		{`/mnt/bad\777`, `/mnt/bad\777`},
		{`/mnt/trailing\`, `/mnt/trailing\`},
	}

	for _, c := range cases {
		if got := unescapeMountField(c.in); got != c.want {
			t.Errorf("unescapeMountField(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
