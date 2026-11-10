---
outline: [2, 3]
description: 开始在 Olares 上使用 Lares。通过 Router 连接模型，选择访问权限，完成第一个任务，并找到联网研究和应用自动化指南。
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Olares Router, 本地 AI 智能体, AI 助手, MCP, 自然语言, Olares 1.12.7
app_version: "0.26.60"
doc_version: "2.0"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/lares.md)为准。
:::

# 开始使用 Lares

Lares 是 Olares 官方 AI 助手。它使用 Router 连接的模型，并通过由你控制的工具操作 Olares 应用、文件和系统服务。

先用一个小型只读任务检查模型和权限。随后可以按需添加联网研究服务，或连接 Blender、YuE2 Studio 等应用，完成特定工作流。

## 学习目标

通过本指南，你将学会：

- 通过 Router 为 Lares 准备对话模型。
- 选择访问权限并运行第一个任务。
- 找到适合联网研究或应用连接的配置指南。

## 前提条件

- Olares 1.12.7 或更高版本。
- 已从 Market 安装 Lares 0.26.60 和 Router。
- 已安装支持工具调用的对话模型，且模型在 Router 中可用。本指南使用 Qwen3.8-27B (llama.cpp)。

## 了解 Lares 如何连接 Olares

Lares 组合了三类能力。每类能力都应在对应位置配置。

| 能力 | 连接方式 | 配置位置 |
|:---|:---|:---|
| 对话与规划 | Router 提供的对话模型 | **Router** > **Default models** |
| Olares 操作 | 内置的 Olares 工具和技能 | Lares 中选择的访问权限 |
| 应用专属操作 | 应用的 MCP 服务或内置集成 | **Lares** > **Settings** > **MCP**，或对应应用指南 |

Router 提供模型，Lares 判断何时调用可用工具。你选择的访问权限会限制 Lares 能代你执行的操作。MCP 连接只会添加某个应用的工具，不会改变其他应用的权限。

## 准备模型

1. 从 Launchpad 打开 Router，进入 **Default models**。
2. 将 Qwen3.8-27B (llama.cpp) 或其他支持工具调用的模型设为默认对话模型。
3. 打开 Lares 前，确认模型显示 **Callable**。

如果使用其他模型，请确认它支持工具调用。仅支持对话的模型可以回答问题，但无法可靠地操作 Olares 或已连接的 MCP 服务。

## 运行第一个任务

1. 从 Launchpad 打开 Lares。
2. 选择对话模型，并检查当前对话的访问权限。

   ![Lares 对话界面](/images/manual/use-cases/lares1.png#bordered)

3. 从只读请求开始：

   ```text
   检查这台设备的配置。汇总操作系统版本、处理器、内存、存储和加速器资源。不要进行任何修改。
   ```

4. 允许执行前检查 Lares 提出的工具调用。确认回复描述了设备信息，且没有修改设置。

:::warning 一次只运行一个高负载任务
在 Olares One 上运行 Qwen3.8-27B (llama.cpp) 时，请等待当前请求完成，再启动另一个模型或媒体生成任务。并发工作负载可能争用内存，导致模型无响应。
:::

## 添加所需能力

围绕要完成的任务配置 Lares。以下每篇指南都包含所需应用、连接方法和第一个实用结果。

- [使用 Lares 开展联网研究](lares-web-research.md)：查找公开来源，并在对话中综合信息。
- [使用 Lares 创建 Blender 场景](blender-lares.md)：搭建并渲染 3D 场景。
- [通过 Lares 使用 FlowStudio](flowstudio-lares.md)：通过预设工作流生成图片和视频。
- [通过 Lares 使用 YuE2 Studio 创作歌曲](yue2-studio-lares.md)：通过 YuE2 Studio MCP 服务生成歌曲。

应用专属指南是端点、身份验证、就绪检查和输出位置的准确信息来源。不要复用其他应用的 MCP 地址或令牌。

## 了解更多

- [将 Olares Router 用作 AI 网关](olares-router.md)：管理 Lares 使用的模型和服务商。
- [在 Olares 上运行本地大模型](llm-base-apps.md)：安装并管理本地对话模型。
