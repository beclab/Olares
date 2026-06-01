---
outline: [2, 3]
description: Set up Firecrawl on Olares to scrape pages, crawl websites, return clean Markdown or structured JSON, and load web pages for Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, Firecrawl, web crawler, web scraping, Firecrawl v2, scrape API, crawl API, Open WebUI, web loader, self-hosted
app_version: "1.0.21"
doc_version: "1.0"
doc_updated: "2026-06-01"
---

# Crawl and scrape websites with Firecrawl

Firecrawl turns web pages into clean Markdown, structured JSON, summaries, and metadata. On Olares, you can test Firecrawl from your browser console, then connect it to apps such as Open WebUI to give AI chats access to web page content.

## Install Firecrawl

1. Open Market and search for "Firecrawl".

   ![Firecrawl](/images/manual/use-cases/firecrawl.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Understand the Bull Dashboard

Open Firecrawl from Launchpad. You will see a Bull Dashboard page with internal queue cards, such as `generateLlmsTxtQueue`, `deepResearchQueue`, `billingQueue`, and `precrawlQueue`. Queue names can differ between Firecrawl builds.

![Firecrawl worker status](/images/manual/use-cases/firecrawl-worker-status.png#bordered)

This dashboard is for internal queue debugging. A normal scrape or crawl request might not appear here, or it might finish too quickly to notice. If all cards stay at `0 Jobs`, check the API response or crawl status URL instead.

## Get the Firecrawl endpoint

The endpoint is the base address of your Firecrawl service on Olares. You will add API paths such as `/v2/scrape` or `/v2/crawl` to this address.

1. Open **Settings**, then navigate to **Applications** > **Firecrawl**.
2. Under **Entrances**, click **Firecrawl**.
3. Under **Endpoint settings**, locate the endpoint URL and copy it.

   ![Firecrawl endpoint](/images/manual/use-cases/alex-firecrawl-endpoint.png#bordered)

In the examples below, replace the sample endpoint with your own:

```text
https://717172b4.alexmiles.olares.com
```

## Use the Firecrawl API

For a first test, you only need two API actions:

| Action | Endpoint | Use it when you want to |
|:-------|:---------|:------------------------|
| Scrape | `/v2/scrape` | Extract content from one specific page. |
| Crawl | `/v2/crawl` | Start from one URL and let Firecrawl discover pages from there. |

The examples below use the browser console so the requests can use your current Olares sign-in session.

:::info Olares endpoint authentication
For Olares-hosted Firecrawl, use `credentials: "include"` from a signed-in browser instead of an `Authorization` header.
:::

### Open the browser console

1. Open Firecrawl from Launchpad.
2. In the same browser, open your browser developer tools.
3. Go to the **Console** tab.
4. Paste one of the examples below and press **Enter**.

:::tip
If your browser blocks pasted code in the console, follow the browser prompt to allow pasting. Only paste code you understand and trust.
:::

### Scrape a single page

Use scrape when you want the content of one page.

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

If the request succeeds, the response includes `data.markdown`. This is the cleaned page text. The response also includes `data.metadata`, such as the page title, source URL, language, and HTTP status code.

### Crawl a page or website

Use crawl when you want Firecrawl to follow links from the starting page. Start with a small `limit` so the result is easy to inspect.

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

A successful request returns a job ID and a status URL:

```json
{
  "success": true,
  "id": "019e6988-6e26-7407-b7a4-8045a8d12269",
  "url": "https://717172b4.alexmiles.olares.com/v2/crawl/019e6988-6e26-7407-b7a4-8045a8d12269"
}
```

The `url` field is the crawl status URL. Open it in your browser to check whether the crawl has finished.

### Read crawl results

| Field | Meaning |
|:------|:--------|
| `status` | Current job state, such as `scraping`, `completed`, or `failed`. |
| `data` | List of crawled pages. Each item usually includes page content and metadata. |
| `markdown` | Cleaned page content. This is the main text you usually send to an AI app. |
| `metadata` | Page information such as title, source URL, language, and HTTP status code. |

:::tip Start with a limit
Large websites can produce hundreds or thousands of pages. Keep `"limit": 1` or `"limit": 10` while testing, then increase it after you confirm the result looks useful.
:::

## Advanced: Generate summaries or structured JSON

Firecrawl can use a configured LLM to summarize a page or return structured JSON.

:::warning LLM provider required for JSON and summary output
Structured JSON and summary output require a configured LLM provider. Local Ollama-based LLM extraction may fail with `Failed to parse URL from /responses`. If this happens, use regular `markdown` output or try an OpenAI-compatible provider.
:::

### Configure model access

1. Open **Settings**, then navigate to **Applications** > **Firecrawl**.
2. Open **Environment variables**.
3. Configure the model provider values you plan to use:

   | Variable | Description |
   |:---------|:------------|
   | `OPENAI_API_KEY` | API key for an OpenAI-compatible provider. |
   | `OPENAI_BASE_URL` | Base URL for an OpenAI-compatible provider. |
   | `OLLAMA_BASE_URL` | Ollama endpoint URL, if you use Ollama. Go to **Settings** > <br>**Applications** > **Ollama** > **Entrances** > **Ollama API** and copy<br> the endpoint. |
   | `MODEL_NAME` | Model name used for LLM-based extraction, such as <br>`qwen3.5:35b-a3b-ud-q4_K_L`. |

4. Click **Apply**.
5. Open Control Hub, select your Firecrawl project under **Browse**, then restart the `worker`, `nuq-worker`, and `firecrawl` deployments to apply the environment variables.

### Return structured JSON


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

| Response field | Meaning |
|:---------------|:--------|
| `data.json` | Structured JSON output was generated successfully. |
| `data.metadata` only | The page was fetched, but structured JSON output was not generated. |
| `error` | Firecrawl returned an error. Read the error message for details. |

### Return a summary

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

## Use Firecrawl with Open WebUI

Open WebUI uses SearXNG as the search engine and Firecrawl as the web page loader. Firecrawl does not replace SearXNG.

Before starting, make sure Open WebUI is installed and connected to a model. See [Set up Open WebUI for local AI chat](openwebui.md). To enable web search with SearXNG, see [Enable web search in Open WebUI](openwebui-search.md).

:::tip GPU resources
If Open WebUI is slow or cannot return a result, your model might not have enough GPU resources. Stop apps that are not in use but still occupy GPU resources, then try again.
:::

1. Open the Open WebUI app.
2. Click your **profile icon** in the bottom-left corner and select **Admin Panel**.
3. Go to **Settings** > **Web Search**.
4. In the loader settings, set **Web Loader Engine** to `firecrawl`.
5. In **Firecrawl API URL**, enter the Firecrawl endpoint you copied from Settings.
6. For **Firecrawl API Key**, enter any non-empty value, such as `fc-test`.

   ![Firecrawl loader](/images/manual/use-cases/firecrawl-openwebui-loader-settings.png#bordered)

7. Click **Save** to save the settings.

Make sure **Bypass Web Loader** is disabled.

## FAQs

### Why does the Bull Dashboard show 0 Jobs?

Queue names and visible jobs can differ by Firecrawl version and Olares app build. Check the API response or crawl status URL instead.

### Why is the crawl result empty or incomplete?

Some websites block automated crawlers, require login, or load content through complex browser interactions. Try a smaller public URL first, keep the crawl limit low, and check the response metadata for errors.

## Learn more

- [Firecrawl crawl API reference](https://docs.firecrawl.dev/api-reference/endpoint/crawl-post): Request and response details for crawling multiple pages.
- [Firecrawl documentation](https://docs.firecrawl.dev/introduction): SDKs, LLM integrations, and more features.
- [Enable web search in Open WebUI](openwebui-search.md): Configure SearXNG and web search in Open WebUI.
