---
outline: [2, 3]
description: 在 Olares 上使用 Excalidraw 创建手绘风格的图表、线框图和草图，打造自托管的虚拟白板。
head:
  - - meta
    - name: keywords
      content: Olares, Excalidraw, whiteboard, diagrams, wireframes, hand-drawn, sketching, self-hosted
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/excalidraw.md)。
:::

# 使用 Excalidraw 创建手绘风格图表

Excalidraw 是一个开源虚拟白板，具有手绘风格的美感。你可以直接在 Olares 设备上用它创建图表、线框图、流程图或任何自由形式的草图。

## 安装 Excalidraw

1. 打开 Market，搜索 "Excalidraw"。
   ![安装 Excalidraw](/images/manual/use-cases/excalidraw.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 使用 Excalidraw

从 Launchpad 打开 Excalidraw 以访问白板画布。

![Excalidraw 画布](/images/manual/use-cases/excalidraw-canvas.png#bordered)

你也可以点击 <i class="material-symbols-outlined">open_in_new</i> 在新浏览器标签页中打开 Excalidraw。

### 创建图表

1. 从工具栏中选择一个形状（矩形、椭圆、箭头、线条等）。
2. 从样式面板中自定义形状的描边颜色、描边宽度、描边样式、背景图案和不透明度。
3. 在画布上点击并拖动以绘制形状。

    ![在画布上绘制](/images/manual/use-cases/excalidraw-drawing.png#bordered)

4. 选择文本工具并在画布上点击以添加文本。

    ![添加文本](/images/manual/use-cases/excalidraw-text.png#bordered)

### 添加 Excalidraw 库

Excalidraw 库是一组可重用的图形元素。你无需从头绘制服务器、数据库或用户图标等常见元素，而是可以从导入的库中拖放使用。

1. 在 Excalidraw 编辑器中，点击右上角的 <span class="material-symbols-outlined">dock_to_left</span> 打开侧边栏。

2. 在库侧边栏中，点击 **Browse libraries** 打开官方 Excalidraw Libraries 网站。
    ![浏览库](/images/manual/use-cases/excalidraw-browse-libraries.png#bordered)

3. 搜索你需要的库，然后点击 **Add to Excalidraw**。
    ![添加到 Excalidraw](/images/manual/use-cases/excalidraw-add-library.png#bordered)

4. 返回编辑器，导入的库将出现在侧边栏中。从中拖动任何元素到画布上。
    ![导入的库](/images/manual/use-cases/excalidraw-imported-library.png#bordered)

### 保存你的作品

Excalidraw 支持将画布本地保存为 `.excalidraw` 文件，或导出为图片。

- **保存到本地**：点击左上角的 <span class="material-symbols-outlined">menu</span>，然后选择 **Save to** > **Save to disk**，将画布保存为 `.excalidraw` 文件，以便稍后重新打开。

    ![保存到磁盘](/images/manual/use-cases/excalidraw-save-to-disk.png#bordered)

- **导出为图片**：点击左上角的 <span class="material-symbols-outlined">menu</span>，然后选择 **Export image**，将画布保存为 PNG 或 SVG 文件，或复制到剪贴板。

    ![导出为图片](/images/manual/use-cases/excalidraw-export-image.png#bordered)

## 已知问题

### 不支持协作和分享

自托管版本的 Excalidraw 不支持实时协作或分享链接。官方自托管镜像仅包含前端客户端，无法连接 Excalidraw 的云后端进行协作和链接分享。Excalidraw 团队计划在未来提供完全可自托管的后端。详情请参见 [excalidraw#1772](https://github.com/excalidraw/excalidraw/issues/1772) 和 [excalidraw#8195](https://github.com/excalidraw/excalidraw/issues/8195)。

## 了解更多

- [Excalidraw 文档](https://docs.excalidraw.com)：官方 Excalidraw 文档和指南。
