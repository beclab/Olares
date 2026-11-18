---
outline: [2, 3]
description: 在 Olares 上安装 Blender，打开串流桌面，并将可编辑项目保存到 Olares Files。
head:
  - - meta
    - name: keywords
      content: Olares, Blender, 3D 创作, 串流桌面, Olares Files, 自托管
app_version: "0.1.27"
doc_version: "2.0"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/blender.md)为准。
:::

# 在 Olares 上开始 Blender 项目

Blender 是一套用于建模、动画和渲染的 3D 创作工具。在 Olares 上，它以串流桌面的形式运行。你可以在浏览器中操作，项目文件则保留在自己的设备上。

本指南介绍如何安装 Blender、打开桌面，并将一个起始项目保存到 Olares Files。之后，你可以继续使用 Blender 界面编辑，也可以让 Lares 通过 MCP 操作同一个场景。

## 学习目标

通过本指南，你将学会：

- 选择计算模式并安装 Blender。
- 打开 Blender 串流桌面。
- 通过 Olares Files 保存并重新打开项目。

## 前提条件

- Olares 1.12.7 或更高版本。

## 安装 Blender

1. 打开 Market，搜索“Blender”。

   <!-- ![Market 中的 Blender](/images/manual/use-cases/blender.png#bordered) -->

2. 点击 **Get**，然后点击 **Install**。
3. 根据设备和工作负载选择计算模式：

   - **Intel:** 使用可用的 Intel 核显处理视口和桌面串流，把独立显卡留给其他应用。
   - **CPU:** 没有合适的 GPU 时使用。桌面串流会采用软件编码，因此占用更多 CPU。
   - **NVIDIA:** 为负载较高的渲染任务分配 NVIDIA GPU。其他应用同时使用同一张 GPU 时，性能可能下降。

4. 确认选择并等待安装完成。

之后可以通过 Olares 资源设置更改分配的加速器。使用其他资源配额重启 Blender 前，请先保存当前项目。

## 打开 Blender

1. 从 Launchpad 打开 Blender。
2. 等待默认场景出现。新场景通常包含 `Camera`、`Cube` 和 `Light`。

浏览器中显示的是串流桌面。需要更低延迟的连接时，请先在电脑上启用 LarePass VPN，再打开 Blender。

:::tip Olares 和电脑在同一局域网？
也可以使用 Olares 的 `.local` 地址。地址格式和 Windows 配置方法请参阅[通过本地地址访问 Olares 服务](../manual/best-practices/local-access.md)。
:::

## 保存起始项目

1. 在 Files 中创建 **Home** > **blender-projects**。
2. 在 Blender 中点击 **File** > **Save As**。
3. 在文件浏览器中打开 `/config/Home/blender-projects`。
4. 将文件名设为 `starter-scene.blend`，点击 **Save As**。
5. 如果 Blender 要求确认，再次点击 **Save As**。
6. 打开 Files，前往 **Home** > **blender-projects**，确认其中出现 `starter-scene.blend`。

Blender 中的 `/config/Home` 对应 Files 中的 **Home**。需要在 Blender 桌面之外预览、下载或复用文件时，请把项目和输出保存到这里。

## 重新打开已保存的项目

1. 从 Launchpad 打开 Blender。
2. 点击 **File** > **Open**。
3. 打开 `/config/Home/blender-projects/starter-scene.blend`。
4. 继续编辑，然后点击 **File** > **Save** 更新同一个项目。

## 了解更多

- [使用 Lares 创建 Blender 场景](blender-lares.md)：通过 MCP 连接 Blender，并在对话中搭建和渲染太阳系场景。
- [通过本地地址访问 Olares 服务](../manual/best-practices/local-access.md)：在局域网中降低串流桌面的延迟。
- [管理加速资源](../manual/olares/settings/gpu-resource.md)：调整分配给 Blender 的 GPU 资源。
