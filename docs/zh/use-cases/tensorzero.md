---
outline: [2, 3]
description: 在 Olares 上安装 TensorZero，将应用连接到 AI 模型，监控性能表现，并在统一平台管理配置。
head:
  - - meta
    - name: keywords
      content: Olares, TensorZero, LLMOps, AI gateway, observability, evaluation, MCP, Ollama, self-hosted
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-05-09"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/tensorzero.md)为准。
:::

# 使用 TensorZero 作为 AI 模型网关和可观测性平台

TensorZero 是一个一体化平台，用于管理、连接和监控你的 AI 模型。它充当一个中央网关，将你的客户端应用连接到本地 AI 模型。它会记录每一次对话和请求，让你能够追踪性能表现，并帮助你测试不同的配置以获得最佳结果。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 TensorZero。
- 理解 TensorZero 如何管理 AI 连接。
- 连接一个聊天模型和一个功能。
- （可选）连接一个嵌入模型。
- 使用内置的 Playground 测试你的配置。
- 将其他应用连接到 TensorZero。
- 通过内置的 MCP 服务器让你的 AI 智能体读取性能数据。

## 前提条件

- 确保已[安装 Ollama](ollama.md)，并至少下载了一个聊天模型（例如 `qwen3.5:9b`）和一个嵌入模型（例如 `nomic-embed-text`）。
- 确保你的客户端应用（如 OpenCode 和 AgentZero）已经安装并完全可用。本指南仅涵盖将它们连接到 TensorZero 所需的特定设置。

## 安装 TensorZero

