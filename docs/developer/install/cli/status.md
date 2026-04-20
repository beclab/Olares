# `status`

## Synopsis

The `status` command prints the current Olares system state by calling the local olaresd daemon's `/system/status` HTTP endpoint.

The endpoint is bound to `127.0.0.1:18088` and only accepts loopback traffic, so this command must run on the same host as the daemon (typically the master node).

```bash
olares-cli status [options]
```

By default, the output is grouped into human-readable sections:

- **Olares**: installation lifecycle, version, names, key timestamps.
- **System**: hardware and OS facts about the host.
- **Network**: wired/Wi-Fi connectivity, internal and external IP addresses.
- **Install / Uninstall**: progress of an in-flight install or uninstall.
- **Upgrade**: progress of an in-flight upgrade (download and install phases).
- **Logs collection**: state of the most recent log collection job.
- **Pressures**: active kubelet node pressure conditions, if any.
- **Other**: FRP, container mode, etc.

Use `--json` to receive the raw payload returned by olaresd, suitable for scripting or piping to tools like `jq`.

## Options

| Option       | Usage                                                                                                                                              | Required | Default                     |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------|----------|-----------------------------|
| `--endpoint` | Base URL of the local olaresd daemon. Override only when olaresd is bound to a non-default address.                                                | No       | `http://127.0.0.1:18088`    |
| `--json`     | Print the raw JSON payload returned by olaresd (the `data` field), suitable for piping to tools like `jq`.                                         | No       | `false`                     |
| `--timeout`  | Maximum time to wait for the olaresd response.                                                                                                     | No       | `5s`                        |
| `--help`     | Show command help.                                                                                                                                 | No       | N/A                         |

## Examples

```bash
# Pretty-printed grouped report (default)
olares-cli status

# Raw JSON payload (forwarded verbatim from olaresd)
olares-cli status --json | jq

# Custom daemon endpoint and longer timeout
olares-cli status --endpoint http://127.0.0.1:18088 --timeout 10s
```

## Field reference

The table below lists the fields returned by `olaresd` (the `data` object of the JSON response) and the labels they appear under in the grouped output.

### Olares

