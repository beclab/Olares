package market

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const consumerModePrompt = `{"consumers":[
	{"consumer":"main","modes":[{"computeType":"nvidia","status":"installable"},{"computeType":"intel","status":"installable"}]},
	{"consumer":"helper","modes":[{"computeType":"nvidia","status":"installable"},{"computeType":"intel","status":"no-matching-node"}]}],
	"pending":["main"]}`

func TestResolveComputeSelectionForConsumers(t *testing.T) {
	sel, err := resolveComputeSelection(json.RawMessage(consumerModePrompt), "llm", "",
		map[string]string{"main": "intel", "helper": "nvidia"}, false)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"main": "intel", "helper": "nvidia"}
	if sel.Mode != "" || !reflect.DeepEqual(sel.Consumers, want) {
		t.Fatalf("selection = %+v, want consumers %v", sel, want)
	}
	if got := sel.String(); got != "helper=nvidia, main=intel" {
		t.Fatalf("String() = %q", got)
	}
}

func TestResolveComputeSelectionNamesThePendingConsumer(t *testing.T) {
	_, err := resolveComputeSelection(json.RawMessage(consumerModePrompt), "llm", "", nil, false)
	if err == nil || !strings.Contains(err.Error(), "--consumer-mode main=<mode>") ||
		!strings.Contains(err.Error(), "nvidia, intel") {
		t.Fatalf("err = %v", err)
	}
}

func TestResolveComputeSelectionKeepsTheLegacyPrompt(t *testing.T) {
	sel, err := resolveComputeSelection(json.RawMessage(`[{"computeType":"nvidia","status":"installable"}]`),
		"llm", "nvidia", nil, false)
	if err != nil || sel.Mode != "nvidia" || sel.Consumers != nil {
		t.Fatalf("selection = %+v, err = %v", sel, err)
	}
}

func TestParseConsumerModeFlags(t *testing.T) {
	got, err := parseConsumerModeFlags([]string{" main = nvidia ", "helper=cpu"})
	if err != nil || !reflect.DeepEqual(got, map[string]string{"main": "nvidia", "helper": "cpu"}) {
		t.Fatalf("got %v, err %v", got, err)
	}
	for _, bad := range [][]string{{"main"}, {"=nvidia"}, {"main="}, {"main=nvidia", "main=cpu"}} {
		if _, err := parseConsumerModeFlags(bad); err == nil {
			t.Fatalf("%q: want an error", bad)
		}
	}
}

func TestParseComputeBindingFlagsWithAConsumer(t *testing.T) {
	got, err := parseComputeBindingFlags([]string{"main=olares:gpu-0:8", "olares:gpu-1"})
	if err != nil {
		t.Fatal(err)
	}
	want := []BindingSelection{
		{NodeName: "olares", DeviceID: "gpu-0", Memory: 8 << 30, Consumer: "main"},
		{NodeName: "olares", DeviceID: "gpu-1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
	if _, err := parseComputeBindingFlags([]string{"=olares:gpu-0"}); err == nil {
		t.Fatal("an empty workload must be refused")
	}
}

func TestInstallRequestCarriesConsumerModes(t *testing.T) {
	body, err := json.Marshal(InstallRequest{AppName: "llm", SelectedGpuTypes: map[string]string{"main": "nvidia"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"selectedGpuTypes":{"main":"nvidia"}`) || strings.Contains(string(body), `"selectedGpuType"`) {
		t.Fatalf("body = %s", body)
	}
}
