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

  但如果数据库集群和内置应用通过[ Service Provider](../develop/advanced/provider.md) 开放了某些接口，社区应用可以通过[声明访问权限](../develop/package/manifest.md)来使用这些服务。

  在这种情况下，系统会在 `user-system` 命名空间下为这些资源提供网络代理，并对来自第三方应用的网络请求进行鉴权。

### 社区应用

社区应用是由第三方开发者创建和维护的应用，涵盖从生产力工具、娱乐应用到数据分析工具等多种用途。

社区应用的命名空间由两部分组成：应用名称和用户的[本地名称](olares-id.md#olares-id-结构)，例如：

```
n8n-alice
gitlab-client-bob
```

### 共享应用

共享应用是 Olares 平台中的一类特殊社区应用，为集群内所有用户提供共享的资源或服务。

共享应用的特点包括：

- **集中管理**：只有管理员可以安装、升级、暂停、恢复和卸载共享应用，并负责在集群内配置和托管对应的服务、资源及运行环境。
- **易于识别**：在 Olares 应用市场中，共享版应用通常带有“共享版”标识。
- **成员开箱即用**：管理员完成安装后，集群内所有成员即可访问该共享应用，无需额外授权或安装任何客户端。
- **统一访问地址，数据隔离**：所有共享应用遵循统一的 URL 访问规则 `https://<应用 ID>.<用户名>.<平台域名>`。每个成员通过自己的用户名访问同一共享应用，系统会根据用户名自动隔离各成员的数据，确保每位成员只能访问属于自己的数据。
- **灵活访问**：共享应用的访问方式取决于其形态。
  - **无界面的后端服务**：对于无图形界面的后端服务类共享应用（如 Ollama），服务会通过共享入口暴露标准 API，可被任何兼容的第三方客户端（如 LobeChat、Open WebUI）调用。成员安装客户端后，将其指向共享应用的 API 端点，该端点可在 Olares **设置** > **应用** > 应用名称 > **共享入口** 中获取。
  - **自带用户界面的完整应用**：对于自带完整用户界面和后端服务的共享版应用（如 ComfyUI 共享版、Dify 共享版），管理员安装后，会在启动台生成一个同名应用入口，集群成员可直接通过该入口访问。

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