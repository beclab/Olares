---
outline: [2, 3]
description: 在 Olares 上将 Bifrost 设置为 AI 网关。将 Ollama 和单模型应用聚合在一个端点后面，然后连接 OpenCode 和 Open WebUI 等客户端。
head:
  - - meta
    - name: keywords
      content: Olares, Bifrost, AI gateway, LLM proxy, Ollama, OpenCode, Open WebUI, self-hosted
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-04-22"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/bifrost.md)为准。
:::

# 将 Bifrost 设置为 AI 模型网关

Bifrost 是一个 AI 网关，位于你的客户端应用和多个模型提供商（如 OpenAI、Anthropic 和本地引擎如 Ollama）之间。它暴露一个兼容 OpenAI 的单一端点，并根据模型名称将每个请求路由到正确的后端。

使用 Bifrost 可以实现高请求吞吐量、内置 MCP 网关访问、语义响应缓存和自动提供商故障转移。

## 学习目标

在本指南中，你将学习如何：

- 安装 Bifrost。
- 在 Bifrost 中将 Ollama 或单模型应用添加为模型提供商。
- 定位 Bifrost 端点 URL。
- 将模型从 Bifrost 路由到 OpenCode。
- 将模型从 Bifrost 路由到 Open WebUI。
- 使用 Bifrost 的可观测性日志验证模型连接。

## 前提条件

确保你使用以下方法之一在 Olares 上运行了本地 AI 模型：
- **Ollama 应用**：一个托管多个模型的应用。确保 [Ollama 已安装](ollama.md) 并至少下载了一个模型，例如 `llama3.1:8b`。
- **单模型应用**：将特定模型作为独立应用运行。确保从 Market 安装了模型应用且模型已完全下载，例如 **Qwen3.5 9B Q4_K_M (Ollama)**。

## 安装 Bifrost

