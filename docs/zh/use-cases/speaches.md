---
outline: [2, 3]
description: 在 Olares 上安装 Speaches，实现语音转文本、文本转语音和 AI 语音聊天。使用 OpenAI-compatible API 将语音服务与其他应用集成。
head:
  - - meta
    - name: keywords
      content: Olares, Speaches, speech-to-text, text-to-speech, STT, TTS, voice chat, OpenAI-compatible, Whisper, Kokoro
app_version: "1.0.7"
doc_version: "1.0"
doc_updated: "2026-04-14"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/speaches.md)为准。
:::

# 使用 Speaches 搭建语音服务

Speaches 是一个兼容 OpenAI API 的语音服务器，支持语音转文本（STT）和文本转语音（TTS）。它预装了模型，开箱即用，也可以轻松作为任何支持 OpenAI SDK 的应用的即插即用后端。

本指南将介绍如何在 Olares 上安装和使用 Speaches，包括语音转文本、文本转语音、Audio Chat、API 访问以及基本的模型管理。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Speaches。
- 使用语音转文本功能转录或翻译音频文件。
- 使用文本转语音功能将文本生成语音。
- 使用 Audio Chat 与 AI 模型进行语音对话。
- 从其他应用访问 Speaches API。
- 管理语音模型。

## 前提条件

- Olares 运行在一台带有 NVIDIA GPU 的设备上。
- [Ollama 已安装并运行](ollama.md)，且至少下载了一个聊天模型（仅 Audio Chat 需要）。

## 安装 Speaches

