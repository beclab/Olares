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

![Olares Router 架构](/images/manual/use-cases/router-archi2.png#bordered)

如架构图所示，整个系统分为三层：

1. **调用方**：无论请求来自 Olares 内的应用、已登录的 Olares 用户，还是 Olares 之外的第三方，都连接到 Router。调用方从不与后端直接交互。

2. **Router**：每个调用都经由这一个网关进入。Router 负责调用方鉴权、访问控制和配额管理。它将不同来源、不同模态的接口归一化成同一种标准 API，再按名字把调用路由到对应的能力。

3. **后端能力**：Router 将每个请求分发到对应的目的地：
   - **本地能力**：本地工作负载的请求按任务路由到对应的后端能力。LLM 和 Audio 运行在由 Model Console 管理的本地推理引擎上，FlowStudio 运行在自己的工作流运行时上，Tools 由安装的工具应用提供服务。
   - **外部提供商**：远程模型的请求被路由到云平台，如 OpenAI、ElevenLabs、Google Vertex AI 和 Tavily。

### 一个网关覆盖所有 AI 能力

LLM 之所以能收敛到标准 API，是因为单个模型通常就能完成任务，而多模态和功能性工具缺少这样的标准化。以音频工作流为例，它通常需要语音分离、增强、语音识别、文字转语音等多个模型。这些模型来自不同厂商，API 结构各不相同，也不存在通用接口。

Router 通过作为覆盖整个 AI 生态的抽象层解决了这种碎片化。它把语言模型、音频和视频模型、实用工具统一暴露为同一个网关后面的系统能力，为所有 AI 工作负载提供一种标准格式。

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
| <nobr>**Olares 应用**</nobr> | 平台会在来自 Olares 内应用的每个请求上盖一个身份头 `X-Olares-App-ID`，调用方无需再提供其他任何东西。 | 无 |
| <nobr>**Olares 用户**</nobr> | 平台会在来自已登录 Olares 用户的每个请求上盖一个身份头 `X-BFL-USER`，调用方无需再提供其他任何东西。 | 无 |
| <nobr>**第三方**</nobr> | 既不是 Olares 应用也不是 Olares 用户的调用方，需要出示 Router 签发的 API 密钥（`Authorization: Bearer <api-key>`）。Router 验证密钥并将调用记到密钥所有者名下。删除密钥，访问随即终止。<br><br>你自己在 Olares 之外的设备，以及任何 Olares 之外的人，都是通过这种方式访问你的模型。 | Router 签发的<br>API 密钥 |

## 理解模型命名规范

在 Router 中查看或配置模型时，你会注意到不同的命名格式。Router 用这些名字来标识模型的提供商、来源和路由逻辑。理解这些规则对配置 API 请求和排查连接问题至关重要。

### 带斜杠的名字

如果模型名包含一个或多个斜杠，Router 会在第一个斜杠处拆分字符串来解析它：

- **Provider**：第一个斜杠之前的部分
- **Model Name**：第一个斜杠之后的全部内容

例如：

- `deepseek/deepseek-v4-flash`：Provider 是 `deepseek`，模型名是 `deepseek-v4-flash`。
- `Olares/unsloth/Qwen3.5-27B-GGUF:Q4_K_M`：Provider 是 `Olares`，表示本地工作负载。剩余的字符串 `unsloth/Qwen3.5-27B-GGUF:Q4_K_M` 就是精确的模型名，对应它在 Hugging Face 上的仓库路径。

### 不带斜杠的名字

如果模型名不包含斜杠，它就不直接对应某个提供商的某个模型。它表示一种自定义的路由配置，只会属于以下三类之一：

- **系统默认名**：以 `default-` 前缀开头的名字，如 `default-chat` 或 `default-tts`，是系统级名称。它们会自动把请求路由到当前分配给该能力的模型。
- **模型组**：为负载均衡创建的统一名称。例如，你可以创建一个名为 `DeepSeek-V4` 的组，把请求分发到本地模型和远端服务。组名本身不包含斜杠。
- **别名**：为方便而创建的自定义短名称。例如，把一个冗长的模型名改成简单的名字。

## 从 Router 获取连接信息

连接 Router 需要三个参数：

- **Base URL**：打开 **LLM** 等能力页面，找到模型，点击右侧的 **View connection example** 图标。

  <!-- ![模型行上的 View connection example 图标](/images/manual/use-cases/router-view-connection-examp.png#bordered) -->

  <!-- ![How to call this model 窗口](/images/manual/use-cases/router-how-to-call-model.png#bordered) -->

- **Model name**：从 **How to call this model** 窗口复制模型名称。或者在 **Default models** 页面为每类能力设置默认模型，然后使用 `default-chat` 这样的系统名称，而不是具体模型名。
- **API key**：在 **API keys** 页面创建。只有局域网或互联网的调用方需要。Olares 中的应用无需填写。
