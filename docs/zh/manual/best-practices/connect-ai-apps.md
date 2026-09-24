---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/manual/best-practices/connect-ai-apps
outline: [2, 3]
description: 将 AI 客户端连接到 Olares Router，获取正确的 Base URL，选择模型名称，配置凭证并测试连接。
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI 应用, default-chat, OpenAI 兼容 API, Base URL, API 密钥
---

# 通过 Olares Router 连接 AI 应用

<VersionRouteSelect />

需要让 LobeHub、OpenCode 等 AI 客户端通过 Olares Router 调用模型时，请使用本指南。这里介绍大多数客户端都需要的连接信息，包括 API 格式、Base URL、模型名称和 API 密钥。

本指南适用于 Olares 1.12.7 及以上版本，以 Qwen3.8-27B (llama.cpp) 为聊天模型。具体客户端中的字段和操作，请参阅对应的[应用教程](#应用教程)。

:::info 想先了解 Router？
Router 的架构、能力分类、鉴权机制和命名规则，请参阅[使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router.md)。本页只介绍如何连接客户端。
:::

## 开始之前

1. 从应用市场安装 AI 客户端和 Qwen3.8-27B (llama.cpp)。
2. 从启动台打开 Router。在 **LLM** 页面确认 Qwen3.8-27B (llama.cpp) 显示 **Callable**。如果模型不可用，请查看状态下方的原因。
3. 在 **Default models** 页面，将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型。

设置默认模型只是指定 `default-chat` 请求的目标，不会启动已停止的模型。

## 选择 API 格式

客户端中的 **Provider** 或 **Engine** 决定它使用哪种 API 格式。请选择 Router 和客户端都支持的格式。

- 使用 OpenAI 兼容接口时，查找 **Custom Provider**、**Custom Endpoint**、**OpenAI** 或 **OpenAI-Compatible**。
- 只有应用教程或 Router 连接示例使用 Ollama API 时，才选择 **Ollama**。模型通过 Ollama 运行，并不意味着所有客户端都必须使用 Ollama 格式。
- 使用其他 API 或工具专用集成时，请按照对应应用教程操作。它可能需要工具应用自身的端点，而不是 Router 地址。

云端提供商的凭证保存在 Router 中。客户端选择 **OpenAI** 来使用其 API 格式，并不代表需要在客户端中填写 OpenAI API 密钥。

## 获取 Router Base URL

1. 在 Router 中打开对应的能力页面，找到模型，点击所在行的 **View connection example** 图标。Qwen3.8-27B 位于 **LLM** 页面。

   ![查看 Qwen3.8-27B 的连接示例](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. 在 **How to call this model** 窗口中，根据客户端的位置选择标签页：

   | 客户端位置 | 标签页 |
   | --- | --- |
   | 安装在 Olares 内的应用 | **Apps in Olares** |
   | 同一局域网内的电脑或其他设备 | **Devices in LAN** |
   | 从局域网外连接的设备 | **Remote** |

   ![Olares 内应用的 Router 连接信息](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. 复制所选标签页中的 **Base URL**。请使用自己设备上显示的地址，截图中的地址仅为示例。

保留客户端 API 格式需要的路径。OpenAI 兼容客户端通常需要末尾的 `/v1`。如果客户端会自动追加 `/v1`，则填写 Router 根地址。例如，Claude Code 的 `ANTHROPIC_BASE_URL` 使用根地址，由 SDK 追加 `/v1/messages`。

## 选择模型名称

通用聊天和 Agent 客户端使用 `default-chat`。请求会转发到 Router 中选定的默认聊天模型，以后更换默认模型时，无需修改使用该名称的客户端。

模型列表 API 不会返回 `default-chat`。如果客户端支持自定义模型 ID，请手动添加。如果客户端只能选择 API 返回的模型，请从 **How to call this model** 复制带提供商前缀的完整模型名称。本指南中 Qwen 模型的完整名称为 `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`。

非聊天能力应使用对应的模型或默认系统名称。例如，设置默认搜索提供商后，搜索可以使用 `default-search`。`default-chat` 不能代替嵌入、语音或搜索模型。

## 配置 API 密钥

请求来源不同，需要的凭证也不同：

| 调用方 | 如何填写 |
| --- | --- |
| Olares 内的应用 | 留空。如果客户端要求必填，可填写 `olares` 等占位值。 |
| 已登录 Olares、通过平台发起请求的用户 | 无需额外填写 Router API 密钥。 |
| 局域网或互联网中的外部客户端 | 在 Router 的 **API keys** 页面创建密钥并填入客户端。仅连接 VPN 不会提供 Olares 用户身份。 |

云端提供商的密钥应保存在 Router 的提供商配置中，而不是填写到通过 Router 连接的客户端中。

## 查看实际上下文大小 {#check-context-window}

Hermes 等客户端需要手动填写上下文大小。请在 Router 中查看当前模型的配置：

<!--@include: ../../reusables/ai-service-connections.md#model-context-window-->

## 配置并测试客户端

打开客户端的提供商或模型设置，填写以下信息：

| 客户端设置 | 填写内容 |
| --- | --- |
| 提供商或 API 格式 | 在[选择 API 格式](#选择-api-格式)中确定的格式 |
| Base URL | 与客户端位置匹配的 Router 地址，并保留客户端要求的路径 |
| 模型名称或模型 ID | `default-chat`，或从 Router 复制的完整模型名称 |
| API 密钥 | Olares 内应用留空或填写占位值。外部客户端填写 Router 签发的密钥。 |

1. 保存设置。如果客户端提供连接测试，请先运行测试。
2. 发送一条简短请求。聊天模型可以新建对话并提一个简单问题。
3. 打开 Router 的 **Usage** 页面，确认请求到达预期模型。

## 排查常见连接错误

| 问题 | 检查方法 |
| --- | --- |
| 模型列表中找不到 `default-chat` | 手动添加。模型列表 API 返回具体模型，不包含默认系统名称。 |
| 默认模型不可用 | 在 **Default models** 中确认选中的模型，再到 **LLM** 页面查看状态。设置默认模型不会启动已停止的模型。 |
| 提示找不到模型 | 使用 `default-chat`，或从 Router 复制带有提供商前缀的完整模型名称。 |
| 鉴权失败 | 外部客户端需要有效的 Router API 密钥，并且该密钥需要有权访问请求的模型。 |
| 地址无法访问或返回 404 | 从与客户端位置匹配的标签页复制地址，并确认 Base URL 是否需要包含 `/v1`。 |
| 浏览器报告 CORS 错误或打开 Olares 登录页 | 确认客户端通过服务器还是浏览器直接发送请求，并按照应用教程选择正确的请求方式。 |

## 连接其他应用服务

有些客户端需要连接工作流服务、网关或文档处理应用自身的 API。请使用对应应用教程提供的端点和凭证。Router Base URL 和 `default-chat` 只适用于通过 Router 调用的能力。

SearXNG、Firecrawl 等工具需要先注册到 Router，客户端才能通过网关调用。完整的搜索配置示例，请参阅[使用 Lares 运行深度研究任务](/zh/use-cases/lares.md#运行深度研究任务)。

## 应用教程

- [使用 LobeHub 构建本地 AI 助手](/zh/use-cases/lobechat.md)
- [设置 Open WebUI 进行本地 AI 聊天](/zh/use-cases/openwebui.md)
- [使用 Dify 定制本地 AI 助手](/zh/use-cases/dify.md)
- [将 OpenCode 设置为 AI 编程助手](/zh/use-cases/opencode.md)
- [使用 Claude Code 编写代码](/zh/use-cases/claude-code.md)

## 了解更多

- [使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router.md)：了解 Router 的架构、能力、身份体系和命名规则。
- [通过引擎基座应用运行本地模型](/zh/use-cases/llm-base-apps.md)：部署和管理本地推理引擎。
- [管理应用入口](../olares/settings/manage-entrance.md)：查找服务端点并配置访问策略。
