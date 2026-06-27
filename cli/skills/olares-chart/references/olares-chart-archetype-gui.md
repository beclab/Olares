# Archetype: GUI desktop application (browser-streamed)

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md) first. Use this when the upstream is a native desktop GUI with no web UI; apply the recipe, then continue with Refine -> lint (the Manifest refinement areas).

## When this archetype fits

A native Linux GUI app (X11/Wayland) with **no web UI** — a browser, office suite, IDE, CAD tool. The Olares challenge: there is no HTTP page to point an entrance at, only a desktop window. The recipe: wrap the app in a **web-desktop base image** that streams the GUI to the browser, then point one visible entrance at that web server.

**Signals**
- Distributed as `.deb`/AppImage/native binary that opens a desktop window; needs an X11/Wayland display (see Xvfb-style headless display).
- LinuxServer.io already ships a container for it (browsers, office, IDEs, media tools), or the upstream is a plain desktop binary.
- No web framework, no HTTP port — `docker run` alone would fail with no display.

## Olares mapping — pick the streaming base

Picking the base image is the core decision. Flow: the user's browser hits a visible entrance on HTTP `:3000` → the **web-desktop base image** serves the web client and streams the GUI of the desktop app running inside the same container. The base image is the only thing you choose; the app itself is unchanged. Two base routes:

| Base | Lineage | Strengths | Costs | Pick when |
|---|---|---|---|---|
| **Selkies** (default, `latest`) | WebRTC / video stream (PixelFlux) | H.264/H.265/AV1, Wayland, iGPU/NVENC zero-copy, ~60fps | `x86_64` needs AVX2 (else auto-falls back to X11); HTTPS required; best with a GPU | video / animation / 3D / high-interaction apps, modern hardware |
| **KasmVNC** (`kasm` tag) | VNC/RFB (a fork of noVNC + TigerVNC, RFB-incompatible) | damage-based pixel transfer, lightweight, no-AVX2 ok | weaker at high-fps video | static / office / form-style UIs, older hardware |

**One line:** Selkies is the default; KasmVNC is the fallback for old hardware / static UIs. `noVNC` and `Xvfb` are lower-level building blocks these bases already absorb — you don't select them directly.

## Templates (copy-pasteable minimum)

