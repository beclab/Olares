package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	secretservice "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/telepresenceio/watchable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	sigsyaml "sigs.k8s.io/yaml"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"k8s.io/klog/v2"
)

const defaultNodeID = "l4-bfl-proxy"

type Config struct {
	Address string
	Port    int
}

type XdsServer struct {
	xdsResources  *message.XdsResources
	snapshotCache cachev3.SnapshotCache
	cfg           *Config
}

func New(xdsResources *message.XdsResources, cfg *Config) *XdsServer {
	return &XdsServer{
		xdsResources:  xdsResources,
		snapshotCache: cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil),
		cfg:           cfg,
	}
}

func (s *XdsServer) Name() string { return "xds-server" }

func (s *XdsServer) Start(ctx context.Context) error {
	klog.Info("xds-server: starting")

	go s.watchAndUpdate(ctx)

	return s.runGRPC(ctx)
}

func (s *XdsServer) watchAndUpdate(ctx context.Context) {
	subscription := s.xdsResources.Subscribe(ctx)
	first := true
	for snapshot := range subscription {
		if first {
			first = false
			for _, val := range snapshot.State {
				if val != nil {
					if err := s.updateSnapshot(ctx, val); err != nil {
						klog.Errorf("xds-server: update snapshot: %v", err)
					}
				}
			}
		}
		for _, update := range snapshot.Updates {
			if update.Delete || update.Value == nil {
				continue
			}
			if err := s.updateSnapshot(ctx, update.Value); err != nil {
				klog.Errorf("xds-server: update snapshot: %v", err)
			}
		}
	}
}

// updateSnapshot pushes a new xDS snapshot to Envoy.
//
// Each resource type (Listeners, Clusters, Routes, Secrets) is versioned
// independently using a SHA-256 content hash. The Delta xDS protocol only
// sends resource types whose hash has changed since the last push, so:
//   - Adding/changing an app (routes/clusters only) → listeners are NOT sent
//     → no Envoy listener drain
//   - Rotating a TLS cert (secret only) → listeners are NOT sent → no drain
//   - Adding a new user (listener structure changes) → listeners ARE sent
//     → Envoy drains the modified listener (unavoidable)
func (s *XdsServer) updateSnapshot(ctx context.Context, xdsSnapshot *message.XdsSnapshot) error {
	snap := cachev3.Snapshot{}
	snap.Resources[cachetypes.Listener] = cachev3.NewResources(hashResources(xdsSnapshot.Listeners), xdsSnapshot.Listeners)
	snap.Resources[cachetypes.Cluster] = cachev3.NewResources(hashResources(xdsSnapshot.Clusters), xdsSnapshot.Clusters)
	snap.Resources[cachetypes.Route] = cachev3.NewResources(hashResources(xdsSnapshot.Routes), xdsSnapshot.Routes)
	snap.Resources[cachetypes.Secret] = cachev3.NewResources(hashResources(xdsSnapshot.Secrets), xdsSnapshot.Secrets)

	//if err := snap.Consistent(); err != nil {
	//	klog.Warningf("xds-server: snapshot not consistent: %v", err)
	//}

	if err := s.snapshotCache.SetSnapshot(ctx, defaultNodeID, &snap); err != nil {
		return fmt.Errorf("set snapshot: %w", err)
	}
	klog.Infof("xds-server: updated snapshot (listener-hash=%s clusters=%d routes=%d secrets=%d)",
		snap.Resources[cachetypes.Listener].Version,
		len(xdsSnapshot.Clusters), len(xdsSnapshot.Routes), len(xdsSnapshot.Secrets))

	if klog.V(3).Enabled() {
		s.logSnapshotYAML(xdsSnapshot)
	}

	return nil
}

// hashResources returns a short hex string (first 16 chars of SHA-256) of
// the deterministically serialised bytes of all resources.  This is used as
// the per-type version string in the xDS snapshot so that the Delta xDS
// protocol only re-sends a resource type when its content has actually changed.
func hashResources(resources []cachetypes.Resource) string {
	h := sha256.New()
	for _, r := range resources {
		if msg, ok := r.(proto.Message); ok {
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(msg)
			if err == nil {
				h.Write(b)
			}
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

func (s *XdsServer) runGRPC(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("xds-server: listen %s: %w", addr, err)
	}

	srv := serverv3.NewServer(ctx, s.snapshotCache, nil)

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    30 * time.Second,
			Timeout: 5 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             15 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	discoverygrpc.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	listenerservice.RegisterListenerDiscoveryServiceServer(grpcServer, srv)
	clusterservice.RegisterClusterDiscoveryServiceServer(grpcServer, srv)
	endpointservice.RegisterEndpointDiscoveryServiceServer(grpcServer, srv)
	routeservice.RegisterRouteDiscoveryServiceServer(grpcServer, srv)
	secretservice.RegisterSecretDiscoveryServiceServer(grpcServer, srv)

	klog.Infof("xds-server: gRPC server listening on %s", addr)

	go func() {
		<-ctx.Done()
		klog.Info("xds-server: shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("xds-server: serve: %w", err)
	}
	return nil
}

var _ = watchable.Snapshot[string, *message.XdsSnapshot]{}

func (s *XdsServer) logSnapshotYAML(xdsSnapshot *message.XdsSnapshot) {
	marshaler := protojson.MarshalOptions{Indent: "  "}

	marshalAll := func(resources []cachetypes.Resource) []json.RawMessage {
		out := make([]json.RawMessage, 0, len(resources))
		for _, r := range resources {
			if msg, ok := r.(proto.Message); ok {
				if data, err := marshaler.Marshal(msg); err == nil {
					out = append(out, data)
				}
			}
		}
		return out
	}

	full := map[string]interface{}{
		"listeners":    marshalAll(xdsSnapshot.Listeners),
		"clusters":     marshalAll(xdsSnapshot.Clusters),
		"routes":       marshalAll(xdsSnapshot.Routes),
		"secret_count": len(xdsSnapshot.Secrets),
	}

	jsonData, err := json.Marshal(full)
	if err != nil {
		klog.Errorf("xds-server: marshal config to json: %v", err)
		return
	}
	var sanitized interface{}
	if err := json.Unmarshal(jsonData, &sanitized); err != nil {
		klog.Errorf("xds-server: unmarshal config json: %v", err)
		return
	}
	maskCertificateChainInlineString(sanitized)
	jsonData, err = json.Marshal(sanitized)
	if err != nil {
		klog.Errorf("xds-server: marshal sanitized config json: %v", err)
		return
	}

	yamlData, err := sigsyaml.JSONToYAML(jsonData)
	if err != nil {
		klog.Errorf("xds-server: convert config to yaml: %v", err)
		return
	}

	klog.Infof("xds-server: full envoy xDS config (YAML):\n---\n%s", string(yamlData))
}

func maskCertificateChainInlineString(v interface{}) {
	switch x := v.(type) {
	case map[string]interface{}:
		if certificateChain, ok := x["certificateChain"].(map[string]interface{}); ok {
			if _, ok := certificateChain["inlineString"]; ok {
				certificateChain["inlineString"] = "*"
			}
		}
		if privateKey, ok := x["privateKey"].(map[string]interface{}); ok {
			if _, ok := privateKey["inlineString"]; ok {
				privateKey["inlineString"] = "*"
			}
		}
		for _, child := range x {
			maskCertificateChainInlineString(child)
		}
	case []interface{}:
		for _, item := range x {
			maskCertificateChainInlineString(item)
		}
	}
}
