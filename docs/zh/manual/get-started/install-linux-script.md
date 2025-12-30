---
description: 在 Linux 系统 Ubuntu 和 Debian 上安装配置 Olares 的完整步骤，包括系统要求、安装命令和激活流程。
---
# 使用命令行安装 Olares
本文介绍如何在 Linux 上使用一行命令行脚本安装 Olares。
:::info 安装遇到问题？
如果你在安装过程中遇到问题，欢迎[提交 GitHub Issue](https://github.com/beclab/Olares/issues/new)。提交时请尽量附上以下信息：

- 你使用的平台或环境（例如 Ubuntu、Docker、WSL 等）。
- 安装方式（脚本安装或 Docker 镜像）。
- 详细的错误信息（包括日志、报错信息或截图）。
:::

## 系统要求
### 必要配置
- **宿主机要求**:
  - **CPU**：4 核及以上。
  - **内存**：至少 8 GB 可用内存。
  - **存储**：至少 150 GB 可用磁盘空间。
   ::: warning 必须使用 SSD
   请勿使用机械硬盘 (HDD)。如果未检测到 SSD，安装将失败。
   :::
  - **支持的系统**：
    - Ubuntu 22.04-25.04 LTS
    - Debian 12 或 13

### 可选硬件

安装 Olares 不需要 GPU，但运行 AI 应用需要 GPU支持。

- **GPU (仅支持 NVIDIA)**：
  - **架构**：Turing 架构或更新版本 (例如 GTX 16 系列、RTX 20 系列)。
  - **验证方法**：运行 `lspci | grep -i nvidia` 并查看 [兼容 GPU 列表](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus)。

## 安装 Olares

在 Linux 命令行中，执行以下命令：

<!--@include: ./reusables.md{4,32}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{34,38}-->