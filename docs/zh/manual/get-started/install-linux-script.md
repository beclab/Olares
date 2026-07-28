---
outline: [2, 3]
description: 在 Linux 系统 Ubuntu 和 Debian 上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活流程。
---
# 在 Linux 设备上使用命令行安装 Olares
本文介绍如何在 Linux 上使用一行命令行脚本安装 Olares。

<!--@include: ./reusables.md#installation-troubleshooting-tip-->

## 系统要求

### 必要配置
- **CPU**：4 核及以上。
- **内存**：至少 8 GB 可用内存。
- **存储**：至少 150 GB 的可用 SSD 磁盘空间。
   :::warning 必须使用 SSD
   使用机械硬盘 (HDD) 会导致安装失败。
   :::
- **支持的系统**：
  - Ubuntu 22.04-25.04 LTS
  - Debian 12 或 13

<!--@include: ./reusables.md#version-compatibility-->


### 可选硬件

<!--@include: ./gpu-requirements.md#gpu-requirements-->

## 安装 Olares

在 Linux 命令行中，执行以下命令：

<!--@include: ./reusables.md#install-script-command-->

<!--@include: ./reusables.md#root-password-tip-->

<!--@include: ./reusables.md#installation-error-tip-->

<!--@include: ./reusables.md#prepare-wizard-heading-->

<!--@include: ./reusables.md#prepare-wizard-details-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md#protect-olares-id-->