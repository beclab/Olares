---
outline: [2, 3]
description: 了解如何在 Mac 上使用 Docker 容器部署运行 Olares 的完整步骤，包括镜像配置和容器设置说明。
---
# 使用 Docker 镜像在 Mac 上安装 Olares

你可以通过 Docker 可以在容器化环境中安装和运行 Olares。本文将带你了解：如何使用 Docker 设置 Olares，准备安装环境，完成激活流程，并管理容器的生命周期。

:::warning 不适用于生产环境
Mac 版 Olares 目前存在以下限制：
- 不支持分布式存储
- 无法添加本地节点

建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求
Mac 设备需满足以下条件：
- 处理器架构：AMD64 或 ARM64
- 内存：可用内存 8 GB 及以上
- 存储空间：可用磁盘空间 90 GB 及以上
- MacOS 版本：Monterey（12）及以上

## 开始之前
开始安装前，请确保：
- 系统中已安装并运行 [Docker](https://docs.docker.com/engine/install/)。
- 已知当前设备的 IP 地址。
  ::: tip 查看 IP 地址
  要查看 Mac 的 IP 地址，可以使用两种方式:
  - 使用图形界面：打开**系统设置**（或**系统偏好设置**）> **网络**，在当前活动的网络连接中查看详细信息。
  - 使用命令行：打开终端窗口，Wi-Fi 网络输入 `ipconfig getifaddr en0`，有线网络输入 `ipconfig getifaddr en1`。
  :::
- 已通过 LarePass [创建 Olares ID](/zh/manual/get-started/create-olares-id.md) 且使用默认的 `olares.cn` 域名。

## 运行 `olaresd-proxy`
::: tip 确认 Mac 芯片  
如果你不确定 Mac 所使用的芯片，请点击苹果菜单并选择**关于本机**，查看芯片类型。
:::
<tabs>
<template #M-系列芯片>

1. 下载`olaresd-proxy`：https://cdn.joinolares.cn/olaresd-proxy-v0.1.0-darwin-arm64.tar.gz 。
2. 解压文件，启动 `olaresd-proxy`。
   :::info 保持 `olaresd-proxy` 在后台运行
   在 Olares 安装和激活期间，保证 `olaresd-proxy` 在后台运行。
   :::
</template>

<template #Intel-芯片>

1. 下载`olaresd-proxy`：https://cdn.joinolares.cn/olaresd-proxy-v0.1.0-darwin-amd64.tar.gz 。
2. 解压文件，启动 `olaresd-proxy`。
   :::info 保持 `olaresd-proxy` 在后台运行
   在 Olares 安装和激活期间，保证 `olaresd-proxy` 在后台运行。
   :::
</template>
</tabs>

## 更新 Docker 的镜像源
添加 Olares 的镜像源，提高镜像拉取速度。以 Docker Desktop 为例：
1. 打开 Docker Desktop，选择 **Settings** > **Docker Engine**。
2. 修改 Docker daemon 的 `json` 文件，添加镜像源：
   ```json{9-11}
   {
     "builder": {
       "gc": {
         "defaultKeepStorage": "20GB",
         "enabled": true
       }
     },
     "experimental": false,
     "registry-mirrors": [
       "https://mirrors.joinolares.cn"
     ]
   }
   ```
3. 点击 **Apply & restart** 保存变更。

## 使用 Docker CLI 运行 Olares

执行以下命令来拉取 Olares 的镜像。

将 `<host ip>` 替换为设备的 IP 地址，将 `<olares version>-cn` 替换为想要使用的 Olares 版本：

```bash{2,9}
docker run -d --privileged -v oic-data:/var \
  -e HOST_IP=<host ip> \
  -p 80:80 \
  -p 443:443 \
  -p 30180:30180 \
  -p 18088:18088 \
  -p 41641:41641/udp \
  --name oic \
  beclab/olares:<olares version>-cn
```
其中：
  - `-d`：以分离模式（detached mode）启动容器，允许其在后台运行。
  - `--privileged`：授予容器完整的系统权限。
  - `-v oic-data:/var`：将 Docker 数据卷（`oic-data`）挂载到容器内的 `/var` 目录以持久化数据。
  - `-e HOST_IP=<host ip>`：设置主机设备的 IP 地址作为环境变量
  - `-p 80:80`：将主机的 `80` 端口映射到容器的 `80` 端口。
  - `-p 443:443`：将主机的 `443` 端口映射到容器的 4`43` 端口。
  - `-p 30180:30180`：将主机的 `30180` 端口映射到容器的 `30180` 端口。
  - `-p 18088:18088`：将宿主机的 `18088` 端口映射到容器的 `18088` 端口。
  - `-p 41641:41641/udp`：将宿主机的 `41641` UDP 端口映射到容器的 `41641` UDP 端口。
  - `--name oic`：将容器命名为 `oic`（Olares in container）方便后续引用。
  - `beclab/olares:<olares version>-cn`：指定 Olares Docker 镜像及版本，例如`beclab/olares:1.11.5-cn`。

容器启动后，你会看到一个容器 ID。
:::warning 请勿添加 `--rm` 参数
`--rm` 参数会在容器停止后自动删除容器。如果发生这种情况，将无法重新启动容器，必须重新安装 Olares 才能再次运行。不使用此参数可以在停止后保留容器，让你能够通过 docker start 命令恢复运行。
:::

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./manage-olares-container.md-->

<!--@include: ./reusables.md{30,34}-->
