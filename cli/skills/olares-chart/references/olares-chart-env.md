# Environment variables — system / user / app level

> **Prerequisite:** read the parent [`../SKILL.md`](../SKILL.md). Env wiring is the configuration half of refinement §3 in [olares-chart-manifest.md](olares-chart-manifest.md) (middleware values are siblings of this).

Olares exposes configuration to an app through env vars at **three levels**. The app declares only the app-level ones (in `OlaresManifest.yaml` `envs[]`); the other two are platform-managed and consumed by reference.

| Level | Names | Who owns / sets it | Scope | App access |
|---|---|---|---|---|
| **System** | `OLARES_SYSTEM_*` | installer / admin, cluster-wide | shared by all users | read-only, via `valueFrom` |
| **User** | `OLARES_USER_*` | each user (SMTP, API keys, tokens) | per-user, isolated | read-only, via `valueFrom` |
| **App** | any app-local name in `envs[]` | the chart author | this app install | declared directly; surfaced as `.Values.olaresEnv.<name>` |

> **Nothing is auto-injected into the container.** Declaring an env (or referencing a system/user var) only puts the value into `.Values.olaresEnv.<name>`. You MUST map it into the workload yourself:
>
> ```yaml
> env:
>   - name: APP_TOKEN
>     value: "{{ .Values.olaresEnv.APP_TOKEN }}"
> ```
>
> This is different from system-injected Helm values like `.Values.postgres.*` / `.Values.userspace.*` / `.Values.os.*` (see [olares-chart-manifest.md](olares-chart-manifest.md)) — those are not env vars and also need explicit mapping.

## When to declare an app-level env

| Scenario | How |
|---|---|
| User must supply a value at install (admin username/password, license key, an API key not covered by a user var) | `required: true` (+ `type: password` for secrets). With no `default`, Olares prompts the user during install. |
| One of a fixed set of choices | `options` (inline dropdown) or `remoteOptions` (list fetched from a URL) |
| Reuse an Olares-managed value (CDN, Hugging Face token/endpoint, SMTP, OpenAI/Anthropic key, GitHub token) | `valueFrom.envName: OLARES_SYSTEM_* / OLARES_USER_*` — don't re-ask the user for something Olares already holds |
| Value should be changeable after install | `editable: true` (add `applyOnChange: true` to restart consumers automatically) |
| Static config that never varies per install | bake it into the image or template — do **not** make it an env |

The classic case — **initialize an app's admin username and password** — is an app-level env with `required: true` and no `default`, so the user is forced to enter them on the install screen.

## Declaration fields

