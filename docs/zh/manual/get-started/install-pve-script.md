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
### 必要配置
- CPU：4 核及以上
- 内存：不少于 8 GB 可用内存
- 存储：不少于 150 GB 的可用磁盘空间，需使用 SSD 硬盘安装。
   ::: warning 必须使用 SSD
   请勿使用机械硬盘 (HDD)。如果未检测到 SSD，安装将失败。
   :::
- 支持的系统版本：PVE 8.2.2

:::info 版本兼容性
虽然以上版本已经过验证，但其他版本也可能正常运行 Olares。根据你的环境可能需要进行调整。如果你在这些平台安装时遇到任何问题，欢迎在 [GitHub](https://github.com/beclab/Olares/issues/new) 上提问。
:::

### 可选硬件
安装 Olares 不需要 GPU，但运行 AI 应用需要 GPU支持。
- GPU (仅支持 NVIDIA)：
  - 架构：Turing 架构或更新版本 (例如 GTX 16 系列、RTX 20 系列及以后)。
  - 验证：在 PVE 宿主机 Shell 中运行 `lspci | grep -i nvidia` 以确认显卡已被识别。
  - 配置：如需使用 GPU，必须配置显卡直通。详细步骤请见[在 PVE 中配置 GPU 直通)](/zh/manual/best-practices/install-olares-gpu-passthrough.md#在-pve-中配置-gpu-直通)。

## 安装 Olares

在 PVE 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{33,37}-->