package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	clistate "github.com/beclab/Olares/cli/pkg/daemon/state"
	"github.com/beclab/Olares/daemon/internel/apiserver/server"
	"github.com/beclab/Olares/daemon/pkg/cluster/inventory"
	"github.com/beclab/Olares/daemon/pkg/cluster/nodestatus"
	"github.com/beclab/Olares/daemon/pkg/cluster/state"
)

func decodeNodeStatus(t *testing.T, body []byte) nodestatus.Status {
	t.Helper()
	var env struct {
		Data nodestatus.Status `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return env.Data
}

func TestNodeStatusRefusesAnUnauthorizedRequest(t *testing.T) {
	asWorker(t)

	resp, body := callRegistered(t, "/system/node-status", nil)

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 without an authorization header: %s", resp.StatusCode, body)
	}
}

func TestNodeStatusAcceptsAnOwnerSignatureForAnOperationPrecheck(t *testing.T) {
	asOwnerSignature(t)
	asWorker(t)

	headers := map[string]string{
		SIGNATURE_HEADER: signatureCarrying(t, ownerBinding("reboot", "client-1")),
	}
	resp, body := callRegistered(t, "/system/node-status", headers)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}
}

func TestNodeStatusReportsThisNode(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "master-1", inventory.RoleMaster, nil)
	hostname := "olares-one"
	device := "Olares One"
	withCurrentState(t, clistate.State{
		TerminusState: clistate.TerminusRunning,
		HostName:      &hostname,
		DeviceName:    &device,
		CpuInfo:       "NVIDIA Grace",
		GPUList:       []string{"NVIDIA GB10"},
	}, time.Now())

	resp, body := callRegistered(t, "/system/node-status", authHeaders())
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}

	got := decodeNodeStatus(t, body)
	if got.NodeName != "master-1" || got.Role != inventory.RoleMaster {
		t.Errorf("identity not reported: %+v", got)
	}
	if got.Hostname != hostname || got.DeviceType != nodestatus.DeviceTypeOlaresOne {
		t.Errorf("host facts not reported: %+v", got)
	}
	if got.Health != nodestatus.HealthHealthy || got.Phase != nodestatus.PhaseRunning {
		t.Errorf("health/phase not derived from the current state: %+v", got)
	}
	if got.Connectivity != nodestatus.ConnectivityOnline {
		t.Errorf("connectivity = %q, want online", got.Connectivity)
	}
	if got.TerminusState != clistate.TerminusRunning {
		t.Errorf("raw terminus state dropped: %+v", got)
	}
	if got.CPU != "NVIDIA Grace" || len(got.GPUs) != 1 {
		t.Errorf("hardware not reported: %+v", got)
	}
}

// The node detail page is the reason this endpoint exists, and memory, disk and
// the olaresd version are three of the fields on it. All three are in the state
// this node already refreshes.
func TestNodeStatusReportsTheDetailFields(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	version := "1.12.6-rc.2"
	ssid := "olares-lab"
	withCurrentState(t, clistate.State{
		TerminusState:  clistate.TerminusRunning,
		Memory:         "128 G",
		Disk:           "3725 G",
		OlaresdVersion: &version,
		OsArch:         "amd64",
		OsKernel:       "6.8.0-60-generic",
		HostIP:         "10.0.0.2",
		WiredConnected: true,
		WifiSSID:       &ssid,
	}, time.Now())

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	got := decodeNodeStatus(t, body)
	if got.Memory != "128 G" || got.Disk != "3725 G" {
		t.Errorf("memory/disk = %q / %q: %s", got.Memory, got.Disk, body)
	}
	if got.OlaresdVersion != version {
		t.Errorf("olaresdVersion = %q, want %q: %s", got.OlaresdVersion, version, body)
	}
	if got.OsArch != "amd64" || got.OsKernel != "6.8.0-60-generic" {
		t.Errorf("os_arch/os_kernel = %q / %q: %s", got.OsArch, got.OsKernel, body)
	}
	if got.HostIP != "10.0.0.2" || !got.WiredConnected || got.WifiSSID != ssid {
		t.Errorf("hostIp/wired/wifi = %q / %v / %q: %s", got.HostIP, got.WiredConnected, got.WifiSSID, body)
	}

	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	for _, key := range []string{"memory", "disk", "olaresdVersion", "observedAt", "deviceType", "os_arch", "os_kernel", "hostIp", "wiredConnected", "wifiSSID"} {
		if _, ok := env.Data[key]; !ok {
			t.Errorf("field %q missing from the node wire format: %s", key, body)
		}
	}
	// There is no node id anywhere in Olares. Serving the node name under one
	// would have clients treat a renameable label as a stable identifier.
	for _, absent := range []string{"nodeId", "node_id", "memoryBytes", "diskBytes"} {
		if _, ok := env.Data[absent]; ok {
			t.Errorf("field %q is not something this node knows: %s", absent, body)
		}
	}
}

func TestNodeStatusOmitsWifiSSIDWhenDisconnected(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	withCurrentState(t, clistate.State{
		TerminusState:  clistate.TerminusRunning,
		HostIP:         "10.0.0.2",
		WiredConnected: true,
	}, time.Now())

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if _, ok := env.Data["wifiSSID"]; ok {
		t.Errorf("wifiSSID must be omitted when disconnected: %s", body)
	}
	if env.Data["hostIp"] != "10.0.0.2" {
		t.Errorf("hostIp = %#v, want 10.0.0.2: %s", env.Data["hostIp"], body)
	}
	if env.Data["wiredConnected"] != true {
		t.Errorf("wiredConnected = %#v, want true: %s", env.Data["wiredConnected"], body)
	}
}

// The timestamp has to be the one the state carries. Stamping the response
// with time.Now() would make a state refreshed hours ago read as current.
func TestNodeStatusReportsWhenTheStateWasObserved(t *testing.T) {
	asAuthorizedUser(t)
	asWorker(t)
	observedAt := time.Now().Add(-7 * time.Minute).Truncate(time.Second)
	withCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning}, observedAt)

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	got := decodeNodeStatus(t, body)
	if got.ObservedAt == nil {
		t.Fatalf("observedAt missing: %s", body)
	}
	if !got.ObservedAt.Equal(observedAt) {
		t.Errorf("observedAt = %v, want the observation time %v", got.ObservedAt, observedAt)
	}
}

// The state refresh holds its mutex across every probe it makes, and on a node
// in trouble those take seconds. This endpoint is what the master polls to
// decide whether a node is reachable, so an answer that waits for the refresh
// is reported upstream as a node that is gone.
func TestNodeStatusAnswersWhileARefreshHoldsTheStateLock(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "worker-1", inventory.RoleWorker, nil)

	state.TerminusStateMu.Lock()
	defer state.TerminusStateMu.Unlock()

	type result struct {
		status int
		err    error
	}
	done := make(chan result, 1)
	go func() {
		req := httptest.NewRequest(http.MethodGet, "/system/node-status", nil)
		req.Header.Set(AUTH_HEADER, testToken)
		resp, err := server.API.App.Test(req)
		if err != nil {
			done <- result{err: err}
			return
		}
		done <- result{status: resp.StatusCode}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("request did not complete while the state was being refreshed: %v", got.err)
		}
		if got.status != http.StatusOK {
			t.Fatalf("status = %d, want 200", got.status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the node-local status endpoint blocked on the state refresh")
	}
}

func TestNodeStatusWithoutAnObservationSaysSo(t *testing.T) {
	asAuthorizedUser(t)
	asWorker(t)
	withCurrentState(t, clistate.State{TerminusState: clistate.Checking}, time.Time{})

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	if got := decodeNodeStatus(t, body); got.ObservedAt != nil {
		t.Errorf("observedAt = %v, want null before the first refresh", got.ObservedAt)
	}
}

// Which capabilities a node declares depends on the machine, so the behaviour
// is pinned in the nodestatus package where the probes can be driven. What the
// handler owes the caller is the field, as an object it can index.
func TestNodeStatusCarriesACapabilityMap(t *testing.T) {
	asAuthorizedUser(t)
	asWorker(t)

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	var env struct {
		Data struct {
			Capabilities map[string]nodestatus.Capability `json:"capabilities"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if env.Data.Capabilities == nil {
		t.Errorf("capabilities must be an object, even when the node declares none: %s", body)
	}
	for name, c := range env.Data.Capabilities {
		if !c.Supported {
			t.Errorf("capability %q declared as unsupported; undeclared is the answer for that", name)
		}
	}
}

// Capabilities are answered for this node in its current deployment, not from
// a constant: a containerized daemon cannot power the machine off.
func TestNodeStatusCapabilitiesFollowTheDeployment(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "worker-1", inventory.RoleWorker, nil)
	container := "docker"
	withCurrentState(t, clistate.State{
		TerminusState: clistate.TerminusRunning,
		ContainerMode: &container,
	}, time.Now())

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	got := decodeNodeStatus(t, body)
	if _, ok := got.Capabilities[nodestatus.CapPowerShutdown]; ok {
		t.Errorf("power.shutdown offered by a containerized daemon: %s", body)
	}
	if _, ok := got.Capabilities[nodestatus.CapPowerReboot]; ok {
		t.Errorf("power.reboot offered by a containerized daemon: %s", body)
	}
}

