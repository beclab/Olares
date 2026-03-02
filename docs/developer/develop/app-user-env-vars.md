---
outline: [2, 3]
description: User environment variables are user-level personalized settings. Apps must reference them via `envs.valueFrom` and map them to `.Values.olaresEnv`.
---

# User environment variables

User environment variables are per-user settings managed by the user themselves. Common examples include timezone, email address, SMTP configuration, mirror endpoints, and API keys. In a cluster with multiple users, each user's values are independent.

Apps cannot read these variables directly. To use one, map it to an app variable in `envs` using `valueFrom`, then reference it in templates via `.Values.olaresEnv.<envName>`.

## Map a user variable to your app

1. In `OlaresManifest.yaml`, declare an app variable under `envs` and set `valueFrom.envName` to the user variable name.

    ```yaml
    # Map user variable OLARES_USER_TIMEZONE to app variable USER_TIMEZONE
    olaresManifest.version: "0.10.0"
    olaresManifest.type: app

    envs:
      - envName: USER_TIMEZONE
        valueFrom:
          envName: OLARES_USER_TIMEZONE
    ```

2. In your Helm template, reference the app variable via `.Values.olaresEnv.<envName>`.

    ```yaml
    # Use USER_TIMEZONE in a container environment variable
    env:
      - name: TZ
        value: "{{ .Values.olaresEnv.USER_TIMEZONE }}"
    ```

At deployment, app-service retrieves the current user's value and injects it into `values.yaml`:

```yaml
# Injected by app-service into values.yaml at deployment
olaresEnv:
  USER_TIMEZONE: "Asia/Shanghai"
```

For the full list of available user variables, see [Variable reference](#variable-reference).

## Variable reference

All user environment variables are editable by the user.

### User information

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_EMAIL` | `string` | None | User email address. |
| `OLARES_USER_USERNAME` | `string` | None | Username. |
| `OLARES_USER_PASSWORD` | `password` | None | User password. |
| `OLARES_USER_TIMEZONE` | `string` | None | User timezone. For example, `Asia/Shanghai`. |

### SMTP settings

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_SMTP_ENABLED` | `bool` | None | Whether to enable SMTP. |
| `OLARES_USER_SMTP_SERVER` | `domain` | None | SMTP server domain. |
| `OLARES_USER_SMTP_PORT` | `int` | None | SMTP server port. Typically `465` or `587`. |
| `OLARES_USER_SMTP_USERNAME` | `string` | None | SMTP username. |
| `OLARES_USER_SMTP_PASSWORD` | `password` | None | SMTP password or authorization code. |
| `OLARES_USER_SMTP_FROM_ADDRESS` | `email` | None | Sender email address. |
| `OLARES_USER_SMTP_SECURE` | `bool` | `"true"` | Whether to use a secure protocol. |
| `OLARES_USER_SMTP_USE_TLS` | `bool` | None | Use TLS. |
| `OLARES_USER_SMTP_USE_SSL` | `bool` | None | Use SSL. |
| `OLARES_USER_SMTP_SECURITY_PROTOCOLS` | `string` | None | Security protocol. Allowed values: `tls`, `ssl`, `starttls`, `none`. |

### Mirror and proxy endpoints

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_HUGGINGFACE_SERVICE` | `url` | `https://huggingface.co/` | Hugging Face service URL. |
| `OLARES_USER_HUGGINGFACE_TOKEN` | `string` | None | Hugging Face access token. |
| `OLARES_USER_PYPI_SERVICE` | `url` | `https://pypi.org/simple/` | PyPI mirror URL. |
| `OLARES_USER_GITHUB_SERVICE` | `url` | `https://github.com/` | GitHub mirror URL. |
| `OLARES_USER_GITHUB_TOKEN` | `string` | None | GitHub personal access token. |

### API keys

| Variable | Type | Default | Description |
| --- | --- | --- | --- |
| `OLARES_USER_OPENAI_APIKEY` | `password` | None | OpenAI API key. |
| `OLARES_USER_CUSTOM_OPENAI_SERVICE` | `url` | None | Custom OpenAI-compatible service URL. |
| `OLARES_USER_CUSTOM_OPENAI_APIKEY` | `password` | None | API key for the custom OpenAI-compatible service. |
