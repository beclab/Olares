---
outline: [2, 3]
description: 在 Olares 上设置 AnythingLLM，使用 RAG 构建本地知识库。上传文档，使用本地模型进行嵌入，并用自然语言查询你的知识。
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, RAG, knowledge base, local LLM, embedding, Ollama
app_version: "1.0.13"
doc_version: "1.0"
doc_updated: "2026-04-13"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/anythingllm.md)为准。
:::

# 使用 AnythingLLM 构建本地知识库

AnythingLLM 是一个开源的、一体化 AI 应用，让你可以使用检索增强生成（RAG）与文档进行对话。它支持多个 LLM 提供商、嵌入引擎和向量数据库，全部在 Olares 设备上本地运行。

## 学习目标

在本指南中，你将学习如何：
- 从 Market 安装聊天模型和嵌入模型。
- 配置 AnythingLLM 以通过共享端点使用这些模型。
- 创建工作空间并上传文档以构建知识库。
- 使用自然语言查询你的知识库。

## 前提条件

- 具有足够磁盘空间和内存的 Olares 设备
- 从 Market 安装共享应用的管理员权限

## 安装 AnythingLLM 和模型应用

要构建本地知识库，需要三个组件：AnythingLLM、用于生成响应的聊天模型，以及用于处理文档的嵌入模型。

本指南使用 "Qwen3.5 9B" 作为聊天模型，"Nomic Embed v1.5" 作为嵌入模型。

1. 打开 Market 并搜索 "AnythingLLM"。

   ![安装 AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. 点击 **Get**，然后点击 **Install**。

3. 搜索 "Qwen3.5 9B" 并安装它。

   ![安装 Qwen3.5 9B](/images/manual/use-cases/qwen35-9b.png#bordered)

4. 搜索 "Nomic Embed v1.5" 并安装它。

   ![安装 Nomic Embed v1.5](/images/manual/use-cases/nomic-embed.png#bordered)

5. 等待所有安装完成。

## 下载模型并获取共享端点

安装后，每个模型应用会自动下载其模型。你必须获取每个模型的共享端点 URL，以便将 AnythingLLM 连接到这些模型。

### 获取聊天模型端点

1. 从 Launchpad 打开 Qwen3.5 9B Q4_K_M (Ollama) 应用，并等待模型下载完成。
2. 打开 Settings，然后前往 **Applications** > **Qwen3.5 9B Q4_K_M (Ollama)**。
3. 在 **Shared entrances** 中，选择 **Qwen3.5 9B Q4_K_M** 查看端点 URL。

   ![Qwen3.5 9B 共享入口](/images/manual/use-cases/anythingllm-qwen359b-shared-entrance.png#bordered){width=80%}

4. 复制共享端点 URL。例如：
   ```plain
   http://bd5355000.shared.olares.com
   ```
### 获取嵌入模型端点

1. 从 Launchpad 打开 Nomic Embed v1.5 应用，并等待模型下载完成。
2. 打开 Settings，然后前往 **Applications** > **Nomic Embed v1.5**。
3. 在 **Shared entrances** 中，选择 **Nomic Embed v1.5** 查看端点 URL。

   ![Nomic Embed v1.5 共享入口](/images/manual/use-cases/anythingllm-nomic-shared-entrance.png#bordered){width=80%}

4. 复制共享端点 URL。例如：
   ```plain
   http://8298761c0.shared.olares.com
   ```

## 配置 AnythingLLM

默认情况下，AnythingLLM 连接到 Ollama 应用的共享端点，用于聊天模型和嵌入器。因为你安装了专用的模型应用，所以必须更新这些端点以指向正确的模型。

这些设置作为所有工作空间的系统默认值。你也可以为单个工作空间自定义不同的模型。

### 设置聊天模型

1. 从 Launchpad 打开 AnythingLLM 应用。
2. 在主页上，点击左下角的 **Open settings** 图标。
3. 在左侧边栏中，选择 **AI Providers** > **LLM**，然后选择 **Ollama** 作为 LLM 提供商。
4. 在 **Ollama Base URL** 字段中，粘贴 Qwen3.5 9B 的共享端点 URL。**qwen3.5:9b** 会自动显示在 **Ollama Model** 中。

   ![配置聊天模型](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered)

5. 点击 **Save changes**。显示 "LLM preferences saved successfully" 消息。

### 设置嵌入模型

1. 在左侧边栏中，选择 **Embedder**，然后选择 **Ollama** 作为嵌入提供商。
2. 在 **Ollama Base URL** 字段中，粘贴 Nomic Embed v1.5 的共享端点 URL。**nomic-embed-text:v1.5** 会自动显示在 **Ollama Embedding Model** 中。

   ![配置嵌入模型](/images/manual/use-cases/anythingllm-configure-embedding.png#bordered)

3. 点击 **Save changes**。显示 "Embedding preferences saved successfully" 消息。
<!--
:::info 默认嵌入模型
AnythingLLM 包含一个内置的嵌入模型（all-MiniLM-L6-v2），无需额外设置即可工作，主要针对英文文档进行训练。如果你更喜欢零配置选项，可以使用默认嵌入器。
:::
-->

## 创建工作空间

1. 点击左上角的 **AnythingLLM** 返回主页。
2. 点击搜索栏旁边的 <span class="material-symbols-outlined">add_2</span>。

   ![创建工作空间](/images/manual/use-cases/anythingllm-create-workspace.png#bordered)

3. 在 **New Workspace** 窗口中，命名你的工作空间，例如 `My test`，然后点击 **Save**。

## 上传文档

1. 点击工作空间名称旁边的 <span class="material-symbols-outlined">upload</span> 打开文档管理器。

   ![打开文档管理器](/images/manual/use-cases/anythingllm-open-upload.png#bordered)

2. 通过上传文件或提交链接来上传你的文档。上传的文档和网页会显示在 **My Documents** 面板中。

   ![上传文档](/images/manual/use-cases/anythingllm-upload-documents.png#bordered)

3. 在 **My Documents** 面板中，选择上传的文档，然后点击 **Move to Workspace** 将它们添加到新创建的工作空间。

   ![移动到工作空间](/images/manual/use-cases/anythingllm-move-to-workspace.png#bordered)

4. 点击 **Save and Embed** 开始嵌入。

   根据文档数量，这可能需要几分钟。嵌入完成后，显示 "Workspace updated successfully" 消息。

## 查询你的知识库

就你的文档提出问题。

1. 返回工作空间聊天视图。
2. 通过聊天发送你的问题。例如：

   ```text
   Olares supports backup or not
   ```

3. AnythingLLM 从你的文档中检索相关部分，并基于内容生成答案。

   ![查询结果](/images/manual/use-cases/anythingllm-query-result.png#bordered)

## 了解更多
- [通过 Ollama 下载和运行本地 AI 模型](ollama.md)
- [官方 AnythingLLM 文档](https://docs.anythingllm.com/)
