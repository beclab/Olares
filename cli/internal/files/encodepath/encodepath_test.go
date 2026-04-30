package encodepath

import "testing"

func TestEncodeURIComponent(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"hello world", "hello%20world"},
		{"a&b=c", "a%26b%3Dc"},
		{"a+b", "a%2Bb"},
		{"a/b", "a%2Fb"},
		{"!*'()", "!*'()"},
		{"-_.~", "-_.~"},
		{"中文.txt", "%E4%B8%AD%E6%96%87.txt"},
		{"~user@host", "~user%40host"},
		{"file (1).txt", "file%20(1).txt"},
		{"100%", "100%25"},
		{"\x00\n\r\t", "%00%0A%0D%09"},
		{"  /  ", "%20%20%2F%20%20"},
		{"a?b#c", "a%3Fb%23c"},
	}
	for _, c := range cases {
		got := EncodeURIComponent(c.in)
		if got != c.want {
			t.Errorf("EncodeURIComponent(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEncodeURL(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"/", "/"},
		{"a/b/c", "a/b/c"},
		{"/a/b/", "/a/b/"},
		{"a b/c d/", "a%20b/c%20d/"},
		{"中文/files/x.txt", "%E4%B8%AD%E6%96%87/files/x.txt"},
		{"//x", "//x"},
		{"x//", "x//"},
		{"/Home/Photos/IMG_001.jpg", "/Home/Photos/IMG_001.jpg"},
		{"/dir/file (1).txt", "/dir/file%20(1).txt"},
		{"/sub/a&b/c=d", "/sub/a%26b/c%3Dd"},
	}
	for _, c := range cases {
		got := EncodeURL(c.in)
		if got != c.want {
			t.Errorf("EncodeURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