1. 打开 Market，搜索 "TensorZero"。

    ![从 Market 搜索 TensorZero](/images/manual/use-cases/tensorzero.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 了解配置要求

TensorZero 不提供图形界面来配置模型。你需要在 Files 中编辑它的配置文件来管理所有设置。

在编辑文件之前，请查看以下规则以避免错误：
- **严格的权限控制**：TensorZero 拒绝直接请求原始模型名称，如 `gpt-4o` 和 `qwen3.5`。你必须为每个要使用的模型定义一个别名。不要在别名中使用点号或冒号。例如，使用 `qwen3_5_9b`，而不是 `qwen3.5:9b`。
- **精确命名**：当你将其他应用连接到 TensorZero 时，必须在模型别名前添加特定前缀，例如 `tensorzero::model_name::<alias>` 和 `tensorzero::function_name::<alias>`。

    :::tip
    对于使用 LiteLLM 框架的应用，你必须在模型名称中包含 `openai/` 前缀。例如，AgentZero 的嵌入功能需要格式为 `openai/tensorzero::embedding_model_name::nomic_embed` 的嵌入模型名称。
    :::

- **格式规则**：配置文件使用 TOML 文本格式。你必须在不同部分之间保持至少一个空行，例如在 `[models]` 和 `[functions]` 之间。如果删除空行，应用可能无法启动。

## 配置聊天模型和功能

要让 TensorZero 工作，你需要两样东西：一个充当 AI 引擎的模型，以及一个作为你的应用与该引擎通信的接入点的功能。

你需要定义模型来告诉 TensorZero AI 在哪里，然后将它链接到一个功能来处理请求。

本示例连接一个本地 Ollama 模型。

1. 打开 Settings，进入 **Applications** > **Ollama** > **Shared entrances** > **Ollama API**，然后复制端点 URL。例如，`http://d54536a50.shared.olares.com`。
2. 打开 Files，然后进入 **Data** > **tensorzero** > **config**。
3. 右键点击 `tensorzero.toml`，然后将其重命名为 `tensorzero.toml.txt`。
4. 双击 `tensorzero.toml.txt`，然后点击 <i class="material-symbols-outlined">edit_square</i>。
5. 在编辑器中，添加以下代码片段：

    - 将 `api_base` 替换为你复制的 Ollama 端点 URL，并追加 `/v1`。
    - 将 `model_name` 替换为你在 Ollama 中下载的模型的确切名称。

    此配置将你的 Ollama 模型注册为别名 `qwen3_5_9b`，并创建一个面向客户端的功能 `general_chat`，将传入的应用请求路由到该模型。

    ```toml
    # models
    [models.qwen3_5_9b]
    routing = ["ollama"]
    [models.qwen3_5_9b.providers.ollama]
    type = "openai"
    api_base = "<ollama-shared-entrance>/v1"
    model_name = "qwen3.5:9b"
    api_key_location = "none"

    # functions
    [functions.general_chat]
    type = "chat"
    [functions.general_chat.variants.my_default_variant]
    type = "chat_completion"
    model = "qwen3_5_9b"
    ```

    ![连接到 Ollama](/images/manual/use-cases/tensorzero-config-ollama.png#bordered)

6. 点击 <i class="material-symbols-outlined">save</i>，然后关闭文件。
7. 将 `tensorzero.toml.txt` 重命名回 `tensorzero.toml`。
8. 打开 Control Hub，进入 **Browse** > **tensorzero-{username}** > **Deployments** > **tensorzero**，然后点击 **Restart** 以应用新设置。

    ![TensorZero pod 重启](/images/manual/use-cases/tensorzero-pod-restart.png#bordered)

## （可选）配置嵌入模型

某些应用需要嵌入模型来搜索文档或构建记忆功能。TensorZero 将嵌入模型与聊天模型分开处理。你必须定义一个专用的嵌入模型。不要为记忆任务使用聊天功能。

1. 在 `tensorzero.toml` 中添加以下代码片段以定义一个嵌入模型：

    - 将 `api_base` 替换为你复制的 Ollama 端点 URL，并追加 `/v1`。
    - 将 `model_name` 替换为你在 Ollama 中下载的嵌入模型的确切名称。

    此配置将你的 Ollama 嵌入模型注册为别名 `nomic_embed`。

    ```toml
    # embedding_models
    [embedding_models.nomic_embed]
    routing = ["ollama"]
    [embedding_models.nomic_embed.providers.ollama]
    type = "openai"
    api_base = "<ollama-shared-entrance>/v1"
    model_name = "nomic-embed-text"
    api_key_location = "none"
    ```

    ![连接到嵌入模型](/images/manual/use-cases/tensorzero-config-embedding.png#bordered)    

2. 在 Control Hub 中重启 **tensorzero** 容器以应用新设置。

## 验证连接

使用内置的 Playground 测试你的功能是否正常工作。

Playground 需要至少一个测试用例（称为 Datapoint）来显示聊天界面。如果你还没有，必须手动创建一个。

1. 从 Launchpad 打开 TensorZero。
2. 从左侧边栏选择 **Datasets**。
3. 点击 **New Datapoint**，然后配置测试用例详情。

    例如，创建一个基础地理测试：
    - **Dataset**：指定一个名称来创建新的测试用例集合。例如，`Baseline tests`。
    - **Function**：选择你之前配置的功能。例如，`general_chat`。
    - **Input**：选择 **+ User Message**，点击 **+ Text**，然后输入一个测试提示。例如，`What is the capital of Spain?`。
    - **Output**：选择 **+ Text**，然后输入你期望模型生成的确切答案。例如，`Madrid`。
    - （可选）**Tags** 和 **Metadata**：输入标签以帮助以后识别此测试用例。例如，添加一个标签，**Key** 设置为 `type`，**Value** 设置为 `QA`。

    ![创建新的数据点](/images/manual/use-cases/tensorzero-new-datapoint.png#bordered)      

4. 点击 **Create Datapoint**。
5. 从左侧边栏选择 **Playground**。
6. 选择你的功能、刚才创建的数据集和你的变体。聊天界面将出现。如果你收到正常回复，说明设置成功。

    ![验证连接](/images/manual/use-cases/tensorzero-playground.png#bordered)   

## 获取 TensorZero 端点

要将其他应用连接到 TensorZero，请获取其入口地址。

1. 打开 Settings，然后进入 **Applications** > **TensorZero** > **Entrances** > **TensorZero**。

    ![TensorZero 端点地址](/images/manual/use-cases/tensorzero-endpoint.png#bordered){width=70%} 

2. 复制端点 URL。例如，`https://ea581361.laresprime.olares.com`。对于兼容 OpenAI 的客户端，你必须在此 URL 后追加 `/openai/v1`。

## 将模型路由到客户端应用

配置你的第三方应用以使用 TensorZero。

### 确定你的模型名称字符串

根据你要调用的资源，使用以下前缀构建正确的模型名称：

| 资源类型 | 必需的字符串格式 | 示例 |
| :--- | :--- | :--- |
| **功能** | `tensorzero::function_name::<alias>` | `tensorzero::function_name::general_chat` |
| **模型** | `tensorzero::model_name::<alias>` | `tensorzero::model_name::qwen3_5_9b` |
| **嵌入** | `tensorzero::embedding_model_name::<alias>` | `tensorzero::embedding_model_name::nomic_embed` |

:::tip
- 不要在别名中使用点号或冒号。例如，使用 `qwen3_5_9b`，而不是 `qwen3.5:9b`。
- 如果模型名称不起作用，请在前面添加 `openai/` 以满足 LiteLLM 框架的要求，然后重试。例如，使用 `openai/tensorzero::embedding_model_name::nomic_embed`。
:::

### 连接你的客户端应用

以下步骤演示如何配置 OpenCode 和 AgentZero，使其通过 TensorZero 路由请求。

<Tabs>
<template #OpenCode>

1. 在 OpenCode 中，点击左下角的 <i class="material-symbols-outlined">settings</i>。
   ![打开 OpenCode 设置](/images/manual/use-cases/opencode-settings.png#bordered)

2. 选择 **Providers**，然后向下滚动并选择 **Custom provider** 旁边的 **Connect**。
   ![选择自定义提供商](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. 输入以下详细信息，然后点击 **Submit**。
   - **Provider ID**：模型提供商的唯一标识符。例如，`olares-ollama-tensorzero`。
   - **Display name**：在提供商列表中显示的名称。例如，`Olares TensorZero`。
    - **Base URL**：输入以 `/openai/v1` 结尾的 TensorZero 端点 URL。例如，`https://ea581361.laresprime.olares.com/openai/v1`。
    - **API key**：输入任意文本。此字段不能为空。
    - **Models**：
        - **Model ID**：输入确切的功能字符串，`tensorzero::function_name::general_chat`。
        - **Display Name**：输入一个描述性名称，以便在界面中轻松识别，例如 `TensorZero Qwen`。

    ![OpenCode 中的 TensorZero 配置](/images/manual/use-cases/tensorzero-opencode.png#bordered){width=70%}     

4. 刷新 OpenCode，进入 **Settings** > **Models**，然后找到你的自定义提供商 **Olares TensorZero**。
5. 验证你添加的模型已启用。

    ![TensorZero 在 OpenCode 中已启用](/images/manual/use-cases/tensorzero-opencode-enable.png#bordered)   

6. 开始一个新的会话，并选择 TensorZero 管理的模型开始聊天。

    ![OpenCode 中的 TensorZero 聊天](/images/manual/use-cases/tensorzero-opencode-chat.png#bordered)

7. 打开 TensorZero，检查可观测性数据。例如，在 **Inferences** 页面上，你发送的每个请求都会显示为一条日志条目，这确认 TensorZero 成功路由了流量。

    ![TensorZero Inferences 页面](/images/manual/use-cases/tensorzero-inferences.png#bordered)

8. 选择一个条目查看详情。

    ![TensorZero Inferences 条目详情](/images/manual/use-cases/tensorzero-inferences-details.png#bordered)
</template>
<template #AgentZero>

1. 打开 Agent Zero，然后进入 **Settings** > **Agent Settings**。
2. 点击 **Chat Model**，按如下方式配置，然后点击 **Save**。

    - **Chat model provider**：选择 **Other OpenAI compatible**。
    - **Chat model name**：输入 `tensorzero::function_name::general_chat`。
    - **Chat model API base URL**：输入以 `/openai/v1` 结尾的 TensorZero 端点 URL。例如，`https://ea581361.laresprime.olares.com/openai/v1`。
    - **API key**：输入任意文本。此字段不能为空。

    ![AgentZero 中的 TensorZero 配置](/images/manual/use-cases/tensorzero-agentzero.png#bordered)

3. 点击 **Embedding Model**，按如下方式配置，然后点击 **Save**。

    - **Embedding model provider**：选择 **Other OpenAI compatible**。
    - **Embedding model name**：输入 `openai/tensorzero::embedding_model_name::nomic_embed`。

        :::tip
        对于使用 LiteLLM 框架的应用（如 AgentZero 的嵌入功能），你必须在模型名称中包含 `openai/` 前缀。
        :::

    - **API key**：输入任意文本。此字段不能为空。
    - **Embedding model API base URL**：输入以 `/openai/v1` 结尾的 TensorZero 端点 URL。例如，`https://ea581361.laresprime.olares.com/openai/v1`。

    ![AgentZero 中的 TensorZero 配置，嵌入模型配置](/images/manual/use-cases/tensorzero-agentzero-embed.png#bordered)    

4. 开始一个新的聊天。

    ![AgentZero 中的 TensorZero 聊天](/images/manual/use-cases/tensorzero-agentzero-chat.png#bordered)

5. 要测试嵌入模型的记忆效果，告诉智能体一个需要记住的具体事实，然后让它回忆该事实。

    ![在 AgentZero 中验证 TensorZero 记忆](/images/manual/use-cases/tensorzero-agentzero-memory.png#bordered)

6. 打开 TensorZero，检查可观测性数据。例如，在 **Inferences** 页面上，你发送的每个请求都会显示为一条日志条目，这确认 TensorZero 成功路由了流量。

    ![TensorZero inferences 页面](/images/manual/use-cases/tensorzero-agentzero-inferences.png#bordered)

7. 选择一个条目查看详情。

    ![AgentZero 中的 TensorZero 聊天推理详情](/images/manual/use-cases/tensorzero-agentzero-inference-details.png#bordered)

</template>
</Tabs>

## 访问内置的 MCP 服务器

TensorZero 在 `/mcp` 端点包含一个内置的 Model Context Protocol (MCP) 服务器。此功能允许你的 AI 智能体查看 TensorZero 中的性能数据。

例如，你可以让你的智能体检索今天 `general_chat` 的平均响应时间，智能体将使用 MCP 连接读取日志并将数据报告给你。

以下示例演示如何配置 OpenCode 以访问此 MCP 工具。

1. 打开 Files，然后进入 **Data** > **opencode** > **.config** > **opencode**。
2. 双击 `opencode.json`，然后点击 <i class="material-symbols-outlined">edit_square</i>。
3. 添加以下 MCP 配置块。确保将 `<tensorzero-endpoint>` 替换为你的实际 TensorZero 端点 URL。

    ```json
    {
    "mcp": {
        "tensorzero": {
        "type": "remote",
        "url": "<tensorzero-endpoint>/mcp",
        "enabled": true
        }
    }
    }
    ```
4. 点击 <i class="material-symbols-outlined">save</i>。
5. 重启 OpenCode 应用以应用更改。在右上角的 **MCP** 标签页中，验证 **tensorzero** 显示为已启用。

    ![TensorZero MCP 在 OpenCode 中已启用](/images/manual/use-cases/tensorzero-mcp.png#bordered){width=50%}

6. 在聊天中直接指示你的 AI 智能体显式使用此工具。例如，输入 `Use the TensorZero MCP tool to analyze the latest inference logs`。

    ![OpenCode MCP 使用](/images/manual/use-cases/tensorzero-mcp-opencode.png#bordered)

## 常见问题

### 连接到模型与连接到功能有什么区别？

在 TensorZero 中，模型和功能都允许你的应用与 AI 通信，但强烈建议将你的应用连接到功能。

- **模型（`tensorzero::model_name::...`）**：这代表原始的 AI 引擎。虽然你可以将客户端应用直接连接到模型，但这样做会绕过 TensorZero 的高级监控功能。
- **功能（`tensorzero::function_name::...`）**：这代表你的应用正在执行的特定任务，例如 `coding_assistant` 或 `text_summarizer`。通过功能连接可以使用 TensorZero 的详细可观测性和统计跟踪。它还允许你将多个不同的功能链接到同一个底层模型，帮助你分别跟踪和优化每个特定任务。

### 错误：model field must start with `tensorzero::function_name::...`

**原因**：你在客户端的模型字段中输入了原始模型名称（如 `qwen3.5:9b`）或格式不正确。

**解决方法**：根据你要连接的内容，始终使用以下三种精确格式之一：

| 你要调用 | 格式 | 示例 |
| :--- | :--- | :--- |
| 功能 | `tensorzero::function_name::<alias>` | `tensorzero::function_name::general_chat` |
| 直接调用模型 | `tensorzero::model_name::<alias>` | `tensorzero::model_name::qwen3_5_9b` |
| 嵌入模型 | `tensorzero::embedding_model_name::<alias>` | `tensorzero::embedding_model_name::nomic_embed` |

### 错误：`litellm.BadRequestError: LLM Provider NOT provided`

**原因**：此错误发生在依赖 LiteLLM 框架的应用中，例如 AgentZero 的嵌入功能。这些特定应用不会自动识别标准的 TensorZero 模型字符串。它们需要显式的提供商前缀来理解如何格式化连接。

**解决方法**：
查看错误消息详情以确定具体哪个模型失败。打开你的应用设置，并在该模型名称的最前面添加 `openai/`。

例如，如果错误提到 `model=tensorzero::embedding_model_name::nomic_embed`，你必须将你的嵌入模型名称或 ID 更改为 `openai/tensorzero::embedding_model_name::nomic_embed`。保存设置并重试请求。

### 编辑配置文件后 TensorZero 无法启动

**原因**：TOML 格式已损坏，通常是由于部分之间缺少空行或别名中包含无效字符导致的。

**解决方法**：
1. 打开 Control Hub，进入 **tensorzero-{username}** > **Deployments** > **tensorzero** > **Pods**，然后点击 tensorzero pod。
2. 在 **Containers** 部分，找到 **gateway**，然后点击它旁边的 <i class="material-symbols-outlined">article</i>。

    ![容器日志](/images/manual/use-cases/tensorzero-container-logs.png#bordered)

3. 查找以下常见错误：

    - `Failed to parse tensorzero.toml`：语法错误。确保在每个部分块（`# models`、`# functions`、`# embedding_models`）之间恰好有一个空行。如果你在粘贴代码时删除了空行，应用将无法启动。
    - `unknown field`：设置名称不正确，例如别名中包含点号或冒号。使用下划线，如 `qwen3_5_9b`，而不是 `qwen3.5:9b`。
    - `provider...not found`：`routing = ["name"]` 行中的提供商名称与紧接其下方定义的块 `[models.alias.providers.name]` 不匹配。例如，如果你写 `routing = ["ollama"]`，你必须有一个匹配的 `[models.xxx.providers.ollama]` 块。

4. 修复语法后，重启 TensorZero 容器。

### 我的配置更改未在 TensorZero UI 中显示

**原因**：UI 缓存、网关未重新加载配置，或重启失败。

**解决方法**：

尝试以下方法：
- 按 Ctrl+Shift+R 或 Cmd+Shift+R 强制刷新浏览器以清除浏览器缓存。
- 检查 **gateway** 容器日志中是否有 `Starting gateway server...`。如果你看到迁移消息，请再等待 30 秒。
- 重启 TensorZero 容器。

### 官方文档中提到的某些页面（Autopilot、Config Editor）缺失

**原因**：这些是高级组件，不包含在默认的 Olares 部署中。Olares 提供核心网关、UI 和可观测性堆栈。

**解决方法**：如果你需要这些功能，请参阅 [TensorZero 官方文档](https://www.tensorzero.com/docs)以自行托管额外服务。

## 了解更多

- [使用 LiteLLM 作为统一的 AI 模型网关](litellm.md)
- [将 Bifrost 设置为 AI 模型网关](bifrost.md)
