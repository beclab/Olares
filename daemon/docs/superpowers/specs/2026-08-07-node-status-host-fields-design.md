# Design: Add host detail fields to `/system/node-status`

**Date:** 2026-08-07  
**Status:** Approved for implementation planning

## Goal

Extend `GET /system/node-status` so each node reports five host details the node detail page needs:

- kernel / CPU architecture
- kernel version
- internal LAN IP
- wired connection status
- Wi-Fi SSID

## Decisions (confirmed)

| Topic | Choice |
|---|---|
| Wired status | Boolean `wiredConnected` (`true` / `false`), same as `/system/status` |
| Wi-Fi when disconnected | Omit `wifiSSID` (`omitempty`) |
| Kernel version visibility | Only on `/system/node-status`; do **not** change `/system/status` wire format |
| JSON naming | Align with `/system/status` keys: `os_arch`, `os_kernel`, `hostIp`, `wiredConnected`, `wifiSSID` |
| Approach | Project from the existing 5s state snapshot; store kernel in `State` with `json:"-"` |

## Wire format

Add fields to `nodestatus.Status`:

| JSON key | Go field | Type | Source | Empty / absent behavior |
|---|---|---|---|---|
| `os_arch` | `OsArch` | `string` | `State.OsArch` | `""` when unknown |
| `os_kernel` | `OsKernel` | `string` | `State.OsKernel` (new, `json:"-"` on State) | `""` when unknown |
| `hostIp` | `HostIP` | `string` | `State.HostIP` | `""` when unknown |
| `wiredConnected` | `WiredConnected` | `bool` | `State.WiredConnected` | `false` when not connected |
| `wifiSSID` | `WifiSSID` | `string` | `State.WifiSSID` | omit when nil / empty |

No route, auth, or aggregation protocol changes. Cluster summary that fans out to `/system/node-status` automatically receives the new fields in each node's body.

## Data flow

```
status refresh (≈5s)
  └─ GetMachineInfo → OsArch, OsKernel (and existing OS fields)
  └─ network probe → HostIP, WiredConnected, WifiSSID
        │
        ▼
CurrentState snapshot
        │
        ▼
localNodeStatus → nodestatus.Build → GET /system/node-status
```

### State change

In `cli/pkg/daemon/state/types.go`, add:

```go
// OsKernel is the kernel release string from GetMachineInfo (OS_KERNEL).
// It is daemon-internal: excluded from /system/status so that endpoint's
// wire format stays unchanged. /system/node-status projects it as os_kernel.
OsKernel string `json:"-"`
```

In `pkg/cluster/state/current.go`, stop discarding the kernel return value:

```go
osType, osInfo, osArch, osVersion, osKernel, err := GetMachineInfo(ctx)
// ...
CurrentState.OsKernel = osKernel
```

`OsArch`, `HostIP`, `WiredConnected`, and `WifiSSID` are already populated by the refresh loop.

### Build mapping

`nodestatus.Build` maps the snapshot into the new Status fields. For `wifiSSID`, set the string only when `State.WifiSSID` is non-nil and non-empty so `omitempty` drops the key when Wi-Fi is down.

`Build` continues to take identity + snapshot only; no per-request host probes.

## Error handling

- `GetMachineInfo` failure: keep existing warning log; leave `OsArch` / `OsKernel` empty; still return HTTP 200.
- Incomplete network info: empty `hostIp`, `wiredConnected=false`, no `wifiSSID` — same semantics as `/system/status`.
- Never invent placeholder values for unknown strings.

## Out of scope

- Changing `/system/status` public JSON (including adding `os_kernel` there)
- Richer wired status (interface name, connection name, etc.)
- Changing capability probes, health, phase, or cluster aggregation logic
- Frontend changes

## Test plan

1. **Build mapping** — snapshot with all five values → Status fields match.
2. **Wire keys** — marshaled JSON contains `os_arch`, `os_kernel`, `hostIp`, `wiredConnected`; contains `wifiSSID` when set.
3. **Wi-Fi omit** — disconnected snapshot → no `wifiSSID` key.
4. **Empty snapshot** — strings empty, `wiredConnected=false`, no `wifiSSID`.
5. **`/system/status` unchanged** — marshaling `State` with `OsKernel` set must not emit `os_kernel` / `OsKernel`.
6. **Handler** — extend `/system/node-status` detail-field tests so the HTTP body includes the new keys.

## Files to touch

- `cli/pkg/daemon/state/types.go` — add `OsKernel` with `json:"-"`
- `daemon/pkg/cluster/state/current.go` — persist kernel from `GetMachineInfo`
- `daemon/pkg/cluster/nodestatus/nodestatus.go` — Status fields + `Build` mapping
- `daemon/pkg/cluster/nodestatus/detail_test.go` (and/or `nodestatus_test.go`) — unit tests
- `daemon/internel/apiserver/handlers/handler_node_status_test.go` — HTTP wire tests
- Possibly a small `cli/pkg/daemon/state` test that `OsKernel` is omitted from JSON
