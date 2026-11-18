---
outline: [2, 3]
description: 在 Olares 上使用 Music Studio 创作歌曲，试听生成结果，并下载音频用于自己的项目。
head:
  - - meta
    - name: keywords
      content: Olares, Music Studio, ACE-Step, AI 音乐, 歌词创作, 纯音乐, 本地音乐生成
app_version: "0.1.77"
doc_version: "1.0"
doc_updated: "2026-11-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/music-studio.md)为准。
:::

# 使用 Music Studio 创作歌曲

Music Studio 是一款基于 ACE-Step XL-SFT 的 AI 音乐创作应用。你可以用歌词和风格描述生成带人声的歌曲，也可以创作纯音乐。完成的歌曲会保存在 Olares 曲库中，供你播放和下载。

本文以创作一首名为 *Morning Light* 的 30 秒歌曲为例。你可以替换示例歌词和风格，创作适合自己项目的音乐。下文保留英文界面的按钮和字段名称，方便对照操作。

## 前提条件

- Olares 1.12.7 或更高版本，且 Router 已安装并运行。
- 配备 NVIDIA GPU 和 24 GiB 显存的 amd64 (x86-64) 设备。
- 至少 40 GiB 可用磁盘空间。

## 安装 Music Studio

1. 打开 Market，搜索“Music Studio”。

   <!-- ![Music Studio 在 Market 中的页面](/images/manual/use-cases/music-studio.jpg#bordered) -->

2. 点击 **Get**，再点击 **Install**，等待安装完成。

:::info 首次启动
首次打开 Music Studio 时，应用会下载约 30 GB 的模型文件。保持网络连接，等待准备完成并进入创作页面。
:::

## 创作音乐

你可以从自己的歌词、歌曲想法或纯音乐描述开始创作。

:::tip 同一时间生成一首歌曲
Music Studio 同一时间只能生成一首歌曲。任务运行时可以点击 **Cancel** 取消。生成时间取决于歌曲时长、质量设置和模型加载情况。
:::

### 使用自己的歌词创作歌曲

1. 打开 **Create**，在 **Lyrics** 中输入以下歌词：

   ```text
   [Verse]
   Morning light across the bay
   Let the shadows drift away
   [Chorus]
   Take a breath and start again
   Every road can find a friend
   ```

2. 在 **Music style** 中输入：

   ```text
   Hopeful acoustic pop, warm piano, gentle guitar and light drums,
   clear English vocals, clean studio production
   ```

   ![为 Morning Light 输入歌词和音乐风格](/images/manual/use-cases/music-studio-create-inputs.jpg#bordered)

3. 将 **Title** 设为 `Morning Light`，将 **Language** 设为 **English**。
4. 在 **Song duration** 中选择 **Fixed**，将 **Duration** 设为 30 秒。此示例包含人声，请保持 **Instrumental** 关闭。
5. 点击 **Create**。Music Studio 会先整理歌词，再在 **In progress** 中显示音频生成任务。等待完成的歌曲出现在 **Recent songs** 中。

### 根据想法创作歌曲

1. 打开 **Create**，在 **Song idea** 中输入简短的歌曲想法。例如：

   ```text
   A hopeful morning song with piano and a simple memorable chorus
   ```

2. 保持 **Instrumental** 关闭，选择语言和时长，然后点击 **Write with AI**。
3. 检查生成的 **Lyrics**、**Music style**、**Title**，以及表单下方的写作建议。如果歌词对所选时长来说过长或重复，请删减。如果音乐描述不符合你的想法，请直接修改。

   ![检查并修改 AI 生成的歌词和音乐风格](/images/manual/use-cases/music-studio-ai-draft.jpg#bordered)

4. 点击 **Create**，等待完成的歌曲出现在 **Recent songs** 中。

如果应用提示歌词重复，可以修改歌曲想法后重试，也可以直接输入自己的歌词继续创作。

### 创作纯音乐

1. 打开 **Create**，开启 **Instrumental**。开启后，**Lyrics** 字段将不可用。
2. 在 **Music style** 中描述乐器、氛围和编排。例如：

   ```text
   Gentle piano and acoustic guitar, warm hopeful mood, light drums,
   clean instrumental arrangement
   ```

3. 在 **Song duration** 中选择 **Fixed**，将 **Duration** 设为 30 秒。

   ![开启 Instrumental 并设置 30 秒时长](/images/manual/use-cases/music-studio-instrumental-options.jpg#bordered)

4. 点击 **Create**，等待完成的歌曲出现在 **Recent songs** 中。

## 播放和管理歌曲

1. 歌曲出现在 **Recent songs** 后，点击封面上的播放按钮。歌曲会开始播放，同时打开详情。你可以使用底部播放器暂停或调整播放位置。

   ![从 Recent songs 播放 Morning Light](/images/manual/use-cases/music-studio-playback.png#bordered)

2. 在歌曲详情中查看歌词和音乐风格。如果带人声的歌曲有歌词时间信息，还可以显示同步歌词。

   ![查看歌曲详情和歌词](/images/manual/use-cases/music-studio-details.jpg#bordered)

:::info 局部重绘
当前版本暂不支持 **Repaint section**。如果想修改歌曲的某一部分，请调整歌词或音乐风格后重新生成。
:::

关闭详情，点击 **Recent songs** 中歌曲旁的 <i class="material-symbols-outlined">more_horiz</i>，然后选择所需操作：

- **Download：** 将音频保存为 WAV 文件。
- **Like：** 收藏歌曲，之后可以在曲库中用 **Liked** 筛选。
- **Reuse seed：** 将歌曲的歌词或纯音乐设置，以及保存的音乐风格载入创作表单。检查标题、语言和时长，修改歌词或风格后点击 **Create**。调整其他设置后，即使复用 seed，也不保证生成相同的音频。
- **Rename：** 修改歌曲标题。
- **Delete：** 删除不再需要的歌曲。如果想保留音频，请先下载。

![从歌曲菜单中选择操作](/images/manual/use-cases/music-studio-download.jpg#bordered)

## 微调音乐

创作前，展开 **Advanced options** 调整声音和音乐方向。第一次创作可以保留默认设置。需要进一步控制时，参考下表：

| 设置 | 选择参考 |
|:--|:--|
| **Quality profile** | <ul><li><code>XL-SFT · 64 steps · ADG</code>：使用 64 步生成并启用 ADG。</li><li><code>XL-SFT · 50 steps</code>：使用 50 步生成，不启用 ADG。</li></ul> |
| **Production profile** | <ul><li><code>Clean studio</code>：引导模型生成更干净的录音室声音。</li><li><code>Textured ambience</code>：引导模型生成更有氛围感的声音。</li></ul> |
| **Style handling** | <ul><li><code>Auto enhance</code>：扩展简短的风格描述。</li><li><code>Strictly preserve</code>：更贴近你输入的描述。</li></ul> |
| **BPM** | 设置速度。较低的值通常听起来更慢，较高的值通常更快。 |
| **Key** | 指定调性，例如 `C major` 或 `A minor`。 |
| **Time signature** | <ul><li><code>Auto</code>：由模型决定。</li><li><code>4/4</code>、<code>3/4</code>、<code>6/8</code> 或 <code>2/4</code>：指定拍号。</li></ul> |
| **Vocal type** | 输入简短的人声描述，例如 `warm female lead`。留空则由模型决定。 |
| **Section structure** | 描述段落顺序，例如 `short intro, verse, chorus, outro`。短歌的段落结构尽量简洁。 |

这些设置用于引导生成，不能保证每项结果都完全符合预期。每次只调整少量设置，便于比较不同作品。

## 在曲库中查找歌曲

打开 **Library** 查找保存的歌曲：

- **搜索：** 在搜索框输入标题、音乐风格或歌词。
- **筛选：** 选择 **Liked** 或 **Instrumental**，也可以使用模型筛选器。
- **排序：** 选择 **Newest first** 或 **Oldest first**。

曲库中的歌曲旁也有相同的 <i class="material-symbols-outlined">more_horiz</i> 菜单。

![在 Library 中查找和整理歌曲](/images/manual/use-cases/music-studio-library.png#bordered)

## 常见问题

### 为什么提示没有可用的音乐模型？

启动或恢复应用后，请等待模型服务完成加载。如果创作页面仍提示没有可用模型，可以刷新页面，并检查 Router 是否正在运行。已有模型任务进行时，请等待任务结束后再开始新的创作。

### 为什么没有同步歌词？

歌词高亮依赖该首人声作品返回的时间信息。没有时间信息时，仍可阅读歌词、播放或下载音频。纯音乐没有演唱歌词。

## 了解更多

- [Olares Router](olares-router.md)：管理 Olares 上的 AI 模型可用性和路由。
