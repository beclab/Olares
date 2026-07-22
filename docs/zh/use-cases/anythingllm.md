---
outline: [2, 3]
description: 在 Olares 上部署 AnythingLLM，用 RAG 构建私有、自托管的知识库。上传文档、用本地模型做嵌入，并用自然语言查询。
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, self-hosted rag, private knowledge base, anythingllm ollama, local LLM, embedding, anythingllm on olares
app_version: "1.0.13"
doc_version: "2.0"
doc_updated: "2026-07-22"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/anythingllm.md)为准。
:::

# 使用 AnythingLLM 构建本地知识库

AnythingLLM 是一个开源的、一体化 AI 应用，让你可以使用检索增强生成（RAG）与文档进行对话。它支持多个 LLM 提供商、嵌入引擎和向量数据库，全部在 Olares 设备上本地运行。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 AnythingLLM。
- 通过模型控制台将 AnythingLLM 连接到这两个模型。
- 创建工作空间并上传文档以构建知识库。
- 使用自然语言查询你的知识库。

## 前提条件

- 具有足够磁盘空间和内存的 Olares 设备
- 从 Market 安装应用的管理员权限
- 一个聊天模型和一个嵌入模型。本指南使用 Qwen3.5 9B 和 Qwen3 Embedding 0.6B。你可以从 Market 安装预构建模型应用，或[使用引擎基座应用托管模型](llm-base-apps.md)。

## 安装 AnythingLLM

1. 打开 Market 并搜索 "AnythingLLM"。

   ![安装 AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

对于本指南中使用的每个模型：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

本例使用 `qwen3.5:9b` 进行聊天，使用 `qwen3-embedding:0.6b` 生成嵌入。AnythingLLM 通过 **Ollama** 提供方连接这两个模型，因此请在各自的模型控制台中查看 **Ollama** 格式并复制对应的 Base URL。

## 配置 AnythingLLM

分别配置聊天模型和嵌入模型。这些设置将成为所有工作空间的系统默认值。

### 设置聊天模型

1. 从 Launchpad 打开 AnythingLLM。
2. 在主页上，点击左下角的 **Open settings** 图标。
3. 在左侧边栏中，选择 **AI Providers** > **LLM**，然后选择 **Ollama** 作为 LLM 提供商。
4. 在 **Ollama Base URL** 中，粘贴 Qwen3.5 9B 模型控制台中的 Base URL。
5. 在 **Ollama Model** 中选择 `qwen3.5:9b`。

   ![配置聊天模型](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered)

6. 点击 **Save changes**。显示 "LLM preferences saved successfully" 消息。

### 设置嵌入模型

1. 在左侧边栏中，选择 **Embedder**，然后选择 **Ollama** 作为嵌入提供商。
2. 在 **Ollama Base URL** 中，粘贴 Qwen3 Embedding 0.6B 模型控制台中的 Base URL。
3. 在 **Ollama Embedding Model** 中选择 `qwen3-embedding:0.6b`。

   ![配置嵌入模型](/images/manual/use-cases/anythingllm-configure-embedding1.png#bordered)

4. 点击 **Save changes**。显示 "Embedding preferences saved successfully" 消息。
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
- [使用引擎基座应用托管本地模型](llm-base-apps.md)
- [官方 AnythingLLM 文档](https://docs.anythingllm.com/)
