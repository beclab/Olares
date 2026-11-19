---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/openwebui-search
outline: deep
description: Enable web search in Open WebUI on Olares using SearXNG and an embedding model for retrieving up-to-date information.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, web search, SearXNG, embedding, RAG
app_version: "1.0.38"
doc_version: "2.1"
doc_updated: "2026-11-18"
---

# Enable web search in Open WebUI

<VersionRouteSelect />

Add web search capabilities to Open WebUI to allow your local AI models to retrieve up-to-date information from the internet. This integration requires a connected embedding model to generate embeddings and SearXNG to fetch web results.

If you want Open WebUI to read full web page content instead of using search result summaries only, configure a web loader such as Firecrawl.

## Learning objectives

In this guide, you will learn how to:

- Use Router for embeddings and get the SearXNG endpoint.
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

## Configure embedding model

<!--@include: ../reusables/openwebui-embeddings.md#configure-embeddings-->

## Connect SearXNG {#enable-web-search-1}

Open WebUI connects directly to SearXNG for search results. Use the SearXNG entrance URL here.

1. Open Olares Settings and go to **Applications** > **SearXNG** > **Entrances**.
2. Click **SearXNG** and copy the endpoint URL. For example:

   ```text
   https://84a93c3c.laresprime.olares.com
   ```

   ![SearXNG endpoint](/images/manual/use-cases/openwebui-searxng-shared-endpoint1.png#bordered){width=70%}

3. In Open WebUI, open **Admin Panel** > **Settings**.
4. On the left sidebar, find **Tools** and select **Web Search**.
5. Set the following fields:

   - **Web Search**: Enable.
   - **Web Search Engine**: Select **SearXNG**.
   - **Searxng Query URL**: Append `/search?q=<query>` to the SearXNG endpoint you copied. For example:

     ```text
     https://84a93c3c.laresprime.olares.com/search?q=<query>
     ```

   - **Bypass Web Loader**: Enable for search result summaries only. Disable to read full web page content.

     To read full pages, first [configure Firecrawl as the web loader](firecrawl.md#configure-open-webui).

   ![SearXNG settings in Open WebUI](/images/manual/use-cases/openwebui-searxng-config1.png#bordered)

6. Leave the other fields at their default values and click **Save**.

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

## Learn more

- [Open WebUI common issues](openwebui-common-issues.md): Resolve common issues with Open WebUI.
