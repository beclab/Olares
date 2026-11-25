package apiserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/emicklei/go-restful/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type deviceInventoryTransport func(*http.Request) (*http.Response, error)

func (f deviceInventoryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClusterDeviceInventory(t *testing.T) {
	const route = "/app-service/v1/hardware/device-inventory"
	const nodes = `{"kind":"NodeList","apiVersion":"v1","items":[
		{"metadata":{"name":"node-a","annotations":{"bytetrade.io/devices":"{\"schemaVersion\":1,\"audio\":[{\"resourceName\":\"audio-1\",\"online\":true,\"selectable\":true}],\"bluetooth\":[{\"resourceName\":\"adapter-1\",\"online\":true,\"selectable\":true,\"accessMode\":\"shared\"}]}"}}},
		{"metadata":{"name":"node-b","annotations":{"bytetrade.io/devices":"{\"schemaVersion\":1,\"video\":[{\"resourceName\":\"video-1\",\"online\":true,\"selectable\":true}]}"}}}
	]}`
	const apps = `{"kind":"ApplicationList","apiVersion":"app.bytetrade.io/v1alpha1","items":[
		{"metadata":{"name":"private-app"},"spec":{"name":"private-app","attachments":[
			{"type":"device.audio","node":"node-a","ref":"audio-1"},{"type":"device.bluetooth","node":"node-a","ref":"adapter-1"}
		]}}
	]}`
	now := time.Now().UTC()
	responses := map[string]string{
		"/api/v1/nodes": nodes,
		"/apis/app.bytetrade.io/v1alpha1/applications":                                                apps,
		"/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/generic-device-plugin-node-a": inventoryLeaseJSON(now),
		"/apis/coordination.k8s.io/v1/namespaces/kube-node-lease/leases/generic-device-plugin-node-b": inventoryLeaseJSON(now.Add(-time.Minute)),
	}
	requests := 0
	handler := &Handler{kubeConfig: &rest.Config{
		Host: "http://kubernetes.test",
		Transport: deviceInventoryTransport(func(req *http.Request) (*http.Response, error) {
			requests++
			body, ok := responses[req.URL.Path]
			if !ok || req.Method != http.MethodGet {
				t.Fatalf("unexpected Kubernetes request %s %s", req.Method, req.URL.Path)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		}),
	}}
	router := restful.NewContainer()
	if err := addServiceToContainer(router, handler); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, route, nil))
	if rec.Code != http.StatusUnauthorized || requests != 0 {
		t.Fatalf("unauthenticated inventory request got %d and made %d Kubernetes requests", rec.Code, requests)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, route, nil)
	req.Header.Set(constants.BflUserKey, "alice")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || strings.Contains(rec.Body.String(), "private-app") {
		t.Fatalf("unexpected cluster device inventory response: %d %s", rec.Code, rec.Body)
	}
	var result ClusterDeviceInventoryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Devices) != 3 {
		t.Fatalf("expected all three device types, got %#v", result.Devices)
	}
	audio, bluetooth, video := result.Devices[0], result.Devices[1], result.Devices[2]
	if audio.Type != "device.audio" || !audio.Online || !audio.Selectable || !audio.Reserved {
		t.Errorf("unexpected reserved audio device: %#v", audio)
	}
	if bluetooth.Type != "device.bluetooth" || !bluetooth.Online || !bluetooth.Selectable || bluetooth.Reserved {
		t.Errorf("unexpected shared Bluetooth device: %#v", bluetooth)
	}
	if video.Type != "device.video" || video.Online || video.Selectable || video.Reason != "device inventory publisher is offline" {
		t.Errorf("unexpected stale video device: %#v", video)
	}
}

func inventoryLeaseJSON(renewTime time.Time) string {
	// Lease.spec.renewTime is a metav1.MicroTime, which only parses exactly six
	// fractional digits. RFC3339Nano emits up to nine (Linux) and trims trailing
	// zeros, either of which makes the lease fail to decode.
	return `{"kind":"Lease","apiVersion":"coordination.k8s.io/v1","spec":{"renewTime":"` + renewTime.UTC().Format(metav1.RFC3339Micro) + `"}}`
}
