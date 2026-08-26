package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
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
		false,
		0,
		false,
		0,
		"table",
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported --quality") {
		t.Fatalf("invalid quality in --extra should fail locally, got %v", err)
	}
}

func TestRunCreateValidatesApp(t *testing.T) {
	err := runCreate(
		context.Background(),
		nil,
		"https://example.com/video",
		"namespace",
		"",
		"",
		"",
		"",
		"",
		"",
		"",
		false,
		0,
		false,
		0,
		"table",
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported --app") {
		t.Fatalf("unknown --app should fail locally, got %v", err)
	}
}

// recordedCall is one request as it went out, kept whole so a test can
// assert on the body and not just the path.
type recordedCall struct {
	method string
	path   string
	body   interface{}
}

// recordingDoer answers the inspect probe and the create POST with
// canned envelopes while keeping every call, so a test can check both
// what was sent and in which order.
type recordingDoer struct {
	inspect    []byte
	inspectErr error
	create     []byte
	// onInspect / onCreate observe the context each call was handed, so
	// a test can assert on deadlines. A non-nil error short-circuits the
	// canned response.
	onInspect func(context.Context) error
	onCreate  func(context.Context) error
	calls     []recordedCall
}

func (r *recordingDoer) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
	r.calls = append(r.calls, recordedCall{method: method, path: path, body: body})
	if strings.HasPrefix(path, "/api/url/inspect") {
		if r.onInspect != nil {
			if err := r.onInspect(ctx); err != nil {
				return err
			}
		}
		if r.inspectErr != nil {
			return r.inspectErr
		}
		return json.Unmarshal(r.inspect, out)
	}
	if r.onCreate != nil {
		if err := r.onCreate(ctx); err != nil {
			return err
		}
	}
	return json.Unmarshal(r.create, out)
}

func (r *recordingDoer) probed() bool {
	for _, c := range r.calls {
		if strings.HasPrefix(c.path, "/api/url/inspect") {
			return true
		}
	}
	return false
}

func (r *recordingDoer) createReq(t *testing.T) NewDownloadReq {
	t.Helper()
	for _, c := range r.calls {
		if c.method != "POST" || c.path != "/api/download" {
			continue
		}
		req, ok := c.body.(NewDownloadReq)
		if !ok {
			t.Fatalf("create body is %T, want NewDownloadReq", c.body)
		}
		return req
	}
	t.Fatalf("no create call recorded, calls=%+v", r.calls)
	return NewDownloadReq{}
}

func inspectEnvelope(t *testing.T, title string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{"provider": "yt-dlp", "title": title},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// TestSubmitCreateSendsInspectTitleAsFileName is the regression guard
// for the CLI/LarePass name mismatch: the same YouTube link showed its
// title in LarePass but landed as the bare URL from the CLI, because
// only LarePass prefilled file_name from the inspect title. Asserting
// on the probe result alone would not have caught it — the bug was that
// nothing carried that title into the create body — so this pins the
// request actually put on the wire.
func TestSubmitCreateSendsInspectTitleAsFileName(t *testing.T) {
	const (
		youtubeURL = "https://www.youtube.com/watch?v=oi2QgPH61JM"
		title      = "Black Myth: Zhong Kui – 15 Minutes Gameplay Trailer"
	)
	d := &recordingDoer{
		inspect: inspectEnvelope(t, title),
		create:  []byte(`{"code":200,"data":{"id":20,"status":"waiting","download_provider":"yt-dlp"}}`),
	}

	task, err := submitCreate(
		context.Background(),
		&preparedClient{doer: d},
		NewDownloadReq{URL: youtubeURL, App: "wise", Path: "drive/Home/Downloads/"},
		nameProbe{url: youtubeURL, enabled: true},
	)
	if err != nil {
		t.Fatalf("submitCreate: %v", err)
	}
	if task.ID != 20 {
		t.Fatalf("task id = %d, want 20", task.ID)
	}
	if got := d.createReq(t).FileName; got != title {
		t.Fatalf("file_name = %q, want the probed title %q", got, title)
	}
	// The probe has to precede the create or its title cannot land.
	if len(d.calls) != 2 || !strings.HasPrefix(d.calls[0].path, "/api/url/inspect") {
		t.Fatalf("expected an inspect then a create, got %+v", d.calls)
	}
}

// A slash in a title used to make the daemon mkdir the head of the
// title and nest the file inside it, so the sanitised form — not the
// raw title — is what may reach the server.
func TestSubmitCreateSanitisesTitleBeforeSending(t *testing.T) {
	d := &recordingDoer{
		inspect: inspectEnvelope(t, "Foo / Bar"),
		create:  []byte(`{"code":200,"data":{"id":1,"status":"waiting"}}`),
	}
	if _, err := submitCreate(
		context.Background(),
		&preparedClient{doer: d},
		NewDownloadReq{URL: "https://host/v", App: "wise"},
		nameProbe{url: "https://host/v", enabled: true},
	); err != nil {
		t.Fatalf("submitCreate: %v", err)
	}
	if got := d.createReq(t).FileName; got != "Foo _ Bar" {
		t.Fatalf("file_name = %q, want the path separator replaced", got)
	}
}

// awaitDeadline blocks until ctx expires, standing in for a probe the
// server is slow to answer. It gives up on its own after a second so a
// missing cap fails the assertion instead of hanging the suite until the
// go test deadline.
func awaitDeadline(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(time.Second):
		return errors.New("probe ran unbounded")
	}
}

