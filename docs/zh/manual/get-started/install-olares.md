---
description: 比较自托管 Linux 硬件、Olares One 和 NVIDIA DGX Spark 的设置路径与安装要求。
outline: [2,4]
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, 系统要求, 安装方式, Linux, Ubuntu, Debian, DGX Spark
---

# 安装 Olares

根据你使用的设备选择设置路径：

- **自有设备**：[在 Linux 上安装 Olares](#linux)。
- **Olares One**：[设置或重装 Olares One](#olares-one)。
- **NVIDIA DGX Spark**：[在 DGX Spark 上安装 Olares](#dgx-spark)。

## Linux

使用以下方式在你自己的兼容硬件上安装 Olares。生产环境推荐使用 Linux。

### 系统要求

:::warning 必须使用 SSD
使用机械硬盘（HDD）会导致安装失败。
:::

| 项目 | 要求 |
| --- | --- |
| CPU | 4 核及以上 |
| 内存 | 8 GB 及以上 |
| 存储 | 150 GB 及以上 SSD 存储空间 |

### 可选 GPU

安装 Olares 无需 GPU，但大多数 AI 应用需要 GPU 才能运行。目前仅支持 NVIDIA GPU。

| 项目 | 要求 |
| --- | --- |
| 架构 | Turing 或更新架构，包括 GTX 16xx，以及 RTX 20xx、30xx、40xx 和 50xx 系列 |
| 显存 | 建议 8 GB 及以上 |

<!--@include: ./gpu-requirements.md#gpu-compatibility-check-->

### 安装方式

| 安装方式 | 适用场景 |
| --- | --- |
| [**ISO 镜像**](install-linux-iso.md)**（推荐）** | 在使用 Intel 或 AMD x86-64 处理器的物理机上全新安装 |
| [**一行命令**](install-linux-script.md) | 在已有 Ubuntu 22.04–25.04 或 Debian 12/13 系统上安装 |
| [**Docker Compose**](install-linux-docker.md) | 在 Ubuntu 22.04–25.04 或 Debian 12/13 上以容器方式安装 |

## Olares One

Olares One 提供专用的设置和恢复流程，以保留其硬件专属功能。

:::warning 请使用 Olares One 专用指南
不要使用通用 Linux ISO 镜像或一行命令安装 Olares One。否则，设备可能被识别为通用硬件，导致部分 Olares One 专属功能无法使用。
:::

| 任务 | 推荐指南 |
| --- | --- |
| 首次设置 Olares One | [首次启动](/zh/one/first-boot) |
| 重装或恢复 Olares OS | [Olares One 专用 ISO](/zh/one/create-bootable-usb) |

## DGX Spark

NVIDIA DGX Spark 是一款紧凑型 AI 开发平台，配备高性能 GPU。Olares 经过优化，可充分发挥 DGX Spark 硬件的性能。

### 安装方式

| 安装方式 | 适用场景 |
| --- | --- |
| [**一行命令**](install-spark-script.md)**（推荐）** | 从现有 DGX OS 安装，需至少 150 GB 可用 SSD 存储空间 |
| [**ISO 镜像**](install-spark-iso.md) | 通过启动 U 盘全新安装 |
