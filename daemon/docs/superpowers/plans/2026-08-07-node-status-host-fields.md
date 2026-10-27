# Node-status host fields Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `os_arch`, `os_kernel`, `hostIp`, `wiredConnected`, and `wifiSSID` to `GET /system/node-status` by projecting the existing 5s state snapshot (kernel stored on State with `json:"-"` so `/system/status` is unchanged).

**Architecture:** The status refresh loop already collects architecture, LAN IP, wired flag, and Wi-Fi SSID. Kernel release is already parsed by `GetMachineInfo` but discarded; persist it on `State.OsKernel` with `json:"-"`. `nodestatus.Build` maps these snapshot fields onto the node-status wire body. No new routes, auth, or probes per request.

**Tech Stack:** Go 1.25, Fiber handlers, shared `cli/pkg/daemon/state` types, `go test`.

**Spec:** `daemon/docs/superpowers/specs/2026-08-07-node-status-host-fields-design.md`

## Global Constraints

- Wire keys on `/system/node-status`: `os_arch`, `os_kernel`, `hostIp`, `wiredConnected`, `wifiSSID` (align with `/system/status` naming where those keys already exist).
- `/system/status` public JSON must not gain `os_kernel` / `OsKernel`.
- `wiredConnected` is a bool; Wi-Fi SSID uses `omitempty` (omit when disconnected / empty).
- Unknown strings stay `""`; never invent placeholders.
- Project from the snapshot inside `nodestatus.Build`; no per-request host probes.

## File map

| File | Responsibility |
|---|---|
| `cli/pkg/daemon/state/types.go` | Add internal `OsKernel` (`json:"-"`) |
| `cli/pkg/daemon/state/types_test.go` | Prove `OsKernel` is omitted from State JSON |
| `daemon/pkg/cluster/state/current.go` | Persist `osKernel` from `GetMachineInfo` into `CurrentState` |
| `daemon/pkg/cluster/nodestatus/nodestatus.go` | Status fields + `Build` mapping |
| `daemon/pkg/cluster/nodestatus/detail_test.go` | Unit tests for mapping / omitempty / empty snapshot |
| `daemon/internel/apiserver/handlers/handler_node_status_test.go` | HTTP wire coverage for the new keys |

---

### Task 1: Persist kernel on State without changing `/system/status`

**Files:**
- Modify: `cli/pkg/daemon/state/types.go` (after `OsVersion`, before `CpuInfo`)
- Modify: `cli/pkg/daemon/state/types_test.go`
- Modify: `daemon/pkg/cluster/state/current.go` (~lines 124–150)
- Test: `cli/pkg/daemon/state/types_test.go`

**Interfaces:**
- Consumes: `GetMachineInfo(ctx) (osType, osInfo, osArch, osVersion, osKernel string, err error)` (already exists)
- Produces: `State.OsKernel string` with tag `` `json:"-"` ``, populated by the refresh loop

- [ ] **Step 1: Write the failing test**

Add to `cli/pkg/daemon/state/types_test.go`:

```go
func TestStateOsKernelIsExcludedFromSystemStatusWire(t *testing.T) {
	raw, err := json.Marshal(State{
		OsArch:   "amd64",
		OsKernel: "6.8.0-60-generic",
		HostIP:   "10.0.0.2",
	})
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if _, exists := got["os_kernel"]; exists {
		t.Fatalf("os_kernel must not appear on /system/status wire, got %#v", got)
	}
	if _, exists := got["OsKernel"]; exists {
		t.Fatalf("OsKernel must not appear on /system/status wire, got %#v", got)
	}
	if got["os_arch"] != "amd64" {
		t.Fatalf("os_arch = %#v, want amd64", got["os_arch"])
	}
	if got["hostIp"] != "10.0.0.2" {
		t.Fatalf("hostIp = %#v, want 10.0.0.2", got["hostIp"])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/cli && go test ./pkg/daemon/state/ -run TestStateOsKernelIsExcludedFromSystemStatusWire -v
```

Expected: FAIL — `State.OsKernel` undefined (or compile error).

- [ ] **Step 3: Add `OsKernel` to State**

In `cli/pkg/daemon/state/types.go`, after the `OsVersion` field block, insert:

```go
	// OsKernel is the kernel release string from GetMachineInfo (OS_KERNEL).
	// It is daemon-internal: excluded from /system/status so that endpoint's
	// wire format stays unchanged. /system/node-status projects it as os_kernel.
	OsKernel string `json:"-"`
```

- [ ] **Step 4: Persist kernel in the refresh loop**

In `daemon/pkg/cluster/state/current.go`, change:

