---
outline: deep
description: 在 Olares 上安装并使用 Open Notebook，收集来源、生成 AI 洞察、与知识库聊天、创建笔记，以及从研究材料生成播客。
head:
  - - meta
    - name: keywords
      content: Olares, Open Notebook, AI notebook, research assistant, sources, notes, RAG, knowledge base, podcast, transformations
app_version: "1.0.4"
doc_version: "2.0"
doc_updated: "2026-07-22"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/open-notebook.md)为准。
:::

# 使用 Open Notebook 构建研究笔记

Open Notebook 是一个 AI 驱动的研究工作空间，用于收集来源材料、生成结构化洞察、与知识库聊天，以及将研究成果转化为可编辑的笔记或播客节目。

本指南以 AI 研究项目为例，带你完成第一个完整的 Open Notebook 工作流。你可以将相同的工作流应用于论文、课程、会议记录、市场研究、产品研究或其他主题。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Open Notebook。
- 设置用于聊天、摘要、检索和播客生成的 AI 模型。
- 创建研究笔记。
- 添加并处理研究来源。
- 查看 AI 生成的洞察。
- 与研究材料聊天。
- 将有用的 AI 回复保存为可编辑的笔记。
- 从选定的来源和笔记生成播客节目。

## 前提条件

开始前，请确保你可以访问想要使用的模型。

- 必需：至少一个语言模型和一个用于向量搜索的嵌入模型。
- 播客生成必需：一个文本转语音（TTS）模型。
- 可选：一个语音转文本（STT）模型，用于处理音频或视频来源。

:::info 推荐的本地 AI 服务
对于 Olares 上的本地 AI 工作流，你可以使用本地语言或嵌入模型服务，并使用 [Speaches](speaches.md) 提供语音转文本和文本转语音功能。
:::

## Open Notebook 的工作原理

Open Notebook 围绕四种主要内容类型组织你的工作：

| 内容类型 | 描述 |
| :-- | :-- |
| **Notebook** | 一个研究主题或项目的工作空间。 |
| **Source** | 添加到 Open Notebook 的原始材料，例如文件、网页、音频、<br>视频或粘贴的文本。 |
| **Insight** | 通过转换从来源创建的 AI 生成输出，例如<br>摘要或关键要点。 |
| **Note** | 保存在笔记中的可编辑知识。笔记可以手动编写、<br>从 AI 回复保存，或从洞察创建。 |

在本指南中，你将创建一个示例 AI 研究笔记，添加来源、生成洞察、与材料聊天、保存笔记，并创建播客。

## 安装 Open Notebook

