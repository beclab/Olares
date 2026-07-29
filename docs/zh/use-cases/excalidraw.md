---
outline: [2, 3]
title: 在 Olares 上使用 Excalidraw 创建图表和线框图
description: 在 Olares 上安装 Excalidraw，在自托管虚拟白板中创建手绘风格图表、线框图、流程图和草图。
head:
  - - meta
    - name: keywords
      content: Olares, Excalidraw, 白板, 图表, 线框图, 手绘风格, 草图, 自托管, 流程图, 虚拟白板
app_version: "1.0.8"
doc_version: "1.1"
doc_updated: "2026-07-29"
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../use-cases/excalidraw.md)。
:::

# 使用 Excalidraw 创建手绘风格图表

Excalidraw 是一个开源虚拟白板，可以生成手绘风格的图表。你可以用它直接在 Olares 设备上绘制线框图、流程图、系统架构图或任何自由形式的草图。

手绘风格让图表更易亲近，帮助团队关注想法本身而非像素级完美效果。由于 Excalidraw 运行在 Olares 上，你的草图保留在自己的基础设施中。

本文档将介绍如何安装 Excalidraw、使用画布、导入素材库以及导出作品。

## 学习目标

在本指南中，你将学习如何：
- 在 Olares 上安装 Excalidraw。
- 在画布上创建和设置形状样式。
- 从 Excalidraw 库导入可复用元素。
- 保存和导出图表。

## 前提条件

- 已设置并运行 Olares 设备。
- 用于打开 Excalidraw 的浏览器。

## 安装 Excalidraw

1. 打开 Market，搜索 "Excalidraw"。
   ![安装 Excalidraw](/images/manual/use-cases/excalidraw.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 打开 Excalidraw

1. 从 Launchpad 打开 Excalidraw。
   
   白板画布将在浏览器中打开。

   ![Excalidraw 画布](/images/manual/use-cases/excalidraw-canvas.png#bordered)

2. 点击 <i class="material-symbols-outlined">open_in_new</i> 在新浏览器标签页中打开 Excalidraw。

## 创建图表

Excalidraw 在画布左侧提供工具栏。你可以选择工具、绘制形状并添加文本。

1. 从工具栏中选择形状工具。选项包括矩形、椭圆、菱形、箭头、线条和自由手绘。
2. 在画布上点击并拖动以绘制形状。
3. 使用样式面板自定义所选形状：
    - **描边颜色**：更改轮廓颜色。
    - **描边宽度**：让轮廓更粗或更细。
    - **描边样式**：使用实线、虚线或点线。
    - **背景**：用颜色或图案填充形状。
    - **不透明度**：调整透明度。

    ![在画布上绘制](/images/manual/use-cases/excalidraw-drawing.png#bordered)

4. 选择文本工具并在画布上点击以添加标签。
5. 拖动形状以重新定位。使用选择工具调整大小或旋转对象。

    ![添加文本](/images/manual/use-cases/excalidraw-text.png#bordered)

:::tip 提示
绘制时按住 **Shift** 键，可将矩形约束为正方形，或将椭圆约束为圆形。
:::

## 使用 Excalidraw 库

Excalidraw 库是可重用图形的集合。你无需从头绘制服务器、数据库或用户图标等常见元素，可以直接导入使用。

1. 在 Excalidraw 编辑器中，点击右上角的 <span class="material-symbols-outlined">dock_to_left</span> 打开库侧边栏。
2. 点击 **Browse libraries** 打开官方 Excalidraw Libraries 网站。
    ![浏览库](/images/manual/use-cases/excalidraw-browse-libraries.png#bordered)

3. 搜索你需要的库，然后点击 **Add to Excalidraw**。
    ![添加到 Excalidraw](/images/manual/use-cases/excalidraw-add-library.png#bordered)

4. 返回编辑器，导入的库将出现在侧边栏中。从中拖动任何元素到画布上。
    ![导入的库](/images/manual/use-cases/excalidraw-imported-library.png#bordered)

:::tip 提示
浏览用户流程图标、云基础设施符号或移动设备模型库，以加快线框图绘制速度。
:::

## 保存和导出作品

Excalidraw 支持将画布本地保存或导出为图片。

### 保存到磁盘

1. 点击左上角的 <span class="material-symbols-outlined">menu</span>。
2. 选择 **Save to** > **Save to disk**。
3. 选择位置并将文件保存为 `.excalidraw`。

    ![保存到磁盘](/images/manual/use-cases/excalidraw-save-to-disk.png#bordered)

你可以稍后通过同一菜单中的 **Open** 重新打开 `.excalidraw` 文件。

### 导出为图片

1. 点击左上角的 <span class="material-symbols-outlined">menu</span>。
2. 选择 **Export image**。
3. 选择 PNG、SVG，或将图片复制到剪贴板。

    ![导出为图片](/images/manual/use-cases/excalidraw-export-image.png#bordered)

SVG 导出可让你的图表在任何尺寸下保持清晰。PNG 导出适合在文档或演示文稿中分享。

## 让图表更清晰的技巧

遵循以下做法，让你的图表更易阅读：

- **使用一致的颜色**。将调色板限制在两到三种颜色加黑色。
- **对齐对象**。从菜单中启用网格，让形状自动对齐。
- **组合相关项目**。选择多个形状并组合，使它们一起移动。
- **为连接器添加标签**。在箭头上添加简短文本标签，让流程显而易见。
- **保持间距均匀**。在相关组周围留出相似数量的空白。

## 已知问题

### 不支持协作和分享

自托管版本的 Excalidraw 不支持实时协作或分享链接。官方自托管镜像仅包含前端客户端，无法连接 Excalidraw 的云后端进行协作和链接分享。Excalidraw 团队计划在未来提供完全可自托管的后端。详情请参见 [excalidraw#1772](https://github.com/excalidraw/excalidraw/issues/1772) 和 [excalidraw#8195](https://github.com/excalidraw/excalidraw/issues/8195)。

## 了解更多

- [Excalidraw 文档](https://docs.excalidraw.com)：官方 Excalidraw 文档和指南。
- [使用 Open Notebook 创建研究笔记](open-notebook.md)：在 Olares 上整理研究材料和笔记。
