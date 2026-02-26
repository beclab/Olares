---
outline: [2, 3]
description: 自定义 Olares 安装过程的环境变量说明。
---
# 使用环境变量自定义 Olares 安装

Olares 支持通过 Shell 层级的环境变量进行高级安装自定义。

这些变量会在安装脚本执行时被解析，并覆盖默认设置，使你能够配置网络行为、硬件支持、Kubernetes 发行版以及系统初始化选项。

:::info
这些变量仅在安装阶段生效。它们不会作为环境变量出现在应用的 Helm chart、容器或运行时环境中。
:::

## 使用示例

在运行安装命令之前，在终端中设置所需的环境变量。

### 使用官方安装脚本

```bash
# 指定安装完整的 Kubernetes (k8s) 而非轻量级 k3s
export KUBE_TYPE=k8s \
&& curl -sSfL https://olares.sh | bash -
```

### 使用本地安装脚本

```bash
# 指定使用完整的 Kubernetes (k8s) 而非轻量级 k3s
export KUBE_TYPE=k8s && bash install.sh
```
两种方式的执行效果相同。环境变量 `KUBE_TYPE` 会传递给安装脚本，并据此修改安装行为。

当然，你也可以组合多个环境变量来实现更灵活的自定义效果。例如中国大陆的用户通过`cn.olares.sh`获取的安装脚本，就是一个在默认安装脚本之上设置了一系列环境变量的脚本：

```bash
curl -fsSL https://cn.olares.sh
#!/bin/bash

export FRP_ENABLE=1 \
    FRP_SERVER="http://frp-bj.api.jointerminus.cn" \
    FRP_PORT=0 \
    JUICEFS=0 \
    FRP_AUTH_METHOD="jws" \
    REGISTRY_MIRRORS="https://mirrors.olares.cn" \
    DOWNLOAD_CDN_URL="https://cdn.olares.cn" 

curl -sSfL https://olares.sh | bash
```

## 环境变量参考

以下列出了安装脚本所支持的环境变量及其默认值、可选值和说明。请根据具体需求进行配置。

### Kubernetes 与基础设施

| 变量  | 类型 | 默认值  | 允许值 | 说明 |
| -- | -- | -- | -- | -- |
| `KUBE_TYPE` | String  | `k3s` | `k3s`，`k8s` | 指定要安装的 Kubernetes 发行版。 |
| `PREINSTALL` | Integer | — | `1` | 仅执行系统依赖初始化阶段。|
| `JUICEFS`| Integer | —  | `1` | 在安装 Olares 时同时安装 [JuiceFS](https://juicefs.com/)。 |
| `REGISTRY_MIRRORS` | String | `https://registry-1.docker.io` | 有效的 URL | 指定用于拉取镜像的自定义 Docker 镜像仓库地址。|

### 网络与访问

| 变量  | 类型 | 默认值  | 允许值 | 说明 |
| -- | -- | -- | -- | -- |
| `PUBLICLY_ACCESSIBLE` | Integer | `0`    | `0`，`1` | 设置为 `1` 时，表示机器具备公网访问能力，将跳过反向代理配置。|
| `CLOUDFLARE_ENABLE` | Integer | `0` | `0`，`1` | 设置为 `1` 时，启用 Cloudflare 代理支持。|
| `FRP_ENABLE` | Integer | `0` | `0`，`1` | 设置为 `1` 时，启用 FRP 内网穿透功能。 |
| `FRP_AUTH_METHOD` | String  | `jws`  | `jws`，`token`，空字符串 | 设置 FRP 认证方式。设置为 `token` 时必须配置 `FRP_AUTH_TOKEN`。设置为空字符串时，不使用认证。 |
| `FRP_AUTH_TOKEN` | String  | — | 任意非空字符串 | FRP 通信所使用的 Token（当 `FRP_AUTH_METHOD=token` 时必填）。 |
| `FRP_PORT` | Integer | `7000` | `1–65535` | 指定 FRP 服务监听端口。若未设置或设置为 `0`，默认使用 `7000`。 |

### GPU 配置

| 变量  | 类型 | 默认值  | 允许值 | 说明 |
| -- | -- | -- | -- | -- |
| `LOCAL_GPU_ENABLE` | Integer | `0` | `0`，`1` | 设置为 `1` 时，启用 GPU 支持并安装相关驱动。 |
| `LOCAL_GPU_SHARE` | Integer | `0` | `0`，`1` | 设置为 `1` 时，启用 GPU 共享（仅在启用 GPU 时有效）。                   |
| `NVIDIA_CONTAINER_REPO_MIRROR` | String  | `nvidia.github.io` | 镜像地址 | 指定用于安装 NVIDIA Container Toolkit 的 APT 镜像源。 |

### 初始账户与域名配置

以下变量用于跳过安装过程中的交互式输入。

| 变量  | 类型 | 默认值  | 允许值 | 说明 |
| -- | -- | -- | -- | -- |
| `TERMINUS_OS_USERNAME` | String  | 交互输入 | 2–250 个字符 | 管理员用户名（不可使用保留关键字）。 |
| `TERMINUS_OS_PASSWORD` | String  | 随机生成 8 位密码 | 6–32 个字符 | 管理员密码。|
| `TERMINUS_OS_EMAIL` | String  | 临时生成邮箱 | 合法邮箱格式 | 管理员邮箱地址。 |
| `TERMINUS_OS_DOMAINNAME` | String  | 交互输入 | 合法域名 | 系统域名。 |
| `TERMINUS_IS_CLOUD_VERSION` | Boolean | — | `true` | 明确标记该机器为云端实例。 |

:::info 保留关键字
`user`、`system`、`space`、`default`、`os`、`kubesphere`、`kube`、`kubekey`、`kubernetes`、`gpu`、`tapr`、`bfl`、`bytetrade`、`project`、`pod`。
:::

### 安全相关

| 变量  | 类型 | 默认值 | 说明 |
| -- | -- | -- | -- |
| `TOKEN_MAX_AGE` | Integer | `31536000` | 认证 Token 的最大有效期（单位：秒）。 |