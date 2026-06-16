package server

import (
	"testing"

	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func TestHealthTracker_NackThenAckRecovers(t *testing.T) {
	h := newHealthTracker()

	if err := h.err(); err != nil {
		t.Fatalf("fresh tracker should be healthy, got %v", err)
	}

	// Envoy NACKs the listener config.
	h.observe(resourcev3.ListenerType, "nonce-1", "duplicate matcher")
	if err := h.err(); err == nil {
		t.Fatal("expected unhealthy after NACK")
	}

	// A subsequent ACK (nonce set, no error) clears it.
	h.observe(resourcev3.ListenerType, "nonce-2", "")
	if err := h.err(); err != nil {
		t.Fatalf("expected healthy after ACK, got %v", err)
	}
}

func TestHealthTracker_InitialRequestIgnored(t *testing.T) {
	h := newHealthTracker()
	// Envoy's first request for a type has no nonce and no error; it must not
	// be treated as an ACK or NACK.
	h.observe(resourcev3.ClusterType, "", "")
	if err := h.err(); err != nil {
		t.Fatalf("initial request should keep tracker healthy, got %v", err)
	}
}

func TestHealthTracker_CallbacksWireUp(t *testing.T) {
	h := newHealthTracker()
	cb := h.callbacks()

	_ = cb.StreamDeltaRequestFunc(1, &discoverygrpc.DeltaDiscoveryRequest{
		TypeUrl:       resourcev3.ListenerType,
		ResponseNonce: "n1",
		ErrorDetail:   &status.Status{Message: "boom"},
	})
	if err := h.err(); err == nil {
		t.Fatal("expected unhealthy after delta NACK callback")
	}

	_ = cb.StreamDeltaRequestFunc(1, &discoverygrpc.DeltaDiscoveryRequest{
		TypeUrl:       resourcev3.ListenerType,
		ResponseNonce: "n2",
	})
	if err := h.err(); err != nil {
		t.Fatalf("expected healthy after delta ACK callback, got %v", err)
	}
}
