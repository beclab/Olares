---
outline: [2, 3]
description: Learn how to enable web search in OpenClaw using SearXNG to give your AI agent access to real-time internet information.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, OpenClaw web search
app_version: "1.0.1"
doc_version: "2.0"
doc_updated: "2026-05-27"
---

# Optional: Enable web search in OpenClaw

By default, OpenClaw answers questions using only its training data. It cannot access current events, real-time news, or live web content. To give your agent internet access, connect it to a web search provider.

This guide uses SearXNG, a privacy-focused meta-search engine that aggregates results from multiple sources without tracking users. You can install SearXNG as a self-hosted instance from the Olares Market.

## Learning objectives

In this guide, you will learn how to:
- Install SearXNG from the Olares Market.
- Obtain the shared endpoint URL for SearXNG.
- Configure OpenClaw to use SearXNG for web search and fetching.
- Verify that the web search tool is working.

## Step 1: Install SearXNG

Install SearXNG and obtain its shared endpoint URL.

1. Open Market, and search for "SearXNG".

   ![SearXNG](/images/manual/use-cases/perplexica-searxng.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.
3. Open Settings, and then go to **Applications** > **SearXNG**.
4. In **Shared entrances**, click **SearXNG**.

   ![Get SearXNG shared endpoint](/images/manual/use-cases/searxng-shared-laresprime.png#bordered){width=90%}

5. Copy the shared endpoint URL. For example:

   ```text
   http://d1236e020.shared.olares.com
   ```

## Step 2: Configure OpenClaw

Connect OpenClaw to SearXNG.

1. Open the OpenClaw CLI.
2. Run the following command to start the configuration wizard:

    ```bash
    openclaw configure --section web
    ```

3. Configure the settings as follows:

   | Settings | Option |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Enable web_search | Yes |
   | Search provider | SearXNG Search |
   | SearXNG Base URL | Paste the shared SearXNG endpoint URL you copied earlier. |
   | Enable web_fetch (keyless HTTP fetch) | Yes |

## Step 3: Verify web search

Test that your agent can retrieve real-time information from the internet.

1. Open the Control UI and start a chat with your agent.
2. Ask a question that requires current information.
3. Check the response. If the agent returns up-to-date information, the web search integration is working.

   :::tip Check tool calls
   Look for **Tool call** blocks in the chat to confirm the agent is using the `web_search` or `web_fetch` tools.
   :::

   ![Web search results using SearXNG](/images/manual/use-cases/openclaw-web-search-results.png#bordered)