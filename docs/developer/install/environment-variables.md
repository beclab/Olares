---
outline: [2, 3]
description: Reference for environment variables used to customize the Olares installation process.
---
# Customize Olares installation with environment variables

Olares supports advanced installation customization through shell-level environment variables.

These variables are evaluated by the installation script and override default settings, allowing you to configure networking behavior, hardware support, Kubernetes distribution, and initial system setup.

:::info 
These variables are only effective during installation. They are not available as environment variables inside application Helm charts, containers, or runtime environments.
:::

## Usage example

Set the environment variables in your terminal before running the installation command.

### Use the official installer
```bash
# Specify Kubernetes (k8s) instead of k3s
export KUBE_TYPE=k8s \
&& curl -sSfL https://olares.sh | bash -
```

### Use a local installation script

```bash
# Specify Kubernetes (k8s) instead of k3s
export KUBE_TYPE=k8s && bash install.sh
```
Both methods achieve the same result. The environment variable `KUBE_TYPE` is passed to the installation script and modifies its behavior accordingly.

## Environment variables reference

### Kubernetes & infrastructure

|Variable| Type | Default | Allowed values | Description |
|--|--|--|--|--|
| `KUBE_TYPE` | String | `k3s` | `k3s`, `k8s` | Determines the Kubernetes distribution to install. |
| `PREINSTALL` | Integer | — | `1` | Runs only the system dependency setup phase.|
| `JUICEFS` | Integer | — | `1` | Installs [JuiceFS](https://juicefs.com/) alongside Olares.|
| `REGISTRY_MIRRORS` | URL | `https://registry-1.docker.io` | Valid URL | Specifies a custom Docker registry mirror for image pulls. |

### Networking & exposure
|Variable| Type | Default | Allowed values | Description |
|--|--|--|--|--|
| `PUBLICLY_ACCESSIBLE` |	Integer |	`0`	| `0`, `1` |If `1`, specifies machine is publicly accessible, and reverse proxy is skipped. |
| `CLOUDFLARE_ENABLE` |	Integer |	`0` |	`0`, `1` | If `1`, enables Cloudflare proxy support. |
| `FRP_ENABLE` | Integer | `0` | `0`, `1` |	If `1`, enables FRP for internal network tunneling.|
| `FRP_AUTH_METHOD` |	String	| `jws` |	`jws`, `token`, empty string | Sets the FRP authentication method. When set to `token`, `FRP_AUTH_TOKEN` is required. When set to empty string, no authentication is used. |
| `FRP_AUTH_TOKEN` | String	| — |	Any non-empty string | Token for FRP communication (required if `FRP_AUTH_METHOD`=`token`).|
| `FRP_PORT` | Integer | `7000` |	`1–65535`	| Specifies the FRP server's listening port. Defaults to `7000` if unset or set to `0`. |

### GPU configuration
|Variable| Type | Default | Allowed values | Description |
|--|--|--|--|--|
| `LOCAL_GPU_ENABLE` | Integer | `0` | `0`, `1` | If `1`, enables GPU support and installs related drivers. |
| `LOCAL_GPU_SHARE` | Integer | `0`| `0`, `1` | If `1`, enables GPU sharing (applies only if GPU is enabled). |
| `NVIDIA_CONTAINER_REPO_MIRROR` | String | `nvidia.github.io` | Mirror URL | Specifies the APT repository mirror for installing the NVIDIA Container Toolkit. |

### Initial account & domain setup
These variables skip interactive prompts during installation.

| Variable | Type | Default | Validation | Description |
|--|--|--|--|--|
| `TERMINUS_OS_USERNAME` | String | Prompt | 2–250 characters | Admin username (no reserved keywords). |
| `TERMINUS_OS_PASSWORD` | String | Random 8-character password | 6–32 characters | Admin password. |
| `TERMINUS_OS_EMAIL` | String | Temporary generated email | Valid email | Admin email address. |
| `TERMINUS_OS_DOMAINNAME` | String | Prompt | Valid domain | System domain name. |
| `TERMINUS_IS_CLOUD_VERSION`| Boolean | — | `true` | Marks the machine explicitly as a cloud instance. |

:::info Reserved keywords
`user`, `system`, `space`, `default`, `os`, `kubesphere`, `kube`, `kubekey`, `kubernetes`, `gpu`, `tapr`, `bfl`, `bytetrade`, `project`, `pod`.
:::

### Security
| Variable | Type | Default | Description |
|--|--|--|--|
| `TOKEN_MAX_AGE` | Integer | `31536000` | Maximum validity period for authentication tokens (in seconds). |