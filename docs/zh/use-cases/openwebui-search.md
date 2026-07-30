---
outline: deep
description: 在 Olares 上使用 SearXNG 和嵌入模型，为 Open WebUI 启用网页搜索，以检索最新信息。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 网页搜索, SearXNG, 嵌入模型, RAG
app_version: "1.0.38"
doc_version: "2.0"
doc_updated: "2026-07-30"
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/openwebui-search.md)。
:::

# 在 Open WebUI 中启用网页搜索

为 Open WebUI 添加网页搜索能力，让本地 AI 模型可以从互联网检索最新信息。该集成需要一个已连接的嵌入模型来生成嵌入向量，并使用 SearXNG 获取网页搜索结果。

如果你希望 Open WebUI 读取完整网页内容，而不是只使用搜索结果摘要，请配置 Firecrawl 等网页加载器。

## 学习目标

在本指南中，你将学习如何：

- 获取嵌入模型和 SearXNG 所需的端点 URL。
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

## 获取服务信息

要将 Open WebUI 与后台服务连接起来，你需要找到嵌入模型和 SearXNG 的连接端点。

### 获取嵌入模型信息

<!--@include: ../reusables/ai-service-connections.md#get-embedding-model-connection-details-openai-->

### 获取 SearXNG 端点

1. 打开 Olares 设置，然后前往**应用** > **SearXNG**。
2. 在**共享入口**下，点击 **SearXNG**，然后复制端点 URL。例如：`http://d1236e020.shared.olares.com`。

   ![SearXNG shared endpoint](/images/manual/use-cases/openwebui-searxng-shared-endpoint1.png#bordered){width=70%}

## 配置 Open WebUI

将获取到的信息填入 Open WebUI 配置面板。

### 设置文档嵌入

配置嵌入模型，使 Open WebUI 能够将文本转换为用于检索的向量表示。

1. 在 Open WebUI 中，选择你的头像图标，然后前往 **Admin Panel** > **Settings**。
2. 在左侧边栏中，找到 **Tools** 部分，然后选择 **Documents**。
3. 在 **Embedding** 区域中，指定以下设置：

   - **Embedding Model Engine**：选择 **OpenAI**。
   - **API Base URL**：输入你从 Model Console 复制的嵌入模型的 **Base URL**。
   - **Embedding Model**：输入你从 Model Console 复制的嵌入模型的 **Model name**。

4. 向下滚动到页面底部，然后点击右下角的 **Reindex** 以应用更改。
5. 选择 **Save**。

### 启用网页搜索

打开网页搜索并将其指向你的 SearXNG 端点。

1. 前往 **Admin Panel** > **Settings**。
2. 在左侧边栏中，找到 **Tools** 部分，然后选择 **Web Search**。
3. 指定以下设置：

   - **Web Search**：启用此设置。
   - **Web Search Engine**：选择 **SearXNG**。
   - **Searxng Query URL**：输入你的 SearXNG 端点 URL，并在末尾追加 `/search?q=<query>`。例如：`http://d1236e020.shared.olares.com/search?q=<query>`。
   - **Bypass Web Loader**：如果你只需要搜索结果摘要，请启用此设置。如果你希望 Open WebUI 通过网页加载器获取完整页面内容，请保持禁用。

      :::tip 全文检索
      若要检索完整网页，请安装 Firecrawl 并将其配置为网页加载器。参阅[将 Firecrawl 用作网页加载器](firecrawl.md#configure-open-webui)。
      :::

   ![SearXNG configurations in Open WebUI](/images/manual/use-cases/openwebui-searxng-config1.png#bordered)

4. 其他字段保持默认值。
5. 选择 **Save**。

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
