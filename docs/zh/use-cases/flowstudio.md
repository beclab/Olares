---
outline: [2, 3]
description: 直接在 FlowStudio 中生成图片和视频，或通过 Lares 调用工作流。了解 FlowStudio、Router 和 Lares 如何协作。
head:
  - - meta
    - name: keywords
      content: Olares, FlowStudio, Lares, Olares Router, 图片生成, 视频生成
app_version: "0.3.78"
doc_version: "1.0"
doc_updated: "2026-09-29"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/flowstudio.md)为准。
:::

# FlowStudio

FlowStudio 提供现成的工作流，用于生成图片、编辑已有图片和生成视频。选择一个工作流创建场景后，FlowStudio 会为你准备所需的模型和运行环境。

## 选择创作方式

你可以直接在 FlowStudio 中创作，也可以让 Lares 通过 Router 调用其能力。

| 方式 | 操作方式 | 教程 |
| --- | --- | --- |
| 在 FlowStudio 中创作 | 在 FlowStudio 中选择场景，输入提示词、提供所需素材，并配置生成参数。 | [在 FlowStudio 中生成图片和视频](flowstudio-create.md) |
| 通过 Lares 创作 | 在对话中描述需求，由 Lares 通过 Router 调用可用的 FlowStudio 能力。 | [通过 Lares 使用 FlowStudio](flowstudio-lares.md) |

## 应用如何协作

无论直接在 FlowStudio 中创作，还是通过 Lares 发起请求，生成工作流都由 FlowStudio 执行。Lares 提供对话界面，通过 Router 发送请求。使用任一方式前，都需要在 FlowStudio 中创建所需场景，并等待模型下载和初始化完成。

| 应用 | 职责 |
| --- | --- |
| FlowStudio | 准备工作流场景，执行图片和视频生成任务。 |
| Router | 将 FlowStudio 的能力提供给 Lares，并路由生成请求。 |
| Lares | 使用大语言模型理解需求，调用可用的生成能力。 |

## 开始使用

先参照[在 FlowStudio 中生成图片和视频](flowstudio-create.md)安装应用并准备图片和视频场景。如需通过对话创作，请参阅[通过 Lares 使用 FlowStudio](flowstudio-lares.md)。

## 了解更多

- [Lares](lares.md)：开始使用 Olares AI 助手。
- [Olares Router](olares-router.md)：了解模型和工具路由。
