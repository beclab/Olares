---
outline: deep
description: Enable web search in Open WebUI on Olares using SearXNG and an embedding model for retrieving up-to-date information.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, web search, SearXNG, embedding, RAG
app_version: "1.0.38"
doc_version: "2.0"
doc_updated: "2026-08-05"
---

# Enable web search in Open WebUI

Add web search capabilities to Open WebUI to allow your local AI models to retrieve up-to-date information from the internet. This integration requires a connected embedding model to generate embeddings and SearXNG to fetch web results.

If you want Open WebUI to read full web page content instead of using search result summaries only, configure a web loader such as Firecrawl.

## Learning objectives

In this guide, you will learn how to:

- Retrieve the required endpoint URLs for your embedding model and SearXNG.
- Configure the document embedding and web search settings in Open WebUI.
- Perform a web-assisted search during a chat session.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one connected local model.
- SearXNG installed.
- Administrator privileges for the Open WebUI instance.
- The following model:
   | Model type | Model | How to get it |
   | :--- | :--- | :--- |
   | Embedding | EmbeddingGemma | Install from Market |

## Retrieve service details

To link Open WebUI with your background services, you need to locate the connection endpoints for both your embedding model and SearXNG.

### Get embedding model details

<!--@include: ../reusables/ai-service-connections.md#get-embedding-model-connection-details-openai-->

### Get SearXNG endpoint

1. Open Olares Settings, and then go to **Applications** > **SearXNG**.
2. Under **Entrances**, click **SearXNG**, and then copy the endpoint URL. For example, `https://84a93c3c.laresprime.olares.com`.

   ![SearXNG endpoint](/images/manual/use-cases/openwebui-searxng-shared-endpoint1.png#bordered){width=70%}

## Configure Open WebUI

Apply the details you retrieved to the Open WebUI configuration panel.

### Set up document embeddings
<!--Note this section is reused in openwebui-knowledge, from line 63 to 72-->

Configure the embedding model so Open WebUI can convert text into vector representations for retrieval.

1. In Open WebUI, select your profile icon, and then go to **Admin Panel** > **Settings**.
2. On the left sidebar, locate the **Tools** section, and then select **Documents**.
3. Under the **Embedding** section, specify the following settings:

   - **Embedding Model Engine**: Select **OpenAI**.
   - **API Base URL**: Enter the embedding model **Base URL** you copied from the Model Console.
   - **Embedding Model**: Enter the embedding **Model name** you copied from the Model Console.

4. Scroll down to the bottom of the page, and then click **Reindex** in the lower-right corner to apply the changes.
5. Select **Save**.

### Enable web search

Turn on web search and point it to your SearXNG endpoint.

1. Go to **Admin Panel** > **Settings**.
2. On the left sidebar, locate the **Tools** section, and then select **Web Search**.
3. Specify the following settings:

   - **Web Search**: Enable this setting.
   - **Web Search Engine**: Select **SearXNG**.
   - **Searxng Query URL**: Enter your SearXNG endpoint URL and append `/search?q=<query>` to the end.

      For example, `https://84a93c3c.laresprime.olares.com/search?q=<query>`.
   - **Bypass Web Loader**: Enable this setting if you only need search result summaries. Leave it disabled if you want Open WebUI to fetch full page content through a web loader.

      :::tip Full-text retrieval
      For full-page retrieval, install Firecrawl and configure it as the web loader. See [Use Firecrawl as a web page loader](firecrawl.md#configure-open-webui).
      :::

   ![SearXNG configurations in Open WebUI](/images/manual/use-cases/openwebui-searxng-config1.png#bordered)

4. Leave the other fields at their default values.
5. Select **Save**.

## Verify the configuration

Test the feature to ensure the AI successfully retrieves up-to-date information from the web.

1. Start a new chat.
2. Select the model.
3. Click the **Integrations** icon under the chat input field, and then enable **Web Search**.

   ![Web search enable in Open WebUI chat](/images/manual/use-cases/openwebui-web-search-enable1.png#bordered)

4. Enter a prompt that requires recent information. For example:

   ```plain
   What’s the latest news about Olares One
   ```
5. Submit the prompt. The AI generates a response that includes the retrieved search results and their source links.

   ![Web search results in Open WebUI](/images/manual/use-cases/openwebui-web-search-results1.png#bordered)

## FAQ

### Why doesn't Open WebUI search the web?

If the AI response does not include web search results, work through the following checks:

1. **Search is enabled for the conversation.**  
   In the chat area, click the **Integrations** icon, and make sure **Web Search** is enabled.

2. **The embedding model is configured and reachable.**  
   Go to **Admin Panel** > **Settings** > **Tools** > **Documents**,  and verify the embedding model settings. If the embedding model is missing or unreachable, Open WebUI cannot process web search results.

3. **SearXNG returns results for your query.**  
   Open SearXNG from the Launchpad and run the same search. If SearXNG returns no results, check that its search engines are enabled and working in **Preferences** > **ENGINES**.

4. **The web loader is being blocked.**  
   If you need full-page content and the default web loader fails because of anti-scraping measures, go to **Admin Panel** > **Settings** > **Tools** > **Web Search**, and then enable **Bypass Web Loader**. This uses search result summaries instead of fetching full pages.

   :::tip
   For reliable full-page retrieval, install and configure Firecrawl as the web loader. See [Use Firecrawl as a web page loader](firecrawl.md#configure-open-webui).
   :::