// TestSubmitCreateBoundsTheProbe pins that a slow inspect cannot decide
// how long a create takes. A yt-dlp probe runs ~10s against a 0.2s
// ordinary round trip, and a channel URL can outrun even the server's
// own deadline — so the probe is capped and the create goes out with no
// name rather than the caller waiting on an optional nicety.
func TestSubmitCreateBoundsTheProbe(t *testing.T) {
	probeCtxErr := make(chan error, 1)
	d := &recordingDoer{
		create: []byte(`{"code":200,"data":{"id":9,"status":"waiting"}}`),
		onInspect: func(ctx context.Context) error {
			err := awaitDeadline(ctx)
			probeCtxErr <- err
			return err
		},
	}

	start := time.Now()
	if _, err := submitCreate(
		context.Background(),
		&preparedClient{doer: d},
		NewDownloadReq{URL: "https://host/slow", App: "wise"},
		nameProbe{url: "https://host/slow", enabled: true, timeout: 20 * time.Millisecond},
	); err != nil {
		t.Fatalf("a probe that runs out of time must not fail the create: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("create waited %s on the probe; the cap did not apply", elapsed)
	}
	if err := <-probeCtxErr; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("probe ended with %v, want a deadline", err)
	}
	if got := d.createReq(t).FileName; got != "" {
		t.Fatalf("file_name = %q, want empty after a timed-out probe", got)
	}
}

// The probe's deadline must not leak into the create request: cancelling
// the create along with the probe would turn a slow inspect into a
// failed task.
func TestSubmitCreateKeepsCreateOutsideTheProbeDeadline(t *testing.T) {
	d := &recordingDoer{
		create:    []byte(`{"code":200,"data":{"id":11,"status":"waiting"}}`),
		onInspect: awaitDeadline,
		onCreate: func(ctx context.Context) error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("create context already cancelled: %w", err)
			}
			return nil
		},
	}
	if _, err := submitCreate(
		context.Background(),
		&preparedClient{doer: d},
		NewDownloadReq{URL: "https://host/slow", App: "wise"},
		nameProbe{url: "https://host/slow", enabled: true, timeout: 10 * time.Millisecond},
	); err != nil {
		t.Fatalf("submitCreate: %v", err)
	}
}

func TestRunCreateValidatesInspectFlags(t *testing.T) {
	cases := []struct {
		name           string
		noInspect      bool
		inspectTimeout time.Duration
		wantIn         string
	}{
		{
			name:           "negative timeout",
			inspectTimeout: -time.Second,
			wantIn:         "unsupported --inspect-timeout",
		},
		{
			name:           "a cap on a probe that will not run",
			noInspect:      true,
			inspectTimeout: time.Second,
			wantIn:         "--inspect-timeout cannot be combined with --no-inspect",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runCreate(
				context.Background(),
				nil,
				"https://example.com/video",
				"", "", "", "", "", "", "", "",
				tc.noInspect,
				tc.inspectTimeout,
				false,
				0,
				"table",
			)
			if err == nil || !strings.Contains(err.Error(), tc.wantIn) {
				t.Fatalf("want an error containing %q, got %v", tc.wantIn, err)
			}
		})
	}
}

