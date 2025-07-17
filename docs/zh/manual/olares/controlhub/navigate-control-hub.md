---
outline: [2, 3]
description: 了解控制面板中的应用、资源、网络配置与中间件服务的。
---

# 了解 Control Hub 界面

本wen将带你逐一了解 Control Hub 中的各个功能模块及其使用方法。

## 浏览

**浏览** 部分将资源组织为两类命名空间：

- 用户项目：包括每个 Olares 用户的独立命名空间。
  - `user-space-*`：用户内置应用。
  - `user-system-*`：与用户相关的系统程序，包括 Olares 应用运行时组件、调度程序和跨应用交互代理。
- 系统项目：包括 Olares 集群的核心应用以及系统级服务程序。

:::info 提示
Olares 成员仅能访问自己的命名空间，而 Olares 管理员可以访问所有用户和系统命名空间。
:::

在第二列中，你可以查看命名空间内的所有资源类型。

### 工作负载

在 Olares 中，工作负载表示运行在集群上的应用，主要对应 Kubernetes 的三种资源类型：

- **部署**（Deployment）
  - Kubernetes 中最常见的工作负载类型。
  - 自动生成 `ReplicaSet` 调度并创建 Pod，实现水平扩展。

- **有状态副本集**（StatefulSet）
  - 部署有状态的 Pod，例如数据库、分布式文件存储或内存缓存的 Pod。
  - 每个 Pod 的状态可能不同，通常按顺序调度。

- **守护进程集**（DaemonSet）
  - 在每个节点上调度并运行一个 Pod。生成的 Pod 数通常等于节点数。
  - 用于节点特定的硬件操作。

### 工作负载详情

![alt text](/images/how-to/olares/controlhub/browse/02.jpg#bordered)

工作负载资源的信息包括：

- **详情**：资源的基本信息。
- **容器组**：Pod 的基本信息。
- **端口**：工作负载 Pod 容器暴露的所有端口集合。
- **环境变量**：在 Pod 模板中为工作负载定义的环境变量。
- **标签**：配置在 `workload` 元数据中，用于控制器管理协调和资源筛选。
- **注解**：功能类似于标签，但更灵活，用于控制器管理工作负载。
- **事件**：最近一小时内与工作负载相关的事件日志，通常显示 Pod 调度状态。

### 容器详情

Pod 的信息包括：

- **信息**：容器的基本信息。
- **容器**：容器的容器列表。
- **卷**：为容器配置的持久卷。
- **环境变量**：为容器定义的环境变量。
- **事件**：按时间顺序记录与容器相关的事件。

![containers](/images/how-to/olares/controlhub/browse/04.jpg#bordered)

### 保密字典

保密字典（Secrets）用于存储密码、凭据和关键配置等敏感数据。在 Kubernetes 中，这些数据默认以 Base64 编码。

![secrets](/images/how-to/olares/controlhub/browse/11.jpg#bordered)

展开**保密字典**可查看应用命名空间下的所有密文数据。

信息包括：

- **信息**：Secret 的基本信息，例如所属命名空间和创建时间。
- **数据**：Secret 的**数据键**和**数据值**。

:::tip 提示
**数据值**默认以 `Base64` 显示，可点击右上角的 **<i class="material-symbols-outlined">visibility</i>预览**按钮查看原文。
:::

### 配置字典

配置字典（ConfigMap）结构类似于**保密字典**，但内容以明文保存。

![configmaps](/images/how-to/olares/controlhub/browse/12.jpg#bordered)

展开**配置字典**可查看应用命名空间下的所有配置。

信息包括：

- **信息**：配置字典的基本信息，如命名空间和创建时间。
- **数据**：配置字典的**数据键**和**数据值**。

### 服务账户

**服务账户**（Service accounts）是 Kubernetes 的一种机制，用于验证集群容器应用并允许其访问 Kubernetes 管理的集群资源。

![Service accounts](/images/how-to/olares/controlhub/browse/13.jpg#bordered)

展开 **服务账户** 部分可查看应用命名空间下的所有服务账户。

信息包括：

- **信息**：服务账户的基本信息，如命名空间、创建时间等。
- **数据**：与服务账户关联的 **Secret** 的**数据键**和**数据值**。
- **Kubeconfig 设置**：服务账户自动生成的 Kubeconfig 配置，可下载用于应用或直接读取 `/var/run/secrets/kubernetes.io/serviceaccount/` 中的配置。

### 服务

**服务**（Service）通过网络服务将运行在单个或一组 Pods 上的应用暴露出来，并根据定义的选择器分发流量。

![service1](/images/how-to/olares/controlhub/browse/14.jpg#bordered)

服务的信息包括：

- **属性**：服务的基本信息，如命名空间、创建时间、选择器、虚拟 IP 等。
- **工作负载**：选择器选择的所有工作负载。
- **端口**：暴露的所有端口信息。
- **容器组**：服务选择的所有 Pods 及其状态。
- **标签**：服务的标签。
- **注解**：服务的注解。
- **事件**：与服务相关的事件。

## 命名空间

**命名空间** 提供基于用户的资源消耗和工作负载状态视图。

![namespace](/images/how-to/olares/controlhub/namespace/01.jpg#bordered)

### 用量排名

按命名空间组织系统资源消耗。

![namespace list](/images/how-to/olares/controlhub/namespace/02.jpg#bordered)

信息包括：

- **配额**：此命名空间的系统资源使用百分比。
- **绒球**：按资源消耗排序的所有 Pods，可通过关键字搜索。

### 资源

显示当前和历史资源利用率图表，可按用户筛选查看。

![resources](/images/how-to/olares/controlhub/namespace/04.jpg#bordered)

## 容器组

**容器组**提供全面的容器状态和资源使用视图。

信息包括：

- **Pod 列表**：Olares 中的所有 Pods。
- **资源**：显示容器的物理资源消耗。

![resources](/images/how-to/olares/controlhub/pods/04.jpg#bordered)

## 资源

包括软件和硬件相关资源。

### 网络

**网络策略**（Network policies）定义网络连接规则，基于命名空间的沙盒机制提供隔离。

#### 入站规则

显示允许进入命名空间的入站流量的规则列表。

#### 出站规则

显示允许离开命名空间的出站流量的规则列表。

### CRDs

列出 Olares 中所有基于 Kubernetes 的自定义资源声明。

![CRDs](/images/how-to/olares/controlhub/resources/02.jpg#bordered)

信息包括：

- **信息**：显示 CRD 的名称、组、作用域和创建时间。
- **自定义资源**：此 CRD 下的所有自定义资源。

## 中间件

管理员可在此管理中间件。

:::tip 提示
仅管理员可访问**中间件**页面。
:::

信息包括：

- **信息**：集群数据，如名称、命名空间、访问地址等。
- **数据库**：各应用使用的数据库概览。

![postgres](/images/how-to/olares/controlhub/middleware/01.jpg#bordered)