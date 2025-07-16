---
outline: [2, 3]
description: 了解如何通过 WSL 2 和 Docker 在 Windows 上安装和运行 Olares，包括系统准备、环境配置以及容器管理。
---
# 使用 Docker 镜像在 Windows 上安装 Olares
你可以通过 Docker 可以在容器化环境中安装和运行 Olares。本文将带你了解：如何使用 Docker 和 WSL 2 设置 Olares，准备安装环境，完成激活流程，并管理容器的生命周期。

:::warning 不适用于生产环境
Windows 版 Olares 目前存在以下限制：
- 不支持分布式存储
- 无法添加本地节点

建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求
Windows 设备需满足以下条件：
- CPU：4 核及以上
- 内存：不少于 16GB 可用内存
- 存储空间：建议使用 SSD，且可用磁盘空间不少于 64GB
- 支持的系统：
    - Windows 10 或 11
    - Linux（WSL 2 环境）：Ubuntu 20.04 LTS 及以上；Debian 11 及以上

## 开始之前
开始安装前，请确保：
- 系统中已安装并运行 [Docker Desktop](https://docs.docker.com/desktop/setup/install/windows-install/)。
   :::info WSL 2 and Hyper-V
   如果 Docker Desktop 的模式为 **Hyper-V**，则无法启用 Olares 的 GPU 支持。请确保 Docker Desktop 在 **WSL 2** 模式下运行。  
   :::
- 已知当前设备的 IP 地址。
  ::: tip 查看 IP 地址
  在 PowerShell 或命令提示符中，使用下列命令确认 IP 地址：
  ```bash
  ipconfig | findstr /i "IPv4.*192"
  ```
  :::
- 已通过 LarePass [创建 Olares ID](/zh/manual/get-started/create-olares-id.md) 且使用默认的 `olares.cn` 域名。

## 配置 WSL 2
1. 打开 PowerShell，运行以下命令确认系统中安装的 WSL 内核版本：
   ```powershell
   wsl --version
   ```
   示例输出：
   ```PowerShell{2}
   WSL 版本： 2.4.8.0
   内核版本： 5.15.167.4-1
   WSLg 版本： 1.0.65
   MSRDC 版本： 1.611.1-81528511
   DXCore 版本： 10.0.26100.1-240331-1435.ge-release
   Windows 版本： 10.0.26100.3475
   ```
2. 下载与 WSL 内核版本相匹配的文件：`https://cdn.joinolares.cn/bzImage-<内核版本号>`。
  例如，`5.15.167.4-1` 版本对应的链接是 [https://cdn.joinolares.cn/bzImage-5.15.167.4](https://cdn.joinolares.cn/bzImage-5.15.167.4)。

   目前支持以下内核版本（`5.15.146.1` 及以上）：
   -  `linux-msft-wsl-5.15.146.1`
   -  `linux-msft-wsl-5.15.150.1`
   -  `linux-msft-wsl-5.15.153.1`
   -  `linux-msft-wsl-5.15.167.4`
   -  `linux-msft-wsl-6.6.75.1`
   -  `linux-msft-wsl-6.6.36.6`
   -  `linux-msft-wsl-6.6.36.3`
3. 设置 WSL 使用的默认版本：
   ```bash
   wsl --set-default-version 2
   ```
4. 在 `C:\Users\<YourUsername>\` 目录下创建文件 `.wslconfig`，填入以下内容：
   ```txt
   [wsl2]
   kernel=c:\\path\\to\\your\\kernel\\bzImage-<version> # 注意：使用双反斜杠 (\\) 作为路径分隔符
   memory=8GB # 建议设置为 16GB
   swap=0GB
   ```
   :::info
   如果在修改 `.wslconfig` 文件之前已经安装了 Docker Desktop，建议先删除 WSL 中已安装的 `docker-desktop` 发行版：
   ```bash
   wsl --unregister docker-desktop
   wsl --unregister docker-desktop-data # 如果存在此版本
   ```
5. 重启 Windows 使变更生效。

## 更新 Docker 的镜像源
添加 Olares 的镜像源，提高镜像拉取速度：
1. 打开 Docker Desktop，选择 **Settings** > **Docker Engine**。
2. 修改 Docker daemon 的 json 文件，添加镜像源：
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
3. 点击 **Apply & restart** 使变更生效。

## 运行 `olaresd-proxy`
1. 下载 `olaresd-proxy`：[https://cdn.joinolares.cn/olaresd-proxy-v0.1.0-windows-amd64.tar.gz](https://cdn.joinolares.cn/olaresd-proxy-v0.1.0-windows-amd64.tar.gz)。
2. 解压文件，打开 `olaresd-proxy`。
   :::info 保持 `olaresd-proxy` 在后台运行
   在安装和激活 Olares 的整个过程中，确保 `olaresd-proxy` 在后台运行。
   :::

## 使用 Docker CLI 运行 Olares
:::warning CUDA 版本要求
如果需要启用 GPU 支持，请确保 CUDA 版本为 12.4 或以上。较低版本不支持 GPU 功能。
:::
使用下列命令拉取 Olares 的镜像。
将 `<host ip>` 替换为设备的 IP 地址，将 `<olares version>-cn` 替换为想要使用的 Olares 版本：
::: code-group
```bash{2,9} [无 GPU]
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
```bash{1,2,9} [支持 GPU]
docker run --gpus all -d --privileged -v oic-data:/var \
  -e HOST_IP=<host ip> \
  -p 80:80 \
  -p 443:443 \
  -p 30180:30180 \
  -p 18088:18088 \
  -p 41641:41641/udp \
  --name oic \
  beclab/olares:<olares version>-cn
```
:::
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
`--rm` 参数会在容器停止后自动删除容器。如果发生这种情况，将无法重新启动容器，必须重新安装 Olares 才能再次运行。不使用此参数可以在停止后保留容器，让你能够通过 `docker start` 命令恢复运行。
:::

<!--@include: ./install-and-activate-olares.md-->

<!--@include: ./manage-olares-container.md-->

<!--@include: ./reusables.md{30,34}-->
