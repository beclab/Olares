---
outline: [2, 3]
description: Olares 提供的环境变量，用于自定义网络、认证、GPU 支持等功能。包含配置示例和详细规格说明。
---
# Olares 环境变量参考

Olares 提供了丰富的环境变量以满足定制化安装需求。通过修改这些变量，你可以覆盖默认安装设置，实现灵活的个性化安装配置。

## 使用示例

你可以在运行安装命令前设置环境变量，以自定义安装流程。例如：

```bash
# 指定安装完整的 Kubernetes (k8s) 而非轻量级 k3s
export KUBE_TYPE=k8s \
&& curl -sSfL https://olares.sh | bash -
```

如果你已预先下载了安装脚本 `install.sh`，也可以使用以下方式：

```bash
# 指定使用完整的 Kubernetes (k8s) 而非轻量级 k3s
export KUBE_TYPE=k8s && bash install.sh
```
两种方式的执行效果相同：环境变量 `KUBE_TYPE` 会传递给安装脚本，脚本会根据这个变量来调整其安装逻辑。

当然，你也可以组合多个环境变量来实现更灵活的自定义效果。例如中国大陆的用户通过`cn.olares.sh`获取的安装脚本，就是一个在默认安装脚本之上设置了一系列环境变量的脚本：

```bash
curl -fsSL https://cn.olares.sh
#!/bin/bash

export FRP_ENABLE=1 \
    FRP_SERVER="http://frp-bj.api.jointerminus.cn" \
    FRP_PORT=0 \
    JUICEFS=0 \
    FRP_AUTH_METHOD="jws" \
    REGISTRY_MIRRORS="https://mirrors.joinolares.cn" \
    DOWNLOAD_CDN_URL="https://cdn.joinolares.cn" \
    MARKET_PROVIDER="appstore-china-server-prod.api.jointerminus.cn" \
    TERMINUS_CERT_SERVICE_API="https://terminus-cert.api.jointerminus.cn" \
    TERMINUS_DNS_SERVICE_API="https://terminus-dnsop.api.jointerminus.cn" \
    DID_GATE_URL="https://did-gate-v3.api.jointerminus.cn/" \
    OLARES_SPACE_URL="https://cloud-api.api.jointerminus.cn/" \
    FIREBASE_PUSH_URL="https://firebase-push-test.api.jointerminus.cn/v1/api/push" \
    FRP_LIST_URL="https://terminus-frp.api.jointerminus.cn/" \
    TAILSCALE_CONTROLPLANE_URL="https://controlplane.api.jointerminus.cn"

curl -sSfL https://olares.sh | bash
```

## 环境变量参考

以下列出了安装脚本所支持的环境变量及其默认值、可选值和说明。请根据具体需求进行配置。

### `CLOUDFLARE_ENABLE`
指定是否启用 Cloudflare 代理。  
- **可选值**：
  - `0`（禁用）
  - `1`（启用）
- **默认值**：`0`（禁用）

### `DID_GATE_URL`
指定 DID 网关服务的地址。
- **可选值**： 
  - `https://did-gate-v3.bttcdn.com` 
  - `https://did-gate-v3.api.jointerminus.cn/`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://did-gate-v3.bttcdn.com/`

### `FIREBASE_PUSH_URL`
指定 Firebase 推送服务的地址。
- **可选值**： 
  - `https://firebase-push-test.bttcdn.com/v1/api/push`  
  - `https://firebase-push-test.api.jointerminus.cn/v1/api/push`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://firebase-push-test.bttcdn.com/v1/api/push`

### `FRP_AUTH_METHOD`
设置 FRP 的认证方式。
- **可选值**：
  - `jws`
  - `token`（需要设置 `FRP_AUTH_TOKEN`）
  - *（空字符串）* —— 不使用认证
- **默认值**：`jws`

### `FRP_AUTH_TOKEN`
当 `FRP_AUTH_METHOD=token` 时，用于指定服务器通信所需的 Token。  
- **可选值**：任意非空字符串  
- **默认值**：无

### `FRP_ENABLE`
指定是否启用 FRP 内网穿透。如果使用自定义 FRP 服务器，还需额外设置相关变量。  
- **可选值**：
  - `0`（禁用）
  - `1`（启用） 
- **默认值**：`0`（禁用）

### `FRP_LIST_URL`
指定 Olares FRP 信息服务的地址。 
- **可选值**： 
  - `https://terminus-frp.snowinning.com`  
  - `https://terminus-frp.api.jointerminus.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://terminus-frp.snowinning.com`

### `FRP_PORT`
设置 FRP 服务端监听端口。
- **可选值**：整数范围 `1～65535`  
- **默认值**：未设置或设为 `0` 时默认为 `7000`

