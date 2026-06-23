---
outline: [2, 3]
description: 学习如何在 Olares 上使用 Whisper-WebUI 进行语音转文字转录、字幕生成、实时录音、字幕翻译和人声分离，支持 96 种语言。
head:
  - - meta
    - name: keywords
      content: Olares, Whisper-WebUI, speech to text, transcription, subtitles, AI, self-hosted, vocal separation
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-28"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/whisper-webui.md)。
:::

# 使用 Whisper-WebUI 转录音频和视频

Whisper-WebUI 是一个基于浏览器的语音转文字工具，用于从音频、视频、YouTube 链接和麦克风录音生成转录文本和字幕文件。它还包含独立的字幕翻译和人声/背景音乐分离工具。

使用本指南在 Olares 上转录媒体文件，使用可选过滤器改进转录结果，翻译现有字幕文件，以及在需要时分离人声和背景音乐。

## 学习目标

在本指南中，您将学习如何：

- 在 Olares 上安装 Whisper-WebUI。
- 转录本地文件、YouTube 视频和麦克风录音。
- 通过移除背景音乐、VAD 和说话人分离来改进转录结果。
- 翻译字幕并分离人声和背景音乐。
- 查找生成的文件并处理自动下载失败时的模型下载。

## 安装 Whisper-WebUI

