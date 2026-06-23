package market

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// failedCheckResp builds the market-wrapped FailedCheckResponse envelope the
// CLI receives: APIResponse.Data carries data.backend_response.data.{type, Data},
// the exact shape parseServerEnvError / parseFailedCheck unwrap.
func failedCheckResp(checkType, dataJSON string) *APIResponse {
	inner := fmt.Sprintf(`{"backend_response":{"code":422,"data":{"type":%q,"Data":%s}}}`, checkType, dataJSON)
	return &APIResponse{Data: json.RawMessage(inner)}
}

func TestInstallRequestWireFormat(t *testing.T) {
	// Without a compute mode the field is omitted entirely (1.12.5 wire
	// stays byte-identical to before SelectedGpuType existed).
	b, _ := json.Marshal(InstallRequest{Source: "market.olares", AppName: "firefox", Version: "1.0.0", Sync: true})
	if strings.Contains(string(b), "selectedGpuType") {
		t.Fatalf("install body must omit selectedGpuType when empty: %s", b)
	}
	// With a mode the field is present.
	b, _ = json.Marshal(InstallRequest{Source: "market.olares", AppName: "comfyui", Version: "1.0.0", Sync: true, SelectedGpuType: "nvidia"})
	if !strings.Contains(string(b), `"selectedGpuType":"nvidia"`) {
		t.Fatalf("install body must carry selectedGpuType when set: %s", b)
	}
}

func TestParseFailedCheck(t *testing.T) {
	checkType, raw := parseFailedCheck(failedCheckResp(checkTypeComputeModeSelect,
		`[{"computeType":"cpu","status":"installable"}]`))
	if checkType != checkTypeComputeModeSelect {
		t.Fatalf("checkType = %q, want %q", checkType, checkTypeComputeModeSelect)
	}
	var plan []computeModePlan
	if err := json.Unmarshal(raw, &plan); err != nil || len(plan) != 1 || plan[0].ComputeType != "cpu" {
		t.Fatalf("unexpected plan from raw %s (err=%v): %#v", raw, err, plan)
	}

	// A non-failed-check response yields an empty type.
	if ct, _ := parseFailedCheck(&APIResponse{Data: json.RawMessage(`{"app_name":"firefox"}`)}); ct != "" {
		t.Fatalf("expected empty checkType for non-failed-check, got %q", ct)
	}
	if ct, _ := parseFailedCheck(nil); ct != "" {
		t.Fatalf("expected empty checkType for nil resp, got %q", ct)
	}
}

func TestResolveComputeMode(t *testing.T) {
	twoInstallable := `[{"computeType":"cpu","status":"installable"},{"computeType":"nvidia","status":"installable"}]`
	_, raw := parseFailedCheck(failedCheckResp(checkTypeComputeModeSelect, twoInstallable))

	// preset that is installable -> returned verbatim.
	if mode, err := resolveComputeMode(raw, "comfyui", "nvidia", false); err != nil || mode != "nvidia" {
		t.Fatalf("preset nvidia: got (%q, %v), want (nvidia, nil)", mode, err)
	}

	// preset declared but not installable -> error explaining why.
	_, raw2 := parseFailedCheck(failedCheckResp(checkTypeComputeModeSelect,
		`[{"computeType":"cpu","status":"installable"},{"computeType":"nvidia","status":"insufficient-resources","reason":"not enough VRAM"}]`))
	_, err := resolveComputeMode(raw2, "comfyui", "nvidia", false)
	if err == nil || !strings.Contains(err.Error(), "not enough VRAM") {
		t.Fatalf("preset non-installable: want error mentioning reason, got %v", err)
	}

	// preset not a declared mode -> error.
	if _, err := resolveComputeMode(raw, "comfyui", "amd", false); err == nil || !strings.Contains(err.Error(), "not a declared mode") {
		t.Fatalf("undeclared preset: want 'not a declared mode' error, got %v", err)
	}

	// no preset, non-interactive -> typed computeModeSelectError listing modes.
	_, err = resolveComputeMode(raw, "comfyui", "", false)
	var modeErr *computeModeSelectError
	if !errors.As(err, &modeErr) {
		t.Fatalf("non-interactive empty preset: want *computeModeSelectError, got %T (%v)", err, err)
	}
	if !reflect.DeepEqual(modeErr.installable, []string{"cpu", "nvidia"}) {
		t.Fatalf("installable = %v, want [cpu nvidia]", modeErr.installable)
	}
}

