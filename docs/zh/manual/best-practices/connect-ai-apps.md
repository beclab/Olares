---
outline: [2, 3]
description: 学习如何在 Olares 上使用模型控制台和应用入口连接 AI 客户端应用到 AI 服务应用。
---

:::warning
本文档由 AI 翻译生成，仅供参考。如有歧义，请参阅[英文原文](/manual/best-practices/connect-ai-apps.md)。
:::

# 连接 AI 应用 <Badge type="tip" text="^ 1.12.6" />

在 Olares 上使用 AI 时，你通常需要同时操作两个应用：一个 AI 服务应用在后台提供 AI 能力，另一个 AI 客户端应用则提供你每天直接交互的聊天界面。要让它们协同工作，你必须先将它们连接起来。

本指南介绍如何识别你的应用、收集所需的连接信息，并完成连接配置。

## 学习目标

完成本教程后，你将能够：

- 区分 AI 服务应用和 AI 客户端应用。
- 理解连接所需的关键参数，以及如何在 Olares 中获取它们。
- 连接常见的 AI 应用。

## 了解你的 AI 应用

在配置连接之前，先确定哪个应用提供 AI 能力、哪个应用使用这些能力。这样可以确保你在后续步骤中知道是去模型控制台还是 Olares 设置里查找连接信息。

- **AI 客户端应用**：提供你直接交互的前端聊天界面或工作流画布，例如 LobeHub 和 Open WebUI。它们依赖 AI 服务应用（通常称为 **provider**）来执行生成文本、提取数据等 AI 任务。
- **AI 服务应用**：通过 API 为兼容的客户端提供 AI 能力，例如聊天、搜索和语音识别。部分 AI 服务应用自带管理 Web 界面，而另一些则主要作为无界面的后端服务运行。

    在 Olares 上，AI 服务应用分为两类：
    - **LLM 服务应用**：托管大型语言模型（LLM），用于文本生成、代码补全和聊天。包括八个预构建模型应用，以及在[引擎基座应用](/zh/use-cases/llm-base-apps.md)上创建的自定义模型实例。
    - **其他 AI 服务应用**：提供 LLM 之外的其他功能的工具应用，例如语音识别（Speaches）和文本提取（PaddleOCR）。

## 准备连接详情

连接 AI 客户端应用通常涉及配置最多四个关键参数。在设置客户端之前，先收集好这些信息。

### Provider 和 API 格式 <Badge type="tip" text="仅 LLM 服务"/>

在大多数 AI 客户端应用中，**provider** 指提供大语言模型的服务或厂商（例如 OpenAI、Anthropic 或 Ollama）。在 Olares 上，你的本地 **LLM 服务应用**就扮演这个角色。客户端应用不再向云厂商发送请求，而是向本地 LLM 服务应用发送请求。

由于不同 provider 使用不同的通信规则，它们依赖特定的 **API 格式**。你可以把这些格式理解为应用之间交流的“语言”。最常见的两种格式是 **OpenAI-Compatible** 和 **Ollama**。

配置 LLM 连接时，先确认客户端应用支持哪些 provider，然后在模型控制台中选择与客户端匹配的 **API 格式**。模型控制台会根据你的选择动态显示对应的 Base URL。

:::info
PaddleOCR 等非 LLM 服务不使用这些通用格式。它们使用自己工具专属的协议进行通信，因此无需为它们配置 provider 格式。
:::

### Base URL

Base URL 是 AI 服务应用接收并处理任务的地址（即端点）。查找方式取决于服务应用的类别：

- **对于 LLM 服务应用**：打开应用启动**模型控制台**。选择与你客户端运行位置匹配的**连接来源（Connection source）**，选择 **API 格式**，然后复制给出的 **Base URL**。请原样复制，包括 `/v1` 等路径后缀。
- **对于其他 AI 服务应用**：打开 Olares **设置**，进入 **应用** > **[应用名称]** > **入口**，然后复制 **Endpoint URL**。请确保入口的**认证级别**设置为 **Internal**，以便其他应用无需登录即可访问。

    :::tip 多个入口
    部分应用会暴露多个入口。请根据客户端的协议或使用场景选择对应的入口。例如，网页访问使用主入口，程序化集成使用专用 API 入口。
    :::

### 模型名称 <Badge type="tip" text="仅 LLM 服务"/>

模型名称是模型的唯一标识。客户端会在每次请求中发送该 ID，以便后端服务知道应该使用哪个模型文件进行处理。

在**模型控制台**中，按照显示的完整内容复制**模型名称**。

:::warning 复制完整模型名称
不要缩写模型名称，也不要删除任何仓库前缀（如 `unsloth/`）或量化标签（如 `UD-Q4_K_XL`）。否则客户端可能会返回类似 “Model not found” 这样的错误。
:::

