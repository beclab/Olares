---
outline: [2, 3]
description: Olares 应用系统的核心概念，包括应用标识符、类型分类和权限体系。阐述系统应用、社区应用和集群范围应用的特性及依赖关系。
---

# 应用

本文介绍 Olares 中应用标识符、类型、权限以及与应用市场集成相关的核心概念。

## 应用标识符

在 Olares 中，每个应用都有两个标识符：应用名称和应用 ID。

### 应用名称

应用名称由 Indexer 分配。Olares 团队维护的 Indexer 仓库是 [apps](https://github.com/beclab/apps)。应用在该仓库中的目录名即为其应用名称。

### 应用 ID

应用 ID 是应用名 MD5 哈希值的前八个字符。例如，如果应用名称为“hello”，则其应用 ID 为“b1946ac9”。

应用对应的端点（Endpoint）会使用该应用 ID。

## 应用类型

Olares 包含多种类型的应用。你可以通过控制面板查看系统的各类应用，并通过命名空间来识别具体的应用类型。

### 系统应用

系统应用包括 Kubernetes、Kubesphere、Olares 组件和必要的硬件驱动。系统级命名空间包括：

```
os-system
kubesphere-monitoring-federated
kubesphere-controls-system
kubesphere-system
kubesphere-monitoring-system
kubekey-system
default
kube-system
kube-public
kube-node-lease
gpu-system
```
其中，`os-system` 是 Olares 开发的组件。集群级的应用以及系统提供的各种数据库中间件都安装在这个命名空间下。

### 用户级系统应用

Olares 支持多用户，并为管理员和普通成员用户提供两个不同的系统应用命名空间：

- **user-space-{本地名称}**

  `user-space` 命名空间用于安装用户日常交互的系统应用，包括：
    - 文件管理器
    - 设置
    - 控制面板
    - 仪表盘
    - 应用市场
    - Profile 
    - Vault

  这些应用之间存在相互调用，同时调用系统底层接口（如 Kubernetes 的 `api-server` 接口）。为了确保系统安全，Olares 将它们统一部署在独立的 `user-space` 命名空间中，通过沙盒机制隔离，防止恶意程序的攻击和非法访问。

- **user-system-{本地名称}**

  系统应用和用户的内置应用通常不允许第三方应用直接访问。

  但如果数据库集群和内置应用通过[ Service Provider](../develop/advanced/provider.md) 开放了某些接口，社区应用可以通过[声明访问权限](../develop/package/manifest.md#sysdata)来使用这些服务。

  在这种情况下，系统会在 `user-system` 命名空间下为这些资源提供网络代理，并对来自第三方应用的网络请求进行鉴权。

### 社区应用

社区应用是由第三方开发者创建和维护的应用，涵盖从生产力工具、娱乐应用到数据分析工具等多种用途。

社区应用的命名空间由两部分组成：应用名称和用户的[本地名称](olares-id.md#olares-id-的组成)，例如：

```
n8n-alice
gitlab-client-bob
```

### 共享应用

**共享应用**是 Olares 平台中的一类特殊社区应用，旨在为 Olares 集群内的所有用户提供统一的、共享的资源或服务。

共享应用的特点包括：

* **集中管理**：只有管理员账户才安装共享版应用的核心服务。管理员负责在 Olares 集群内**安装、配置和托管**应用的服务、资源以及运行环境。
* **易于识别**：在 Olares 应用市场中，共享版应用通常带有 "Shared" 标识以便用户区分。
* **灵活访问**：访问共享版应用的方式取决于共享应用本身的形态：
    * **无界面的后端服务**: 对于通常在后端提供服务没有直接用户界面的共享应用（如 Ollama），用户通常需要通过安装一个**授权应用**作为访问入口。例如，可以通过 Open WebUI 或 LobeChat 访问 Ollama 服务。
    * **自带用户界面的完整应用**: 对于共享版应用本身就包含完整用户界面和后端服务的（例如，ComfyUI 共享版 或 Dify 共享版），管理员和集群中的其他用户都可通过直接安装该共享版应用本身获取服务的访问入口。

### 授权应用

授权应用是指具有 Olares 中特定共享应用访问权限的应用。它们通常为用户提供可交互界面，以便访问被授权共享应用的 API 或服务。

例如，Open WebUI、LobeChat 和 n8n 是 Ollama 的授权应用。Dify Shared 是它自身的授权应用。

### 依赖项
依赖项是某些应用正常运行所必需的前置应用。安装带有依赖项的应用前，用户必须确保集群中已安装所有必需的依赖项。

## Service Provider

Service Provider 机制使社区应用能够与系统应用、其他社区应用的服务进行交互。

![Service Provider](/images/overview/olares/image3.jpeg)

该机制包含三个步骤：

1. Provider 声明：开发者必须[将其应用声明为特定服务接口的 Provider](../../developer/develop/advanced/provider#申明-Provider)。
   系统包含内置的 Provider。

2. 权限请求：需要使用 Service 接口的应用必须明确[申请 Provider 的权限](../../developer/develop/advanced/provider#申请-Provider-的访问权限)。

3. 请求处理：调用时，`user-system` 下的 `system-server` 服务作为代理，处理传入请求并执行必要的权限验证。

## 了解更多

- 用户

  [管理应用](../../manual/olares/market/market.md)<br>

- 开发者

  [在 Olares 上开发应用程序](../develop/index.md)<br>