```go
osType, osInfo, osArch, osVersion, _, err := GetMachineInfo(ctx)
```

to:

```go
osType, osInfo, osArch, osVersion, osKernel, err := GetMachineInfo(ctx)
```

And after `CurrentState.OsArch = osArch` (near the other OS assignments), add:

```go
CurrentState.OsKernel = osKernel
```

Keep the existing assignments for `OsInfo`, `OsVersion`, `OsType` unchanged.

- [ ] **Step 5: Run test to verify it passes**

Run:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/cli && go test ./pkg/daemon/state/ -run TestStateOsKernelIsExcludedFromSystemStatusWire -v
```

Expected: PASS

Also confirm daemon still builds against the local cli replace:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/daemon && go build ./pkg/cluster/state/
```

Expected: success (no compile errors).

- [ ] **Step 6: Commit**

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os
git add cli/pkg/daemon/state/types.go cli/pkg/daemon/state/types_test.go daemon/pkg/cluster/state/current.go
git commit -m "$(cat <<'EOF'
feat(daemon): keep kernel release on state without status wire change

EOF
)"
```

---

### Task 2: Project host fields in `nodestatus.Build`

**Files:**
- Modify: `daemon/pkg/cluster/nodestatus/nodestatus.go`
- Modify: `daemon/pkg/cluster/nodestatus/detail_test.go`
- Test: `daemon/pkg/cluster/nodestatus/detail_test.go`

**Interfaces:**
- Consumes: `State.OsArch`, `State.OsKernel`, `State.HostIP`, `State.WiredConnected`, `State.WifiSSID *string`
- Produces: `Status` fields:
  - `OsArch string \`json:"os_arch"\``
  - `OsKernel string \`json:"os_kernel"\``
  - `HostIP string \`json:"hostIp"\``
  - `WiredConnected bool \`json:"wiredConnected"\``
  - `WifiSSID string \`json:"wifiSSID,omitempty"\``

- [ ] **Step 1: Write the failing tests**

Append to `daemon/pkg/cluster/nodestatus/detail_test.go`:

```go
func TestBuildReportsHostDetailFields(t *testing.T) {
	ssid := "olares-lab"
	st := clistate.State{
		TerminusState:  clistate.TerminusRunning,
		OsArch:         "amd64",
		OsKernel:       "6.8.0-60-generic",
		HostIP:         "10.0.0.2",
		WiredConnected: true,
		WifiSSID:       &ssid,
	}

	got := Build(Identity{}, st, nil, time.Now())

	if got.OsArch != "amd64" {
		t.Errorf("os_arch = %q, want amd64", got.OsArch)
	}
	if got.OsKernel != "6.8.0-60-generic" {
		t.Errorf("os_kernel = %q, want 6.8.0-60-generic", got.OsKernel)
	}
	if got.HostIP != "10.0.0.2" {
		t.Errorf("hostIp = %q, want 10.0.0.2", got.HostIP)
	}
	if !got.WiredConnected {
		t.Errorf("wiredConnected = false, want true")
	}
	if got.WifiSSID != ssid {
		t.Errorf("wifiSSID = %q, want %q", got.WifiSSID, ssid)
	}
}

func TestBuildOmitsWifiSSIDWhenDisconnected(t *testing.T) {
	st := clistate.State{
		TerminusState:  clistate.TerminusRunning,
		WiredConnected: true,
		HostIP:         "10.0.0.2",
	}

	raw, err := json.Marshal(Build(Identity{}, st, nil, time.Now()))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	if _, ok := fields["wifiSSID"]; ok {
		t.Errorf("wifiSSID must be omitted when disconnected: %s", raw)
	}
	for _, key := range []string{"os_arch", "os_kernel", "hostIp", "wiredConnected"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("field %q missing: %s", key, raw)
		}
	}
}

func TestBuildLeavesUnknownHostDetailsEmpty(t *testing.T) {
	got := Build(Identity{}, clistate.State{TerminusState: clistate.TerminusRunning}, nil, time.Now())

	if got.OsArch != "" || got.OsKernel != "" || got.HostIP != "" || got.WifiSSID != "" {
		t.Errorf("want empty host detail strings: %+v", got)
	}
	if got.WiredConnected {
		t.Errorf("wiredConnected = true, want false when unknown")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/daemon && go test ./pkg/cluster/nodestatus/ -run 'TestBuildReportsHostDetailFields|TestBuildOmitsWifiSSIDWhenDisconnected|TestBuildLeavesUnknownHostDetailsEmpty' -v
```

Expected: FAIL — `Status` missing the new fields / compile errors.

