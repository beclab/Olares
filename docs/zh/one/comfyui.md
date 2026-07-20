---
noindex: true
outline: [2,3]
description: 了解如何安装 ComfyUI，通过 ComfyUI Launcher 管理模型，以及在 Olares One 上生成高性能图像和视频。
head:
  - - meta
    - name: keywords
      content: 本地 AI, comfyui, nunchaku, wan
---

:::warning
本页面由 AI 自动翻译，部分技术术语可能与中文习惯存在差异。如有疑问，请以[英文原文](../../one/comfyui.md)为准。
:::

# 使用 ComfyUI 生成图像和视频 <Badge type="tip" text="1 h" />
ComfyUI 是一个强大的基于节点的 Stable Diffusion 界面，将 AI 媒体生成转变为可视化编程体验。

Olares 提供 ComfyUI Shared，允许多个用户在集群内共享模型、插件和工作流资源。它还配备了 ComfyUI Launcher，为管理员用户提供一种简单的方式来管理 ComfyUI 资源和运行时环境。

## 学习目标
- 安装和配置 ComfyUI 服务。
- 使用 ComfyUI Launcher 下载优化的模型包。
- 使用 Z-Image Turbo 工作流生成图像。
- 使用 Wan 2.2 模型生成视频。

<!--
## 开始之前
对于图像生成：
- Olares One 配备了 NVIDIA RTX 5090 mobile GPU。这允许你利用 Nunchaku Flux.1-dev 模型以比标准 FP16 或 FP8 版本更快的速度生成图像。
- Nunchaku Flux.1-dev 是一种使用 SVDQuant 量化（NVFP4）的优化模型。它旨在该特定硬件上提供高性能推理，同时保持最小的视觉质量损失。
-->

## 前提条件
**硬件** <br>
- 连接到稳定网络的 Olares One。
- 足够的磁盘空间来下载模型。

**用户权限**
- 从 Market 安装 ComfyUI 以及为集群启动或停止 ComfyUI 服务的管理员权限。

