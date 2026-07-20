---
outline: [2, 3]
description: 在 Olares 上安装 ACE-Step AI 的分步指南，使用歌词或器乐生成歌曲，通过 retake 和 repainting 优化音频，以及使用 Audio2Audio 将参考音频转化为新音乐。
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/ace-step.md)。
:::

# 使用 ACE-Step 创作你自己的 AI 音乐

ACE-Step 由 ACE Studio 和 StepFun 开发，是一个开源模型，可以根据你提供的歌词和风格标签生成音乐，让你通过简单的文本输入创作歌曲、人声和器乐。借助其内置工具，你还可以通过调整或重新生成特定部分来完善曲目，而无需从头开始。

本指南展示如何在 Olares 上安装 ACE-Step，生成你的第一首曲目，探索不同的音乐风格，以及使用应用内置的编辑功能增强音频。

## 学习目标

完成本教程后，你将学会如何：
- 在 Olares 设备上安装 ACE-Step。
- 使用歌词、标签和风格控制生成歌曲。
- 查找并下载生成的音频文件。
- 通过调整风格、编辑片段、延长歌曲或使用参考片段重塑来完善曲目。

## 前提条件

开始前，请确保：
- Olares 运行在配备 NVIDIA GPU 的机器上。

## 安装并设置 ACE-Step

准备好 Olares 设备后，按照以下步骤安装 ACE-Step 并开始生成音乐。

### 安装 ACE-Step

按照以下步骤安装 ACE-Step。

