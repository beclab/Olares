package clusterexec

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHandshakeErrorCarriesTheServerReason(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   string
	}{
		{http.StatusBadRequest, "container nope is not valid for pod p-1\n", "(HTTP 400): container nope is not valid for pod p-1: the container must exist and be running"},
		{http.StatusInternalServerError, `{"message":"container not running (4f2a)"}`, "(HTTP 500): container not running (4f2a): the container"},
		{http.StatusInternalServerError, "", "(HTTP 500): the container must exist and be running"},
	} {
		resp := &http.Response{StatusCode: tc.status, Body: io.NopCloser(strings.NewReader(tc.body))}
		err := handshakeError(errors.New("bad handshake"), resp)
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("HTTP %d %q: got %q, want it to contain %q", tc.status, tc.body, err, tc.want)
		}
		if strings.Contains(err.Error(), "upgrade Olares") {
			t.Errorf("HTTP %d: the version gate already passed, so the error must not blame the version: %q", tc.status, err)
		}
	}
}