### `JUICEFS`
随 Olares 一起安装 [JuiceFS](https://juicefs.com/)。
- **可选值**：`1`  
- **默认值**：无（若不设置则不安装 JuiceFS）

### `KUBE_TYPE`
指定要使用的 Kubernetes 发行版。
- **可选值**：
  - `k8s`（完整的 Kubernetes）
  - `k3s`（Kubernetes 的轻量级发行版）
- **默认值**：`k3s`

### `LOCAL_GPU_ENABLE`
指定是否启用 GPU 并安装相关驱动。  
- **可选值**：
  - `0`（禁用）
  - `1`（启用）
- **默认值**：`0`（禁用）

### `LOCAL_GPU_SHARE`
指定是否启用 GPU 共享功能。仅在已启用 GPU 时生效。  
- **可选值**：
  - `0`（禁用）
  - `1`（启用）
- **默认值**：`0`（禁用）

### `MARKET_PROVIDER`
指定应用市场（Market）后端服务所使用的域名，可根据网络环境选择合适的域名以优化访问速度。  
- **可选值**： 
  - `appstore-server-prod.bttcdn.com`  
  - `appstore-china-server-prod.api.jointerminus.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`appstore-server-prod.bttcdn.com`

### `NVIDIA_CONTAINER_REPO_MIRROR`
配置 `nvidia-container-toolkit` 的 APT 安装镜像源。
- **可选值**：
  - `nvidia.github.io`
  - `mirrors.ustc.edu.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`nvidia.github.io`

### `OLARES_SPACE_URL`
指定 Olares Space 服务的地址。
- **可选值**： 
  - `https://cloud-api.bttcdn.com/`  
  - `https://cloud-api.api.jointerminus.cn/`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://cloud-api.bttcdn.com/`

### `PREINSTALL`
仅执行预安装阶段（系统依赖配置），不进行完整的 Olares 安装。
- **可选值**：`1`  
- **默认值**：无（若不设置则执行完整安装）

### `PUBLICLY_ACCESSIBLE`
明确指定该机器可以在互联网上公开访问，同时不设置反向代理。
- **可选值**： 
  - `0` (否)
  - `1` (是)
- **默认值**: `0`

### `REGISTRY_MIRRORS`
设置 Docker 镜像加速地址。
- **可选值**：`https://mirrors.joinolares.cn` 或其他镜像源地址  
- **默认值**：`https://registry-1.docker.io`

### `TAILSCALE_CONTROLPLANE_URL`
指定 Olares Tailscale 控制平面（control-plane）服务的地址。
- **可选值**： 
  - `https://controlplane.snowinning.com`  
  - `https://controlplane.api.jointerminus.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://controlplane.snowinning.com`

### `TERMINUS_CERT_SERVICE_API`
指定 Olares HTTPS 证书服务的地址。
- **可选值**： 
  - `https://terminus-cert.snowinning.com`  
  - `https://terminus-cert.api.jointerminus.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://terminus-cert.snowinning.com`

### `TERMINUS_DNS_SERVICE_API`
指定 Olares DNS 服务的地址。
- **可选值**： 
  - `https://terminus-dnsop.snowinning.com`  
  - `https://terminus-dnsop.api.jointerminus.cn`（推荐中国大陆用户使用，连接性更好）
- **默认值**：`https://terminus-dnsop.snowinning.com`

### `TERMINUS_IS_CLOUD_VERSION`
明确将此机器标记为云端实例（cloud instance）。  
- **可选值**：`true`  
- **默认值**：无

### `TERMINUS_OS_DOMAINNAME`
在安装前预先设置域名，会跳过安装过程中的交互式提示。  
- **可选值**：任意有效域名  
- **默认值**：无（若不设置则会提示输入域名）

### `TERMINUS_OS_EMAIL`
在安装前预先设置邮箱地址，会跳过安装过程中的交互式提示。  
- **可选值**：任意有效邮箱地址  
- **默认值**：无（若不设置则自动生成临时邮箱）

### `TERMINUS_OS_PASSWORD`
在安装前预先设置密码，会跳过安装过程中的交互式提示。  
- **可选值**：6～32 个字符的有效密码
- **默认值**：随机生成 8 位密码

### `TERMINUS_OS_USERNAME`
在安装前预先设置用户名，会跳过安装过程中的对应交互式提示。
- **可选值**：任意有效用户名（长度 2～250，且不与保留关键词冲突）  
- **默认值**：无（若不设置则会提示输入用户名）
- **验证规则**：保留关键词包括 `user`、`system`、`space`、`default`、`os`、`kubesphere`、`kube`、`kubekey`、`kubernetes`、`gpu`、`tapr`、`bfl`、`bytetrade`、`project`、`pod`

### `TOKEN_MAX_AGE`
设置 Token 的最大有效时长（单位：秒）。  
- **可选值**：任意整数（单位：秒）  
- **默认值**：`31536000`（365 天）