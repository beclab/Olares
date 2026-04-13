---
outline: [2, 3]
description: Set up AnythingLLM on Olares to build a local knowledge base with RAG. Upload documents, embed them with a local model, and query your knowledge using natural language.
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, RAG, knowledge base, local LLM, embedding, Ollama
app_version: "1.0.13"
doc_version: "1.0"
doc_updated: "2026-04-13"
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

## Install AnythingLLM and model apps

You need three apps: AnythingLLM, a chat model for generating responses, and an embedding model for processing documents. 

This guide uses "Qwen3.5 9B" as the chat model and "Nomic Embed v1.5" as the embedding model.

1. Open Market and search for "AnythingLLM".

   ![Install AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. Click **Get**, and then click **Install**.

3. Search for "Qwen3.5 9B" and install it.

   ![Install Qwen3.5 9B](/images/manual/use-cases/qwen35-9b.png#bordered)

4. Search for "Nomic Embed v1.5" and install it.

   ![Install Nomic Embed v1.5](/images/manual/use-cases/nomic-embed.png#bordered)

5. Wait for all installations to finish.

## Download models and get shared endpoints

After installation, each model app downloads its model automatically. You must obtain the shared endpoint URL for each model to connect AnythingLLM to these models.

### Get the chat model endpoint

1. Open the Qwen3.5 9B Q4_K_M (Ollama) app from Launchpad and wait for the model download to complete.
2. Open Settings, and then go to **Applications** > **Qwen3.5 9B Q4_K_M (Ollama)**.
3. In **Shared entrances**, select **Qwen3.5 9B Q4_K_M** to view the endpoint URL.

   ![Qwen3.5 9B shared entrance](/images/manual/use-cases/anythingllm-qwen359b-shared-entrance.png#bordered){width=80%}

4. Copy the shared endpoint URL. For example:
   ```plain
   http://bd5355000.shared.olares.com
   ```
### Get the embedding model endpoint

1. Open the Nomic Embed v1.5 app from Launchpad and wait for the model download to complete.
2. Open Settings, and then go to **Applications** > **Nomic Embed v1.5**.
3. In **Shared entrances**, select **Nomic Embed v1.5** to view the endpoint URL.

   ![Nomic Embed v1.5 shared entrance](/images/manual/use-cases/anythingllm-nomic-shared-entrance.png#bordered){width=80%}

4. Copy the shared endpoint URL. For example:
   ```plain
   http://8298761c0.shared.olares.com
   ```

## Configure AnythingLLM

By default, AnythingLLM connects to the Ollama app's shared endpoint for both the chat model and the embedder. Because you installed dedicated model apps, you must update these endpoints to point to the correct models.

These settings apply as the system default for all workspaces. You can also customize individual workspaces to use different models.

### Set up the chat model

1. Open the AnythingLLM app from Launchpad.
2. On the home page, click the **Open settings** icon in the bottom-left.
3. In the left sidebar, select **AI Providers** > **LLM**, and then select **Ollama** as the LLM provider.
4. In the **Ollama Base URL** field, paste the shared endpoint URL for Qwen3.5 9B. **qwen3.5:9b** is automatically displayed in **Ollama Model**.
   
   ![Configure chat model](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered)

5. Click **Save changes**. The "LLM preferences saved successfully" message is displayed.

### Set up the embedding model

1. In the left sidebar, select **Embedder**, and then select **Ollama** as the embedding provider.
2. In the **Ollama Base URL** field, paste the shared endpoint URL for Nomic Embed v1.5. **nomic-embed-text:v1.5** is automatically displayed in **Ollama Embedding Model**.

   ![Configure embedding model](/images/manual/use-cases/anythingllm-configure-embedding.png#bordered)

3. Click **Save changes**. The "Embedding preferences saved successfully" message is displayed.
<!--
:::info Default embedding model
AnythingLLM includes a built-in embedding model (all-MiniLM-L6-v2) that works without additional setup and is primarily trained on English documents. If you prefer a zero-configuration option, you can use the default embedder instead.
:::
-->

## Create a workspace

1. Click **AnythingLLM** in the upper-left corner to return to the home page.
2. Click <span class="material-symbols-outlined">add_2</span> next to the search bar.

   ![Create workspace](/images/manual/use-cases/anythingllm-create-workspace.png#bordered)

3. In the **New Workspace** window, name your workspace such as `My test`, and then click **Save**.

## Upload documents

1. Click <span class="material-symbols-outlined">upload</span> next to the workspace name to open the document manager.

   ![Open document manager](/images/manual/use-cases/anythingllm-open-upload.png#bordered)

2. Upload your documents by uploading files or by submitting links. The uploaded documents and webpages are displayed in the **My Documents** panel.

   ![Upload documents](/images/manual/use-cases/anythingllm-upload-documents.png#bordered)

3. In the **My Documents** panel, select the uploaded documents, and then click **Move to Workspace** to add them to the newly created workspace.

   ![Move to workspace](/images/manual/use-cases/anythingllm-move-to-workspace.png#bordered)

4. Click **Save and Embed** to start embedding.

   This might take a few minutes depending on the number of documents. When the embedding finishes, the "Workspace updated successfully" message is displayed.

## Query your knowledge base

Ask questions about your documents.

1. Return to the workspace chat view.
2. Send your question through the chat. For example:

   ```text
   Olares supports backup or not
   ```

3. AnythingLLM retrieves relevant sections from your documents and generates an answer based on the content.

   ![Query result](/images/manual/use-cases/anythingllm-query-result.png#bordered)

## Learn more
- [Download and run local AI models via Ollama](ollama.md)
- [Official AnythingLLM documentation](https://docs.anythingllm.com/)
