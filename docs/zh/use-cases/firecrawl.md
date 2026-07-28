---
outline: [2, 3]
description: 在 Olares 上设置 Firecrawl，作为 Open WebUI 等应用的网页加载器，或使用其 API 抓取和爬取网站。
head:
  - - meta
    - name: keywords
      content: Olares, Firecrawl, web crawler, web scraping, Firecrawl v2, scrape API, crawl API, Open WebUI, web loader, self-hosted
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-06-02"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/firecrawl.md)为准。
:::

# 使用 Firecrawl 作为网页加载器

Firecrawl 是一个无头网页数据服务，可将网页转换为干净的 Markdown、结构化 JSON、摘要和元数据。在 Olares 上，Open WebUI 等应用可以使用 Firecrawl 在找到搜索结果后加载完整的网页内容。

你也可以直接调用 Firecrawl API 来测试抓取和爬取功能。

## 安装 Firecrawl

1. 打开 Market，搜索 "Firecrawl"。

   ![Firecrawl](/images/manual/use-cases/firecrawl.png#bordered)

2. 点击 **获取**，然后点击 **安装**，等待安装完成。

## 在其他应用中使用 Firecrawl

Firecrawl 通常在后台运行。其他应用在需要获取和清理网页内容时，通过其端点 URL 调用它。

### 获取 Firecrawl 端点

端点是 Firecrawl 服务在 Olares 上的基础地址。你将在此地址后添加 `/v2/scrape` 或 `/v2/crawl` 等 API 路径。

1. 打开 **Settings**，然后导航至 **Applications** > **Firecrawl**。
2. 在 **Entrances** 下，点击 **Firecrawl**。
3. 在 **Endpoint settings** 下，找到端点 URL 并复制它。

   ![Firecrawl 端点](/images/manual/use-cases/alex-firecrawl-endpoint.png#bordered)

在以下示例中，将示例端点替换为你自己的：

```text
https://717172b4.alexmiles.olares.com
```

### 配置 Open WebUI

要在 Open WebUI 中使用 Firecrawl，首先[将 Open WebUI 连接到模型](openwebui.md#配置模型后端)并配置网络搜索。[Open WebUI 网络搜索指南](openwebui-search.md#启用网络搜索-1)使用 SearXNG 作为示例。然后按以下步骤手动配置 Firecrawl 作为网页加载器。

:::tip GPU 资源
如果 Open WebUI 运行缓慢或无法返回结果，你的模型可能没有足够的 GPU 资源。停止不使用的但仍占用 GPU 资源的应用，然后重试。
:::

1. 打开 Open WebUI 应用。
2. 点击左下角的 **profile icon**，然后选择 **Admin Panel**。
3. 前往 **Settings** > **Web Search**。
4. 在加载器设置中，将 **Web Loader Engine** 设为 `firecrawl`。
5. 在 **Firecrawl API URL** 中，输入你从 Settings 复制的 Firecrawl 端点。
6. 对于 **Firecrawl API Key**，输入任意非空值，例如 `fc-test`。

   ![Firecrawl 加载器](/images/manual/use-cases/firecrawl-openwebui-loader-settings.png#bordered)

7. 点击 **Save** 保存设置。

确保 **Bypass Web Loader** 已禁用。

## 使用 API 测试 Firecrawl

本节是可选的。当你想确认 Firecrawl 是否可以直接抓取或爬取页面时使用。

### 了解 Bull Dashboard

从 Launchpad 打开 Firecrawl。你将看到一个 Bull Dashboard 页面，其中包含内部队列卡片，例如 `generateLlmsTxtQueue`、`deepResearchQueue`、`billingQueue` 和 `precrawlQueue`。队列名称可能因 Firecrawl 构建版本而异。

![Firecrawl 工作器状态](/images/manual/use-cases/firecrawl-worker-status.png#bordered)

此仪表板用于内部队列调试。正常的抓取或爬取请求可能不会出现在这里，或者完成得太快而注意不到。如果所有卡片都保持为 `0 Jobs`，请检查 API 响应或爬取状态 URL。

对于首次测试，你只需要两个 API 操作：

| 操作 | 端点 | 使用场景 |
|:-------|:---------|:------------------------|
| Scrape | `/v2/scrape` | 从单个特定页面提取内容。 |
| Crawl | `/v2/crawl` | 从一个 URL 开始，让 Firecrawl 从中发现页面。 |

以下示例使用浏览器控制台，以便请求可以使用你当前的 Olares 登录会话。

:::info Olares 端点认证
对于 Olares 托管的 Firecrawl，请使用已登录浏览器中的 `credentials: "include"`，而不是 `Authorization` 请求头。
:::

### 打开浏览器控制台

1. 从 Launchpad 打开 Firecrawl。
2. 在同一浏览器中，打开浏览器开发者工具。
3. 前往 **Console** 标签页。
4. 粘贴以下示例之一，然后按 **Enter**。

:::tip
如果你的浏览器阻止在控制台中粘贴代码，请按照浏览器提示允许粘贴。只粘贴你理解并信任的代码。
:::

### 抓取单个页面

当你想要单个页面的内容时，使用 scrape。

```javascript
const endpoint = "https://717172b4.alexmiles.olares.com";

const response = await fetch(`${endpoint}/v2/scrape`, {
  method: "POST",
  credentials: "include",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://docs.olares.com/manual/overview.html",
    formats: ["markdown"]
  })
});

const data = await response.json();
console.log(data);
```

如果请求成功，响应将包含 `data.markdown`。这是清理后的页面文本。响应还包含 `data.metadata`，例如页面标题、来源 URL、语言和 HTTP 状态码。

### 爬取页面或网站

当你希望 Firecrawl 从起始页面跟踪链接时，使用 crawl。开始时使用较小的 `limit`，以便结果易于检查。

```javascript
const endpoint = "https://717172b4.alexmiles.olares.com";

const response = await fetch(`${endpoint}/v2/crawl`, {
  method: "POST",
  credentials: "include",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://docs.olares.com/manual/overview.html",
    limit: 1
  })
});

const data = await response.json();
console.log(data);
```

成功的请求返回一个作业 ID 和一个状态 URL：

```json
{
  "success": true,
  "id": "019e6988-6e26-7407-b7a4-8045a8d12269",
  "url": "https://717172b4.alexmiles.olares.com/v2/crawl/019e6988-6e26-7407-b7a4-8045a8d12269"
}
```

`url` 字段是爬取状态 URL。在浏览器中打开它以检查爬取是否已完成。

### 读取爬取结果

| 字段 | 含义 |
|:------|:--------|
| `status` | 当前作业状态，例如 `scraping`、`completed` 或 `failed`。 |
| `data` | 爬取页面列表。每个项目通常包含页面内容和元数据。 |
| `markdown` | 清理后的页面内容。这是你通常发送给 AI 应用的主要文本。 |
| `metadata` | 页面信息，例如标题、来源 URL、语言和 HTTP 状态码。 |

:::tip 从 limit 开始
大型网站可能产生数百或数千个页面。测试时保持 `"limit": 1` 或 `"limit": 10`，确认结果有用后再增加。
:::

## 高级：生成摘要或结构化 JSON

Firecrawl 可以使用配置的 LLM 来总结页面或返回结构化 JSON。

:::warning JSON 和摘要输出需要 LLM 提供商
结构化 JSON 和摘要输出需要配置的 LLM 提供商。基于本地 Ollama 的 LLM 提取可能会失败，并显示 `Failed to parse URL from /responses`。如果发生这种情况，请使用常规 `markdown` 输出，或尝试 OpenAI 兼容的提供商。
:::

### 配置模型访问

1. 打开 **Settings**，然后导航至 **Applications** > **Firecrawl**。
2. 打开 **Environment variables**。
3. 配置你计划使用的模型提供商值：

   | 变量 | 描述 |
   |:---------|:------------|
   | `OPENAI_API_KEY` | OpenAI 兼容提供商的 API 密钥。 |
   | `OPENAI_BASE_URL` | OpenAI 兼容提供商的基础 URL。 |
   | `OLLAMA_BASE_URL` | 如果你使用 Ollama，则为 Ollama 端点 URL。前往 **Settings** > <br>**Applications** > **Ollama** > **Entrances** > **Ollama API** 并复制端点。 |
   | `MODEL_NAME` | 用于基于 LLM 提取的模型名称，例如 <br>`qwen3.5:35b-a3b-ud-q4_K_L`。 |

4. 点击 **Apply**。
5. 打开 Control Hub，在 **Browse** 下选择你的 Firecrawl 项目，然后重启 `worker`、`nuq-worker` 和 `firecrawl` 部署以应用环境变量。

### 返回结构化 JSON


```javascript
const endpoint = "https://717172b4.alexmiles.olares.com";

const response = await fetch(`${endpoint}/v2/scrape`, {
  method: "POST",
  credentials: "include",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://docs.olares.com/manual/overview.html",
    formats: [
      {
        type: "json",
        prompt: "Read the page carefully and answer: what is Olares in one sentence, and list 3 main features.",
        schema: {
          type: "object",
          properties: {
            one_liner: { type: "string" },
            features: { type: "array", items: { type: "string" } }
          },
          required: ["one_liner", "features"]
        }
      }
    ]
  })
});

const data = await response.json();
console.log(data.data.json);
```

| 响应字段 | 含义 |
|:---------------|:------------|
| `data.json` | 结构化 JSON 输出已成功生成。 |
| 仅 `data.metadata` | 页面已获取，但结构化 JSON 输出未生成。 |
| `error` | Firecrawl 返回错误。请阅读错误消息了解详情。 |

### 返回摘要

```javascript
const endpoint = "https://717172b4.alexmiles.olares.com";

const response = await fetch(`${endpoint}/v2/scrape`, {
  method: "POST",
  credentials: "include",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    url: "https://docs.olares.com/zh/manual/overview.html",
    formats: ["markdown", "summary"]
  })
});

const data = await response.json();
console.log(data.data.summary);
console.log("markdown length:", data.data.markdown?.length);
```

## FAQs

### 为什么 Bull Dashboard 显示 0 Jobs？

队列名称和可见作业可能因 Firecrawl 版本和 Olares 应用构建而异。请检查 API 响应或爬取状态 URL。

### 为什么爬取结果为空或不完整？

某些网站会阻止自动爬虫、需要登录，或通过复杂的浏览器交互加载内容。首先尝试较小的公共 URL，保持较低的爬取限制，并检查响应元数据中的错误。

## 了解更多

- [Firecrawl crawl API 参考](https://docs.firecrawl.dev/api-reference/endpoint/crawl-post)：爬取多个页面的请求和响应详情。
- [Firecrawl 文档](https://docs.firecrawl.dev/introduction)：SDK、LLM 集成和更多功能。
- [在 Open WebUI 中启用网络搜索](openwebui-search.md)：在 Open WebUI 中配置 SearXNG 和网络搜索。
