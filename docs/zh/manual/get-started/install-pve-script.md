---
description: 在 PVE 虚拟化平台上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活过程。
---
# 在 PVE 上使用脚本安装 Olares
Proxmox 虚拟环境（PVE）是一个基于 Debian Linux 的开源虚拟化平台。本文将介绍如何在 PVE 环境中使用脚本安装 Olares。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

## 系统要求
请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：不少于 150GB 的可用磁盘空间，需使用 SSD 硬盘安装。
- 支持的系统版本：PVE 8.2.2

:::warning 注意
使用 HDD（机械硬盘）可能会导致安装失败。
:::

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

在 PVE 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{33,37}-->