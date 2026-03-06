package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	clusterservice "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoverygrpc "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointservice "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerservice "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routeservice "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/telepresenceio/watchable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	sigsyaml "sigs.k8s.io/yaml"

	"bytetrade.io/web3os/bfl/internal/ingress/message"
	"k8s.io/klog/v2"
)

const defaultNodeID = "bfl-ingress"

type Config struct {
	Address string
	Port    int
}

type XdsServer struct {
	xdsResources  *message.XdsResources
	snapshotCache cache.SnapshotCache
	cfg           *Config
	version       uint64
}

func New(xdsResources *message.XdsResources, cfg *Config) *XdsServer {
	return &XdsServer{
		xdsResources:  xdsResources,
		snapshotCache: cache.NewSnapshotCache(false, cache.IDHash{}, nil),
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

func (s *XdsServer) updateSnapshot(ctx context.Context, xdsSnapshot *message.XdsSnapshot) error {
	version := atomic.AddUint64(&s.version, 1)
	versionStr := fmt.Sprintf("%d", version)

	snap, err := cache.NewSnapshot(versionStr, map[resource.Type][]types.Resource{
		resource.ListenerType: xdsSnapshot.Listeners,
		resource.ClusterType:  xdsSnapshot.Clusters,
		resource.RouteType:    xdsSnapshot.Routes,
		resource.SecretType:   xdsSnapshot.Secrets,
	})
	if err != nil {
		return fmt.Errorf("create snapshot: %w", err)
	}

	if err := snap.Consistent(); err != nil {
		klog.Warningf("xds-server: snapshot not consistent: %v", err)
	}

	if err := s.snapshotCache.SetSnapshot(ctx, defaultNodeID, snap); err != nil {
		return fmt.Errorf("set snapshot: %w", err)
	}
	klog.Infof("xds-server: updated snapshot version=%s, listeners=%d, clusters=%d",
		versionStr, len(xdsSnapshot.Listeners), len(xdsSnapshot.Clusters))

	s.logSnapshotYAML(xdsSnapshot)

	return nil
}

func (s *XdsServer) logSnapshotYAML(xdsSnapshot *message.XdsSnapshot) {
	marshaler := protojson.MarshalOptions{Indent: "  "}

	marshalAll := func(resources []types.Resource) []json.RawMessage {
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

	yamlData, err := sigsyaml.JSONToYAML(jsonData)
	if err != nil {
		klog.Errorf("xds-server: convert config to yaml: %v", err)
		return
	}

	klog.Infof("xds-server: full envoy xDS config (YAML):\n---\n%s", string(yamlData))
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

// Suppress unused import warnings
var _ = watchable.Snapshot[string, *message.XdsSnapshot]{}