| Field             | JSON key           | Meaning                                                                                                  |
|-------------------|--------------------|----------------------------------------------------------------------------------------------------------|
| State             | `terminusState`    | High-level system state. See [State values](#state-values).                                              |
| Olaresd state     | `terminusdState`   | Lifecycle of the olaresd daemon itself: `initialize` while bootstrapping, `running` once ready.          |
| Name              | `terminusName`     | Olares ID of the admin user, e.g. `alice@olares.cn`.                                                     |
| Version           | `terminusVersion`  | Installed Olares version (semver).                                                                       |
| Olaresd version   | `olaresdVersion`   | Running olaresd binary version. Useful for spotting drift after partial upgrades.                        |
| Installed at      | `installedTime`    | Unix epoch (seconds) when Olares finished installing on this node.                                       |
| Initialized at    | `initializedTime`  | Unix epoch (seconds) when the admin user finished initial activation.                                    |

### System

| Field       | JSON key      | Meaning                                                              |
|-------------|---------------|----------------------------------------------------------------------|
| Device      | `device_name` | User-friendly device or chassis name.                                |
| Hostname    | `host_name`   | Kernel hostname.                                                     |
| OS          | `os_type` / `os_arch` / `os_info` | OS family, CPU architecture, distro string.        |
| OS version  | `os_version`  | OS version string, e.g. `22.04`.                                     |
| CPU         | `cpu_info`    | CPU model name.                                                      |
| Memory      | `memory`      | Total physical memory, formatted as `<N> G`.                         |
| Disk        | `disk`        | Total filesystem size of the data partition, formatted as `<N> G`.   |
| GPU         | `gpu_info`    | GPU model name when one is detected.                                 |

### Network

| Field       | JSON key         | Meaning                                                              |
|-------------|------------------|----------------------------------------------------------------------|
| Wired       | `wiredConnected` | `yes` when an Ethernet connection is active.                         |
| Wi-Fi       | `wifiConnected`  | `yes` when the active default route is over Wi-Fi.                   |
| Wi-Fi SSID  | `wifiSSID`       | SSID of the connected Wi-Fi network, when applicable.                |
| Host IP     | `hostIp`         | Internal LAN IPv4 address used by Olares to reach other nodes.       |
| External IP | `externalIp`     | Public IPv4 address as observed by an external probe (refreshed at most once per minute). |

### Install / Uninstall

| Field          | JSON key                | Meaning                                                                       |
|----------------|-------------------------|-------------------------------------------------------------------------------|
| Installing     | `installingState`       | Lifecycle of the in-flight install: `in-progress`, `completed`, `failed`.     |
|                | `installingProgress`    | Free-form description of the current install step (shown as the inline note). |
| Uninstalling   | `uninstallingState`     | Lifecycle of the in-flight uninstall.                                         |
|                | `uninstallingProgress`  | Free-form description of the current uninstall step.                          |

### Upgrade

| Field          | JSON key                       | Meaning                                                                           |
|----------------|--------------------------------|-----------------------------------------------------------------------------------|
| Target         | `upgradingTarget`              | Target version of the in-flight upgrade.                                          |
| State          | `upgradingState`               | Lifecycle of the install phase of the upgrade.                                    |
|                | `upgradingProgress`            | Free-form progress message (inline note).                                         |
| Step           | `upgradingStep`                | Name of the current upgrade step.                                                 |
| Last error     | `upgradingError`               | Most recent error from the install phase.                                         |
| Download state | `upgradingDownloadState`       | Lifecycle of the download phase.                                                  |
|                | `upgradingDownloadProgress`    | Free-form progress message for the download (inline note).                        |
| Download step  | `upgradingDownloadStep`        | Name of the current download step.                                                |
| Download error | `upgradingDownloadError`       | Most recent error from the download phase.                                        |
| Retry count    | `upgradingRetryNum`            | Number of times the upgrader has retried after a transient failure (only shown when > 0). |
| Next retry at  | `upgradingNextRetryAt`         | Wall-clock time at which the next retry will fire (only shown when set).          |

### Logs collection

| Field       | JSON key                 | Meaning                                                          |
|-------------|--------------------------|------------------------------------------------------------------|
| State       | `collectingLogsState`    | Lifecycle of the most recent log collection job triggered through olaresd. |
|             | `collectingLogsError`    | Error from the most recent log collection job (inline note).     |

### Pressures

The `pressures` array lists kubelet node-condition pressures that are currently true on this node. The grouped output shows `(none)` when the node is healthy.

| Field   | JSON key  | Meaning                                                                |
|---------|-----------|------------------------------------------------------------------------|
| Type    | `type`    | Kubernetes node condition type, e.g. `MemoryPressure`, `DiskPressure`. |
| Message | `message` | Human-readable explanation provided by kubelet.                        |

### Other

| Field          | JSON key            | Meaning                                                                |
|----------------|---------------------|------------------------------------------------------------------------|
| FRP enabled    | `frpEnable`         | Whether the FRP-based reverse tunnel is turned on (sourced from `FRP_ENABLE`). |
| FRP server     | `defaultFrpServer`  | FRP server address (sourced from `FRP_SERVER`).                        |
| Container mode | `containerMode`     | Set when olaresd is running inside a container (sourced from `CONTAINER_MODE`). |

## State values

The `terminusState` field can take one of the following values. The CLI sources its descriptions from the same enumeration, so the table below stays in sync with the CLI output.

| Value                | Meaning                                                                  |
|----------------------|--------------------------------------------------------------------------|
| `checking`           | olaresd has not finished the first status probe yet.                     |
| `network-not-ready`  | No usable internal IPv4 address detected.                                |
| `not-installed`      | Olares is not installed on this node.                                    |
| `installing`         | Olares is currently being installed.                                     |
| `install-failed`     | The most recent install attempt failed.                                  |
| `uninitialized`      | Olares is installed but the admin user has not been activated yet.       |
| `initializing`       | The admin user activation is in progress.                                |
| `initialize-failed`  | The admin user activation failed.                                        |
| `terminus-running`   | Olares is running normally.                                              |
| `restarting`         | The node was just restarted; status will stabilize shortly.              |
| `invalid-ip-address` | The node IP changed since install; run `change-ip` to recover.           |
| `ip-changing`        | A `change-ip` operation is in progress.                                  |
| `ip-change-failed`   | The most recent `change-ip` attempt failed.                              |
| `system-error`       | One or more critical pods are not running.                               |
| `self-repairing`     | olaresd is attempting automatic recovery.                                |
| `adding-node`        | A worker node is being joined.                                           |
| `removing-node`      | A worker node is being removed.                                          |
| `uninstalling`       | Olares is being uninstalled.                                             |
| `upgrading`          | An upgrade is being applied.                                             |
| `disk-modifing`      | The storage layout is being modified.                                    |
| `shutdown`           | The system is shutting down.                                             |
