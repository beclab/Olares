package pod

import (
	"encoding/json"
	"testing"
)

func TestStatusReasonSurfacesAFailingInitContainer(t *testing.T) {
	for _, tc := range []struct {
		status string
		want   string
	}{
		{`{"phase":"Pending","initContainerStatuses":[{"name":"init","state":{"waiting":{"reason":"CrashLoopBackOff"}}}],"containerStatuses":[{"name":"app","state":{"waiting":{"reason":"PodInitializing"}}}]}`, "Init:CrashLoopBackOff"},
		{`{"phase":"Pending","initContainerStatuses":[{"name":"init","state":{"terminated":{"exitCode":1,"reason":"Error"}}}]}`, "Init:Error"},
		{`{"phase":"Running","initContainerStatuses":[{"name":"init","state":{"terminated":{"exitCode":0,"reason":"Completed"}}}],"containerStatuses":[{"name":"app","state":{"waiting":{"reason":"CrashLoopBackOff"}}}]}`, "CrashLoopBackOff"},
		{`{"phase":"Running","initContainerStatuses":[{"name":"init","state":{"terminated":{"exitCode":0,"reason":"Completed"}}}],"containerStatuses":[{"name":"app","state":{"running":{}}}]}`, "Running"},
	} {
		var p Pod
		if err := json.Unmarshal([]byte(`{"status":`+tc.status+`}`), &p); err != nil {
			t.Fatal(err)
		}
		if got := p.statusReason(); got != tc.want {
			t.Errorf("statusReason() = %q, want %q for %s", got, tc.want, tc.status)
		}
	}
}
