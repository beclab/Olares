---
outline: [2, 3]
description: System environment variables are cluster-wide global settings maintained by the administrator. Apps must reference them via `envs.valueFrom` and map them to `.Values.olaresEnv`.
---

# System environment variables

System environment variables are cluster-wide settings configured at Olares install time or managed by the cluster administrator. All users in the cluster share the same values. Common examples include the CDN endpoint, Docker Hub mirror, and Olares root path.

Apps cannot modify these variables. To use one, map it to an app variable in `envs` using `valueFrom`, then reference it in templates via `.Values.olaresEnv.<envName>`.

## Map a system variable to your app

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

At deployment, app-service resolves the system variable and injects the value into `values.yaml`:

```yaml
# Injected by app-service into values.yaml at deployment
olaresEnv:
  APP_CDN_ENDPOINT: "https://cdn.olares.com"
```

For the full list of available system variables, see [Variable reference](#variable-reference).

## Variable reference

The `editable` and `required` columns describe the system variable's own properties, not something your app controls.

| Variable | Type | Default | editable | required | Description |
| --- | --- | --- | --- | --- | --- |
| `OLARES_SYSTEM_REMOTE_SERVICE` | `url` | `https://api.olares.com` | `true` | `true` | Remote service endpoint for Olares, such as Market and Olares Space. |
| `OLARES_SYSTEM_CDN_SERVICE` | `url` | `https://cdn.olares.com` | `true` | `true` | CDN endpoint for system resources. |
| `OLARES_SYSTEM_DOCKERHUB_SERVICE` | `url` | None | `true` | `false` | Docker Hub mirror or accelerator endpoint. |
| `OLARES_SYSTEM_ROOT_PATH` | `string` | `/olares` | `false` | `true` | Olares root directory path. |
| `OLARES_SYSTEM_ROOTFS_TYPE` | `string` | `fs` | `false` | `true` | Olares filesystem type. |
| `OLARES_SYSTEM_CUDA_VERSION` | `string` | None | `false` | `false` | Host CUDA version. |
