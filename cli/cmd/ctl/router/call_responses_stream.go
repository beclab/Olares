package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

// GET /v1/responses (WebSocket)
//
// The same Responses call over a socket instead of a POST. It exists on Router
// because a client that wants the answer as it is written needs somewhere to
// receive it, and the HTTP route answers only when the whole thing is done.
//
// One socket, one request here. Router's protocol allows several `response.create`
// frames down one connection and holds it open for an hour, which is a session
// — the same thing `call responses` declines to open over HTTP, for the same
// reason: a conversation spread across invocations is a design, not a flag.
// What this adds is watching one long answer arrive.
//
// The frames are the Responses event stream, unchanged: `response.output_text.delta`
// carries the text a piece at a time, the reasoning deltas carry the thinking,
// and one of `response.completed` / `response.incomplete` / `response.failed`
// ends it. An `error` frame is Router refusing the request itself, before any
// model saw it.

// responsesEvent is the part of a server frame this verb reads. Everything
// else in the event — item ids, part indices, annotations — describes structure
// a streaming reader does not rebuild.
type responsesEvent struct {
	Type     string           `json:"type"`
	Delta    string           `json:"delta,omitempty"`
	Response *responsesAnswer `json:"response,omitempty"`
	Error    *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func runResponsesStream(ctx context.Context, f *cmdutil.Factory, input string,
	opts responsesOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}
	format, err := parseFormat(opts.OutputIn)
	if err != nil {
		return err
	}
	pc, err := prepare(ctx, f)
	if err != nil {
		return err
	}
	token, err := f.ValidAccessToken(ctx)
	if err != nil {
		return err
	}
	conn, err := dialRouterSocket(ctx, pc, token, epResponses, strings.TrimSpace(opts.Model), opts.APIKey)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	create := map[string]any{
		"type": "response.create",
		"response": responsesRequest{
			Model:           strings.TrimSpace(opts.Model),
			Input:           input,
			Instructions:    strings.TrimSpace(opts.Instructions),
			MaxOutputTokens: opts.MaxTokens,
		},
	}
	frame, err := json.Marshal(create)
	if err != nil {
		return err
	}
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		return fmt.Errorf("send the request: %w", err)
	}
	answer, err := readResponsesStream(conn, os.Stdout, os.Stderr, format)
	// The close is sent whatever happened: Router settles the spend row when
	// the connection ends, and a socket left to time out settles it a minute
	// late for no reason.
	_ = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}
	if format == FormatJSON {
		return printJSON(os.Stdout, answer)
	}
	if answer.Error != nil && strings.TrimSpace(answer.Error.Message) != "" {
		return fmt.Errorf("%s: %s", nonEmpty(answer.Error.Code), answer.Error.Message)
	}
	if opts.Quiet {
		return nil
	}
	return printResponsesFooter(answer)
}

// readResponsesStream prints the text as it arrives and returns the terminal
// response object. Text goes to stdout so it stays pipeable, and reasoning to
// stderr, which is what `call chat --stream` does with the same two things.
//
// Under -o json nothing is printed while it streams: a document assembled from
// deltas is not the document Router sends, and the terminal frame carries the
// real one.
func readResponsesStream(conn *websocket.Conn, out, progress io.Writer,
	format Format) (*responsesAnswer, error) {
	streaming := format != FormatJSON
	wroteText := false
	inReasoning := false
	for {
		kind, payload, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure,
				websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				return nil, fmt.Errorf("the connection closed before the answer was finished")
			}
			return nil, fmt.Errorf("read the answer: %w", err)
		}
		if kind != websocket.TextMessage {
			continue
		}
		var event responsesEvent
		if json.Unmarshal(payload, &event) != nil {
			continue
		}
		switch event.Type {
		case "response.output_text.delta":
			if !streaming || event.Delta == "" {
				continue
			}
			if inReasoning {
				if _, err := fmt.Fprintln(progress); err != nil {
					return nil, err
				}
				inReasoning = false
			}
			wroteText = true
			if _, err := io.WriteString(out, event.Delta); err != nil {
				return nil, err
			}
		case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
			if !streaming || event.Delta == "" {
				continue
			}
			if !inReasoning {
				if _, err := fmt.Fprint(progress, "[reasoning] "); err != nil {
					return nil, err
				}
				inReasoning = true
			}
			if _, err := io.WriteString(progress, event.Delta); err != nil {
				return nil, err
			}
		case "response.completed", "response.incomplete", "response.failed":
			if streaming && wroteText {
				if _, err := fmt.Fprintln(out); err != nil {
					return nil, err
				}
			}
			if event.Response == nil {
				return nil, fmt.Errorf("%s arrived without the response it ends", event.Type)
			}
			if streaming && !wroteText {
				if err := renderResponses(out, event.Response, true); err != nil {
					return nil, err
				}
			}
			return event.Response, nil
		case "error":
			// Router's own refusal, before the model. It is not a response
			// object, so there is nothing to fall through to.
			if event.Error == nil {
				return nil, fmt.Errorf("the request was refused without a reason")
			}
			return nil, fmt.Errorf("%s: %s", nonEmpty(event.Error.Code), event.Error.Message)
		}
	}
}
