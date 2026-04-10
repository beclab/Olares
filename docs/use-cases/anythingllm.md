---
outline: [2, 3]
description: Set up AnythingLLM on Olares to build a local knowledge base with RAG. Upload documents, embed them with a local model, and query your knowledge using natural language.
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, RAG, knowledge base, local LLM, embedding, Ollama
app_version: "1.0.13"
doc_version: "1.0"
doc_updated: "2026-04-10"
---

# Build a local knowledge base with AnythingLLM

AnythingLLM is an open-source, all-in-one AI application that lets you chat with documents using Retrieval-Augmented Generation (RAG). It supports multiple LLM providers, embedding engines, and vector databases, all running locally on your Olares device.

## Learning objectives

In this guide, you will learn how to:
- Install a chat model and an embedding model from Market.
- Configure AnythingLLM to use these models via shared endpoints.
- Create a workspace and upload documents to build a knowledge base.
- Query your knowledge base using natural language.

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Install the model apps and AnythingLLM

You need three apps: a chat model for generating responses, an embedding model for processing documents, and AnythingLLM itself. This guide uses Cogito 14B as the chat model and Nomic Embed v1.5 as the embedding model.

You can also use other chat and embedding models available in Market, or install models through the [Ollama](ollama.md) app.

1. Open Market and search for "AnythingLLM".

   ![Install AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. Click **Get**, and then click **Install**.

3. Search for "Cogito 14B" and install it.

   ![Install Cogito 14B](/images/manual/use-cases/cogito-14b.png#bordered)

4. Search for "Nomic Embed v1.5" and install it.

   ![Install Nomic Embed v1.5](/images/manual/use-cases/nomic-embed.png#bordered)

5. Wait for all installations to finish.

## Download the models and get shared endpoints

After installation, each model app downloads its model automatically. You also need the shared endpoint URLs to connect AnythingLLM to these models.

### Get the chat model endpoint

1. Open the Cogito 14B app from Launchpad and wait for the model download to complete.
2. Open Settings, then navigate to **Application** > **Cogito 14B (Ollama)**.
3. In **Shared entrances**, select **Cogito 14B** to view the endpoint URL.

   <!-- ![Cogito 14B shared entrance](/images/manual/use-cases/anythingllm-cogito-shared-entrance.png#bordered) -->

4. Copy the shared endpoint URL. For example:
   ```plain
   http://d7837bc80.shared.olares.com
   ```

### Get the embedding model endpoint

1. Open the Nomic Embed v1.5 app from Launchpad and wait for the model download to complete.
2. Open Settings, then navigate to **Application** > **Nomic Embed v1.5**.
3. In **Shared entrances**, select **Nomic Embed v1.5** to view the endpoint URL.

   <!-- ![Nomic Embed v1.5 shared entrance](/images/manual/use-cases/anythingllm-nomic-shared-entrance.png#bordered) -->

4. Copy the shared endpoint URL. For example:
   ```plain
   http://8298761c0.shared.olares.com
   ```

## Configure AnythingLLM

By default, AnythingLLM connects to the Ollama app's shared endpoint for both the chat model and the embedder. Since you installed dedicated model apps, you need to update these endpoints to point to the correct models. These settings apply as the system default for all workspaces. You can also customize individual workspaces to use different models.

### Set up the chat model

1. Open AnythingLLM from Launchpad.
2. On the main page, click the tool icon in the bottom-left to open settings.
3. In the left sidebar, select **AI Providers** > **LLM**, and select **Ollama** as the LLM provider.
4. In the **Ollama Base URL** field, paste the shared endpoint URL for Cogito 14B.
5. Select the chat model from the **Ollama Model** dropdown.
   <!-- ![Configure chat model](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered) -->

### Set up the embedding model

1. In the left sidebar, select **Embedder**, and select **Ollama** as the embedding provider.
2. In the **Ollama Base URL** field, paste the shared endpoint URL for Nomic Embed v1.5.
3. Select the embedding model from the dropdown.

   <!-- ![Configure embedding model](/images/manual/use-cases/anythingllm-configure-embedding.png#bordered) -->

:::info Default embedding model
AnythingLLM includes a built-in embedding model (all-MiniLM-L6-v2) that works without additional setup and is primarily trained on English documents. If you prefer a zero-configuration option, you can use the default embedder instead.
:::

## Create a workspace

1. Click the back arrow to return to the AnythingLLM home screen.
2. Click **+** next to the search bar.
3. In the pop-up **New Workspace** window, name your workspace (for example, "Olares Docs"), and click **Save**.

   <!-- ![Create workspace](/images/manual/use-cases/anythingllm-create-workspace.png#bordered) -->

## Upload documents

1. Click the upload icon <span class="material-symbols-outlined">upload</span> next to the workspace name to open the document manager.

   <!-- ![Open document manager](/images/manual/use-cases/anythingllm-open-upload.png#bordered) -->

2. Upload your documentation files. AnythingLLM supports Markdown, PDF, TXT, DOCX, and other text-based formats.

   <!-- ![Upload documents](/images/manual/use-cases/anythingllm-upload-documents.png#bordered) -->

3. Select the uploaded files and click **Move to Workspace** to add them to the workspace.

   <!-- ![Move to workspace](/images/manual/use-cases/anythingllm-move-to-workspace.png#bordered) -->

4. Click **Save and Embed** to start embedding. This might take a few minutes depending on the number of documents.

   <!-- ![Embedding progress](/images/manual/use-cases/anythingllm-embedding-progress.png#bordered) -->

## Query your knowledge base

Once embedding is complete, you can ask questions about your documents.

1. Return to the workspace chat view.
2. Type your question in the chat input. For example:

   ```text
   How to enable LarePass VPN?
   ```

3. AnythingLLM retrieves relevant sections from your documents and generates an answer based on the content.

   <!-- ![Query result](/images/manual/use-cases/anythingllm-query-result.png#bordered) -->

## Learn more
- [Ollama](ollama.md): Set up Ollama on Olares for running local AI models.
- [Open WebUI](openwebui.md): Chat with local LLMs using a web-based interface.
- [LobeHub (LobeChat)](lobechat.md): Another AI chat interface for Olares.
