---
outline: [2, 3]
description: Build a private knowledge base with AnythingLLM on Olares. Add documents, create embeddings, and query them with RAG.
head:
  - - meta
    - name: keywords
      content: Olares, AnythingLLM, self-hosted rag, private knowledge base, anythingllm ollama, local LLM, embedding, anythingllm on olares
app_version: "1.0.13"
doc_version: "2.0"
doc_updated: "2026-07-27"
---

# Build a local knowledge base with AnythingLLM

AnythingLLM is an open-source, all-in-one AI application that lets you chat with documents using Retrieval-Augmented Generation (RAG). It supports multiple LLM providers and vector databases, all running locally on your Olares device.

## Learning objectives

In this guide, you will learn how to:
- Install AnythingLLM on Olares.
- Configure the chat model and embedding provider in AnythingLLM.
- Create a workspace and upload documents to build a knowledge base.
- Query your knowledge base using natural language.

## Prerequisites

Before you begin, you need:

- An Olares device with sufficient disk space and memory.
- Admin privileges to install apps from Market.
- The following models:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.6-27B (llama.cpp) | Install from Market |
  | Embedding | all-MiniLM-L6-v2 | The built-in embedder from AnythingLLM |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install AnythingLLM

1. Open Market and search for "AnythingLLM".

   ![Install AnythingLLM](/images/manual/use-cases/anythingllm.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Get model connection details

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For Qwen3.6-27B (llama.cpp):
<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

## Configure AnythingLLM

Configure the chat model and embedding provider. These settings become the system defaults for all workspaces.

### Set up the chat model

1. Open AnythingLLM from Launchpad.
2. On the home page, click the **Open settings** icon in the bottom-left.
3. In the left sidebar, select **AI Providers** > **LLM**, and then select **Generic OpenAI** as the LLM provider.
4. In **Base URL**, paste the Base URL from the Qwen3.6-27B (llama.cpp) Model Console.
5. In **Selected Model**, select `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.

   ![Configure chat model](/images/manual/use-cases/anythingllm-configure-chat-model.png#bordered)

6. Click **Save changes**. The "LLM preferences saved successfully" message is displayed.

### Set up the embedding model

1. In the left sidebar, select **Embedder**, and then select **AnythingLLM Embedder** as the embedding provider.
2. In **Model Preference**, select **all-MiniLM-L6-v2** as the embedding model.

   ![Configure embedding model](/images/manual/use-cases/anythingllm-configure-embedding-model.png#bordered)

3. Click **Save changes**. The "Embedding preferences saved successfully" message is displayed.

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
- [Host local models with Engine Base apps](llm-base-apps.md)
- [Official AnythingLLM documentation](https://docs.anythingllm.com/)
