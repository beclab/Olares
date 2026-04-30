package files

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestFormatSize(t *testing.T) {
	cases := []struct {
		name  string
		n     int64
		isDir bool
		want  string
	}{
		{name: "directory always dash", n: 12345, isDir: true, want: "-"},
		{name: "zero bytes", n: 0, want: "0B"},
		{name: "one byte", n: 1, want: "1B"},
		{name: "just under 1K", n: 1023, want: "1023B"},
		{name: "exactly 1K", n: 1024, want: "1.0KB"},
		{name: "fractional KB", n: 1536, want: "1.5KB"},
		{name: "just under 1M", n: 1024*1024 - 1, want: "1024.0KB"},
		{name: "exactly 1M", n: 1024 * 1024, want: "1.0MB"},
		{name: "1.5GB", n: 1536 * 1024 * 1024, want: "1.5GB"},
		{name: "1TB", n: 1 << 40, want: "1.0TB"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatSize(c.n, c.isDir); got != c.want {
				t.Errorf("formatSize(%d, %v) = %q, want %q", c.n, c.isDir, got, c.want)
			}
		})
	}
}

func TestFormatHTTPError(t *testing.T) {
	const url = "https://files.alice.olares.com/api/resources/drive/Home/"
	const olaresID = "alice@olares.com"

	cases := []struct {
		name           string
		status         int
		body           string
		wantSubstrs    []string
		wantNotSubstrs []string
	}{
		{
			name:        "401 surfaces re-login CTA with olaresId",
			status:      http.StatusUnauthorized,
			body:        `{"error":"unauthorized"}`,
			wantSubstrs: []string{"HTTP 401", "profile login", olaresID},
			// the body's error text shouldn't leak through on auth failures —
			// the CTA is more useful than the raw 401 reason
			wantNotSubstrs: []string{"unauthorized"},
		},
		{
			name:        "403 also routes to re-login CTA",
			status:      http.StatusForbidden,
			body:        ``,
			wantSubstrs: []string{"HTTP 403", "profile login", olaresID},
		},
		{
			name:        "500 with {error} surfaces backend message",
			status:      http.StatusInternalServerError,
			body:        `{"error":"node missing"}`,
			wantSubstrs: []string{"HTTP 500", "node missing", url},
		},
		{
			name:        "500 with code+message surfaces both",
			status:      http.StatusInternalServerError,
			body:        `{"code":1,"message":"Directory not exist."}`,
			wantSubstrs: []string{"HTTP 500", "code=1", "Directory not exist."},
		},
		{
			name:        "non-JSON body falls back to raw (truncated)",
			status:      http.StatusBadGateway,
			body:        "<html>upstream gone</html>",
			wantSubstrs: []string{"HTTP 502", "upstream gone"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := formatHTTPError(c.status, []byte(c.body), olaresID, url)
			if err == nil {
				t.Fatalf("formatHTTPError(%d): want error, got nil", c.status)
			}
			msg := err.Error()
			for _, s := range c.wantSubstrs {
				if !strings.Contains(msg, s) {
					t.Errorf("formatHTTPError(%d): want %q in %q", c.status, s, msg)
				}
			}
			for _, s := range c.wantNotSubstrs {
				if strings.Contains(msg, s) {
					t.Errorf("formatHTTPError(%d): did NOT want %q in %q", c.status, s, msg)
				}
			}
		})
	}
}

