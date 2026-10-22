package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

// No Kubernetes Service lookup or write is needed to publish Files routes.
// A fresh installation without per-user Services/RBAC must include all nodes.
func TestFileserverNodesWithoutLegacyResources(t *testing.T) {
	data := &fileserverGlobalData{
		podMap: map[string]*corev1.Pod{
			"worker": {Status: corev1.PodStatus{PodIP: "10.233.1.2"}},
			"master": {Status: corev1.PodStatus{PodIP: "10.233.1.1"}},
		},
		masterNodes: map[string]bool{"master": true},
	}
	nodes := fileserverNodes(data)
	require.Len(t, nodes, 2)
	assert.Equal(t, "master", nodes[0].NodeName)
	assert.Equal(t, "10.233.1.1", nodes[0].PodIP)
	assert.True(t, nodes[0].IsMaster)
	assert.Equal(t, "worker", nodes[1].NodeName)
	assert.False(t, nodes[1].IsMaster)
	// Snapshot content must not change when the backing Pod object changes.
	data.podMap["master"].Status.PodIP = "10.233.2.1"
	assert.Equal(t, "10.233.1.1", nodes[0].PodIP)
	delete(data.podMap, "worker")
	require.Len(t, fileserverNodes(data), 1)
}

func TestFileserverNodesEmpty(t *testing.T) {
	assert.Empty(t, fileserverNodes(&fileserverGlobalData{}))
}