1. 打开 Market 并搜索 "Speaches"。

   ![Speaches in Market](/images/manual/use-cases/speaches.png#bordered){width=95%}

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

安装完成后，你将在 Launchpad 上看到两个图标：
- Speaches：语音转文本、文本转语音和音频聊天的主界面。
- Speaches Terminal：用于管理模型的命令行终端。

:::info  首次启动时的模型设置
首次打开 Speaches 时，它会下载并初始化内置模型。根据你的网络连接情况，此过程可能需要一些时间。

如果初始化在 30 分钟内未完成，可能会超时并自动取消。如果发生这种情况，请等待网络连接稳定后，再次打开 Speaches 以重试初始化。
:::

## 使用 Speaches

Speaches 预装了两个开箱即用的模型：

| Model | Type | Purpose |
|:------|:-----|:--------|
| `Systran/faster-whisper-small` | STT | 语音识别和翻译 |
| `speaches-ai/Kokoro-82M-v1.0-ONNX` | TTS | 语音合成 |

### 转录音频

1. 打开 Speaches 并点击 **Speech-to-Text** 标签页。
2. 在 **Model** 下，选择一个 STT 模型，例如 `Systran/faster-whisper-small`。
3. 在 **Task** 下，选择 **transcribe**。
4. 上传一个音频文件，或点击 <i class="material-symbols-outlined">mic</i> 从麦克风录制音频。
5. （可选）如果你希望在转录过程中接收部分结果，请启用 **Stream**。
6. 点击 **Generate**。

   ![Speech-to-text transcription](/images/manual/use-cases/speaches-stt-transcribe.png#bordered){width=90%}

处理完成后，转录文本将显示在 **Textbox** 中。

### 将音频翻译为英文

Speaches 可以自动检测音频语言并将其翻译为英文。

1. 打开 Speaches 并点击 **Speech-to-Text** 标签页。
2. 在 **Model** 下，选择一个 STT 模型，例如 `Systran/faster-whisper-small`。
3. 在 **Task** 下，选择 **translate**。
4. 上传一个音频文件，或点击 <i class="material-symbols-outlined">mic</i> 从麦克风录制音频。
5. （可选）如果你希望在翻译过程中接收部分结果，请启用 **Stream**。
6. 点击 **Generate**。
   ![Speech-to-text translation](/images/manual/use-cases/speaches-stt-translate.png#bordered){width=90%}

处理完成后，英文翻译将显示在 **Textbox** 中。

### 从文本生成语音

1. 打开 Speaches 并点击 **Text-to-Speech** 标签页。
2. 在 **Input Text** 中输入要转换的文本。
3. 在 **Model** 下，选择一个 TTS 模型。
4. 在 **Voice** 中选择一个语音。
5. 在 **Response Format** 下，选择一个输出格式。
6. 点击 **Generate Speech**。

   ![Text-to-speech generation](/images/manual/use-cases/speaches-tts.png#bordered){width=90%}

7. 播放生成的音频，并在需要时下载。

### 使用语音与 AI 聊天

使用 **Audio Chat** 可以通过语音、文本或音频文件与 AI 模型对话。Speaches 首先将你的语音转换为文本，然后将文本发送给聊天模型，并可以将回复转换回语音。


:::info
- Audio Chat 需要 Ollama 已安装，且至少下载了一个聊天模型。
- 音频播放目前仅支持英文回复。对于其他语言，回复仅以文本形式显示。
:::

#### 开始语音对话

1. 打开 Speaches 并点击 **Audio Chat** 标签页。
2. 在 **Chat Model** 下，选择一个 Ollama 模型，例如 `qwen2.5:7b`。
3. 使用以下任一方式发送消息：
   - **Audio file**：上传一个音频文件。
   - **Text**：在麦克风图标旁的输入框中输入你的消息并发送。
   - **Voice**：点击 <i class="material-symbols-outlined">mic</i> 录制你的消息，然后点击 <i class="material-symbols-outlined">send</i> 发送。

   ![Audio Chat interface](/images/manual/use-cases/speaches-audio-chat.png#bordered){width=90%}

4. 等待 Speaches 生成回复。

:::warning
完整的语音流程（STT、LLM、TTS）需要一定时间来完成。在回复生成过程中请勿刷新页面，否则可能会看到 UI 闪烁。
:::

#### 可选：提高 Audio Chat 的转录准确性

Audio Chat 默认使用预装的 `Systran/faster-whisper-small` 语音转文本模型。为了获得更好的转录准确性，你可以切换到更大的模型，例如 `Systran/faster-whisper-large-v3`。

:::info  可能需要更多 GPU 资源
更大的模型需要更多的 GPU 资源。如果在切换到更大的模型后生成任务开始失败，请参阅[为什么切换到更大的模型后任务会失败](#为什么切换到更大的模型后任务会失败)。
:::

1. 打开 Speaches Terminal 并下载模型：

   ```bash
   hf download Systran/faster-whisper-large-v3
   ```
   
   如果你看到关于 `HF_TOKEN` 的警告，可以忽略它。即使不设置此项，模型下载仍然可以继续。

2. 前往 **Settings** > **Applications** > **Speaches** > **Manage environment variables**。
3. 点击 `SPEACHES_WHISPER_MODEL` 旁边的 <i class="material-symbols-outlined">edit_square</i>。
4. 将值设置为你下载的模型，例如 `Systran/faster-whisper-large-v3`，然后点击 **Confirm**。
   ![Update STT model](/images/manual/use-cases/speaches-update-stt-model.png#bordered){width=90%}

5. 点击 **Apply** 保存更改。

Speaches 会自动重启以应用更改。

:::tip 等待服务初始化
应用显示为运行状态后，请再等待一段时间再使用，因为服务可能仍在初始化中。
:::

<!-- ## Use the Speaches API

This section is for connecting other apps to Speaches. If you only want to use Speaches in its own interface, you can skip this section.

Speaches is fully compatible with the OpenAI API format. Any app that supports the OpenAI SDK can call it directly.-->

<!-- ### Get the endpoint

<Tabs>
<template #Access-within-Olares>

1. Go to **Settings** > **Applications** > **Speaches**.
2. Under **Shared entrances**, click **Speaches API**.
3. On the **Set up endpoint** page, copy the URL next to **Endpoint**. 

   For example:
   ```
   http://edd26bab0.shared.olares.com
   ```

![Speaches shared entrance](/images/manual/use-cases/speaches-shared-entrance.png#bordered){width=90%}

</template>
<template #Access-outside-of-Olares>

Before you start, make sure [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) is enabled on your device.
1. Go to **Settings** > **Applications** > **Speaches**.
2. Under **Entrances**, click **Speaches API**.
3. On the **Endpoint settings** panel, copy the URL next to **Endpoint**. 

   For example:
   ```
   https://a8259cf22.laresprime.olares.com
   ```

![Speaches API entrance](/images/manual/use-cases/speaches-api-entrance.png#bordered){width=90%}

</template>
</Tabs> -->

<!-- ### Connect Open Notebook to Speaches

You can use Speaches as the STT/TTS backend for Open Notebook.

1. Open Open Notebook and navigate to its settings page.
2. Add a new model provider:
   - **Name**: Name your configuration (for example, `Speaches`).
   - **Base URL**: Enter the shared entrance URL with `/v1` appended. For example:
     ```plain
     http://d54536a50.shared.olares.com/v1
     ```

3. Configure STT and TTS separately under the new provider, selecting the models you have downloaded in Speaches.

   ![Configure STT/TTS in Open Notebook](/images/manual/use-cases/speaches-open-notebook-config.png#bordered) -->

## 管理模型

当你想使用不同的模型、提高质量或释放存储空间时，可以管理模型。

### 查看已下载的模型

要查看所有已下载的模型，请打开 Speaches Terminal 并运行：

```bash
hf cache list
```

### 下载新模型

1. 打开 Speaches Terminal 并运行：

   ```bash
   hf download <model-name>
   ```

   例如：

   ```bash
   # 下载更大的 Whisper 模型以获得更高的准确性
   hf download Systran/faster-whisper-medium
   # 准确性最高的 Whisper 模型，需要更多显存
   hf download Systran/faster-whisper-large-v3
   ```

   :::tip 共享模型存储
   模型下载到 Olares Files 的 `/Home/Huggingface/speaches/` 目录。如果 Olares 上的其他应用也使用 Hugging Face 模型，它们会共享此目录。
   :::

2. 刷新 Speaches 页面以将新模型加载到列表中。

### 删除模型

要释放存储空间，你可以删除不再需要的模型：

1. 打开 Speaches Terminal 并运行：
```bash
hf cache rm model/<model_name>
```

例如：

```bash
hf cache rm model/Systran/faster-whisper-medium
```

2. 刷新 Speaches 页面以更新模型列表。

## 切换到 CPU 模式

Speaches 默认使用 GPU 模式。如果需要，你可以切换到 CPU 模式。CPU 模式速度较慢，主要适用于小型任务。

要切换到 CPU 模式：

1. 前往 **Settings** > **Applications** > **Speaches** > **Manage environment variables**。
2. 点击 `SPEACHES_GPU` 旁边的 <i class="material-symbols-outlined">edit_square</i>，将其值更改为 `false`，然后点击 **Confirm**。

   ![Switch to CPU mode](/images/manual/use-cases/speaches-cpu-mode.png#bordered){width=90%}

3. 点击 **Apply** 保存更改。

Speaches 会自动以 CPU 模式重新部署。与 GPU 模式相比，处理速度会更慢。

## 常见问题

### 为什么 Audio Chat 显示错误？

Audio Chat 需要 Ollama 正在运行，且至少下载了一个聊天模型。如果 Ollama 未安装或没有可用模型，Audio Chat 将显示错误。

要修复此问题，请安装 Ollama 并按照 [Ollama 指南](ollama.md)下载一个聊天模型。Speaches 会自动检测 Ollama，因此无需重启 Speaches。

### 为什么切换到更大的模型后任务会失败？

此问题通常发生在 GPU 处于 **Memory slicing** 模式时。

更大的模型需要更多的 VRAM。如果 Speaches 只分配了少量 VRAM，在切换到更大的模型后，生成任务可能会失败。

要修复此问题：
- 在 **Memory slicing** 模式下增加分配给 Speaches 的 VRAM。
- 或将 GPU 切换到其他模式。

详细说明请参阅[管理 GPU 资源](/zh/manual/olares/settings/gpu-resource.md)。

### 我可以为 Audio Chat 使用不同的 Ollama 实例吗？

可以。更新部署配置中的 `CHAT_COMPLETION_BASE_URL`：

1. 打开 Control Hub 并导航到 **Browse** > **System** > **speachesserver-shared** > **Deployments** > **speaches**。
2. 点击 <i class="material-symbols-outlined">edit_square</i> 编辑 YAML 文件。

   ![Navigate to Speaches deployment](/images/manual/use-cases/speaches-controlhub-deployment.png#bordered){width=90%}

3. 在 **Edit YAML** 中，找到 `CHAT_COMPLETION_BASE_URL`，并将其值更新为你的 Ollama 端点。确保 URL 以 `/v1` 结尾。
   
   ![Edit CHAT_COMPLETION_BASE_URL](/images/manual/use-cases/speaches-edit-base-url.png#bordered){width=90%}

4. 前往 **Settings** > **Applications** > **Speaches**，点击 **Stop**，然后点击 **Resume** 以重启 Speaches。

## 了解更多

- [Speaches 官方文档](https://speaches.ai/)：完整的 API 参考和模型兼容性说明。
- [Ollama](ollama.md)：下载和运行本地 AI 模型。
- [Open WebUI](openwebui.md)：可以使用 Speaches 作为语音后端的聊天界面。
- [IndexTTS2](indextts2.md)：通过零样本语音克隆从文本生成语音。