func TestRenderListing(t *testing.T) {
	fp, err := ParseFrontendPath("drive/Home/Documents/")
	if err != nil {
		t.Fatalf("ParseFrontendPath: %v", err)
	}
	parentMod := time.Date(2026, 4, 17, 11, 31, 51, 0, time.UTC)

	t.Run("empty shows header line then (empty)", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderListing(&buf, fp, listingResponse{
			Name:     "Documents",
			NumDirs:  0,
			NumFiles: 0,
			Modified: parentMod,
		}); err != nil {
			t.Fatalf("renderListing: %v", err)
		}
		out := buf.String()
		if strings.Contains(out, "MODE") || strings.Contains(out, "NAME") {
			t.Errorf("empty listing should not print table header, got: %q", out)
		}
		if !strings.Contains(out, "drive/Home/Documents/") {
			t.Errorf("expected requested path in header, got: %q", out)
		}
		if !strings.Contains(out, "0 dirs") || !strings.Contains(out, "0 files") {
			t.Errorf("expected zero dir/file counts in header, got: %q", out)
		}
		if !strings.Contains(out, "(empty)") {
			t.Errorf("missing (empty) marker in output: %q", out)
		}
	})

	t.Run("counts pluralize correctly (1 dir, 1 file)", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderListing(&buf, fp, listingResponse{
			NumDirs:  1,
			NumFiles: 1,
			Modified: parentMod,
			Items: []listingItem{
				{Name: "sub", IsDir: true, Mode: 0x80000000 | 0o755},
				{Name: "a.txt", IsDir: false, Mode: 0o644, Type: "text", Size: 12},
			},
		}); err != nil {
			t.Fatalf("renderListing: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "1 dir,") || !strings.Contains(out, "1 file,") {
			t.Errorf("expected singular forms '1 dir, 1 file' in header, got: %q", out)
		}
	})

	t.Run("falls back to counting items when envelope counts are zero", func(t *testing.T) {
		var buf bytes.Buffer
		if err := renderListing(&buf, fp, listingResponse{
			// NumDirs/NumFiles intentionally zero; backend should still tell us
			// what's in the dir via Items.
			Items: []listingItem{
				{Name: "x", IsDir: true},
				{Name: "y", IsDir: true},
				{Name: "z.txt", IsDir: false},
			},
		}); err != nil {
			t.Fatalf("renderListing: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "2 dirs") || !strings.Contains(out, "1 file") {
			t.Errorf("expected derived '2 dirs, 1 file' counts, got: %q", out)
		}
	})

	t.Run("dirs first then files alphabetical, with mode + type columns", func(t *testing.T) {
		mod := time.Date(2026, 4, 1, 12, 30, 0, 0, time.UTC)
		items := []listingItem{
			{Name: "zebra.txt", IsDir: false, Size: 10, Modified: mod, Mode: 0o644, Type: "text"},
			{Name: "Alpha", IsDir: true, Modified: mod, Mode: 0x80000000 | 0o755},
			{Name: "beta.md", IsDir: false, Size: 1024, Modified: mod, Mode: 0o644, Type: "text"},
			{Name: "movie.mp4", IsDir: false, Size: 1 << 20, Modified: mod, Mode: 0o644, Type: "video"},
			{Name: "zeta", IsDir: true, Modified: mod, Mode: 0x80000000 | 0o755},
		}
		var buf bytes.Buffer
		if err := renderListing(&buf, fp, listingResponse{
			NumDirs:  2,
			NumFiles: 3,
			Modified: parentMod,
			Items:    items,
		}); err != nil {
			t.Fatalf("renderListing: %v", err)
		}
		out := buf.String()

		// Header line then column header then 5 rows.
		if !strings.Contains(out, "MODE") || !strings.Contains(out, "TYPE") {
			t.Errorf("expected MODE and TYPE columns, got: %q", out)
		}
		if !strings.Contains(out, "drwxr-xr-x") || !strings.Contains(out, "-rw-r--r--") {
			t.Errorf("expected decoded mode strings, got: %q", out)
		}
		if !strings.Contains(out, "video") {
			t.Errorf("expected 'video' type for movie.mp4, got: %q", out)
		}

		// Order: drop the banner line and the column header, then check NAME col.
		lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
		if len(lines) < 7 { // banner + col header + 5 rows
			t.Fatalf("expected banner + header + 5 rows, got %d lines: %q", len(lines), out)
		}
		names := make([]string, 0, len(lines)-2)
		for _, ln := range lines[2:] {
			fields := strings.Fields(ln)
			names = append(names, fields[len(fields)-1])
		}
		want := []string{"Alpha/", "zeta/", "beta.md", "movie.mp4", "zebra.txt"}
		for i, n := range want {
			if names[i] != n {
				t.Errorf("row %d: name = %q, want %q (full output: %q)", i, names[i], n, out)
			}
		}
	})
}

