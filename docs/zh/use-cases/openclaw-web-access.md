---
outline: [2, 3]
description: 学习如何在 OpenClaw 中启用 SearXNG 网页搜索，让 AI 助手获取实时互联网信息。
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw 教程, OpenClaw 学习, OpenClaw 网页搜索
app_version: "1.0.8"
doc_version: "2.2"
doc_updated: "2026-07-29"
---

# 可选：在 OpenClaw 中启用网页搜索

默认情况下，OpenClaw 仅使用训练数据回答问题，无法获取时事新闻、实时资讯或在线网页内容。如需让助手具备联网搜索能力，可以为其接入网页搜索服务。

本指南以 SearXNG 为例。这是一款注重隐私的元搜索引擎，能够聚合多个来源的搜索结果，且不会追踪用户。你可以从 Olares 应用市场安装自托管的 SearXNG 实例。

## 学习目标

在本指南中，你将学习如何：
- 从 Olares 应用市场安装 SearXNG。
- 获取 SearXNG 的 Endpoint。
- 配置 OpenClaw 使用 SearXNG 进行网页搜索并获取搜索结果。
- 验证网页搜索工具是否正常工作。

## 步骤 1：安装 SearXNG

从应用市场安装 SearXNG。

1. 打开应用市场，搜索 "SearXNG"。

   ![SearXNG](/images/zh/manual/use-cases/searxng.png#bordered)

2. 点击**获取**，然后点击**安装**。等待安装完成。

## 步骤 2：获取 SearXNG Endpoint

OpenClaw 需要使用 SearXNG Endpoint 连接其搜索服务。

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

1. 前往 Olares **设置** > **应用** > **SearXNG** > **入口**。
2. 选择 **SearXNG**，然后复制 **Endpoint** URL。

## 步骤 3：配置 OpenClaw

将 OpenClaw 连接到 SearXNG。

1. 打开 OpenClaw CLI。
2. 运行以下命令，下载安装 `searxng` 插件：

   ```bash
   openclaw plugins install searxng
   ```

3. 运行以下命令重启网关，加载新安装的插件：

   ```bash
   restart-gateway
   ```

4. 重启完成后，运行以下命令启动配置向导：

    ```bash
    openclaw configure --section web
    ```

5. 按如下方式配置：

    | 配置 | 选项 |
    |:---------|:-------|
    | Where will the Gateway run | 选择 **Local (this machine)**。 |
    | Enable web_search | 选择 **Yes**。 |
    | Search provider | 选择 **SearXNG Search** 。|
    | SearXNG Base URL | 填写步骤 2 中复制的 SearXNG Endpoint URL。 |
    | Enable web_fetch (keyless HTTP fetch) | 选择 **Yes**。 |

## 步骤 4：验证网页搜索

测试助手是否能够从互联网获取实时信息。

1. 打开 Control UI，与助手开始对话。
2. 提出一个需要最新信息的问题。
3. 检查回复。如助手返回了最新信息，说明网页搜索集成已正常工作。

   ![使用 SearXNG 的网页搜索结果](/images/manual/use-cases/openclaw-web-search-results1.png#bordered)

:::tip 全文检索
SearXNG 仅返回标题、URL 和摘要，不会返回完整页面内容。获取完整文本可能受反爬取机制限制。如需让助手读取网页完整内容，建议使用在线网页服务。推荐 Firecrawl 和 Tavily，二者可返回完整文本或答案摘要，并提供免费搜索额度。
:::
