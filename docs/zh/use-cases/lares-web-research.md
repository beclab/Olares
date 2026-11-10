---
outline: [2, 3]
description: 为 Lares 连接联网研究服务，搜索公开来源，检查返回的链接，并在对话中综合研究结果。
head:
  - - meta
    - name: keywords
      content: Olares, Lares, 联网研究, SearXNG, Firecrawl, Tavily, 网页搜索, 本地 AI
app_version: "0.26.60"
doc_version: "1.0"
doc_updated: "2026-11-07"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/lares-web-research.md)为准。
:::

# 使用 Lares 开展联网研究

Lares 可以通过 Router 中配置的搜索和网页读取服务使用公开网页内容。搜索服务用于查找相关页面，抓取服务用于读取指定页面，供 Lares 分析或总结。

本指南将 SearXNG 连接为搜索服务，先通过一个小型来源查找任务检查连接，再使用搜索结果完成范围更大的研究请求。

## 学习目标

通过本指南，你将学会：

- 区分搜索和抓取服务。
- 在 Router 中配置联网研究服务。
- 在 Lares 中选择搜索服务并检查来源。

## 前提条件

- 已[为 Lares 准备支持工具调用的对话模型](lares.md)。
- 一个联网研究服务。本指南使用从 Market 安装的 SearXNG。

## 选择联网研究服务

Router 按提供的能力为服务商添加标签。

| 能力 | 服务商 | Lares 可以完成的操作 |
|:---|:---|:---|
| `search` | SearXNG、Serper、Tavily | 查找页面并返回标题、摘要和链接。 |
| `scrape` | Firecrawl、Tavily、Jina Reader | 读取指定 URL，用于分析、总结或保存。 |

SearXNG 和 Firecrawl 以 Olares 应用形式运行。从 Market 安装应用，然后在 Olares **Settings** > **Applications** > **[应用名]** > **Entrances** 中复制端点。Serper、Tavily 和 Jina Reader 需要从服务商处获取凭据。

## 在 Router 中连接 SearXNG

1. 打开 Olares **Settings** > **Applications** > **SearXNG** > **Entrances**。
2. 选择 SearXNG 入口，复制 **Endpoint** 地址。
3. 打开 Router，进入 **Tools** > **Manage providers**。
4. 选择 **SearXNG**，填写：

   - **Provider name:** `localsearxng`
   - **SearXNG instance URL:** 第 2 步复制的端点

5. 点击 **Add**。服务商会出现在 **Available** 下。
6. 点击 <i class="material-symbols-outlined">add</i> 启用它。

   ![在 Router 中启用 SearXNG](/images/manual/use-cases/router-search-tool-enable.png#bordered)

   服务商会出现在 **Configured** 下。

   ![Router 中已启用 SearXNG](/images/manual/use-cases/router-search-tool-enabled.png#bordered)

## 在 Lares 中选择搜索服务

带 `search` 能力的服务需要在 Lares 中选中。只有抓取能力的服务会在你要求 Lares 读取指定 URL 时调用，不会显示为默认搜索服务。

1. 打开 Lares，进入 **Settings** > **Web Search**。
2. 点击 **Refresh**。
3. 在 **Default search service** 中选择 `localsearxng`。

## 检查连接

先发送一个小型请求，以便逐个检查返回的链接：

```text
用联网搜索找到 3 个关于使用 AI 写作工具时保持个人风格的公开来源。每个来源返回标题、URL 和一句话摘要。
```

打开链接，确认页面内容与摘要相符。如果没有返回链接，请先确认 SearXNG 可以直接搜索，再刷新 Lares 中的服务商列表。

## 开展范围更大的研究

来源检查成功后，可以让 Lares 搜索、阅读并综合资料。明确写出问题、来源要求和输出形式。例如：

```text
研究使用 AI 编辑工具时保持作者个人风格的实用方法。使用至少 5 个公开来源，其中至少 2 个来自写作或出版机构的一手来源。比较这些方法，指出取舍，并生成一份带有行内来源链接的简明报告。
```

使用报告前，请检查引用页面。搜索连接成功只能说明 Lares 能获取结果，不能证明生成的每项说法都得到来源支持。

## 了解更多

- [开始使用 Lares](lares.md)：准备模型并运行第一个只读任务。
- [设置 SearXNG](searxng.md)：安装并直接使用搜索服务。
- [使用 Firecrawl](firecrawl.md)：准备读取网站内容的服务。