func TestNodeStatusUnknownIdentityIsNotFabricated(t *testing.T) {
	asAuthorizedUser(t)
	asNode(t, "", "", context.DeadlineExceeded)
	withCurrentState(t, clistate.State{TerminusState: clistate.Checking}, time.Now())

	_, body := callRegistered(t, "/system/node-status", authHeaders())

	got := decodeNodeStatus(t, body)
	if got.Role != inventory.RoleUnknown {
		t.Errorf("role = %q, want unknown when the node cannot resolve itself: %s", got.Role, body)
	}
	if got.Health != nodestatus.HealthUnknown || got.Phase != nodestatus.PhaseUnknown {
		t.Errorf("want unknown health/phase while still checking: %+v", got)
	}
}

// The endpoint TermiPass and olares-cli already poll has to keep answering,
// unauthenticated, while clients migrate to the node-local API.
func TestSystemStatusIsUnchanged(t *testing.T) {
	withCurrentState(t, clistate.State{TerminusState: clistate.TerminusRunning}, time.Now())

	resp, body := callRegistered(t, "/system/status", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want the legacy endpoint to keep answering: %s", resp.StatusCode, body)
	}

	var env struct {
		Data clistate.State `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if env.Data.TerminusState != clistate.TerminusRunning {
		t.Errorf("legacy payload changed shape: %s", body)
	}
}
