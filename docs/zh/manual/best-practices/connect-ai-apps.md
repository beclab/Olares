---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/manual/best-practices/connect-ai-apps
outline: [2, 3]
description: 通过 Olares Router 为应用接入模型和工具。了解 API 格式、连接地址、模型名称与密钥，并配置聊天模型和搜索服务。
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI 应用, AI 工具, default-chat, default-search, OpenAI 兼容 API, Base URL, API 密钥
---

# 为应用接入 AI 能力

<VersionRouteSelect />

Olares Router 让应用通过统一入口调用本地模型、云端模型和搜索等工具。客户端把请求发送给 Router，由 Router 转发给指定的模型或工具。使用 `default-chat` 等默认系统名称时，以后只需在 Router 中更换默认模型，无需逐个修改客户端。

本指南适用于 Olares 1.12.7 及以上版本，分别以 Qwen3.8-27B (llama.cpp) 和 SearXNG 演示聊天模型与搜索工具的连接。具体应用的字段和操作步骤，请参阅下方的[应用教程](#应用教程)。

## 理解连接中的概念

### AI 客户端与服务

AI 客户端提供你直接操作的界面或工作流，例如 LobeHub 的聊天界面。AI 服务通过 API 提供文本生成、语音识别或网页搜索等能力。

Olares 上的本地模型服务包括应用市场中的预置模型应用，以及通过[引擎基座应用](/zh/use-cases/llm-base-apps.md)创建的模型实例。SearXNG、Firecrawl 等工具应用则提供搜索和网页内容获取能力。完成配置后，兼容的客户端可以通过 Router 调用这些能力。

### 提供商与 API 格式

客户端中的 **Provider** 或 **Engine** 决定它使用哪种 API 格式。通过 Router 连接时，选择双方都支持的格式：

- 使用 OpenAI 兼容接口时，查找 **Custom Provider**、**Custom Endpoint**、**OpenAI** 或 **OpenAI-Compatible**，并将 Base URL 设为 Router 的地址。
- 只有连接示例使用 Ollama API 时，才选择 **Ollama**。后端通过 Ollama 运行模型，并不意味着客户端必须使用这种格式。
- 其他 API 或工具应遵循 Router 中的连接示例和客户端教程。某些工具专用的提供商选项需要填写工具应用自身的端点，而非 Router 地址。

云端提供商的凭证保存在 Router 中。客户端选择 **OpenAI** 来使用其 API 格式，并不代表这里需要填写 OpenAI API 密钥。

### 连接参数

| 参数 | 用途 | 获取位置 |
| --- | --- | --- |
| Base URL | 客户端发送请求的 Router 地址 | 在 **How to call this model** 中选择与客户端位置匹配的标签页 |
| 模型名称 | 指定 Router 应调用哪个模型或能力 | 从 **How to call this model** 复制完整名称，或使用在 **Default models** 中配置的默认系统名称 |
| API 密钥 | 识别调用方并控制访问 | 外部客户端从 **API keys** 页面获取。Olares 内应用无需 Router 签发的密钥。 |

Router 将能力分为 **LLM**、**Audio**、**Creative** 和 **Tools**。工具也可以有模型名称。例如，名为 `localsearxng` 的 SearXNG 提供商使用 `localsearxng/search` 标识搜索能力。

## 准备模型或工具

### 准备聊天模型

1. 从应用市场安装 AI 客户端和 Qwen3.8-27B (llama.cpp)。
2. 从启动台打开 Router，在 **LLM** 页面找到模型并查看状态。发送请求前，模型需要显示 **Callable**。如果不可用，请查看状态下方的原因。
3. 在 **Default models** 页面，将 Qwen3.8-27B (llama.cpp) 设为默认聊天模型，供本例中的 `default-chat` 使用。

设置默认模型只是指定 `default-chat` 请求的目标，不会启动已停止的模型。

### 注册工具应用

部分工具需要先在 Router 中添加为提供商，客户端才能通过网关调用。以 SearXNG 为例：

1. 从应用市场安装 SearXNG。
2. 打开 Settings，前往 **Applications** > **SearXNG** > **Entrances**。复制服务的 **Endpoint** URL，并确认访问策略允许 Olares 内其他应用访问。
3. 打开 Router，前往 **Tools** > **Manage providers**。
4. 选择 **SearXNG**，填写：

   | 设置 | 值 |
   | --- | --- |
   | **Provider name** | `localsearxng` |
   | **SearXNG instance URL** | 从 Settings 复制的端点 |

5. 点击 **Add**。在 **Available** 列表中，点击该工具的添加图标以启用它。

   ![在 Router 中启用 SearXNG](/images/manual/use-cases/router-search-tool-enable.png#bordered)

6. 确认工具出现在 **Configured** 列表中。

   ![SearXNG 已加入 Configured 列表](/images/manual/use-cases/router-search-tool-enabled.png#bordered)

客户端通过 Router 调用搜索时，可以使用 `localsearxng/search`，也可以在 **Default models** 中设置默认搜索模型后使用 `default-search`。客户端需要支持 Router 的搜索 API。在 Lares 中使用搜索的方法，请参阅[运行深度研究任务](/zh/use-cases/lares.md#运行深度研究任务)。

## 获取 Router Base URL

1. 在 Router 中打开所需能力的页面，找到模型或工具，点击所在行的 **View connection example** 图标。Qwen 示例使用 **LLM** 页面，SearXNG 使用 **Tools** 页面。

   ![查看 Qwen3.8-27B 的连接示例](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. 在 **How to call this model** 窗口中，根据客户端的位置选择标签页：

   | 客户端位置 | 标签页 |
   | --- | --- |
   | 安装在 Olares 内的应用 | **Apps in Olares** |
   | 同一局域网内的电脑或其他设备 | **Devices in LAN** |
   | 从局域网外连接的设备 | **Remote** |

   ![Olares 内应用的 Router 连接信息](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. 复制所选标签页中的 **Base URL**。请使用自己设备上的地址，截图中的地址仅为示例。

保留所用 API 的连接示例中显示的路径。OpenAI 兼容聊天客户端通常需要末尾的 `/v1`。如果客户端会自动追加 `/v1`，则应填写 Router 根地址。例如，Claude Code 的 `ANTHROPIC_BASE_URL` 使用根地址，由 SDK 追加 `/v1/messages`。具体以各应用教程为准。

## 选择模型名称

通用聊天和 Agent 示例使用 `default-chat`，请求会转发给 Router 中设置的默认聊天模型。修改默认模型会影响所有使用此名称的客户端，因此应确认新模型支持客户端所需的工具调用、图片输入等能力。

`default-chat` 不会出现在模型列表 API 的返回结果中。如果客户端自动获取模型列表，需要手动添加它。如果客户端只允许选择列表中的模型，请改用 Router 中显示的完整模型名称。

如果希望客户端始终使用指定模型，请复制 **How to call this model** 窗口中的完整模型名称，保留提供商前缀。上图中的名称为 `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`。

搜索、嵌入、语音等能力需要使用各自的模型或默认系统名称，例如搜索使用 `default-search`。不能用 `default-chat` 代替嵌入模型。创建和查询知识库时，应保持嵌入模型一致。

## 查看实际上下文大小 {#check-context-window}

部分客户端（例如 Hermes）需要手动填写上下文大小。在 Router 中查看所用模型的配置：

<!--@include: ../../reusables/ai-service-connections.md#model-context-window-->

## 配置 API 密钥

Router 根据调用方身份进行鉴权，是否需要密钥取决于请求来源：

| 调用方 | 如何填写 |
| --- | --- |
| Olares 内的应用 | 无需 Router API 密钥。允许留空时留空。客户端要求必填时，可填写 `olares` 等占位值。 |
| 已登录 Olares、通过平台发起请求的用户 | 平台会提供用户身份，无需额外填写 Router API 密钥。 |
| 局域网或互联网中的外部客户端 | 在 Router 的 **API keys** 页面创建密钥，并填入客户端。仅连接 VPN 不会自动提供 Olares 用户身份。 |

云端模型提供商的密钥应保存在 Router 的提供商配置中。通过 Router 连接的客户端使用上表中的调用方凭证。

## 配置并测试客户端

打开客户端的提供商或模型设置，填写以下内容：

| 客户端设置 | 填写内容 |
| --- | --- |
| 提供商或 API 格式 | Router 与客户端都支持的格式，参阅[提供商与 API 格式](#提供商与-api-格式) |
| Base URL | 与客户端位置匹配的 Router 地址，保留客户端要求的路径后缀 |
| 模型名称或模型 ID | 能力的完整名称，或已配置的默认系统名称，例如聊天使用 `default-chat` |
| API 密钥 | Olares 内应用留空或填写占位值。外部客户端填写 Router 签发的密钥 |

1. 保存设置。如果客户端提供连接测试，先运行测试。
2. 根据配置的能力发送一条简短请求。聊天模型可以新建对话并提一个问题，搜索工具则执行一次搜索。
3. 打开 Router 的 **Usage** 页面，检查请求是否到达预期模型或工具。

## 排查常见连接错误

| 问题 | 检查方法 |
| --- | --- |
| 模型列表中找不到 `default-chat` | 手动添加。模型列表返回具体模型，不包含默认路由。 |
| 默认模型不可用 | 在 **Default models** 中确认选中的模型，再到能力页面查看状态。聊天模型在 **LLM** 页面检查。设置默认模型不会启动已停止的模型。 |
| 提示找不到模型 | 从 Router 复制带有提供商前缀的完整名称，或使用为该能力配置的默认系统名称。直接填写引擎模型名可能无法匹配 Router 中的提供商。 |
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

- [管理应用入口](../olares/settings/manage-entrance.md)：查找服务端点并配置访问策略。
- [使用 Olares Router 作为 AI 网关](/zh/use-cases/olares-router.md)：了解能力分类、调用方身份和模型命名。
- [通过引擎基座应用运行本地模型](/zh/use-cases/llm-base-apps.md)：部署和管理本地推理引擎。
