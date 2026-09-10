package router

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// A completion is the one verb in this tree that can be minutes away through
// no fault of the prompt, and it was the only long-running one still on the
// 30-second client. That made the caller the first link on the chain to give
// up, on a chain where nothing behind it stops when it does.
func TestChatAllowsALongWaitAndSaysHowLong(t *testing.T) {
	cmd := newCallChatCommand(nil)
	flag := cmd.Flag("timeout")
	if flag == nil {
		t.Fatal("chat has no --timeout")
	}
	if flag.DefValue != (10 * time.Minute).String() {
		t.Errorf("default timeout = %s, want 10m0s", flag.DefValue)
	}
	// Giving up here does not recall the request, and a user who believes it
	// does will retry into a queue they are already in.
	for _, want := range []string{"--timeout", "queue"} {
		if !strings.Contains(cmd.Long, want) {
			t.Errorf("chat help does not mention %q", want)
		}
	}
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestSilenceIsReportedAndAnAnswerStopsIt(t *testing.T) {
	var spoke syncBuffer
	stop := noticeWhileWaiting(context.Background(), &spoke, time.Millisecond, 10*time.Minute)
	deadline := time.Now().Add(2 * time.Second)
	for spoke.String() == "" && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	stop()
	if !strings.Contains(spoke.String(), "waiting for the model") {
		t.Errorf("nothing was said about the wait: %q", spoke.String())
	}
	// The budget belongs in the line: "still waiting" alone leaves the reader
	// with the question they already had.
	if !strings.Contains(spoke.String(), "10m0s") {
		t.Errorf("the line does not say how long the wait may last: %q", spoke.String())
	}

	var quick syncBuffer
	stopFast := noticeWhileWaiting(context.Background(), &quick, time.Hour, time.Minute)
	stopFast()
	stopFast() // idempotent: the streaming path calls it on every first token
	time.Sleep(5 * time.Millisecond)
	if quick.String() != "" {
		t.Errorf("an answered request still printed a waiting line: %q", quick.String())
	}
}

// --quiet already suppresses the token footer, which is the other thing this
// command writes to stderr.
func TestQuietSaysNothingAboutWaiting(t *testing.T) {
	if got := noticeTarget(true); got == nil {
		t.Fatal("nil writer")
	}
	var loud syncBuffer
	stop := noticeWhileWaiting(context.Background(), noticeTarget(true), time.Millisecond, time.Minute)
	defer stop()
	time.Sleep(10 * time.Millisecond)
	if loud.String() != "" {
		t.Error("quiet wrote somewhere it should not have")
	}
}

func TestGivingUpNamesTheDeadlineAndSaysTheWorkContinues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	<-ctx.Done()

	err := chatWaitErr(ctx, context.DeadlineExceeded, 90*time.Second)
	if err == nil {
		t.Fatal("no error")
	}
	for _, want := range []string{"1m30s", "not", "withdrawn", "--timeout"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%q does not contain %q", err.Error(), want)
		}
	}

	// A failure that is not the deadline keeps its own words.
	live, cancelLive := context.WithCancel(context.Background())
	defer cancelLive()
	if got := chatWaitErr(live, context.Canceled, time.Minute); got != context.Canceled {
		t.Errorf("a non-deadline failure was rewritten: %v", got)
	}
	if got := chatWaitErr(ctx, nil, time.Minute); got != nil {
		t.Errorf("success was turned into %v", got)
	}
}

func TestRetryAfterIsCarriedOffTheHeader(t *testing.T) {
	seconds := http.Header{"Retry-After": []string{"3"}}
	err := withRetryAfter(&RouterError{Status: 503, Code: "model_at_capacity"}, seconds)
	if got := routerErrorOf(err).RetryAfter; got != 3*time.Second {
		t.Errorf("seconds: got %s want 3s", got)
	}

	// RFC 9110 allows a date, and a proxy on the way may rewrite the count as
	// one. A header nobody can read is the same as no header.
	date := http.Header{"Retry-After": []string{time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)}}
	if got := routerErrorOf(withRetryAfter(&RouterError{Status: 503}, date)).RetryAfter; got <= 0 {
		t.Errorf("date: got %s, want a positive wait", got)
	}

	for _, h := range []http.Header{
		{},
		{"Retry-After": []string{"soon"}},
		{"Retry-After": []string{"-5"}},
		{"Retry-After": []string{time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)}},
	} {
		if got := routerErrorOf(withRetryAfter(&RouterError{Status: 503}, h)).RetryAfter; got != 0 {
			t.Errorf("header %v produced %s, want no wait", h, got)
		}
	}

	// Nothing to attach it to, and nothing to panic over.
	if err := withRetryAfter(nil, seconds); err != nil {
		t.Errorf("nil error became %v", err)
	}
}

// A full engine and an exhausted budget both stop a call, and the two have
// opposite remedies: one is waited out, the other is a number somebody raises.
// Router separates them by status and code; this is where that separation
// becomes something a person reads.
//
// Router now holds the caller in a soft queue before answering this, so the
// sentence has to say the wait already happened. Told it was "refused rather
// than queued", someone waits and retries into the same wall.
func TestAFullEngineReadsAsAWaitRatherThanALimit(t *testing.T) {
	err := callErr(&RouterError{
		Status: 503, Type: "upstream_error", Code: "model_at_capacity",
		Message:    "this model is already serving the 1 requests it was configured for",
		RetryAfter: time.Second,
	})
	msg := err.Error()
	for _, want := range []string{"Router waited for a slot", "1s", "provider get"} {
		if !strings.Contains(msg, want) {
			t.Errorf("%q does not contain %q", msg, want)
		}
	}
	if strings.Contains(msg, "router quota list") {
		t.Error("a saturated engine was explained as a quota; no quota was reached")
	}

	// Without the header the sentence has to stop rather than guess.
	bare := callErr(&RouterError{Status: 503, Code: "model_at_capacity"}).Error()
	if strings.Contains(bare, "Router asks for") {
		t.Errorf("a wait was invented with no Retry-After: %q", bare)
	}

	// The pre-existing 502/503/504 branch must not swallow the new code: it
	// says the endpoint did not answer, and this one answered clearly.
	generic := callErr(&RouterError{Status: 503, Code: "upstream_unreachable"}).Error()
	if !strings.Contains(generic, "did not answer") {
		t.Errorf("an unreachable upstream lost its explanation: %q", generic)
	}
}
