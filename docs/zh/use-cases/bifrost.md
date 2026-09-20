---
outline: [2, 3]
description: 在 Olares 上将 Bifrost 设置为 AI 网关。将模型聚合在一个端点后面，然后连接 OpenCode 和 Open WebUI 等客户端。
head:
  - - meta
    - name: keywords
      content: Olares, Bifrost, AI gateway, LLM proxy
app_version: "1.0.11"
doc_version: "2.0"
doc_updated: "2026-08-05"
---

# 将 Bifrost 设置为 AI 模型网关

Bifrost 是一个 AI 网关，位于你的客户端应用和多个模型提供商（如 OpenAI、Anthropic 和本地引擎）之间。它暴露一个兼容 OpenAI 的单一端点，并根据模型名称将每个请求路由到正确的后端。

使用 Bifrost 可以实现高请求吞吐量、内置 MCP 网关访问、语义响应缓存和自动提供商故障转移。

## 学习目标

在本指南中，你将学习如何：

- 安装 Bifrost。
- 在 Bifrost 中添加模型提供商。
- 获取 Bifrost 端点 URL。
- 将模型从 Bifrost 路由到 OpenCode 和 Open WebUI。
- 使用 Bifrost 的可观测性日志验证模型连接。

## 前提条件

开始前，你需要以下模型：
| 模型类型 | 模型 | 获取方式 |
| :--- | :--- | :--- |
| 对话 | Qwen3.6-27B (llama.cpp) | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Bifrost

