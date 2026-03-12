---
outline: [2, 3]
description: 在 Linux 系统 Ubuntu 和 Debian 上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活流程。
---
# 使用命令行安装 Olares
本文介绍如何在 Linux 上使用一行命令行脚本安装 Olares。

<!--@include: ./reusables.md{44,51}-->

## 系统要求

### 必要配置
- **CPU**：4 核及以上。
- **内存**：至少 8 GB 可用内存。
- **存储**：至少 150 GB 可用磁盘空间。
  ::: warning 必须使用 SSD
  请勿使用机械硬盘 (HDD)。如果未检测到 SSD，安装将失败。
  :::
- **支持的系统**：
  - Ubuntu 22.04-25.04 LTS
  - Debian 12 或 13

<!--@include: ./reusables.md{63,65}-->


### 可选硬件

<!--@include: ./gpu-requirements.md{5,}-->

## 安装 Olares

在 Linux 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,36}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{38,42}-->