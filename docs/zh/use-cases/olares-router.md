---
outline: [2, 3]
title: 使用 Olares Router 作为你的 AI 网关
description: 了解 Olares Router 是什么，如何通过一个统一访问层调用 LLM、音频、创意和工具能力，以及不同调用方应使用哪种凭证。
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI 网关, OpenAI 兼容 API, 模型端点, API 密钥, 多模态
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/olares-router.md)为准。
:::

# 使用 Olares Router 作为你的 AI 网关

Olares Router 是内置于 Olares 的 AI 网关，随 v1.12.7 版本提供。它将 Olares 上的所有 AI 能力，无论本地还是云端，都通过同一个访问层暴露出来。客户端从不直接连接本地模型或云厂商，而是连接 Router，由 Router 将每个请求路由到正确的后端。

## 什么是 Olares Router

要理解 Router 如何简化你的 AI 工作流，可以先看它在平台中的位置，以及它如何把不同的 AI 工作负载统一成标准能力。

### Router 的位置

Router 是整个 AI 生态的中央调度层。所有流量都经由这一个网关，而不是直接面对分散的后端。

![Olares Router 架构](/images/manual/use-cases/router-archi.png#bordered)

如架构图所示，整个系统分为三层：

1. **调用方（入向请求）**：无论请求来自 Olares 内的应用、已登录的 Olares 用户，还是 Olares 之外的第三方，都连接到 Router。调用方从不与后端基础设施直接交互。

2. **Router（网关）**：请求到达后，Router 负责调用方鉴权、配额管理和能力路由。对于外部云模型，Router 使用保存在平台内部的凭证，确保这些密钥不会暴露给调用方。

3. **模型后端（出向路由）**：Router 充当流量调度者，将请求分发到正确的目的地：
   - **本地模型**：本地能力的请求被路由到已安装的模型应用（如 vLLM、SGLang 或 llama.cpp）。每个应用都带有一个 `model_console` 适配器，统一生命周期和 API 协议，并直接从平台的共享存储拉取模型权重。
   - **云厂商**：远程模型的请求经由互联网安全地路由到 OpenAI 或 Anthropic 等厂商，由 Router 保存的凭证完成认证。

### 一个网关覆盖所有 AI 能力

LLM 之所以能收敛到标准 API，是因为单个模型通常就能完成任务，而多模态和功能性工具缺少这样的标准化。以音频工作流为例，它通常需要语音分离、增强、语音识别、文字转语音等多个模型。这些模型来自不同厂商，API 结构各不相同，也不存在通用接口。

Router 作为 AI 生态的通用抽象层解决了这种碎片化。它把语言模型、音频和视频模型、实用工具统一暴露为同一个网关后面的系统能力，为所有 AI 工作负载提供一种标准格式。

为了让整个生态井然有序、易于触达，Router 把这些统一的能力划分为四大类：

| 分类 | 包含内容 | 示例 |
| --- | --- | --- |
| **LLM** | 聊天和响应模型，本地或远程 | DeepSeek、OpenAI、Anthropic、Google Gemini，以及运行在你自己 GPU 上的本地 Qwen 模型 |
| **Audio** | 语音识别、合成以及音频管线的其余部分 | 语音转文字（STT）、文字转语音（TTS）、语音活动检测（VAD）、说话人分离、对齐、说话人嵌入、音频增强、音效、变声 |
| **Creative** | 图像和视频生成 | 文生图、文生视频、音乐生成、3D |
| **Tools** | 智能体在模型之外用到的一切 | 嵌入、网页搜索、网页抓取、重排序、光学字符识别（OCR）、翻译 |

![Router 的能力分类](/images/manual/use-cases/router-capabilities.png#bordered)

## 为什么选择 Olares Router

Router 直接集成在 Olares 平台中，与你的应用和模型同处一套系统。它是一个平台级网关，统一管理访问、生命周期和多模态路由：

- **身份与访问**：Router 复用 Olares 的身份体系。内部调用方使用平台身份完成鉴权，外部调用方使用 Router 签发的 API 密钥。底层云厂商的凭证隔离保存在 Router 内部，不会暴露给调用方。
- **模型可观测性**：Router 与 Model Console 自动同步，发现已安装的模型。它提供界面调整或重启底层引擎，并在 Usage 页面按调用方追踪所有请求指标。
- **统一的多模态能力**：Router 把语言、音频、视频模型以及搜索、嵌入等实用工具聚合在同一个访问层后面。客户端使用一种标准接口，按名字调用任何系统能力。

## 调用方如何向 Router 鉴权

通过 Router 的每一次调用都有明确的调用方，调用方需要提供什么，取决于它是谁。

| 调用方 | Router 如何识别 | <nobr>所需凭证</nobr> |
| --- | --- | --- |
| <nobr>**Olares 应用**</nobr> | 来自 Olares 内部的每个应用请求都先经过平台，平台验证调用方后会在请求上盖一个身份头 `X-Olares-App-ID`。只有平台能添加这个头，所以调用方无需再提供其他任何东西。 | 无 |
| <nobr>**Olares 用户**</nobr> | 来自已登录 Olares 用户的每个请求同样先经过平台，平台验证调用方后会在请求上盖一个身份头 `X-BFL-USER`。只有平台能添加这个头，所以调用方无需再提供其他任何东西。Olares CLI 以同样的方式工作，用你的 Olares ID 认证一次，之后的每次调用都携带你的身份。 | 无 |
| <nobr>**第三方**</nobr> | 既不是 Olares 应用也不是 Olares 用户的调用方，需要出示 Router 签发的 API 密钥（`Authorization: Bearer <api-key>`）。Router 验证密钥并将调用记到密钥所有者名下。删除密钥，访问随即终止。<br><br>你自己在 Olares 之外的设备，以及任何 Olares 之外的人，都是通过这种方式访问你的模型。 | Router 签发的<br>API 密钥 |

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
- **API key**：在 **API keys** 页面创建。只有局域网或互联网的调用方需要。Olares 中的应用无需填写。

## 通过 Router 调用模型

下面的示例展示每类调用方如何连接 Router 并调用模型。

### Olares 中的应用

OpenClaw 是一个运行在 Olares 上的应用，它在自定义提供商设置中连接 Router。平台身份已经覆盖鉴权，所以只需要填写 base URL 和模型名：

- **API Base URL**：`https://router.<your-olares-id>.olares.com/v1`
- **Model ID**：`default-chat`

会话随后以 `default-chat` 应答。

![OpenClaw 通过 Router 使用 default-chat 聊天](/images/manual/use-cases/router-client-connect-openclaw.png#bordered)

### Olares 用户

Olares 用户也可以用 Olares CLI 调用模型，不需要 base URL，不需要 API 密钥，只靠平台盖戳的用户身份。

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

### Olares 之外的调用方

Olares 之外的任何调用方都需要两样东西，一个匹配所在位置的 base URL，和一个 Router 签发的 API 密钥。base URL 在局域网和互联网之间不同，但密钥在任何地方都通用。

:::tip
对于通过互联网访问 Router 的调用方，需要在 Olares Settings 中将 Router 入口的 **Authentication level** 设置为 **Public**。局域网访问不受影响。
:::

这个示例使用公网 URL。在同一局域网的设备上，改用连接信息中的 `.local` base URL。

```bash
curl https://router.<your-olares-id>.olares.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-api-key>" \
  -d '{"model": "default-chat", "messages": [{"role": "user", "content": "Hello"}]}'
```
