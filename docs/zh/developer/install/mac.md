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

<!--@include: ./reusables.md{36,41}-->

## 系统要求
Mac 设备需满足以下条件：
- 处理器架构：X86-64 或 ARM64
- 内存：可用内存 8 GB 及以上
- 存储空间：可用磁盘空间 90 GB 及以上
- MacOS 版本：Monterey（12）及以上

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

<!--@include: ./reusables.md{4,18}-->
## 配置 Wizard
在安装过程结束时，你需要提供下列信息：
1. 检查 Mac 的 IP 地址（例如，`192.168.x.x`）。

   如果自动获取的 IP 地址正确，请按 `Y` 确认。如果需要修改，请按 `R` 并输入正确的地址。
   ::: tip 查看 IP 地址
   要查看 Mac 的 IP 地址，可以使用两种方式:
   - 使用图形界面：打开**系统设置**（或**系统偏好设置**）> **网络**，在当前活动的网络连接中查看详细信息。
   - 使用命令行：打开终端窗口，Wi-Fi 网络输入 `ipconfig getifaddr en0`，有线网络输入 `ipconfig getifaddr en1`。
   :::

2. 如果你的 Olares ID 为 `alice123@olares.cn`，输入 `alice123` 即可。

   ![输入 Olares ID](/images/zh/manual/get-started/enter-olares-id.png)

<!--@include: ./reusables.md{26,28}-->

<!--@include: ./activate-olares.md-->

<!--@include: ./log-in-to-olares.md-->

<!--@include: ./reusables.md{30,34}-->