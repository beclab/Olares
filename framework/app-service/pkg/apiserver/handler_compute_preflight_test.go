package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/compute"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/prometheus"
	"github.com/beclab/Olares/framework/app-service/pkg/utils"
	apputils "github.com/beclab/Olares/framework/app-service/pkg/utils/app"
	"github.com/emicklei/go-restful/v3"
)

type fakePreflightCollector struct {
	snapshot compute.PreflightSnapshot
	err      error
	owners   []string
}

func (f *fakePreflightCollector) Collect(_ context.Context, _ string, owners []string) (compute.PreflightSnapshot, error) {
	f.owners = append([]string(nil), owners...)
	return f.snapshot, f.err
}

func callComputePreflight(t *testing.T, h *Handler, user, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := restful.NewRequest(httptest.NewRequest(http.MethodPost, "/app-service/v1/compute-resources/preflight", strings.NewReader(body)))
	req.Request.Header.Set("Content-Type", restful.MIME_JSON)
	req.SetAttribute(constants.UserContextAttribute, user)
	rec := httptest.NewRecorder()
	resp := restful.NewResponse(rec)
	resp.SetRequestAccepts(restful.MIME_JSON)
	h.computeResourcesPreflight(req, resp)
	return rec
}

func installableSnapshot() compute.PreflightSnapshot {
	return compute.PreflightSnapshot{
		Nodes: []compute.Node{{
			NodeName: "gpu-a",
			GPUTypes: []string{utils.NvidiaCardType},
			Devices: []compute.Device{{
				ID: "gpu0", NodeName: "gpu-a", Mode: utils.NvidiaCardType,
				Memory: 16 << 30, Health: "yes", SupportType: compute.SupportTypeExclusive,
			}},
		}},
		Pressure: compute.PressureSnapshot{
			Threshold: 0.9,
			UsageByNode: map[string]prometheus.NodeResourceUsage{
				"gpu-a": {
					CPUCapacity: 8000, MemoryCapacity: 64 << 30, MemoryAvailable: 48 << 30,
					DiskCapacity: 1 << 40, DiskAvailable: 900 << 30,
				},
			},
		},
		Cluster: &prometheus.ClusterMetrics{
			CPU:    prometheus.Value{Total: 64},
			Memory: prometheus.Value{Total: 128 << 30},
			Disk:   prometheus.Value{Total: 1 << 40},
		},
		Owners:       map[string]*prometheus.ClusterMetrics{"alice": {}, "bob": {}},
		K8sAvailable: apputils.ResourceState{CPU: 64000, Memory: 128 << 30},
	}
}

func validDemandJSON(id, owner string) string {
	return `{
		"id":"` + id + `","appId":"app-id","application":"model","owner":"` + owner + `",
		"mode":"nvidia","requiredCPU":"500m","requiredGPU":"8Gi","limitedGPU":"8Gi",
		"requiredMemory":"1Gi","limitedMemory":"2Gi","requiredDisk":"10Gi",
		"supportMultiCards":false,"supportMultiNodes":false
	}`
}

func testPreflightHandler(snapshot compute.PreflightSnapshot, admin bool) *Handler {
	return &Handler{
		preflightCollector: &fakePreflightCollector{snapshot: snapshot},
		preflightIsAdmin:   func(context.Context, string) (bool, error) { return admin, nil },
	}
}