1. 打开 Market 并搜索 "Open Notebook"。

   ![Open Notebook in Market](/images/manual/use-cases/open-notebook.png#bordered){width=90%}

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

安装完成后，在开始第一个研究笔记之前，配置所需的提供方和模型。

## 设置 AI 模型

Open Notebook 使用 AI 模型进行摘要、聊天、检索和播客生成。你只需要设置一次。

### 获取提供方连接信息

连接信息的获取方式取决于你要连接独立模型，还是另一个 Olares 应用。

#### 连接独立模型

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

对于本指南中使用的每个独立模型：

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

本例使用 `qwen3.5-9b` 和 `qwen3-embedding:0.6b`。Open Notebook 通过 **Ollama** 提供方连接它们，因此请在各自的模型控制台中查看 **Ollama** 格式并复制对应的 Base URL。

#### 连接应用

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

本指南使用 Speaches 作为 TTS 和 STT 提供方：

1. 前往 Olares **Settings** > **Applications** > **Speaches** > **Entrances**。
2. 选择 **Speaches API**，然后复制 **Endpoint** URL。

### 添加提供方配置

前往 **Manage** > **Models**。针对每个服务，找到匹配的提供方并点击 **Add Configuration**。

| 服务 | 提供方 | Base URL |
| :-- | :-- | :-- |
| Qwen 语言模型 | **Ollama** | 对应模型控制台中的 Base URL |
| Qwen 嵌入模型 | **Ollama** | 对应模型控制台中的 Base URL |
| Speaches | **OpenAI Compatible** | Speaches endpoint，并在末尾添加 `/v1` |

输入容易识别的配置名称和 Base URL。如果必须填写 API key，请输入 `olares`，然后保存配置。

### 添加模型

在每项配置中点击 **Models**，然后添加以下模型：

| 配置 | 类型 | Model ID |
| :-- | :-- | :-- |
| Qwen 语言模型 | **Language** | `qwen3.5-9b` |
| Qwen 嵌入模型 | **Embedding** | `qwen3-embedding:0.6b` |
| Speaches | **TTS** | `speaches-ai/Kokoro-82M-v1.0-ONNX` |
| Speaches | **STT** | `Systran/faster-whisper-small` |

### 分配默认模型

在 **Default Model Assignments** 下按如下方式分配模型：

| 插槽 | 模型 |
| :-- | :-- |
| Chat Model | `qwen3.5-9b` |
| Embedding Model | `qwen3-embedding:0.6b` |
| Text-to-Speech Model | `speaches-ai/Kokoro-82M-v1.0-ONNX` |
| Speech-to-Text Model | `Systran/faster-whisper-small` |
| Transformation Model | `qwen3.5-9b` |
| Tools Model | `qwen3.5-9b` |
| Large Context Model | `qwen3.5-9b` |

如果 **Auto-assign Defaults** 可用，可以用它自动填充插槽，然后检查选择。

![Model assignments](/images/manual/use-cases/open-notebook-set-models-result1.png#bordered)

## 创建你的第一个研究笔记

笔记是一个主题、项目、课程或研究问题的工作空间。在本指南中，你将创建一个关于生成式 AI 的学习笔记。

1. 前往 **Process** > **Notebooks**。
2. 点击 **New** > **Notebook**。
3. 输入笔记名称和描述。
    :::tip 撰写有用的描述
    笔记描述有助于 AI 理解项目的上下文。请尽可能清楚地描述主题、目的和预期用途。
    :::
4. 点击 **Create New Notebook**。

![Create a notebook](/images/manual/use-cases/open-notebook-create-a-notebook.png#bordered){width=70%}

## 添加你的第一个来源

来源是你希望 Open Notebook 处理的原始材料。对于首次运行，请使用轻量级文本来源，以避免因外部网站、大型 PDF 或不可用的视频转录而导致失败。

本指南提供了关于生成式 AI 的示例文本。将其添加为 **Text** 来源。首次工作流成功后，你可以添加更多文本来源或尝试外部 URL、PDF、YouTube 视频或音频/视频文件。

### 添加文本来源

1. 打开你刚刚创建的笔记。
2. 在 **Source** 区域，点击 **Add Source** > **Add Source**。
3. 点击 **Enter Text** 标签页。
4. 粘贴以下内容：
   ```plain
   Generative AI is a type of artificial intelligence that can create new content based on patterns learned from data. It can generate text, images, audio, video, code, and other forms of content. In everyday work, generative AI is often used to draft documents, summarize long materials, rewrite text for different audiences, brainstorm ideas, create outlines, and answer questions.

   A common example is using a language model to help write a product document. The user can provide rough notes, requirements, or meeting records, and the model can turn them into a clearer draft. The user still needs to review the result, check accuracy, and decide whether the writing fits the intended audience.

   Generative AI is useful because it can reduce the time spent on repetitive writing and analysis tasks. It can also help users explore unfamiliar topics by explaining concepts, comparing viewpoints, and suggesting follow-up questions.

   However, generative AI has limitations. It may produce inaccurate or unsupported statements. It may miss important context. It may also sound confident even when the answer is incomplete or wrong. For this reason, users should treat generative AI as an assistant rather than a final authority.

   A reliable workflow is to combine AI output with source verification. Users can collect original materials, generate summaries, ask targeted questions, save useful answers as notes, and manually review the final result before using it in real work.
   ```

5. 输入以下标题：
   ```plain
   What generative AI can do
   ```
6. 点击 **Next**。
7. 如果提示，将来源链接到你的笔记。
8. 点击 **Next** 打开处理设置。
9. 在 **Transformations** 下，为首次运行选择 **Dense Summary**。

   你可以根据目标选择不同的转换：
   
   | 转换 | 使用场景 |
   |:---------------|:------------|
   | **Dense Summary** | 你想要一个紧凑、信息丰富的来源概述。<br>推荐用于首次运行。 |
   | **Simple Summary** | 你想要一个更短的摘要，以便详细阅读。 |
   | **Key Insights** | 你想要主要要点、声明或发现。 |
   | **Paper Analysis** | 你正在处理学术论文。 |
   | **Reflection Questions** | 你想要用于讨论或后续研究的问题。 |
   | **Table of Contents** | 你想要了解长来源的结构。 |

10. 保持 **Enable search vector embedding** 选中。
11. 点击 **Done**。

![Add first source](/images/manual/use-cases/open-notebook-add-first-source.png#bordered){width=70%}

Open Notebook 开始处理来源。处理完成后，你可以将其用于洞察、聊天、笔记和引用。

### 添加其他来源类型

首次工作流成功后，你可以以类似方式添加其他材料。

| 来源类型 | 支持的内容 |
| :-- | :-- |
| **Upload file** | 文档、图片、压缩包和媒体文件。<br>音频或视频文件需要语音转文本模型。 |
| **Add URL** | 网页和其他支持的在线内容。 |
| **Enter text** | 直接粘贴或键入的内容。 |

:::warning 避免大量处理
使用本地模型时，请先处理一个来源并使用一个转换。同时处理多个大型来源或应用多个转换可能会导致处理缓慢、超时或任务失败。

你可以稍后从来源的 **Insights** 标签页使用 **Generate New Insight** 生成额外的洞察。
:::

## 查看生成的洞察

洞察是从来源创建的 AI 生成输出。例如，**Dense Summary** 帮助你在详细阅读之前快速了解来源的内容。

### 查看洞察

1. 打开一个已处理的来源。
2. 点击 **Insights** 标签页。
3. 点击 **View Insight** 查看生成的洞察。

![Review insight](/images/manual/use-cases/open-notebook-review-insight.png#bordered){width=90%}

使用洞察来决定来源是否有用，以及是否应将其包含在笔记聊天中。

### 生成另一个洞察

如果你想以另一种方式分析同一来源：

1. 打开已处理的来源。
2. 点击 **Insights** 标签页。
3. 在 **Generate New Insight** 下，选择一个转换。

   ![New insight](/images/manual/use-cases/open-notebook-new-insight.png#bordered){width=90%}

4. 点击 **New**。

## 与研究材料聊天

在你的来源处理完成后，你可以基于笔记中的材料提问。

1. 打开你的笔记。
2. 在 **Chat with Notebook** 中，选择你要使用的模型。
3. 点击每个来源旁边的图标，选择 AI 可以使用每个来源的多少内容：
      | 图标 | 上下文级别 | 推荐用途 |
      | :-- | :--  | :-- |
      | <i class="material-symbols-outlined">news</i> | 完整内容 | 用于最重要的来源，当你需要详细答案和引用时。 |
      | <i class="material-symbols-outlined">lightbulb_2</i> | 仅洞察 | 用于背景来源，当摘要足够时。 |
      | <i class="material-symbols-outlined">visibility_off</i>| 不包含在聊天中| 用于不相关、敏感或不必要的来源。  |
   
4. 输入你的问题并发送。

![Chat with AI](/images/manual/use-cases/open-notebook-chat.png#bordered){width=90%}

Open Notebook 根据当前聊天上下文中包含的来源回答。

:::tip 验证引用
当答案包含引用时，点击它们打开引用的来源段落。将答案与原始内容进行比较，以检查 AI 回复是否得到你的来源支持。
:::

## 创建笔记

笔记是可编辑的项目，用于摘要、大纲、问题、草稿或结论。

### 将 AI 答案保存为笔记

当你在聊天中收到有用的答案时：

1. 点击 AI 回复下方的 <i class="material-symbols-outlined">save</i> 图标。
2. 在 **Notes** 区域，点击带有 `AI Generated` 标签的保存笔记以查看它。
  
   ![AI generated note](/images/manual/use-cases/open-notebook-ai-note.png#bordered){width=70%}

3. 在需要时更新标题或内容，然后点击 **Save Note**。

你可以将保存的笔记用作未来笔记上下文的一部分，或将其包含在播客生成中。

### 手动创建笔记

你也可以手动创建笔记。

1. 打开你的笔记。
2. 前往 **Notes** 区域。
3. 点击 **Write Note**。
4. 输入标题并编写笔记内容。支持 Markdown。
5. 点击 **Create Note**。

你的笔记将出现在 **Notes** 区域，带有 `Human` 标签。

![Manually created note](/images/manual/use-cases/open-notebook-manual-note.png#bordered){width=70%}

## 生成播客

在你拥有来源、洞察和笔记后，你可以将研究材料转化为播客节目。

播客生成需要：

- 一个用于大纲生成的语言模型。
- 一个用于转录生成的语言模型。
- 一个用于音频生成的文本转语音模型。
- 已处理的来源或笔记，用作上下文。

### 配置播客配置文件

在生成播客之前，在播客配置文件中配置所需的模型和语音。

1. 前往 **Create** > **Podcasts**，然后点击 **Profiles** 标签页。
2. 打开任何标记为 **Needs Configuration** 或带有警告图标的配置文件。
3. 对于 **Speaker Profile**，选择一个语音模型并输入该模型支持的语音 ID。

   :::warning 使用所选 TTS 模型支持的语音 ID
   扬声器配置文件中的默认语音 ID 可能不受你选择的 TTS 模型支持。例如，`nova` 在 `speaches-ai/Kokoro-82M-v1.0-ONNX` 中不可用。请使用支持的 Kokoro 语音 ID，例如 `af_heart`。
   :::

4. 对于 **Episode Profile**，选择扬声器配置文件、大纲模型、转录模型、语言、片段数和简介。
5. 保存你的更改。

### 生成音频

1. 点击 **Episodes** 标签页，然后点击 **Generate Podcast**。
2. 选择要包含的来源或笔记。
3. 选择剧集配置文件。
4. 设置剧集名称。
5. 如果需要，添加额外说明。
   :::tip 匹配语言和语音
   某些 TTS 语音最适合特定语言。请确保播客语言与所选语音匹配。如果你使用英语语音，请添加额外说明，例如：`Generate the entire podcast script in ENGLISH only.`
   :::
6. 点击 **Generate**。

![Generate podcast](/images/manual/use-cases/open-notebook-generate-podcast.png#bordered){width=90%}

剧集完成后，你可以：

- 在浏览器中播放。
- 下载音频文件。
- 在 **Details** 中查看生成的转录。

![Generated podcast](/images/manual/use-cases/open-notebook-podcast-result.png#bordered){width=90%}


## 探索更多功能

### 搜索你的知识库

当你想要跨来源和笔记查找信息时，前往 **Process** > **Ask and Search**。

#### 提问

当你想要综合答案时，使用 **Ask**。

1. 打开 **Process** > **Ask and Search**。
2. 点击 **Ask** 标签页。
3. 输入你的问题。

   示例：

   ```plain
   Based on my sources, what are the main benefits and risks of using generative AI in business?
   ```

4. 点击 **Ask**。

Open Notebook 根据知识库中的匹配内容返回综合答案。

![Ask a question](/images/manual/use-cases/open-notebook-ask-result.png#bordered){width=90%}

#### 搜索来源片段

当你想要自己检查匹配片段时，使用 **Search**。

1. 打开 **Process** > **Ask and Search**。
2. 点击 **Search** 标签页。
3. 选择搜索类型：
   - **Text Search**：当你记得确切的单词或短语时使用。
   - **Vector Search**：当你记得含义但不记得确切措辞时使用。
4. 输入你的查询。

:::warning 需要嵌入模型
向量搜索需要配置的嵌入模型。来源在处理期间也需要启用嵌入。如果嵌入缺失或配置不正确，向量搜索可能返回无有用的结果。
:::

### 自定义转换

转换是可重用的 AI 提示，将来源内容转化为结构化洞察，例如摘要、关键要点、论文分析或反思问题。

:::tip 如何应用转换
使用 **Manage** > **Transformations** 查看、编辑、测试或创建模板。

要将转换应用于来源：
- 添加来源时选择转换。
- 对于现有来源，打开来源，前往 **Insights** 标签页，然后点击 **Generate New Insight**。
:::

你可以编辑现有转换或创建新转换。

#### 编辑现有转换

当内置转换接近你的需求时使用此方法。

1. 前往 **Manage** > **Transformations**。
2. 找到你要调整的转换，然后点击 **Edit**。
3. 修改标题、描述或提示。
4. 点击 **Edit Transformation** 保存你的更改。

编辑转换仅更改下次应用时它的工作方式。现有洞察不会自动更新。要获取新结果，请从来源的 **Insights** 标签页再次运行转换。

#### 创建新转换

当你想要为特定分析任务创建单独的模板时使用此方法。

1. 前往 **Manage** > **Transformations**。
2. 点击 **Create New**。
3. 输入名称、标题、描述和提示。
4. 点击 **Create New**。

   ![New transformation](/images/manual/use-cases/open-notebook-new-trans.png#bordered){width=90%}

#### 测试转换

在将转换应用于真实来源之前，先用短样本测试它。

1. 在 **Manage** > **Transformations** 中，找到你要测试的转换。
2. 点击 **Playground**。
3. 粘贴来源的短摘录。
4. 运行转换并查看输出。
5. 如果输出不符合你的预期，点击 **Edit** 优化提示，然后再次测试。

:::tip 在应用于完整来源之前测试
使用 **Playground** 在将转换应用于完整来源之前检查输出格式、长度和准确性。
:::

## 常见问题

### 为什么处理缓慢或不稳定？

处理缓慢或不稳定有两个常见原因：

- 你正在同时处理多个来源或应用多个转换。
- 其他 GPU 密集型应用正在使用 Open Notebook 或其配置的模型服务所需的 GPU 资源。

要提高性能：

1. 先处理一个来源并使用一个转换。
2. 停止或暂停其他 GPU 密集型应用。仅保持 **Manage** > **Models** 中分配的模型服务运行。有关详细信息，请参阅 [管理 GPU 资源](/zh/manual/olares/settings/gpu-resource.md)。

### 为什么向量搜索返回无有用的结果？

向量搜索需要：

- 配置的 **Embedding Model**。
- 处理时启用嵌入的来源。

如果向量搜索返回无有用的结果：

1. 前往 **Manage** > **Models** 并检查 **Embedding Model** 分配。
2. 确保所选模型是嵌入模型，而不是语言模型。
3. 检查来源是否使用 **Enable search vector embedding** 处理。
4. 重新处理来源。

### 为什么播客生成失败？

如果所需模型缺失、所选来源未准备就绪、TTS 提供方不可用、转录语言与所选语音不匹配，或扬声器语音 ID 不受所选 TTS 模型支持，播客生成可能会失败。

检查以下内容：

- 剧集配置文件具有大纲模型和转录模型。
- 扬声器配置文件具有有效的 TTS 模型。
- 每个扬声器使用所选 TTS 模型支持的语音 ID。
- 所选来源已处理并准备就绪。
- 播客语言与所选语音匹配。
- TTS 提供方正在运行。

如果错误消息列出了支持的语音 ID，请使用其中一个 ID 更新扬声器配置文件，然后创建新的播客生成任务。

## 了解更多

- [Ollama](ollama.md)：在 Olares 上下载并运行本地语言模型。
- [Speaches](speaches.md)：设置本地语音转文本和文本转语音服务。
- [管理 GPU 资源](/zh/manual/olares/settings/gpu-resource.md)：为本地 AI 应用分配 GPU 资源。
