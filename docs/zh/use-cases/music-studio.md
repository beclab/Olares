---
outline: [2, 3]
description: 在 Olares 上使用 Music Studio 创作歌曲，试听生成结果，并下载音频用于自己的项目。
head:
  - - meta
    - name: keywords
      content: Olares, Music Studio, ACE-Step, AI 音乐, 歌词创作, 纯音乐, 本地音乐生成
app_version: "0.1.76"
doc_version: "1.0"
doc_updated: "2026-11-12"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/music-studio.md)为准。
:::

# 使用 Music Studio 创作歌曲

Music Studio 是基于 ACE-Step XL-SFT 的 AI 音乐创作应用。你可以输入歌词和音乐风格描述，生成带人声的歌曲，也可以创作纯音乐。完成的作品会保存在你的 Olares 曲库中，供你播放和下载。

本文以创作一首名为 *Morning Light* 的 30 秒歌曲为例。你可以将示例歌词和风格替换成适合自己视频或音乐项目的内容。为便于对照截图，下文保留英文界面中的按钮和字段名称。

## 学习目标

- 使用自己的歌词和风格描述创作短歌。
- 使用 AI 起草歌词，并审阅和修改生成的建议。
- 播放、整理和下载完成的作品。

## 前提条件

- Olares 1.12.7 或更高版本，且 Router 已安装并运行。
- 配备 NVIDIA GPU 的 amd64 设备，GPU 显存为 24 GiB。这是当前应用安装包声明的要求。
- 足够的磁盘空间，用于首次下载约 30 GB 模型文件，以及保存生成的音频。

## 安装 Music Studio

1. 打开应用市场，搜索“Music Studio”。
2. 点击 **Get**，再点击 **Install**，等待安装完成。
3. 从启动台打开 Music Studio。首次启动时，模型准备页面会下载约 30 GB 的必要模型文件，并显示下载进度。等待下载和模型加载完成后，应用会自动进入创作页面。

应用已包含所需的音乐模型，创作时不需要选择模型。首次下载需要保持网络可用，并预留足够的磁盘空间。下载和加载耗时取决于网络与设备情况。后续启动可以复用已有的模型缓存。

## 创作歌曲

### 使用自己的歌词

1. 打开 **Create**，在 **Lyrics** 中输入：

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

   这段描述指定了充满希望的原声流行风格、温暖的钢琴、轻柔的吉他和鼓点，以及清晰的英语人声。

3. 将 **Title** 设为 `Morning Light`，将 **Language** 设为 **English**。
4. 在 **Song duration** 中选择 **Fixed**，将 **Duration** 设为 30 秒。这个示例需要人声，因此不要勾选 **Instrumental**。
5. 点击 **Create**。应用会先整理歌词，然后在 **In progress** 区域显示音频生成任务。等待完成的歌曲出现在 **Recent songs** 中。

![在 Music Studio 中填写歌词和音乐风格](/images/manual/use-cases/music-studio-create.png#bordered)

Music Studio 同一时间处理一个创作任务。如果要停止正在生成的歌曲，可以点击任务上的 **Cancel**。生成耗时会随歌曲时长、质量设置和模型加载情况变化。

### 让 AI 起草歌词

在 **Song idea** 中简要描述创作想法，选择语言和时长，然后点击 **Write with AI**。例如：

```text
A hopeful morning song with piano and a simple memorable chorus
```

这个想法描述了一首以钢琴伴奏、包含简洁易记副歌的清晨歌曲。你也可以换成自己的主题。

点击 **Create** 前，先检查生成的 **Lyrics**、**Music style** 和 **Title**。AI 建议可能出现重复内容，或者包含与目标语言不一致的音乐描述。你可以直接编辑这些字段，保留需要的部分。

如果应用提示歌词重复度过高，可以修改创作想法后重试，也可以手动输入歌词。一次写词失败不会阻止你使用自己的内容继续创作。

### 创作纯音乐

如果需要没有人声的背景音乐，勾选 **Instrumental**，在 **Music style** 中描述乐器、情绪和编排方式。设置时长后点击 **Create**。完成后，可以在 **Library** 中使用 **Instrumental** 筛选找到作品。

### 调整音乐方向

需要更多控制时，展开 **Advanced options**。**Quality profile** 提供 XL-SFT 的 50 步和 64 步 ADG 两种设置。**Production profile** 可选择干净的录音室风格或带有环境质感的声音。**Style handling** 用于选择增强风格描述，或更严格地保留原有描述。

你还可以填写 BPM、调性、拍号、人声类型和段落结构。这些设置用于引导生成，并不保证结果完全符合每项要求。建议每次只调整少量设置，便于比较不同版本。

## 试听和保存作品

1. 点击已完成歌曲旁的播放按钮。使用底部播放器暂停、切换歌曲，或拖动播放进度。
2. 点击歌曲打开详情，查看歌词和音乐风格。带人声的歌曲在有时间轴数据时可以显示同步歌词。纯音乐不需要歌词时间轴。
3. 打开 **Library** 搜索已保存的作品。使用 **Liked** 或 **Instrumental** 缩小范围，也可以调整模型筛选和排序方式。
4. 打开歌曲的 **More actions** 菜单，选择 **Like**、**Rename** 或 **Download**，分别用于收藏、重命名和下载。生成的音频以 WAV 文件下载。

![查看作品详情并使用播放器](/images/manual/use-cases/music-studio-playback.png#bordered)

在同一菜单中选择 **Reuse seed**，可以将作品的创作设置填回表单。修改歌词或风格后，再生成一个版本。更改其他设置后，即使复用 seed，也不保证得到完全相同的音频。

对于不再需要的歌曲，可以选择 **Delete**。如果还想保留音频，请先下载一份副本。

:::info 局部重绘暂不可用
当前版本的局部重绘暂不可用。你可以修改歌词或风格描述，重新生成一个版本。
:::

## 切换界面语言

打开 **Settings**，找到 **Display language**。选择 **English**、**简体中文**，或选择 **Auto** 跟随 Olares 的语言设置。这个设置只改变界面语言，歌曲语言需要在创作表单中单独选择。

## 常见问题

### 为什么提示没有可用的音乐模型？

启动或恢复应用后，请等待模型服务完成加载。如果创作页面仍提示没有可用模型，可以刷新页面，并检查 Router 是否正在运行。已有模型任务进行时，请等待任务结束后再开始新的创作。

### 为什么没有同步歌词？

歌词高亮依赖该首人声作品返回的时间轴数据。没有时间轴时，仍可阅读歌词、播放或下载音频。纯音乐没有演唱歌词。

## 了解更多

- [Olares Router](olares-router.md)：管理 Olares 上的 AI 模型可用性和路由。