## 步骤 1：安装 ComfyUI
1. 打开 Market，搜索 "ComfyUI"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。

   ![安装 ComfyUI](/images/one/comfyui.png#bordered)

安装完成后，你可以在 Launchpad 上看到两个 ComfyUI 图标：
- **ComfyUI**：ComfyUI 的客户端界面。
- **ComfyUI Launcher**：核心管理仪表板。你必须使用此工具启动 ComfyUI 服务，然后才能使用客户端。

## 步骤 2：下载模型包
要生成内容，你必须首先下载特定的模型包。

1. 打开 ComfyUI Launcher，向下滚动到 **Resource Package**。
2. 对于图像生成，找到 **Z-Image Turbo Package** 并点击 **GET ALL** 下载所需模型。
3. 对于视频生成，选择 **Wan 2.2 Text to Video 14B Package** 并点击 **GET ALL** 下载所需模型。

   ![安装模型包](/images/one/comfyui-install-model-package2.png#bordered){width=90%}
   <!-- ![安装模型包](/images/one/comfyui-install-model-package.png#bordered) -->

4. 下载完成后，在 **Home** 页面点击 **RESTART** 以使更改生效。

## 步骤 3：启动 ComfyUI 服务
1. 在 ComfyUI Launcher 中，点击右上角的 **START**。
   ![启动 ComfyUI](/images/one/comfyui-start.png#bordered)
   :::tip 初始化时间
   首次启动通常需要 10-20 秒，因为环境正在初始化。
   :::
2. 状态变为 "Running" 后，点击 **OPEN**。这将在新的浏览器标签页中启动 ComfyUI 客户端。

## 步骤 4：生成图像
本节使用 Z-Image Turbo 工作流来帮助你入门。

1. 在 ComfyUI 客户端中，点击左侧导航栏的 **Templates**，然后在 **GENERATION TYPE** 下选择 **Image**。
2. 搜索 "Z-Image"，然后从结果中选择 **Z-Image-Turbo Text to Image** 以打开工作流。
   ![Z-Image Turbo 模板](/images/one/comfyui-tti-template.png#bordered)

3. 在同一节点中更新文本提示词，描述你想要生成的图像。
4. 点击工具栏中的 **Run** 开始生成。
   ![生成的图像](/images/manual/use-cases/comfyui-z-image-result.png#bordered)

<!--
## 步骤 4：生成图像
本节使用 `nunchaku-flux.1-dev-qencoder` 工作流来帮助你入门。

1. 在 ComfyUI 客户端中，点击左上角的 **ComfyUI** 图标打开菜单。
2. 选择 **View** > **Browse Templates**。
3. 在 **EXTENSIONS** 下，选择 **ComfyUI-nunchaku**。
4. 选择模板：**nunchaku-flux.1-dev-qencoder**。
   ![打开 Nunchaku 工作流](/images/one/comfyui-nunchaku-templates.png#bordered)

5. 在每个模型加载器节点的文件名前添加 `public/`。例如：
   - **默认**：`clip_I.safetensors`
   - **更改为**：`public/clip_I.safetensors`
   :::info 共享模型路径
   Olares 中的 ComfyUI 使用与标准安装不同的文件结构。此更改允许模型在 ComfyUI 和 SD Web UI 之间共享。
   :::
   ![更改模型路径](/images/one/comfyui-nunchaku-change-model-path.png#bordered)

6. 替换 **CLIP Text Encode** 中的文本以更新图像的提示词。例如：
    ```plain
   8-bit cyberpunk: Blocky pixel cat holds "olares is fast!" neon on pixel street.
    ```
7. 点击工具栏中的 **Run** 开始生成。
-->

## 步骤 5：生成视频
本节使用 Wan 2.2 工作流。

1. 在 ComfyUI 客户端中，点击左侧导航栏的 **Templates**，然后在 **GENERATION TYPE** 下选择 **Video**。
2. 搜索 "Wan 2.2"，然后从结果中选择 **Wan 2.2 14B Text to Video** 以打开工作流。
   ![打开 Wan 2.2 工作流](/images/one/comfyui-video-templates.png#bordered)

3. 在工作流中，找到 **CLIP Text Encode (Positive Prompt)** 节点并根据需要编辑提示词。你可以修改完整提示词或只修改你想更改的部分。如果需要，你也可以调整 **CLIP Text Encode (Negative Prompt)** 节点。

   例如：

    ```plain
    A woman with long brown hair and light skin smiles at another woman with long blonde hair. The woman with brown hair wears a black jacket and has a small, barely noticeable mole on her right cheek. The camera angle is a close-up, focused on the woman with brown hair's face. The lighting is warm and natural, likely from the setting sun, casting a soft glow on the scene. The scene appears to be real-life footage.
    ```

4. 点击工具栏中的 **Run** 开始生成。视频生成比图像生成耗时显著更长。
   ![生成的视频](/images/manual/use-cases/comfyui-w-video-result.png#bordered)

## 步骤 6：下载输出文件
你可以将所有输出图像和视频从 Olares One 下载到你的本地电脑。
1. 打开 Files 应用。
2. 导航到以下目录：
    ```plain
    External/olares/ai/output/comfyui
    ```
   ![查看输出文件](/images/one/comfyui-output.png#bordered)
   
3. 选择你要保存的文件。
4. 右键点击并选择 **Download** 将它们保存到你的本地电脑。

## 故障排除

### CPU 温度异常升高

运行大型工作流可能导致 Olares One 的 CPU 温度异常升高。请参阅 [Olares One 上 CPU 温度异常升高](/use-cases/comfyui-common-issues#cpu-temperature-rises-unusually-high-on-olares-one)了解解决方法。

## 资源
- [ComfyUI 官方文档](https://docs.comfy.org/)
- [管理模型](/use-cases/comfyui-launcher.md#manage-models)
- [使用 ComfyUI Launcher 管理 ComfyUI](../use-cases/comfyui-launcher.md)
