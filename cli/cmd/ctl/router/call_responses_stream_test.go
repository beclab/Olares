package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// serveResponsesFrames plays a fixed script of server frames down one socket
// and hands back a client connection reading it.
func serveResponsesFrames(t *testing.T, frames []string) *websocket.Conn {
	t.Helper()
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		for _, frame := range frames {
			if err := conn.WriteMessage(websocket.TextMessage, []byte(frame)); err != nil {
				return
			}
		}
		// Held open: a reader that has seen a terminal frame must not need the
		// close to know the answer ended.
		_, _, _ = conn.ReadMessage()
	}))
	t.Cleanup(server.Close)
	client, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatalf("dial the test socket: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// Text goes to stdout and reasoning to stderr, so a pipe receives the answer
// and nothing else. This is what `call chat` does with the same two things.
func TestAStreamedAnswerSeparatesTheThinkingFromTheAnswer(t *testing.T) {
	conn := serveResponsesFrames(t, []string{
		`{"type":"response.created","sequence_number":1}`,
		`{"type":"response.reasoning_summary_text.delta","delta":"weighing "}`,
		`{"type":"response.reasoning_summary_text.delta","delta":"options"}`,
		`{"type":"response.output_text.delta","delta":"Consensus "}`,
		`{"type":"response.output_text.delta","delta":"is agreement."}`,
		`{"type":"response.completed","response":{"id":"resp_1","model":"gpt-5",` +
			`"status":"completed","usage":{"input_tokens":9,"output_tokens":4,"total_tokens":13}}}`,
	})
	var out, progress bytes.Buffer
	answer, err := readResponsesStream(conn, &out, &progress, FormatTable)
	if err != nil {
		t.Fatalf("the stream did not finish: %v", err)
	}
	if got := out.String(); got != "Consensus is agreement.\n" {
		t.Errorf("stdout carries more than the answer: %q", got)
	}
	if got := progress.String(); !strings.Contains(got, "weighing options") {
		t.Errorf("the reasoning did not reach stderr: %q", got)
	}
	if answer == nil || answer.ID != "resp_1" || answer.Usage == nil || answer.Usage.TotalTokens != 13 {
		t.Errorf("the terminal frame's response did not come back: %+v", answer)
	}
}

// Under -o json the deltas are not printed, because a document assembled from
// them is not the document Router sends. The terminal frame carries the real
// one, and that is what is returned.
func TestUnderJSONNothingIsPrintedWhileItStreams(t *testing.T) {
	conn := serveResponsesFrames(t, []string{
		`{"type":"response.output_text.delta","delta":"half an answer"}`,
		`{"type":"response.completed","response":{"id":"resp_2","status":"completed"}}`,
	})
	var out, progress bytes.Buffer
	answer, err := readResponsesStream(conn, &out, &progress, FormatJSON)
	if err != nil {
		t.Fatalf("the stream did not finish: %v", err)
	}
	if out.Len() != 0 || progress.Len() != 0 {
		t.Errorf("something was printed under -o json: %q / %q", out.String(), progress.String())
	}
	if answer == nil || answer.ID != "resp_2" {
		t.Errorf("the response did not come back: %+v", answer)
	}
}

// An `error` frame is Router refusing the request before a model saw it, so
// there is no response object to fall through to and no partial answer to
// keep. It has to surface as the failure it is.
func TestARefusalOnTheSocketIsAFailureRatherThanAnEmptyAnswer(t *testing.T) {
	conn := serveResponsesFrames(t, []string{
		`{"type":"error","error":{"code":"model_not_allowed","message":"the key does not allow it"}}`,
	})
	var out, progress bytes.Buffer
	_, err := readResponsesStream(conn, &out, &progress, FormatTable)
	if err == nil {
		t.Fatal("a refusal was read as an answer")
	}
	if !strings.Contains(err.Error(), "model_not_allowed") ||
		!strings.Contains(err.Error(), "the key does not allow it") {
		t.Errorf("the refusal lost its reason: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("a refusal printed something on stdout: %q", out.String())
	}
}

// `response.failed` and `response.incomplete` end the stream the same way
// `response.completed` does. Waiting past them is waiting for a frame that is
// not coming.
func TestAnAnswerThatDidNotFinishStillEndsTheStream(t *testing.T) {
	for _, terminal := range []string{"response.failed", "response.incomplete"} {
		conn := serveResponsesFrames(t, []string{
			`{"type":"` + terminal + `","response":{"id":"resp_3","status":"incomplete",` +
				`"incomplete_details":{"reason":"max_output_tokens"}}}`,
		})
		var out, progress bytes.Buffer
		answer, err := readResponsesStream(conn, &out, &progress, FormatTable)
		if err != nil {
			t.Fatalf("%s did not end the stream: %v", terminal, err)
		}
		if answer == nil || answer.Status != "incomplete" {
			t.Errorf("%s: the response did not come back: %+v", terminal, answer)
		}
	}
}

// A socket that closes before a terminal frame is not an empty answer: the
// answer is unfinished and unknown, and printing nothing would read as the
// model having said nothing.
func TestASocketThatClosesEarlyIsNotAnEmptyAnswer(t *testing.T) {
	conn := serveResponsesFrames(t, nil)
	// The scripted server holds the socket open on a read; closing from here
	// is what ends it, which is the shape of a connection dropped mid-answer.
	_ = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	var out, progress bytes.Buffer
	_, err := readResponsesStream(conn, &out, &progress, FormatTable)
	if err == nil {
		t.Fatal("a connection that closed mid-answer was read as a finished one")
	}
}

// Both the streaming route and the one-shot one are the same path. The
// difference is the method and the upgrade, which is why there is one
// constant.
func TestTheStreamRunsOnTheSamePathAsThePost(t *testing.T) {
	if epResponses != dataPlaneAPI+"/responses" {
		t.Errorf("the responses route is spelled %q", epResponses)
	}
	got, err := routerSocketURL("https://olares.example", epResponses, "gpt-5")
	if err != nil {
		t.Fatal(err)
	}
	if got != "wss://olares.example"+epResponses+"?model=gpt-5" {
		t.Errorf("the socket URL is %q", got)
	}
}
