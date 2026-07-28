---
outline: deep
description: 在 Olares 上使用 SearXNG 和嵌入模型，为 Open WebUI 启用网页搜索，以检索最新信息。
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, 网页搜索, SearXNG, 嵌入模型, RAG
app_version: "1.0.20"
doc_version: "1.1"
doc_updated: "2026-06-02"
---

# 在 Open WebUI 中启用网页搜索

为 Open WebUI 添加网页搜索能力，让本地 AI 模型可以从互联网检索最新信息。该集成需要一个可用的嵌入模型来处理文档，并使用 SearXNG 获取网页搜索结果。

如果你希望 Open WebUI 读取完整网页内容，而不是只使用搜索结果摘要，请配置 Firecrawl 等网页加载器。

## 学习目标

在本指南中，你将学习如何：

- 获取嵌入模型和 SearXNG 所需的端点 URL。
- 在 Open WebUI 中配置文档嵌入和网页搜索设置。
- 在聊天过程中执行带网页辅助的搜索。

## 前提条件

开始前，确保已满足以下条件：

- 已安装并配置 [Open WebUI](openwebui.md)，且至少有一个可用的模型后端。
- 已安装 SearXNG。
- 已安装嵌入模型应用，例如 **Qwen3 Embedding 0.6B (Ollama)**。
- 拥有 Open WebUI 实例的管理员权限。

## 获取服务信息

要将 Open WebUI 与后台服务连接起来，你需要找到嵌入模型和 SearXNG 的连接端点。

### 获取嵌入模型信息
1. 从启动台打开 Qwen3 Embedding 0.6B (Ollama)。
2. 记录主页面上显示的准确模型名称。例如：`qwen3-embedding:0.6b`。

   ![Qwen3 Embedding 0.6B](/images/manual/use-cases/qwen3-embedding.png#bordered)

3. 打开 Olares 设置，然后前往**应用** > **Qwen3 Embedding 0.6B (Ollama)**。
4. 在**共享入口**下，点击 **Qwen3 Embedding 0.6B**，然后复制端点 URL。例如：`http://eae5afcf0.shared.olares.com`。

### 获取 SearXNG 端点

1. 打开 Olares 设置，然后前往**应用** > **SearXNG**。
2. 在**共享入口**下，点击 **SearXNG**，然后复制端点 URL。例如：`http://d1236e020.shared.olares.com`。

   ![SearXNG shared endpoint](/images/manual/use-cases/openwebui-searxng-shared-endpoint.png#bordered){width=70%}

## 配置 Open WebUI

将获取到的信息填入 Open WebUI 配置面板。

### 设置文档嵌入
1. 在 Open WebUI 中，选择你的头像图标，然后前往 **Admin Panel** > **Settings** > **Documents**。
2. 在 **Embedding** 区域中，指定以下设置：

   - **Embedding Model Engine**：选择 **Ollama**。
   - **API Base URL**：输入你之前记录的嵌入模型端点 URL。
   - **Embedding Model**：输入你之前记录的嵌入模型名称。

3. 向下滚动到页面底部，然后点击右下角的 **Reindex** 以应用更改。
4. 选择 **Save**。

### 启用网页搜索

1. 前往 **Admin Panel** > **Settings** > **Web Search**。
2. 指定以下设置：

   - **Web Search**：启用此设置。
   - **Web Search Engine**：选择 **SearXNG**。
   - **Searxng Query URL**：输入你的 SearXNG 端点 URL，并在末尾追加 `/search?q=<query>`。

      例如：`http://d1236e020.shared.olares.com/search?q=<query>`。
   - **Bypass Web Loader**：如果你只需要搜索结果摘要，请启用此设置。如果你希望 Open WebUI 通过网页加载器获取完整页面内容，请保持禁用。

      :::tip 全文检索
      若要检索完整网页，请安装 Firecrawl 并将其配置为网页加载器。参阅[将 Firecrawl 用作网页加载器](firecrawl.md#configure-open-webui)。
      :::

   ![SearXNG configurations in Open WebUI](/images/manual/use-cases/openwebui-searxng-config.png#bordered)

3. 其他字段保持默认值。
4. 选择 **Save**。

## 验证配置

测试该功能，确保 AI 可以成功从网页检索最新信息。

1. 开始一个新聊天。
2. 选择模型。
3. 点击消息输入框下方的 **Integrations** 图标，然后启用 **Web Search**。

   ![Web search enable in Open WebUI chat](/images/manual/use-cases/openwebui-web-search-enable.png#bordered)

4. 输入一个需要最新信息的提示词。例如：

   ```plain
   What’s the latest news about Olares One
   ```
5. 提交提示词。AI 会生成包含检索到的搜索结果及其来源链接的回复。

   ![Web search results in Open WebUI](/images/manual/use-cases/openwebui-web-search-results.png#bordered)
