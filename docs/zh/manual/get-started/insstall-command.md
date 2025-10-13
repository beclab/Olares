---
description: 在 Linux 系统 Ubuntu 和 Debian 上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活流程。
---
# 使用命令行安装 Olares
本文介绍如何在 Linux 上使用一行命令行脚本安装 Olares。

## 系统要求

请确保设备满足以下配置要求：

- CPU：4 核及以上
- 内存：不少于 8GB 可用内存
- 存储：不少于 150GB 的可用磁盘空间，需要使用SSD硬盘安装，使用HDD（机械硬盘）将会导致安装失败
- 支持的系统版本：
  - Ubuntu 22.04-25.04 LTS
  - Debian 12, 13

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台上安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

## 安装 Olares

在 Linux 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,32}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{34,38}-->