- [ ] **Step 3: Extend `Status` and `Build`**

In `daemon/pkg/cluster/nodestatus/nodestatus.go`, add fields to `Status` after `OlaresdVersion` (before `ObservedAt`):

```go
	// Host detail fields projected from the local state snapshot. Keys match
	// /system/status where those already exist; os_kernel is node-status only.
	OsArch         string `json:"os_arch"`
	OsKernel       string `json:"os_kernel"`
	HostIP         string `json:"hostIp"`
	WiredConnected bool   `json:"wiredConnected"`
	WifiSSID       string `json:"wifiSSID,omitempty"`
```

In `Build`, before the `return Status{...}`, resolve Wi-Fi:

```go
	var wifiSSID string
	if st.WifiSSID != nil {
		wifiSSID = *st.WifiSSID
	}
```

And include in the returned struct literal:

```go
		OsArch:         st.OsArch,
		OsKernel:       st.OsKernel,
		HostIP:         st.HostIP,
		WiredConnected: st.WiredConnected,
		WifiSSID:       wifiSSID,
```

- [ ] **Step 4: Run tests to verify they pass**

Run:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/daemon && go test ./pkg/cluster/nodestatus/ -run 'TestBuildReportsHostDetailFields|TestBuildOmitsWifiSSIDWhenDisconnected|TestBuildLeavesUnknownHostDetailsEmpty' -v
```

Expected: PASS

Also run the existing detail tests to ensure no regressions:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/daemon && go test ./pkg/cluster/nodestatus/ -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os
git add daemon/pkg/cluster/nodestatus/nodestatus.go daemon/pkg/cluster/nodestatus/detail_test.go
git commit -m "$(cat <<'EOF'
feat(daemon): expose host details on node-status build

EOF
)"
```

---

### Task 3: Cover the HTTP `/system/node-status` wire

**Files:**
- Modify: `daemon/internel/apiserver/handlers/handler_node_status_test.go`
- Test: same file (`TestNodeStatusReportsTheDetailFields` and a small Wi-Fi omit case)

**Interfaces:**
- Consumes: `nodestatus.Status` fields from Task 2; `withCurrentState` test helper already installs a snapshot
- Produces: HTTP `data` object containing the five keys (SSID omitted when unset)

- [ ] **Step 1: Extend the detail-fields handler test (failing first)**

In `TestNodeStatusReportsTheDetailFields`, expand the seeded state and assertions.

Update the `withCurrentState` seed to:

```go
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
```

After decoding `got`, add:

```go
	if got.OsArch != "amd64" || got.OsKernel != "6.8.0-60-generic" {
		t.Errorf("os_arch/os_kernel = %q / %q: %s", got.OsArch, got.OsKernel, body)
	}
	if got.HostIP != "10.0.0.2" || !got.WiredConnected || got.WifiSSID != ssid {
		t.Errorf("hostIp/wired/wifi = %q / %v / %q: %s", got.HostIP, got.WiredConnected, got.WifiSSID, body)
	}
```

Extend the required-key loop to:

```go
	for _, key := range []string{"memory", "disk", "olaresdVersion", "observedAt", "deviceType", "os_arch", "os_kernel", "hostIp", "wiredConnected", "wifiSSID"} {
```

- [ ] **Step 2: Add a Wi-Fi omit handler test**

Append:

```go
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
```

- [ ] **Step 3: Run tests**

If Task 2 is done, these should pass. If run before Task 2 implementation, expect FAIL on missing fields.

Run:

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os/daemon && go test ./internel/apiserver/handlers/ -run 'TestNodeStatusReportsTheDetailFields|TestNodeStatusOmitsWifiSSIDWhenDisconnected' -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
cd /Users/liuyu/workspace/bytetrade.io/terminus-os
git add daemon/internel/apiserver/handlers/handler_node_status_test.go
git commit -m "$(cat <<'EOF'
test(daemon): cover node-status host detail wire fields

EOF
)"
```

---

## Spec coverage checklist

| Spec requirement | Task |
|---|---|
| `os_arch` on node-status | Task 2, 3 |
| `os_kernel` on node-status only | Task 1 (`json:"-"`), Task 2, 3 |
| `hostIp` | Task 2, 3 |
| `wiredConnected` bool | Task 2, 3 |
| `wifiSSID` omitempty when disconnected | Task 2, 3 |
| Persist kernel from `GetMachineInfo` | Task 1 |
| No `/system/status` wire change | Task 1 test |
| Empty / unknown → empty strings, false wired | Task 2 |
| Handler HTTP coverage | Task 3 |
| No route/auth/aggregation changes | (explicitly out of scope; no task) |
