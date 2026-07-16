---
outline: [2, 3]
description: 学习如何在 Olares 上使用模型控制台和应用入口连接 AI 客户端应用到 AI 服务应用。
---

:::warning
本文档由 AI 翻译生成，仅供参考。如有歧义，请参阅[英文原文](/manual/best-practices/connect-ai-apps-new.md)。
:::

# 连接 AI 应用 <Badge type="tip" text="^1.12.6" />

Olares 上的许多 AI 应用都遵循一种标准的连接模式：一个 **AI 服务应用** 通过 API 提供 AI 能力，另一个 **AI 客户端应用** 则提供你每天交互的聊天界面。理解这个模式后，你就可以连接几乎任何兼容的应用组合。

本教程介绍核心概念，并通过几个实际示例带你完成连接。

:::tip 适用于 Olares 1.12.5 及更早版本
本指南适用于 Olares v1.12.6 及更高版本。AI 应用的连接方式在 v1.12.6 中随模型控制台的引入而发生了变化。如果你使用的是 Olares 1.12.5 或更早版本，请参阅 [旧版 AI 应用连接指南](./connect-ai-apps.md)。
:::

## 学习目标

完成本教程后，你将能够：

- 区分 AI 服务应用和 AI 客户端应用。
- 了解 Olares 中 AI 服务应用的类型。
- 为不同类型的 AI 服务应用找到正确的端点来源。
- 将常见的 AI 应用连接起来。

## 核心概念

### AI 服务应用与 AI 客户端应用

- **AI 服务应用**：通过 API 为兼容的客户端提供 AI 能力，例如聊天、搜索和语音识别。部分 AI 服务应用自带管理 Web 界面，而另一些则主要作为无界面的后端服务运行。
- **AI 客户端应用**：提供你直接交互的聊天界面或工作流画布，但它们依赖 AI 服务应用来执行生成文本、搜索网页或识别文字等 AI 任务。例如 **LobeHub**（前身为 LobeChat）和 **Open WebUI**。

### AI 服务应用的类型

在 Olares 中，AI 服务应用主要分为以下几类：

- **LLM 服务应用**：托管大型语言模型，用于文本生成、代码补全和聊天。包括：
    - **[引擎基座应用](/zh/use-cases/llm-base-apps.md)**：四个可复用的基座应用，分别是 Ollama Engine Base、vLLM Engine Base、SGLang Engine Base 和 llama.cpp Engine Base。你可以将它们克隆为独立的模型实例并自行配置。
    - **预构建模型应用**：八个即开即用的应用，将特定模型与特定引擎打包在一起，例如 Qwen3.6-27B（llama.cpp）和 Gemma 4 26B（Ollama）。
- **其他 AI 服务应用**：提供 LLM 之外的其他 AI 能力，例如 Speaches（语音识别）和 PaddleOCR（文本识别）。

### AI 服务应用如何暴露端点

端点是访问应用入口的 URL。在配置 AI 客户端应用时，这个端点通常被称为 **Base URL**。在 Olares 中，AI 服务应用根据所提供能力的类型，通过两种不同路径暴露端点：

| 服务类型 | 端点位置 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| **LLM 服务** | 模型控制台 | 根据客户端位于 Olares 内部、本地网络还是远程，提供动态、网络优化的 API。<br><br>打开应用启动模型控制台，然后获取 **Base URL**。 | <ul><li>Qwen3.6-27B<br>(llama.cpp)</li><li>Gemma 4 26B (Ollama)</li></ul> |
| **其他 AI 服务** | 应用设置 | 使用标准 HTTPS 端点。<br><br>打开 Olares **设置**，进入 **应用** > **[应用名称]** > **入口**，然后复制 **Endpoint URL**。 | <ul><li>SearXNG</li><li>PaddleOCR</li></ul> |

:::tip 多个入口
部分应用会暴露多个入口。请根据客户端的协议或使用场景选择对应的入口。例如，网页访问使用主入口，程序化集成使用专用 API 入口。
:::

### 认证级别

Olares 为应用入口提供以下访问级别：

- **Internal（推荐）**：允许应用之间无需登录提示即可通信，也允许通过本地网络或 LarePass VPN 访问。
- **Public**：向互联网上的任何人开放。不建议用于私有服务。

如何应用这些访问级别取决于 AI 服务应用的类型：

- 对于 LLM 服务应用，模型控制台通过你选择的 **Connection source** 来处理访问控制。
- 对于端点来自 **设置 > 应用 > [应用名称] > 入口** 的其他 AI 服务应用，在连接客户端之前，需将入口的**认证级别**设置为 **Internal**。

:::info 非 AI 应用也使用相同模式
相同的内部入口模式也适用于非 AI 应用之间的连接。例如，*Arrs 媒体栈使用内部入口 URL 连接 Sonarr、Radarr、Prowlarr、Bazarr 和 qBittorrent。详见[使用 *Arrs 生态管理媒体库](/zh/use-cases/arrs.md)。
:::

## 示例

以下示例展示如何连接 AI 服务应用与 AI 客户端应用。示例默认相关应用已经安装并配置完成。

### 将 LLM 服务应用连接到 LobeHub

在本示例中，预构建模型应用 Gemma 4 26B（Ollama）是 LLM 服务应用，LobeHub 是客户端应用。

1. 从启动台打开 **Gemma 4 26B (Ollama)**，启动其模型控制台。
2. 确保 **Model** 显示 **Ready**，**Engine** 显示 **Running**。
3. 指定以下设置：

    - **Connection source**：选择 **Apps in Olares**，因为 LobeHub 直接安装在 Olares 集群中。
    - **API format**：选择 **Ollama**，因为 LobeHub 的 Ollama 提供方需要这种格式。

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

### 将 SearXNG 连接到 Vane

在本示例中，SearXNG 是提供网页搜索能力的 AI 服务应用，Vane（前身为 Perplexica）是客户端应用。

1. 打开 Olares **设置**，然后进入 **应用** > **SearXNG** > **入口** > **SearXNG**。
2. 在 **访问策略** 区域，确保**认证级别**设置为 **Internal**。
3. 在 **Endpoint settings** 区域，复制端点 URL，例如 `https://84a93c3c.alice2026.olares.com`。

    ![设置中的 SearXNG 端点](/images/manual/tutorials/connect-apps-searxng-endpoint.png#bordered){width=70%}

4. 在 Vane 主页，点击左下角的 **Settings** 图标，然后选择 **Search**。
5. 输入你复制的 SearXNG 内部端点 URL，例如 `https://84a93c3c.alice2026.olares.com`。

    ![Vane 中的 SearXNG 设置](/images/manual/tutorials/connect-apps-searxng-vane.png#bordered)

6. 点击 **Save**。
7. 在 Vane 聊天框中，输入一个需要联网搜索的问题，例如 `What is the latest news about OpenAI`。

    如果 SearXNG 连接正确，Vane 会显示 **Sources** 和 **Found X results**，然后返回带有网页引用来源的答案。

    ![Vane 使用 SearXNG 搜索结果作答](/images/manual/tutorials/connect-apps-searxng-vane-working.png#bordered)

## 了解更多

- [应用示例](/zh/use-cases/index.md)
- [使用引擎基座应用托管本地大语言模型](/zh/use-cases/llm-base-apps.md)
- [管理应用入口](/zh/manual/olares/settings/manage-entrance.md)
