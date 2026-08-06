package handlers

import (
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"k8s.io/klog/v2"
)

func init() {
	s := server.API
	cluster := s.App.Group("cluster")

	// The overview above the node list. It is the master's answer for the
	// same reason the directory is: a worker holds a partial view, and a
	// summary of part of a cluster reads exactly like a summary of all of it.
	cluster.Get("/", handlers.RequireAuthorization(
		handlers.RequireMaster(handlers.GetCluster)))

	cluster.Get("/nodes", handlers.RequireAuthorizationOrOwnerSignature(
		handlers.RequireMaster(handlers.RequireLocal(handlers.GetClusterNodes))))

	// Powering the cluster is the owner's decision, so the scan callback's
	// operation-bound owner signature is sufficient. Read routes still require
	// an access token.
	cluster.Post("/operations", handlers.RequireSignature(
		handlers.RequireOwner(
			handlers.RequireMaster(handlers.PostClusterOperation))))

	// Following an operation is an ordinary read: whoever is signed in may
	// watch what the owner started.
	cluster.Get("/operations/by-request", handlers.RequireAuthorization(
		handlers.RequireMaster(handlers.GetClusterOperationByRequest)))
	cluster.Get("/operations/by-request/:requestId", handlers.RequireAuthorization(
		handlers.RequireMaster(handlers.GetClusterOperationByRequest)))
	cluster.Get("/operations/:id", handlers.RequireAuthorization(
		handlers.RequireMaster(handlers.GetClusterOperation)))

	klog.V(8).Info("cluster handlers initialized")
}
