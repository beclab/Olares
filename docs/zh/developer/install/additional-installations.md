---
description: 在 macOS、Windows、PVE、树莓派等平台的容器或虚拟环境中安装 Olares，用于开发和测试场景的安装指南。不适用于生产环境。
---

# 其他安装方式

本部分文档介绍如何在 Linux、macOS、Windows、PVE 或 Raspberry Pi 平台上的容器化或虚拟环境下安装和运行 Olares，**仅适用于开发或测试环境**。

::: tip 生产环境安装建议
Olares 已针对 Linux 系统（Ubuntu 或 Debian）进行了优化。我们推荐[通过脚本在 Linux 上安装 Olares](/manual/get-started/install-olares.md)，以在生产环境中获得最佳的性能与稳定性。
:::

在安装之前，请先[创建 Olares ID](../../manual/get-started/create-olares-id.md)，并确保操作系统与硬件满足最低要求。

请选择你的平台开始安装：
- [在 Linux 上安装（通过 Docker 镜像）](linux-via-docker-compose.md)
- [在 macOS 上安装](mac.md)
- [在 Windows（WSL 2）上安装](windows.md)
- [在 PVE 上安装](pve.md)
- [在 LXC 上安装](lxc.md)
- [在树莓派上安装](raspberry-pi.md)