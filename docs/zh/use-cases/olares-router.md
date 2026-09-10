---
outline: [2, 3]
title: 使用 Olares Router 作为你的 AI 网关
description: 了解 Olares Router 是什么，如何通过一个 OpenAI 兼容端点调用 LLM、音频、创意和工具能力，以及不同调用方应使用哪种凭证。
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI 网关, OpenAI 兼容 API, 模型端点, API 密钥, 多模态
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/olares-router.md)为准。
:::

# 使用 Olares Router 作为你的 AI 网关

Olares Router 是内置于 Olares 的 AI 网关，随 v1.12.7 版本提供。它将 Olares 上的所有 AI 能力（无论本地还是云端）通过一个 OpenAI 兼容端点暴露出来。客户端从不直接连接本地模型或云厂商，而是连接 Router，由 Router 将每个请求路由到正确的后端。

## 为什么选择 Olares Router

许多用户在需要把应用连接到多家 AI 服务提供商时，会部署 LiteLLM 这类独立的 LLM 代理。有了 Router，这层额外的组件就不再需要了。Router 在 Olares 内部解决了同样的核心问题，并更进一步：

- **内置，而非外挂**：Router 随 Olares 一同到来。它将 AI 的访问和管理带入运行你应用和本地 AI 负载的同一套系统。LiteLLM 式的代理则是又一个需要部署、配置和维护的组件。
- **本地 AI，自动发现**：Router 自动发现并同步已安装的模型和实用工具应用。代理则要求每个模型都必须先注册才能使用。
- **内用身份，外用密钥**：Olares 内的应用使用平台身份，无需存储 API 密钥。Olares 外的客户端使用 Router 签发的 API 密钥，云厂商凭证始终留在 Router 内部。用代理时，每个调用者都是又一把需要签发、保存和轮换的密钥。
- **新 AI 能力，同一端点**：Olares 支持的新模型、新模态和新 AI 实用工具都能通过 Router 暴露，无需任何新的部署。客户端使用同一访问层，平台侧集成由 Router 完成。代理也可以扩展，但每增加一项都是新一轮配置。

## 一个网关覆盖所有 AI 能力

安装在 Olares 上的 LLM 服务应用和 AI 实用工具应用会自动出现在 Router 中。云厂商添加后也会出现，并且在同一个地方管理。

Router 将所有这些 AI 能力分为四类：

| 分类 | 包含内容 | 示例 |
| --- | --- | --- |
| **LLM** | 聊天和响应模型，本地或远程 | DeepSeek、OpenAI、Anthropic、Google Gemini，以及运行在你自己 GPU 上的本地 Qwen 模型 |
| **Audio** | 语音识别、合成以及音频管线的其余部分 | 语音转文字（STT）、文字转语音（TTS）、语音活动检测（VAD）、说话人分离、对齐、说话人嵌入、音频增强、音效、变声 |
| **Creative** | 图像和视频生成 | 文生图、文生视频、音乐生成、3D |
| **Tools** | 智能体在模型之外用到的一切 | 嵌入、网页搜索、网页抓取、重排序、光学字符识别（OCR）、翻译 |

