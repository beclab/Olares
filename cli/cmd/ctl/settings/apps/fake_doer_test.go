package apps

import (
	"context"
	"encoding/json"
	"fmt"
)

// fakeDoer is the package-shared test helper. It records every DoJSON
// call (method + path + body) for assertions and supports a small,
// FIFO-style response queue for verbs that issue more than one wire
// call (e.g. RMW writes that GET then POST).
//
// Tests can either:
//
//   - leave responses empty and let the doer return nil for every call
//     (the body assertion path),
//
//   - queue typed responses with `enqueueRespond([]byte)` to drive
//     reads back through bflEnvelope unwrapping, or
//
//   - set wantErr to fail the next call.
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
	if out == nil {
		return nil
	}
	if len(f.responses) == 0 {
		return nil
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	if len(resp) == 0 {
		return nil
	}
	return json.Unmarshal(resp, out)
}

// enqueueEnvelope wraps `data` in a {code: 0, data: <data>} BFL
// envelope. Useful when faking GET responses that doGetEnvelope
// unwraps.
func (f *fakeDoer) enqueueEnvelope(data interface{}) {
	raw, err := json.Marshal(map[string]interface{}{"code": 0, "data": data})
	if err != nil {
		panic(fmt.Sprintf("enqueueEnvelope: %v", err))
	}
	f.responses = append(f.responses, raw)
}

// enqueueEmptyEnvelope is a {code: 0} envelope with no data field.
// Suitable for write responses where the verb's success message is
// driven entirely by a non-error return.
func (f *fakeDoer) enqueueEmptyEnvelope() {
	f.responses = append(f.responses, []byte(`{"code":0}`))
}

func (f *fakeDoer) lastCall() recordedCall {
	if len(f.calls) == 0 {
		return recordedCall{}
	}
	return f.calls[len(f.calls)-1]
}
