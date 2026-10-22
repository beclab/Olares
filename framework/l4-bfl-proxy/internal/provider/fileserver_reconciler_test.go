package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/beclab/l4-bfl-proxy/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type filesInformer struct {
	ctrlcache.Informer
	handler toolscache.ResourceEventHandler
}

func (i *filesInformer) AddEventHandler(h toolscache.ResourceEventHandler) (toolscache.ResourceEventHandlerRegistration, error) {
	i.handler = h
	return nil, nil
}

type filesEndpointCache struct {
	ctrlcache.Cache
	pods         corev1.PodList
	podInformer  filesInformer
	nodeInformer filesInformer
}

func (c *filesEndpointCache) GetInformer(_ context.Context, obj client.Object, _ ...ctrlcache.InformerGetOption) (ctrlcache.Informer, error) {
	switch obj.(type) {
	case *corev1.Pod:
		return &c.podInformer, nil
	case *corev1.Node:
		return &c.nodeInformer, nil
	default:
		return nil, fmt.Errorf("unexpected informer %T", obj)
	}
}
func (c *filesEndpointCache) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) error {
	switch dst := list.(type) {
	case *corev1.PodList:
		*dst = *c.pods.DeepCopy()
	case *corev1.NodeList:
		*dst = corev1.NodeList{}
	default:
		return fmt.Errorf("unexpected list %T", list)
	}
	return nil
}

// A real informer handler schedules the Provider snapshot rebuild. Endpoint
// discovery then reads the latest Pod IP rather than retaining the old one.
func TestFileserverReconcilerRefreshesPodEndpoints(t *testing.T) {
	ctx := context.Background()
	old := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "files-old", Labels: map[string]string{"app": "files"}},
		Spec:       corev1.PodSpec{NodeName: "worker"},
		Status:     corev1.PodStatus{PodIP: "10.233.1.1", Phase: corev1.PodRunning},
	}
	c := &filesEndpointCache{pods: corev1.PodList{Items: []corev1.Pod{old}}}
	p := New(c, nil, &Config{})
	p.synced.Store(true)
	r := NewFileserverReconciler(c)
	r.OnReconciled = p.NotifyChanged
	require.NoError(t, r.SetupWithManager(ctx))
	readIP := func() string {
		gd, err := p.getFileserverGlobalData(ctx)
		require.NoError(t, err)
		nodes := fileserverNodes(gd)
		if len(nodes) == 0 {
			return ""
		}
		require.Len(t, nodes, 1)
		return nodes[0].PodIP
	}
	refresh := func() {
		select {
		case <-p.debounceCh:
		default:
			t.Fatal("Files change did not schedule a snapshot refresh")
		}
	}
	c.podInformer.handler.OnAdd(&old, false)
	refresh()
	assert.Equal(t, "10.233.1.1", readIP())
	replacement := old.DeepCopy()
	replacement.Name = "files-new"
	replacement.Status.PodIP = "10.233.2.9"
	c.pods.Items = []corev1.Pod{*replacement}
	c.podInformer.handler.OnUpdate(&old, replacement)
	refresh()
	assert.Equal(t, "10.233.2.9", readIP())
	terminating := old.DeepCopy()
	now := metav1.Now()
	terminating.DeletionTimestamp = &now
	c.pods.Items = append(c.pods.Items, *terminating)
	c.podInformer.handler.OnUpdate(&old, terminating)
	refresh()
	assert.Equal(t, "10.233.2.9", readIP(), "terminating old pod must not replace the new endpoint")
	c.pods.Items = nil
	c.podInformer.handler.OnDelete(toolscache.DeletedFinalStateUnknown{Obj: replacement})
	refresh()
	assert.Empty(t, readIP())
	unrelated := old.DeepCopy()
	unrelated.Labels = map[string]string{"app": "other"}
	c.podInformer.handler.OnAdd(unrelated, false)
	assert.Empty(t, p.debounceCh)
	c.nodeInformer.handler.OnUpdate(&corev1.Node{}, &corev1.Node{})
	refresh()
}

// A newly joined node can appear before its Files pod has been scheduled or
// assigned an IP. Publish its route only after the Pod IP becomes available.
func TestFileserverReconcilerNewNodeAndPod(t *testing.T) {
	ctx := context.Background()
	c := &filesEndpointCache{}
	p := New(c, nil, &Config{})
	p.synced.Store(true)
	r := NewFileserverReconciler(c)
	r.OnReconciled = p.NotifyChanged
	require.NoError(t, r.SetupWithManager(ctx))
	refresh := func() []*message.FileserverNodeInfo {
		select {
		case <-p.debounceCh:
		default:
			t.Fatal("new node/pod did not schedule a snapshot refresh")
		}
		gd, err := p.getFileserverGlobalData(ctx)
		require.NoError(t, err)
		return fileserverNodes(gd)
	}
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "new-worker"}}
	c.nodeInformer.handler.OnAdd(node, false)
	assert.Empty(t, refresh(), "node without a Files pod has no endpoint yet")
	pending := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "files-new-worker", Labels: map[string]string{"app": "files"}},
		Spec:       corev1.PodSpec{NodeName: node.Name},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	}
	c.pods.Items = []corev1.Pod{pending}
	c.podInformer.handler.OnAdd(&pending, false)
	assert.Empty(t, refresh(), "pod without an IP must not publish an empty target")
	running := pending.DeepCopy()
	running.Status.Phase = corev1.PodRunning
	running.Status.PodIP = "10.233.3.8"
	c.pods.Items = []corev1.Pod{*running}
	c.podInformer.handler.OnUpdate(&pending, running)
	nodes := refresh()
	require.Len(t, nodes, 1)
	assert.Equal(t, "new-worker", nodes[0].NodeName)
	assert.Equal(t, "10.233.3.8", nodes[0].PodIP)
}
