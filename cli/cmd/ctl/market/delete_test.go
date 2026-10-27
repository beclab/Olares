package market

import (
	"encoding/json"
	"testing"
)

// TestDeletedRowCount covers the three cases runDelete branches on. The
// backend answers a delete of an app it does not hold with HTTP 200 and
// success=true, reporting the work it did in data.deleted_rows, so a zero
// there is the only signal that "deleted" removed nothing.
func TestDeletedRowCount(t *testing.T) {
	cases := []struct {
		name      string
		data      string
		wantCount int
		wantKnown bool
	}{
		{name: "rows removed", data: `{"app_name":"probe","deleted_rows":2}`, wantCount: 2, wantKnown: true},
		{name: "nothing removed", data: `{"app_name":"probe","deleted_rows":0}`, wantCount: 0, wantKnown: true},
		{name: "field absent", data: `{"app_name":"probe"}`, wantKnown: false},
		{name: "field not a number", data: `{"deleted_rows":"two"}`, wantKnown: false},
		{name: "empty data", data: ``, wantKnown: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &APIResponse{Success: true}
			if tc.data != "" {
				resp.Data = json.RawMessage(tc.data)
			}
			count, known := deletedRowCount(resp)
			if known != tc.wantKnown {
				t.Fatalf("known = %v, want %v", known, tc.wantKnown)
			}
			if known && count != tc.wantCount {
				t.Errorf("count = %d, want %d", count, tc.wantCount)
			}
		})
	}

	if _, known := deletedRowCount(nil); known {
		t.Error("a nil response must not report a known row count")
	}
}
