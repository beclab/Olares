---
outline: [2, 3]
description: Environment variables available in Olares for customizing networking, authentication, GPU support and other features. Includes configuration examples and specifications.
---
# Olares environment variables

Olares provides a highly customizable installation process through the use of environment variables. These variables can override default settings, enabling advanced configurations to suit your specific requirements.

## Usage examples

To customize the installation process, you can set the environment variables before running the installation command. For example:

```bash
# Specify Kubernetes (k8s) instead of k3s
export KUBE_TYPE=k8s \
&& curl -sSfL https://olares.sh | bash -
```
Or, if you have already downloaded the installation script `install.sh`:

```bash
# Specify Kubernetes (k8s) instead of k3s
export KUBE_TYPE=k8s && bash install.sh
```
Both methods achieve the same result. The environment variable `KUBE_TYPE` will be passed to the script, and the script will use it to modify its behavior.

## Environment variables reference

The section lists all the environment variables, along with their default values, optional values, and descriptions. Configure them as needed.

### `CLOUDFLARE_ENABLE`
Specifies whether to enable the Cloudflare proxy.
- **Valid values**:
  - `0` (disable)
  - `1` (enable)
- **Default**: `0`

### `DID_GATE_URL`
Specifies the endpoint for the DID gateway.
- **Valid values**:
  - `https://did-gate-v3.bttcdn.com`
  - `https://did-gate-v3.api.jointerminus.cn/` (recommended for better connectivity in mainland China)
- **Default**: `https://did-gate-v3.bttcdn.com`

### `FIREBASE_PUSH_URL`
Specifies the endpoint for Firebase push services.
- **Valid values**:
  - `https://firebase-push-test.bttcdn.com/v1/api/push`
  - `https://firebase-push-test.api.jointerminus.cn/v1/api/push` (recommended for better connectivity in mainland China)
- **Default**: `https://firebase-push-test.bttcdn.com/v1/api/push`

### `FRP_AUTH_METHOD`
Sets the FRP authentication method.
- **Valid values**:
  - `jws`
  - `token` (requires `FRP_AUTH_TOKEN`)
  - (empty) – No authentication
- **Default**: `jws`

### `FRP_AUTH_TOKEN`
Specifies the token for FRP communication (required if `FRP_AUTH_METHOD=token`).
- **Valid values**: Any non-empty string
- **Default**: None

### `FRP_ENABLE`
Specifies whether to enable FRP for internal network tunneling. Requires additional FRP-related variables if using a custom server.
- **Valid values**:
  - `0` (disable)
  - `1` (enable)
- **Default**: `0`

### `FRP_LIST_URL`
Specifies the endpoint for the Olares FRP information service.
- **Valid values**:
  - `https://terminus-frp.snowinning.com`
  - `https://terminus-frp.api.jointerminus.cn` (recommended for better connectivity in mainland China)
- **Default**: `https://terminus-frp.snowinning.com`

### `FRP_PORT`
Specifies the FRP server's listening port.
- **Valid values**: An integer in the range `1–65535`
- **Default**: `7000` (if not set or set to `0`)

