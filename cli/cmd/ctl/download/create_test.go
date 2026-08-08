package download

import (
	"context"
	"strings"
	"testing"
)

func TestNormalizeSelectFiles(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		wantCSV string
		wantOK  bool
		wantErr bool
	}{
		{name: "blank omits", in: "   ", wantOK: false},
		{name: "all omits", in: "all", wantOK: false},
		{name: "all case-insensitive omits", in: "ALL", wantOK: false},
		{name: "single index", in: "3", wantCSV: "3", wantOK: true},
		{name: "csv preserved order", in: "1,3,5", wantCSV: "1,3,5", wantOK: true},
		{name: "spaces trimmed", in: " 1 , 3 ", wantCSV: "1,3", wantOK: true},
		{name: "zero rejected", in: "0", wantErr: true},
		{name: "negative rejected", in: "-1", wantErr: true},
		{name: "non-integer rejected", in: "1,x", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			csv, ok, err := normalizeSelectFiles(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("normalizeSelectFiles(%q) expected error, got csv=%q ok=%v", tc.in, csv, ok)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeSelectFiles(%q) unexpected error: %v", tc.in, err)
			}
			if ok != tc.wantOK {
				t.Fatalf("normalizeSelectFiles(%q) ok=%v want %v", tc.in, ok, tc.wantOK)
			}
			if ok && csv != tc.wantCSV {
				t.Fatalf("normalizeSelectFiles(%q) csv=%q want %q", tc.in, csv, tc.wantCSV)
			}
		})
	}
}

func TestValidateYTDLPQuality(t *testing.T) {
	for _, valid := range []string{"", "best", "2160p", "1080p", "720p", "480p", "360p", "audio"} {
		if err := validateYTDLPQuality(valid, false); err != nil {
			t.Fatalf("validateYTDLPQuality(%q) unexpected error: %v", valid, err)
		}
	}

	if err := validateYTDLPQuality("", true); err == nil || !strings.Contains(err.Error(), "--quality is required") {
		t.Fatalf("required empty quality should fail locally, got %v", err)
	}
	err := validateYTDLPQuality("4k", false)
	if err == nil {
		t.Fatal("unsupported quality should fail locally")
	}
	for _, want := range []string{"unsupported --quality", ytdlpQualityValues} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("quality error %q missing %q", err, want)
		}
	}
}

func TestRunCreateValidatesQualityFromExtra(t *testing.T) {
	err := runCreate(
		context.Background(),
		nil,
		"https://example.com/video",
		"",
		"",
		"",
		"",
		"",
		`{"ytdlp_quality":"4k"}`,
		"",
		"",
		"table",
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported --quality") {
		t.Fatalf("invalid quality in --extra should fail locally, got %v", err)
	}
}
