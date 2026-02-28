---
outline: [2, 3]
description: Declare and validate app configuration via envs in `OlaresManifest.yaml`, and reference values in templates through `.Values.olaresEnv`.
---
# Custom environment variables

Developers can declare the configuration parameters required by an app under `envs` in `OlaresManifest.yaml`. During deployment or upgrade, App Service injects the final resolved values into `.Values.olaresEnv` in the app's `values.yaml`.

:::tip Injection notes
- All variables declared in `envs` use the same template path: <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>
- If `values.yaml` contains a field with the same name, it will be overwritten by the system-injected value.
:::

## How to use

### Declare in the manifest

**Example**:

```yaml
olaresManifest.version: '0.10.0'
olaresManifest.type: app

envs:
  - envName: ADMIN_PASSWORD
    type: password
    required: true
    editable: false
    description: "Password must be at least 6 characters long"
    regex: '^[\w\-!@#$%^&*()+={}\[\]:,.?~]{6,}$'
```

### Reference in templates

**Example**:

```yaml
env:
  - name: ADMIN_PASSWORD
    value: "{{ .Values.olaresEnv.ADMIN_PASSWORD }}"
```

## Field Reference

The following fields define where a variable's value comes from, how it is validated, and whether it can be edited.

### applyOnChange

Boolean. When `true`, changing this variable automatically restarts all apps/components that use it so the change takes effect.

:::info How changes take effect
When `applyOnChange` is `false`, the change will not take effect even if you manually stop/start the app. It only takes effect after the app is upgraded or reinstalled.
:::

### default

The default value of the variable. Developers can provide it when authoring the app, and users cannot modify it. When the value is not provided by user input or `valueFrom`, `default` is used.

### description

Describes the purpose of the variable and the meaning of valid values.

### editable

Boolean. When `true`, the variable can be modified after installation.

### envName

The key injected into `values.yaml`. Template reference: <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>.

### options

An allowlist of values. The variable's value must be chosen from the list, and the system usually provides a selection UI for users.

**Example**:
```yaml
envs:
  - envName: VERSION
    options:
      - title: "Windows 11 Pro"
        value: "iso/Win11_24H2_English_x64.iso"
      - title: "Windows 7 Ultimate"
        value: "iso/win7_sp1_x64_1.iso"
```

### regex

Regex validation. Only values matching `regex` are allowed. Validation failure may cause setting to fail, or cause installation/upgrade to fail.

### remoteOptions

Provides the options list via a URL. The response body must be a JSON-encoded options array.

**Example**:

```yaml
envs:
  - envName: VERSION
    remoteOptions: https://app.cdn.olares.com/appstore/windows/version_options.json
```

### required

Boolean. When `true`, the variable is required for installation:
- If `default` is not set, the system will prompt the user to enter a value before installation.
- After installation, the value cannot be changed to empty.

### type

The value type used for validation. If validation fails, the value cannot be set. Supported types:

- `int`
- `bool`
- `url`
- `ip`
- `domain`
- `email`
- `string`
- `password`

### value

The value of an environment variable. It does not support directly hardcoding arbitrary constants. The value can come from `default`, user input, or by referencing other variables via `valueFrom`.

### valueFrom

References the value of an Olares system environment variable or user environment variable. 

**Example**:
```yaml
envs:
  - envName: APP_CDN_ENDPOINT
    required: true
    applyOnChange: true
    valueFrom:
      envName: OLARES_SYSTEM_CDN_SERVICE
```

:::info Inheritance rules for references
When `valueFrom` is used, the current variable inherits all attribute fields from the referenced variable (such as `type`, `editable`, `regex`, etc.). Any fields with the same names defined on the current variable will be ignored. In addition, `default` and `options` are not effective in this case.
:::