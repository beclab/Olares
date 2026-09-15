---
outline: [2, 3]
title: Lares
description: 认识 Lares，Olares 官方 AI 助手。通过自然语言管理应用、文件和系统，深入展开研究。
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Olares Router, Olares 本地 AI, 本地 AI 智能体, AI 助手, 自然语言, Olares 1.12.7
---

# Lares

Lares 是 Olares 的官方 AI 助手，在 v1.12.7 中推出。有了 Router 和一个已连接的模型，你用自然语言说出一个目标，Lares 就会在你的设备上规划并执行任务。

作为完全自主的智能体，Lares 既能处理快速的日常任务，也能完成复杂的自动化。它不只是执行单条指令，而是拆解高层目标，动态串联工具，在设备上直接执行端到端的完整计划。从第一个请求到最终结果，全部在对话中完成。

## 开始之前

请先确认以下条件和注意事项。

- **前置条件**：确保已安装 Router，且至少连接了一个模型。
- **一次一个任务**：本地 AI 模型通过时间片共享加速卡资源，一次只能处理一个请求。
    - **Lares 内的排队**：Lares 一次处理一个任务。任务运行期间，新请求会排队等待，当前任务结束后自动开始。
    - **同时运行多个智能体**：要同时运行 Lares 和其他智能体，请将它们连接到不同的模型。

## 了解 Lares 的工作原理

为了在设备上安全地执行任务，Lares 基于两个核心机制运作：内置的模型路由和严格的访问控制。

- **与 Router 集成**：Lares 自动接入你在 Router 中配置的 AI 模型和工具。它通过集群内身份完成认证，无需个人登录，也无需 API 密钥。
- **权限可控的执行**：Lares 通过 Olares CLI Agent Skills 代你执行任务，但权限范围由你掌控。你可以为它指定具体的权限级别，比如 **Read Only**、**Write** 或 **Full Access**，精确界定它在系统上可以执行哪些操作。

## 开始使用

1. 打开 Lares，选择模型，并确认它工作时的权限级别。

    ![Lares 聊天界面](/images/manual/use-cases/lares.png#bordered)

2. 提出你的第一个问题：

   ```text
   检查一下这台设备的配置
   ```

    :::warning 一次一个请求
    Lares 一次处理一个请求。如果你在任务完成前又发了一条消息，它会排队等候。要同时运行 Lares 和其他智能体，请将它们连接到不同的模型。
    :::

## 运行深度研究任务

默认情况下，Lares 只使用 Olares 上已有的内容。要使用公网网页资料研究一个主题，先在 Router 中配置 Web Research 工具，然后在 Lares 中选择并使用它。

### 1. 选择 Web Research 工具

Web research 有两种能力，Router 以标签形式展示在每个 Web Research 工具上：

| 标签 | 工具 | 能力 |
| ---- | ---- | ---- |
| `search` | <nobr>SearXNG、Serper、Tavily</nobr> | 找到相关页面，返回标题、摘要和来源链接。 |
| `scrape` | <nobr>Firecrawl、Tavily、Jina Reader</nobr> | 读取你提供的页面，把内容变成 Lares 可以分析、总结或保存的材料。 |

### 2. 获取连接信息

连接要求取决于你选择的工具：

- **SearXNG 和 Firecrawl**：它们以应用形式运行在你的 Olares 上。从 Market 安装应用，然后在 Olares 的 **Settings** > **[App-Name]** > **Entrances** > **Endpoint** 中找到 entrance URL。
- **Serper、Tavily 和 Jina Reader**：到服务商官网获取 API key。

### 3. 在 Router 中配置 Web Research 工具

以下步骤以 SearXNG 为例。

1. 打开 Router，进入 **Tools** > **Manage providers**。
2. 选择 **SearXNG**，然后填写以下设置：

    - **Provider name**：`localsearxng`
    - **SearXNG instance URL**：在第 2 步准备好的 entrance URL

3. 点击 **Add**。工具会出现在 **Available** 列表中。
4. 点击 <i class="material-symbols-outlined">add</i> 启用它。

    ![在 Router 中启用 SearXNG](/images/manual/use-cases/router-search-tool-enable.png#bordered)

5. 确认配置好的工具已启用，并出现在 **Configured** 列表中。

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

Lares 会返回所选搜索服务的来源链接和摘要。确认搜索可用后，就可以让 Lares 完成结合资料发现、页面阅读和综合分析的复杂研究任务。