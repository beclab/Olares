---
outline: [2, 3]
description: Declare and validate app configuration via envs in `OlaresManifest.yaml`, and reference values in templates through `.Values.olaresEnv`.
---
# Declarative environment variables

Use `envs` in `OlaresManifest.yaml` to declare the configuration parameters your app needs, such as passwords, API endpoints, or feature flags. During deployment, app-service resolves the values and injects them into `.Values.olaresEnv` in `values.yaml`. Reference them in templates as <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>.

## Field reference

The following fields are available under each `envs` entry.

### envName

The name of the variable as injected into `values.yaml`. Reference it in templates as <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>.

### value

The resolved value of the variable. You cannot set a fixed constant directly. The value comes from `default`, user input, or a referenced variable via `valueFrom`.

### default

The default value for the variable. Provided by the developer at authoring time. Users cannot modify it. Used when no value is supplied by the user or by `valueFrom`.

### valueFrom

Maps this variable to a system or user environment variable. When set, the current variable inherits all fields from the referenced variable (`type`, `editable`, `regex`, and so on). Any fields defined locally on the current variable are ignored. `default` and `options` have no effect when `valueFrom` is used.

**Example**: map the app variable `APP_CDN_ENDPOINT` to the system variable `OLARES_SYSTEM_CDN_SERVICE`.

```yaml
# Map app env APP_CDN_ENDPOINT to system variable OLARES_SYSTEM_CDN_SERVICE
envs:
  - envName: APP_CDN_ENDPOINT
    required: true
    applyOnChange: true
    valueFrom:
      envName: OLARES_SYSTEM_CDN_SERVICE
```

### required

Boolean. When `true`, the variable must have a value for installation to proceed. If no `default` is set, the user is prompted to enter one. After installation, the value cannot be set to empty.

### editable

Boolean. When `true`, the variable can be modified after installation.

### applyOnChange

Boolean. When `true`, changing this variable automatically restarts all apps or components that use it. When `false`, a change only takes effect after the app is upgraded or reinstalled. Stopping and starting the app manually has no effect.

### type

The expected type of the value. Used for validation before the value is accepted. Supported types: `int`, `bool`, `url`, `ip`, `domain`, `email`, `string`, `password`.

### regex

A regular expression the value must match. If validation fails, the value cannot be set and installation or upgrade may fail.

### options

Restricts the variable to a fixed list of allowed values. The system presents users with a selection UI.

**Example**: a dropdown where the display title and the stored value differ.

```yaml
# Dropdown: title shown in UI, value stored internally
envs:
  - envName: VERSION
    options:
      - title: "Windows 11 Pro"
        value: "iso/Win11_24H2_English_x64.iso"
      - title: "Windows 7 Ultimate"
        value: "iso/win7_sp1_x64_1.iso"
```

### remoteOptions

Loads the options list from a URL instead of defining it inline. The response body must be a JSON-encoded array in the same format as `options`.

**Example**: options fetched from a remote endpoint.

```yaml
# Options list fetched from remote URL at install time
envs:
  - envName: VERSION
    remoteOptions: https://app.cdn.olares.com/appstore/windows/version_options.json
```

### description

A human-readable description of the variable's purpose and valid values. Displayed in the Olares interface.
