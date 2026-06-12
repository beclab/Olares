---
outline: deep
description: 使用 Olares 上的 Speaches 应用，在 Open WebUI 中启用语音转文字和文字转语音。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, STT, TTS, 语音, Speaches, 音频
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-14"
---

# 在 Open WebUI 中配置语音交互

通过将 Open WebUI 连接到 Speaches 应用，你可以启用免手动输入的语音交流。该集成提供语音转文字（STT）用于听写提示词，也提供文字转语音（TTS）用于朗读 AI 回复。
## 学习目标

在本指南中，你将学习如何：

- 从 Speaches 获取所需的 STT 和 TTS 配置信息。
- 配置 Open WebUI，使用 Speaches 作为音频后端。
- 验证语音转文字、文字转语音和连续语音模式。

## 前提条件

开始前，请确保已满足以下条件：

- 已安装并配置 [Open WebUI](openwebui.md)，且至少有一个可用的模型后端。
- 已安装 [Speaches](speaches.md#install-speaches)。
- 拥有 Open WebUI 实例的管理员权限。
- 大约 14 GB 可用 VRAM，用于同时运行 LLM、STT 和 TTS 模型。

## 获取 Speaches 配置信息

要连接 Open WebUI 和 Speaches，你需要获取 Speaches 共享端点 URL，并确认 Speaches 中使用的 STT 和 TTS 模型名称。

### 获取共享端点 URL

1. 打开 Olares 设置，然后前往**应用** > **Speaches**。
2. 在**共享入口**中，点击 **Speaches API**，并记录端点 URL。

   例如：`http://edd26bab0.shared.olares.com`。

   ![Speaches shared entrance](/images/manual/use-cases/openwebui-speaches-shared-entrance.png#bordered){width=70%}

### 查找模型和语音名称

1. 从启动台打开 Speaches。
2. 前往 **Speech-to-Text** 标签页，点击 **Model** 下拉列表，并记录默认 STT 模型名称 `Systran/faster-whisper-small`。
3. 前往 **Text-to-Speech** 标签页，点击 **Model** 下拉列表，选择默认 TTS 模型 `speaches-ai/Kokoro-82M-v1.0-ONNX`，然后记录模型名称。
4. 从 **Voice** 下拉列表中，为 AI 朗读回复选择一个语音，并记录语音名称。例如：`am_eric`。

   ![Text-to-speech generation](/images/manual/use-cases/speaches-tts.png#bordered){width=90%}

## 在 Open WebUI 中配置音频设置

1. 在 Open WebUI 中，点击你的头像图标，然后前往 **Admin Panel** > **Settings** > **Audio**。
2. 在 **Speech-to-Text** 区域中，指定以下设置：

   - **Speech-to-Text Engine**：选择 **OpenAI**。
   - **API Base URL**：输入 Speaches 共享端点 URL，并在末尾追加 `/v1`。例如：`http://edd26bab0.shared.olares.com/v1`。
   - **API Key**：输入任意文本。不要留空。
   - **STT Model**：输入你之前记录的 STT 模型名称，即 `Systran/faster-whisper-small`。

3. 在 **Text-to-Speech** 区域中，指定以下设置：

   - **Text-to-Speech Engine**：选择 **OpenAI**。
   - **API Base URL**：输入 Speaches 共享端点 URL，并在末尾追加 `/v1`。例如：`http://edd26bab0.shared.olares.com/v1`。
   - **API Key**：输入任意文本。不要留空。
   - **TTS Voice**：输入你之前记录的语音名称。例如：`am_eric`。
   - **TTS Model**：输入你之前记录的 TTS 模型名称，即 `speaches-ai/Kokoro-82M-v1.0-ONNX`。

   ![Audio settings in Open WebUI](/images/manual/use-cases/openwebui-audio-settings.png#bordered)

4. 点击 **Save**。

## 验证配置

分别测试各项音频功能，确保集成正常工作。

:::tip 在新标签页中运行 Open WebUI 以使用音频
现代浏览器会阻止在 Olares 桌面窗口内运行的应用访问麦克风。为避免收到 "Permission denied" 错误，请选择 Open WebUI 窗口右上角的 <i class="material-symbols-outlined">open_in_new</i>，在新的浏览器标签页中打开应用。请在该新标签页中完成以下测试。
:::

### 测试语音转文字

1. 在 Open WebUI 中开始一个新聊天。
2. 选择模型。
3. 点击消息输入框旁边的 <i class="material-symbols-outlined">mic</i>。

   ![Dictate button](/images/manual/use-cases/openwebui-dictate-button.png#bordered)

4. 浏览器提示时，允许麦克风访问。
5. 对着麦克风说话。你的语音会被转写到文本框中。

### 测试文字转语音

1. 向模型发送消息并等待回复。
2. 点击回复下方的 <i class="material-symbols-outlined">volume_up</i>。回复内容会被朗读出来。

   ![Read aloud](/images/manual/use-cases/openwebui-read-aloud.png#bordered)

### 测试连续语音模式

1. 在聊天界面中，点击 <i class="material-symbols-outlined">graphic_eq</i>。首次加载时，模型初始化可能需要一些时间。

   ![Voice mode](/images/manual/use-cases/openwebui-voice-mode.png#bordered)

2. 自然说话。系统会转写你的语音、生成回复，并自动朗读回复。

:::warning 资源占用
使用音频功能会同时调用 LLM、STT 和 TTS 模型。确保设备有足够的 VRAM 和内存，以便三个模型顺利加载和切换。如果资源不足，Olares 可能会为保护系统而停止应用，导致短暂不可用。

在生产环境中，建议将 GPU 模式设置为**显存切片**，以避免模型之间争抢资源。
:::

:::tip 非英语语音
默认 STT 和 TTS 模型对非英语语言的效果可能不佳。如有需要，你可以在 Speaches Playground 中切换到其他模型。关于更换模型的说明，可参阅[在 Speaches 中管理模型](speaches.md#manage-models)。
:::