### API key

API key（也称为“Auth Token”或“API Token”）是一种密钥凭证。AI 客户端应用将其提供给 AI 服务应用，以证明自己的身份并获取使用 AI 能力的权限。

对于部署在 Olares 本地的 AI 服务应用，通常不需要真实的 API key。内部入口默认信任来自同一集群中其他应用的请求。

然而，大多数 AI 客户端应用仍要求该字段必须有值。此时可以输入任意占位文本，例如 `olares` 或 `local`，即可继续。

## 示例

以下示例展示如何在真实客户端应用中收集并应用上述连接参数。开始之前，请确保相关应用已安装。

### 将 LLM 服务应用连接到 LobeHub

在本示例中，预构建模型应用 Gemma 4 26B（Ollama）是 LLM 服务应用，LobeHub（前身为 LobeChat）是客户端应用。

1. 从启动台打开 **Gemma 4 26B (Ollama)**，启动其模型控制台。
2. 确保 **Model** 显示 **READY**，**Engine** 显示 **RUNNING**。
3. 选择与你客户端应用匹配的连接选项，以获取正确的 Base URL：

    - **Connection source**：选择 **Apps in Olares**，因为 LobeHub 直接安装在 Olares 集群中。
    - **API format**：选择 **Ollama**，因为 LobeHub 的 Ollama provider 需要这种格式。

4. 记下以下信息：

    - **Model name**：`gemma4:26b`
    - **Base URL**：`https://74bfa5ee.alice2026.olares.com`

    ![LobeHub 的模型控制台](/images/manual/tutorials/connect-app-exp-model-console.png#bordered)

5. 打开 **LobeHub**，然后进入 **Settings** > **AI Service Provider** > **Ollama**。
6. 在 **Interface proxy address** 字段中，粘贴你复制的 **Base URL**。
7. 确保 **Use Client Request Mode** 已关闭。

    :::warning 禁用 Client Request Mode
    不要在 LobeHub 中启用 **Use Client Request Mode**。启用后，应用会改为通过浏览器发起前端调用，可能触发跨域（CORS）限制或 Olares 安全认证提示。保持关闭可确保安全的后端到后端通信。
    :::

8. 在 **Model List** 旁边，点击 **Fetch models**。模型名称 `gemma4:26b` 会出现在列表中。
9. 打开 `gemma4:26b` 的开关以启用该模型。
10. 在 **Connectivity Check** 右侧，从下拉列表中选择该模型名称，然后点击 **Check**。当显示 **Check Passed** 时，连接即建立成功。

    ![LobeHub 成功连接模型](/images/manual/tutorials/connect-app-eg-lobehub2.png#bordered)

### 将 PaddleOCR 连接到 Open WebUI

在本示例中，PaddleOCR 是提供图片文字识别能力的 AI 服务应用，Open WebUI 是客户端应用。

1. 打开 Olares **设置**，然后进入**应用** > **PaddleOCR** > **入口** > **PaddleOCR**。

   ![PaddleOCR 入口](/images/manual/use-cases/paddleocr-entrances.png#bordered){width=75%}

2. 确保**认证级别**设置为 **Internal**。
3. 复制端点 URL。例如：`https://17b4c78a.alice2026.olares.com`。

   ![PaddleOCR 端点](/images/manual/use-cases/paddleocr-endpoint.png#bordered){width=75%}

4. 在 Open WebUI 中，点击左下角的头像图标，然后进入 **Admin Panel** > **Settings** > **Documents**。
5. 在 **General** 区域中，按如下方式配置设置：

    - **Content Extraction Engine**：选择 **PaddleOCR-vl**。
    - **API Base URL**：输入你刚刚复制的 PaddleOCR 端点 URL。
    - **API Token**：输入任意占位文本。请勿留空。

   ![Open WebUI 中的 PaddleOCR 配置](/images/manual/use-cases/openwebui-paddleocr-config1.png#bordered)

6. 点击右下角的 **Save**。

## 常见问题

### 如何连接非 AI 应用？

相同的内部入口模式也适用于非 AI 应用之间的连接。例如：

- *Arrs 媒体栈使用内部入口 URL 连接 Sonarr、Radarr、Prowlarr、Bazarr 和 qBittorrent。详见[使用 *Arrs 生态管理媒体库](/zh/use-cases/arrs.md)。
- SearXNG 本身不是 AI 模型，但它可以连接到 Vane 等 AI 客户端，用于私有增强搜索。详见[将 SearXNG 连接到 Vane](/zh/use-cases/perplexica.md)。

## 了解更多

- [应用示例](/zh/use-cases/index.md)
- [使用引擎基座应用托管本地大语言模型](/zh/use-cases/llm-base-apps.md)
- [管理应用入口](/zh/manual/olares/settings/manage-entrance.md)
