---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/manual/best-practices/connect-ai-apps
outline: [2, 3]
description: 了解 AI 客户端如何通过 Olares Router 连接本地或远程模型。
head:
  - - meta
    - name: keywords
      content: Olares, AI 客户端, 模型服务, LarePass VPN, Router, default-chat, Base URL, API key
---

# AI 应用如何连接 Olares 模型

<VersionRouteSelect />

许多 AI 应用只提供交互界面或工作流，模型能力由独立的服务提供。LobeHub、OpenCode 和 Claude Code 都属于这类应用。

连接 AI 应用，就是告诉应用使用哪种 API、将请求发送到哪里、调用哪个模型，以及如何完成鉴权。

## 客户端、Router 与模型服务

Olares 上的 AI 连接包含三个部分：

- **客户端应用**：发送模型请求的聊天界面、编程工具或工作流应用。
- **Router**：Olares AI 能力的统一入口，负责验证调用方身份并转发请求。
- **模型服务**：本地模型应用中运行的模型，或 Router 已接入的远程模型服务商。

客户端统一连接 Router。Router 将请求发送给选定的模型，再把结果返回给客户端。本地模型和远程模型共用同一个 Router 服务，客户端使用的 Base URL 则取决于请求来源。

本地模型应用会在 Olares 内提供共享入口，供 Router 调用模型引擎。这个地址属于 Router 与模型后端之间的连接。客户端应使用 Router 提供的 Base URL。

## 请求来源决定连接方式

选择连接方式时，应以发送 API 请求的组件及其网络位置为准。

- **服务端请求**：安装在 Olares 上的应用通常由应用自身的服务进程发送请求，请求始终位于 Olares 内部网络。
- **客户端直连**：桌面应用、命令行工具、IDE 扩展或浏览器客户端从电脑发送请求，通过局域网或 LarePass VPN 访问 Olares。

部分网页应用可以在服务端请求和浏览器请求之间切换。例如，启用 **Client Request Mode** 一类的设置后，请求来源会从 Olares 内的应用转移到浏览器。遇到这类选项时，应按照对应应用的教程选择请求模式。

| 请求来源 | Router 连接选项 | 访问方式 |
| --- | --- | --- |
| Olares 内的应用进程 | **Apps in Olares** | 平台注入应用身份 |
| 与 Olares 位于同一局域网的电脑 | **Devices in LAN** | 局域网直连并使用 Router API key |
| 位于局域网外的电脑 | **Remote** | 通过 LarePass VPN 接入并使用 Router API key |

### 从局域网外连接

同一局域网内的电脑可以直接访问 Router。电脑位于其他网络时，使用 **Remote** 中的连接信息，并先通过 LarePass VPN 接入 Olares 私有网络。

VPN 和 API key 解决的是连接中的不同问题。LarePass VPN 在电脑与 Olares 之间建立加密的网络通道，Router API key 则用于识别外部客户端，并授权其使用模型能力。因此，远程客户端需要同时使用两者。

## 连接参数

各 AI 应用使用的字段名称有所差异，模型连接通常包含以下参数：

| 参数 | 作用 |
| --- | --- |
| Provider 或 API 格式 | 定义客户端和 Router 之间的请求格式 |
| Base URL | 指定 Router 的访问地址和 API 路径 |
| 模型名称或 Model ID | 指定要调用的模型或路由规则 |
| API key | 验证 Olares 外部客户端的身份 |

### Provider 和 API 格式

“Provider”在客户端和 Router 中表示不同的对象：

- 在客户端中，Provider 通常是负责组织请求格式的适配器。例如，OpenAI 兼容的 Provider 可以调用运行在 Olares 本地的模型。
- 在 Router 中，Provider 是 Router 转发请求的后端，可以是本地模型应用，也可以是云服务商。

请使用应用教程中指定的 Provider 或适配器。即使调用同一个模型，不同 API 格式使用的请求结构和路径也可能不同。

### Base URL

Base URL 告诉客户端将请求发送到哪里。Router 分别为 **Apps in Olares**、**Devices in LAN** 和 **Remote** 提供地址，因为三类请求通过不同的网络路径到达 Router。

应根据请求来源复制完整地址，包括界面中显示的 `/v1` 等路径。部分客户端会自动追加 API 路径，因此具体教程可能要求移除或修改该后缀。

### API key

Olares 内部应用由平台完成身份验证。电脑上的客户端使用在 Router 中创建的 API key，包括来自同一局域网的请求。

Olares 内的应用遇到必填的 API key 字段时，请填写该应用教程中给出的占位值。

## 模型选择

模型名称决定 Router 如何处理请求：

- **`default-chat`** 将请求转发给 Router **Default models** 页面中设置的聊天模型。之后更换默认模型时，使用该名称的应用会自动跟随。
- **完整模型名称**将请求发送给一个指定模型。需要让应用始终使用同一模型时，请使用这种方式。

`default-chat` 是路由名称，模型列表 API 则返回具体模型。如果客户端根据该 API 生成模型列表，需要手动添加 `default-chat`。只允许选择列表内模型的客户端应使用完整模型名称。

## 应用配置教程

以下教程介绍各客户端所需的 Provider、URL 格式和配置字段：

- [使用 LobeHub 构建本地 AI 助手](/zh/use-cases/lobechat.md)
- [将 OpenCode 设置为 AI 编程助手](/zh/use-cases/opencode.md)
- [使用 Claude Code 编写代码](/zh/use-cases/claude-code.md)

## 了解更多

- [使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router.md)
- [通过 LarePass VPN 连接 Olares 私有网络](../larepass/private-network.md)
