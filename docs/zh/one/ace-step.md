---
outline: [2, 3]
description: 在 Olares 上安装 ACE-Step AI 的分步指南，生成带歌词或纯器乐的歌曲，使用 retake 和 repainting 优化音频，以及使用 Audio2Audio 将参考音频转换为新音乐。
head:
  - - meta
    - name: keywords
      content: ACE-Step, AI 音乐
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/ace-step.md)。
:::

# 使用 ACE-Step 创作 AI 音乐 <Badge type="tip" text="15 min" />

ACE-Step 是由 ACE Studio 和 StepFun 开发的开源音乐生成模型。它可以根据歌词和风格标签生成音乐，并支持 retake、repainting 和 Audio2Audio 等编辑工具。

本指南将引导你在 Olares One 上完成安装、首次生成和基本编辑工作流。

## 学习目标

完成本教程后，你将学会如何：
- 使用歌词、标签和风格控制生成歌曲。
- 查找并下载生成的音频文件。
- 通过调整风格、编辑段落、扩展歌曲或使用参考片段重塑来优化音轨。

## 前提条件

开始之前，请确保你已具备：
- 连接到稳定网络的 Olares One。
- 足够的磁盘空间来下载模型。

## 安装并设置 ACE-Step

准备好 Olares 设备后，按照以下步骤安装 ACE-Step 并开始生成音乐。

### 安装 ACE-Step

按照以下步骤安装 ACE-Step。

1. 打开 Market，搜索 "ACE-Step"。
    ![安装 ACE-Step](/images/one/ace-market.png#bordered)

2. 点击 **Get**，然后点击 **Install**。  
3. 等待几分钟，直至安装完成。

### 首次启动时下载所需模型

安装完成后，从 Launchpad 打开 ACE-Step。

Olares 将自动下载并安装所需模型。一个 **Download Manager** 窗口将出现，显示模型大小和下载进度。  
   ![ACE-Step Download Manager](/images/manual/use-cases/ace-step-download-manager.png#bordered){width=500}

下载完成后，ACE-Step 生成界面将自动打开。

## 生成你的第一首音轨

按照以下步骤设置参数并开始音乐生成。

### 设置基本参数

- **Audio Duration**：拖动滑块选择音轨长度（最长 240 秒）。如果保持默认值（`-1`），音频长度将是随机的。
- **Format**：从 `MP3`、`ogg`、`wav` 和 `flac` 中选择音频格式。
    :::tip 推荐使用 MP3
    建议将默认输出格式更改为 MP3。这样可以获得更小的文件大小、更快的加载速度和更好的整体体验。
    :::
- **Lora Name or Path**：如果有可用的 LoRA 模型，请选择。目前默认只提供中文说唱 LoRA。
- **Tags**：输入风格、情绪、节奏或乐器的描述词，用逗号分隔。例如：

    ```plain
    Chinese Rap, J-Pop, Anime, kawaii future bass, Female vocals, EDM, Super Fast
    ```
- **Lyrics**：输入你的歌词。使用结构标签来组织歌曲流程：
    - `[verse]` 用于主歌。
    - `[chorus]` 用于副歌。
    - `[bridge]` 用于桥段。
    
    :::tip 生成**纯器乐**音轨
    在 **Lyrics** 区域输入标签 `[instrumental]` 或 `[inst]` 即可生成无人声的音乐。
    :::

### 开始生成

1.  设置好所有参数后，点击 **Generate**。 
2.  生成完成后，点击 **Play** 按钮预览你的音轨。
   ![生成音频](/images/manual/use-cases/ace-step-generate.png#bordered){width=80%}

### 保存生成的音乐

你可以通过两种方式保存生成的音乐。

<tabs>
<template #直接下载>

点击右上角的 <i class="material-symbols-outlined">download</i> 按钮，直接将音频文件保存到你的本地设备。
</template>
<template #从 Olares Files>

1. 打开 Files。
2. 导航至：`/Home/AI/output/acestepv2`。
3. 右键点击生成的音频文件并将其保存到你的本地设备。
</template>
</tabs>

## 优化你的音频

ACE-Step 提供强大的工具来优化和修改生成音轨的特定部分。

### 重新生成整段音轨

你可以生成整首音轨的新版本。

1. 点击 **retake** 标签页。
2. 调整 **variance** 滑块：
    - 较高的值：创建显著不同的版本。
    - 较低的值：使新版本更接近原版。
3. 点击 **Retake** 并等待生成。

    ![预览 retake](/images/manual/use-cases/ace-step-retake.png#bordered){width=90%}

### 重新生成特定段落

你可以只更新选定的时间范围，同时保持音轨其余部分不变。

1. 点击 **repainting** 标签页。
2. 调整 **variance** 滑块：
    - 较高的值：创建选定段落的显著不同版本。
    - 较低的值：使新段落更接近原版。
3. 调整 **Repaint Start Time** 和 **Repaint End Time** 滑块以选择你要重新生成的段落。
4. 选择 repainting 的源：
    - `text2music`：通过 Text2Music 生成的原始歌曲。
    - `last_repaint`：之前的 repainted 版本。
    - `upload`：你上传的音频。
5. 点击 **Repaint** 并等待生成。

    ![预览 repaint](/images/manual/use-cases/ace-step-repaint.png#bordered){width=90%}

### 编辑歌词

你可以编辑歌词来修改特定行，而不会影响音轨的其余部分。

1. 点击 **edit** 标签页。
2. 复制原始歌词并粘贴到 **Edit Lyrics** 区域。
3. 只修改你希望更改的特定行。
4. 在 **Edit Type** 下，选择 `only_lyrics`。
5. 点击 **Edit** 并等待生成。

    ![编辑歌词](/images/manual/use-cases/ace-step-edit-lyrics.png#bordered){width=90%}

### 编辑标签

你可以编辑标签来重置音轨的风格或音色。

1. 点击 **edit** 标签页。
2. 在 **Edit Tags** 区域输入新的风格或音色标签（例如 `hard rock` 或 `male tenor vocals`）。
3. 在 **Edit Type** 中，选择 `remix`。
4. 点击 **Edit** 并等待生成。

    ![编辑标签](/images/manual/use-cases/ace-step-edit-tags.png#bordered)

### 扩展音频

你可以通过在原始音频之前或之后添加新音频来扩展音轨长度。

1. 点击 **extend** 标签页。
2. 调整滑块：
    - **Left Extend Length**：在原始音频*之前*添加新音频。
    - **Right Extend Length**：在原始音频*之后*添加新音频。
3. 选择要扩展的源：
    - `text2music`：通过 Text2Music 生成的原始歌曲。
    - `last_extend`：之前的扩展版本。
    - `upload`：你上传的音频。
4. 点击 **Extend** 并等待生成。

    ![扩展标签](/images/manual/use-cases/ace-step-extend.png#bordered)

## Audio2Audio

你可以基于上传的**参考音频**片段创建新音轨。AI 会分析音色、节奏和风格等特征，以生成具有相似感觉的音轨。
1. 勾选 **Enable Audio2Audio**。
2. 上传一个现有的音乐片段作为参考。
3. 调整 **Refer audio strength** 滑块。较高的值会产生更接近参考音轨的音乐。
4. 选择一个 **Preset** 风格，或保持默认。
5. 根据需要设置其他参数。
6. 点击 **Generate** 创建与参考音频氛围相似的新音乐。

    ![Audio2Audio](/images/manual/use-cases/ace-step-audio2audio.png#bordered)