func TestFormatMode(t *testing.T) {
	cases := []struct {
		name      string
		mode      uint32
		isDir     bool
		isSymlink bool
		want      string
	}{
		{name: "regular 0644", mode: 0o644, want: "-rw-r--r--"},
		{name: "exec 0775", mode: 0o775, want: "-rwxrwxr-x"},
		{name: "directory 0755", mode: 0x80000000 | 0o755, want: "drwxr-xr-x"},
		{name: "directory weird high bits as observed live (FileMode 2147484141)", mode: 2147484141, want: "drwxr-xr-x"},
		{name: "fallback to flags when mode=0 and dir", mode: 0, isDir: true, want: "d---------"},
		{name: "fallback to flags when mode=0 and symlink", mode: 0, isSymlink: true, want: "L---------"},
		{name: "fallback to flags when mode=0 and regular", mode: 0, want: "----------"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := formatMode(c.mode, c.isDir, c.isSymlink); got != c.want {
				t.Errorf("formatMode(%d, dir=%v, sym=%v) = %q, want %q", c.mode, c.isDir, c.isSymlink, got, c.want)
			}
		})
	}
}

// TestListingResponseDecode_CloudDriveAwss3 covers the awss3 envelope
// shape from a real backend response: children under top-level `data`
// (NOT `items`), `mode` and `modified` as empty strings on every
// child, no `numDirs` / `numFiles` summary, and `fileSize` populated
// alongside `size`. The decoder must surface a fully-populated
// listingResponse the renderer can use without per-namespace branches.
func TestListingResponseDecode_CloudDriveAwss3(t *testing.T) {
	const body = `{
		"data": [
			{
				"name": "03 (1).avi",
				"isDir": false,
				"isSymlink": false,
				"size": 5788048,
				"fileSize": 5788048,
				"mode": "",
				"modified": "",
				"path": "/03 (1).avi",
				"type": "",
				"meta": {"key": "03 (1).avi", "last_modified": ""}
			},
			{
				"name": "datasets",
				"isDir": true,
				"isSymlink": false,
				"size": 0,
				"fileSize": 0,
				"mode": "",
				"modified": "",
				"path": "/datasets",
				"type": ""
			}
		],
		"fileExtend": "AKIAVJDTX4VSSYHHRWAU",
		"filePath": "/",
		"fileType": "awss3",
		"name": "",
		"status_code": "SUCCESS"
	}`
	var got listingResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal cloud envelope: %v", err)
	}
	if len(got.Items) != 2 {
		t.Fatalf("want 2 items merged from `data`, got %d", len(got.Items))
	}
	first := got.Items[0]
	if first.Name != "03 (1).avi" || first.IsDir || first.Size != 5788048 {
		t.Errorf("first item mismatch: %+v", first)
	}
	if !first.Modified.IsZero() {
		t.Errorf("modified should decode to zero time for empty-string input, got %v", first.Modified)
	}
	if first.Mode != 0 {
		t.Errorf("mode should decode to 0 for empty-string input, got %d", first.Mode)
	}
	second := got.Items[1]
	if second.Name != "datasets" || !second.IsDir {
		t.Errorf("second item mismatch: %+v", second)
	}
}

// TestListingResponseDecode_CloudDriveFileSizeOnly: when the server
// only populates `fileSize` (no `size`), Items[].Size must still
// surface the right byte count — the renderer's SIZE column would
// otherwise show 0B for every cloud-drive entry.
func TestListingResponseDecode_CloudDriveFileSizeOnly(t *testing.T) {
	const body = `{
		"data": [
			{"name":"big.bin","isDir":false,"fileSize":17236328572,"mode":"","modified":""}
		]
	}`
	var got listingResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Size != 17236328572 {
		t.Errorf("want fileSize fallback, got %+v", got.Items)
	}
}

