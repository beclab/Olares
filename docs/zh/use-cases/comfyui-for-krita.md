---
outline: [2, 3]
description: 将 ComfyUI 与 Krita 集成，用于 AI 驱动的数字艺术创作。将你 Olares 上托管的 ComfyUI 连接到 Krita，并无缝生成 AI 艺术作品。
head:
  - - meta
    - name: keywords
      content: Olares, ComfyUI, Krita, AI art, digital painting, Krita AI Diffusion, image generation
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-03-20"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/comfyui-for-krita.md)为准。
:::

# 使用 ComfyUI 和 Krita 创作 AI 艺术

ComfyUI 提供强大的 AI 图像生成能力，但要让它真正有用，你需要将其集成到你的创意工作流程中。本指南展示如何将运行在 Olares 上的 ComfyUI 连接到你电脑上的 Krita，以便你可以直接在数字绘画环境中生成 AI 艺术作品。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 中部署和配置 ComfyUI，以最大化性能和资源效率。
- 安装和配置 Krita AI Diffusion 插件。
- 将 Krita 连接到你 Olares 托管的 ComfyUI 实例。
- 在 Krita 中使用文本提示生成 AI 艺术作品。

## 前提条件

- 已安装并运行 [ComfyUI Shared](comfyui.md) 的 Olares 工作实例。
- 你的电脑上已安装 [Krita](https://krita.org/en/download/)。
- Olares 设备上有足够的系统资源来下载模型。

## 设置 ComfyUI

打开 ComfyUI Launcher 并点击 **START** 以确保服务正在运行。

:::tip 最大化 GPU 性能
你可以在 **Settings** > **GPU** 中将 GPU 模式设置为 **App exclusive** 并分配 ComfyUI 完整的 GPU 访问权限，以确保最大性能。
:::

## 获取 ComfyUI 的端点

1. 在 Olares 上，打开 Settings，然后前往 **Application** > **ComfyUI Shared**。
2. 在 **Entrances** 下，点击 **ComfyUI**。
3. 确保其 **Authentication level** 设置为 **Internal**。
4. 在 **Endpoint** 旁边，点击 <i class="material-symbols-outlined">content_copy</i> 复制显示的端点 URL。

   ![设置端点](/images/manual/use-cases/comfyui-set-up-endpoint1.png#bordered){width=90%}

## 下载并启用 AI Diffusion 插件
1. 下载 [Krita AI Diffusion 插件](https://github.com/Acly/krita-ai-diffusion/releases)。
2. 启动 Krita，从工具栏中选择 **Tools** > **Scripts** > **Import Python Plugin from File**。

   ![导入 AI 插件](/images/manual/use-cases/krita-import-plugin1.png#bordered)

3. 选择下载的 ZIP 包。
4. 出现提示时，确认插件激活并重启 Krita。

   ![确认插件激活](/images/manual/use-cases/krita-comfirm-plugin-activation.png#bordered){width=40%}

5. 重启后，在 **Krita** > **Preferences** > **Python Plugin Manager** 中验证安装。

   ![验证 AI 插件](/images/manual/use-cases/krita-verify-plugin.png#bordered)

## 将 Krita 连接到 ComfyUI

连接步骤取决于你的电脑和 Olares 设备是否在同一网络上。

<tabs>
<template #使用-.local-域名（局域网，推荐）>

如果你的电脑与 Olares 在同一本地网络上，你可以使用 `.local` 域名连接而无需 LarePass VPN。以下步骤以 macOS 为例，其中 `.local` 域名无需额外设置即可原生工作。

:::info Windows 用户
在 Windows 上，多级 `.local` 域名需要一些额外设置。尝试以下方法之一：
- **在 LarePass 中导入 hosts**：打开 LarePass 桌面应用，使用内置选项将 Olares hosts 导入到你的系统。
- **使用单级域名**：将 `https://806ba3e40.{username}.olares.com` 更改为 `http://806ba3e40-{username}-olares.local`。

有关详细信息，请参阅[本地访问 Olares 服务](../manual/best-practices/local-access.md)。
:::

1. 在 Krita 中创建一个新文档。

   :::tip 画布尺寸
   从 512 x 512 像素的画布开始，以优化性能并高效管理图形内存。
   :::

2. 点击 **Settings** > **Dockers** > **AI Image Generation** 以启用插件。你可以将面板放置在方便的位置。

   ![启用 AI 插件](/images/manual/use-cases/krita-enable-plugin.png#bordered)

3. 点击 **Configure** 访问插件设置。

   ![配置 AI 插件](/images/manual/use-cases/krita-configure-plugin1.png#bordered){width=70%}

4. 设置 ComfyUI 连接：

   a. 在 **Connection** 中，选择 **Custom Server**，并粘贴你的 ComfyUI URL。

   b. 将 URL 更改为使用 `.local` 域名和 `http`。例如，如果复制的 URL 是：
      ```plain
      https://806ba3e40.laresprime.olares.com
      ```
      将其更改为：
      ```plain
      http://806ba3e40.laresprime.olares.local
      ```

   c. 点击 **Connect** 验证连接。

   ![连接 ComfyUI](/images/manual/use-cases/krita-missing-required-nodes-local.png#bordered)

   你可能会看到错误消息，指示连接已建立但服务器缺少必需的自定义节点或模型。这是预期的。继续前往[准备模型和插件](#准备模型和插件)下载所需的资源。

</template>

<template #使用-.com-域名>

如果你的电脑与 Olares 不在同一本地网络上，请启用 LarePass VPN 以确保安全连接。

1. 在 LarePass 桌面客户端上启用 VPN：

   a. 打开 LarePass 应用，点击左上角的头像打开用户菜单。

   b. 打开 **VPN connection** 的开关。

   启用后，确保连接状态为 **Intranet**（局域网）或 **P2P**（局域网外）。

2. 在 Krita 中创建一个新文档。

   :::tip 画布尺寸
   从 512 x 512 像素的画布开始，以优化性能并高效管理图形内存。
   :::

3. 点击 **Settings** > **Dockers** > **AI Image Generation** 以启用插件。你可以将面板放置在方便的位置。

   ![启用 AI 插件](/images/manual/use-cases/krita-enable-plugin.png#bordered)

4. 点击 **Configure** 访问插件设置。

   ![配置 AI 插件](/images/manual/use-cases/krita-configure-plugin1.png#bordered){width=70%}

5. 设置 ComfyUI 连接：

   a. 在 **Connection** 中，选择 **Custom Server**，并粘贴你的 ComfyUI URL。

   b. 点击 **Connect** 验证连接。

   ![连接 ComfyUI](/images/manual/use-cases/krita-missing-required-nodes-com.png#bordered)

   你可能会看到错误消息，指示连接已建立但服务器缺少必需的自定义节点或模型。这是预期的。继续前往[准备模型和插件](#准备模型和插件)下载所需的资源。

</template>
</tabs>

## 准备模型和插件

### 安装必需的自定义节点

Krita AI Diffusion 插件需要以下自定义节点：
- ControlNet 预处理器: https://github.com/Fannovel16/comfyui_controlnet_aux
- IP-Adapter: https://github.com/cubiq/ComfyUI_IPAdapter_plus
- Inpaint 节点: https://github.com/Acly/comfyui-inpaint-nodes
- 外部工具节点: https://github.com/Acly/comfyui-tooling-nodes

要安装它们：

1. 打开 ComfyUI Launcher，然后前往 **Plugins** > **Custom Install**。
2. 粘贴自定义节点的 GitHub URL（例如 `https://github.com/Acly/comfyui-tooling-nodes`），然后点击 **INSTALL PLUGIN**。
3. 对上面列出的其余每个自定义节点重复操作。
   ![下载自定义节点](/images/manual/use-cases/comfyui-download-custom-nodes.png#bordered)
4. 返回 ComfyUI Launcher 中的 **Home** 页面，然后点击 **RESTART** 使更改生效。

5. 可选：如果你返回 Krita 并再次点击 **Connect**，你应该会看到错误消息，指示必需的模型仍然缺失。

   ![缺少必需的模型](/images/manual/use-cases/krita-missing-required-models.png#bordered)

### 安装必需的模型

该插件需要这些实用模型才能正常工作。没有它们，某些功能将无法正常运行。预先安装它们可以确保更流畅的体验。

- NMKD Superscale: https://huggingface.co/gemasai/4x_NMKD-Superscale-SP_178000_G/resolve/main/4x_NMKD-Superscale-SP_178000_G.pth
- OmniSR X2: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X2_DIV2K.safetensors
- OmniSR X3: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X3_DIV2K.safetensors
- OmniSR X4: https://huggingface.co/Acly/Omni-SR/resolve/main/OmniSR_X4_DIV2K.safetensors
- MAT Inpaint: https://huggingface.co/Acly/MAT/resolve/main/MAT_Places512_G_fp16.safetensors

通过 ComfyUI Launcher 下载它们：

1. 打开 ComfyUI Launcher，然后前往 **Models** > **Custom Download**。
2. 下载放大模型：

   a. 粘贴 NMKD Superscale URL，将 **Destination folder** 设置为 **Upscale Models**，然后点击 **DOWNLOAD MODEL**。
      ![下载必需的放大模型](/images/manual/use-cases/comfyui-download-upscale-models.png#bordered)

   b. 使用相同的目标文件夹对三个 OmniSR 模型重复操作。

3. 下载 inpaint 模型：

   a. 粘贴 MAT Inpaint URL。

   b. 将 **Destination folder** 设置为 **Custom Directory**。

   c. 输入 `inpaint` 作为 **Directory Name**。

   d. 点击 **DOWNLOAD MODEL**。
      ![下载必需的 inpaint 模型](/images/manual/use-cases/comfyui-download-inpaint-model.png#bordered)
4. 返回 ComfyUI Launcher 中的 **Home** 页面，然后点击 **RESTART** 使更改生效。
5. 可选：如果你返回 Krita 并再次点击 **Connect**，你应该会看到错误消息，指示基础模型仍然缺失。

   ![缺少必需的模型](/images/manual/use-cases/krita-missing-base-models.png#bordered)

### 安装基础扩散模型

至少需要一个扩散模型（通常称为 "checkpoint"）。本指南使用 Z-Image Turbo 作为示例。Z-Image Turbo 是一个中等大小的模型，在质量和速度之间取得平衡，生成逼真的图像而不需要过多的内存。

1. 打开 ComfyUI Launcher，向下滚动到 **Package installation** 部分。
2. 找到 **Z-Image Turbo Package** 并点击 **VIEW**。

   ![Z-Image Turbo 包](/images/manual/use-cases/comfyui-zimage-turbo-package.png#bordered)

3. 在包详情页面上，点击 **GET ALL** 开始下载。你可以在状态栏中跟踪进度。

   ![下载进度](/images/manual/use-cases/comfyui-download-progress-z-image.png#bordered)

4. 返回 ComfyUI Launcher 中的 **Home** 页面，然后点击 **RESTART** 使更改生效。
5. 在 Krita 中，前往 **Connection** > **Server Configuration** 并再次点击 **Connect**。绿色的 "Connected" 指示器确认连接成功。你应该会在基础模型列表中看到 Z-Image 标记为 "supported"。

   ![检测到 Z-Image](/images/manual/use-cases/comfyui-z-image-detected.png#bordered)

## 添加风格

在生成图像之前，你需要创建一个 Style Preset，告诉 Krita 使用哪个模型。

1. 在 Krita 中打开 **Configure Image Diffusion** 对话框，然后前往 **Styles** 选项卡。
2. 对于 **Style Presets**，从内置风格中选择 **Z-Image Turbo**。

   ![选择内置 Z-Image Turbo 风格](/images/manual/use-cases/krita-select-built-in-style.png#bordered)

3. 点击复制图标以创建当前风格的副本。

   ![复制风格](/images/manual/use-cases/krita-duplicate-style.png#bordered)

4. 对于 **Model Checkpoint**，选择 Z-Image 模型。模型名称应该是 `public/z_image_turbo_bf16`。

   ![选择 Z-Image 模型](/images/manual/use-cases/krita-select-z-image-model.png#bordered)

5. 点击刷新图标以刷新可用风格列表。
   ![刷新风格列表](/images/manual/use-cases/krita-refresh-style-list.png#bordered)

6. 保持其他设置的默认值，然后点击 **Ok** 保存更改。
   :::warning
   如果你不熟悉 Krita，建议使用默认设置。更改默认设置可能会产生意想不到的结果。
   :::

## 使用文本提示创作 AI 艺术

1. 在 **AI Image Generation** 面板中，确认已选择 Z-Image Turbo 风格。

2. 在文本框中输入你的提示。例如：

   ```plain
   A person relaxing on a sandy beach, basking in the warm sunlight, with the calm blue ocean in the background.
   ```

3. 点击 **Generate**。生成的图像出现在画布上。
   ![生成图像](/images/manual/use-cases/krita-generated-image-1.png#bordered)

4. 再次点击 **Generate** 生成新图像。
   ![生成图像](/images/manual/use-cases/krita-generated-image-2.png#bordered)

5. 选择一个首选结果，然后点击 **Apply** 将其添加到图层中。

## Inpaint

要完善生成图像的特定区域，请使用 inpainting。这允许你在保持其余部分完整的同时修改图像的某些部分。

1. 选择自由手选择工具并在要修改的区域周围绘制。
   ![使用自由手选择工具](/images/manual/use-cases/krita-use-selection-tool.png#bordered)

2. 输入你想要在选定区域中看到的内容的描述。例如：
   ```plain
   Seagulls can be seen flying in the distant sky.
   ```

3. 点击 **Fill**。几个填充候选将出现在面板中。
   ![填充候选](/images/manual/use-cases/krita-fill.png#bordered)
4. 点击每个候选以在画布上预览它。

5. 当你找到喜欢的结果时，点击 **Apply** 将其添加到图层中。
   ![选择 inpaint 候选](/images/manual/use-cases/krita-select-inpaint-candidate.png#bordered)

## 故障排查

### 无法从 Krita 连接到 ComfyUI

如果 Krita 显示连接错误：

| 检查 | 解决方法 |
|:------|:-----------|
| 网络连接 | 确保你的电脑和 Olares 在同一网络上。 |
| ComfyUI 认证级别 | 在 **Settings** > **Application** > **ComfyUI Shared** 中，确认设置为 **Internal**。 |
| `.com` URL 的 LarePass VPN | 在 LarePass 桌面应用中启用 **VPN connection**。 |
| 干扰的代理/VPN | 暂时禁用其他 VPN 或代理软件。 |
| ComfyUI 服务状态 | 打开 ComfyUI Launcher 并验证服务正在 **Running**。 |
| GPU 访问 | 在 **Settings** > **GPU** 中，验证 ComfyUI 已绑定到具有足够<br> VRAM 分配的 GPU。 |
| 必需的插件和模型 | 确保所有自定义节点、实用模型和基础扩散模型<br> 已下载并且 ComfyUI 已重启。 |

## 了解更多

- [ComfyUI 快速入门指南](comfyui.md)：安装 ComfyUI 并生成你的第一张图像。
- [使用 ComfyUI Launcher 管理 ComfyUI](comfyui-launcher.md)：控制 ComfyUI 服务、管理模型并配置环境。
- [Krita AI Diffusion 文档](https://github.com/Acly/krita-ai-diffusion/wiki)：探索高级功能和工作流程。