1. 打开 Market 并搜索 "Bifrost"。

   ![Market 中的 Bifrost](/images/manual/use-cases/bifrost.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 在 Bifrost 中添加模型提供商

在 Bifrost 中，模型提供商代表托管你的 AI 模型的引擎。你通过提供运行模型的应用的端点 URL 来配置提供商。

### 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 在 Bifrost 中配置模型提供商

1. 从 Launchpad 打开 Bifrost，前往 **Models** > **Model Providers** > **Add provider**，然后选择 **Custom provider**。

   ![选择 Custom provider](/images/manual/use-cases/bifrost-add-provider.png#bordered)

2. 在 **Add Custom Provider** 面板中，配置以下设置：

   - **Name**：例如 local-qwen36
   - **Base Format**：选择 **OpenAI**。
   - **Base URL**：输入你从 Model Console 复制的 **Base URL**，不包含 `/v1`。例如：`https://e46e044d.laresprime.olares.com`。
   - **Allow Private Network**：启用它以连接托管在私有或内部网络上的模型。
   - **Is Keyless**：启用它，因为该提供商不需要 API key。

   ![编辑 custom provider 配置](/images/manual/use-cases/bifrost-single-model-config2.png#bordered){width=90%}

3. 点击 **Add**。

## 获取 Bifrost 端点

客户端应用通过 Bifrost 端点 URL 连接到 Bifrost，而不是你之前配置的模型提供商 URL。

1. 打开 Olares **Settings**，前往 **Applications** > **Bifrost** > **Entrances** > **Bifrost**，然后复制端点 URL。例如：

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

在 OpenCode 中，将 Bifrost 注册为自定义提供商，并在其下添加你的示例模型。

### 步骤 1：将 OpenCode 连接到 Bifrost

1. 打开 OpenCode，前往 **Settings** > **Providers** > **Custom provider**，然后点击右侧的 **Connect**。
2. 输入以下详细信息：
   - **Provider ID**：提供商的唯一标识符。例如，`olares-bifrost`。
   - **Display name**：在提供商或模型列表中显示的名称。例如，`Olares Bifrost`。
   - **Base URL**：附加了 `/v1` 的 Bifrost 端点 URL。例如：`https://44039dc0.laresprime.olares.com/v1`。
   - **model-id**：输入你从 Model Console 复制的 **Model name**。例如：`unsloth/Qwen3.6-27B-GGUF:Q4_K_M`。
   - **Display Name**：指定一个友好的标签来标识该模型，例如 `Qwen3.6 27B`。
   - 要添加多个模型，点击 **Add model**。

   ![在 OpenCode 中添加模型](/images/manual/use-cases/bifrost-opencode-add-model1.png#bordered){width=70%}

3. 点击 **Submit**。显示一条通知，提示提供商已连接。
4. 前往 **Settings** > **Models** > **Olares Bifrost**，然后验证你添加的模型已启用。

   ![OpenCode 中已启用的添加模型](/images/manual/use-cases/bifrost-opencode-add-model-enabled1.png#bordered){width=70%}

### 步骤 2：聊天并验证

1. 在 OpenCode 中开始新会话，然后选择 Bifrost 管理的模型开始聊天。

   ![在 OpenCode 中聊天](/images/manual/use-cases/bifrost-opencode-chat1.png#bordered)

2. 打开 Bifrost，然后前往 **Observability** > **LLM Logs**。

   你发送的每个请求都会显示为一个日志条目，这确认 Bifrost 成功路由了流量。

   ![Bifrost LLM 日志](/images/manual/use-cases/bifrost-llm-logs1.png#bordered)

## 将模型路由到 Open WebUI

在 Open WebUI 中，将 Bifrost 添加为直接外部连接，并在其下添加示例模型。

### 步骤 1：将 Open WebUI 连接到 Bifrost

1. 在 Open WebUI 中，点击你的用户头像，然后选择 **Admin Panel**。
2. 点击 **Settings** 选项卡，找到 **AI** 部分，然后选择 **Connections**。
3. 在 **Manage OpenAI Connections** 右侧，点击 <span class="material-symbols-outlined">add</span> 以添加新连接。
4. 在 **Add Connection** 窗口中，指定以下设置：
   - **URL**：粘贴附加了 `/v1` 的 Bifrost 端点 URL。
   - **Auth**：选择 **None**。
   - **Add a Model ID**：展开 **Advanced**，输入你从 Model Console 复制的 **Model name**，然后点击 <span class="material-symbols-outlined">add</span>。

   ![Open WebUI 添加连接窗口](/images/manual/use-cases/bifrost-openwebui-connection-form1.png#bordered){width=50%}

5. 点击 <span class="material-symbols-outlined">refresh</span> 验证连接，然后点击 **Save**。
6. 确保已启用 **Direct Connection**。
7. 点击 **Save**。

### 步骤 2：聊天并验证

1. 在 Open WebUI 中，前往 **New Chat** 页面。
2. 选择已配置的模型，然后开始对话。

   ![Open WebUI 聊天](/images/manual/use-cases/bifrost-openwebui-chat1.png#bordered)

3. 打开 Bifrost，然后前往 **Observability** > **LLM Logs**。

   你发送的每个请求都会显示为一个日志条目，这确认 Bifrost 成功路由了流量。

   ![Open WebUI 的 Bifrost 日志](/images/manual/use-cases/bifrost-openwebui-log1.png#bordered)

## 常见问题

### 使用 Bifrost 还是 LiteLLM？

Olares 提供多个 AI 网关。如果你需要高请求吞吐量、内置 MCP 网关访问、语义缓存或高级速率限制，请使用 Bifrost。对于不需要这些高级功能的更简单设置，请考虑使用 [LiteLLM](litellm.md)。

### 为什么 OpenCode 连接到 Bifrost 时返回错误？

确保你在客户端配置中的 Bifrost 端点 URL 后附加了 `/v1`。没有 `/v1` 后缀，来自兼容 OpenAI 的客户端的请求将失败。

### 为什么在 OpenCode 中通过 Bifrost 调用模型时会出错？

某些模型具有自己的原生输出格式，如自定义标签或推理块，或缺乏客户端期望的功能支持，如工具调用。当 Bifrost 路由这些请求时，模型可能会返回兼容 OpenAI 的客户端（如 OpenCode）无法解析的响应，从而导致失败。

如果你遇到此问题：
- 查看模型文档以了解特殊输出格式或功能限制。
- 验证模型是否支持你客户端请求的特定功能。
- 切换到完全符合 OpenAI API 标准的模型。

## 了解更多

- [将 OpenCode 设置为你的 AI 编码助手](opencode.md)：完整的 OpenCode 设置和项目工作流程。
- [使用 Open WebUI 与本地 LLM 聊天](openwebui.md)：针对 Olares 托管模型的 Open WebUI 配置。
- [使用 LiteLLM 作为统一的 AI 模型网关](litellm.md)：与 Bifrost 比较以选择适合你栈的网关。
- [Bifrost 官方文档](https://docs.getbifrost.ai)：提供商、MCP、缓存和治理功能的完整参考。
