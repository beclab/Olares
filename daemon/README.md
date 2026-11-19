
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
| GET    | `/containerd/registry/configs/`           | List saved registry configurations  |
| GET    | `/containerd/registry/configs/:registry`  | Get registry configuration          |
| PUT    | `/containerd/registry/configs/:registry`  | Update supplied configuration entries |
| DELETE | `/containerd/registry/configs/:registry`  | Delete complete registry configuration |

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
### Registry configuration

Olaresd manages registry settings on the node receiving the request. Settings
must apply the configuration to every node that may run the application. These
routes require authorization and apply only to the local node:

- `GET /containerd/registry/configs/`: return saved configurations keyed by
  registry hostname, including registries without mirrors or downloaded images.
- `GET /containerd/registry/configs/:registry`: return `server`, `skip_verify`,
  `has_auth`, and the ordered mirror list `endpoint`. Passwords/tokens and the
  Authorization header are never returned.
- `PUT /containerd/registry/configs/:registry`: create or update the supplied
  configuration entries and return the resulting complete configuration.
- `DELETE /containerd/registry/configs/:registry`: remove the complete registry
  configuration, including upstream, mirrors, and credentials.

The path parameter is normalized to a hostname with an optional port. URL-style
input such as `https://gitlab.example.com:5050` is accepted: its scheme is stripped
and the registry namespace becomes `gitlab.example.com:5050`. URL-encode the path
parameter when sending a full address (for example, `encodeURIComponent(registry)`);
unescaped scheme URLs are also accepted by olaresd. Repository paths, userinfo,
query strings and fragments are rejected after normalization. The input scheme
does not set the upstream protocol: use the `server` body entry for HTTP/HTTPS;
omitting `server` still preserves its current value or defaults to HTTPS.
All configuration is managed through these routes; the old
`/containerd/registry/mirrors` routes have been removed. The existing Settings
frontend needs to switch to the config routes. Clearing mirrors must use
`PUT {"endpoint":[]}`, rather than DELETE, to preserve upstream credentials/TLS.

A complete GitLab example request body is:

```json
{
  "server": "https://gitlab.example.com",
  "skip_verify": false,
  "auth": {"username": "alice", "password": "<read_registry token>"},
  "endpoint": ["https://mirror.example.com"]
}
```

Each top-level entry has independent update semantics:

- Omit an entry to preserve its current value. A request containing only
  `{"auth":{"username":"alice","password":"<new token>"}}` preserves the server,
  TLS setting, and mirrors. A request containing only `endpoint` preserves auth.
- `server`: a supplied string replaces the upstream URL. An empty string resets
  it to the canonical HTTPS endpoint; `null` is rejected. The hostname must match
  the registry (Docker Hub uses `registry-1.docker.io`); HTTP may be explicit.
- `skip_verify`: a supplied boolean replaces the value, including explicit
  `false`. `null` is rejected.
- `auth`: supply the complete username/password-or-token pair to replace it,
  send `null` to remove credentials, or omit it to retain them. Credentials are
  write-only; `has_auth` in responses indicates whether they are configured.
  A supplied credential object must pass an upstream login check before any
  changes are saved. Omitting `auth` or sending `auth:null` skips this check.
- `endpoint`: a supplied list replaces all mirrors in priority order, highest
  first. `[]` or `null` clears mirrors. Retained hosts keep their existing TLS,
  headers, capabilities, and path settings; newly added hosts default to
  `pull` and `resolve`. Duplicate or invalid endpoints are rejected before saving.

For a new registry, omitted entries default to its canonical HTTPS server,
`skip_verify:false`, no credentials, and `endpoint:[]`. A GET for an unconfigured
registry returns the same defaults without creating a file. `PUT {}` leaves the
configuration untouched. Invalid updates leave the previous hosts file intact.

Credential validation uses the resulting server and TLS settings, contacts
`/v2/` directly without mirrors, and handles Basic or Bearer authentication with
a 15-second total timeout. It does not download an image or verify access to a
particular repository. Existing upstream CA/client certificate settings are
honored. If verification fails, neither credentials nor any other supplied
configuration entries are saved.

Validation errors use the standard error envelope with these codes:

- `400`: the registry rejects the credentials/access, or its TLS file settings
  cannot be loaded.
- `422`: credentials cannot be verified (the registry permits anonymous access,
  offers no supported challenge, or returns an invalid authentication URL or an
  HTTPS-to-HTTP authentication downgrade).
- `502`: connection/TLS/authentication-service failure or an unexpected response.
- `504`: validation timed out.

Redirects are not followed. Bearer credentials are sent to the authentication
service selected by the configured registry's `WWW-Authenticate` challenge,
including services on other hostnames or ports, without a platform allowlist.
The realm must be an absolute HTTP/HTTPS URL without userinfo or a fragment.
HTTPS registries require an HTTPS authentication service; explicit HTTP
registries may use HTTP or HTTPS authentication services.
Remote error bodies, passwords and temporary tokens are not returned to callers.

A successful PUT/GET response uses the standard olaresd response envelope:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "server": "https://gitlab.example.com",
    "skip_verify": false,
    "has_auth": true,
    "endpoint": ["https://mirror.example.com"]
  }
}
```

The collection GET uses the same configuration objects under registry keys:
`{"code":200,"message":"success","data":{"gitlab.example.com":{...}}}`.
An empty collection is `{}` in `data`; empty mirror lists are always `[]`.

Hosts configuration is written atomically to
`/etc/containerd/certs.d/<registry>/hosts.toml`. Existing mirror order and retained
hosts' TLS, headers, and capabilities are preserved. Clearing mirror endpoints
no longer deletes separately configured upstream credentials. Configured
registries appear in `/containerd/registries` before any image has been pulled.

Registry credentials are stored only in the upstream's `hosts.toml` Authorization
header as `Basic base64(username:password-or-token)`. Hosts files have mode `0600`.
Containerd reads hosts configuration on every pull; adding, updating, or removing
credentials does not write CRI auth drop-ins or restart containerd. Olares ships
containerd 2.1.3, which also passes hosts headers to Bearer token requests. This
supports GitLab's standard Basic-authenticated token exchange, using an account
password or an access token with `read_registry` permission. Accounts with 2FA
must use an access token. Host configuration is loaded again on the next pull.

App-service's metadata inspection and image-service pre-pull use the same
containerd hosts resolver for **all** registries. They keep original image names
and honor per-host TLS, HTTP, path overrides, capabilities, mirror order and
upstream fallback. Basic credentials from a hosts header also participate in
Basic/Bearer challenge authentication. Changes are read on the next request,
without an app-service/image-service restart. CronWorkflow admission keeps the
original registry namespace (normalizing shorthand Docker references only), so
node-local containerd applies the same registry configuration to workflow Pods.

Settings still needs to expose an add/edit registry form and call these routes;
this change supplies the olaresd and app-service backend support. Integration
validation should use a private GitLab/Gitea/GHCR repository, both image metadata
inspection and pre-pull, and a Pod with `imagePullPolicy: Always` on each node.

Requirement references:
[Market Source registry scenarios](https://w8wvyn24o5.feishu.cn/wiki/Ex7hwdGFrijJnlkDKescJo5Yn2d),
[Registry UI requirements](https://w8wvyn24o5.feishu.cn/wiki/Z99ewwRlRi2uK4kVi5lcx3IJnmd),
[containerd registry hosts](https://github.com/containerd/containerd/blob/v2.1.3/docs/hosts.md),
[GitLab registry authentication](https://docs.gitlab.com/user/packages/container_registry/authenticate_with_container_registry/).
