---
outline: [2, 3]
description: 使用 Olares 上的 FastH3，根据文本、首尾帧和参考素材生成带同步音频的视频，并管理生成结果与模型权重。
head:
  - - meta
    - name: keywords
      content: Olares, FastH3, MiniMax H3, AI 视频生成, 文生视频, 图生视频, 参考素材生成视频, 本地部署
app_version: "0.1.15"
doc_version: "1.0"
doc_updated: "2026-11-26"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/fasth3.md)为准。
:::

<script setup>
import { withBase } from 'vitepress'
</script>

# 使用 FastH3 生成视频

FastH3 是 Olares 上的一款本地视频生成应用，使用 MiniMax H3 的 8 步蒸馏模型同步生成视频和音频。你可以根据文本描述生成视频，让静态图片动起来，或使用参考图片、视频和音频引导视频生成。

通过可视化界面，你无需自行搭建 ComfyUI 工作流，就能创作和调整视频。AI 提示词优化功能通过 Router 调用对话模型，帮助你完善创作描述。

## 学习目标

- 安装 FastH3 并准备视频生成模型。
- 根据文本、首尾帧和参考素材生成视频。
- 提取并复用视频帧，创作新的视频。
- 查找和管理媒体素材。

## 前提条件

- Olares 1.12.8 或更高版本，以及用于安装应用的管理员账号。
- 应用支持的 NVIDIA GPU。应用安装包要求 22 GiB 显存和 32 GiB 系统内存。
- 足够的磁盘空间，用于存储应用、媒体素材和模型权重。默认 NVFP4 模型权重需要下载约 47 GB 的文件。

## 安装 FastH3

1. 打开 Market，搜索“FastH3”。
2. 点击 **Get**，再点击 **Install**，等待安装完成。

## 准备模型

从 Launchpad 打开 FastH3。首次启动时，应用默认自动下载所需的 NVFP4 模型权重。

如需查看下载进度，在 FastH3 中打开 **Settings** > **Models**，点击 **Open the Model Console**。等待控制台显示 **Ready**，且生成服务准备就绪后，再提交任务。

FastH3 与 Olares 上的其他 ComfyUI 应用及 FlowStudio 共享模型文件，可以复用已有的模型文件。

FastH3 是一款[共享应用](../manual/olares/market/shared-apps.md)。管理员只需安装一次，同一 Olares 上的所有用户就能使用。每位用户的个人生成队列、上传素材和生成结果仅自己可见。

## 生成视频

FastH3 支持三种视频生成模式：

| 模式 | 输入 | 使用场景 |
| --- | --- | --- |
| Text | 文本提示词 | 无需已有素材，直接根据创意生成视频。 |
| Frames | 起始帧、可选的结束帧及提示词 | 让图片动起来，或引导两个画面之间的过渡。 |
| References | 参考图片、视频、音频片段及提示词 | 使用已有媒体素材引导视频生成。 |

下文以制作香水广告为例，先根据文本生成视频，再复用视频帧创作不同的镜头，最后结合产品图片与新的参考素材生成视频。你可以按需替换产品、调整创作内容。

三个案例的 **Seed** 均使用 **Random**，未使用 **AI optimize**。你生成的视频会与案例有所不同。

### 根据文本生成视频

使用 **Text** 模式，无需上传参考素材，就能根据描述生成视频。

1. 前往 **Create**，选择 **Text**。
2. 在 **Prompt** 中描述视频主体、场景、动作和视觉风格。例如：

   ```text
   A cinematic product shot of a minimalist transparent glass perfume bottle filled with amber liquid, standing on a cream-colored stone pedestal.

   The front of the bottle features only one word: "PERFUME", printed in simple, clean, black uppercase sans-serif letters. The text is centered, clearly legible, and remains unchanged throughout the video. No brand names, logos, additional text, numbers, or decorative typography.

   Warm sunlight casts soft shadows across the background. The camera slowly pushes toward the bottle as delicate reflections shift across the glass. Minimalist composition, realistic glass materials, refined luxury aesthetic, and soft studio lighting.
   ```

3. 可选：使用 **AI optimize** 优化描述。该功能支持三种处理方式：整理提示词结构并翻译成英文、整理结构但保留原语言，以及仅翻译成英文。

   此功能需要 [Olares Router](olares-router.md) 中有可用的对话模型。你可以在 FastH3 的 **Settings** > **AI optimize** 中选择模型。如需准备本地对话模型，请参阅[使用 Engine Base 运行本地大模型](llm-base-apps.md)。

