---
outline: deep
description: Enable web search in Open WebUI on Olares using SearXNG and an embedding model for retrieving up-to-date information.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, web search, SearXNG, embedding, RAG
app_version: "1.0.20"
doc_version: "1.1"
doc_updated: "2026-06-02"
---

# Enable web search in Open WebUI

Add web search capabilities to Open WebUI to allow your local AI models to retrieve up-to-date information from the internet. This integration requires an active embedding model to process documents and SearXNG to fetch web results.

If you want Open WebUI to read full web page content instead of using search result summaries only, configure a web loader such as Firecrawl.

## Learning objectives

In this guide, you will learn how to:

- Retrieve the required endpoint URLs for your embedding model and SearXNG.
- Configure the document embedding and web search settings in Open WebUI.
- Perform a web-assisted search during a chat session.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one active model backend.
- SearXNG installed.
- An embedding model application installed, such as **Qwen3 Embedding 0.6B (Ollama)**.
- Administrator privileges for the Open WebUI instance.

## Retrieve service details

To link Open WebUI with your background services, you need to locate the connection endpoints for both your embedding model and SearXNG.

### Get embedding model details
<!--Note this section is reused in openwebui-knowledge, from line 38 to 46-->
1. Open Qwen3 Embedding 0.6B (Ollama) from the Launchpad.
2. Note down the exact model name displayed on the main page. For example, `qwen3-embedding:0.6b`.

   ![Qwen3 Embedding 0.6B](/images/manual/use-cases/qwen3-embedding.png#bordered)

3. Open Olares **Settings**, and then go to **Applications** > **Qwen3 Embedding 0.6B (Ollama)**.
4. Under **Shared entrances**, click **Qwen3 Embedding 0.6B**, and then copy the endpoint URL. For example, `http://eae5afcf0.shared.olares.com`.

### Get SearXNG endpoint

1. Open Olares Settings, and then go to **Applications** > **SearXNG**.
2. Under **Shared entrances**, click **SearXNG**, and then copy the endpoint URL. For example, `http://d1236e020.shared.olares.com`.

   ![SearXNG shared endpoint](/images/manual/use-cases/openwebui-searxng-shared-endpoint.png#bordered){width=70%}

## Configure Open WebUI

Apply the details you retrieved to the Open WebUI configuration panel.

### Set up document embeddings
<!--Note this section is reused in openwebui-knowledge, from line 61 to 69-->
1. In Open WebUI, select your profile icon, and then go to **Admin Panel** > **Settings** > **Documents**.
2. Under the **Embedding** section, specify the following settings:

   - **Embedding Model Engine**: Select **Ollama**.
   - **API Base URL**: Enter the embedding model endpoint URL you noted earlier.
   - **Embedding Model**: Enter the embedding model name you noted earlier.

3. Scroll down to the bottom of the page, and then click **Reindex** in the lower-right corner to apply the changes.
4. Select **Save**.

### Enable web search

1. Go to **Admin Panel** > **Settings** > **Web Search**.
2. Specify the following settings:

   - **Web Search**: Enable this setting.
   - **Web Search Engine**: Select **SearXNG**.
   - **Searxng Query URL**: Enter your SearXNG endpoint URL and append `/search?q=<query>` to the end.

      For example, `http://d1236e020.shared.olares.com/search?q=<query>`.
   - **Bypass Web Loader**: Enable this setting if you only need search result summaries. Leave it disabled if you want Open WebUI to fetch full page content through a web loader.

      :::tip Full-text retrieval
      For full-page retrieval, install Firecrawl and configure it as the web loader. See [Use Firecrawl as a web page loader](firecrawl.md#configure-open-webui).
      :::

   ![SearXNG configurations in Open WebUI](/images/manual/use-cases/openwebui-searxng-config.png#bordered)

3. Leave the other fields at their default values.
4. Select **Save**.

## Verify the configuration

Test the feature to ensure the AI successfully retrieves up-to-date information from the web.

1. Start a new chat.
2. Select the model.
3. Click the **Integrations** icon under the message input field, and then enable **Web Search**.

   ![Web search enable in Open WebUI chat](/images/manual/use-cases/openwebui-web-search-enable.png#bordered)

4. Enter a prompt that requires recent information. For example:

   ```plain
   What’s the latest news about Olares One
   ```
5. Submit the prompt. The AI generates a response that includes the retrieved search results and their source links.

   ![Web search results in Open WebUI](/images/manual/use-cases/openwebui-web-search-results.png#bordered)
