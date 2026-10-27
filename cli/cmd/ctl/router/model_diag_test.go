package router

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
)

func TestModelDiagNoLongerOffersPerf(t *testing.T) {
	cmd := newModelDiagCommand(&cmdutil.Factory{}, addressByModel)
	for _, child := range cmd.Commands() {
		if child.Name() == "perf" {
			t.Fatal("removed server diagnostic is still exposed")
		}
	}
	if strings.Contains(cmd.Long, "diag perf") {
		t.Error("help still advertises perf")
	}
}

func TestLocalEndpointKeepsAsyncAndIgnoresRetiredInputLimit(t *testing.T) {
	raw := []byte(`{
		"method":"POST",
		"path":"/v1/audio/transcriptions",
		"category":"audio",
		"description":"transcribe",
		"available":true,
		"async_supported":true,
		"max_input_seconds":600
	}`)
	var endpoint localEndpoint
	if err := json.Unmarshal(raw, &endpoint); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(endpoint)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"async_supported":true`)) {
		t.Errorf("JSON dropped async_supported: %s", encoded)
	}
	if bytes.Contains(encoded, []byte("max_input_seconds")) {
		t.Errorf("retired engine-only limit was retained: %s", encoded)
	}

	var out bytes.Buffer
	if err := renderLocalEndpoints(&out, "audio", []localEndpoint{endpoint}, false); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(out.String()), "async") {
		t.Errorf("human output dropped async metadata: %q", out.String())
	}
	if strings.Contains(out.String(), "MAX INPUT") || strings.Contains(out.String(), "10m") {
		t.Errorf("human output retained the retired input limit: %q", out.String())
	}
}

func TestOldLocalEndpointResponseStillDecodes(t *testing.T) {
	var endpoint localEndpoint
	if err := json.Unmarshal([]byte(`{
		"method":"GET",
		"path":"/healthz",
		"category":"ops",
		"description":"health",
		"available":true
	}`), &endpoint); err != nil {
		t.Fatal(err)
	}
	if endpoint.AsyncSupported != nil {
		t.Errorf("missing optional fields became values: %+v", endpoint)
	}
	var out bytes.Buffer
	if err := renderLocalEndpoints(&out, "llamacpp", []localEndpoint{endpoint}, false); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "ASYNC") || strings.Contains(out.String(), "MAX INPUT") {
		t.Errorf("legacy endpoint table grew audio-only columns: %q", out.String())
	}
}
