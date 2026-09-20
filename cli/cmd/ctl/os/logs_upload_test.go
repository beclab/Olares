package os

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr swaps os.Stderr for a pipe (never a TTY) and returns whatever
// fn wrote to it.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	w.Close()
	return <-done
}

func TestUploadBodyNonTTYEmitsNoCarriageReturns(t *testing.T) {
	payload := strings.Repeat("x", 512*1024)
	var body io.Reader

	out := captureStderr(t, func() {
		reader, finish := uploadBody(strings.NewReader(payload), int64(len(payload)))
		body = reader
		got, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if len(got) != len(payload) {
			t.Errorf("body length = %d, want %d", len(got), len(payload))
		}
		finish()
	})

	if strings.Contains(out, "\r") {
		t.Errorf("non-TTY stderr contains carriage returns: %q", out)
	}
	if strings.Contains(out, "\033[") {
		t.Errorf("non-TTY stderr contains ANSI escapes: %q", out)
	}
	if out != "uploading log archive...\n" {
		t.Errorf("non-TTY stderr = %q, want single plain status line", out)
	}
	if _, wrapped := body.(*progressReader); wrapped {
		t.Error("non-TTY stderr should not wrap the body in a progressReader")
	}
}

func TestProgressReaderDrawClearsLine(t *testing.T) {
	p := &progressReader{r: strings.NewReader(""), total: 1000, read: 1000, label: "Uploading", last: -1}
	out := captureStderr(t, func() { p.draw(100) })

	if !strings.HasPrefix(out, "\r\033[K") {
		t.Errorf("draw output = %q, want it to start with \\r\\033[K to clear the line", out)
	}
}

func TestProgressReaderFinishNoOpWhenNothingDrawn(t *testing.T) {
	p := &progressReader{r: strings.NewReader(""), total: 0, label: "Uploading", last: -1}
	if out := captureStderr(t, p.finish); out != "" {
		t.Errorf("finish with nothing drawn wrote %q, want no output", out)
	}

	captureStderr(t, func() { p.draw(0) })
	out := captureStderr(t, p.finish)
	if out != "\n" {
		t.Errorf("finish after a draw wrote %q, want a single newline", out)
	}
}