1. 打开 Market 并搜索 "Bifrost"。

   ![Market 中的 Bifrost](/images/manual/use-cases/bifrost.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 在 Bifrost 中添加模型提供商

在 Bifrost 中，模型提供商代表托管你的 AI 模型的引擎。你通过提供运行模型的应用的端点 URL 来配置提供商。

你可以连接 Ollama 应用以路由其中的每个模型，或连接单模型应用以仅暴露该特定模型。

在本教程中，由于两个示例模型都在 Ollama 引擎上运行，因此为两种场景选择 **Ollama** 作为提供商类型。

<tabs>
<template #Ollama-app>

使用此方法通过 Bifrost 路由 Ollama 实例中每个已下载的模型。

1. 打开 **Settings**，前往 **Applications** > **Ollama** > **Entrances** > **Ollama API**，然后复制端点 URL。例如：

   ```plain
   https://a5be22681.laresprime.olares.com
   ```

   ![Settings 中的 Ollama 端点](/images/manual/use-cases/bifrost-ollama-endpoint.png#bordered){width=80%}

2. 从 Launchpad 打开 Bifrost，前往 **Models** > **Model Providers** > **Add provider**，然后选择 **Ollama**。

   ![选择 Ollama 作为提供商](/images/manual/use-cases/bifrost-add-provider-ollama.png#bordered)

3. 点击右上角的 **Edit Provider Config**。
4. 在 **Base URL** 中，输入你复制的 Ollama 端点 URL。

   ![为 Ollama 编辑提供商配置](/images/manual/use-cases/bifrost-config-provider-ollama.png#bordered){width=90%}

5. 点击 **Save Network Configuration**。显示 "Provider configuration updated successfully" 消息。
6. 关闭 **Ollama Provider configuration** 窗口。
</template>
<template #Single-model-app>

当模型作为其自己的 Olares 应用运行时（如 Qwen3.5 9B Q4_K_M (Ollama)），使用此方法。

1. 打开 **Settings**，前往 **Applications** > **Qwen3.5 9B Q4_K_M (Ollama)** > **Entrances**，点击 **Shared entrances** 下的模型名称，然后记下端点 URL。

   在本例中，它是：

   ```plain
   http://bd5355000.shared.olares.com
   ```

   ![Settings 页面上的模型端点](/images/manual/use-cases/litellm-model-endpoint.png#bordered){width=80%}

2. 从 Launchpad 打开 Bifrost，前往 **Models** > **Model Providers** > **Add provider**，然后选择 **Ollama**。

   ![选择 Ollama 作为提供商](/images/manual/use-cases/bifrost-add-provider-ollama.png#bordered)

3. 点击右上角的 **Edit Provider Config**。
4. 配置以下设置：
   - **Base URL**: 粘贴你复制的端点 URL。确保 Base URL 不以 `/v1` 结尾。
   - **Timeout (seconds)**: 设置为 `300`。单模型应用比运行中的 Ollama 实例需要更长的预热时间。

   ![为单模型应用编辑提供商配置](/images/manual/use-cases/bifrost-single-model-config.png#bordered){width=90%}

5. 点击 **Save Network Configuration**。显示 "Provider configuration updated successfully" 消息。
6. 关闭 **Ollama Provider configuration** 窗口。
</template>
</tabs>

## 获取 Bifrost 端点

客户端应用通过 Bifrost 端点 URL 连接到 Bifrost，而不是你之前配置的模型提供商 URL。

1. 打开 **Settings**，前往 **Applications** > **Bifrost** > **Entrances** > **Bifrost**，然后复制端点 URL。例如：

   ```plain
   https://44039dc0.laresprime.olares.com
   ```

   ![Settings 中的 Bifrost 端点](/images/manual/use-cases/bifrost-endpoint.png#bordered){width=70%}

2. 配置客户端时，始终在此 Bifrost 端点 URL 后附加 `/v1`。例如：

   ```plain
   https://44039dc0.laresprime.olares.com/v1
   ```

:::warning
`/v1` 后缀对于兼容 OpenAI 的客户端是必需的。没有它，请求将失败。
:::

## 将模型路由到 OpenCode

在 OpenCode 中，将 Bifrost 注册为自定义提供商，并在其下添加你的示例模型（来自 Ollama 和单模型应用）。

### 步骤 1：将 OpenCode 连接到 Bifrost

1. 打开 OpenCode，然后前往 **Settings** > **Providers** > **Custom provider** > **Connect**。

   <!--![OpenCode 中的自定义提供商](/images/manual/use-cases/bifrost-opencode-custom-provider.png#bordered)-->

2. 输入以下详细信息：
   - **Provider ID**: 唯一标识符。例如，`olares-bifrost`。
   - **Display name**: 提供商列表中显示的名称。例如，`Olares Bifrost`。
   - **Base URL**: 粘贴附加了 `/v1` 的 Bifrost 端点 URL。

3. 每行添加一个模型。点击 **Add model** 插入更多行，并按如下方式指定每行：
   - **Model ID**: 使用 `ollama/<model-name>` 格式，其中 `<model-name>` 是后端上的精确模型名称。
     - 对于 **Ollama 模型**，使用 Ollama 中显示的名称。例如，`ollama/llama3.1:8b`。
     - 对于 **单模型应用**，使用应用页面上显示的模型名称。例如，`ollama/qwen3.5:9b`。
         ![模型应用页面上的模型名称](/images/manual/use-cases/litellm-model-name.png#bordered){width=55%}
   - **Display name**: 任何友好的标签，如 `Llama 3.1 8B` 或 `Qwen3.5 9B`。
         ![在 OpenCode 中添加模型](/images/manual/use-cases/bifrost-opencode-add-model.png#bordered){width=70%}

   :::warning
   - 你必须在 Bifrost URL 后附加 `/v1`。没有它，OpenCode 会返回错误。
   - 你必须在模型 ID 上包含 `ollama/` 前缀。没有它，API 调用会失败。
   - 你输入的模型名称必须与 Ollama 实例中下载的模型名称完全匹配。要查找已下载模型的精确名称，请在 Ollama 终端中运行 `ollama list`。
   :::

4. 点击 **Submit**。显示 "Olares Bifrost connected" 消息。
5. 返回 OpenCode，然后前往 **Settings** > **Models** > **Olares Bifrost**。
6. 验证你添加的模型已启用。

   ![OpenCode 中已启用的添加模型](/images/manual/use-cases/bifrost-opencode-add-model-enabled.png#bordered){width=70%}

### 步骤 2：聊天并验证

1. 在 OpenCode 中开始新会话，并选择一个 Bifrost 管理的模型开始聊天。

   ![在 OpenCode 中聊天](/images/manual/use-cases/bifrost-opencode-chat.png#bordered)

2. 打开 Bifrost，然后前往 **Observability** > **LLM Logs**。

   你发送的每个请求都会显示为一个日志条目，这确认 Bifrost 成功路由了流量。

   ![Bifrost LLM 日志](/images/manual/use-cases/bifrost-llm-logs.png#bordered)

## 将模型路由到 Open WebUI

在 Open WebUI 中，将 Bifrost 添加为直接外部连接，并在其下添加两个示例模型。

### 步骤 1：将 Open WebUI 连接到 Bifrost

1. 在 Open WebUI 中，点击你的用户头像，然后选择 **Admin Panel**。
2. 点击 **Settings** 选项卡，然后选择 **Connections**。
3. 启用 **Direct Connection**，然后点击 **Manage OpenAI Connections** 右侧的 <span class="material-symbols-outlined">add</span>。

   ![直接连接开关](/images/manual/use-cases/bifrost-openwebui-direct-connection.png#bordered)

4. 在 **Add Connection** 窗口中，指定以下设置：
   - **URL**: 粘贴附加了 `/v1` 的 Bifrost 端点 URL。
   - **Auth**: 选择 **None**。
   - **Add a Model ID**: 以 `ollama/<model-name>` 格式输入每个模型 ID，然后点击 <span class="material-symbols-outlined">add</span> 添加它。例如：
     - `ollama/llama3.1:8b`
     - `ollama/qwen3.5:9b`

   ![Open WebUI 连接表单](/images/manual/use-cases/bifrost-openwebui-connection-form.png#bordered){width=50%}

5. 点击 <span class="material-symbols-outlined">refresh</span> 验证连接，然后点击 **Save**。

### 步骤 2：聊天并验证

1. 在 Open WebUI 中，前往 **New Chat** 页面。
2. 选择其中一个配置的模型，然后开始对话。

   ![Open WebUI 聊天](/images/manual/use-cases/bifrost-openwebui-chat.png#bordered)

3. 打开 Bifrost，然后前往 **Observability** > **LLM Logs**。

   你发送的每个请求都会显示为一个日志条目，这确认 Bifrost 成功路由了流量。

   ![Open WebUI 的 Bifrost 日志](/images/manual/use-cases/bifrost-openwebui-log.png#bordered)

## 常见问题

### 使用 Bifrost 还是 LiteLLM？

Olares 提供多个 AI 网关。如果你需要高请求吞吐量、内置 MCP 网关访问、语义缓存或高级速率限制，请使用 Bifrost。对于不需要这些高级功能的更简单设置，请考虑使用 [LiteLLM](litellm.md)。

### 为什么 OpenCode 连接到 Bifrost 时返回错误？

确保你在客户端配置中的 Bifrost 端点 URL 后附加了 `/v1`。没有 `/v1` 后缀，来自兼容 OpenAI 的客户端的请求将失败。

### 为什么即使连接成功，我的模型调用也会失败？

- **检查模型 ID**：你必须在模型 ID 上包含 `ollama/` 前缀。例如，`ollama/llama3.1:8b`。
- **检查模型名称**：确保模型名称与你的 Ollama 实例中下载的名称完全匹配。

### 为什么在 OpenCode 中通过 Bifrost 调用模型时会出错？

某些模型具有自己的原生输出格式，如自定义标签或推理块，或缺乏客户端期望的功能支持，如工具调用。当 Bifrost 路由这些请求时，模型可能会返回兼容 OpenAI 的客户端（如 OpenCode）无法解析的响应，从而导致失败。

如果你遇到此问题：
- 查看模型文档以了解特殊输出格式或功能限制。
- 验证模型是否支持你客户端请求的特定功能。
- 切换到完全符合 OpenAI API 标准的模型。

## 了解更多

- [通过 Ollama 下载和运行本地 AI 模型](ollama.md)：安装 Ollama 并拉取模型供 Bifrost 路由。
- [将 OpenCode 设置为你的 AI 编码助手](opencode.md)：完整的 OpenCode 设置和项目工作流程。
- [使用 Open WebUI 与本地 LLM 聊天](openwebui.md)：针对 Olares 托管模型的 Open WebUI 配置。
- [使用 LiteLLM 作为统一的 AI 模型网关](litellm.md)：与 Bifrost 比较以选择适合你栈的网关。
- [Bifrost 官方文档](https://docs.getbifrost.ai)：提供商、MCP、缓存和治理功能的完整参考。
