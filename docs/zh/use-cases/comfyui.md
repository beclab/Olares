---
outline: [2, 3]
description: 在 Olares 上本地运行 ComfyUI：一键部署 ComfyUI 和模型，用默认工作流生成 AI 图像，所有产出都保留在你自己的设备上。
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, Stable Diffusion, AI image generation, self-hosted comfyui, comfyui alternative, comfyui workflow, comfyui on olares
app_version: "1.0.34"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../use-cases/comfyui.md)为准。
:::

# 本地运行 ComfyUI 进行 AI 图像生成

ComfyUI 是一个强大的、基于节点的 Stable Diffusion 界面，将 AI 图像生成转化为可视化编程体验。通过像积木一样连接不同的节点，你可以精确控制生成过程的每个方面，从提示词和模型到后处理效果。

## 学习目标

在本指南中，你将学习如何：
- 安装 ComfyUI Shared 并了解其组件。
- 通过 ComfyUI Launcher 下载基本的 Stable Diffusion 模型包。
- 启动 ComfyUI 服务并使用默认工作流生成你的第一张图像。

## 前提条件
- 已安装 GPU 和足够磁盘空间来下载模型的 Olares 工作实例。
- 从 Market 安装 ComfyUI 并启动 ComfyUI 服务的管理员权限。

## 安装 ComfyUI

1. 打开 **Market** 并搜索 "ComfyUI"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。

   ![安装 ComfyUI](/images/one/comfyui.png#bordered)

安装后，你将在 Launchpad 上看到两个图标：
- **ComfyUI**：你构建工作流和生成图像的客户端界面。
- **ComfyUI Launcher**：管理员的管理仪表板。你必须使用此工具在集群中的任何人可以使用客户端之前启动 ComfyUI 服务。

:::info 成员用户
成员用户只会看到 ComfyUI 客户端图标。管理员必须从 Launcher 启动服务，然后成员才能访问 ComfyUI。
:::

## 下载基本模型包

在生成图像之前，你需要准备模型。本指南使用 Stable Diffusion v1.5 作为示例。ComfyUI Launcher 提供一键包，包含所有基本基础模型。

1. 从 Launchpad 打开 **ComfyUI Launcher**。
2. 向下滚动到 **Resource Package** 部分。
3. 找到 **Stable Diffusion base package** 并点击 **VIEW**。

   ![Stable Diffusion 基础包](/images/manual/use-cases/comfyui-base-package2.png#bordered){width=90%}
   
4. 在包详情页面上，点击 **GET ALL** 开始下载。你可以在状态栏中跟踪进度。

   ![下载进度](/images/manual/use-cases/comfyui-download-progress1.png#bordered){width=90%}

## 启动 ComfyUI 服务

1. 在 ComfyUI Launcher 中，点击右上角的 **START**。

   ![启动 ComfyUI](/images/manual/use-cases/comfyui-start-service.png#bordered)

   :::tip 初始化时间
   初始启动通常需要 10–20 秒进行环境初始化。
   :::

2. 状态变为 "Running" 后，点击 **OPEN** 在新浏览器标签页中启动 ComfyUI 客户端。
3. 当提示此工作流缺少模型时，只需关闭窗口。

## 生成你的第一张图像

ComfyUI 加载时带有默认的文生图工作流。此工作流包含生成图像所需的所有基本节点。

![默认工作流](/images/manual/use-cases/comfyui-default-workflow.png#bordered)

:::tip 了解默认工作流
要理解每个节点及其功能，请参阅 [ComfyUI 文生图工作流节点说明](https://docs.comfy.org/tutorials/basic/text-to-image#comfyui-text-to-image-workflow-node-explanation)。
:::

1. 点击工具栏中的 **Run** 以使用默认提示词生成图像。生成的图像出现在 **Save Image** 节点中。

   ![生成的图像](/images/manual/use-cases/comfyui-generated-image.png#bordered)

2. 尝试修改 **CLIP Text Encode** 节点中的文本，然后再次点击 **Run** 查看输出如何变化。

3. 右键点击 **Save Image** 节点中的图像以将其保存到本地，或在 Files 应用的 `External/olares/ai/output/comfyui` 中找到所有输出文件。

   ![在 Files 中查看生成的图像](/images/manual/use-cases/comfyui-check-generated-image-in-files.png#bordered)

## 了解更多

- [在 Olares 上管理 ComfyUI](comfyui-launcher.md)：控制 ComfyUI 服务、管理模型、插件、Python 依赖和故障排查任务。
- [使用 ComfyUI 和 Krita 创作 AI 艺术](comfyui-for-krita.md)：将 ComfyUI 与 Krita 集成用于 AI 驱动的数字艺术工作流程。
