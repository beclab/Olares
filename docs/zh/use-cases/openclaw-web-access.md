---
outline: [2, 3]
description: 学习如何在 OpenClaw 中启用 SearXNG 网页搜索，让 AI 助手获取实时互联网信息。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 学习, OpenClaw 网页搜索
app_version: "1.0.2"
doc_version: "2.0"
doc_updated: "2026-06-09"
---

# 可选：在 OpenClaw 中启用网页搜索

默认情况下，OpenClaw 仅使用训练数据回答问题，无法获取时事新闻、实时资讯或在线网页内容。如需让助手具备联网搜索能力，可以为其接入网页搜索服务。

本指南以 SearXNG 为例。这是一款注重隐私的元搜索引擎，能够聚合多个来源的搜索结果，且不会追踪用户。你可以从 Olares 应用市场安装自托管的 SearXNG 实例。

## 学习目标

在本指南中，你将学习如何：
- 从 Olares 应用市场安装 SearXNG。
- 获取 SearXNG 的共享端点 URL。
- 配置 OpenClaw 使用 SearXNG 进行网页搜索并获取搜索结果。
- 验证网页搜索工具是否正常工作。

## 步骤 1：安装 SearXNG

安装 SearXNG 并获取其共享端点 URL。

1. 打开应用市场，搜索 "SearXNG"。

   ![SearXNG](/images/zh/manual/use-cases/searxng.png#bordered)

2. 点击**获取**，然后点击**安装**。等待安装完成。
3. 打开设置，进入**应用** > **SearXNG**。
4. 在**共享入口**中，点击 **SearXNG**。

   ![获取 SearXNG 共享端点](/images/zh/manual/use-cases/searxng-shared-laresprime.png#bordered){width=90%}

5. 复制保存共享端点 URL。例如：

   ```text
   http://d1236e020.shared.olares.com
   ```

## 步骤 2：配置 OpenClaw

将 OpenClaw 连接到 SearXNG。

1. 打开 OpenClaw CLI。
2. 运行以下命令启动配置向导：

    ```bash
    openclaw configure --section web
    ```

3. 按如下方式配置：

    | 配置 | 选项 |
    |:---------|:-------|
    | Where will the Gateway run | 选择 **Local (this machine)**。 |
    | Enable web_search | 选择 **Yes**。 |
    | Search provider | 选择 **SearXNG Search** 。|
    | SearXNG Base URL | 填写[步骤 1](#步骤-1-安装-searxng) 中获取的 SearXNG 共享端点 URL。 |
    | Enable web_fetch (keyless HTTP fetch) | 选择 **Yes**。 |

## 步骤 3：验证网页搜索

测试助手是否能够从互联网获取实时信息。

1. 打开 Control UI，与助手开始对话。
2. 提出一个需要最新信息的问题。
3. 检查回复。如助手返回了最新信息，说明网页搜索集成已正常工作。

   ![使用 SearXNG 的网页搜索结果](/images/manual/use-cases/openclaw-web-search-results1.png#bordered)

:::tip 全文检索
SearXNG 仅返回标题、URL 和摘要，不会返回完整页面内容。获取完整文本可能受反爬取机制限制。如需让助手读取网页完整内容，建议使用在线网页服务。推荐 Firecrawl 和 Tavily，二者可返回完整文本或答案摘要，并提供免费搜索额度。
:::
