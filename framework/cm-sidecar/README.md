# cm-sidecar

Watch ConfigMaps and write their contents to files as soon as they change.

Mounting a ConfigMap as a volume works, but kubelet refreshes the projected
files on its own sync loop, so the moment a change lands on disk is not
predictable (and with `subPath` the file is never refreshed at all). This
sidecar watches the API instead: every event triggers a write, so the delay is
the watch latency rather than a sync period.

## Behaviour

- Each ConfigMap key becomes a file name, its value the file content. Both
  `data` and `binaryData` are handled; `binaryData` is written as raw bytes.
- A file is rewritten only when its content actually differs, so the
  modification time is untouched for no-op updates.
- Writes go to a temporary file in the same directory and are then renamed, so
  a reader never sees a half-written file.
- A file is removed when its key disappears from the ConfigMap, when the
  ConfigMap is deleted, or when the ConfigMap stops matching the selector.

> **The target directory belongs to this sidecar.** Every sync compares the
> directory against the keys of all matching ConfigMaps and removes the top
> level regular files that are left over. Point `CONFIGMAP_TARGET_DIR` at a
> dedicated volume, not at a directory shared with other writers.
> Subdirectories are left untouched.

That is also what makes cleanup reliable. A watch event for a deleted ConfigMap
no longer carries the object, so its keys are unknown by then; comparing against
the directory as a whole removes the files anyway. Files left behind by a
previous run are cleaned up on startup for the same reason, including the case
where the ConfigMap was deleted while the sidecar was not running.

## Configuration

| Variable                   | Description                                                                     | Required | Default                    |
| -------------------------- | ------------------------------------------------------------------------------- | -------- | -------------------------- |
| `CONFIGMAP_TARGET_DIR`     | Directory the files are written to. Created if missing.                          | yes      | -                          |
| `CONFIGMAP_LABEL_SELECTOR` | Kubernetes label selector used to pick ConfigMaps.                               | no       | `bytetrade.io/cm-sidecar`  |
| `CONFIGMAP_NAMESPACE`      | Namespace to watch.                                                              | no       | the sidecar's own namespace |

`CONFIGMAP_LABEL_SELECTOR` takes the standard selector syntax, so all of these
work:

- `bytetrade.io/cm-sidecar` matches any ConfigMap carrying the label, whatever
  its value.
- `bytetrade.io/cm-sidecar=true` matches an exact value.
- `app=nginx,tier!=canary` combines requirements.

A `/healthz` endpoint is served on port `8081`. It reports ready only once the
watch cache has synced, which makes it usable to gate a dependent container on
the files being present.

## Usage

The sidecar and the application share a writable volume. The sidecar writes to
its own mount path; the application reads the same files through its own.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample
spec:
  selector:
    matchLabels:
      app: sample
  template:
    metadata:
      labels:
        app: sample
    spec:
      serviceAccountName: cm-sidecar
      containers:
        - name: app
          image: bash:5.2.15
          command: ["watch", "ls", "-l", "/etc/app-config"]
          volumeMounts:
            - name: config
              mountPath: /etc/app-config
        - name: cm-sidecar
          image: beclab/cm-sidecar:latest
          env:
            - name: CONFIGMAP_TARGET_DIR
              value: /etc/app-config
            - name: CONFIGMAP_LABEL_SELECTOR
              value: bytetrade.io/cm-sidecar
          volumeMounts:
            - name: config
              mountPath: /etc/app-config
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8081
      volumes:
        - name: config
          emptyDir: {}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: sample-config
  labels:
    bytetrade.io/cm-sidecar: "true"
data:
  app.conf: |
    listen = 8080
```

This writes `/etc/app-config/app.conf` in both containers.

### RBAC

The service account needs to read ConfigMaps in the watched namespace:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cm-sidecar
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cm-sidecar
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cm-sidecar
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cm-sidecar
subjects:
  - kind: ServiceAccount
    name: cm-sidecar
```

Setting `CONFIGMAP_NAMESPACE` to another namespace requires a Role and
RoleBinding in that namespace instead.

## Development

```bash
make          # fmt, vet and build into output/cm-sidecar
make test     # unit tests, no cluster needed
make tidy
```

Running locally against the current kubecontext:

```bash
CONFIGMAP_TARGET_DIR=/tmp/cm CONFIGMAP_NAMESPACE=default make run
```
