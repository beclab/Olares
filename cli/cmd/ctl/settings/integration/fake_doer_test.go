package integration

import (
	"context"
	"encoding/json"
	"fmt"
)

// fakeDoer records every DoJSON call and answers reads from a FIFO
// queue, so a verb that reads before it writes (cookie import --merge)
// can be driven end to end without a server.
type fakeDoer struct {
	calls     []recordedCall
	responses [][]byte
	wantErr   error
}

type recordedCall struct {
	method string
	path   string
	body   interface{}
}

func (f *fakeDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	f.calls = append(f.calls, recordedCall{method: method, path: path, body: body})
	if f.wantErr != nil {
		return f.wantErr
	}
	if out == nil || len(f.responses) == 0 {
		return nil
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	if len(resp) == 0 {
		return nil
	}
	return json.Unmarshal(resp, out)
}

// enqueueEnvelope wraps data in the {code, message, data} envelope every
// user-service route returns.
func (f *fakeDoer) enqueueEnvelope(data interface{}) {
	raw, err := json.Marshal(map[string]interface{}{"code": 0, "data": data})
	if err != nil {
		panic(fmt.Sprintf("enqueueEnvelope: %v", err))
	}
	f.responses = append(f.responses, raw)
}

func (f *fakeDoer) lastCall() recordedCall {
	if len(f.calls) == 0 {
		return recordedCall{}
	}
	return f.calls[len(f.calls)-1]
}