Each entry under `envs:` supports these fields ([app-env-vars.md](https://docs.olares.com/developer/develop/app-env-vars.html)):

| Field | Meaning |
|---|---|
| `envName` | App-local name; injected as `.Values.olaresEnv.<envName>`. **Must not start with `OLARES_USER`** — this is enforced because the skill sets `apiVersion: v3` ([oac/internal/manifest/envs.go](../../../../framework/oac/internal/manifest/envs.go); see [olares-chart-versioning.md](olares-chart-versioning.md)). Map user vars via `valueFrom` instead. |
| `default` | Developer-supplied fallback. Users cannot edit it. Used when no user value and no `valueFrom`. |
| `valueFrom.envName` | Map to a system/user variable. The entry then **inherits** `type` / `editable` / `regex` / etc. from the referenced var; local `default` / `options` / `type` are ignored. |
| `required` | `true` → must resolve to a non-empty value for install to proceed (see below). |
| `editable` | `true` → value can be changed after install. |
| `applyOnChange` | `true` → changing it auto-restarts apps/components that use it; `false` → takes effect only on upgrade/reinstall. |
| `type` | Value format for validation (see below). |
| `regex` | Regular expression the value must match. |
| `options` | Inline fixed list (`{title, value}`) → selection UI. |
| `remoteOptions` | URL returning a JSON array in the same shape as `options`. |
| `description` | Human-readable purpose shown in the UI. |

## Validation semantics

How `required`, `type`, and `regex` actually behave at install, from [`CheckAppEnvs`](../../../../framework/app-service/pkg/utils/app/validate.go) (`pkg/utils/app/validate.go`).

### required vs optional — two code paths

- **Plain var (no `valueFrom`):** the effective value is `value || default`.
  - `required: true` and effective value is empty → reported as **`missingValues`**; install is blocked and the UI prompts the user (this is the init username/password case).
  - `required: false` (default) → empty is allowed; install proceeds.
- **Mapped var (`valueFrom`):**
  - `required: true` → app-service lists `SystemEnv` and the owner's `UserEnv`, and checks the referenced name is **present and non-empty**; otherwise reported as **`missingRefs`**.
  - optional `valueFrom` → **not** existence-checked; an empty reference resolves to empty.

> `required` with no `default` prompts the user at install, and after install the value **cannot be set back to empty**.

### Field types (`type`)

`int` | `bool` | `url` | `ip` | `domain` | `email` | `string` | `password`. `password` is rendered masked in the UI. The type is used to format-validate the value before it is accepted.

### Choice types

`options` (inline) and `remoteOptions` (fetched from a URL) restrict the value to a fixed list and present a selection UI instead of free text.

### regex

The value must match `regex`; this is checked by `ValidateValue(effectiveValue)`, and a failure is reported as **`invalidValues`**. `type` and `regex` **stack** — e.g. a username can be `type: string` plus `regex: '^[a-z0-9]{3,20}$'`. Validation runs on the resolved value, so for an optional var that may be left empty, pair a strict `regex`/`type` with a `default` (or make it `required`) to avoid surprising empty-value behavior.

### How failures surface

App-service does **not** return a 5xx. It returns HTTP 200 with an embedded `code: 422` and payload `type: appenv`, split into three buckets:

```json
{
  "code": 422,
  "data": {
    "type": "appenv",
    "data": {
      "missingValues": [ /* required + empty */ ],
      "missingRefs":   [ /* required valueFrom, referenced var missing/empty */ ],
      "invalidValues": [ /* type/regex/options validation failed */ ]
    }
  }
}
```

The CLI renders these (e.g. as "missing required env var(s): …"). **`lint` does not check any of this** — it neither validates env values nor verifies that the template actually maps `.Values.olaresEnv.<name>` into a container `env:`. Only app-service validates, at install time.

### required / type / regex at a glance

| Declaration | Outcome at install |
|---|---|
| `required: true`, no `default`, no `valueFrom` | user is prompted; empty → `missingValues` |
| `required: true`, has `default` | `default` used if user leaves it; never empty |
| `required: true`, `valueFrom` → var unset/empty | `missingRefs` |
| `required: false`, no `default` | allowed empty; install proceeds |
| any, value fails `type`/`regex`/`options` | `invalidValues` |

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

From [build/system-env.yaml](../../../../build/system-env.yaml). Reference these via `valueFrom.envName`; an app cannot change them.

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

From [build/user-env.yaml](../../../../build/user-env.yaml). All are user-editable and per-user. Reference via `valueFrom.envName`.

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

## Caveats

- **Map it in the template.** `.Values.olaresEnv.<name>` does not reach the container until you write it into `env:`.
- **`lint` won't catch env mistakes.** Missing mappings, wrong `valueFrom` names, and bad values are only caught by app-service at install (the `appenv` 422 above).
- **`valueFrom` inherits** `type`/`editable`/`regex` from the referenced var; local `default`/`options`/`type` are ignored.
- **Install-time override.** A value can be supplied at install via the CLI `--env KEY=VALUE` (see [`olares-market`](../../olares-market/SKILL.md)).
- **`applyOnChange: false`** means edits only take effect on upgrade/reinstall — stopping and starting the app does nothing.