![Router 的能力分类](/images/manual/use-cases/router-capabilities.png#bordered)

## 鉴权方式

Router 如何验证调用方，调用方需要提供什么凭证，取决于请求来自哪里。

| 调用方位置 | 工作原理 | <nobr>所需凭证</nobr> |
| --- | --- | --- |
| <nobr>**Olares 中的应用**</nobr> | Olares 内部的每个请求都先经过平台，平台验证调用方后会在请求上盖一个身份头：对运行在 Olares 中的应用是 `X-Olares-App-ID`，对已登录用户是 `X-BFL-USER`。只有平台能添加这个头，因此无法伪造。Router 因此始终知道谁在调用，调用方无需再提供其他任何东西。 | 无 |
| <nobr>**局域网设备**</nobr> | 请求经局域网到达，没有平台盖戳的身份。调用方出示一个 API 密钥（`Authorization: Bearer <api-key>`），Router 验证密钥并将调用记到密钥所有者名下。 | Router 签发的<br>API 密钥 |
| **远程** | 通过互联网也是如此。调用方出示一个 API 密钥，Router 验证密钥并将调用记到密钥所有者名下。 | Router 签发的<br>API 密钥 |

:::info 两种 API 密钥
你为云厂商添加的密钥留在 Router 内部，永远不会到达调用方。调用方出示的始终是 Router 签发的密钥，Router 与云厂商通信时使用自己保存的凭证。
:::

## 从 Router 获取连接信息

连接 Router 一般需要三个参数，Router 为每个参数都提供了获取途径。

- **Base URL**：打开 **LLM** 等能力页面，找到模型，点击右侧的 **View connection example** 图标。

  ![模型行上的 View connection example 图标](/images/manual/use-cases/router-view-connection-examp.png#bordered)

  在 **How to call this model** 窗口中，三种调用方位置各有一个标签页，对应各自的 base URL，可以直接复制到你的客户端。

  | 调用方位置 | Base URL |
  | --- | --- |
  | Olares 中的应用 | `https://router.<your-olares-id>.olares.com/v1` |
  | 局域网设备 | <ul><li>Windows、Linux：`http://router-<your-olares-id>-olares.local/v1`</li><li>macOS：`http://router.<your-olares-id>.olares.local/v1`</li></ul> |
  | 远程 | `https://router.<your-olares-id>.olares.com/v1` |

  ![How to call this model 窗口](/images/manual/use-cases/router-how-to-call-model.png#bordered)

- **Model name**：从 **How to call this model** 窗口复制模型名称。或者在 **Default models** 页面为每类能力设置默认模型，然后使用 `default-chat` 这样的系统名称，而不是具体模型名。
- **API key**：在 **API Keys** 页面创建。只有局域网或互联网的调用方需要。Olares 中的应用可以填任何占位符。

## 配置客户端

拿到连接信息后，下面的示例展示如何在每种调用方位置配置客户端。

### Olares 中的应用

**配置**：OpenClaw，一个运行在 Olares 上的应用。在自定义提供商设置中，将模型指向 Router，无需 API 密钥：

- **API Base URL**：`https://router.<your-olares-id>.olares.com/v1`
- **Model ID**：`default-chat`

**结果**：状态栏显示该会话以 `default-chat` 应答。

![OpenClaw 通过 Router 使用 default-chat 聊天](/images/manual/use-cases/router-client-connect-openclaw.png#bordered)

### 局域网设备

**配置**：OpenCode，运行在与你 Olares 处于同一网络的 Mac 上。在配置文件 `opencode.jsonc` 中添加一个提供商条目，指向 `.local` 地址，并使用在 Router 中创建的 API 密钥：

```jsonc
{
  "provider": {
    "router": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Router",
      "options": {
        "baseURL": "http://router.<your-olares-id>.olares.local/v1",
        "apiKey": "<your-api-key>"
      },
      "models": {
        "default-chat": { "name": "Router-chat" }
      }
    }
  }
}
```

**结果**：Mac 上的 OpenCode 通过 `.local` 地址与 Router 通信，每次构建都显示配置中的显示名称 `Router-chat`。这些调用也会出现在 Router 的 **Usage** 页面，记在你名下。

![OpenCode 通过 Router 的 default-chat 聊天](/images/manual/use-cases/router-client-connect-opencode.png#bordered)

### 远程客户端

:::tip
这种场景下，要确保在 Olares Settings 中将 Router 入口的 **Authentication level** 设置为 **Public**。
:::

**配置**：本地网络之外的客户端，比如连接公共 Wi-Fi 的笔记本电脑。使用公网 base URL `https://router.<your-olares-id>.olares.com/v1`，搭配在 Router 中创建的 API 密钥。同一个密钥在任何地方都通用。

## 跳过客户端：使用 Olares CLI 调用

有时候你只是想看看有什么可以调用，或者马上试一个模型，而不是配置客户端。Olares CLI 是最快的方式，不需要 base URL，不需要 API 密钥，只用你已登录的身份（`X-BFL-USER`）。只需要三条命令：

1. 使用你的 Olares ID 认证 Olares CLI。CLI 会保存地址和你的身份，供后续调用使用。

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

2. 查看 Olares 上所有可调用的内容及其就绪状态。

   ```bash
   olares-cli router call models
   ```

    示例输出：

    ```text
    NAME                                        MODE    SUPPORTS                                                 READINESS  SERVED BY
    Olares/onnx-community/silero-vad            audio   vad                                                      ready      Olares
    Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL  chat    function_calling,parallel_function_calling,reasoning,+3  ready      Olares
    deepseek/deepseek-v4-flash                  chat    assistant_prefill,function_calling,native_streaming,+7   ready      deepseek
    deepseek/deepseek-v4-pro                    chat    assistant_prefill,function_calling,native_streaming,+7   ready      deepseek
    myfirecrawl/scrape                          scrape  -                                                        ready      myfirecrawl
    myjina/reader                               scrape  -                                                        ready      myjina
    mysearxng/search                            search  -                                                        ready      mysearxng
    myserper/search                             search  -                                                        ready      myserper
    mytavily/extract                            scrape  -                                                        ready      mytavily
    mytavily/search                             search  -                                                        ready      mytavily
    mytavily/search-advanced                    search  search                                                   ready      mytavily
    ```

3. 发送一条聊天消息。

   ```bash
   olares-cli router call chat "explain what is AI gateway in one sentence"
   ```

    示例输出：

    ```text
    [reasoning] User asks: "explain what is AI gateway in one sentence". Need answer one sentence. Need final only one sentence.

    An AI gateway is a central access point that routes, secures, manages, and monitors requests to AI models and APIs.

    unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL
    ```