func TestComputeResourcesPreflightReturnsOverallReport(t *testing.T) {
	collector := &fakePreflightCollector{snapshot: installableSnapshot()}
	h := &Handler{
		preflightCollector: collector,
		preflightIsAdmin:   func(context.Context, string) (bool, error) { return false, nil },
	}
	rec := callComputePreflight(t, h, "alice", `{"demands":[`+validDemandJSON("one", "alice")+`]}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var response ComputePreflightResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !response.Data.Installable || response.Data.FailedDemandID != "" {
		t.Fatalf("unexpected report: %#v", response.Data)
	}
	if strings.Contains(rec.Body.String(), "placement") || strings.Contains(rec.Body.String(), "demands") {
		t.Fatalf("response leaked internal/per-demand details: %s", rec.Body.String())
	}
	if len(collector.owners) != 1 || collector.owners[0] != "alice" {
		t.Fatalf("owners=%v", collector.owners)
	}
}

func TestComputeResourcesPreflightOwnerAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name       string
		admin      bool
		owner      string
		wantStatus int
	}{
		{name: "regular user own", owner: "alice", wantStatus: http.StatusOK},
		{name: "regular user cross owner", owner: "bob", wantStatus: http.StatusForbidden},
		{name: "admin cross owner", admin: true, owner: "bob", wantStatus: http.StatusOK},
		{name: "admin shared", admin: true, owner: "shared", wantStatus: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := callComputePreflight(t, testPreflightHandler(installableSnapshot(), tc.admin), "alice",
				`{"demands":[`+validDemandJSON("one", tc.owner)+`]}`)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestComputeResourcesPreflightRejectsInvalidWireDemands(t *testing.T) {
	h := testPreflightHandler(installableSnapshot(), false)
	tests := []string{
		`{"demands":[]}`,
		`{"demands":[` + validDemandJSON("same", "alice") + `,` + validDemandJSON("same", "alice") + `]}`,
		`{"demands":[` + strings.Replace(validDemandJSON("one", "alice"), `"mode":"nvidia"`, `"mode":"mystery"`, 1) + `]}`,
		`{"demands":[` + strings.Replace(validDemandJSON("one", "alice"), `"requiredMemory":"1Gi"`, `"requiredMemory":"wat"`, 1) + `]}`,
		`{"demands":[` + strings.Replace(validDemandJSON("one", "alice"), `"requiredDisk":"10Gi"`, `"requiredDisk":"-1Gi"`, 1) + `]}`,
		`{"demands":[` + strings.Replace(validDemandJSON("one", "alice"), `"id":"one"`, `"id":"`+strings.Repeat("x", 257)+`"`, 1) + `]}`,
	}
	for _, body := range tests {
		rec := callComputePreflight(t, h, "alice", body)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}

func TestComputeResourcesPreflightLimitsRequestSizeAndCount(t *testing.T) {
	h := testPreflightHandler(installableSnapshot(), false)
	demands := make([]string, 65)
	for i := range demands {
		demands[i] = validDemandJSON(string(rune('a'+i%26))+string(rune('A'+i/26)), "alice")
	}
	rec := callComputePreflight(t, h, "alice", `{"demands":[`+strings.Join(demands, ",")+`]}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("count status=%d body=%s", rec.Code, rec.Body.String())
	}
	body := `{"demands":[` + validDemandJSON("one", "alice") + `],"padding":"` + strings.Repeat("x", 1<<20) + `"}`
	rec = callComputePreflight(t, h, "alice", body)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("size status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestComputePreflightDemandNormalizesMultiNodeToMultiCard(t *testing.T) {
	wire := ComputePreflightDemand{
		ID: "app", Application: "app", Owner: "alice", Mode: utils.NvidiaCardType,
		RequiredCPU: "0", RequiredGPU: "8Gi", LimitedGPU: "8Gi",
		RequiredMemory: "1Gi", LimitedMemory: "1Gi", RequiredDisk: "0",
		SupportMultiNodes: true,
	}
	demand, err := wire.toCompute()
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if !demand.Requirement.SupportMultiNodes || !demand.Requirement.SupportMultiCards {
		t.Fatalf("multi-node must imply multi-card: %#v", demand.Requirement)
	}

	wire.Mode = utils.CPUType
	demand, err = wire.toCompute()
	if err != nil {
		t.Fatalf("convert cpu: %v", err)
	}
	if demand.Requirement.SupportMultiNodes || demand.Requirement.SupportMultiCards {
		t.Fatalf("non-nvidia mode must ignore multi-card flags: %#v", demand.Requirement)
	}
}

func TestComputeResourcesPreflightInternalFailuresAreServerErrors(t *testing.T) {
	tests := []*Handler{
		{
			preflightCollector: &fakePreflightCollector{err: errors.New("metrics unavailable")},
			preflightIsAdmin:   func(context.Context, string) (bool, error) { return false, nil },
		},
		testPreflightHandler(compute.PreflightSnapshot{}, false),
		{preflightCollector: &fakePreflightCollector{snapshot: installableSnapshot()}},
		{preflightIsAdmin: func(context.Context, string) (bool, error) { return false, nil }},
		{
			preflightCollector: &fakePreflightCollector{snapshot: installableSnapshot()},
			preflightIsAdmin: func(context.Context, string) (bool, error) {
				return false, restful.ServiceError{Code: http.StatusBadRequest, Message: "role lookup failed"}
			},
		},
	}
	for _, h := range tests {
		rec := callComputePreflight(t, h, "alice", `{"demands":[`+validDemandJSON("one", "alice")+`]}`)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	}
}
