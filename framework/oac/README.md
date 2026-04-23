# oac API & OlaresManifest Validation Rules

`oac` lints an Olares App (aka `oac`) Helm chart directory: it validates `OlaresManifest.yaml`, dry-runs the chart, collects the image list, and checks container resource limits and ServiceAccount RBAC permissions.

This document covers two things:

1. [Public API](#public-api): how external projects should call into the library
2. [OlaresManifest.yaml validation rules](#olaresmanifestyaml-validation-rules): what is currently being checked

---

## Public API

All exported symbols live in a single public package:

| Package                                    | Role                                                                              |
|---|---|
| `github.com/beclab/Olares/framework/oac`   | Main entry point: build a `Checker`, run lint/validate, load config, list images; `AppConfiguration` is re-exported as an alias |

Sub-packages under `internal/` are not public — external callers should not (and cannot) import them directly.

> **Schema types come straight from upstream**: `oac.AppConfiguration` and all of its nested structs are type aliases from **`github.com/beclab/api/manifest`** (`Entrance` / `ServicePort` / `TailScale` / `ACL` come from `github.com/beclab/api/api/app.bytetrade.io/v1alpha1`, `AppEnvVar` from `github.com/beclab/api/api/sys.bytetrade.io/v1alpha1`, and `LabelSelector` / `LabelSelectorRequirement` from `k8s.io/apimachinery/.../meta/v1`). `*oac.AppConfiguration` and `*apimanifest.AppConfiguration` share the same `reflect.Type` — when you need to build a `manifest.Chart{...}` / `appv1.Entrance{...}` value, **import the upstream package directly**; the `oac` package no longer re-exports them.

The `Manifest` interface ("the smallest common surface for a parsed manifest"):

```go
type Manifest interface {
    APIVersion() string        // "v1" / "v2" / "v3"
    ConfigVersion() string     // olaresManifest.version
    ConfigType() string        // olaresManifest.type, typically "app"
    AppName() string           // metadata.name
    AppVersion() string        // metadata.version
    Entrances() []EntranceInfo
    OptionsImages() []string   // options.images
    PermissionAppData() bool
    Raw() any                  // type-assert: m.Raw().(*oac.AppConfiguration)
                                // or just use oac.AsAppConfiguration(m)
                                // underlying type: *github.com/beclab/api/manifest.AppConfiguration
}
```

When you need the full `AppConfiguration` struct to access detailed fields, **don't** reach for `Raw()` — use the typed API in §4 below.

### 2. Constructor

```go
// Build a reusable Checker
c := oac.New(opts...)

// Load without validating
m, err  := c.LoadManifestFile(path)
m, err  := c.LoadManifestContent(bytes)

// Validate the manifest only (no helm, no folder check)
err := c.ValidateManifestFile(path)
err := c.ValidateManifestContent(bytes)
err := c.ValidateAppConfiguration(cfg)              // *AppConfiguration already in memory

// Full lint: folder layout + manifest + resources + optional checks
err := c.Lint(path)

// Individual checks
err := c.CheckChartFolder(path)                    // folder layout
err := c.CheckResources(path)                      // helm dry-run + container limits (§3.1); upload mount / workload naming are only enforced by Lint
err := c.CheckServiceAccountRules(path)            // RBAC forbidden rules
err := c.CheckSameVersion(path, m /* may be nil */) // Chart.yaml ↔ manifest version match

// List every image (helm-rendered workload images ∪ options.images, deduped and sorted)
imgs, err := c.ListImages(path)
// Same, but render the chart with .Values.GPU.Type = mode so GPU-gated
// templates contribute their workloads (empty mode == ListImages).
imgs, err := c.ListImagesForMode(path, "nvidia")
```

> Every helm dry-run (`Lint`, `ListImages` / `ListImagesForMode`, `CheckResources`, `CheckServiceAccountRules`, ...) deep-merges any values registered via `oac.WithValues(map[string]interface{}{...})` on top of the scaffold built by `helmrender.BuildValues`. External keys win on conflicts; map keys recurse so siblings the caller did not override survive. The per-mode `.Values.GPU.Type` set by `ListImagesForMode` and the resource-limit per-mode loop runs **after** the merge, so it still wins on `GPU.Type`.

### 3. Top-level convenience functions

If you don't want to hold on to a `Checker`, use these shortcuts (each is just `New(opts...).Xxx(...)`):

```go
err  := oac.Lint(path, opts...)
err  := oac.ValidateManifestFile(path, opts...)
err  := oac.ValidateManifestContent(bytes, opts...)
imgs, err := oac.ListImagesFromOAC(path, opts...)
imgs, err := oac.ListImagesFromOACForMode(path, "nvidia", opts...)

// Parse straight into a typed *AppConfiguration (no validation)
cfg, err  := oac.LoadAppConfiguration(path, opts...)
cfg, err  := oac.LoadAppConfigurationContent(bytes, opts...)

// Validate an already-built *AppConfiguration (same rules as ValidateManifestFile)
err := oac.ValidateAppConfiguration(cfg, opts...)

// Run both admin==owner and admin!=owner scenarios (Lint once for each)
err := oac.LintBothOwnerScenarios(path, extraOpts...)

// Peek the two top-level version fields without a full parse
v, err := oac.PeekManifestVersions(bytes)
// v.APIVersion             -> apiVersion            ("v1" / "v2" / "v3" / ...)
// v.OlaresManifestVersion  -> olaresManifest.version ("0.12.0" / ...)

// Resolve the effective resource envelope for a single spec.resources[] mode
// (see §7 below)
lim, err := oac.ResourceLimitsForResourceMode(cfg, "nvidia")
```

### 4. Typed access to `AppConfiguration`

`*Manifest.Raw()` returns `any`; you need a type assertion to get at the fields. Most callers just want "one line to get a typed cfg", so the root package provides a thin wrapper:

```go
// Alias: oac.AppConfiguration == github.com/beclab/api/manifest.AppConfiguration
// (same underlying type, same reflect.Type)
type AppConfiguration = apimanifest.AppConfiguration

// Checker methods (parse only, no validation)
func (c *Checker) LoadAppConfiguration(oacPath string) (*AppConfiguration, error)
func (c *Checker) LoadAppConfigurationContent(content []byte) (*AppConfiguration, error)

// Top-level convenience functions
func LoadAppConfiguration(oacPath string, opts ...Option) (*AppConfiguration, error)
func LoadAppConfigurationContent(content []byte, opts ...Option) (*AppConfiguration, error)

// Validate a parsed / hand-built *AppConfiguration with the exact same rules as
// ValidateManifestFile (field-level + cross-field; no helm / folder / resource-level
// checks and no customValidators — those need an oacPath and chart templates, which
// a bare *AppConfiguration can't provide).
//
// Behavior:
//   - Honors SkipManifestCheck(): returns nil immediately when enabled.
//   - Does not panic on cfg == nil; returns a *ValidationError the caller can
//     unwrap with errors.As.
//   - Failures are wrapped in *ValidationError; Version is taken from
//     cfg.APIVersion (defaults to "v1" when empty).
func (c *Checker) ValidateAppConfiguration(cfg *AppConfiguration) error

// Top-level shortcut (equivalent to New(opts...).ValidateAppConfiguration(cfg))
func ValidateAppConfiguration(cfg *AppConfiguration, opts ...Option) error

// Extract a *AppConfiguration from an existing Manifest (nil-safe; today's
// Strategy always returns true).
func AsAppConfiguration(m Manifest) (*AppConfiguration, bool)
```

If you only need to **read** fields, `oac.AppConfiguration` is enough:

```go
cfg, _ := oac.LoadAppConfiguration("./my-app")
fmt.Println(cfg.Metadata.Name, cfg.Spec.Resources[0].Mode)
for _, e := range cfg.Entrances { fmt.Println(e.Name, e.Port) }
```

When you need to **construct** nested structs (generating a manifest, writing test fixtures, patching a field), import the upstream type packages directly — the `oac` package does not forward them:

```go
import (
    oac "github.com/beclab/Olares/framework/oac"

    appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
    "github.com/beclab/api/manifest"
)

cfg := &oac.AppConfiguration{ // equivalent to *manifest.AppConfiguration
    APIVersion: manifest.APIVersionV1, // or just the literal "v1" if not exported upstream
    Metadata:   manifest.AppMetaData{Name: "demo", Title: "Demo", Version: "1.0.0"},
    Entrances:  []appv1.Entrance{{Name: "web", Host: "demo", Port: 8080}},
    Spec: manifest.AppSpec{
        SubCharts: []manifest.Chart{{Name: "extras", Shared: true}},
        Resources: []manifest.ResourceMode{{
            Mode: "nvidia",
            ResourceRequirement: manifest.ResourceRequirement{
                RequiredCPU: "200m", LimitedCPU: "1",
                RequiredGPU: "1",   LimitedGPU: "1",
            },
        }},
    },
    Options: manifest.Options{Upload: manifest.Upload{Dest: "/data", LimitedSize: 1024}},
}
```

**Schema type layout**

| Source package                                                       | Types it covers                                                     |
|---|---|
| `github.com/beclab/api/manifest`                                     | The vast majority: `AppConfiguration`, `AppMetaData`, `AppSpec`, `Chart`, `Provider`, `Permission`, `ProviderPermission`, `Policy`, `Dependency`, `Conflict`, `Options`, `ResetCookie`, `AppScope`, `WsConfig`, `Upload`, `OIDC`, the `Middleware` family (`Database`, `PostgresConfig`, `ArgoConfig`, `MinioConfig`, `Bucket`, `RabbitMQConfig`, `VHost`, `ElasticsearchConfig`, `Index`, `RedisConfig`, `MongodbConfig`, `MariaDBConfig`, `MySQLConfig`, `ClickHouseConfig`, `NatsConfig`, `Subject`, `Export`, `Ref`, `RefSubject`, `PermissionNats`), `Hardware` / `CpuConfig` / `GpuConfig` / `SupportClient`, `ConfigOverlay`, `ResourceRequirement` / `ResourceMode` / `SpecialResource` |
| `github.com/beclab/api/api/app.bytetrade.io/v1alpha1` (`appv1`)      | `Entrance`, `ServicePort`, `TailScale`, `ACL` |
| `github.com/beclab/api/api/sys.bytetrade.io/v1alpha1` (`sysv1alpha1`) | `AppEnvVar` |
| `k8s.io/apimachinery/pkg/apis/meta/v1` (`metav1`)                    | `LabelSelector`, `LabelSelectorRequirement` |

**Note**: `LoadAppConfiguration*` does **not** validate — it only runs the parse pipeline (legacy goes through template rendering, modern is parsed literally). If you need "parse + validate + get the config", call `LoadAppConfiguration` first, then `ValidateAppConfiguration(cfg)` (or `ValidateManifestContent(bytes)`).

**About the old `(*AppConfiguration).Validate()` method**: it has been removed — `AppConfiguration` is now a type alias of `github.com/beclab/api/manifest.AppConfiguration`, and Go does not allow methods on aliases. Use the `(*Checker).ValidateAppConfiguration(cfg)` method or the **package-level** `oac.ValidateAppConfiguration(cfg, opts...)` instead; the rules are identical.

**How options affect `ValidateAppConfiguration`**: currently only `SkipManifestCheck()` affects this entry point (returns nil immediately when on). Other options (owner/admin, SkipResourceCheck, SkipAppDataCheck / WithAppDataValidator, ...) need helm rendering, the chart directory, or chart templates — none of which a bare `*AppConfiguration` can provide — so passing them is a silent no-op.

**Migrating from `appcfg`**: the same types were once also re-exported through the `github.com/beclab/Olares/framework/oac/appcfg` sub-package, which has been removed. Callers should import the three upstream packages listed above directly; the aliases match, so the types are fully compatible — just swap the import and prefix (`appcfg.Chart` → `manifest.Chart`, `appcfg.Entrance` → `appv1.Entrance`, `appcfg.LabelSelector` → `metav1.LabelSelector`).

---

### 5. Options

| Option | Default     | Description |
|---|-------------|---|
| `WithOwner(name)` | empty       | Sets `.Values.bfl.username` in the chart template; empty values are ignored |
| `WithAdmin(name)` | empty       | Sets `.Values.admin` in the chart template; empty values are ignored |
| `WithOwnerAdmin(name)` | —           | owner and admin take the same value ("admin self-install" scenario) |
| `WithAutoOwnerScenarios()` | off         | `Lint` ignores any explicit `WithOwner*` above and automatically renders + checks workloads for both `owner==admin` and `owner!=admin` installation scenarios; both must pass. Manifest validation, folder checks, and same-version checks run only once (they don't depend on owner). Equivalent to `LintBothOwnerScenarios` |
| `WithoutAutoOwnerScenarios()` | —           | Clears `autoOwner` and falls back to a single run using the explicit owner/admin values. Useful for turning the auto mode off at a specific call site when composing option sets |
| `SkipFolderCheck()` | off         | Skip the folder-layout check |
| `SkipManifestCheck()` | off         | Skip manifest structural validation |
| `SkipResourceCheck()` | off         | Skip **container resource limit** checks. Note: upload mount point and workload naming are structural integrity checks and are always enforced by `Lint`, regardless of this option |
| `SkipSameVersionCheck()` | **on**      | Disable the Chart.yaml ↔ manifest version match check. It is on by default; turn it off when you want to align version numbers separately before publishing to the app store |
| `WithSameVersionCheck()` | —           | Turn the match check back on (useful when composing option sets to re-enable it at a specific call site) |
| `WithServiceAccountRulesCheck()` | off         | Enable the ServiceAccount RBAC forbidden-rules check |
| `WithCustomValidator(fn)` | —           | Register a custom `CustomValidator`; can be called multiple times to append |
| `SkipAppDataCheck()` | **on**      | Disable the built-in `.Values.userspace.appdata` cross-check. The check is on by default; turn it off only when a chart legitimately renders appdata via a non-standard path |
| `WithAppDataValidator()` | —           | Re-enable the built-in `.Values.userspace.appdata` cross-check. Kept as a back-compat alias — the check now runs by default and is disabled with `SkipAppDataCheck()`. Calling this on a fresh checker is a no-op |

> Note: `New()`'s default behavior is "run everything except `ServiceAccountRules`" — the Chart.yaml ↔ manifest version match check and the `.Values.userspace.appdata` cross-check both run by default. Use `Skip*` to turn off a built-in check, and `With*Check` to turn on one that is off by default.

### 6. Typical usage

**Run a full lint in one line**:

```go
import oac "github.com/beclab/Olares/framework/oac"

if err := oac.Lint("./my-app",
    oac.WithOwnerAdmin("root"),
    // same-version check is on by default; use SkipSameVersionCheck()
    // here if you don't want Chart.yaml ↔ manifest alignment enforced.
); err != nil {
    log.Fatal(err)
}
```

**Reuse a Checker across multiple charts**:

```go
c := oac.New(
    oac.WithOwner("alice"),
    oac.WithAdmin("root"),
    oac.WithServiceAccountRulesCheck(),
    // .Values.userspace.appdata cross-check runs by default; pass
    // oac.SkipAppDataCheck() here only if you need to bypass it.
)
for _, p := range paths {
    if err := c.Lint(p); err != nil {
        log.Printf("%s: %v", p, err)
    }
}
```

**Just the image list**:

```go
images, err := oac.ListImagesFromOAC("./my-app", oac.WithOwnerAdmin("root"))
```

**Validate manifest content only (no chart directory needed)**:

```go
if err := oac.ValidateManifestContent(yamlBytes); err != nil {
    var ve *oac.ValidationError
    if errors.As(err, &ve) {
        fmt.Printf("apiVersion=%s, field=%s, reason=%s\n", ve.Version, ve.Field, ve.Reason)
    }
}
```

**Pass both admin-install and user-install scenarios** (two equivalent forms):

```go
err := oac.LintBothOwnerScenarios("./my-app")
// or via the option directly:
err = oac.Lint("./my-app", oac.WithAutoOwnerScenarios())
```

**Typed read of the config only (no validation)**:

```go
cfg, err := oac.LoadAppConfiguration("./my-app")
if err != nil { log.Fatal(err) }
for _, r := range cfg.Spec.Resources {
    fmt.Println(r.Mode, r.RequiredCPU, r.LimitedCPU)
}
```

**Switch from an existing Manifest to a typed view**:

```go
m, _ := oac.New().LoadManifestFile("./my-app")
if cfg, ok := oac.AsAppConfiguration(m); ok {
    _ = cfg.Options.Upload.Dest
}
```

**Hand-built / programmatically generated cfg, then validate**:

```go
cfg := &oac.AppConfiguration{ /* ... */ } // equivalent to *manifest.AppConfiguration

// Top-level convenience for a one-shot validation
if err := oac.ValidateAppConfiguration(cfg); err != nil {
    var ve *oac.ValidationError
    if errors.As(err, &ve) { /* ... */ }
}

// Method form when reusing a Checker alongside other entry points (Lint / LoadAppConfiguration)
c := oac.New()
if err := c.ValidateAppConfiguration(cfg); err != nil { /* ... */ }
```

### 7. Lightweight helpers

These helpers cover lookups that don't need a full lint or helm render.

#### 7.1 `PeekManifestVersions`

Extracts `apiVersion` and `olaresManifest.version` from raw YAML using the same line-oriented regex probe as the version-dispatch pipeline (see §9 — tolerant of unrendered Helm template blocks, quoted / commented values, and CRLF). Useful when you only want to decide which pipeline or schema applies, without paying for parse + validate.

```go
// Versions is a trivial DTO; field names match the YAML keys (apiVersion /
// olaresManifest.version).
type ManifestVersions struct {
    APIVersion            string
    OlaresManifestVersion string
}

func PeekManifestVersions(content []byte) (ManifestVersions, error)
```

- Missing keys collapse to empty strings rather than raising an error.
- Only low-level scanner failures (e.g. an unreadable buffer) produce a non-nil `error`.
- Lines whose first byte is `' '`, `'\t'`, `'-'`, or `'#'` are ignored, so nested / commented re-definitions don't poison the result.

```go
v, err := oac.PeekManifestVersions(yaml)
if err != nil { /* ... */ }
switch {
case v.APIVersion == "v2":
    // v2-specific path
case v.OlaresManifestVersion == "":
    // treat as legacy
}
```

#### 7.2 `ResourceLimitsForResourceMode`

Resolves the CPU / memory / disk / GPU envelope that applies to one `spec.resources[]` row. Returns the inline `ResourceRequirement` declared on the matched mode, exposed as a pure function so that external tooling can compute "what should the limits be for this mode" without invoking helm.

```go
// Same eight-field shape as ResourceRequirement: cpu / memory / disk / gpu,
// each with a required and a limited entry. Fields not declared on the
// manifest collapse to "".
type ManifestResourceLimits = manifest.ResourceRequirementLimits

func ResourceLimitsForResourceMode(
    cfg *AppConfiguration,
    mode string, // e.g. "nvidia", "amd-apu" — matched case-insensitively
) (ManifestResourceLimits, error)
```

Errors:

- `cfg == nil` → `"oac: AppConfiguration is nil"`.
- No `spec.resources[]` entry whose `mode` matches → `"oac: no spec.resources entry with mode ..."`.

```go
lim, err := oac.ResourceLimitsForResourceMode(cfg, "nvidia")
if err != nil { /* ... */ }
fmt.Println(lim.RequiredCPU, lim.LimitedCPU, lim.RequiredGPU, lim.LimitedGPU)
```

---

### 8. Error model

- Validation errors are returned as `*oac.ValidationError` (unwrap with `errors.As` to get the fields)
- `AggregateErrors([]error)` merges multiple errors into one (nil / empty slice returns nil)
- All non-validation errors (IO, helm render, RBAC parsing, etc.) are returned as plain `error`
- For legacy manifests (`olaresManifest.version < 0.12.0`), validation runs parse+validate twice — once with `admin==owner` template rendering and once with `admin!=owner` — and aggregates failures into a single `ValidationError`

### 9. Version dispatch pipeline

Callers don't need to worry about this, but it's worth knowing:

- **Probe**: `manifest.Peek` uses regex to pull `olaresManifest.version` / `apiVersion` from the raw YAML (tolerates unrendered `{{ ... }}`)
- **Dispatch**: `manifest.NewPipeline(olaresVersion, defaultStrategy)`
  - `< 0.12.0` → `dualOwnerPipeline`: must render the chart through helm; validation runs parse+validate in both admin==owner and admin!=owner scenarios
  - `>= 0.12.0` / empty / malformed → `singlePipeline`: parses the YAML literally, no template rendering
- **Strategy instance**: the root package only wires a single `&manifest.ManifestStrategy{}` (stateless, goroutine-safe), responsible for parse + validate of v1 / v2 / v3 `apiVersion`
- `oac.ManifestFileName` exports the `"OlaresManifest.yaml"` constant

---

## OlaresManifest.yaml validation rules

Validation is organized in three layers:

1. **Structural rules** (`internal/manifest.ValidateAppConfiguration`, exposed by the root package as `oac.ValidateAppConfiguration(cfg, opts...)` and `(*Checker).ValidateAppConfiguration(cfg)`): field-level validation driven by `ozzo-validation`, including `spec` rules that depend on `olaresManifest.version` (legacy flat quantities required `< 0.12.0` vs modern `spec.resources[]` required `>= 0.12.0`) plus version-independent cross-field rules (Rule 7 mutual exclusion, mode → supportArch, completeness)
2. **Cross-field rules** (`checkSubCharts` only): v2 `spec.subCharts` layout rules that require looking at `apiVersion` and `spec`
3. **Resource-level rules** (the `resources` sub-package): after a helm dry-run, additional checks on the generated Kubernetes resources

> `Lint` also runs a folder-layout check (`chartfolder`) up front — see [§3.5](#35-folder-layout-check-checklayout-on-by-default).

---

### 1. Structural rules (field-level)

#### 1.1 Top-level `AppConfiguration`

| Field | Rule |
|---|---|
| `olaresManifest.version` | **required** |
| `apiVersion` | Must be `"v1"`, `"v2"`, or `"v3"` when non-empty; empty defaults to v1 |
| `metadata` | Recursive validation (see §1.2) |
| `entrances` | **required**, length `1..10`, `name` must be unique (`uniqueEntranceNames`) |
| `spec` | **required**, recursive validation (see §1.3) |
| `permission` | Recursive (no additional rules at the permission level today) |
| `options` | Recursive validation (see §1.4) |

#### 1.2 `metadata` (`AppMetaData`)

| Field | Rule |
|---|---|
| `name` | **required**, 1–30 characters |
| `icon` | **required** |
| `description` | **required** |
| `title` | **required**, 1–30 characters |
| `version` | **required**, must be a valid [SemVer](https://semver.org) (e.g. `1.2.3`) |

#### 1.3 `spec` (`AppSpec`)

Validation pivots on **`olaresManifest.version`** (`ConfigVersion`):

| `olaresManifest.version` | Flat `spec.required*` / `spec.limited*` (cpu/memory/disk) | `spec.resources[]` |
|---|---|---|
| **&lt; 0.12.0** | `requiredCpu` / `limitedCpu` / `requiredMemory` / `limitedMemory` / `requiredDisk` are **required** and quantity-checked; `limitedDisk` is optional and quantity-checked | Not enforced as required. Per-element rules in §2.2 still run on whatever entries are present |
| **≥ 0.12.0** | Not enforced at the field level — these fields must instead be empty by virtue of the mutual-exclusion rule below (Rule 7) when `spec.resources[]` is set | `spec.resources` is **required** (must be non-empty). Each entry follows §2.2 |

`spec.requiredGpu` and `spec.limitedGpu` are **always optional** and quantity-checked when set, regardless of version.

In addition, regardless of version, **Rule 7** (mutual exclusion) applies: when `spec.resources[]` is non-empty, none of the eight legacy flat fields (`spec.requiredCpu` / `spec.limitedCpu` / `spec.requiredMemory` / `spec.limitedMemory` / `spec.requiredDisk` / `spec.limitedDisk` / `spec.requiredGpu` / `spec.limitedGpu`) may be set. The two are alternative shapes of the same envelope — pick one.

| Field | Rule |
|---|---|
| `requiredMemory` / `requiredDisk` / `requiredCpu` / `limitedMemory` / `limitedCpu` | Required + quantity regex when `< 0.12.0`; must be empty when `>= 0.12.0` and `spec.resources[]` is set (Rule 7) |
| `limitedDisk` | Optional + quantity regex when `< 0.12.0`; must be empty when `>= 0.12.0` and `spec.resources[]` is set (Rule 7) |
| `requiredGpu` / `limitedGpu` | Always optional + quantity regex when non-empty; must be empty when `spec.resources[]` is set (Rule 7) |
| `resources[]` | Required when `>= 0.12.0`. Recursive: each `ResourceMode` follows the rules in §2.2 |

The Kubernetes Quantity regex covers common forms: `100m`, `1.5`, `512Mi`, `2Gi`, `1e9`, etc.

**Empty-spec optimisation**: instead of cascading every per-field "is required" rule when the user supplies nothing, `validateAppSpec` collapses the missing-everything case into a single, version-tagged guidance message:

- **`< 0.12.0`** (legacy): when all five required legacy fields (`requiredCpu`, `limitedCpu`, `requiredMemory`, `limitedMemory`, `requiredDisk`) are empty, the per-field cascade is suppressed and the validator emits one consolidated message: `spec.requiredCpu / spec.limitedCpu / spec.requiredMemory / spec.limitedMemory / spec.requiredDisk are required for olaresManifest.version < 0.12.0; populate the legacy resource envelope`. Partial fills (one or more legacy fields set) still produce pinpointed per-field "is required" errors.
- **`≥ 0.12.0`** (modern): when `spec.resources` is empty, the validator emits one message: `spec.resources is required for olaresManifest.version >= 0.12.0; declare at least one entry`.

The cross-field rules in §2.2 still run as appropriate for the version: Rule 7 (mutual exclusion) always fires when `spec.resources[]` is non-empty, while Rule 1 (mode → supportArch) and Rule 4-empty (entry completeness) are scoped to `>= 0.12.0`. The optimisation only suppresses the noisy per-field "Required" cascade when there is nothing to attribute it to.

#### 1.4 Each `entrances[i]` (`Entrance`)

| Field | Rule |
|---|---|
| `name` | Matches `^[a-z0-9A-Z-]*$`, length ≤ 63 |
| `host` | Matches `^[a-z]([-a-z0-9]*[a-z0-9])$`, length ≤ 63 |
| `port` | Must be > 0 (`ozzo`'s `Min` skips unset zero values, so `0` still passes, but negatives like `-1` are rejected) |
| `icon` | When non-empty, must be a valid `http://` / `https://` URL |
| `title` | **required**, 1–30 characters, matches `^[a-z0-9A-Z-\s]*$` |
| `authLevel` | One of `""`, `"public"`, `"private"` |
| `openMethod` | One of `""`, `"default"`, `"iframe"`, `"window"` |

Additionally, all entrances' `name` fields must be unique (enforced by the top-level `uniqueEntranceNames` rule).

#### 1.5 `options`

| Sub-field | Rule |
|---|---|
| `policies[]` | Recursive: each `Policy` requires `uriRegex` and `level`; `validDuration` must match `^((?:[-+]?\d+(?:\.\d+)?([smhdwy]\|us\|ns\|ms))+)$` when non-empty |
| `resetCookie` | Placeholder (no rules) |
| `dependencies[]` | Each `Dependency` requires `name` and `version`; `type` must be `"system"` or `"application"` |
| `appScope` | Placeholder |
| `websocket` | Placeholder |

---

### 2. Cross-field rules

#### 2.1 `checkSubCharts` (unconditionally triggered by `apiVersion=v2`)

- `spec.subCharts` must exist and be non-empty
- At least one entry in `spec.subCharts[]` must have `shared: true`

> This rule **only looks at `apiVersion`** and is independent of `olaresManifest.version`. Even if an older chart is on `olaresManifest.version: 0.11.0`, as long as it declares `apiVersion: v2`, `checkSubCharts` still fires — don't confuse it with the `olaresManifest.version` threshold in §1.3 for flat fields vs `spec.resources[]`.

#### 2.2 `spec.resources[]` per-element and modern-only cross-field rules (see §1.3)

Each `spec.resources[i]` is a `ResourceMode`:

```yaml
- mode: nvidia
  # inline ResourceRequirement (required/limited for cpu/memory/disk/gpu)
  requiredCpu: 100m
  limitedCpu: 200m
  requiredMemory: 128Mi
  limitedMemory: 256Mi
  requiredDisk: 1Gi
  limitedDisk: 2Gi
  requiredGpu: 8Gi
  limitedGpu: 8Gi
```

**Valid `mode` values**: `cpu`, `amd-apu`, `amd-gpu`, `apple-m`, `nvidia`, `nvidia-gb10`, `mthreads-m1000`.

**Per-element base rules** (`ValidateResourceMode`, dispatched onto each `ResourceMode` via `validation.Each`):

- `mode` is **required** and must be within the enum above
- Every quantity field (`requiredCpu`…`limitedGpu`) in the inline `ResourceRequirement` must match the k8s Quantity regex when non-empty

**Cross-field rules** (numbers match the code comments):

| # | Name | Description |
|---|---|---|
| Rule 1 | mode → supportArch | GPU chip families imply a CPU architecture:<br>• `amd-gpu`, `nvidia` → must include `amd64`<br>• `nvidia-gb10`, `mthreads-m1000` → must include `arm64` |
| Rule 3 | Non-GPU modes must not declare GPU | `cpu` / `amd-apu` / `apple-m` / `nvidia-gb10` / `mthreads-m1000` must leave `requiredGpu` / `limitedGpu` empty (via `ensureNoGPUSection`). **`requiredDisk` / `limitedDisk` are allowed on all modes.** |
| Rule 4 | Section-completeness envelope (`ensureSectionComplete`) | If the inline `ResourceRequirement` has **at least one non-empty quantity field** (`hasAnyQuantity`), it must be filled out completely:<br>• `requiredCpu`, `limitedCpu`, `requiredMemory`, `limitedMemory`, `requiredDisk`, `limitedDisk` — all six **required**<br>• If `mode` ∈ `{nvidia, amd-gpu}`, `requiredGpu` and `limitedGpu` are additionally required as a **complete pair** |
| Rule 4-empty | Empty entry rejected (`requireResourceEntryFields`) | If the inline `ResourceRequirement` has **no quantity fields at all**, every standard field is reported as missing so that an empty `mode`-only declaration cannot slip through unnoticed. GPU pairs are reported alongside cpu/memory/disk on `nvidia` / `amd-gpu` |
| Rule 5 | `limited >= required` | For each dimension (cpu/memory/disk/gpu), when both required and limited are declared, limited must be ≥ required |
| Rule 7 | Mutual exclusion of legacy flat fields (`ensureLegacyAndResourcesAreMutuallyExclusive`) | When `spec.resources[]` is non-empty, the eight legacy flat fields `spec.requiredCpu` / `spec.limitedCpu` / `spec.requiredMemory` / `spec.limitedMemory` / `spec.requiredDisk` / `spec.limitedDisk` / `spec.requiredGpu` / `spec.limitedGpu` **must all be empty**. The two are alternative expressions of the same envelope and cannot coexist on a single manifest. Each violating field reports its own error and `errors.Join` aggregates them. Applies regardless of `olaresManifest.version`. |

**Versioning notes**: `specResourceCrossFieldRules` is partially version-gated.

- Rule 7 (mutual exclusion) runs at every `olaresManifest.version`. A legacy manifest that mixes `spec.required*` / `spec.limited*` with `spec.resources[]` is still rejected so users on the legacy schema cannot accidentally straddle both shapes.
- Rule 1 (mode → supportArch) and Rule 4-empty (empty-entry completeness) only run when `olaresManifest.version >= 0.12.0`. `spec.resources[]` is not part of the legacy schema, so its inner contents are intentionally not inspected below the gate; legacy users are instead steered toward the flat fields by §1.3 (consolidated guidance when the legacy envelope is missing).
- The version dispatch for "which shape is required" lives in `validateAppSpec`, not in `specResourceCrossFieldRules`. Per-element base rules (the `mode` enum, quantity validity, Rules 3 / 4 / 5) are fired by ozzo-validation via `validation.Each` on `ValidateResourceMode`; that wiring is also only added in the modern branch of `validateAppSpec`, so they never fire on legacy manifests either.

---

### 3. Resource-level rules (require a helm dry-run)

`Checker.Lint` first calls `helmrender.BuildValues` and `helmrender.Render` to produce the **default** `kube.ResourceList`. §3.2, §3.3 and optionally §3.4 run on that same list; when `SkipResourceCheck()` is not set, `checkResourceLimits` (§3.1) is invoked afterwards (modern manifests re-render **additionally** per mode, independent of the default list).

`Checker.CheckResources` also does one `BuildValues` + `Render` (except it returns immediately for **`apiVersion: v2`**, which skips container limit checks), then forwards the chart path and manifest to `checkResourceLimits`; for modern manifests the list produced by that render is not used by the limit branch (limit checks are entirely driven by per-mode re-renders), so it **does not include** §3.2 / §3.3.

For **`apiVersion: v2`** the parent OAC root is not a renderable workload chart in the multi-chart install layout, so the helm dry-run for **`Lint` / `CheckServiceAccountRules` / `ListImages`** iterates `spec.subCharts[]` and concatenates the per-subchart `kube.ResourceList`s. **`apiVersion: v3`** follows the same single-chart render path as **v1** (whole chart at the OAC root). A non-empty `apiVersion` outside `v1` / `v2` / `v3` fails validation and resource checks with **不支持该版本**.

> `Lint` always runs the helm render. Upload mount and workload naming (§3.2, §3.3) are **always** enforced; container limits (§3.1) are gated by `SkipResourceCheck()`; RBAC (§3.4) is gated by `WithServiceAccountRulesCheck()`.

#### 3.1 Container limits (`CheckResourceLimits`, skippable)

For each `Deployment` and `StatefulSet` in the render, on the primary container:

- Manifest side: `requiredCpu <= limitedCpu`, `requiredMemory <= limitedMemory`
- Container side: every container **must** declare both `requests` and `limits` for CPU and memory
- Container-side `requests.cpu/memory <= limits.cpu/memory`
- Sum of all containers' `requests.cpu/memory` ≤ manifest-side `requiredCpu/Memory`
- Sum of all containers' `limits.cpu/memory` ≤ manifest-side `limitedCpu/Memory`

**Where the manifest-side limits come from (dispatched by `olaresManifest.version` and `apiVersion`)**:

- **Legacy (`< 0.12.0`)**: read directly from the four flat fields `spec.requiredCpu` / `spec.limitedCpu` / `spec.requiredMemory` / `spec.limitedMemory`. `Lint` renders the chart once (sharing the render with §3.2 / §3.3) and compares that single `kube.ResourceList` against the four fields.
- **Modern (`>= 0.12.0`) with `apiVersion` unset, `v1`, or `v3`**: `validateAppSpec` requires `spec.resources[]` and Rule 7 forbids mixing it with the flat fields, so limits come from `spec.resources[]`. For each `ResourceMode rm`:
  1. `BuildValues(...)` + `SetGPUType(values, rm.Mode)` (`.Values.GPU.Type = <mode>`).
  2. `helmrender.Render(...)` at `oacPath` produces the `kube.ResourceList` for that mode.
  3. Limits use the **inline** `ResourceRequirement` on that mode (`requiredCpu`, `limitedCpu`, `requiredMemory`, `limitedMemory`).
  4. `CheckResourceLimits(list, limits)`. Failure prefix: `resources mode=<mode>:`.
- **Modern (`>= 0.12.0`) with `apiVersion` v2**: the container limit check is **skipped** (`checkResourceLimits` returns nil when `spec.resources` is non-empty; `CheckResources` also returns nil without rendering for limits).
- Failures across modes are aggregated via `errors.Join`.
- If `spec.resources` is empty (a modern manifest without any `ResourceMode`), this check is **skipped entirely** — such a chart has no limits to compare against, so there is nothing more to validate.
- The upload-mount and workload-naming checks (§3.2 / §3.3) run **only on the default render**; per-mode re-renders serve **only** §3.1 and do not re-run §3.2 / §3.3.

#### 3.2 Upload mount (`CheckUploadConfig`, always enforced)

If `options.upload.dest` is set in the manifest: the primary container of any rendered Deployment/StatefulSet must mount the same path in its `volumeMounts` (compared after `filepath.Clean`). Missing mounts are reported as an error.

#### 3.3 Workload naming (`CheckDeploymentName`, always enforced)

When `olaresManifest.type == "app"`: at least one Deployment or StatefulSet's **rendered** name must equal `metadata.name`. Non-app types skip this check.

During dry-run, helm's `Release.Name` is set to `metadata.name` (matching the "release name == app name" convention in production Olares), so templates using `name: {{ .Release.Name }}` pass this check as well.

#### 3.4 ServiceAccount RBAC (`CheckServiceAccountRules`, off by default; enable with `WithServiceAccountRulesCheck()`)

For every `RoleBinding` / `ClusterRoleBinding` rendered by the chart:

- Find bindings where `subject.kind == ServiceAccount` and collect their `roleRef.name`s
- Pull the corresponding `Role` / `ClusterRole` `rules`
- Compare them against the default forbidden set (`DefaultForbiddenRules`, overridable internally via `LoadForbiddenRules("custom yaml")`):

```yaml
rules:
- apiGroups: ['*']
  resources:  [nodes, networkpolicies]
  verbs:      [create, update, patch, delete, deletecollection]
```

If any ServiceAccount binding grants any of those actions, the check fails.

#### 3.5 Folder-layout check (`CheckLayout`, on by default)

The `Lint` entry point runs `chartfolder.CheckLayout(path)` first:

- Directory name matches `^[a-z0-9]{1,30}$`
- Directory exists
- Contains `Chart.yaml`, `values.yaml`, `templates/`, `OlaresManifest.yaml`

The further `CheckConsistency` (called by `CheckSameVersion`) also validates:

- directory name == `Chart.yaml`'s `name` == `metadata.name`
- `metadata.version` == `Chart.yaml`'s `version`

`CheckWithTitle` (used in the PR flow, not exposed at the root package) additionally requires:

- Every entry in `metadata.categories` is within the fixed enum (`AI` / `Blockchain` / `Utilities` / `Social Network` / `Data` / `Entertainment` / `Productivity` / `Lifestyle` / `Developer` / `Multimedia`)
- The directory name is not in the reserved list (`user` / `system` / `space` / `default` / `os` / `kubesphere` / `kube` / `kubekey` / `kubernetes` / `gpu` / `tapr` / `bfl` / `bytetrade` / `project` / `pod`)

---

### 4. Custom validators

Functions registered via `WithCustomValidator(fn)` are invoked after the built-in structural validation and before the resource-level checks:

```go
type CustomValidator func(oacPath string, m Manifest) error
```

A built-in equivalent runs as a separate, always-on step (not via `customValidators`): if any file under the chart's `templates/*.yaml` references `.Values.userspace.appdata`, `OlaresManifest.yaml` must declare `permission.appData: true`, otherwise the check fails. It runs by default in `Lint`; pass `SkipAppDataCheck()` to disable it for a particular call. The legacy entry point `WithAppDataValidator()` is kept as a back-compat alias that simply re-asserts the on-by-default state.

---

### 5. Version compatibility matrix

| `olaresManifest.version` | `apiVersion` | Pipeline | Extra rules |
|---|---|---|---|
| Any (including empty / malformed) | `v1` / empty | Parsed as v1 directly, or after template rendering | — |
| Any | `v2` | Same as above, still against the v1 schema | `checkSubCharts` active |
| `< 0.12.0` | Any | dualOwnerPipeline (helm template render in both owner scenarios) | — |
| `>= 0.12.0` | Any | singlePipeline (literal parse) | — |
| `< 0.12.0` | Any | — | §1.3 `spec` validation: legacy flat fields (`requiredCpu` / `limitedCpu` / `requiredMemory` / `limitedMemory` / `requiredDisk`) **required** + quantity-checked; `limitedDisk` optional; `requiredGpu` / `limitedGpu` optional |
| `>= 0.12.0` | Any | — | §1.3 `spec` validation: `spec.resources[]` **required** (each entry follows §2.2 base rules); `requiredGpu` / `limitedGpu` still optional at the spec level |
| Any | Any | — | §1.3 `spec` validation (cross-field, version-independent): Rule 7 mutual exclusion (no legacy flat fields when `spec.resources[]` is set) |
| `>= 0.12.0` | Any | — | §1.3 `spec` validation (cross-field, modern-only): Rule 1 mode → supportArch and Rule 4-empty completeness on every declared `spec.resources[]` entry. Skipped on `< 0.12.0` because `spec.resources[]` is not part of the legacy schema |

These four axes are orthogonal: v2 always triggers `checkSubCharts`; `>=0.12.0` makes `spec.resources[]` mandatory at the field level; v1 + `<0.12.0` does not trigger `checkSubCharts` and requires the legacy flat fields. Inside `specResourceCrossFieldRules`, Rule 7 (mutual exclusion) runs regardless of version and only fires when `spec.resources[]` is non-empty; Rule 1 and Rule 4-empty are gated to `>=0.12.0` so the `spec.resources[]` payload is left untouched on legacy manifests.