func TestSubmitCreateLeavesFileNameAlone(t *testing.T) {
	cases := []struct {
		name         string
		req          NewDownloadReq
		probe        nameProbe
		inspectErr   error
		wantFileName string
		wantProbed   bool
	}{
		{
			name:         "an explicit --name is never overwritten by the probe",
			req:          NewDownloadReq{URL: "https://host/v", App: "wise", FileName: "clip.mp4"},
			probe:        nameProbe{url: "https://host/v", enabled: true},
			wantFileName: "clip.mp4",
		},
		{
			// A cookie-gated probe must not fail the create: the
			// provider still writes the real name back later.
			name:         "a failed probe omits the field and still creates",
			req:          NewDownloadReq{URL: "https://host/v", App: "wise"},
			probe:        nameProbe{url: "https://host/v", enabled: true},
			inspectErr:   errors.New("daemon returned code 505 (network)"),
			wantFileName: "",
			wantProbed:   true,
		},
		{
			name:         "a magnet is not probed and stays nameless",
			req:          NewDownloadReq{URL: "magnet:?xt=urn:btih:abc", App: "wise"},
			probe:        nameProbe{url: "magnet:?xt=urn:btih:abc"},
			wantFileName: "",
		},
		{
			name:         "--no-inspect skips the probe",
			req:          NewDownloadReq{URL: "https://host/v", App: "wise"},
			probe:        nameProbe{url: "https://host/v"},
			wantFileName: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &recordingDoer{
				inspect:    inspectEnvelope(t, "should not be used"),
				inspectErr: tc.inspectErr,
				create:     []byte(`{"code":200,"data":{"id":7,"status":"waiting"}}`),
			}
			if _, err := submitCreate(context.Background(), &preparedClient{doer: d}, tc.req, tc.probe); err != nil {
				t.Fatalf("submitCreate: %v", err)
			}
			if got := d.createReq(t).FileName; got != tc.wantFileName {
				t.Fatalf("file_name = %q, want %q", got, tc.wantFileName)
			}
			if d.probed() != tc.wantProbed {
				t.Fatalf("probed = %v, want %v", d.probed(), tc.wantProbed)
			}
		})
	}
}

// TestRunCreateRejectsNameForRenameLockedProviders pins that the flag
// fails locally rather than being dropped on the way to the server: for
// these providers a custom name only ever reaches the task row, so
// silently accepting it would have the list advertise a filename that
// does not exist on disk.
func TestRunCreateRejectsNameForRenameLockedProviders(t *testing.T) {
	cases := []struct {
		name        string
		rawURL      string
		torrentFile string
		extra       string
		wantIn      string
	}{
		{name: "magnet", rawURL: "magnet:?xt=urn:btih:abc", wantIn: "magnet link"},
		{name: "torrent upload", torrentFile: "./x.torrent", wantIn: "torrent upload"},
		{name: "huggingface", rawURL: "https://huggingface.co/org/repo", wantIn: "HuggingFace repo"},
		{name: "hf cache via extra", rawURL: "https://host/x", extra: `{"_hf_dest":"cache"}`, wantIn: "HuggingFace repo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runCreate(
				context.Background(),
				nil,
				tc.rawURL,
				"",
				"",
				"my-name.mp4",
				"",
				"",
				tc.extra,
				tc.torrentFile,
				"",
				false,
				0,
				false,
				0,
				"table",
			)
			if err == nil {
				t.Fatal("--name should be rejected before any HTTP call")
			}
			if !strings.Contains(err.Error(), "--name is not supported") || !strings.Contains(err.Error(), tc.wantIn) {
				t.Fatalf("error should name the provider, got %v", err)
			}
		})
	}
}

// TestEmitCreatedReportsWriteFailure pins the recovery contract: the task
// id is the only handle a caller has for a resume or cleanup, so a failed
// stdout write must not look like success.
func TestEmitCreatedReportsWriteFailure(t *testing.T) {
	readOnly, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { readOnly.Close() })
	orig := os.Stdout
	os.Stdout = readOnly
	t.Cleanup(func() { os.Stdout = orig })

	for _, format := range []Format{FormatTable, FormatJSON} {
		err := emitCreated(format, DownloadTask{ID: 42, Status: "downloading"})
		if err == nil {
			t.Fatalf("%s: failed stdout write should be reported", format)
		}
		if !strings.Contains(err.Error(), "42") {
			t.Fatalf("%s: error must keep the task id: %v", format, err)
		}
	}
}

func TestReadTorrentFile(t *testing.T) {
	dir := t.TempDir()
	good := dir + "/ok.torrent"
	if err := os.WriteFile(good, []byte("d4:infod4:name1:aee"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readTorrentFile(good, "--torrent"); err != nil {
		t.Fatalf("valid torrent should pass: %v", err)
	}

	yamlPath := dir + "/test.yaml"
	if err := os.WriteFile(yamlPath, []byte("a: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := func() error { _, e := readTorrentFile(yamlPath, "--file"); return e }()
	if err == nil || !strings.Contains(err.Error(), "need a .torrent file") || !strings.Contains(err.Error(), "--file") || strings.Contains(err.Error(), "--torrent") {
		t.Fatalf("non-torrent should fail with --file quote hint, got %v", err)
	}

	empty := dir + "/empty.torrent"
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := readTorrentFile(empty, "--torrent"); err == nil || !strings.Contains(err.Error(), "file is empty") {
		t.Fatalf("empty torrent should fail, got %v", err)
	}

	missing := dir + "/missing with spaces.torrent"
	_, err = readTorrentFile(missing, "--torrent")
	if err == nil || !strings.Contains(err.Error(), "quote it") || !strings.Contains(err.Error(), "--torrent") {
		t.Fatalf("missing torrent should hint quoting, got %v", err)
	}
}
