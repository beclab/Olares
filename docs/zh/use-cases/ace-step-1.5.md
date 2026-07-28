---
outline: [2, 3]
description: 了解如何在 Olares 上安装 ACE-Step 1.5，通过提示词或歌词生成音乐，并使用参考音频、Cover、Repaint 和 LoRA 工作流来完善你的创意。
head:
  - - meta
    - name: keywords
      content: Olares, ACE-Step 1.5, AI music generation, text-to-music, audio editing, LoRA training
app_version: "1.0.1"
doc_version: "1.0"
doc_updated: "2026-04-09"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/ace-step-1.5.md)为准。
:::

# 使用 ACE-Step 1.5 创作 AI 音乐

ACE-Step 1.5 是一款 AI 音乐生成应用，可以将你的文本、歌词和音频引导转化为完整的歌曲。在 Olares 上，它运行在一个开箱即用的工作空间中，让你可以专注于音乐创作。

本指南涵盖生成、编辑和管理曲目的日常工作流程。

## 学习目标

完成本教程后，你将学会如何：

- 在 Olares 上安装并启动 ACE-Step 1.5。
- 理解两步生成工作流程。
- 完成从创意到成品曲目的端到端工作流程。
- 使用 Cover 功能重新演绎现有歌曲的风格。
- 使用 Repaint 功能修复曲目的特定片段。
- 查看、保存并继续迭代生成的结果。

## 前提条件

开始前，请确保：
- Olares 运行在配备 NVIDIA GPU 的设备上。
- 你的设备有足够的可用存储空间用于初始模型下载。
- 你有稳定的网络连接。

## 安装并启动 ACE-Step 1.5

