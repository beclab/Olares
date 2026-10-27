---
outline: [2, 3]
description: 使用 FlowStudio 中准备好的工作流，通过 Lares 对话生成图片和视频。
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Lares, Olares Router, 对话生成图片, 视频生成
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/flowstudio-lares.md)为准。
:::

# 通过 Lares 使用 FlowStudio

你可以在 Lares 对话中请求生成图片和视频。Router 将请求连接到可用的 FlowStudio 能力，再由 FlowStudio 执行生成工作流。

## 前提条件

- Olares OS 1.12.7 或更高版本。
- 已从 Market 安装 FlowStudio、Router 和 Lares。已安装的应用需更新至最新可用版本。
- 已安装 Qwen3.8-27B (llama.cpp)，并完成模型下载。
- 已创建所需的 FlowStudio 场景，并完成模型下载和初始化。具体操作请参阅[在 FlowStudio 中生成图片和视频](flowstudio-create.md)。

:::warning 重要：一次只运行一个任务
在 Olares One 上运行 Qwen3.8-27B (llama.cpp) 时，建议一次只处理一个请求，以在 102K 上下文窗口和当前模型精度下获得最佳体验。
:::

:::warning 避免同时运行 FlowStudio 生成任务和 Qwen3.8-27B 对话
在 Lares 中使用 Qwen3.8-27B (llama.cpp) 时，请等待直接在 FlowStudio 中启动的图片或视频生成任务完成，再发送消息。同时运行两者可能导致显存冲突，使模型无法响应。
:::

## 开始使用

1. 打开 Lares，选择 **Qwen3.8-27B (llama.cpp)**，并确认权限级别。

   ![Lares 对话界面](/images/manual/use-cases/lares1.png#bordered)

2. 询问 Lares 当前有哪些可用的图片和视频生成能力：

   ```text
   What image and video generation capabilities are currently available through FlowStudio?
   ```

3. 查看可用能力，然后描述你想创作的内容。如果需要使用特定的生成模型，请在提示词中写明模型名称；否则，让 Lares 根据需求和可用模型进行选择。

## 生成图片

1. 描述你想生成的图片。以下示例指定了生成模型：

   ```text
   Use Qwen-Image 2.1 NVFP4 · Text to Image to generate a 16:9 automotive poster. Show a red vintage coupe on a coastal road, with a seaside village and turquoise sea in the background. Use warm late-afternoon lighting and add the headline "TIMELESS DRIVE".
   ```

2. 等待生成完成，然后打开图片查看效果。

## 生成视频

1. 描述你想要的场景、动作和声音。以下示例由 Lares 选择可用的生成模型：

   ```text
   Generate a 5-second video at 480p. A red vintage coupe drives slowly along a coastal road at sunset. The camera tracks alongside it, keeping the whole car visible as its wheels turn. Add a gentle engine sound and soft instrumental music. No dialogue or text.
   ```

2. 等待生成完成，然后播放视频查看效果。

## 了解更多

- [FlowStudio](flowstudio.md#应用如何协作)：了解 FlowStudio、Router 和 Lares 如何协作。
- [Lares](lares.md)：了解如何使用 Lares 管理 Olares 并开展研究。
- [在 FlowStudio 中生成图片和视频](flowstudio-create.md)：直接使用场景并准备示例素材。