// TestListingResponseDecode_DriveStillWorks ensures the standard
// Drive/Sync envelope (children under `items`, `mode` as a number,
// `modified` as RFC3339) keeps decoding correctly after the cloud-
// envelope tolerance was added.
func TestListingResponseDecode_DriveStillWorks(t *testing.T) {
	const body = `{
		"name": "Documents",
		"path": "/drive/Home/Documents",
		"modified": "2026-04-17T19:31:51Z",
		"mode": 2147484141,
		"numDirs": 1,
		"numFiles": 1,
		"items": [
			{"name":"sub","isDir":true,"size":0,"mode":2147484141,"modified":"2026-04-01T12:30:00Z"},
			{"name":"a.txt","isDir":false,"size":12,"mode":420,"modified":"2026-04-01T12:30:00Z","type":"text"}
		]
	}`
	var got listingResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal Drive envelope: %v", err)
	}
	if got.NumDirs != 1 || got.NumFiles != 1 {
		t.Errorf("want NumDirs=1 NumFiles=1, got %d/%d", got.NumDirs, got.NumFiles)
	}
	if got.Modified.IsZero() {
		t.Error("Drive envelope must decode `modified` as a real timestamp")
	}
	if got.Mode == 0 {
		t.Error("Drive envelope must decode `mode` as a non-zero uint32")
	}
	if len(got.Items) != 2 {
		t.Fatalf("want 2 items, got %d", len(got.Items))
	}
	if got.Items[1].Name != "a.txt" || got.Items[1].Size != 12 || got.Items[1].Mode != 420 {
		t.Errorf("file item mismatch: %+v", got.Items[1])
	}
}

// TestListingResponseDecode_ItemsWinsWhenBothPresent: defensive —
// if a backend transition emits both shapes, the canonical `items`
// field wins so a hybrid server can't double-count.
func TestListingResponseDecode_ItemsWinsWhenBothPresent(t *testing.T) {
	const body = `{
		"items":[{"name":"from-items","isDir":false,"size":1,"mode":420,"modified":"2026-04-01T12:30:00Z"}],
		"data":[{"name":"from-data","isDir":false,"size":2,"mode":"","modified":""}]
	}`
	var got listingResponse
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Items) != 1 || got.Items[0].Name != "from-items" {
		t.Errorf("want items-wins, got %+v", got.Items)
	}
}

// TestListingResponseDecode_BadModified: a malformed timestamp must
// surface as an error rather than silently zeroing the field —
// hiding server-side bugs would leave the renderer's MODIFIED
// column blank with no diagnostic.
func TestListingResponseDecode_BadModified(t *testing.T) {
	const body = `{
		"items": [
			{"name":"x","isDir":false,"size":1,"mode":420,"modified":"not-a-time"}
		]
	}`
	var got listingResponse
	err := json.Unmarshal([]byte(body), &got)
	if err == nil {
		t.Fatal("expected decode error for bad timestamp")
	}
	if !strings.Contains(err.Error(), "modified") {
		t.Errorf("error should mention `modified`: %v", err)
	}
}

func TestFormatType(t *testing.T) {
	cases := []struct {
		t     string
		isDir bool
		want  string
	}{
		{t: "video", want: "video"},
		{t: "text", want: "text"},
		{t: "", want: "-"},
		{t: "", isDir: true, want: "-"},
		{t: "video", isDir: true, want: "-"},
	}
	for _, c := range cases {
		if got := formatType(c.t, c.isDir); got != c.want {
			t.Errorf("formatType(%q, dir=%v) = %q, want %q", c.t, c.isDir, got, c.want)
		}
	}
}
