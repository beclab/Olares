package apiserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/beclab/Olares/daemon/pkg/cluster/state"
	"github.com/beclab/Olares/daemon/pkg/onboarding"
	"github.com/beclab/Olares/daemon/pkg/utils"
	"k8s.io/klog/v2"
)

// StartOnboarding serves the first-user APIs on the master only. The regular
// management API stays on 18088; it is never exposed through this listener.
func StartOnboarding(ctx context.Context) {
	go func() {
		var server *http.Server
		var service *onboarding.Service
		var mu sync.Mutex
		getService := func(ctx context.Context) (*onboarding.Service, error) {
			snapshot, _ := state.Snapshot()
			switch snapshot.TerminusState {
			case state.Uninitialized, state.Initializing, state.InitializeFailed, state.TerminusRunning:
			default:
				return nil, errors.New("system installation is not complete")
			}
			kube, err := utils.GetKubeClient()
			if err != nil {
				return nil, err
			}
			_, _, role, err := utils.GetThisNodeName(ctx, kube)
			if err != nil {
				return nil, err
			}
			if role != "master" {
				return nil, errors.New("onboarding is only available on the master")
			}
			dynamic, err := utils.GetDynamicClient()
			if err != nil {
				return nil, err
			}
			mu.Lock()
			defer mu.Unlock()
			if service == nil || service.Kube != kube || service.Dynamic != dynamic {
				service = &onboarding.Service{Kube: kube, Dynamic: dynamic, ResetPassword: onboarding.ResetWithAuthelia(kube)}
			}
			return service, nil
		}
		stop := func() {
			if server != nil {
				shutdown, cancel := context.WithTimeout(context.Background(), 35*time.Second)
				_ = server.Shutdown(shutdown)
				cancel()
				server = nil
			}
		}
		defer stop()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				probe, cancel := context.WithTimeout(ctx, 5*time.Second)
				s, err := getService(probe)
				var status onboarding.Status
				if err == nil {
					status, err = s.Status(probe)
				}
				cancel()
				if errors.Is(err, onboarding.ErrConflict) || status.State == onboarding.Completed {
					stop()
					continue
				}
				if err != nil {
					stop()
					continue
				}
				if server != nil {
					continue
				}
				listener, err := net.Listen("tcp", ":30180")
				if err != nil {
					klog.Error("start onboarding listener: ", err)
					continue
				}
				server = &http.Server{Handler: onboarding.Handler(getService), ReadHeaderTimeout: 5 * time.Second,
					ReadTimeout: 35 * time.Second, WriteTimeout: 35 * time.Second, IdleTimeout: 30 * time.Second,
					BaseContext: func(net.Listener) context.Context { return ctx }}
				go func(s *http.Server) {
					if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
						klog.Error("onboarding listener: ", err)
					}
				}(server)
			}
		}
	}()
}