### `JUICEFS`
Installs [JuiceFS](https://juicefs.com/) alongside Olares.
- **Valid values**: `1`
- **Default**: None (does not install JuiceFS if not set)

### `KUBE_TYPE`
Determines the Kubernetes distribution to install.
- **Valid values**:
  - `k8s` (full Kubernetes)
  - `k3s` (lightweight Kubernetes)
- **Default**: `k3s`

### `LOCAL_GPU_ENABLE`
Specifies whether to enable GPU support and install related drivers.
- **Valid values**:
  - `0` (disable)
  - `1` (enable)
- **Default**: `0`

### `LOCAL_GPU_SHARE`
Specifies whether to enable GPU sharing. Applies only if GPU is enabled.
- **Valid values**:
  - `0` (disable)
  - `1` (enable)
- **Default**: `0`

### `MARKET_PROVIDER`
Specifies the backend domain used by the application marketplace (Market).
- **Valid values**:
  - `appstore-server-prod.bttcdn.com`
  - `appstore-china-server-prod.api.jointerminus.cn` (recommended for better connectivity in mainland China)
- **Default**: `appstore-server-prod.bttcdn.com`

### `NVIDIA_CONTAINER_REPO_MIRROR`
Specifies the APT repository mirror for installing NVIDIA Container Toolkit.
- **Valid values**:
  - `nvidia.github.io`
  - `mirrors.ustc.edu.cn` (recommended for better connectivity in mainland China)
- **Default**: `nvidia.github.io`

### `OLARES_SPACE_URL`
Specifies the endpoint for the Olares Space service.
- **Valid values**:
  - `https://cloud-api.bttcdn.com/`
  - `https://cloud-api.api.jointerminus.cn/` (recommended for better connectivity in mainland China)
- **Default**: `https://cloud-api.bttcdn.com/`

### `PREINSTALL`
Runs only the pre-installation phase (system dependency setup) without proceeding to the full Olares installation.
- **Valid values**: `1`
- **Default**: None (performs full installation if not set)

### `PUBLICLY_ACCESSIBLE`
Explicitly specifies that this machine is accessible publicly on the internet, and a reverse proxy should not be used.
- **Valid values**: 
  - `0` (false)
  - `1` (true)
- **Default**: `0`


### `REGISTRY_MIRRORS`
Specifies a custom Docker registry mirror for faster image pulls.
- **Valid values**: `https://mirrors.joinolares.cn` or any other valid URL
- **Default**: `https://registry-1.docker.io`

### `TAILSCALE_CONTROLPLANE_URL`
Specifies the endpoint for the Olares Tailscale control-plane service.
- **Valid values**:
  - `https://controlplane.snowinning.com`
  - `https://controlplane.api.jointerminus.cn` (recommended for better connectivity in mainland China)
- **Default**: `https://controlplane.snowinning.com`

### `TERMINUS_CERT_SERVICE_API`
Specifies the endpoint for the Olares HTTPS certificate service.
- **Valid values**:
  - `https://terminus-cert.snowinning.com`
  - `https://terminus-cert.api.jointerminus.cn` (recommended for better connectivity in mainland China)
- **Default**: `https://terminus-cert.snowinning.com`

### `TERMINUS_DNS_SERVICE_API`
Specifies the endpoint for the Olares DNS service.
- **Valid values**:
  - `https://terminus-dnsop.snowinning.com`
  - `https://terminus-dnsop.api.jointerminus.cn` (recommended for better connectivity in mainland China)
- **Default**: `https://terminus-dnsop.snowinning.com`

### `TERMINUS_IS_CLOUD_VERSION`
Marks the machine explicitly as a cloud instance.
- **Valid values**: `true`
- **Default**: None

### `TERMINUS_OS_DOMAINNAME`
Sets the domain name before installation to skip the interactive prompt.
- **Valid values**: Any valid domain name
- **Default**: None (prompts for domain name if not set)

### `TERMINUS_OS_EMAIL`
Specifies the email address to use instead of a generated one.
- **Valid values**: Any valid email address
- **Default**: None (a temporary email is generated if not set)

### `TERMINUS_OS_PASSWORD`
Specifies the password to use instead of a generated one.
- **Valid values**: A valid password with 6–32 characters
- **Default**: A randomly generated 8-character password

### `TERMINUS_OS_USERNAME`
Specifies the username before installation to skip the interactive prompt.
- **Valid values**: Any valid username (2–250 characters, excluding reserved keywords)
- **Default**: None (prompts for username if not set)
- **Validation**: Reserved keywords include `user`, `system`, `space`, `default`, `os`, `kubesphere`, `kube`, `kubekey`, `kubernetes`, `gpu`, `tapr`, `bfl`, `bytetrade`, `project`, `pod`

### `TOKEN_MAX_AGE`
Sets the maximum validity period for a token (in seconds).
- **Valid values**: Any integer (in seconds)
- **Default**: `31536000` (365 days)