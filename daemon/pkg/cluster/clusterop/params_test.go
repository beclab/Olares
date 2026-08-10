package clusterop

import (
	"encoding/json"
	"testing"
)

func TestCanonicalParamsTreatsMissingAndWhitespaceAsEmptyObject(t *testing.T) {
	for _, raw := range []json.RawMessage{
		nil,
		json.RawMessage(""),
		json.RawMessage(" \n\t "),
		json.RawMessage(`{}`),
	} {
		canonical, err := CanonicalParams(raw)
		if err != nil {
			t.Fatalf("CanonicalParams(%q): %v", string(raw), err)
		}
		if string(canonical) != "{}" {
			t.Fatalf("CanonicalParams(%q) = %s, want {}", string(raw), canonical)
		}
	}
}

func TestDigestParamsEmptyAndObjectAreEqual(t *testing.T) {
	omitted, err := DigestParams(nil)
	if err != nil {
		t.Fatal(err)
	}
	emptyObject, err := DigestParams(json.RawMessage(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if omitted != emptyObject {
		t.Fatalf("digests differ: %q != %q", omitted, emptyObject)
	}
}

func TestDigestParamsCanonicalizesObjectOrder(t *testing.T) {
	left, err := DigestParams(json.RawMessage(`{"a":1,"b":2}`))
	if err != nil {
		t.Fatal(err)
	}
	right, err := DigestParams(json.RawMessage(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatalf("digests differ: %q != %q", left, right)
	}
}

func TestDigestParamsChangesWhenValuesOrArraysChange(t *testing.T) {
	first, err := DigestParams(json.RawMessage(`{"value":1,"items":["a","b"]}`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestParams(json.RawMessage(`{"value":2,"items":["a","b"]}`))
	if err != nil {
		t.Fatal(err)
	}
	third, err := DigestParams(json.RawMessage(`{"value":1,"items":["b","a"]}`))
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("changed value kept the same digest")
	}
	if first == third {
		t.Fatal("changed array order kept the same digest")
	}
}

func TestDigestParamsRejectsInvalidJSON(t *testing.T) {
	for _, raw := range []json.RawMessage{
		json.RawMessage(`{`),
		json.RawMessage(`{"a":1} {"b":2}`),
	} {
		if _, err := DigestParams(raw); err == nil {
			t.Fatalf("invalid JSON accepted: %q", string(raw))
		}
	}
}