1. 在 Olares 网页界面中打开 **Market** 应用。
2. 使用搜索栏并输入 "ACE-Step"。
3. 点击 **Get**，然后点击 **Install**。
   ![ACE-Step 安装](/images/manual/use-cases/ace-step-install.png#bordered)
4. 等待几分钟，直到安装完成。

### 首次启动时下载所需模型

安装完成后，从 Launchpad 打开 ACE-Step。

Olares 将自动下载并安装所需模型。一个 **Download Manager** 窗口将出现，显示模型大小和下载进度。
   ![ACE-Step 下载管理器](/images/manual/use-cases/ace-step-download-manager.png#bordered){width=500}

下载完成后，ACE-Step 生成界面将自动打开。

## 生成你的第一首曲目

按照以下步骤设置参数并开始音乐生成。

### 设置基本参数

- **Audio Duration**: 拖动滑块选择曲目长度（最长 240 秒）。如果保持默认值（`-1`），音频长度将是随机的。
- **Format**: 从 `MP3`、`ogg`、`wav` 和 `flac` 中选择音频格式。
    :::tip 推荐 MP3
    建议将默认输出格式更改为 MP3。这将产生更小的文件大小、更快的加载速度和更好的用户体验。
    :::
- **Lora Name or Path**: 如果有可用的 LoRA 模型，请选择一个。目前仅支持中文说唱 LoRA。
- **Tags**: 输入风格、情绪、节奏或乐器的描述符，用逗号分隔。例如：
-
    ```plain
    Chinese Rap, J-Pop, Anime, kawaii future bass, Female vocals, EDM, Super Fast`
    ```
- **Lyrics**: 输入你的歌词，确保使用结构标签以获得最佳的组织和流畅性：
    - `[verse]` 用于主歌部分
    - `[chorus]` 用于副歌部分
    - `[bridge]` 用于桥段

    :::tip 生成纯器乐曲目
    在 **Lyrics** 区域输入标签 `[instrumental]` 或 `[inst]`。
    :::

### 开始生成

1.  设置好所有参数后，点击 **Generate**。
2.  生成完成后，点击 **Play** 按钮预览你的曲目。
   ![生成音频](/images/manual/use-cases/ace-step-generate.png#bordered)

### 保存生成的音乐

你可以通过两种方式保存生成的音乐：

- **直接下载**: 点击右上角 <i class="material-symbols-outlined">download</i> 按钮，直接将音频文件保存到本地设备。

- **从 Olares Files**：
    1. 打开 **Files**。
    2. 前往以下路径：`/Home/AI/output/acestepv2`。
    3. 右键点击生成的音频文件并将其保存到本地设备。


## 优化你的音频

ACE-Step 提供强大的工具来完善和修改生成曲目的特定部分。

### 重新生成整个片段

你可以生成整个曲目的新版本。

1. 点击 **retake** 选项卡。
2. 调整 **variance** 滑块以控制新版本的不同程度。数值越高，歌曲差异越大。
3. 点击 **Retake** 并等待生成。
4. 点击下方的 **Play** 按钮预览风格变化。
    ![预览 retake](/images/manual/use-cases/ace-step-retake.png#bordered)

### 重新生成特定片段

你可以只更新选定的时间范围，同时保持曲目的其余部分不变。

1. 点击 **repainting** 选项卡。
2. 调整 **Variance** 滑块以控制新生成中的变化程度。数值越高，歌曲差异越大。
3. 调整 **Repaint Start Time** 和 **Repaint End Time** 下方的滑块，设置你想要重新生成的片段时间段。
4. 选择 repainting 的源：
    - `text2music`: 通过 Text2Music 生成的原始歌曲。
    - `last_repaint`: 之前的 repainted 版本。
    - `upload`: 你上传的音频。
5. 点击 **Repaint** 并等待生成。
6. 点击下方的 **Play** 按钮预览结果。
    ![预览 repaint](/images/manual/use-cases/ace-step-repaint.png#bordered)

### 编辑歌词

你可以编辑歌词来修改特定行，而不会影响曲目的其余部分。

1. 点击 **edit** 选项卡。
2. 复制原始歌词并粘贴到 **Edit Lyrics** 区域。
3. 只修改你想要更改的特定歌词行。
4. 在 **Edit Type** 下，选择 `only_lyrics`。
5. 点击 **Edit** 并等待生成。
6. 点击下方的 **Play** 按钮预览变化。
    ![编辑歌词](/images/manual/use-cases/ace-step-edit-lyrics.png#bordered)

### 编辑标签

你可以编辑标签来重置曲目的风格或音色。

1. 点击 **edit** 选项卡。
2. 在 **Edit Tags** 区域输入新的风格或音色标签（例如 `hard rock` 或 `male tenor vocals`）。
3. 在 **Edit Type** 中，选择 `remix`。
4. 点击 **Edit** 并等待生成。
5. 点击下方的 **Play** 按钮预览变化。
    ![编辑标签](/images/manual/use-cases/ace-step-edit-tags.png#bordered)

### 延长音频

你可以通过在原始曲目之前或之后添加新音频来延长其长度。

1. 点击 **extend** 选项卡。
2. 调整 **Left Extend Length** 下方的滑块，在原始音频*之前*添加新生成。
3. 调整 **Right Extend Length** 下方的滑块，在原始音频*之后*添加新生成。
4. 选择延长的源：
    - `text2music`: 通过 Text2Music 生成的原始歌曲。
    - `last_extend`: 之前的延长版本。
    - `upload`: 你上传的音频。
5. 点击 **Extend** 并等待生成。
6. 点击下方的 **Play** 按钮预览变化。
    ![延长标签](/images/manual/use-cases/ace-step-extend.png#bordered)

## Audio2Audio

你可以基于上传的**参考音频**片段创建新曲目。它分析音色、节奏和风格等特征，生成具有相似感觉的曲目。
1. 勾选 **Enable Audio2Audio** 复选框。
2. 上传一个现有的音乐片段作为参考。
3. 调整 **Refer audio strength** 滑块。数值越高，生成的音乐与参考曲目越相似。
4. 选择一个 **Preset** 风格，或保持默认。
5. 根据需要设置其他参数。
6. 点击 **Generate** 创建与参考音频氛围相似的新音乐。
    ![Audio2Audio](/images/manual/use-cases/ace-step-audio2audio.png#bordered)
