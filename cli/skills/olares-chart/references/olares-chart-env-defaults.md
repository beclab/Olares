# Environment variables — worked examples & default variables

> **Prerequisite:** read [olares-chart-env.md](olares-chart-env.md) first for the three levels, declaration fields, and validation semantics. This file is the reference half: copy-pasteable examples plus the default `OLARES_SYSTEM_*` / `OLARES_USER_*` variables you can map via `valueFrom`.

## Worked examples

### Initialize admin credentials at install

```yaml
# OlaresManifest.yaml
envs:
  - envName: ADMIN_USERNAME
    required: true
    type: string
    editable: false
    regex: '^[a-z0-9]{3,20}$'
    description: Admin username created on first launch
  - envName: ADMIN_PASSWORD
    required: true
    type: password
    editable: false
    regex: '^.{8,}$'
    description: Admin password (min 8 chars)
```

```yaml
# templates/deployment.yaml
        env:
        - name: APP_ADMIN_USER
          value: "{{ .Values.olaresEnv.ADMIN_USERNAME }}"
        - name: APP_ADMIN_PASSWORD
          value: "{{ .Values.olaresEnv.ADMIN_PASSWORD }}"
```

Both are `required` with no `default`, so the user must fill them in on the install screen.

### Reuse a user-level variable (don't re-ask)

```yaml
envs:
  - envName: HF_TOKEN
    required: false
    applyOnChange: true
    valueFrom:
      envName: OLARES_USER_HUGGINGFACE_TOKEN   # type/editable/regex inherited
```

```yaml
        env:
        - name: HUGGING_FACE_HUB_TOKEN
          value: "{{ .Values.olaresEnv.HF_TOKEN }}"
```

### Optional value with a default (no prompt)

```yaml
envs:
  - envName: API_BASE_URL
    required: false
    type: url
    default: "https://api.example.com"
    editable: true
```

## Default system-level variables

Olares ships these system-level variables. Reference them via `valueFrom.envName`; an app cannot change them.

| Variable | Type | Default | Editable | Required |
|---|---|---|---|---|
| `OLARES_SYSTEM_REMOTE_SERVICE` | url | `https://api.olares.com` | yes | yes |
| `OLARES_SYSTEM_CDN_SERVICE` | url | `https://cdn.olares.com` | yes | yes |
| `OLARES_SYSTEM_DOCKERHUB_SERVICE` | url | — | yes | no |
| `OLARES_SYSTEM_ROOT_PATH` | — | `/olares` | no | yes |
| `OLARES_SYSTEM_ROOTFS_TYPE` | — | `fs` | no | yes |
| `OLARES_SYSTEM_CUDA_VERSION` | — | — | no | no |
| `OLARES_SYSTEM_HUGGINGFACE_SERVICE` | url | `https://huggingface.co/` | yes | no |
| `OLARES_SYSTEM_HUGGINGFACE_TOKEN` | password | — | yes | no |

`REMOTE_SERVICE` is the unified base for several legacy endpoints (DID, Olares Space, FRP, Tailscale control plane, Market provider, cert/DNS service). `ROOT_PATH` / `ROOTFS_TYPE` / `CUDA_VERSION` are also read directly at render time by app-service (e.g. `.Values.rootPath`, `.Values.GPU.Cuda`).

## Default user-level variables

Olares ships these user-level variables. All are user-editable and per-user. Reference them via `valueFrom.envName`.

**User info**

| Variable | Type | Default |
|---|---|---|
| `OLARES_USER_EMAIL` | string | — |
| `OLARES_USER_USERNAME` | string | — |
| `OLARES_USER_PASSWORD` | password | — |
| `OLARES_USER_TIMEZONE` | string | — |

**SMTP**

| Variable | Type | Default |
|---|---|---|
| `OLARES_USER_SMTP_ENABLED` | bool | — |
| `OLARES_USER_SMTP_SERVER` | domain | — |
| `OLARES_USER_SMTP_PORT` | number | — |
| `OLARES_USER_SMTP_USERNAME` | string | — |
| `OLARES_USER_SMTP_PASSWORD` | password | — |
| `OLARES_USER_SMTP_FROM_ADDRESS` | email | — |
| `OLARES_USER_SMTP_SECURE` | bool | `true` |
| `OLARES_USER_SMTP_USE_TLS` | bool | — |
| `OLARES_USER_SMTP_USE_SSL` | bool | — |
| `OLARES_USER_SMTP_SECURITY_PROTOCOLS` | string (options: `tls`/`ssl`/`starttls`/`none`) | — |

**AI keys**

| Variable | Type | Default |
|---|---|---|
| `OLARES_USER_OPENAI_APIKEY` | password | — |
| `OLARES_USER_CUSTOM_OPENAI_SERVICE` | url | — |
| `OLARES_USER_CUSTOM_OPENAI_APIKEY` | password | — |
| `OLARES_USER_ANTHROPIC_APIKEY` | password | — |

**Mirrors & tokens**

| Variable | Type | Default |
|---|---|---|
| `OLARES_USER_HUGGINGFACE_SERVICE` | url | `https://huggingface.co/` |
| `OLARES_USER_HUGGINGFACE_TOKEN` | password | — |
| `OLARES_USER_PYPI_SERVICE` | url | `https://pypi.org/simple/` |
| `OLARES_USER_GITHUB_SERVICE` | url | `https://github.com/` |
| `OLARES_USER_GITHUB_TOKEN` | password | — |
