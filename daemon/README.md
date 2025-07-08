
# Olares System Daemon (`olaresd`)

`olaresd` is the foundational process that boots on every Olares node. It runs as a `systemd` service on port `18088`, exposing a secure REST API for hardware abstraction, network orchestration, storage management, and turnkey cluster operations—all before Kubernetes starts. 

Olaresd is installed as a systemd service in `/etc/systemd/system/olaresd.service`.

## Key features

- **System monitoring**: Continuous health checks of cluster and node status.
- **Cluster lifecycle management**: Automated install, upgrade, IP-switching, restart, and maintenance operations.
- **Hardware Abstraction**: USB auto-mounting, storage provisioning, and management.
- **Network Management**: mDNS service discovery, WiFi onboarding, and network interface control.

## REST API reference

The daemon provides an authenticated REST API (using signature-based auth):

**Base URL**: `http://<node-ip>:18088`

### System commands `/command/`

**Lifecycle operations**

| Method | Endpoint                    | Description                  |
|--------|-----------------------------|------------------------------|
| POST   | `/command/install`          | Install Olares               |
| POST   | `/command/uninstall`        | Uninstall Olares             |
| POST   | `/command/upgrade`          | Upgrade Olares               |
| DELETE | `/command/upgrade`          | Cancel upgrade               |
| POST   | `/command/reboot`           | Reboot node                  |
| POST   | `/command/shutdown`         | Shutdown node                |

**Network configuration**

| Method | Endpoint                    | Description                  |
|--------|-----------------------------|------------------------------|
| POST   | `/command/connect-wifi`     | Connect to WiFi              |
| POST   | `/command/change-host`      | Change Olares IP binding     |

**Storage management**

| Method | Endpoint                          | Description                        |
|--------|-----------------------------------|------------------------------------|
| POST   | `/command/mount-samba`            | Mount SMB shares                   |
| POST   | `/command/v2/mount-samba`         | Enhanced SMB mounting              |
| POST   | `/command/umount-samba`           | Unmount SMB shares                 |
| POST   | `/command/umount-samba-incluster` | Cluster-wide SMB unmount           |
| POST   | `/command/umount-usb`             | Unmount USB device                 |
| POST   | `/command/umount-usb-incluster`   | Cluster-wide USB unmount           |

**System Maintenance**

| Method | Endpoint                    | Description                         |
|--------|-----------------------------|-------------------------------------|
| POST   | `/command/collect-logs`     | Collect system logs for diagnostics |

---

### System information (`/system/`)

**System status**

| Method | Endpoint                 | Description                 |
|--------|--------------------------|-----------------------------|
| GET    | `/system/status`         | Get full system status      |
| GET    | `/system/ifs`            | List network interfaces     |
| GET    | `/system/hosts-file`     | View `/etc/hosts`           |
| POST   | `/system/hosts-file`     | Update `/etc/hosts`         |

**Mount information**

| Method | Endpoint                        | Description                    |
|--------|---------------------------------|--------------------------------|
| GET    | `/system/mounted-usb`           | Mounted USB devices            |
| GET    | `/system/mounted-hdd`           | Mounted hard drives            |
| GET    | `/system/mounted-smb`           | Mounted SMB shares             |
| GET    | `/system/mounted-path`          | All mount points               |

**Cluster-wide mounts**

| Method | Endpoint                             | Description                      |
|--------|--------------------------------------|----------------------------------|
| GET    | `/system/mounted-usb-incluster`      | USB mounts in cluster            |
| GET    | `/system/mounted-hdd-incluster`      | HDD mounts in cluster            |
| GET    | `/system/mounted-smb-incluster`      | SMB mounts in cluster            |
| GET    | `/system/mounted-path-incluster`     | All cluster mounts               |

---

### Container management (`/containerd/`)

**Registry Management**

| Method | Endpoint                                  | Description                         |
|--------|-------------------------------------------|-------------------------------------|
| GET    | `/containerd/registries`                  | List registries                     |
| GET    | `/containerd/registry/mirrors/`           | List registry mirrors               |
| GET    | `/containerd/registry/mirrors/:registry`  | Get specific registry mirror        |
| PUT    | `/containerd/registry/mirrors/:registry`  | Update registry mirror              |
| DELETE | `/containerd/registry/mirrors/:registry`  | Delete registry mirror              |

**Image Management**

| Method | Endpoint                         | Description                    |
|--------|----------------------------------|--------------------------------|
| GET    | `/containerd/images/`            | List container images          |
| DELETE | `/containerd/images/:image`      | Delete specific image          |
| POST   | `/containerd/images/prune`       | Remove unused images           |


## Build from source

### Prerequisites

* Go 1.24+
* GoReleaser (Optional, for creating release artifacts)

### Steps

1.  **Navigate to the daemon directory:**

    ```bash
    cd daemon
    ```

2.  **Build for your host OS/architecture:**

    ```bash
    go build -o olaresd ./cmd/olaresd/main.go
    ```

3.  **Cross-compile for another target (e.g., Linux AMD64):**

    ```bash
    GOOS=linux GOARCH=amd64 go build -o olaresd ./cmd/olaresd/main.go
    ```

4.  **Produce release artifacts (optional):**

    ```bash
    goreleaser release --snapshot --clean
    ```

## Extend `olaresd`

To add a new command API:

1.  **Define command**: Add a new command struct in `pkg/commands/`.
2.  **Implement handler**: Create the corresponding HTTP handler logic in `internal/apiserver/handlers/`.
3.  **Register route**: Register the new API route in `internal/apiserver/server.go`.
4.  **Update state**: If the command modifies the cluster's state, ensure you update the logic in `pkg/cluster/state/`.
5.  **Validate**: Run `go vet ./... && go test ./...` to check for issues and ensure all tests pass before opening a pull request.


### Test a custom build

1.  Copy the binary to your Olares node.

2.  On the node, replace the existing binary:

    ```bash
    # Move the new binary into place
    sudo cp -f /tmp/olaresd /usr/local/bin/

3. Restart the daemon to apply changes:
    
   ``` 
   sudo systemctl restart olaresd
   ```