func TestResolveComputeBindingNonInteractive(t *testing.T) {
	data := `{"availability":{"scope":"card","nodes":[{"nodeName":"node-1","gpuType":"nvidia","status":"available","devices":[` +
		`{"nodeName":"node-1","deviceId":"gpu-0","supportType":"Exclusive","capacity":17179869184,"available":17179869184,"operable":true,"health":"yes"},` +
		`{"nodeName":"node-1","deviceId":"gpu-1","supportType":"Exclusive","capacity":17179869184,"available":0,"operable":false,"health":"yes"}]}]},` +
		`"validation":{"ok":false,"reason":"binding required"}}`
	_, raw := parseFailedCheck(failedCheckResp(checkTypeComputeBindingRequired, data))

	_, err := resolveComputeBinding(raw, "comfyui", false)
	var bindErr *computeBindingError
	if !errors.As(err, &bindErr) {
		t.Fatalf("want *computeBindingError, got %T (%v)", err, err)
	}
	if bindErr.reason != "binding required" {
		t.Fatalf("reason = %q, want %q", bindErr.reason, "binding required")
	}
	// Only the operable device is offered.
	if !reflect.DeepEqual(bindErr.options, []string{"node-1:gpu-0"}) {
		t.Fatalf("options = %v, want [node-1:gpu-0]", bindErr.options)
	}
}

func TestComputeBindingRejected(t *testing.T) {
	data := `{"availability":{"scope":"card","nodes":[{"nodeName":"node-1","devices":[` +
		`{"nodeName":"node-1","deviceId":"gpu-0","supportType":"Exclusive","operable":true}]}]},` +
		`"validation":{"ok":false,"code":"node-pressure:node-1","reason":"node under memory pressure"}}`
	_, raw := parseFailedCheck(failedCheckResp(checkTypeComputeBindingUnavailable, data))

	err := computeBindingRejected(raw, "comfyui", []BindingSelection{{NodeName: "node-2", DeviceID: "gpu-9"}})
	msg := err.Error()
	for _, want := range []string{"node-2:gpu-9", "rejected", "node under memory pressure", "node-1:gpu-0"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("rejected message %q missing %q", msg, want)
		}
	}
}

func TestParseComputeBindingFlags(t *testing.T) {
	got, err := parseComputeBindingFlags([]string{"node-1:gpu-0", "node-1:gpu-1:8", "node-1:gpu-2:512Mi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []BindingSelection{
		{NodeName: "node-1", DeviceID: "gpu-0"},
		{NodeName: "node-1", DeviceID: "gpu-1", Memory: 8 * giBytes},
		{NodeName: "node-1", DeviceID: "gpu-2", Memory: 512 * miBytes},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseComputeBindingFlags = %#v, want %#v", got, want)
	}

	for _, bad := range []string{"nodeonly", "node:dev:8:extra", "node::8", ":dev", "node-1:gpu-0:notanumber", "node-1:gpu-0:-3"} {
		if _, err := parseComputeBindingFlags([]string{bad}); err == nil {
			t.Fatalf("expected error for %q, got nil", bad)
		}
	}

	if got, err := parseComputeBindingFlags(nil); err != nil || got != nil {
		t.Fatalf("nil input: got (%#v, %v), want (nil, nil)", got, err)
	}
}

func TestParseBindingMemory(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"8", 8 * giBytes},      // bare number defaults to Gi
		{"8Gi", 8 * giBytes},    // explicit Gi
		{"8gi", 8 * giBytes},    // case-insensitive suffix
		{"1.5", 3 * giBytes / 2}, // fractional Gi
		{"512Mi", 512 * miBytes}, // Mi suffix
		{"512mi", 512 * miBytes},
		{" 4 Gi ", 4 * giBytes}, // surrounding / inner whitespace tolerated
	}
	for _, tc := range cases {
		got, err := parseBindingMemory(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("parseBindingMemory(%q) = (%d, %v), want (%d, nil)", tc.in, got, err, tc.want)
		}
	}
	for _, bad := range []string{"", "0", "-1", "abc", "Gi", "8Ti", "8 GB"} {
		if _, err := parseBindingMemory(bad); err == nil {
			t.Fatalf("expected error for %q", bad)
		}
	}
}
