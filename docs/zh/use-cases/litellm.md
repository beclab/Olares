---
outline: [2, 3]
description: 在 Olares 上设置 LiteLLM，将多个 AI 模型提供商统一到一个 OpenAI 兼容的 API 后面，然后将其连接到 Open WebUI 等应用。
head:
  - - meta
    - name: keywords
      content: Olares, LiteLLM, AI gateway, model proxy, OpenAI-compatible, Ollama, Open WebUI, self-hosted
app_version: "1.0.8"
doc_version: "1.1"
doc_updated: "2026-07-29"
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
- 在 LiteLLM 中添加和配置 OpenAI-compatible 本地模型。
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

开始前，你需要：

- Olares 管理员权限。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.6-27B (llama.cpp) | 从 Market 安装 |

## 安装 LiteLLM

1. 打开 Market 并搜索 "LiteLLM"。

   ![LiteLLM in Market](/images/manual/use-cases/litellm.png#bordered)

2. 点击 **获取**，然后点击 **安装**。
3. 出现提示时，设置环境变量：

   - **UI_USERNAME**：指定管理员账户的用户名。
   - **UI_PASSWORD**：指定管理员账户的密码。
4. 点击 **确认** 并等待安装完成。

## 添加模型

### 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

对于 Qwen3.6-27B (llama.cpp)，LiteLLM 使用 OpenAI-compatible API 格式。在模型控制台中选择 **OpenAI-Compatible**，然后按照以下步骤操作：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### 将模型添加到 LiteLLM

1. 从 Launchpad 打开 LiteLLM，然后使用安装时设置的管理员凭据登录。

2. 从左侧边栏选择 **Models + Endpoints**，然后点击 **Add Model** 标签页。

   ![添加模型标签页](/images/manual/use-cases/litellm-add-model-tab.png#bordered)

3. 配置以下设置：

   - **Provider**：选择 **OpenAI**。
   - **LiteLLM Model Name(s)**：输入从模型控制台复制的准确 Model name。本示例中为 `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`。
   - **Public Model Name**：输入容易识别的别名，例如 `qwen3.6-27b`。
   - **API Base**：粘贴从 Qwen3.6-27B 模型控制台复制的 Base URL。请按显示内容原样使用。
   - **API Key**：输入占位值，例如 `local`。

4. 点击页面底部的 **Test Connect**。
5. 当 **Connection Test Results** 窗口显示连接成功消息时，关闭窗口。

6. 点击 **Test Connect** 旁边的 **Add Model**。你现在可以在 **All Models** 标签页中查看新添加的模型。

## 测试模型

1. 从左侧边栏选择 **Playground**。
2. 在 **聊天** 标签页上，配置以下设置：
   - **虚拟密钥来源**：保持默认的 **当前 UI 会话**。
   - **自定义代理基础地址**：留空。填写它会导致错误。
   - **端点类型**：选择与你的模型匹配的模式。对于聊天模型，选择 **v1/chat/completions**。
   - **Select Model**：选择刚刚添加的模型。本示例中为 **qwen3.6-27b**。

3. 在 **测试密钥** 面板中，在聊天中发送提示以评估模型的性能。

   例如：

   ```text
   写一个关于机器人发现被遗忘图书馆的 3 段科幻故事
   ```

   你可以查看首 token 时间（TTFT）、总延迟以及输入/输出 token 数等指标。

4. 要检查模型支持的功能和参数，从左侧边栏选择 **AI Hub**，然后在 **模型中心** 标签页上点击 **详情**。

   你可以在模型概览页面上查看详情。

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
4. 在 **Save your Key** 窗口中，复制虚拟密钥供后续使用。

   ![复制虚拟密钥](/images/manual/use-cases/litellm-copy-key.png#bordered){width=60%}

### 获取 LiteLLM API Endpoint

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

对于 LiteLLM：

1. 前往 Olares **Settings** > **Applications** > **LiteLLM** > **Entrances**。
2. 选择 **LiteLLM API**，然后复制 **Endpoint** URL。

在 Open WebUI 中，将该 Endpoint 用作 **API Base URL**。

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

### 聊天和监控使用情况

1. 在 Open WebUI 中开始新聊天并选择你的 LiteLLM 管理模型，以验证它在对话中正确响应。

   ![在 Open WebUI 中聊天](/images/manual/use-cases/litellm-openwebui-chat.png#bordered)

2. 返回 LiteLLM 监控你的使用数据。

   - 要查看图形使用统计，从左侧边栏选择 **使用**。

   ![LiteLLM 使用统计](/images/manual/use-cases/litellm-usage.png#bordered)

   - 要查看详细的 API 请求记录，从左侧边栏选择 **日志**。

   ![LiteLLM 日志](/images/manual/use-cases/litellm-logs.png#bordered)

## 了解更多

- [使用引擎基座应用托管本地大语言模型](llm-base-apps.md)
- [使用 Open WebUI 与本地 LLM 聊天](openwebui.md)
- [LiteLLM 官方文档](https://docs.litellm.ai/docs/)
