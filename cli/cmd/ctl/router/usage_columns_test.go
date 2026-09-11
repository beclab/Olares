package router

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// One row of every non-token shape Router writes. The point is the pointers: a
// quantity that arrived and a quantity that did not are different facts, and a
// struct that decoded both as 0 would report unmeasured audio as free.
const liveSpendRow = `{
  "id": 4711,
  "attempt_id": "att_01HZX",
  "request_id": "req_01HZX",
  "model_name": "FlowStudio/wan2.2",
  "mode": "video_generation",
  "op": "generate",
  "prompt_tokens": 0,
  "completion_tokens": 0,
  "total_tokens": 0,
  "cost_usd": 0.42,
  "status": "success",
  "http_status": 200,
  "latency_ms": 91234,
  "job_ms": 273000,
  "ttft_ms": 800,
  "queue_ms": 120,
  "streamed": false,
  "videos": 1,
  "video_seconds": 5.5,
  "session_id": "task-4711",
  "tags": ["unpriced"],
  "created_at": "2026-08-27T10:00:00Z"
}`

func TestASpendRowDecodesTheColumnsRouterAnswersWith(t *testing.T) {
	var it spendLog
	if err := json.Unmarshal([]byte(liveSpendRow), &it); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if it.Op == nil || *it.Op != "generate" {
		t.Fatalf("op did not decode: %+v", it.Op)
	}
	if it.SessionID == nil || *it.SessionID != "task-4711" {
		t.Fatalf("session_id did not decode: %+v", it.SessionID)
	}
	if it.TTFTMS == nil || *it.TTFTMS != 800 {
		t.Fatalf("ttft_ms did not decode: %+v", it.TTFTMS)
	}
	if it.QueueMS == nil || *it.QueueMS != 120 {
		t.Fatalf("queue_ms did not decode: %+v", it.QueueMS)
	}
	if it.JobMS == nil || *it.JobMS != 273000 {
		t.Fatalf("job_ms did not decode: %+v", it.JobMS)
	}
	if it.Videos == nil || *it.Videos != 1 {
		t.Fatalf("videos did not decode: %+v", it.Videos)
	}
	if it.VideoSeconds == nil || *it.VideoSeconds != 5.5 {
		t.Fatalf("video_seconds did not decode: %+v", it.VideoSeconds)
	}
	if it.AudioInputSeconds != nil {
		t.Fatal("a column Router did not send must stay absent, not become zero")
	}
}

// The MODE column has to say which operation was asked for. A mode alone reads
// the same for a picture that was made and one that was edited.
func TestTheOperationIsShownBesideTheMode(t *testing.T) {
	edit := "edit"
	it := &spendLog{Mode: "image_generation", Op: &edit}
	if got := spendOp(it); got != "image_generation/edit" {
		t.Fatalf("got %q, want the operation beside the mode", got)
	}
	generate := "image_generation"
	if got := spendOp(&spendLog{Mode: "image_generation", Op: &generate}); got != "image_generation" {
		t.Fatalf("an operation equal to the mode should not be repeated, got %q", got)
	}
	if got := spendOp(&spendLog{Mode: "chat"}); got != "chat" {
		t.Fatalf("got %q, want the bare mode", got)
	}
}

