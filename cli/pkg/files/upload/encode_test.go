package upload

import "testing"

// Test cases derived from JavaScript's encodeURIComponent reference output.
// If you ever change encode.go, re-run these examples in a Node REPL to
// re-confirm parity:
//
//	> encodeURIComponent("hello world")  // 'hello%20world'
//	> encodeURIComponent("a&b=c")         // 'a%26b%3Dc'
//	> encodeURIComponent("中文.txt")       // '%E4%B8%AD%E6%96%87.txt'
//
// We deliberately test the boundary characters where Go's url.QueryEscape
// would diverge from JS (' ' as '+', and the !*'() set escaping) so any
// regression to "just use url.QueryEscape" gets caught here.
func TestEncodeURIComponent(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"abc", "abc"},
		{"hello world", "hello%20world"},                       // space → %20 (NOT '+')
		{"a&b=c", "a%26b%3Dc"},                                 // query metacharacters
		{"a+b", "a%2Bb"},                                       // '+' percent-encoded
		{"a/b", "a%2Fb"},                                       // '/' is reserved
		{"!*'()", "!*'()"},                                     // unreserved-extras (NOT escaped)
		{"-_.~", "-_.~"},                                       // RFC 3986 unreserved
		{"中文.txt", "%E4%B8%AD%E6%96%87.txt"},                   // UTF-8 multibyte
		{"~user@host", "~user%40host"},                         // '@' encoded
		{"file (1).txt", "file%20(1).txt"},                     // spaces + parens together
		{"100%", "100%25"},                                     // '%' itself
		{"\x00\n\r\t", "%00%0A%0D%09"},                         // control chars
		{"  /  ", "%20%20%2F%20%20"},                           // multiple spaces around '/'
		{"a?b#c", "a%3Fb%23c"},                                 // '?' and '#'
	}
	for _, c := range cases {
		got := encodeURIComponent(c.in)
		if got != c.want {
			t.Errorf("encodeURIComponent(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// EncodeURL is the path-segment-aware variant: '/' separators are kept,
// each segment is encodeURIComponent'd. The interesting cases are
// preserving leading/trailing slashes (the backend uses them as directory
// hints) and the empty-segment edge case from "//".
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
		{"//x", "//x"},                                         // empty leading segment kept
		{"x//", "x//"},                                         // empty trailing segment kept
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
