---
outline: [2, 3]
description: System environment variables are cluster-wide global settings maintained by the administrator. Apps must reference them via `envs.valueFrom` and map them to `.Values.olaresEnv`.
---

# System environment variables

System environment variables are global settings for each Olares cluster instance. They are configured during system installation or maintained by the cluster administrator. All users in the cluster share the same set of system environment variables.

:::info
System environment variables are a "variable pool". Apps cannot modify them directly—they are maintained by the administrator. Apps must map them to their own `envName` via `envs.valueFrom`, and then use them in templates via `.Values.olaresEnv.<envName>`.
:::

## How to Use

The following example shows how to reference a system environment variable `APP_CDN_ENDPOINT` in an app.

1. Declare the mapping in `OlaresManifest.yaml`. Under `envs`, declare an app variable and use `valueFrom` to reference the system environment variable.

    **Example**:
    ```yaml
    olaresManifest.version: '0.10.0'
    olaresManifest.type: app

    envs:
      - envName: APP_CDN_ENDPOINT
        required: true
        applyOnChange: true
        valueFrom:
          envName: OLARES_SYSTEM_CDN_SERVICE
    ```

2. In your Helm template where the variable is needed, reference it using the `.Values.olaresEnv` path.

    **Example**:
    ```yaml
    value: "{{ .Values.olaresEnv.APP_CDN_ENDPOINT }}"
    ```

3. When the app is deployed through App Service, the system injects the corresponding system environment variable value into `values.yaml`.

    **Example**:
    ```yaml
    olaresEnv:
      APP_CDN_ENDPOINT: "https://cdn.olares.com"
    ```
## System environment variables reference

### OLARES_SYSTEM_REMOTE_SERVICE

Remote service endpoint for the Olares system (such as the Market, Olares Space, etc.)

- Type: `url`
- Default: `https://api.olares.com`
- Editable:  Yes
- Required: Yes

### OLARES_SYSTEM_CDN_SERVICE

CDN endpoint for system resources

- Type: `url`
- Default: `https://cdn.olares.com`
- Editable: Yes
- Required: Yes

### OLARES_SYSTEM_DOCKERHUB_SERVICE

Docker Hub mirror/accelerator endpoint

- Type: `url`
- Editable: Yes
- Required: No

### OLARES_SYSTEM_ROOT_PATH

Olares root directory path

- Type: `string`
- Default: `/olares`
- Editable: No
- Required: Yes

### OLARES_SYSTEM_ROOTFS_TYPE

Olares filesystem type

- Type: `string`
- Default: `fs`
- Editable: No
- Required: Yes

### OLARES_SYSTEM_CUDA_VERSION

Host CUDA version

- Type: `string`
- Editable: No
- Required: No

## Full Structure Example
```yaml
systemEnvs:
    
  - envName: OLARES_SYSTEM_REMOTE_SERVICE
    default: "https://api.olares.com"
    type: url
    editable: true
    required: true
    
  - envName: OLARES_SYSTEM_CDN_SERVICE
    default: "https://cdn.olares.com"
    type: url
    editable: true
    required: true
    
  - envName: OLARES_SYSTEM_DOCKERHUB_SERVICE
    type: url
    editable: true
    required: false
    
  - envName: OLARES_SYSTEM_ROOT_PATH
    default: /olares
    editable: false
    required: true
    
  - envName: OLARES_SYSTEM_ROOTFS_TYPE
    default: fs
    editable: false
    required: true
    
  - envName: OLARES_SYSTEM_CUDA_VERSION
    editable: false
    required: false
```