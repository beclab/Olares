---
outline: [2, 3]
description: Olares 网络架构设计说明，阐述应用入口类型、本地访问机制、端点配置和内部网络安全策略的基本原理。
---
# 网络

Olares 为用户提供无障碍且安全灵活的网络解决方案。本文档介绍与应用访问相关的核心概念。

## 入口

每个 Olares 应用可以配置一个或多个入口来接入外部访问。入口分为三种类型：

### 公开入口

- 托管博客、社交媒体等需要外部访问的服务
- 无需认证即可访问
- 通过 Cloudflare 提供基础安全防护

### 私有入口

- 专门为个人、家庭或团队提供服务
- 适用于阅读器、娱乐、生产力工具、桌面应用等
- 需要通过[认证](account.md#多因素认证mfa)才能访问

### 内部入口
- 功能与私有入口类似。
- 通过 LarePass VPN 访问应用时，无需身份验证。

## 端点

端点是用户与应用交互的访问地址或接入点（access point）。简单来说，就是在浏览器地址栏中输入的 URL，用于访问特定的 Olares 应用或其功能。

典型的 Olares 应用端点格式如下：

    https://{routeID}.{domain}

例如：`https://vault.alice123.olares.cn`，其中：

- `vault` 是系统应用的路由 ID
- `alice123.olares.cn` 是由 Olares ID `alice123@olares.cn` 演变而来

## 路由 ID

路由 ID 是用于标识特定应用或应用入口的唯一标识符。系统会根据以下规则自动生成默认路由 ID：

- 系统应用
  - 使用预设的易记路由 ID
  - 示例：`desktop`（桌面）、`market`（应用市场）
- 社区应用
  - 使用 8 位随机字符串 + 入口索引（从 0 开始）
  - 示例：对于路由 ID 为 `92d76a13` 且有两个入口的应用，第一个入口为`92d76a130`，访问 URL 为 `92d76a130.alice.olares.cn`

::: tip 注意

- 应用地址包含 Olares ID
- 入口索引指的是入口在 [`OlaresManifest.yaml`](../develop/package/manifest.md) 中定义的多个入口中的位置。
  :::

## Olares 内部网络

Olares 在网关架构中采用多层代理路由设计。流量经过多个层级分发：

`集群` -> `用户` -> `应用` -> `服务组件`

![alt text](/images/overview/olares/image4.jpeg)

在应用内部，Olares 实现了多层安全防护。

- **命名空间隔离**

  - 每个应用运行在独立命名空间中
  - 所有资源限定在命名空间内
  - 应用无法将 "`ClusterRole`" 连接到 "`ServiceAccount`"
  - 禁止跨命名空间访问资源

- **网络策略控制**
  - 每个命名空间有专属网络策略
  - 入站网络请求仅限于用户的集群应用和系统应用
  - 用户级网络隔离：
    - 不同用户之间的应用相互隔离
    - 统一用户的第三方应用之间相互隔离
- **Pod 限制**
  - Pod 不能使用 `hostNetwork` 服务或 `NodePort` 服务
  - 流量访问必须通过声明的入口服务和系统提供的入口代理
  - 声明为入口的 Pod 将被强制加入 Envoy 的沙箱 Sidecar，以对入站流量进行认证和授权

## 了解更多

- [为应用设置自定义域名](../../manual/olares/settings/custom-app-domain.md#自定义域名)
- [通过专用网络访问 Olares 应用](../../manual/larepass/private-network.md)