1. 打开 Market 并搜索 "Whisper-WebUI"。
   ![Install Whisper-WebUI](/images/manual/use-cases/whisper-webui.png#bordered){width=80%}

2. 点击 **Get**，然后 **Install**，等待安装完成。

安装后，您将在 Launchpad 上看到两个图标：

- Whisper-WebUI：用于转录、字幕翻译和背景音乐分离的主界面。
- Whisper-WebUI Terminal：用于管理模型的命令行终端。

## 了解基础知识

### 主要工作流程

Whisper-WebUI 包含五个主要选项卡，分为两类：转录和独立工具。

| 选项卡 | 类型 | 最适合 |
| :--- | :--- | :--- |
| **File** | 转录 | 为本地媒体文件生成字幕，最大 500 MB。 |
| **Youtube** | 转录 | 通过 URL 转录在线视频，无需手动下载。 |
| **Mic** | 转录 | 直接在浏览器中录音并转录。 |
| **T2T Translation** | 独立工具 | 翻译现有字幕文件。 |
| **BGM Separation** | 独立工具 | 导出单独的人声和伴奏音轨。 |

![Whisper-WebUI interface](/images/manual/use-cases/whisper-webui-interface.png#bordered)

### 选择输出格式

使用转录选项卡时，根据您计划如何使用结果来选择输出格式。

| 格式 | 最适合 |
|:--|:--|
| SRT | 视频播放器和编辑器的标准字幕文件。 |
| WebVTT | 网页视频字幕和基于浏览器的播放。 |
| TXT | 无时间戳的纯文本转录。 |
| LRC | 音乐播放器和音频应用的同步歌词。 |

### 选择转录模型

对于大多数任务，从 `large-v2` 开始。它已预装，适用于一般转录。

当您需要不同的速度、准确性和资源使用平衡时，更换模型：

| 需求 | 推荐模型 | 说明 |
| :--- | :--- | :--- |
| 更快的处理 | `small` 或 `medium` | 当大型模型太慢时使用。准确性可能较低。 |
| 最低资源使用 | `tiny` 或 `base` | 仅用于快速测试或简单音频。准确性有限。 |
| 更高的准确性 | `large-v3` | 更适合复杂、嘈杂或非英语音频，但使用更多资源。 |
| 更快的大型模型转录 | `large-v3-turbo` | 比 `large-v3` 更快，有一些准确性权衡。 |
| 仅英语音频 | 以 `.en` 结尾的模型 | 仅在源音频为英语时使用。 |

:::info 首次下载
只有 `large-v2` 是预装的。其他模型在您首次选择时会自动下载。下载可能需要一些时间，具体取决于您的网络和模型大小。
:::

## 转录音频和视频

**File**、**Youtube** 和 **Mic** 选项卡遵循相同的转录工作流程，并共享核心设置，例如模型、语言、输出格式和高级设置。

每次转录任务完成后，Whisper-WebUI 会在输出区域显示转录文本，并提供可下载的字幕或文本文件。生成的文件也会保存在 Files 中的 `/External/olares/ai/output/whisperwebui/` 下。

### 转录本地文件

1. 点击 **File** 选项卡。
2. 点击上传区域并选择音频或视频文件。文件大小限制为 500 MB。
3. 在 **Model** 下，选择转录模型。
4. 在 **Language** 下，指定源语言或使用 **Automatic Detection**。
   :::tip
   指定语言可以提高准确性，特别是对于短音频或非英语内容。
   :::
5. 在 **File Format** 下，选择您的首选输出格式。
6. 可选：展开下方面板以[移除背景音乐](#在转录前移除背景音乐)、[使用 VAD 检测语音](#使用-vad-检测语音片段)或[识别说话人](#识别多说话人音频中的说话人)。
7. 点击 **GENERATE SUBTITLE FILE**。

![File transcription result](/images/manual/use-cases/whisper-webui-file-result.png#bordered){width=90%}

### 转录 YouTube 视频

:::warning YouTube 访问限制
如果 YouTube 阻止自动访问或下载请求，或者网络环境无法访问该视频，YouTube 转录可能会失败。
:::

1. 点击 **Youtube** 选项卡。
2. 将 YouTube 视频 URL 粘贴到输入字段中。Whisper-WebUI 在可用时检测视频的缩略图、标题和描述。

   ![YouTube URL input](/images/manual/use-cases/whisper-webui-youtube-input.png#bordered){width=90%}

3. 在 **Model** 下，选择转录模型。
4. 在 **Language** 下，指定视频的语言。
5. 在 **File Format** 下，选择您的首选输出格式。
6. 可选：展开下方面板以[移除背景音乐](#在转录前移除背景音乐)、[使用 VAD 检测语音](#使用-vad-检测语音片段)或[识别说话人](#识别多说话人音频中的说话人)。
7. 点击 **GENERATE SUBTITLE FILE**。

![YouTube transcription result](/images/manual/use-cases/whisper-webui-youtube-result.png#bordered){width=90%}

### 使用麦克风录音并转录

:::info 麦克风访问要求
麦克风录音需要浏览器麦克风权限和 HTTPS 或 localhost 访问。如果录音不起作用，请检查浏览器权限和您打开 Whisper-WebUI 的方式。
:::

1. 点击 **Mic** 选项卡。
2. 点击录音按钮开始录音。您可以随时暂停。

   ![Mic recording](/images/manual/use-cases/whisper-webui-mic-recording.png#bordered){width=90%}

3. 点击 **Stop** 结束录音。您可以预览和修剪音频。

   ![Mic recorded](/images/manual/use-cases/whisper-webui-mic-recorded.png#bordered){width=90%}

4. 选择转录的 **Model**、**Language** 和 **File Format**。
5. 可选：展开下方面板以[移除背景音乐](#在转录前移除背景音乐)、[使用 VAD 检测语音](#使用-vad-检测语音片段)或[识别说话人](#识别多说话人音频中的说话人)。
6. 点击 **GENERATE SUBTITLE FILE**。

### 可选转录过滤器

**File**、**YouTube** 和 **Mic** 选项卡包含可选过滤器，可以针对特定音频类型改进结果。在点击 **GENERATE SUBTITLE FILE** 之前配置这些过滤器。

#### 在转录前移除背景音乐

当语音与音乐或背景音频混合时，使用此功能。

启用方法：

1. 展开 **Background Music Remover Filter**。
2. 勾选 **Enable Background Music Remover Filter**。
3. 除非需要自定义设置，否则保持默认模型和片段大小。

Whisper-WebUI 首先分离人声轨道，然后转录处理后的音频。

:::tip BGM 分离与背景音乐移除
**BGM Separation** 选项卡仅将音频分离为人声和伴奏音轨。它不会转录结果。

背景音乐移除是转录工作流程的一部分。它首先分离人声，然后转录人声轨道。
:::

#### 使用 VAD 检测语音片段

对于长录音、会议、播客或带有长静音片段的音频，使用 VAD。VAD 可以跳过静音、加速转录并减少来自静音音频的幻觉文本。

启用方法：

1. 展开 **Voice Detection Filter**。
2. 勾选 **Enable Silero VAD Filter**。

#### 识别多说话人音频中的说话人

说话人分离在转录文本中标记不同的说话人，例如 `SPEAKER_00` 和 `SPEAKER_01`。

首次使用前，完成 Hugging Face 设置，以便 Whisper-WebUI 可以下载所需的模型：

1. 展开 **Diarization** 面板。
2. 在 **HuggingFace Token** 字段下，点击提供的两个 `pyannote` 模型链接。
3. 在 Hugging Face 上，登录或创建免费账户，然后接受访问两个模型的条件。
4. 在您的 Hugging Face 账户设置中，创建具有 **Read** 权限的访问令牌。
5. 返回 Whisper-WebUI，勾选 **Enable Diarization**。
6. 将您的 Hugging Face 令牌粘贴到 **HuggingFace Token** 输入字段中。
7. 照常运行转录。

:::info 首次分离下载
首次启用说话人分离时，Whisper-WebUI 使用您的令牌下载所需的模型。这可能需要一些时间。模型下载后，可以在未来的转录中重复使用。
:::

## 使用独立工具

除了转录之外，Whisper-WebUI 还提供专门的字幕翻译和音频分离选项卡。

### 翻译字幕

使用 **T2T Translation** 选项卡翻译现有字幕文件。

Whisper-WebUI 提供两种翻译方法：

| 方法 | 要求 | 最适合 |
| :--- | :--- | :--- |
| **NLLB** | 首次使用时下载本地翻译模型。 | 无需外部 API 密钥的本地翻译。 |
| **DeepL API** | 需要 DeepL API 密钥。 | 使用 DeepL 的在线翻译。 |

<Tabs>
<template #使用-NLLB-翻译>

首次翻译可能需要几分钟，因为 Whisper-WebUI 需要先下载选定的 NLLB 模型。模型下载后，可以在以后重复使用。

1. 点击 **T2T Translation** 选项卡。
2. 上传您想要翻译的字幕文件。
3. 在 **NLLB** 子选项卡下，选择翻译模型。

   ![T2T translation model](/images/manual/use-cases/whisper-webui-t2t-model.png#bordered){width=90%}

4. 设置 **Source Language** 和 **Target Language**。
5. 点击 **TRANSLATE SUBTITLE FILE**。

完成后，您可以预览结果、下载生成的文件，或在 Files 中的 `/External/olares/ai/output/whisperwebui/` 查看它。

</template>

<template #使用-DeepL-API-翻译>

:::info 需要 API 密钥
您必须拥有有效的 DeepL API 密钥才能使用此方法。如果密钥缺失或无效，翻译将失败。
:::

1. 点击 **T2T Translation** 选项卡。
2. 上传您想要翻译的字幕文件。
3. 在 **DeepL API** 子选项卡下，输入您的 **DeepL API key**。
4. 如果您使用的是 DeepL Pro 订阅，勾选 **Pro account**。
5. 设置 **Source Language** 和 **Target Language**。
6. 点击 **TRANSLATE SUBTITLE FILE**。


完成后，您可以预览结果、下载生成的文件，或在 Files 中的 `/External/olares/ai/output/whisperwebui/` 查看它。

</template>

</Tabs>

### 分离人声和背景音乐

使用 **BGM Separation** 选项卡将音频文件拆分为单独的人声和伴奏音轨。此独立工具不会转录结果。

1. 点击 **BGM Separation** 选项卡。
2. 上传您想要处理的音频文件。
3. 在 **Device** 下，根据您的硬件选择处理设备。
4. 在 **Model** 下，选择分离模型。
5. 点击 **SEPARATE BACKGROUND MUSIC**。

![BGM separation result](/images/manual/use-cases/whisper-webui-bgm-result.png#bordered){width=90%}

完成后，您可以预览结果、下载生成的文件，或在 Files 中查看：

- 伴奏：
`/External/olares/ai/output/whisperwebui/UVR/instrumental`
- 人声：`/External/olares/ai/output/whisperwebui/UVR/vocals`

## 高级：从终端管理模型下载

大多数用户不需要手动管理模型。仅在自动模型下载失败、超时或您需要检查模型是否已下载时，才使用 Whisper-WebUI Terminal。

点击 Launchpad 上的 **Whisper-WebUI Terminal** 图标打开 Web 终端。

### 检查已下载的模型

打开 Whisper-WebUI Terminal，然后运行：

```bash
find /Whisper-WebUI/models -maxdepth 3 -type d
```

检查特定模型文件夹：

```bash
# Whisper 转录模型
ls -la /Whisper-WebUI/models/Whisper/faster-whisper/

# NLLB 翻译模型
ls -la /Whisper-WebUI/models/NLLB/

# UVR 背景音乐分离模型
ls -la /Whisper-WebUI/models/UVR/MDX_Net_Models/

# 说话人分离模型
ls -la /Whisper-WebUI/models/Diarization/
```

### 手动下载模型

虽然 Whisper-WebUI 自动下载模型，但您可以在 UI 下载超时时手动触发下载：

例如，要下载 Whisper 转录模型，将仓库名称替换为您需要的模型：
```bash
hf download Systran/faster-whisper-large-v3 \
  --cache-dir /Whisper-WebUI/models/Whisper/faster-whisper
```

下载完成后，刷新 Whisper-WebUI 并从模型列表中选择该模型。

## 常见问题

### 为什么 T2T Translation 使用 NLLB 时会失败？

如果模型下载中断或模型文件夹不完整，NLLB 翻译可能会失败。

重置 NLLB 下载：

1. 在 Files 中打开 `External/olares/ai/whisperwebui/NLLB/`。
2. 删除文件夹内的内容，但保留文件夹本身。
3. 返回 Whisper-WebUI，然后重新下载模型。

### 为什么说话人分离会失败？

说话人分离可能因以下原因失败：

- Hugging Face 令牌缺失或无效。
- 未接受所需的 pyannote 模型条款。
- 由于网络问题导致模型下载失败。

检查以下内容：
- Hugging Face 令牌具有 **Read** 权限。
- 您使用相同的 Hugging Face 账户接受了两个 pyannote 模型的条款。
- Whisper-WebUI 下载模型时您的网络稳定。

### 为什么切换到更大的模型后任务会失败？

切换模型后，任务可能因以下原因之一失败：

- 选定的模型尚未完成下载。
- 该模型需要比 Whisper-WebUI 当前拥有的更多的 GPU 内存。

解决此问题：

- 等待首次模型下载完成，然后重试任务。
- 在 Memory slicing 模式下为 Whisper-WebUI [分配更多 VRAM](/zh/manual/olares/settings/single-gpu.md#adjust-vram-allocation)。
- 将 GPU 切换到另一种合适的模式。
- 选择较小的模型。

## 了解更多
- [Open WebUI](openwebui.md)：将 Whisper-WebUI 用作聊天输入的语音转文字后端。
