---
outline: [2, 3]
title: Lares
description: 认识 Lares，Olares 官方 AI 助手。通过自然语言管理应用、文件和系统，深入展开研究。
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Olares Router, Olares 本地 AI, 本地 AI 智能体, AI 助手, 自然语言, Olares 1.12.7
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/lares.md)为准。
:::

# Lares

Lares 是 Olares 的官方 AI 助手，在 v1.12.7 中推出。有了 Router 和一个已连接的模型，你用自然语言说出一个目标，Lares 就会在你的设备上规划并执行任务。

作为完全自主的智能体，Lares 既能处理快速的日常任务，也能完成复杂的自动化。它不只是执行单条指令，而是拆解高层目标，动态串联工具，在设备上直接执行端到端的完整计划。从第一个请求到最终结果，全部在对话中完成。

## 学习目标

看完本页后，你将学会：

- 了解 Lares 如何连接 Router，并在你指定的权限下工作。
- 先在 Lares 里问一个简单的问题。
- 配置 Web Research 工具，运行复杂的研究任务。

## 前置条件

- **系统**：Olares OS v1.12.7 或更高版本
- **应用**：已从 Market 安装 Lares、Router 和 Qwen3.8-27B（llama.cpp）

## 了解 Lares 的工作原理

为了在设备上安全地执行任务，Lares 基于两个核心机制运作：内置的模型路由和严格的访问控制。

- **与 Router 集成**：Lares 自动接入你在 Router 中配置的 AI 模型和工具。它通过集群内身份完成认证，无需个人登录，也无需 API 密钥。
- **权限可控的执行**：Lares 通过 Olares CLI Agent Skills 代你执行任务，但范围由你掌控。你可以为它指定具体的权限级别，比如 **Read Only**、**Write** 或 **Full Access**，精确界定它在系统上可以执行哪些操作。

## 开始使用

1. 打开 Lares，选择模型，并确认它工作时的权限级别。

    ![Lares 聊天界面](/images/manual/use-cases/lares.png#bordered)

2. 从一个简单的问题开始：

   ```text
   检查一下这台设备的配置
   ```

    :::warning 重要：一次运行一个任务
    在 Olares One 上运行 Qwen3.8-27B（llama.cpp）时，建议一次只运行一个请求，以保证 100K 上下文窗口和模型精度的最佳体验。
    :::

## 运行深度研究任务

默认情况下，Lares 只使用 Olares 上已有的内容。要扩展它的能力，使用公网网页资料研究一个主题，可以在 Router 中配置 Web Research 工具，然后在 Lares 中选择并使用它。

### 1. 选择 Web Research 工具

Web research 有两种能力，Router 以标签形式展示在每个 Web Research 工具上：

| 标签 | 工具 | 能力 |
| ---- | ---- | ---- |
| `search` | <nobr>SearXNG、Serper、Tavily</nobr> | 找到相关页面，返回标题、摘要和来源链接。 |
| `scrape` | <nobr>Firecrawl、Tavily、Jina Reader</nobr> | 读取指定的 URL，把内容变成 Lares 可以分析、总结或保存的材料。 |

### 2. 获取连接信息

连接要求取决于你选择的工具：

- **SearXNG 和 Firecrawl**：它们以应用形式运行在你的 Olares 上。从 Market 安装应用，然后在 Olares 的 **Settings** > **[App-Name]** > **Entrances** > **Endpoint** 中找到 entrance URL。
- **Serper、Tavily 和 Jina Reader**：到服务商官网获取 API key。

### 3. 在 Router 中配置 Web Research 工具

以下步骤以 SearXNG 为例。

1. 打开 Router，进入 **Tools** > **Manage providers**。
2. 选择 **SearXNG**，然后填写以下设置：

    - **Provider name**：`localsearxng`
    - **SearXNG instance URL**：在第 2 步获取的 entrance URL

3. 点击 **Add**。工具会出现在 **Available** 列表中。
4. 点击 <i class="material-symbols-outlined">add</i> 启用它。

    ![在 Router 中启用 SearXNG](/images/manual/use-cases/router-search-tool-enable.png#bordered)

    工具会出现在 **Configured** 列表中。

    ![SearXNG 已启用](/images/manual/use-cases/router-search-tool-enabled.png#bordered)

### 4. 在 Lares 中选择搜索工具

只有带 `search` 能力的工具需要这一步。

:::info
所有 Web Research 工具都在 Router 中配置，但带 `search` 标签的工具还需要在 Lares 的 **Settings** > **Web Search** 中多走一步。只有 `scrape` 能力的工具不需要这一步，你让 Lares 读取指定 URL 时它会自动被调用。
:::

1. 打开 Lares，进入 **Settings** > **Web Search**。
2. 点击 **Refresh**，然后在 **Default search service** 列表中选择配置好的搜索工具。

### 5. 研究一个主题

先用一个简单的请求确认搜索工具工作正常。例如：

```text
用联网搜索找到 3 个关于使用 AI 写作工具时保持个人风格的公开来源。每个来源返回标题、URL 和一句话摘要。
```

Lares 会返回所选搜索服务的来源链接和摘要。

确认搜索可用后，就可以让 Lares 完成结合资料发现、页面阅读和综合分析的复杂研究任务。
