package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	secretv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestPublishedEntryVersionChangesOnRotation(t *testing.T) {
	resources := &message.XdsResources{}
	s := New(resources, &Config{})
	ctx := t.Context()

	withSecret := func(path string) *message.XdsSnapshot {
		return &message.XdsSnapshot{
			Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "https_443"}, &listenerv3.Listener{Name: "stream_tcp_a"}},
			Secrets: []cachetypes.Resource{&secretv3.Secret{
				Name: "l4-app-gateway-client",
				Type: &secretv3.Secret_TlsCertificate{TlsCertificate: &secretv3.TlsCertificate{
					WatchedDirectory: &corev3.WatchedDirectory{Path: path},
				}},
			}},
		}
	}
	key := resourceKey{resourcev3.SecretType, "l4-app-gateway-client"}
	currentVersion := func() string {
		t.Helper()
		s.health.mu.RLock()
		defer s.health.mu.RUnlock()
		if _, ok := s.health.expected[resourceKey{resourcev3.ListenerType, "stream_tcp_a"}]; ok {
			t.Fatal("an unrelated TCP listener entered the readiness set")
		}
		return s.health.expected[key]
	}

	resources.Store(defaultNodeID, withSecret("/initial"))
	go s.watchAndUpdate(ctx)
	var initial string
	deadline := time.Now().Add(time.Second)
	for initial == "" && time.Now().Before(deadline) {
		initial = currentVersion()
		time.Sleep(time.Millisecond)
	}
	if initial == "" {
		t.Fatal("initial Entry version was not published")
	}
	resources.Store(defaultNodeID, withSecret("/rotated"))
	for deadline := time.Now().Add(time.Second); time.Now().Before(deadline); {
		if rotated := currentVersion(); rotated != "" && rotated != initial {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("rotation did not change the version required for readiness")
}

func TestDeltaRemovalClearsAcceptedEntryResource(t *testing.T) {
	h := newHealthTracker()
	h.openDeltaStream(1)
	kind := resourcev3.SecretType
	key := resourceKey{kind, "l4-app-gateway-ca"}
	h.received(1, &discoverygrpc.DeltaDiscoveryRequest{TypeUrl: kind, InitialResourceVersions: map[string]string{key.name: "old"}})
	if h.streams[1].applied[key] != "old" {
		t.Fatal("expected accepted resource on reconnect")
	}
	h.sent(1, &discoverygrpc.DeltaDiscoveryResponse{TypeUrl: kind, Nonce: "1", RemovedResources: []string{key.name}})
	h.received(1, &discoverygrpc.DeltaDiscoveryRequest{TypeUrl: kind, ResponseNonce: "1"})
	if h.streams[1].applied[key] != "" {
		t.Fatal("removed trust anchor still counted as accepted")
	}
}

func TestDeltaADSUpdatesAcknowledgedVersion(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
	defer cancel()
	s := New(&message.XdsResources{}, &Config{})
	if err := s.updateSnapshot(ctx, &message.XdsSnapshot{
		Listeners: []cachetypes.Resource{&listenerv3.Listener{Name: "https_443"}},
	}); err != nil {
		t.Fatal(err)
	}
	listener := bufconn.Listen(1 << 20)
	errCh := make(chan error, 1)
	go func() { errCh <- s.serveGRPC(ctx, listener) }()
	conn, err := grpc.NewClient("passthrough:///envoy",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	stream, err := discoverygrpc.NewAggregatedDiscoveryServiceClient(conn).DeltaAggregatedResources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	kind := resourcev3.ListenerType
	if err := stream.Send(&discoverygrpc.DeltaDiscoveryRequest{TypeUrl: kind, Node: &corev3.Node{Id: defaultNodeID}}); err != nil {
		t.Fatal(err)
	}
	response, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if len(response.GetResources()) != 1 || response.GetResources()[0].GetName() != "https_443" || response.GetNonce() == "" {
		t.Fatalf("listener was not delivered with an ACK nonce: %v", response)
	}
	if err := stream.Send(&discoverygrpc.DeltaDiscoveryRequest{TypeUrl: kind, ResponseNonce: response.GetNonce()}); err != nil {
		t.Fatal(err)
	}
	key := resourceKey{kind, "https_443"}
	applied := func() bool {
		s.health.mu.RLock()
		defer s.health.mu.RUnlock()
		for _, state := range s.health.streams {
			if state.applied[key] == response.GetResources()[0].GetVersion() {
				return true
			}
		}
		return false
	}
	deadline := time.Now().Add(time.Second)
	for !applied() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if !applied() {
		t.Fatal("the real ADS ACK did not mark the delivered version accepted")
	}
	_ = conn.Close()
	cancel()
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}
