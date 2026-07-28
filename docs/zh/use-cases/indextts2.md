---
outline: deep
description: 在 Olares 上使用 IndexTTS2 从文本生成自然语音，支持零样本语音克隆。提供一段短音频样本，输入你的文本，即可获得与参考语音匹配的语音。
head:
  - - meta
    - name: keywords
      content: Olares, IndexTTS2, text-to-speech, TTS, voice cloning, zero-shot, AI speech synthesis
app_version: "1.0.6"
doc_version: "1.0"
doc_updated: "2026-04-20"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/indextts2.md)。
:::

# 使用 IndexTTS2 克隆语音

IndexTTS2 是一个零样本文本转语音（TTS）系统，可以从短音频参考生成自然语音。它将说话人身份与情感分离，让你可以独立控制语音音色、说话风格和语速。

在 Olares 上运行 IndexTTS2 可以将你的语音数据和生成的音频完全保留在你自己的硬件上。

使用本指南快速测试 IndexTTS2 的内置示例语音，从你自有的参考音频克隆语音，或调整生成结果的情感设置。

## 学习目标

在本指南中，你将学习如何：

- 安装 IndexTTS2。
- 使用内置示例语音或你自己的参考音频从文本生成语音。
- 调整情感设置以改变生成语音的情感基调。

## Prerequisites

- Olares 运行在配备 NVIDIA GPU 的设备上（最低 9 GB VRAM）。
- 设备使用 x86_64（amd64）处理器。
- 你有稳定的网络连接用于初始模型下载。

## 安装 IndexTTS2

1. 打开 Market，搜索 "IndexTTS2"。

   ![IndexTTS2](/images/manual/use-cases/indextts2.png#bordered){width=90%}

2. 点击 **Get**，然后点击 **Install**，等待安装完成。
3. 打开 IndexTTS2。

首次启动时，IndexTTS2 从 Hugging Face 下载所需的模型文件，然后在本地初始化它们。这可能需要几分钟，具体取决于你的网络速度和设备性能。

   ![IndexTTS2 模型下载](/images/manual/use-cases/indextts2-model-download.png#bordered){width=90%}

如果下载未完成或页面长时间卡住，请检查你的网络是否可以访问 Hugging Face，然后重新打开 IndexTTS2 并重试。

## 生成语音

IndexTTS2 提供两种入门方式：
- 使用内置示例语音快速测试应用。
- 使用你自己的参考音频克隆特定语音。

![IndexTTS2 界面](/images/manual/use-cases/indextts2-interface.png#bordered){width=90%}

### 使用示例语音

使用内置示例快速测试语音合成，无需准备你自己的音频。

1. 在 **Examples** 中，选择一个样本语音。
2. 在 **Text** 字段中，保留默认文本或输入你自己的文本。
3. （可选）要更改情感，请参阅[调整情感](#可选调整情感)。
4. 点击 **Synthesize**。

   ![从示例生成](/images/manual/use-cases/indextts2-example.png#bordered){width=90%}

生成完成后，结果将显示在输出音频播放器中。你可以直接在浏览器中播放，或点击 <i class="material-symbols-outlined">download</i> 下载。

### 使用你自己的参考音频

当你希望生成的语音与特定说话人匹配时，使用此选项。

1. 将参考音频文件上传到 **Voice Reference** 区域，或点击 <i class="material-symbols-outlined">mic</i> 录制一个。

   :::tip 选择良好的参考片段
   为获得最佳效果，请使用 5 到 15 秒的干净录音，背景噪音最小，且只有单个说话人。
   :::

2. 在 **Text** 字段中，输入你要合成的文本。
3. （可选）要更改情感，请参阅[调整情感](#可选调整情感)。
4. 点击 **Synthesize**。

   ![从自定义语音生成](/images/manual/use-cases/indextts2-custom.png#bordered){width=90%}

生成完成后，结果将显示在输出音频播放器中。你可以直接在浏览器中播放，或点击 <i class="material-symbols-outlined">download</i> 下载。

### 可选：调整情感

默认情况下，IndexTTS2 使用主参考音频中的情感。你可以在不更改任何情感设置的情况下生成语音。

如果你想更改情感，请展开 **Settings**，然后在 **Emotion control method** 下选择一种方法。

#### 使用情感参考音频

当你希望保留一个说话人的语音，但借用另一个片段的情感时使用此选项。

1. 选择 **Use emotion reference audio**。
2. 在 **Upload emotion reference audio** 中上传音频片段，或点击 <i class="material-symbols-outlined">mic</i> 录制一个。
3. 调整 **Emotion control weight** 以控制情感参考对生成语音的影响强度。

#### 使用情感向量

当你希望直接控制情感强度时使用此选项。

1. 选择 **Use emotion vectors**。
2. 调整一个或多个情感滑块，例如 **Happy**、**Angry**、**Sad**、**Afraid**、**Disgusted**、**Melancholic**、**Surprised** 和 **Calm**。
3. 调整 **Emotion control weight** 以控制情感的应用强度。

   ![使用情感向量](/images/manual/use-cases/indextts2-emotion-vectors.png#bordered){width=90%}

:::tip 逐渐调整情感权重
我们建议从大约 0.6 的值开始。较高的值增加情感强度，而较低的值保留原始语音的自然音色。
:::

## FAQs

### 为什么音频在文本说完之前就被截断了？

如果生成的音频在完整文本被说出之前停止，**Advanced generation parameter settings** 中的 **max_mel_tokens** 可能太低。

要解决此问题：
1. 展开 **Advanced generation parameter settings**。
2. 增加 **max_mel_tokens**。
3. 重新生成音频。
4. 如果文本很长，也稍微增加 **Max tokens per generation segment**，然后重试。

### 为什么长文本听起来断断续续或在尴尬的地方停顿？

长文本在生成之前被分成较小的段落。如果文本分割得太激进，结果可能听起来不够流畅，或在不自然的地方停顿。

要提高连续性：
1. 展开 **Advanced generation parameter settings**。
2. 查看 **Preview of the audio generation segments** 以了解文本是如何被分割的。
3. 逐渐增加 **Max tokens per generation segment**。
4. 重新生成音频并比较结果。

:::tip
如果文本很长，请考虑在生成之前手动将其分成较小的段落。
:::

### 为什么输出包含重复的单词或短语？

如果生成的语音不自然地重复单词或短语，当前的解码设置可能导致过多的变化。

要减少重复：
1. 展开 **Advanced generation parameter settings**。
2. 稍微增加 **repetition_penalty**。
3. 重新生成音频。
4. 如果重复仍然存在，请尝试稍微降低 **temperature**，然后再次测试。

逐渐调整这些值。大的更改可能会使结果听起来不够自然。

## 了解更多

- [IndexTTS2 on GitHub](https://github.com/index-tts/index-tts)：源代码和技术细节。
- [Speaches](speaches.md)：一个应用中集成语音转文本、文本转语音和语音聊天。
