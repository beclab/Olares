---
description: 在 PVE 虚拟化平台上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活过程。
---
# 在 PVE 上使用脚本安装 Olares
Proxmox 虚拟环境（PVE）是一个基于 Debian Linux 的开源虚拟化平台。本文将介绍如何在 PVE 环境中使用脚本安装 Olares。

::: warning 不适用于生产环境
该部署方式当前仍有功能限制，建议仅用于开发或测试环境。
:::

<!--@include: ./reusables.md{44,51}-->

## 系统要求

### 必要配置

- **CPU**：4 核及以上
- **内存**：不少于 8 GB 可用内存
- **存储**：不少于 200 GB 的 SSD 可用磁盘空间。
   :::warning 必须使用 SSD
   使用机械硬盘 (HDD) 会导致安装失败。
   :::
- **支持的系统版本**：PVE 8.2.2


<!--@include: ./reusables.md{63,65}-->


### 可选硬件
<!--@include: ./gpu-requirements.md{5,}-->

:::tip 显卡直通
PVE 中如需使用 GPU，必须配置显卡直通。详细步骤见[在 PVE 中配置 GPU 直通](/zh/manual/best-practices/install-olares-gpu-passthrough.md#在-pve-中配置-gpu-直通)。
:::

## 安装 Olares

在 PVE 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,37}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->