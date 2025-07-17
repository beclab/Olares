---
outline: [2, 3]
description: Olares 部署的核心架构，涵盖系统原生层、容器编排曾和容器化层。深入解析 Olares 各层级之间的技术交互。
---
# Olares 安装概述 

本文档从宏观的角度介绍了 Olares 的安装流程，重点介绍其整体架构和核心组件，旨在为系统管理员和开发者提供 Olares 运行原理及安装方法的基本了解。

## Olares 安装的三层结构
Olares 的安装流程分为三个关键层级：

- **原生层**：处理 Linux 系统配置与基础环境依赖安装。
- **容器编排层**：部署 Kubernetes 集群，实现服务的自动化管理与扩展。
- **容器化层**：启动容器化的系统核心服务和用户应用，提供最终运行时环境。

安装过程由命令行工具 `olares-cli` 管理。该工具负责协调所有组件的安装、配置以及生命周期管理。

![Install arch](/images/developer/install/olares-install.png)

::: tip 提示
要了解安装的详细过程，请参考 [Olares 安装流程详解](installation-process.md)文档。
:::

## 原生层
Olares 的安装从原生层开始，确保底层 Linux 环境支持分布式存储、容器运行时和 Kubernetes 集群管理。

该层的配置涵盖核心 Linux 环境设置、文件系统初始化、容器运行时安装以及关键系统服务部署。

### 环境配置

安装首先配置基本的 Linux 安装环境，如配置域名解析系统（DNS）、安全外壳远程访问协议（SSH）、网络时间协议（NTP）等服务，以确保时间同步与远程管理。

同时，也会安装必要依赖，如 GNU 编译器集合（GCC）、网络工具（net-tools）等系统工具，确保运行时环境。

### 文件系统配置

文件系统（rootfs）用于存取系统核心组件与用户数据。根据部署需求，Olares 支持以下两种文件系统：

- **LocalFS**（默认）：使用本地 Linux 磁盘进行存储，适用于无需网络共享且需要高数据吞吐量的单节点部署。

- **JuiceFS**：为多节点集群提供分布式文件系统。文件数据可以存储在本地安装的 MinIO 实例中，也可以存储在 Amazon S3 这一类远程存储桶中。该配置支持不同存储节点共享统一的存储视图。

:::tip 启用 JuiceFS
JuiceFS 和 MinIO 默认不会安装。如需启用，需在安装前设置必要的[环境变量](environment-variables.md#juicefs)或用 [olares-cli 命令](cli/prepare#选项)配合 JuiceFS 相应参数安装。
:::

### 容器运行时：containerd
Olares 使用轻量级容器运行时 containerd 进行容器化部署，其主要功能包括：
- **容器镜像管理**：
  - 从内容分发网络（CDN）下载已打包的容器镜像
  - 在“准备”阶段将下载的镜像导入至 containerd
  - 在“安装”环阶段以容器进程的方式启动镜像
- **容器生命周期管理**：启动、停止、重启和监控容器化应用服务。

::: tip 兼容性问题
如果你的机器之前安装过 containerd（比如通过 Docker 安装），可能会与 Olares 安装的 containerd 有兼容性问题。请在安装 Olares 前卸载现有的 containerd。
:::

### 系统守护进程：olaresd
olaresd 是系统守护进程，在后台运行，提供以下关键管理功能：
- **自动配置更新**：当系统发生变化（如 IP 地址变更）时，自动检测并调整相关配置项。
- **远程系统管理**：根据来自 LarePass 客户端或 `olares-cli` 的命令执行远程系统操作，例如 Olares 安装和激活。

### CUDA 支持
为了让本地 AI 模型和应用启用 GPU 加速，Olares 支持通过 [`olares-cli`](./cli/gpu.md) 自动检测并安装 CUDA 工具包和相关驱动程序。

## 容器编排层
容器编排层通过 Kubernetes 将系统组件集成到高效的运行时环境中。

### Kubernetes 的角色
Kubernetes 是容器编排层的核心，负责实现多组件服务的自动部署、运行、扩展和管理。

与 Docker Compose 或 Docker Swarm 等工具相比，Kubernetes 具有以下优势：
- 高可扩展性和生产级可靠性：适用于大规模集群部署和关键任务环境。
- 丰富的社区支持：拥有活跃的社区和丰富的生态系统，可通过 Helm Charts、Operators 和自定义资源定义（CRDs）集成各种应用程序。

### Olares 支持的 Kubernetes 版本
Olares 支持以下 Kubernetes 部署方式：
- **K3s**（默认）：轻量级 Kubernetes 发行版，可优化本地硬件上的资源利用效率。
- **Kubernetes**：完整功能的 Kubernetes 发行版，适用于高级或自定义部署需求。
- **minikube**（仅限 macOS）：设置单节点 Kubernetes 集群的工具，确保一致的功能和用户体验。

## 容器化层

在容器化层，Olares 各组件和应用协同工作提供系统完整功能。所有 Olares 组件和用户应用都在容器中运行，整个生命周期由 Kubernetes 管理。这确保了系统的高效性、稳定性和可扩展性。

在安装并激活 Olares 后，你就可以通过**控制面板**应用的图形界面查看正在运行的容器。

![控制面板中查看运行的容器](/images/developer/install/running-pods.png#bordered){width=90%}

## 了解更多

- [Olares 安装流程详解](installation-process.md)
- [Olares Home 概述](olares-home.md)
- [`olares-cli` 命令行参考](../install/cli/olares-cli.md)
- [Olares 环境变量](environment-variables.md)