---
outline: [2, 4]
description: Declare and validate app configuration via envs in `OlaresManifest.yaml`, and reference values in templates through `.Values.olaresEnv`.
---
# Declarative environment variables

Use `envs` in `OlaresManifest.yaml` to declare the configuration parameters, such as passwords, API endpoints, or feature flags. During deployment, app-service resolves the values and injects them into `.Values.olaresEnv` in `values.yaml`. Reference them in Helm templates as <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>.

## Variable sources

Declarative variables can obtain values from configurations managed outside the application:

- **System variables**: Environment variables defined at the Olares cluster level. They are set during system installation or centrally managed by administrators, and are shared by all users within the cluster.
- **User variables**: Environment variables defined at the Olares user level. They are managed individually by each user, and are isolated from one another within the same cluster.

Applications cannot modify these variables directly. To use them, map the variable via the `valueFrom` field.

## Map environment variables

Both system environment variables and user environment variables use the same mapping mechanism via `valueFrom`.

The following example maps the system variable `OLARES_SYSTEM_CDN_SERVICE` to an application variable `APP_CDN_ENDPOINT`:

1. In `OlaresManifest.yaml`, declare an app variable under `envs` and set `valueFrom.envName` to the system variable name.

    ```yaml
    # Map system variable OLARES_SYSTEM_CDN_SERVICE to app variable APP_CDN_ENDPOINT
    olaresManifest.version: '0.10.0'
    olaresManifest.type: app

    envs:
      - envName: APP_CDN_ENDPOINT
        required: true
        applyOnChange: true
        valueFrom:
          envName: OLARES_SYSTEM_CDN_SERVICE
    ```

2. In your Helm template, reference the app variable via `.Values.olaresEnv.<envName>`.

    ```yaml
    # Use APP_CDN_ENDPOINT in a container environment variable
    env:
      - name: CDN_ENDPOINT
        value: "{{ .Values.olaresEnv.APP_CDN_ENDPOINT }}"
    ```

At deployment, app-service resolves the referenced variable and injects the value into `values.yaml`:

```yaml
# Injected by app-service into values.yaml at deployment
olaresEnv:
  APP_CDN_ENDPOINT: "https://cdn.olares.com"
```

For the full list of available environment variables, see [Variable references](#variable-references).

## Declaration fields

The following fields are available under each `envs` entry.

### envName

The name of the variable as injected into `values.yaml`. Reference it in templates as <code v-pre>{{ .Values.olaresEnv.&lt;envName&gt; }}</code>.

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

**Example**: a dropdown list of supported Windows versions for installation.

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

## Variable references

### System environment variables

The following table lists system-level environment variables that can be referenced via `valueFrom`.

| Variable | Type | Default | Editable | Required | Description |
| --- | --- | --- | --- | --- | --- |
| `OLARES_SYSTEM_REMOTE_SERVICE` | `url` | `https://api.olares.com` | Yes | Yes | Remote service endpoint for Olares, such as Market and Olares Space. |
| `OLARES_SYSTEM_CDN_SERVICE` | `url` | `https://cdn.olares.com` | Yes | Yes | CDN endpoint for system resources. |
| `OLARES_SYSTEM_DOCKERHUB_SERVICE` | `url` | None | Yes | No | Docker Hub mirror or accelerator endpoint. |
| `OLARES_SYSTEM_ROOT_PATH` | `string` | `/olares` | No | Yes | Olares root directory path. |
| `OLARES_SYSTEM_ROOTFS_TYPE` | `string` | `fs` |  No | Yes | Olares filesystem type. |
| `OLARES_SYSTEM_CUDA_VERSION` | `string` | None | No | No | Host CUDA version. |

### User environment variables

All user environment variables are editable by the user.

#### User information

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_EMAIL` | `string` | None | User email address. |
| `OLARES_USER_USERNAME` | `string` | None | Username. |
| `OLARES_USER_PASSWORD` | `password` | None | User password. |
| `OLARES_USER_TIMEZONE` | `string` | None | User timezone. For example, `Asia/Shanghai`. |

#### SMTP settings

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_SMTP_ENABLED` | `bool` | None | Whether to enable SMTP. |
| `OLARES_USER_SMTP_SERVER` | `domain` | None | SMTP server domain. |
| `OLARES_USER_SMTP_PORT` | `int` | None | SMTP server port. Typically `465` or `587`. |
| `OLARES_USER_SMTP_USERNAME` | `string` | None | SMTP username. |
| `OLARES_USER_SMTP_PASSWORD` | `password` | None | SMTP password or authorization code. |
| `OLARES_USER_SMTP_FROM_ADDRESS` | `email` | None | Sender email address. |
| `OLARES_USER_SMTP_SECURE` | `bool` | `"true"` | Whether to use a secure protocol. |
| `OLARES_USER_SMTP_USE_TLS` | `bool` | None | Whether to use TLS. |
| `OLARES_USER_SMTP_USE_SSL` | `bool` | None | Whether to use SSL. |
| `OLARES_USER_SMTP_SECURITY_PROTOCOLS` | `string` | None | Security protocol. Allowed values: `tls`, `ssl`, `starttls`, `none`. |

#### Mirror and proxy endpoints

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_HUGGINGFACE_SERVICE` | `url` | `https://huggingface.co/` | Hugging Face service URL. |
| `OLARES_USER_HUGGINGFACE_TOKEN` | `string` | None | Hugging Face access token. |
| `OLARES_USER_PYPI_SERVICE` | `url` | `https://pypi.org/simple/` | PyPI mirror URL. |
| `OLARES_USER_GITHUB_SERVICE` | `url` | `https://github.com/` | GitHub mirror URL. |
| `OLARES_USER_GITHUB_TOKEN` | `string` | None | GitHub personal access token. |

#### API keys

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_OPENAI_APIKEY` | `password` | None | OpenAI API key. |
| `OLARES_USER_CUSTOM_OPENAI_SERVICE` | `url` | None | Custom OpenAI-compatible service URL. |
| `OLARES_USER_CUSTOM_OPENAI_APIKEY` | `password` | None | API key for the custom OpenAI-compatible service. |