A single Deployment (`strategy: Recreate` — the `/config` mount can't roll), uid-1000 via `PUID/PGID`, `/config` persisted into userspace, a memory `emptyDir` for `/dev/shm` (the image's `shm_size` requirement), and an HTTP entrance on **3000** (Olares' entrance terminates TLS, so point at 3000, not the self-signed 3001):

```yaml
# templates/<app>.yaml — Deployment (CPU-only baseline)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <app>
  namespace: {{ .Release.Namespace }}
  labels: { io.kompose.service: <app> }
spec:
  replicas: {{ .Values.workloads.<app>.replicaCount }}
  strategy: { type: Recreate }
  selector:
    matchLabels: { io.kompose.service: <app> }
  template:
    metadata:
      labels: { io.kompose.service: <app> }
    spec:
      enableServiceLinks: false   # avoid k8s injecting <SVC>_PORT=tcp://... that can clobber app config env; see the Env area
      containers:
        - name: <app>
          image: "docker.io/<your-namespace>/<app>:<pinned-tag>"   # never :latest
          env:
            - { name: PUID, value: "1000" }
            - { name: PGID, value: "1000" }
            - { name: TZ, value: Etc/UTC }
          ports:
            - { name: http, containerPort: 3000, protocol: TCP }
            - { containerPort: 3001 }
          volumeMounts:
            - { mountPath: /config, name: config }
            - { mountPath: /dev/shm, name: dshm }
      volumes:
        - name: config
          hostPath:
            type: DirectoryOrCreate
            path: {{ .Values.userspace.appCache }}/config
        - name: dshm
          emptyDir: { medium: Memory, sizeLimit: 1Gi }
```

```yaml
# OlaresManifest.yaml — one visible window entrance
permission:
  appCache: true    # matches the /config mount below
entrances:
- name: <app>
  host: <app>
  port: 3000        # HTTP — Olares' entrance does TLS; do not use 3001
  title: <App>
  openMethod: window
```

> **Where `/config` lives.** The example follows official chromium and maps `/config` into `appCache` (node-local `/Cache/<app>`, fast but evictable). If the app's profile/settings must survive node migration, mount `appData` (`/Data/<app>`) instead and declare `permission.appData: true` — `appCache`/`appData` already include the `/<app>` suffix, so `{{ .Values.userspace.appCache }}/config` just organizes within it.

## Integrated GPU acceleration (iGPU / VAAPI, self-contained)

Integrated graphics ride along with the CPU/SoC (e.g. the Intel iGPU on Olares One). This is plain **`/dev/dri` device passthrough + VAAPI video encoding**, a capability of the streaming base itself — it is **not** the NVIDIA CUDA / `spec.accelerator` path in the GPU / models capability, and this section does not depend on it.

Gate the acceleration on the device using the injected `.Values.deviceName` (documented in the system-injected Helm values reference §A); other devices fall through to the CPU-only baseline above so the chart stays portable:

- **Rendering vs encoding:** `DRINODE` = render node for EGL/3D, `DRI_NODE` = encode node for VAAPI. Pointing both at the same `/dev/dri/renderD128` lets Selkies enable **zero-copy** encoding (large CPU/latency drop).
- **Intel iGPU:** set `LIBVA_DRIVER_NAME=iHD` (intel-media-driver) and pass the base's zero-copy launch flags (`--use-gl=egl --enable-zero-copy --enable-features=VaapiVideoDecoder,VaapiVideoEncoder,...`); the accelerated branch also opens the Selkies websocket port `8082`. The flag-passing env var is **app-specific** — chromium uses `CHROME_CLI`, other apps expose their own (or none).
- **Node access:** only the accelerated branch sets `securityContext.privileged: true` and `hostPath`-mounts `/dev/dri`.
- **Fallback:** Selkies on `x86_64` needs AVX2 or it auto-falls back to X11; a device with no iGPU keeps CPU encoding and still works, just at higher CPU usage.

Add these device-gated branches to the baseline spec above — wrap each in `{{- if eq .Values.deviceName "Olares One" }} … {{- end }}` so other devices keep the CPU-only path:

```yaml
# container: privileged + encode env + the 8082 websocket port + the /dev/dri mount
securityContext: { privileged: true }
env:
  - { name: DRINODE, value: /dev/dri/renderD128 }    # EGL/3D render node
  - { name: DRI_NODE, value: /dev/dri/renderD128 }   # VAAPI encode node (same -> zero-copy)
  - { name: LIBVA_DRIVER_NAME, value: iHD }
  - name: CHROME_CLI                                 # app-specific flag var (chromium); others differ or none
    value: "--ignore-gpu-blocklist --enable-zero-copy --use-gl=egl --enable-features=VaapiVideoDecoder,VaapiVideoEncoder"
ports:
  - { containerPort: 8082 }                          # Selkies websocket
volumeMounts:
  - { mountPath: /dev/dri, name: dev-dri }
# pod volumes:
volumes:
  - { name: dev-dri, hostPath: { path: /dev/dri } }
```

## Canonical example

[chromium](https://github.com/beclab/apps/tree/main/chromium) (`appVersion: …-selkies`) — the exact `{{- if eq .Values.deviceName "Olares One" }}` iGPU branch above is lifted from its [template](https://github.com/beclab/apps/blob/main/chromium/templates/chromium.yaml) and [OlaresManifest.yaml](https://github.com/beclab/apps/blob/main/chromium/OlaresManifest.yaml); [firefox](https://github.com/beclab/apps/tree/main/firefox) is the same shape for the KasmVNC-era base.

## Hard rules
- Entrance points at **HTTP 3000**, not the self-signed `3001` — Olares' entrance handles TLS.
- The base image's built-in `CUSTOM_USER`/`PASSWORD` is "keep the kids out" only; real exposure relies on Olares' entrance auth.
- Only `/config` persists — anything written elsewhere is lost on recreate. Map app data/settings under `/config`.
- Pin the image tag (never `:latest`) and keep `supportArch` consistent with the base (NVIDIA is not available on the Alpine-based bases).
- The iGPU branch (`privileged` + `/dev/dri`) is **device-gated** — never make it unconditional, or the chart won't run on devices without that GPU.
