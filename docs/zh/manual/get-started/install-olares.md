---
description: 在 Linux 系统上通过一键脚本，快速上手 Olares。
---
# 安装 Olares

本文介绍如何在 Linux 系统上安装 Olares。我们推荐在 **Linux 系统**（如 Ubuntu 或 Debian）上部署 Olares，以在生产环境中获得最佳的性能与稳定性。

开始安装前，请先[创建 Olares ID](create-olares-id.md)，并确认操作系统与硬件已满足最低要求。

:::info 安装遇到问题？
如果安装过程中遇到问题，[可以提交 GitHub Issue](https://github.com/beclab/Olares/issues/new)。提交时请提供以下信息：
- 使用的平台或环境（如 Ubuntu、Docker、WSL 等）。
- 安装方式（脚本安装或 Docker 镜像）。
- 详细的错误信息（包括日志、错误提示或截图）。
:::

## 系统要求

请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：不少于 150GB 的可用磁盘空间，需使用 SSD 硬盘安装。
- 支持的系统版本：
    - Ubuntu 22.04 LTS 及以上
    - Debian 12 及以上

:::warning 注意
使用 HDD（机械硬盘）可能会导致安装失败。
:::

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台上安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

在 Linux 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,32}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{34,38}-->
