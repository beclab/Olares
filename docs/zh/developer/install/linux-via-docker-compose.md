---
outline: [2, 3]
description: 了解如何使用 Docker Compose 在 Linux 服务器上部署 Olares。本安装指南涵盖系统要求、配置、安装、激活以及容器管理的相关内容。
---
# 使用 Docker Compose 在 Linux 上安装 Olares
通过 Docker 可以在容器化环境中安装和运行 Olares。本文将介绍如何使用 Docker 设置 Olares、准备安装环境、完成激活过程以及管理容器生命周期。

::: warning 不适用于生产环境
该部署方式仅适用于开发或测试环境。我们推荐[通过脚本方式在 Linux 上安装 Olares](/zh/manual/get-started/install-olares.md)，以获得最佳的性能与稳定性。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求

请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：建议使用 SSD，且可用磁盘空间不少于 64GB
- 支持的系统版本：
   - Ubuntu 20.04 LTS 及以上
   - Debian 11 及以上

## 开始之前
开始安装前，请确保：
- 系统中已安装并运行 [Docker](https://docs.docker.com/engine/install/) 和 [Docker Compose](https://docs.docker.com/compose/install/)。
- 已知当前设备的 IP 地址。
  :::tip 查看 IP 地址
  如需确认 IP 地址，在终端中运行以下命令：
  ```bash
  ip r
  ```
  找到以 `default via` 开头的行，对应默认网关和正在使用的网络接口。
  :::
- 已通过 LarePass [创建 Olares ID](/zh/manual/get-started/create-olares-id.md) 且使用默认的 `olares.cn` 域名。

## 创建文件夹
创建文件夹存储 Olares 的配置文件。例如，用如下命令创建名为 `olares-config` 的文件夹：

```bash
mkdir ~/olares-config
cd ~/olares-config
```
## 准备 `docker-compose.yaml`
1. 在 `olares-config` 目录中创建 `docker-compose.yaml` 文件。
2. 根据是否启用 GPU，填入对应的内容：
   :::code-group
   <<< @/code-snippets/docker-compose.yaml
   <<< @/code-snippets/docker-compose-GPU.yaml
   :::
3. 保存 `docker-compose.yaml` 文件。

## 更新 Docker 的镜像源
添加 Olares 的镜像源，提高镜像拉取速度：
1. 打开 `/etc/docker/daemon.json` 文件。
2. 编辑文件，加上以下内容：

   <<< @/code-snippets/docker-daemon.json
3. 重启 Docker 服务以应用更改。
   ```bash
   sudo systemctl restart docker
   ```
4. 验证配置文件是否修改成功：
   ```bash
   docker info
   ```
   在输出的结果中，如输出结果包含如下内容，表示修改成功：

   ```bash
   Registry Mirrors:
   https://mirrors.joinolares.cn/
   ```
## 设置环境变量并启动容器

1. 在 `olares-config` 目录，运行以下命令设置环境变量并启动容器：
   ```bash
   VERSION=<olares version>-cn HOST_IP=<host ip> docker compose up -d
   ```
   - `VERSION=<olares version>-cn`：指定 Olares 镜像的版本。将 `<olares version>-cn` 替换为实际版本，如 `1.11.5-cn`。
   - `HOST_IP=<host ip>`：指定当前主机设备的 IP 地址。将 `<host ip>` 替换为实际地址。

   运行完成后，输出结果如下：
   ```bash
   [+] Running 20/20
   ✔ olaresd-proxy Pulled                                                                           67.8s
   ✔ 688513194d7a Pull complete                                                                    6.8s
   ✔ bfb59b82a9b6 Pull complete                                                                    6.9s
   ✔ efa9d1d5d3a2 Pull complete                                                                    9.5s
   ✔ a62778643d56 Pull complete                                                                    9.6s
   ✔ 7c12895b777b Pull complete                                                                    9.6s
   ✔ 3214acf345c0 Pull complete                                                                   13.6s
   ✔ 5664b15f108b Pull complete                                                                   14.1s
   ✔ 0bab15eea81d Pull complete                                                                   14.2s
   ✔ 4aa0ea1413d3 Pull complete                                                                   15.0s
   ✔ da7816fa955e Pull complete                                                                   15.1s
   ✔ 9aee425378d2 Pull complete                                                                   15.1s
   ✔ 701c983262e9 Pull complete                                                                   36.2s
   ✔ 221438ca359c Pull complete                                                                   36.3s
   ✔ f3d0ed3b32e0 Pull complete                                                                   36.4s
   ✔ 70d5c1f325f6 Pull complete                                                                   43.2s
   ✔ olares Pulled                                                                                5863.6s
   ✔ 2d5815038f40 Pull complete                                                                 5759.0s
   ✔ 13788179ee16 Pull complete                                                                 5831.6s
   ✔ 5a9b10c3302f Pull complete                                                                 5831.7s
    ```

2. 确认容器是否正常运行：
   ```bash
   docker ps
   ```
   输出结果如下：
   ```bash
   CONTAINER ID   IMAGE                         COMMAND                  CREATED              STATUS              PORTS                   NAMES
   28e86c473750   beclab/olaresd:proxy-v0.1.0   "/mdns-agent"            About a minute ago   Up About a minute                           olares-olaresd-proxy-1
   5fd68a8709ad   beclab/olares:1.11.5-cn       "/usr/local/bin/entr…"   2 minutes ago        Up About a minute   0.0.0.0:80->80/tcp...   olares-olares-1
   ```

<!--@include: ./install-and-activate-olares.md-->

## 管理 Olares 容器
在运行任何命令之前，请确保你位于包含 `docker-compose.yaml` 文件的目录中。
### 停止容器
要停止当前正在运行的容器：
```bash
docker compose stop
```

### 重启容器
容器停止后，使用以下命令重启：
```bash
docker compose start
```
容器重启后，所有服务可能需要 6–7 分钟才能完全初始化。在此时间内请耐心等待。

### 卸载容器
要完全删除容器：
```bash
docker compose down
```

<!--@include: ./reusables.md{30,34}-->
   
   

