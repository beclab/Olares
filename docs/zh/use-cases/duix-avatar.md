---
noindex: true
outline: [2, 3]
description: 学习在 Olares 上部署 Duix.Avatar，从模型训练到视频合成，创建文本驱动的数字人视频。
head:
  - - meta
    - name: keywords
      content: Olares, Duix.Avatar, HeyGem, digital avatar, AI avatar video, self-hosted digital human, Duix.Avatar on Olares
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/duix-avatar.md)为准。
:::

# 使用 Duix.Avatar 创建数字人

Duix.Avatar（前身为 HeyGem）是一个开源 AI 工具包，用于生成数字人，专注于离线视频创建和数字克隆。

本指南介绍如何在 Olares 上部署和使用 Duix.Avatar，涵盖从模型训练到视频合成的完整流程，以生成文本驱动的数字人视频。

## 学习目标

在本指南中，你将学习如何：
- 为数字人克隆准备和处理视频及音频素材。
- 使用 Olares 上的 Hoppscotch 调用 Duix.Avatar API 集合来训练模型、合成音频和创建视频。

## Prerequisites

在开始之前，请确保：
- Olares 1.11 或更高版本。
- Olares 运行在配备 NVIDIA GPU 的机器上。

## 安装 Duix.Avatar

1. 在 **Market** 中搜索 "Duix.Avatar"。
   ![Duix.Avatar](/images/manual/use-cases/duix-avatar.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 安装 Hoppscotch

除了 Duix.Avatar，你还需要 Hoppscotch，一个开源 API 开发环境，用于与 Duix.Avatar 服务交互。
1. 在 **Market** 中搜索 "Hoppscotch"。
   ![Hoppscotch](/images/manual/use-cases/hoppscotch.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 准备媒体文件

生成数字人需要一个源视频作为面部和声音的模板。你需要一段 10-20 秒的视频片段，视频中的人应面对镜头清晰说话。

然后你必须将此源视频分离为两个文件：无声视频和纯音频文件。本指南使用 `ffmpeg` 完成此步骤。

:::info 确保已安装 ffmpeg
要按照本指南使用 `ffmpeg` 命令，请确保它已安装在你的本地电脑上。参见 https://www.ffmpeg.org/download.html。
:::
1. 打开终端，`cd` 进入包含你视频的文件夹，然后运行以下命令：
   ```bash
   # 将 input.mp4 替换为你的实际文件名
    ffmpeg -i input.mp4 -c:v copy -an output_video.mp4 -c:a pcm_s16le -f wav output_audio.wav
   ```
   这将在同一文件夹中创建两个新文件：
   - `output_video.mp4`（无声视频）
   - `output_audio.wav`（音频）
2. Duix.Avatar 服务从特定目录读取文件。将刚刚生成的两个文件上传到 Olares **Files** 应用中的指定位置。
   1. 将 `output_audio.wav` 上传至：
   ```plain
   /Data/heygem/voice/data/ 
   ```
   ![上传源音频](/images/manual/use-cases/duix-avatar-upload-source-audio.png#bordered)

3. 将 `output_video.mp4` 上传至：
   ```plain
   /Data/heygem/face2face-data/temp/
   ```
   ![上传源视频](/images/manual/use-cases/duix-avatar-upload-source-video.png#bordered)

## 将 API 集合导入 Hoppscotch

预配置的 Hoppscotch 集合可用于简化 API 调用。
1. 在终端中运行以下命令下载 API 集合文件：
    ```bash
    curl -o duix.avatar.json https://cdn.olares.com/app/demos/en/duix/duix.avatar.json
    ```
2. 在 Olares 中打开 Hoppscotch 应用。
3. 在右侧的集合面板中，点击 **Import** > **Import from Hoppscotch**，然后选择你刚刚下载的 `duix.avatar.json` 文件。
   ![从 Hoppscotch 导入](/images/manual/use-cases/duix-avatar-import-from-hoppscotch.png#bordered)

导入后，你将看到一个名为 `duix.avatar` 的新集合，包含四个预配置的请求。
   ![检查集合](/images/manual/use-cases/duix-avatar-check-collection.png#bordered)

## 通过 API 训练数据

现在你将按顺序调用四个 API 来生成数字人。

:::tip
Duix.Avatar API 地址与你的 Olares ID 绑定。在以下所有 API 请求中，你必须将 URL 中的 `<OLARES_ID_PREFIX>` 替换为你自己的 Olares ID 前缀。例如，如果你的 Olares 访问 URL 是 `https://app.alice123.olares.com`，你的前缀就是 `alice123`。
:::

### 步骤 1：模型训练

此步骤预处理你上传的音频，提取特征以准备语音克隆。

1. 在 Hoppscotch 中，展开 `duix.avatar` 集合并选择 **1. Model training**。
2. 修改请求 URL，将 `<OLARES_ID_PREFIX>` 替换为你的 Olares ID 前缀。
   :::info
   请求体已预设为指向你上传的 `output_audio.wav` 文件，因此无需更改。
   :::
3. 点击 **Send** 开始预训练。
   成功的请求将返回 JSON 响应。复制 `reference_audio_text` 和 `asr_format_audio_url` 的值以备后用。
   ![预训练](/images/manual/use-cases/duix-avatar-pretrain.png#bordered)

### 步骤 2：音频合成

此步骤使用你在步骤 1 中训练的语音模型，从文本提示合成新音频。
1. 点击 **2. Audio synthesis**。
2. 修改请求 URL 中的 Olares ID。
3. 在请求体中，修改以下字段：
   * `text`：输入你希望数字人说的文本。
   * `reference_audio`：粘贴步骤 1 中的 `asr_format_audio_url` 值。
   * `reference_text`：粘贴步骤 1 中的 `reference_audio_text` 值。
   * 其他参数可以保持默认值。
   ![编辑音频参数](/images/manual/use-cases/duix-avatar-edit-audio-parameters.png#bordered)

4. 点击 **Send** 合成音频。成功的请求将返回一个音频文件。

5. 在响应区域中，点击 <span class="material-symbols-outlined">more_vert</span> 以 MP3 格式下载音频。
   ![生成音频文件](/images/manual/use-cases/duix-avatar-generate-audio-file.png#bordered)

6. 将下载的文件重命名为 `new.mp3`。在同一文件夹中，使用 `ffmpeg` 将其转换为 `.wav`：
    ```bash
   ffmpeg -i new.mp3 new.wav
   ```
7. 将新的 `new.wav` 文件上传至：
   ```plain
   /Data/heygem/face2face-data/temp/
    ```
   ![上传音频](/images/manual/use-cases/duix-avatar-upload-audio.png#bordered)

### 步骤 3：视频合成

现在你将使用合成的音频（`new.wav`）和原始无声视频（`output_video.mp4`）来合成最终的数字人视频。

1. 点击 **3. Video synthesis**。
2. 修改请求 URL 中的 Olares ID。
3. 在请求体中，将 `code` 字段更改为一个新的唯一任务标识符。你将使用此 ID 来检查合成进度。
   :::info
   请求体中的 `audio_url` 和 `video_url` 已预设为 `new.wav` 和 `output_video.mp4`，与你上传的文件匹配。无需更改。
   :::
4. 确认设置并点击 **Send**。成功的响应将返回 `"success": true`，表示任务已提交。
   ![提交任务](/images/manual/use-cases/duix-avatar-submit-task.png#bordered)

### 步骤 4：查询视频合成进度

视频合成是一个耗时的任务。使用此 API 查询其处理状态。
1. 点击 **4. Query progress**。
2. 修改请求 URL 中的 Olares ID。
3. 在 **Params** 部分，将 `code` 值更改为你步骤 3 中设置的唯一标识符。
4. 点击 **Send** 检查当前进度。
5. 重复此查询，直到响应中的 `progress` 字段达到 `100`，表示视频合成完成。
   ![任务完成](/images/manual/use-cases/duix-avatar-task-completed.png#bordered)
   :::tip
   视频合成所需时间取决于你的 GPU 性能和视频长度。可能需要几分钟或更长时间。
   :::
6. 成功后，响应中的 `result` 字段将包含输出视频的文件名。你可以在 Olares Files 应用的以下位置找到最终生成的视频：
    ```plain
   /Data/heygem/face2face-data/temp/
    ```
   ![在 Files 中检查视频](/images/manual/use-cases/duix-avatar-check-video-in-files.png#bordered)

## FAQ

### 进度卡住或合成失败

如果进度查询长时间停滞或 API 返回错误，请前往 Control Hub，找到名为 `heygemgenvideo` 的容器，并检查其日志以获取详细的错误信息。
![在 Control Hub 中检查 Duix.Avatar](/images/manual/use-cases/duix-avatar-check-in-controlhub.png#bordered)

### API 请求失败

请确认以下事项：
- 你已在请求 URL 中将默认的 Olares ID（`<OLARES_ID_PREFIX>`）正确替换为你自己的 ID。
- 所有媒体文件（`output_audio.wav`、`output_video.mp4`、`new.wav`）都已上传到正确的目录，且文件名完全匹配。

### 媒体已更新，但仍生成旧视频

确保你为视频合成使用了新的唯一 `code` 参数。系统会缓存结果，因此重复使用 `code` 将返回之前缓存的视频。
