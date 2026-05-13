---
outline: deep
description: Enable web search in Open WebUI on Olares using SearXNG and an embedding model for retrieving up-to-date information.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, web search, SearXNG, embedding, RAG
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

# Enable web search

You can equip Open WebUI with web search capabilities so your local LLM can retrieve current information from the internet. This requires an embedding model and the SearXNG search engine.

## Prerequisites

- [Open WebUI and a model backend installed](openwebui.md) on Olares
- An embedding model service installed, such as qwen3-embedding
- Admin privileges on Open WebUI

## Configure embedding model

1. In Open WebUI, click your **profile icon** and select **Admin Panel**.
2. Navigate to **Settings** > **Documents**.
3. Set **Embedding Model Engine** to **Ollama**. Do not use the default SentenceTransformers engine.
   <!-- ![Embedding settings](/images/manual/use-cases/openwebui/embedding-settings.png#bordered) -->
4. Get the embedding service endpoint:

   a. Open Olares Settings, then navigate to **Applications** > **[Embedding App]**.

   b. In **Shared entrances**, copy the endpoint URL.

   c. Open the embedding app from Launchpad and note the model name shown on the main page.

5. Return to Open WebUI and fill in the fields:
   - **Ollama Base URL**: Paste the embedding endpoint URL.
   - **Embedding Model**: Enter the model name you noted.
6. Click **Save**.

## Install SearXNG

1. Open Market and search for "SearXNG".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open Olares Settings and navigate to **Applications** > **SearXNG**.
4. In **Shared entrances**, copy the endpoint URL. For example:
   ```plain
   http://d1236e020.shared.olares.com
   ```

## Configure web search

1. In the Open WebUI Admin Panel, navigate to **Settings** > **Web Search**.
2. Turn on the toggle in the top-right corner.
3. Set the following fields:
   - **Search Engine**: Select **SearXNG**.
   - **Searxng Query URL**: Enter your SearXNG endpoint URL followed by `/search?q=<query>`. For example:
     ```plain
     http://d1236e020.shared.olares.com/search?q=<query>
     ```
4. Leave other fields at their defaults.
5. Click **Save**.

## Verify the configuration

1. Start a new chat in Open WebUI.
2. Enable the **Web Search** toggle near the message input field.
   <!-- ![Web search toggle](/images/manual/use-cases/openwebui/web-search-toggle.png#bordered) -->
3. Ask a question that requires recent information. For example:
   ```plain
   Search the latest news about Olares One
   ```
4. The response should include search results with source links.
   <!-- ![Search results](/images/manual/use-cases/openwebui/search-results.png#bordered) -->
