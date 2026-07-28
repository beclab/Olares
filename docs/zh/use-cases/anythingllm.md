---
outline: [2, 3]
description: 在 Olares 上使用 AnythingLLM 构建私有知识库。添加文档、创建嵌入，并使用 RAG 进行查询。
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, self-hosted rag, private knowledge base, anythingllm ollama, local LLM, embedding, anythingllm on olares
app_version: "1.0.13"
doc_version: "2.0"
doc_updated: "2026-07-27"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/anythingllm.md)为准。
:::

# 使用 AnythingLLM 构建本地知识库

AnythingLLM 是一个开源的一体化 AI 应用，让你可以使用检索增强生成（RAG）与文档对话。它支持多个 LLM 提供方和向量数据库，并可在 Olares 设备上本地运行。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 AnythingLLM。
- 在 AnythingLLM 中配置聊天模型和嵌入提供方。
- 创建工作空间并上传文档以构建知识库。
- 使用自然语言查询你的知识库。

## 前提条件

开始前，你需要：

- 一台具有足够磁盘空间和内存的 Olares 设备。
- 从 Market 安装应用的管理员权限。
- 以下模型：

  | 模型类型 | 模型 | 获取方式 |
  | :--- | :--- | :--- |
  | 聊天 | Qwen3.6-27B (llama.cpp) | 从 Market 安装 |
  | 嵌入 | all-MiniLM-L6-v2 | AnythingLLM 内置的嵌入模型 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 AnythingLLM

1. 打开 Market 并搜索 "AnythingLLM"。

   ![安装 AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

对于 Qwen3.6-27B (llama.cpp)：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

## 配置 AnythingLLM

配置聊天模型和嵌入提供方。这些设置将成为所有工作空间的系统默认值。

### 设置聊天模型

1. 从 Launchpad 打开 AnythingLLM。
2. 在主页上，点击左下角的 **Open settings** 图标。
3. 在左侧边栏中，选择 **AI Providers** > **LLM**，然后选择 **Generic OpenAI** 作为 LLM 提供方。
4. 在 **Base URL** 中，粘贴 Qwen3.6-27B (llama.cpp) 模型控制台中的 Base URL。
5. 在 **Selected Model** 中选择 `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`。

   ![配置聊天模型](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered)

6. 点击 **Save changes**。显示 "LLM preferences saved successfully" 消息。

### 设置嵌入模型

1. 在左侧边栏中选择 **Embedder**，然后选择 **AnythingLLM Embedder** 作为嵌入提供方。
2. 在 **Model Preference** 中选择 **all-MiniLM-L6-v2** 作为嵌入模型。

   ![配置嵌入模型](/images/manual/use-cases/anythingllm-configure-embedding-model.png#bordered)

3. 点击 **Save changes**。显示 "Embedding preferences saved successfully" 消息。

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
