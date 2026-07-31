---
outline: deep
title: Chat with documents in Open WebUI
description: Upload documents and create a knowledge base in Open WebUI on Olares for retrieval-augmented generation (RAG).
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, knowledge base, RAG, document upload, PDF
app_version: "1.0.38"
doc_version: "2.0"
doc_updated: "2026-07-31"
---

# Chat with documents and knowledge bases in Open WebUI

Open WebUI supports retrieval-augmented generation (RAG) to help local AI models answer questions based on your uploaded documents or curated knowledge bases.

This guide explains how to analyze individual documents during a chat session and how to build persistent knowledge collections for reuse.

## Learning objectives

In this guide, you will learn how to:

- Configure an embedding model to process document text.
- Upload and analyze individual documents in a chat session.
- Build and manage a persistent knowledge base.
- (Optional) Configure an advanced content extraction engine for complex document layouts.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one connected local model.
- Administrator privileges for the Open WebUI instance.
- The following model:
   | Model type | Model | How to get it |
   | :--- | :--- | :--- |
   | Embedding | EmbeddingGemma | Install from Market |

## Configure embedding model

Document understanding requires an embedding model to convert text into vector data. To configure Open WebUI, you must first retrieve your embedding model details.

### Get embedding model details

<!--@include: ../reusables/ai-service-connections.md#get-embedding-model-connection-details-openai-->

### Apply embedding settings in Open WebUI

<!--@include: ./openwebui-search.md{63,72}-->

## Analyze individual documents

Attach documents directly to a chat session for one-off analysis and summarization.

1. Start a new chat.
2. Select the model.
3. Click <i class="material-symbols-outlined">add_2</i> under the message input field, and then select **Upload Files**.

   ![Upload files in Open WebUI](/images/manual/use-cases/openwebui-upload-files1.png#bordered)

4. Upload a PDF or a text file.
5. Enter a prompt asking the model to analyze the document. For example:

   ```plain
   Summarize the main points of this document.
   ```

6. Submit the prompt. If the generated response includes file citations, Open WebUI successfully added the document to the context.

   ![File summary](/images/manual/use-cases/openwebui-file-summary1.png#bordered)

## Build a knowledge base

For documents you want to reuse across multiple chats, create a persistent knowledge base.

1. In Open WebUI, click your profile icon, and then go to **Workspace** > **Knowledge**.
2. Click **Create**.
3. In the **What are you working on** field, enter a name for your knowledge base. For example: `Product FAQs`.
4. In the **What are you trying to achieve** field, enter a description. For example: `Frequently asked questions and support guides for Olares products`.

   ![Create knowledge](/images/manual/use-cases/openwebui-create-knowledge1.png#bordered)

5. Click **Create Knowledge** to save the collection.
6. Click <i class="material-symbols-outlined">add</i> > **Upload files**, and then upload your files to populate the knowledge base.

   ![Populate knowledge base](/images/manual/use-cases/openwebui-populate-knowledge1.png#bordered)

## Attach a knowledge base to a chat

Once you have created a knowledge base, attach it to a chat so the model can reference its content.

1. Start a new chat.
2. Select the model.
3. Click <i class="material-symbols-outlined">add_2</i> under the message input field, and then select **Attach Knowledge**.
4. Choose the knowledge collection you want to use.

   ![Attach knowledge base to chat](/images/manual/use-cases/openwebui-attach-knowledge-base1.png#bordered)

5. Ask questions related to the knowledge base content. The model will retrieve relevant passages and cite them in its response.

## (Optional) Configure an advanced extraction engine

By default, Open WebUI uses a simple text extraction engine. For complex document layouts containing tables or complicated formatting, switch to PaddleOCR for better accuracy.

:::warning Performance impact
PaddleOCR requires more GPU VRAM and processes documents slower than the default engine. Use this engine only when document layout quality is critical.
:::

1. Install the PaddleOCR app from Market.

   ![PaddleOCR installation](/images/manual/use-cases/paddleocr.png#bordered)

2. Get the PaddleOCR endpoint URL:

   a. Open Olares Settings, and then go to **Applications** > **PaddleOCR** > **Entrances**.
   
   b. Copy the endpoint URL. For example, `https://17b4c78a.laresprime.olares.com`.

3. In Open WebUI, go to **Admin Panel** > **Settings**.
4. On the left sidebar, locate the **Tools** section, and then select **Documents**.
5. Under the **Content Extraction** section, configure as follows:

   a. **Content Extraction Engine**: Select **PaddleOCR-vl** in the drop-down list.

   b. **API Base URL**: Enter the PaddleOCR endpoint URL.

   c. **API Token**: Enter any text such as `local`. Do not leave this field empty.
   
   ![PaddleOCR config in Open WebUI](/images/manual/use-cases/openwebui-paddleocr-config1.png#bordered)

6. Click **Save**.