4. 配置视频生成参数：

   | 参数 | 本教程设置 | 如何选择 |
   | --- | --- | --- |
   | Duration | 5 s | 设置视频时长，支持 2 至 15 秒。 |
   | Aspect | 16:9 | 根据使用场景选择宽高比，或使用 Auto。 |
   | Quality | 480p | 如需更高分辨率的视频，选择 720p。 |
   | Video count | ×1 | 每次提交可生成 1 至 4 个视频。 |
   | Seed | Random | 在 Advanced settings 中启用 Lock seed，可在多次生成时使用相同的随机种子。 |

5. 点击 **Generate** 提交任务。如果有可用的耗时估算，按钮上会显示预计生成时间。
6. 在 **Results** 中查看生成进度、已用时间和预计剩余时间。
7. 任务完成后，打开结果预览视频。

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="根据文本提示词生成的香水视频">
  <source :src="withBase('/images/manual/use-cases/fasth3-text.mp4')" type="video/mp4" />
  你的浏览器不支持嵌入式视频。<a :href="withBase('/images/manual/use-cases/fasth3-text.mp4')">下载示例视频</a>。
</video>

### 使用 Frames 让图片动起来

使用 **Frames** 模式将已有图片转化为动态视频。本例复用上一节视频中的一帧，创作具有不同镜头运动的视频。

1. 打开上一节生成的视频，定位到能清晰展示香水瓶的画面。
2. 点击 **Grab frame**，选择 **Use as first frame**，将提取的图片用作起始帧。
3. 在 **Prompt** 中描述新的镜头运动：

   ```text
   The camera slowly moves from left to right around the perfume bottle. Warm sunlight creates subtle shifting reflections across the glass and amber liquid. Keep the bottle's original shape, proportions, color, and label consistent throughout the video. Elegant product cinematography with smooth, controlled camera movement.
   ```

4. 将 **Duration** 设置为 5 s、**Quality** 设置为 480p、**Video count** 设置为 ×1。

   在 **Frames** 模式下，输出视频的宽高比由起始帧决定。

5. 点击 **Generate**，预览生成结果。

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="使用提取的视频帧生成的香水视频">
  <source :src="withBase('/images/manual/use-cases/fasth3-frames.mp4')" type="video/mp4" />
  你的浏览器不支持嵌入式视频。<a :href="withBase('/images/manual/use-cases/fasth3-frames.mp4')">下载示例视频</a>。
</video>

你也可以上传自己的图片作为起始帧。每个帧位置支持一张不超过 512 MB 的图片，格式支持 JPG、PNG、WebP、静态 GIF 和 HEIC/HEIF。

如需引导视频的结束画面，可以添加可选的 **End frame**，并在提示词中描述起始帧与结束帧之间的过渡。

### 根据参考素材生成视频

使用 **References** 模式，通过已有媒体素材引导视频内容，而非指定视频的起始帧和结束帧。

本例将上一节视频中的香水瓶与白色茉莉花参考图片相结合，创作新的广告场景。

References 模式支持以下媒体素材：

| 素材类型 | 数量上限 | 时长 |
| --- | --- | --- |
| 图片 | 5 张 | — |
| 视频 | 1 段 | 2–15 秒 |
| 音频片段 | 1 段 | 2–15 秒 |

最多可添加 7 个参考文件，单个文件不超过 512 MB。如需查看支持的文件格式，点击 **Add media** 下的 **File formats**。

1. 打开上一节生成的视频，定位到能清晰展示香水瓶的画面。
2. 点击 **Grab frame**，选择 **Use as reference image**。
3. 下载下方链接中的茉莉花图片。在 **Add media** 中点击 **+**，上传下载的图片。

   <a :href="withBase('/images/manual/use-cases/fasth3-jasmine.jpg')" download="fasth3-jasmine.jpg">下载茉莉花参考图片（JPG）</a>

4. 在 **Prompt** 中输入 `@`，选择参考素材，并描述各素材在视频中的用途。例如：

   ```text
   Create an elegant cinematic perfume commercial featuring the perfume bottle from @image_1 and the white jasmine flowers from @image_2.

   Preserve the original design, shape, and amber color of the perfume bottle. Arrange the jasmine flowers naturally around its base on a cream-colored stone surface. A few delicate white petals drift slowly through the air.

   Warm golden sunlight, soft shadows, realistic glass reflections, and a slow cinematic camera pullback. Refined luxury fragrance advertising aesthetic.
   ```

