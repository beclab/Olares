---
description: 在 Mac 系统上安装配置 Olares 的完整步骤，包括环境准备、安装过程和系统激活。
---
# 在 Mac 上使用脚本安装 Olares
本文介绍如何在 Mac 上使用脚本安装 Olares。

::: warning 不适用于生产环境
Mac 版 Olares 目前存在以下限制：
- 不支持分布式存储
- 无法添加本地节点

建议仅用于开发或测试环境。
:::
:::info 安装遇到问题？
如果你在安装过程中遇到问题，欢迎[提交 GitHub Issue](https://github.com/beclab/Olares/issues/new)。提交时请尽量附上以下信息：

- 你使用的平台或环境（例如 Ubuntu、Docker、WSL 等）。
- 安装方式（脚本安装或 Docker 镜像）。
- 详细的错误信息（包括日志、报错信息或截图）。
:::

## 系统要求
Mac 设备需满足以下条件：
- 处理器架构：X86-64 或 ARM64。
- 内存：至少 8 GB 可用内存。
- 存储空间：至少 150 GB 可用磁盘空间。
   ::: warning 必须使用 SSD
   请勿使用机械硬盘 (HDD)。如果未检测到 SSD，安装将失败。
   :::
- MacOS 版本：Monterey（12）及以上。
::: info GPU 限制
Olares 的 GPU 加速功能目前仅支持 NVIDIA GPU。因此，在 macOS 设备上无法为 AI 工作负载管理或使用 GPU 资源。
:::
## 前置准备
请确保已安装以下软件：
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [MiniKube](https://minikube.sigs.k8s.io/docs/start/?arch=%2Fmacos%2Farm64%2Fstable%2Fhomebrew)
  ::: tip
  推荐通过 `homebrew` 安装 minikube。
  :::

## 配置系统环境
1. 打开 Docker Desktop，进入 **Settings** > **Resources**，按以下要求配置资源：
    - **CPU limit**：至少设置为 4 核
    - **Memory limit**：至少设置为 9 GB
    - **Virtual disk limit**：至少设置为 80 GB

   ![更新资源配置示例](/images/manual/get-started/docker-resources-settings.png#bordered)
2. 点击 **Apply & restart** 使配置生效。
## 安装 Olares
在终端中运行以下命令：

<!--@include: ./reusables.md{4,21}-->
## 配置 Wizard
在安装过程结束时，你需要提供下列信息：
1. 检查 Mac 的 IP 地址（例如，`192.168.x.x`）。

   如果自动获取的 IP 地址正确，请按 `Y` 确认。如果需要修改，请按 `R` 并输入正确的地址。
   ::: tip 查看 IP 地址
   要查看 Mac 的 IP 地址，可以使用两种方式:
   - 使用图形界面：打开**系统设置**（或**系统偏好设置**）> **网络**，在当前活动的网络连接中查看详细信息。
   - 使用命令行：打开终端窗口，Wi-Fi 网络输入 `ipconfig getifaddr en0`，有线网络输入 `ipconfig getifaddr en1`。
   :::

2. <!--@include: ./reusables.md{25,31}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{33,37}-->