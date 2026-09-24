---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/manual/best-practices/connect-ai-apps
outline: [2, 3]
description: 通过 Olares Router 连接 AI 应用，设置默认聊天模型，获取连接地址，并为 Olares 内应用和外部客户端配置模型名称与 API 密钥。
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI 应用, default-chat, OpenAI 兼容 API, Base URL, API 密钥
---

# 通过 Olares Router 连接 AI 应用

<VersionRouteSelect />

Olares Router 为 AI 应用提供统一的本地和云端模型入口。在 Router 中设置默认聊天模型后，客户端只需配置 Router 的 Base URL 和 `default-chat`。以后更换默认模型，无需逐个修改客户端的连接设置。

本指南适用于 Olares 1.12.7 及以上版本，以 Qwen3.8-27B (llama.cpp) 为本地聊天模型。具体应用的字段和操作步骤，请参阅下方的[应用教程](#应用教程)。

## 开始之前

- 从应用市场安装 AI 客户端和 Qwen3.8-27B (llama.cpp)。
- 从启动台打开 Router。发送请求前，确认模型在 **LLM** 页面显示 **Callable**；如果不可用，查看状态下方显示的原因。
- 在 **Default models** 页面，将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型。

## 获取 Router Base URL

1. 在 Router 中打开 **LLM** 页面，找到模型，点击所在行的 **View connection example** 图标。

   ![查看 Qwen3.8-27B 的连接示例](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. 在 **How to call this model** 窗口中，根据客户端的位置选择标签页：

   | 客户端位置 | 标签页 |
   | --- | --- |
   | 安装在 Olares 内的应用 | **Apps in Olares** |
   | 同一局域网内的电脑或其他设备 | **Devices in LAN** |
   | 从局域网外连接的设备 | **Remote** |

   ![Olares 内应用的 Router 连接信息](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. 复制所选标签页中的 **Base URL**。请使用自己设备上的地址，截图中的地址仅为示例。

OpenAI 兼容客户端通常需要保留末尾的 `/v1`。如果客户端会自动追加 `/v1`，则应填写 Router 根地址。例如，Claude Code 的 `ANTHROPIC_BASE_URL` 使用根地址，由 SDK 追加 `/v1/messages`。具体以各应用教程为准。

## 选择模型名称

通用聊天和 Agent 示例使用 `default-chat`，请求会转发给 Router 中设置的默认聊天模型。修改默认模型会影响所有使用此名称的客户端，因此应确认新模型支持客户端所需的工具调用、图片输入等能力。

`default-chat` 不会出现在模型列表 API 的返回结果中。如果客户端自动获取模型列表，需要手动添加它。如果客户端只允许选择列表中的模型，请改用 Router 中显示的完整模型名称。

如果希望客户端始终使用指定模型，请复制 **How to call this model** 窗口中的完整模型名称，保留提供商前缀。上图中的名称为 `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`。

嵌入、语音等能力需要使用各自的模型或默认路由，不能用 `default-chat` 代替。创建和查询知识库时，应保持嵌入模型一致。

## 查看实际上下文大小 {#check-context-window}

部分客户端（例如 Hermes）需要手动填写上下文大小。在 Router 中查看所用模型的配置：

<!--@include: ../../reusables/ai-service-connections.md#model-context-window-->

## 配置 API 密钥

Router 根据调用方身份进行鉴权，是否需要密钥取决于请求来源：

| 调用方 | 如何填写 |
| --- | --- |
| Olares 内的应用 | 无需 Router API 密钥。允许留空时留空；客户端要求必填时，可填写 `olares` 等占位值。 |
| 已登录 Olares、通过平台发起请求的用户 | 平台会提供用户身份，无需额外填写 Router API 密钥。 |
| 局域网或互联网中的外部客户端 | 在 Router 的 **API keys** 页面创建密钥，并填入客户端。仅连接 VPN 不会自动提供 Olares 用户身份。 |

云端模型提供商的密钥应保存在 Router 的提供商配置中。通过 Router 连接的客户端使用上表中的调用方凭证。

## 配置并测试客户端

打开客户端的提供商或模型设置，填写以下内容：

| 客户端设置 | 填写内容 |
| --- | --- |
| 提供商或 API 格式 | 大多数客户端使用 **OpenAI-compatible**；具体格式以应用教程为准 |
| Base URL | 与客户端位置匹配的 Router 地址，保留客户端要求的路径后缀 |
| 模型名称或模型 ID | `default-chat`，需要时手动添加 |
| API 密钥 | Olares 内应用留空或填写占位值；外部客户端填写 Router 签发的密钥 |

保存设置，运行客户端的连接测试，再发送一条简短消息。在 Router 的 **Usage** 页面检查请求是否到达预期模型。

## 排查常见连接错误

| 问题 | 检查方法 |
| --- | --- |
| 模型列表中找不到 `default-chat` | 手动添加。模型列表返回具体模型，不包含默认路由。 |
| 默认模型不可用 | 检查 **Default models**，确认选中的聊天模型在 **LLM** 页面显示 **Callable**。 |
| 提示找不到模型 | 使用 `default-chat`，或从 Router 复制带有提供商前缀的完整模型名称。直接填写引擎模型名可能无法匹配 Router 中的提供商。 |
| 鉴权失败 | 外部客户端需要 Router 签发的 API 密钥。确认密钥有效，且允许访问请求的模型。 |
| 地址无法访问或返回 404 | 从正确的连接标签页重新复制地址，确认客户端是否要求 Base URL 带有 `/v1`。 |
| 浏览器报告 CORS 错误或打开 Olares 登录页 | 确认请求由客户端服务器还是浏览器直接发出，按应用教程选择请求模式。 |

## 连接其他应用服务

部分教程需要连接网关、工作流服务或文档处理应用自身的 API。这类连接应使用应用教程中的端点和鉴权设置。`default-chat` 仅用于通过 Router 发起的聊天请求。

## 应用教程

- [使用 LobeHub 构建本地 AI 助手](/zh/use-cases/lobechat.md)
- [设置 Open WebUI 进行本地 AI 聊天](/zh/use-cases/openwebui.md)
- [使用 Dify 定制本地 AI 助手](/zh/use-cases/dify.md)
- [将 OpenCode 设置为 AI 编程助手](/zh/use-cases/opencode.md)
- [使用 Claude Code 编写代码](/zh/use-cases/claude-code.md)

## 了解更多

- [使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router.md)：了解能力分类、调用方身份和模型命名。
- [通过引擎基座应用运行本地模型](/zh/use-cases/llm-base-apps.md)：部署和管理本地推理引擎。
