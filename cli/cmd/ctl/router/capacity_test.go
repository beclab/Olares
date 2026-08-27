package router

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// A queued request and a slow one look the same from outside, so the reading has
// to arrive whole: what is running, what is behind it, how wide the engine is,
// and how old the numbers are.
func TestAProviderShowsWhatItsEngineIsHolding(t *testing.T) {
	const raw = `{
  "id": "11111111-1111-1111-1111-111111111111",
  "name": "Olares",
  "provider_type": "openai-compatible",
  "base_url": "http://localhost",
  "status": "active",
  "source": "olares",
  "olares_app_name": "llamacppqwen3",
  "credentials_version": 1,
  "engine_load": {
    "engine_kind": "llama.cpp",
    "processing": 1,
    "deferred": 18,
    "slots": 1,
    "observed_at": "2026-08-27T10:00:00Z"
  },
  "created_at": "2026-08-01T00:00:00Z",
  "updated_at": "2026-08-01T00:00:00Z"
}`
	var p providerRow
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if p.EngineLoad == nil {
		t.Fatal("engine_load did not decode")
	}
	var buf bytes.Buffer
	if err := renderProviderRow(&buf, &p); err != nil {
		t.Fatalf("renderProviderRow: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ENGINE LOAD", "1 of 1 slots busy", "18 queued behind", "llama.cpp", "ago"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in:\n%s", want, out)
		}
	}
}

// A provider with no reading says nothing rather than claiming an idle queue: a
// cloud account has no engine of ours to report one.
func TestAProviderWithNoReadingClaimsNothing(t *testing.T) {
	var buf bytes.Buffer
	if err := renderProviderRow(&buf, &providerRow{Name: "OpenAI", Source: "manual", Status: "active"}); err != nil {
		t.Fatalf("renderProviderRow: %v", err)
	}
	if strings.Contains(buf.String(), "ENGINE LOAD") {
		t.Fatalf("nothing reported a queue here:\n%s", buf.String())
	}
}

// Slots absent means the engine kept its own default, which is not a number
// readable from out here. Saying "0 of 0" would read as serving nobody.
func TestAnEngineThatDidNotDeclareItsWidthIsNotGivenOne(t *testing.T) {
	load := &engineLoad{Processing: 2, Deferred: 0, ObservedAt: time.Now()}
	got := load.describe()
	if !strings.Contains(got, "2 processing") {
		t.Fatalf("got %q, want the count without a width", got)
	}
	if strings.Contains(got, "slots") {
		t.Fatalf("nothing declared a width, got %q", got)
	}
}

func TestAReadingWithNoTimestampSaysSo(t *testing.T) {
	load := &engineLoad{Processing: 1, Deferred: 1}
	if !strings.Contains(load.describe(), "no timestamp") {
		t.Fatalf("an undated reading should say so, got %q", load.describe())
	}
}

// The width only exists for a local engine, so the column comes and goes with
// the rows. A permanent column of dashes reads as a figure nobody filled in.
func TestTheWidthColumnAppearsOnlyWhereItIsKnown(t *testing.T) {
	ptr := func(s string) *string { return &s }
	local := []adminModelRow{{
		ProviderName: "Olares", ProviderType: "openai-compatible", ProviderSource: "olares",
		ProviderStatus: "active", ProviderOlaresStatus: ptr("running"), ModelConsoleStatus: ptr("ready"),
		Model: providerModelRow{Name: "qwen3", Mode: "chat", Enabled: true, Status: "active", MaxConcurrency: 4},
	}}
	var buf bytes.Buffer
	if err := renderModelList(&buf, local, 1, 100, 0); err != nil {
		t.Fatalf("renderModelList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "AT ONCE") || !strings.Contains(out, "4") {
		t.Fatalf("expected the width column, got:\n%s", out)
	}

	cloud := []adminModelRow{{
		ProviderName: "OpenAI", ProviderType: "openai", ProviderSource: "manual", ProviderStatus: "active",
		Model: providerModelRow{Name: "gpt-4o", Mode: "chat", Enabled: true, Status: "active"},
	}}
	var plain bytes.Buffer
	if err := renderModelList(&plain, cloud, 1, 100, 0); err != nil {
		t.Fatalf("renderModelList: %v", err)
	}
	if strings.Contains(plain.String(), "AT ONCE") {
		t.Fatalf("no cloud model declares a width:\n%s", plain.String())
	}
}

func TestOneModelReportsItsWidth(t *testing.T) {
	var buf bytes.Buffer
	err := renderProviderModel(&buf,
		&providerRow{Name: "Olares", Source: "olares"},
		&providerModelRow{Name: "qwen3", Mode: "chat", Enabled: true, Status: "active", MaxConcurrency: 4})
	if err != nil {
		t.Fatalf("renderProviderModel: %v", err)
	}
	if !strings.Contains(buf.String(), "AT ONCE") || !strings.Contains(buf.String(), "4 requests") {
		t.Fatalf("expected the width, got:\n%s", buf.String())
	}
}

// The data plane's own list carries the width too, and it is the list a caller
// reads before sending. Its json has to keep the card's figures rather than drop
// them on re-serialization.
func TestTheCallableListKeepsTheCardsFigures(t *testing.T) {
	const raw = `{"id":"Olares/qwen3","object":"model","mode":"chat","supports":["chat"],
	"readiness":"ready","owned_by":"llamacppqwen3","context_size":32768,
	"max_output_tokens":4096,"max_concurrency":4}`
	var m modelObject
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if m.ContextSize != 32768 || m.MaxOutputTokens != 4096 || m.MaxConcurrency != 4 {
		t.Fatalf("the card's figures did not decode: %+v", m)
	}
	var buf bytes.Buffer
	if err := renderModelsList(&buf, []modelObject{m}, false); err != nil {
		t.Fatalf("renderModelsList: %v", err)
	}
	if !strings.Contains(buf.String(), "AT ONCE") {
		t.Fatalf("expected the width column, got:\n%s", buf.String())
	}
}