5. 输入 `@` 并选择对应素材，将示例中的 `@image_1` 和 `@image_2` 替换为实际的素材引用。
6. 将 **Duration** 设置为 5 s、**Aspect** 设置为 16:9、**Quality** 设置为 480p、**Video count** 设置为 ×1。
7. 点击 **Generate**，预览参考素材在生成视频中的呈现效果。

<video controls playsinline preload="metadata" style="display: block; width: 100%; aspect-ratio: 16 / 9; object-fit: contain; background: #000;" aria-label="结合香水瓶和茉莉花参考图片生成的视频">
  <source :src="withBase('/images/manual/use-cases/fasth3-references.mp4')" type="video/mp4" />
  你的浏览器不支持嵌入式视频。<a :href="withBase('/images/manual/use-cases/fasth3-references.mp4')">下载示例视频</a>。
</video>

你也可以添加视频或音频片段作为参考素材，引导新视频的画面内容、动作或声音。在提示词中说明各个参考素材的用途。

## 管理视频和媒体素材

### 复用和延续视频创作

打开已完成的视频，可以执行以下操作：

- **Reuse**：恢复之前任务的输入素材和生成参数，按需调整后重新生成。
- **Continue**：使用已生成视频的最后一帧作为新视频的起始帧。
- **Grab frame**：提取选中的视频帧，并选择用途：
  - **Use as first frame**：用作起始帧。
  - **Use as last frame**：用作结束帧。
  - **Use as reference image**：用作参考图片。
  - **Library only**：保存到 Library，供后续使用。
- **Download**：下载生成的视频。

### 浏览和下载媒体素材

**Results** 面板显示生成任务及其结果。你可以通过搜索或 **Filters**，根据生成模式、分辨率或任务状态查找任务。

前往 **Library** 浏览已上传的媒体素材和生成内容。你可以根据媒体类型和素材来源筛选。

生成队列、上传素材和生成结果均存储在你的 Olares 上，可在不同设备间访问。重启应用后，已上传的素材和生成的视频仍会保留。

### 使用个人空间和项目空间

通过工作空间选择器，在个人空间和项目空间之间切换。

- **My space**：个人工作空间，包含仅自己可见的生成队列、上传素材和生成结果。
- **New project**：创建项目，将项目内容共享给指定用户或 Olares 上的所有用户。

FastH3 将文件存储在 `Common/Studio` 下，每位用户拥有独立的文件夹。

## 管理模型权重

FastH3 提供两种模型权重格式：

| 精度 | 下载大小 | 说明 |
| --- | --- | --- |
| NVFP4 | 约 47 GB | 默认模型集，针对更快的生成速度进行了优化。 |
| INT8 | 约 63 GB | 另一种可选的模型权重格式。 |

在 FastH3 中，前往 **Settings** > **Models**，查看各模型集的安装状态。

你可以下载其他模型权重，在可用精度之间切换，也可以删除不再使用的模型集，释放磁盘空间。

删除模型集前，先检查其他 ComfyUI 应用或 FlowStudio 是否需要这些权重。删除模型集会移除共享模型文件夹中的权重，即使其他应用正在使用这些文件，也会受到影响。

## 在 ComfyUI 中打开工作流

如需进一步自定义视频生成流程，在 FastH3 中前往 **Settings** > **Open in ComfyUI**，选择 **Text**、**Frames** 或 **References**，在节点编辑器中打开对应工作流。

更多信息，请参阅[本地运行 ComfyUI 进行 AI 图像生成](comfyui.md)。

## 常见问题

### 为什么打开应用后无法生成视频？

所需模型可能仍在下载。在 FastH3 中前往 **Settings** > **Models**，点击 **Open the Model Console** 查看下载状态。等待控制台显示 **Ready** 后，再开始生成视频。

### 为什么无法使用 AI optimize？

确保 [Router](olares-router.md) 中有可用的对话模型。在 FastH3 中前往 **Settings** > **AI optimize**，点击 **Refresh**，选择可用的对话模型。

未启用 AI 优化时，你仍可以直接输入提示词生成视频。
