package webhook

import (
	"context"
	"testing"
	"time"

	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestEndpointSliceHasReadyPodIP(t *testing.T) {
	slice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-service-xyz",
			Namespace: "os-framework",
			Labels:    map[string]string{"kubernetes.io/service-name": "app-service"},
		},
		Ports: []discoveryv1.EndpointPort{{
			Name: ptr.To("webhook"),
			Port: ptr.To[int32](8433),
		}},
		Endpoints: []discoveryv1.Endpoint{{
			Addresses: []string{"10.233.98.55"},
			Conditions: discoveryv1.EndpointConditions{
				Ready: ptr.To(true),
			},
		}},
	}
	if !endpointSliceHasReadyPodIP(slice, "10.233.98.55", 8433) {
		t.Fatal("expected ready match for .55:8433")
	}
	if endpointSliceHasReadyPodIP(slice, "10.233.98.192", 8433) {
		t.Fatal("did not expect match for stale IP")
	}
	if endpointSliceHasReadyPodIP(slice, "10.233.98.55", 6755) {
		t.Fatal("did not expect match for wrong port")
	}

	notReady := slice.DeepCopy()
	notReady.Endpoints[0].Conditions.Ready = ptr.To(false)
	if endpointSliceHasReadyPodIP(notReady, "10.233.98.55", 8433) {
		t.Fatal("did not expect match for not-ready endpoint")
	}
}

func TestWaitForServiceEndpointReadySeesSlice(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	slice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-service-xyz",
			Namespace: "os-framework",
			Labels:    map[string]string{"kubernetes.io/service-name": "app-service"},
		},
		Ports: []discoveryv1.EndpointPort{{Port: ptr.To[int32](8433)}},
		Endpoints: []discoveryv1.Endpoint{{
			Addresses:  []string{"10.233.98.55"},
			Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)},
		}},
	}
	if _, err := client.DiscoveryV1().EndpointSlices("os-framework").Create(ctx, slice, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	err := WaitForServiceEndpointReady(ctx, client, "os-framework", "app-service", "10.233.98.55", 8433, time.Second, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForServiceEndpointReady: %v", err)
	}
}

func TestWaitForServiceEndpointReadyTimeoutStillNil(t *testing.T) {
	ctx := context.Background()
	client := fake.NewSimpleClientset()
	err := WaitForServiceEndpointReady(ctx, client, "os-framework", "app-service", "10.233.98.55", 8433, 150*time.Millisecond, 40*time.Millisecond)
	if err != nil {
		t.Fatalf("timeout should not fail hard: %v", err)
	}
}

func TestWaitForServiceEndpointReadyEmptyPodIP(t *testing.T) {
	err := WaitForServiceEndpointReady(context.Background(), fake.NewSimpleClientset(), "os-framework", "app-service", "", 8433, time.Second, time.Millisecond)
	if err != nil {
		t.Fatalf("empty pod IP should skip: %v", err)
	}
}
