---
description: 概览 Olares 支持的安装方式。推荐在 Linux 环境下通过 ISO 镜像或一行命令安装。其他平台（如 macOS、Windows、PVE、Raspberry Pi）适用于测试和开发。
outline: [2,4]
---

# 安装 Olares

本文介绍 Olares 支持的安装方式。 在开始安装前，请确保已完成以下准备工作：
- 已创建 [Olares ID](create-olares-id.md)。
- 确认操作系统和硬件满足具体安装文档里列出的最低要求。

## 选择合适的安装方式

Olares 支持多平台、多部署方式。请根据你的使用场景选择最合适的安装方式。

### 生产环境推荐方式

推荐在 **Linux（Ubuntu 或 Debian）** 系统上运行 Olares，以获得最佳性能和稳定性。

- [**ISO 镜像**](install-linux-iso.md) —  通过官方 ISO 镜像在物理机上全新安装 Olares，自动配置宿主（Linux）<br/>环境、容器运行时、驱动及核心依赖。
- [**一行命令**](install-linux-script.md) —  在现有 Linux 系统中快速安装 Olares。

:::tip 推荐
在主机上使用 ISO 安装方式可确保最佳的兼容性、性能与系统级优化，是运行 Olares 的首选方案。
:::

### 其他安装方式

以下方式仅适用于**开发测试**或**轻量级环境**。

#### Linux
- [**Docker 镜像**](install-linux-docker.md) — 在 Linux 上通过 Docker Compose 以容器化方式运行 Olares。

#### Windows
- [**一行命令**](install-windows-script.md) — 在 WSL2 虚拟化环境中安装 Olares。
- [**Docker 镜像**](install-windows-docker.md) — 在 WSL2 的 Docker 容器中运行 Olares。

#### macOS
- [**一行命令**](install-mac-script.md) — 使用 MiniKube 在容器化环境中安装 Olares。
- [**Docker 镜像**](install-mac-docker.md) — 在 macOS 上通过 Docker 部署 Olares。

#### PVE
- [**ISO 镜像**](install-pve-iso.md) — 在 Proxmox VE 中通过 ISO 镜像以虚拟机方式部署 Olares。
- [**一行命令**](install-pve-script.md) — 直接在 PVE 节点上安装 Olares。
- [LXC](install-lxc.md) - 在 PVE 环境下使用 Linux 容器 （LXC）部署 Olares。

#### Raspberry Pi（ARM）

- [**一行命令**](install-raspberry-pi.md)：在基于 ARM 架构的 Raspberry Pi 设备上安装 Olares。