func TestEachRowReportsTheQuantityItWasPricedBy(t *testing.T) {
	images := int64(2)
	videos := int64(1)
	videoSeconds := 5.5
	audioIn := 12.5
	audioOut := 0.0
	queries := int64(3)
	pages := int64(7)
	objects := int64(1)

	cases := map[string]struct {
		row  spendLog
		want string
	}{
		"tokens are the fallback":    {spendLog{Mode: "chat", TotalTokens: 1200}, "1200 tok"},
		"pictures":                   {spendLog{Mode: "image_generation", Images: &images}, "2 img"},
		"a video and its length":     {spendLog{Mode: "video_generation", Videos: &videos, VideoSeconds: &videoSeconds}, "1 vid 5.5s"},
		"audio in":                   {spendLog{Mode: "audio", AudioInputSeconds: &audioIn}, "12.5s in"},
		"a measured zero is a zero":  {spendLog{Mode: "audio", AudioOutputSeconds: &audioOut}, "0s out"},
		"queries":                    {spendLog{Mode: "search", Queries: &queries}, "3 q"},
		"OCR pages":                  {spendLog{Mode: "ocr", Pages: &pages}, "7 pg"},
		"3D objects":                 {spendLog{Mode: "model3d_generation", Objects: &objects}, "1 obj"},
		"nothing measured is a dash": {spendLog{Mode: "music_generation"}, "-"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := spendQuantity(&tc.row); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// A row is written when the call is admitted and priced when it ends. Showing $0
// in between would be read as free, which is the one thing it is not.
func TestARunningCallHasNoCostYet(t *testing.T) {
	if got := spendCost(&spendLog{Status: statusInProgress}); got != "pending" {
		t.Fatalf("got %q, want pending", got)
	}
	if got := spendCost(&spendLog{Status: "success", CostUSD: 0}); got != "$0" {
		t.Fatalf("a settled free call really is $0, got %q", got)
	}
}

// A row records the name the caller wrote, and a call that named no model wrote
// a category. Reading `default-tts-clone` off a bill tells nobody which engine
// spent the time.
func TestTheModelColumnNamesWhatAnswered(t *testing.T) {
	cases := map[string]struct {
		row  spendLog
		want string
	}{
		"a category resolves to the model behind it": {
			spendLog{ModelName: "default-tts-clone", ServedModelName: "Olares/Breeze-TTS-2"},
			"Olares/Breeze-TTS-2",
		},
		"a named model is shown as named": {
			spendLog{ModelName: "Olares/Breeze-TTS-2", ServedModelName: "Olares/Breeze-TTS-2"},
			"Olares/Breeze-TTS-2",
		},
		"a refusal that reached no model keeps the caller's own name": {
			spendLog{ModelName: "default-rerank"},
			"default-rerank",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := spendModel(&tc.row); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// The table is one line per call, so the substitution has to be said somewhere:
// a reader who cannot find the name they typed would take the page for the
// wrong account.
func TestASubstitutedCategoryIsReportedOnce(t *testing.T) {
	var buf bytes.Buffer
	rows := []spendLog{
		{Mode: "tts", Status: "success", ModelName: "default-tts-clone",
			ServedModelName: "Olares/Breeze-TTS-2"},
		{Mode: "chat", Status: "success", ModelName: "Olares/Qwen3-8B",
			ServedModelName: "Olares/Qwen3-8B"},
	}
	if err := spendZeroNotes(&buf, rows); err != nil {
		t.Fatalf("spendZeroNotes: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "1 call named a category") {
		t.Fatalf("the notes should count the one substituted row, got:\n%s", out)
	}
	if strings.Count(out, "named a category") != 1 {
		t.Fatalf("the note belongs on the page once, got:\n%s", out)
	}
}

// Four different facts share one glyph in the money column, so the page says
// which of them it is showing.
func TestTheZerosOnThePageAreExplained(t *testing.T) {
	rows := []spendLog{
		{Mode: "chat", Status: statusInProgress},
		{Mode: "music_generation", Status: "success", Tags: []string{"unpriced"}},
		{Mode: "audio", Status: "success", Tags: []string{"audio_unmetered"}},
	}
	var buf bytes.Buffer
	if err := spendZeroNotes(&buf, rows); err != nil {
		t.Fatalf("spendZeroNotes: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"still running", "unpriced", "audio_unmetered"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the notes should mention %q, got:\n%s", want, out)
		}
	}
	var quiet bytes.Buffer
	if err := spendZeroNotes(&quiet, []spendLog{{Mode: "chat", Status: "success", CostUSD: 1}}); err != nil {
		t.Fatalf("spendZeroNotes: %v", err)
	}
	if quiet.String() != "" {
		t.Fatalf("a page with nothing to explain should say nothing, got:\n%s", quiet.String())
	}
}

// Fixed token columns would report a window of image work as three zeros beside
// a real cost, which reads as broken arithmetic rather than another unit.
func TestASummaryShowsTheColumnsItsBucketsCarry(t *testing.T) {
	var buf bytes.Buffer
	items := []spendSummaryRow{
		{Key: "m", Label: "wan2.2", Requests: 4, CostUSD: 1.2, Images: 8, VideoSeconds: 20, Videos: 4},
	}
	if err := renderSummaryBuckets(&buf, "model", items); err != nil {
		t.Fatalf("renderSummaryBuckets: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"IMAGES", "VIDEOS", "VIDEO S"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing the %s column in:\n%s", want, out)
		}
	}
	if strings.Contains(out, "TOKENS") {
		t.Fatalf("no bucket carried a token, so the column should be gone:\n%s", out)
	}

	var work bytes.Buffer
	if err := renderSummaryBuckets(&work, "model", []spendSummaryRow{
		{Key: "m", Requests: 3, Pages: 12, Objects: 2},
	}); err != nil {
		t.Fatalf("renderSummaryBuckets: %v", err)
	}
	for _, want := range []string{"PAGES", "OBJECTS"} {
		if !strings.Contains(work.String(), want) {
			t.Fatalf("missing the %s column in:\n%s", want, work.String())
		}
	}

	var tokens bytes.Buffer
	if err := renderSummaryBuckets(&tokens, "model", []spendSummaryRow{
		{Key: "m", Requests: 2, TotalTokens: 100, PromptTokens: 60, CompletionTok: 40},
	}); err != nil {
		t.Fatalf("renderSummaryBuckets: %v", err)
	}
	if !strings.Contains(tokens.String(), "TOKENS") {
		t.Fatalf("a token bucket should keep its columns:\n%s", tokens.String())
	}
	if strings.Contains(tokens.String(), "IMAGES") {
		t.Fatalf("nothing made a picture here:\n%s", tokens.String())
	}
}

func TestTheTotalsNameEveryQuantityThatArrived(t *testing.T) {
	var buf bytes.Buffer
	err := renderSummaryTotals(&buf, &spendTotals{
		TotalRequests: 9, TotalSuccessRequests: 9, TotalCostUSD: 2,
		TotalAudioSeconds: 61.5, TotalImages: 3, TotalVideos: 1, TotalVideoSeconds: 5,
		TotalQueries: 4, TotalPages: 12, TotalObjects: 2,
	})
	if err != nil {
		t.Fatalf("renderSummaryTotals: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"61.5s of audio", "3 images", "1 video", "5s of video", "4 queries",
		"12 pages", "2 3D objects",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("the totals should state %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "tokens,") {
		t.Fatalf("no call here was priced by the token:\n%s", out)
	}
}

// A window with nothing measurable in it still reports tokens, so the shape of
// the line does not depend on what happened to be measured.
func TestTotalsWithNothingMeasuredStillReportTokens(t *testing.T) {
	var buf bytes.Buffer
	if err := renderSummaryTotals(&buf, &spendTotals{TotalRequests: 1, TotalSuccessRequests: 0}); err != nil {
		t.Fatalf("renderSummaryTotals: %v", err)
	}
	if !strings.Contains(buf.String(), "0 tokens") {
		t.Fatalf("expected a token figure, got:\n%s", buf.String())
	}
}

func TestTheNewFiltersReachRouterUnderItsOwnNames(t *testing.T) {
	q, err := resolveSpendQuery(context.Background(), nil, spendFilter{
		Status: "IN_PROGRESS", Mode: "Image_Generation", SessionID: "task-1",
		SortBy: "cost_usd", SortOrder: "ASC",
	})
	if err != nil {
		t.Fatalf("resolveSpendQuery: %v", err)
	}
	want := map[string]string{
		"status":     "in_progress",
		"mode":       "image_generation",
		"session_id": "task-1",
		"sort_by":    "cost_usd",
		"sort_order": "asc",
	}
	for key, value := range want {
		if got := q.Get(key); got != value {
			t.Fatalf("%s: got %q, want %q", key, got, value)
		}
	}
}

// Speech synthesis is its own mode, and this list is a gate rather than a hint:
// a mode missing from it is a mode nobody can ask about from here, however
// happily Router would answer.
func TestEveryModeRouterWritesCanBeAskedFor(t *testing.T) {
	for _, mode := range spendModes {
		q, err := resolveSpendQuery(context.Background(), nil, spendFilter{Mode: mode})
		if err != nil {
			t.Fatalf("--mode %s: %v", mode, err)
		}
		if got := q.Get("mode"); got != mode {
			t.Fatalf("--mode %s reached Router as %q", mode, got)
		}
	}
	if !containsString(spendModes, "tts") {
		t.Fatal("speech synthesis is its own mode since ADR-63; without it no TTS call can be filtered for")
	}
}

// One capability is not one mode: the ElevenLabs-shaped routes are `tts` and
// the rest of the audio work stayed `audio`, so "what did audio cost me" needs
// both at once.
func TestSeveralModesTravelInOneRequest(t *testing.T) {
	q, err := resolveSpendQuery(context.Background(), nil, spendFilter{Mode: " Audio , TTS ,audio"})
	if err != nil {
		t.Fatalf("resolveSpendQuery: %v", err)
	}
	if got := q.Get("mode"); got != "audio,tts" {
		t.Fatalf("got %q, want the modes normalized and deduplicated", got)
	}
}

func TestASummaryCanBeRankedByAQuantityRatherThanCost(t *testing.T) {
	q, err := resolveSpendQuery(context.Background(), nil, spendFilter{RankBy: "Requests"})
	if err != nil {
		t.Fatalf("resolveSpendQuery: %v", err)
	}
	if got := q.Get("rank_by"); got != "requests" {
		t.Fatalf("rank_by: got %q, want requests", got)
	}
}

// An async call is opened by one request and settled by another, and the
// settling request is not what took the time.
func TestTheDurationShownIsTheWorkRatherThanTheRequest(t *testing.T) {
	job := int64(273000)
	if got := spendDuration(&spendLog{LatencyMS: 5, JobMS: &job}); got != "4m33s" {
		t.Fatalf("got %q, want the job duration", got)
	}
	if got := spendDuration(&spendLog{LatencyMS: 843}); got != "843ms" {
		t.Fatalf("a synchronous call is its own request, got %q", got)
	}
	if got := spendDuration(&spendLog{LatencyMS: 12400}); got != "12.4s" {
		t.Fatalf("got %q, want seconds", got)
	}
}

func TestAValueRouterWouldRefuseIsRefusedHere(t *testing.T) {
	cases := map[string]spendFilter{
		"a status that is not one":  {Status: "pending"},
		"a mode that is not one":    {Mode: "chatting"},
		"a mode list of nothing":    {Mode: " , "},
		"a column nobody sorts by":  {SortBy: "user_agent"},
		"an order that is not one":  {SortOrder: "sideways"},
		"a ranking that is not one": {RankBy: "latency"},
	}
	for name, fl := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := resolveSpendQuery(context.Background(), nil, fl); err == nil {
				t.Fatal("expected a refusal before the request")
			}
		})
	}
}
