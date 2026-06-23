---
outline: [2, 3]
description: 在 Olares 上设置 LiteLLM，将多个 AI 模型提供商统一到一个 OpenAI 兼容的 API 后面，然后将其连接到 Open WebUI 等应用。
head:
  - - meta
    - name: keywords
      content: Olares, LiteLLM, AI gateway, model proxy, OpenAI-compatible, Ollama, Open WebUI, self-hosted
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-09"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/litellm.md)为准。
:::

# 使用 LiteLLM 作为统一的 AI 模型网关

LiteLLM 是一个 AI 网关，将来自不同模型提供商（如 OpenAI、Anthropic、Google 和本地引擎如 Ollama）的 API 统一到一个 OpenAI 兼容的接口中。它自动将请求参数转换为目标提供商期望的格式，并将请求路由到正确的后端。

在 Olares 上运行 LiteLLM 可为你提供一个集中管理所有模型配置的地方，在远程和本地提供商之间自由切换，并为其他应用提供一个单一的 API 端点。

## 学习目标

在本指南中，你将学习如何：
- 安装 LiteLLM。
- 在 LiteLLM 中添加和配置来自 Ollama 等提供商的 AI 模型。
- 使用内置 Playground 测试模型连接。
- 生成虚拟密钥并将 LiteLLM 连接到 Open WebUI。
- 监控 API 调用日志和模型使用统计。

## 了解 LiteLLM 网关

LiteLLM 位于你的应用和模型提供商之间，充当代理层：
- **统一接口**：LiteLLM 将来自 OpenAI、Anthropic、Google 和本地引擎（Ollama、vLLM）的不同 API 格式规范化为单一的 OpenAI 兼容标准。
- **自动格式转换**：当你使用标准参数发送请求时，LiteLLM 将它们转换为目标提供商期望的特定参数名称和数据结构。
- **请求路由**：根据请求中的模型名称，LiteLLM 决定将其转发到远程云提供商还是本地模型服务器。

