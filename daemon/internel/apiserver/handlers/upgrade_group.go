package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API

	// The node-local half of a cluster upgrade. Every node serves it, and it
	// acts only on the node it reaches.
	//
	// Neither route goes through RequireSignature or RequireAuthorization.
	// Both are guarded instead by the per-operation upgrade token, checked
	// inside the handler: an upgrade outlives any signature's permitted
	// lifetime and outlives this daemon's process, so neither of the two
	// credentials the other routes use can cover it. See
	// clusterop.UpgradeDeps.Auth.
	s.App.Post("/command/upgrade-node", handlers.PostUpgradeStage)
	s.App.Get("/node/upgrade-stage", handlers.GetUpgradeStageStatus)
	s.App.Get("/node/upgrade-readiness", handlers.GetUpgradeReadiness)

	klog.V(8).Info("upgrade stage handlers initialized")
}
