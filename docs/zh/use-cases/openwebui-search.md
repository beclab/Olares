---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/openwebui-search
outline: deep
description: 在 Olares 上使用 SearXNG 和嵌入模型，为 Open WebUI 启用网页搜索，以检索最新信息。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 网页搜索, SearXNG, 嵌入模型, RAG
app_version: "1.0.38"
doc_version: "2.1"
doc_updated: "2026-11-18"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/openwebui-search.md)为准。
:::

# 在 Open WebUI 中启用网页搜索

<VersionRouteSelect />

为 Open WebUI 添加网页搜索能力，让本地 AI 模型可以从互联网检索最新信息。该集成需要一个已连接的嵌入模型来生成嵌入向量，并使用 SearXNG 获取网页搜索结果。

如果你希望 Open WebUI 读取完整网页内容，而不是只使用搜索结果摘要，请配置 Firecrawl 等网页加载器。

## 学习目标

在本指南中，你将学习如何：

- 为嵌入模型使用 Router，并获取 SearXNG 的入口地址。
- 在 Open WebUI 中配置文档嵌入和网页搜索设置。
- 在聊天过程中执行带网页辅助的搜索。

## 前提条件

开始前，确保已满足以下条件：

- 已安装并配置 [Open WebUI](openwebui.md)，且至少连接了一个本地模型。
- 已安装 SearXNG。
- 拥有 Open WebUI 实例的管理员权限。
- 以下模型：
   | 模型类型 | 模型 | 获取方式 |
   | :--- | :--- | :--- |
   | Embedding | EmbeddingGemma | 从应用市场安装 |

## 配置嵌入模型

<!--@include: ../reusables/openwebui-embeddings.md#configure-embeddings-->

## 连接 SearXNG {#启用网络搜索-1}

Open WebUI 直接连接 SearXNG 获取搜索结果。此处使用 SearXNG 入口地址。

1. 打开 Olares 设置，进入**应用** > **SearXNG** > **入口**。
2. 点击 **SearXNG**，复制入口地址。例如：

   ```text
   https://84a93c3c.laresprime.olares.com
   ```

   ![SearXNG 入口](/images/manual/use-cases/openwebui-searxng-shared-endpoint1.png#bordered){width=70%}

3. 在 Open WebUI 中打开 **Admin Panel** > **Settings**。
4. 在左侧边栏找到 **Tools**，选择 **Web Search**。
5. 填写以下设置：

   - **Web Search**：启用。
   - **Web Search Engine**：选择 **SearXNG**。
   - **Searxng Query URL**：在刚才复制的 SearXNG 入口地址后添加 `/search?q=<query>`。例如：

     ```text
     https://84a93c3c.laresprime.olares.com/search?q=<query>
     ```

   - **Bypass Web Loader**：只需要搜索结果摘要时启用。需要读取完整网页内容时关闭。

     若要读取完整网页，先[配置 Firecrawl 网页加载器](firecrawl.md#configure-open-webui)。

   ![Open WebUI 中的 SearXNG 设置](/images/manual/use-cases/openwebui-searxng-config1.png#bordered)

6. 其他字段保持默认值，点击 **Save**。

## 验证配置

测试该功能，确保 AI 可以成功从网页检索最新信息。

1. 开始一个新聊天。
2. 选择模型。
3. 点击聊天输入框下方的 **Integrations** 图标，然后启用 **Web Search**。

   ![Web search enable in Open WebUI chat](/images/manual/use-cases/openwebui-web-search-enable1.png#bordered)

4. 输入一个需要最新信息的提示词。例如：

   ```plain
   What’s the latest news about Olares One
   ```
5. 提交提示词。AI 会生成包含检索到的搜索结果及其来源链接的回复。

   ![Web search results in Open WebUI](/images/manual/use-cases/openwebui-web-search-results1.png#bordered)

## 了解更多

- [Open WebUI 常见问题](openwebui-common-issues.md#open-webui-无法搜索网页)：解决 Open WebUI 的常见问题。
