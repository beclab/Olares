---
description: 在 PVE 虚拟化平台上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活过程。
---
# 在 PVE 上安装 Olares
Proxmox 虚拟环境（PVE）是一个基于 Debian Linux 的开源虚拟化平台。本文将介绍如何在 PVE 环境中安装 Olares。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{36,41}-->

## 系统要求
请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：建议使用 SSD，且可用磁盘空间不少于 64GB
- 支持的系统版本：PVE 8.2.2

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

在 PVE 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{30,34}-->