1. 打开 Market 并搜索 "ACE-Step 1.5"。
    ![安装 Ace-Step 1.5](/images/manual/use-cases/install-ace-step-1.5.png#bordered){width=90%}

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

3. 打开 ACE-Step 1.5 并等待所需模型下载完成。根据网络条件，此过程可能需要一些时间。
   ![下载所需模型](/images/manual/use-cases/ace-step-1.5-download-models.png#bordered){width=60%}

当主工作空间出现时，ACE-Step 1.5 已准备就绪。

## 了解 ACE-Step 1.5 的工作原理

ACE-Step 1.5 通常分两个阶段工作：

1. 你描述音乐创意。
2. ACE-Step 将创意转化为生成输入，再转化为音频。

开始前，请使用下表作为常用控件的快速指南。

| 控件 | 用途 |
| --- | --- |
| **Song Description** | 在 **Simple** 模式下从高层次创意开始。 |
| **Music Caption** | 描述风格、乐器、情绪、人声和声音方向。 |
| **Create Sample** | 起草 **Music Caption**、**Lyrics** 和相关设置。 |
| **Generate Music** | 生成实际音频。 |
| **Reference Audio** | 添加风格引导，但不直接修改曲目。 |
| **Source Audio** | 提供 **Cover** 或 **Repaint** 使用的曲目。 |
| **Audio Cover Strength** | 控制 **Cover** 对原始结构的遵循程度。 |

## 创建并完善曲目

这是 **text2music** 最常用的工作流程。

1. 起草创意。

    a. 在 **Task Type** 中，选择 **text2music**。

    b. 在 **Generation Mode** 中，选择 **Simple**。

    c. 在 **Song Description** 中输入高层次创意。

    例如：
    ```text
    upbeat pop rock with electric guitars, driving drums, and catchy synth hooks
    ```

    d. 点击 **Create Sample**。
    ![创建样本](/images/manual/use-cases/ace-step-1.5-create-sample.png#bordered){width=90%}
2. 完善草稿。

    a. 查看 AI 在 **Music Caption** 和 **Lyrics** 框中生成的文本。
    ![编辑音乐描述](/images/manual/use-cases/ace-step-1.5-music-caption.png#bordered){width=90%}

    b. 根据需要编辑歌词，并确保包含 `[Verse]` 和 `[Chorus]` 等结构标签。

    c. 检查 **Music Caption** 和 **Lyrics** 的内容是否冲突。例如，如果你在描述中添加了 "acoustic guitar"，就不要在歌词中放入 `[Heavy Metal Guitar Solo]`。

3. 生成并试听。

    a. 点击 **Generate Music**。

    b. 在 **Results** 区域预览结果。
    ![查看结果](/images/manual/use-cases/ace-step-1.5-results.png#bordered){width=90%}
    :::tip
    始终生成几个变体。AI 音乐生成具有随机性。如果你不喜欢第一首曲目，点击 **Generate Music** 再次生成同一提示词的不同演绎。
    :::

4. 如需修改曲目。如果曲目接近预期但仍需调整：
    - 当你想保留大部分结构但改变风格时，使用 [Cover](#使用-cover-重新演绎现有曲目)。
    - 当你只想替换某个片段时，使用 [Repaint](#使用-repaint-重新生成曲目片段)。

5. 保存你想要保留的结果，然后从最有潜力的版本继续迭代。

## 使用更多控制选项生成

熟悉基本工作流程后，你可以探索更精确的控制选项。

### 在 Custom 模式下生成

当你想跳过起草步骤，直接输入自己的歌词和设置时，使用 **Custom** 模式。

1. 在 **Task Type** 中，选择 **text2music**。
2. 在 **Generation Mode** 中，选择 **Custom**。
3. 在 **Music Caption** 中填写目标风格、流派、乐器和情绪。
4. 可选地点击 **Format** 将简单的手写描述扩展为更丰富的描述。
5. 在 **Lyrics** 中输入你的文本。
6. 需要时设置 **BPM**、**Key Scale**、**Time Signature** 或 **Audio Duration** 等元数据。
    :::details 需要设置音乐元数据的帮助？

    如果你没有音乐理论背景，可以使用以下指南来自定义歌曲的情绪、速度和节奏：

    - **Key Scale (情绪):**
        - `Major` (例如 `C Major`): 明亮、阳光或令人振奋的曲目。
        - `Minor` (例如 `A Minor`): 悲伤、忧郁或冷酷的曲目。
    - **BPM (速度):**
        - `60–80`: 慢速民谣和 lo-fi。
        - `90–120`: 中速流行和摇滚。
        - `130–180`: 快节奏电子、trap 或高能摇滚。
    - **Time Signature (节奏/律动):**
        - `4`: 4/4 拍。流行、摇滚和大多数现代音乐的标准。如果不确定，这总是安全的选择。
        - `3`: 3/4 拍。营造经典的华尔兹或舞蹈节奏。
        - `2`: 2/4 拍。强劲有力，非常适合进行曲或快速乡村音乐。
        - `6`: 6/8 拍。营造轻柔的摇摆感，非常适合慢情歌或蓝调民谣。
    :::
7. 点击 **Generate Music**。
    ![自定义模式](/images/manual/use-cases/ace-step-1.5-custom-mode.png#bordered){width=90%}

### 使用参考音频添加风格引导

当你希望结果更紧密地遵循现有片段的感觉，而不直接修改该片段时，使用此功能。

1. 前往 **Audio Uploads**。
2. 上传一个片段到 **Reference Audio**，或使用麦克风图标录制一个。
    ![参考音频](/images/manual/use-cases/ace-step-1.5-reference-audio.png#bordered){width=90%}
3. 根据需要填写 **Music Caption** 和 **Lyrics**。
4. 点击 **Generate Music**。

## 修改现有曲目

### 使用 Cover 重新演绎现有曲目

当你想创建一首歌曲的新版本，同时保留其核心旋律结构和节奏时，使用 **Cover** 任务类型。

1. 在 **Task Type** 中，选择 **Cover**。
2. 在 **Audio Uploads** 中，将原始曲目上传到 **Source Audio**。如果你想继续处理刚刚生成的曲目，点击 **Results** 区域中的 **Send To Src Audio** 即可。
3. 输入描述你想要的新风格或声音的 **Music Caption**。
4. 在 **Advanced Settings** 中，调整 **Audio Cover Strength** 滑块。一般来说，较低的 **Audio Cover Strength** 值允许更多变化，而较高的值使结果更接近原始结构。
    ![使用 Cover 重新演绎](/images/manual/use-cases/ace-step-1.5-cover.png#bordered){width=90%}

5. 点击 **Generate Music**。

### 使用 Repaint 重新生成曲目片段

当只有曲目的某个特定片段需要更改时，使用 **Repaint** 任务类型。

1. 在 **Task Type** 中，选择 **Repaint**。
2. 在 **Audio Uploads** 中，将曲目上传到 **Source Audio**。
3. 设置 **Repainting Start** 和 **Repainting End** 以隔离需要重新生成的片段。如果希望编辑持续到曲目末尾，在 **Repainting End** 中使用 `-1`。
    ![使用 Repaint 重新生成](/images/manual/use-cases/ace-step-1.5-repaint.png#bordered){width=90%}

4. 输入描述更新片段应该听起来怎样的 **Music Caption**。
5. 点击 **Generate Music**。

## 查看、保存和复用结果

生成完成后：

1. 试听结果。
2. 比较几个版本。
3. 决定保留或完善哪个曲目。
4. 使用 **Results** 区域中的工具继续：
   - **Send To Src Audio**: 将当前结果直接移动到 **Source Audio** 槽位，以便立即开始 **Cover** 或 **Repaint** 任务。
   - **Apply These Settings to UI**: 将 promising 曲目的参数恢复到工作空间，以便生成类似的变体。
   - **Score**: 显示自动对齐分数，用于比较多个版本。
   - **Save**: 保留当前结果以供日后复用。

## 使用 LoRA 训练自定义风格

当你希望 ACE-Step 1.5 从你自己的数据集中学习更一致的风格时，使用 **LoRA Training**。这是一个高级工作流程，日常音乐生成不需要它。

:::info
训练需要至少 16 GB 的 VRAM。对于更长的歌曲，建议 20 GB 或更多。
:::

![LoRA 训练](/images/manual/use-cases/ace-step-1.5-lora-training.png#bordered){width=90%}

训练 LoRA：

1. 准备包含音频文件、歌词和注释的数据集。
2. 在 **Dataset Builder** 中，扫描或加载你的数据集。
3. 如有需要，查看并编辑检测到的元数据。
4. 保存数据集并将其预处理为张量。
5. 切换到 **Train LoRA** 选项卡并开始训练。如果你不熟悉训练参数，默认值通常就可以了。
6. 训练完成后，加载训练好的 LoRA 并在生成中使用它。

有关详细的数据集要求、参数参考和完整训练步骤，请参阅官方 [ACE-Step 1.5 LoRA training](https://github.com/ace-step/ACE-Step-1.5/blob/main/docs/en/LoRA_Training_Tutorial.md) 文档。

## 常见问题排查

### 生成缓慢或失败

生成速度取决于你的硬件和当前系统负载。

如果生成缓慢或失败：

- 等待当前任务完成。
- 尝试生成更短的片段。
- 关闭 Olares 上的其他重负载工作。

在某些设备上，有限的 GPU 内存也可能影响稳定性。

### 结果不符合预期

AI 音乐生成通常需要迭代。

如果结果不符合你的意图：

- 让你的 **Music Caption** 更具体。
- 确保你的 **Lyrics** 使用清晰的结构标签，如 `[Verse]` 和 `[Chorus]`。
- 检查 **Music Caption** 和 **Lyrics** 的内容是否冲突。
- 多次点击 **Generate Music** 以探索不同版本。

## 了解更多

有关生成、高级工作流程和 LoRA 训练的更多详细信息，请参阅官方 ACE-Step 文档：

- [ACE-Step 1.5 Ultimate Guide](https://github.com/ace-step/ACE-Step-1.5/blob/main/docs/en/Tutorial.md): 了解更多关于两步工作流程、参数行为和生成概念。
- [ACE-Step 1.5 — A Musician's Guide](https://github.com/ace-step/ACE-Step-1.5/blob/main/docs/en/ace_step_musicians_guide.md): 探索提示词创意、硬件指导和结构标签的实用技巧。
- [ACE-Step 1.5 LoRA Training Tutorial](https://github.com/ace-step/ACE-Step-1.5/blob/main/docs/en/LoRA_Training_Tutorial.md): 遵循数据集准备和 LoRA 训练的完整工作流程。