![LiteLLM 网关示意图](/images/manual/use-cases/litellm-gateway.png#bordered){width=80%}

由于这个统一层，你的客户端应用只需要一个 API 端点即可访问所有配置的模型。

## 前提条件

- 一个或多个从 Market 安装的模型应用。本教程使用 **Qwen3.5 9B Q4_K_M (Ollama)** 应用作为示例。
- Olares 管理员权限。

## 安装 LiteLLM

1. 打开 Market 并搜索 "LiteLLM"。

   ![LiteLLM in Market](/images/manual/use-cases/litellm.png#bordered)

2. 点击 **获取**，然后点击 **安装**。
3. 出现提示时，设置环境变量：

   - **UI_USERNAME**：指定管理员账户的用户名。
   - **UI_PASSWORD**：指定管理员账户的密码。
4. 点击 **确认** 并等待安装完成。

## 添加模型

本示例使用模型应用 "Qwen3.5 9B Q4_K_M (Ollama)"。其他提供商的流程类似。

1. 从 Launchpad 打开 Qwen3.5 9B Q4_K_M (Ollama) 应用，然后准确记下页面上显示的模型名称。在本例中，它是 `qwen3.5:9b`。

   ![模型应用页面上的模型名称](/images/manual/use-cases/litellm-model-name.png#bordered){width=55%}

2. 打开 **设置**，前往 **应用** > **Qwen3.5 9B Q4_K_M (Ollama)**，点击 **共享入口** 下的模型名称，然后记下端点 URL。在本例中，它是 `http://bd5355000.shared.olares.com`。

   ![设置页面上的模型端点](/images/manual/use-cases/litellm-model-endpoint.png#bordered){width=80%}

3. 从 Launchpad 打开 LiteLLM，然后使用安装时设置的管理员凭据登录。

4. 从左侧边栏选择 **模型 + 端点**，然后点击 **添加模型** 标签页。

   ![添加模型标签页](/images/manual/use-cases/litellm-add-model-tab.png#bordered)

5. 配置以下设置：

   - **提供商**：选择驱动模型应用的引擎。例如，如果模型应用名称包含 "Ollama"，选择 **Ollama**。
   - **LiteLLM 模型名称**：输入你记下的确切模型名称。在本例中，它是 `qwen3.5:9b`。
   - （可选）**公共模型名称**：指定一个更短的别名，用于外部客户端应用。
   - **API 基础地址**：输入你记下的模型应用共享端点 URL。在本例中，它是 `http://bd5355000.shared.olares.com`。

      :::warning
      不要在 API 基础地址 URL 后附加 `/v1`。添加它会导致连接失败。
      :::

6. 点击页面底部的 **测试连接**。
7. 当 **连接测试结果** 窗口显示连接成功消息时，关闭窗口。

   ![测试连接](/images/manual/use-cases/litellm-test-connection.png#bordered){width=60%}

8. 点击 **测试连接** 旁边的 **添加模型**。你现在可以在 **所有模型** 标签页上查看新添加的模型。

   ![所有模型](/images/manual/use-cases/litellm-all-models.png#bordered)

## 测试模型

1. 从左侧边栏选择 **Playground**。
2. 在 **聊天** 标签页上，配置以下设置：
   - **虚拟密钥来源**：保持默认的 **当前 UI 会话**。
   - **自定义代理基础地址**：留空。填写它会导致错误。
   - **端点类型**：选择与你的模型匹配的模式。对于聊天模型，选择 **v1/chat/completions**。
   - **选择模型**：选择你刚刚添加的模型。在本例中，它是 **qwen3.5:9b**。

   ![Playground 配置](/images/manual/use-cases/litellm-playground.png#bordered)

3. 在 **测试密钥** 面板中，在聊天中发送提示以评估模型的性能。

   例如：

   ```text
   写一个关于机器人发现被遗忘图书馆的 3 段科幻故事
   ```

   你可以查看首 token 时间（TTFT）、总延迟以及输入/输出 token 数等指标。

   ![Playground 测试结果](/images/manual/use-cases/litellm-playground-test.png#bordered)

4. 要检查模型支持的功能和参数，从左侧边栏选择 **AI Hub**，然后在 **模型中心** 标签页上点击 **详情**。

   ![查看模型详情](/images/manual/use-cases/litellm-view-model-details.png#bordered)

   你可以在模型概览页面上查看详情。

   ![模型概览](/images/manual/use-cases/litellm-model-overview.png#bordered)

## 将 LiteLLM 与 Open WebUI 一起使用

本节使用 Open WebUI 作为示例。相同的方法适用于任何支持 OpenAI 兼容 API 的客户端应用。

### 生成虚拟密钥

1. 在 LiteLLM 中，从左侧边栏选择 **虚拟密钥**，然后点击 **创建新密钥**。
2. 在密钥所有权窗口中，配置以下设置：

   - **密钥名称**：输入一个描述性名称以便识别。
   - **模型**：选择此密钥允许访问的模型。
   - 保持所有其他选项为默认值。

   ![创建虚拟密钥](/images/manual/use-cases/litellm-create-key.png#bordered)

3. 点击 **创建密钥**。
4. 在 **保存你的密钥** 窗口中，复制虚拟密钥供后续使用。在本例中，它是 `sk-ZSkc399qrcc3VXutDfxhpA`。

   ![复制虚拟密钥](/images/manual/use-cases/litellm-copy-key.png#bordered){width=60%}

### 获取 LiteLLM API 端点

1. 打开 **设置**，前往 **应用** > **LiteLLM** > **入口** > **LiteLLM API**。
2. 复制 **端点** URL。在本例中，它是 `https://6aead52a1.laresprime.olares.com`。

   ![LiteLLM API 入口](/images/manual/use-cases/litellm-api-entrance.png#bordered){width=80%}

:::info 内部与公开访问
LiteLLM API 端点的 **认证级别** 默认设置为 **内部**，这意味着只有同一本地网络上的应用才能访问它。如果你需要从本地网络外部访问 LiteLLM，请将认证级别更改为 **公开**。LiteLLM 的 API 密钥认证将控制访问。
:::

### 将 Open WebUI 连接到 LiteLLM

1. 启动 Open WebUI，点击左下角的用户头像，然后选择 **管理面板**。
2. 点击 **设置** 标签页，然后点击 **连接**。

   ![Open WebUI 连接页面](/images/manual/use-cases/litellm-openwebui-connection.png#bordered)

3. 在 **OpenAI API** 下，点击 <span class="material-symbols-outlined">add</span> 添加新连接。
4. 在 **添加连接** 窗口中，配置以下设置：

   - **连接类型**：点击 **外部** 将其切换为 **本地**。
   - **API 基础地址**：输入你之前记下的 LiteLLM API URL。
   - **API 密钥**：输入你之前复制的虚拟密钥。

   ![Open WebUI 连接设置](/images/manual/use-cases/litellm-openwebui-connection-setup.png#bordered){width=60%}

5. 点击 <span class="material-symbols-outlined">cached</span> 验证连接。
6. 当你看到 "Server connection verified" 消息时，点击 **保存**。
7. 在 **连接** 下，选择 **模型** 以确认 LiteLLM 中配置的模型现在可用，并显示你之前设置的公共模型名称。

   ![Open WebUI 中的模型](/images/manual/use-cases/litellm-openwebui-models.png#bordered)

### 聊天和监控使用情况

1. 在 Open WebUI 中开始新聊天并选择你的 LiteLLM 管理模型，以验证它在对话中正确响应。

   ![在 Open WebUI 中聊天](/images/manual/use-cases/litellm-openwebui-chat.png#bordered)

2. 返回 LiteLLM 监控你的使用数据。

   - 要查看图形使用统计，从左侧边栏选择 **使用**。

   ![LiteLLM 使用统计](/images/manual/use-cases/litellm-usage.png#bordered)

   - 要查看详细的 API 请求记录，从左侧边栏选择 **日志**。

   ![LiteLLM 日志](/images/manual/use-cases/litellm-logs.png#bordered)

## 了解更多

- [通过 Ollama 下载和运行本地 AI 模型](ollama.md)
- [使用 Open WebUI 与本地 LLM 聊天](openwebui.md)
- [LiteLLM 官方文档](https://docs.litellm.ai/docs/)
