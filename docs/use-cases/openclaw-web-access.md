---
outline: [2, 3]
description: Learn how to enable web search in OpenClaw using SearXNG to give your AI agent access to real-time internet information.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw web search
app_version: "1.0.36"
doc_version: "2.4"
doc_updated: "2026-09-04"
---

# Optional: Enable web search in OpenClaw

By default, OpenClaw answers questions using only its training data. It cannot access current events, real-time news, or live web content. To give your agent internet access, connect it to a web search provider.

This guide uses SearXNG, a privacy-focused meta-search engine that aggregates results from multiple sources without tracking users. You can install SearXNG as a self-hosted instance from the Olares Market.

## Learning objectives

In this guide, you will learn how to:
- Install SearXNG from the Olares Market.
- Get the SearXNG endpoint.
- Configure OpenClaw to use SearXNG for web search and fetching.
- Verify that the web search tool is working.

## Step 1: Install SearXNG

Install SearXNG from Market.

1. Open Market, and search for "SearXNG".

   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Step 2: Get the SearXNG endpoint

OpenClaw needs the SearXNG endpoint to connect to its search service.

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

1. Go to Olares **Settings** > **Applications** > **SearXNG** > **Entrances**.
2. Select **SearXNG**, then copy the **Endpoint** URL.

## Step 3: Connect OpenClaw to SearXNG

Connect OpenClaw to SearXNG.

1. Open the OpenClaw CLI.
2. Run the following command to start the configuration wizard:

   ```bash
   openclaw configure --section web
   ```

3. Configure the settings as follows:

   | Setting | Option |
   |:---------|:-------|
   | <nobr>Enable the web_search tool</nobr> | Yes |
   | <nobr>Search provider</nobr> | SearXNG Search |
   | <nobr>SearXNG Base URL</nobr> | Paste the SearXNG Endpoint URL copied in Step 2. |
   | <nobr>Install SearXNG plugin</nobr> | Download from npm (@openclaw/searxng-plugin) |
   | Enable the web_fetch tool (keyless HTTP fetch) | Yes |

## Step 4: Verify web search

Test that your agent can retrieve real-time information from the internet.

1. Open the Control UI and start a chat with your agent.
2. Ask a question that requires current information. For example:

   ```text
   Tell me the latest top 3 news on this web page
   https://news.ycombinator.com/newest using your web search tool
   ```
3. Check the response. If the agent returns up-to-date information, the web search integration is working. 

:::tip Full-text retrieval
SearXNG returns only titles, URLs, and snippets, not full page content. Fetching the full text might be blocked by anti-scraping measures. If you need the agent to read the full contents of web pages, use an online web service. We recommend Firecrawl and Tavily. They return full text or answer snippets and offer free quotas for web search.
:::
