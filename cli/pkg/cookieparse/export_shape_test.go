package cookieparse

import "testing"

// Structural mirror of a real browser-extension export that carries NO
// #HttpOnly_ markers. Values are fake; only the shape matters.
const noPrefixExport = `# Netscape HTTP Cookie File
# https://curl.haxx.se/rfc/cookie_spec.html
# This is a generated file! Do not edit.

.youtube.com	TRUE	/	TRUE	1806549705	LOGIN_INFO	AFmmF2s:QUQ3MjNmeFlTaFpU
.youtube.com	TRUE	/	FALSE	1819290749	SID	g.a000AgnV9j--ffFZ
.youtube.com	TRUE	/	TRUE	1819290749	__Secure-3PSID	g.a000AgnV2_BXfSfsgR7E
.youtube.com	TRUE	/	FALSE	1819290749	HSID	ACTBMTz799m9DmDvG
.youtube.com	TRUE	/	TRUE	1819290749	SSID	A-v3fztR-xwjPHWXz
.youtube.com	TRUE	/	FALSE	1819290749	APISID	F_T3aoxoHNRkCZOJ/AFe0pXXkuliMEUyJN
.youtube.com	TRUE	/	TRUE	1821612966	PREF	f7=4000&tz=Asia.Shanghai&f6=40000000
.youtube.com	TRUE	/	TRUE	0	YSC	ZSwvbV5T6_E
.youtube.com	TRUE	/	TRUE	1802604970	VISITOR_PRIVACY_METADATA	CgJVUxIEGgAgQw%3D%3D
`

func TestShapeOfAnExportWithoutHttpOnlyMarkers(t *testing.T) {
	res, format, err := Parse(noPrefixExport, FormatAuto, "")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if format != FormatNetscape {
		t.Fatalf("format = %q, want netscape", format)
	}
	if len(res.InvalidLines) != 0 {
		t.Fatalf("rejected lines: %v", res.InvalidLines)
	}
	if got := res.Count(); got != 9 {
		t.Fatalf("parsed %d records, want 9", got)
	}

	// Nothing carries the marker, so nothing is flagged httpOnly.
	for _, rec := range res.Cookies[".youtube.com"] {
		if rec.HttpOnly {
			t.Errorf("%s: httpOnly=true, want false (no marker in input)", rec.Name)
		}
	}

	// Values keep their punctuation intact.
	for _, tc := range []struct{ name, want string }{
		{"APISID", "F_T3aoxoHNRkCZOJ/AFe0pXXkuliMEUyJN"},
		{"PREF", "f7=4000&tz=Asia.Shanghai&f6=40000000"},
		{"VISITOR_PRIVACY_METADATA", "CgJVUxIEGgAgQw%3D%3D"},
		{"LOGIN_INFO", "AFmmF2s:QUQ3MjNmeFlTaFpU"},
	} {
		if got := recordByName(t, res, ".youtube.com", tc.name).Value; got != tc.want {
			t.Errorf("%s value = %q, want %q", tc.name, got, tc.want)
		}
	}

	if got := recordByName(t, res, ".youtube.com", "YSC").ExpiryUnix(); got != 0 {
		t.Errorf("YSC expiry = %d, want 0 (session cookie)", got)
	}
	if got := recordByName(t, res, ".youtube.com", "SID").ExpiryUnix(); got != 1819290749 {
		t.Errorf("SID expiry = %d, want 1819290749", got)
	}
	if rec := recordByName(t, res, ".youtube.com", "SID"); rec.Secure {
		t.Errorf("SID secure=true, want false (column 4 is FALSE)")
	}
}
