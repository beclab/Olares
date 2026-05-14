---
outline: deep
description: Upload documents and create a knowledge base in Open WebUI on Olares for retrieval-augmented generation (RAG).
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, knowledge base, RAG, document upload, PDF
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-14"
---

# Chat with documents and knowledge bases

Open WebUI supports retrieval-augmented generation (RAG) to help local AI models answer questions based on your uploaded documents or curated knowledge bases.

This guide explains how to analyze individual documents during a chat session and how to build persistent knowledge collections for repeated use.

## Learning objectives

In this guide, you will learn how to:

- Configure an embedding model to process document text.
- Upload and analyze individual documents in a chat session.
- Build and manage a persistent knowledge base.
- Configure an advanced content extraction engine for complex document layouts.

## Prerequisites

Before you begin, ensure you have the following in place:

- [Open WebUI](openwebui.md) installed and configured with at least one active model backend.
- An embedding model application (for example, Qwen3 Embedding) installed from Market.
- Administrator privileges for the Open WebUI instance.

## Configure embedding model

Document understanding requires an embedding model to convert text into vector data. To configure Open WebUI, you must first retrieve your embedding model details.

<!--@include: ./openwebui-search.md{38,46}-->

### Apply embedding settings in Open WebUI

<!--@include: ./openwebui-search.md{61,69}-->

## Analyze individual documents

Attach documents directly to a chat session for one-off analysis and summarization.

1. Start a new chat in Open WebUI.
2. Select the model.
3. Click <i class="material-symbols-outlined">add_2</i> in the message input area, and then select **Upload Files**.

   ![Upload files in Open WebUI](/images/manual/use-cases/openwebui-upload-files.png#bordered) -->

3. Upload a PDF or a text file.
4. Enter a prompt asking the model to analyze the document. For example:

   ```plain
   Summarize the main points of this document.
   ```

5. Submit the prompt. If the generated response includes file citations, Open WebUI successfully added the document to the context.

   ![File summary](/images/manual/use-cases/openwebui-file-summary.png#bordered)

## Build a knowledge base

For documents you want to reuse across multiple chats, create a persistent knowledge base.

1. In Open WebUI, click your profile icon, and then go to **Workspace** > **Knowledge**.
2. Click **New Knowledge**.
3. In the **What are you working on?** field, enter a name for your knowledge base. For example: `Product FAQs`.
4. In the **What are you trying to achieve?** field, enter an optional description. For example: `Frequently asked questions and support guides for Olares products`.

   ![Create knowledge](/images/manual/use-cases/openwebui-create-knowledge.png#bordered)

5. Click **Create Knowledge** to save the collection.
6. Click <i class="material-symbols-outlined">add</i> > **Upload files**, and then upload your files to populate the knowledge base.

   ![Populate knowledge base](/images/manual/use-cases/openwebui-populate-knowledge.png#bordered)

## Attach a knowledge base to a chat

1. Start a new chat.
2. Select the model.
3. Click <i class="material-symbols-outlined">add_2</i> in the message input area, and then select **Attach Knowledge**.
4. Choose the knowledge collection you want to use.

   ![Attach knowledge base to chat](/images/manual/use-cases/openwebui-attach-knowledge-base.png#bordered)

5. Ask questions related to the knowledge base content. The model will retrieve relevant passages and cite them in its response.

   ![Search results from attached knowledge base](/images/manual/use-cases/openwebui-search-results-from-knowledge-base.png#bordered)

## (Optional) Configure an advanced extraction engine

By default, Open WebUI uses a simple text extraction engine. For complex document layouts containing tables or intricate formatting, switch to PaddleOCR for better accuracy.

:::warning Performance impact
PaddleOCR requires more GPU VRAM and processes documents slower than the default engine. Use this engine only when document layout quality is critical.
:::

1. Install the PaddleOCR app from Market.

   ![PaddleOCR installation](/images/manual/use-cases/paddleocr.png#bordered)

2. Get the PaddleOCR endpoint URL:

   a. Open Olares Settings, and then go to **Applications** > **PaddleOCR** > **Shared entrances** > **PaddleOCR API**.
   
   b. Copy the endpoint URL. For example, `http://6b2a6fc50.shared.olares.com`.

3. In Open WebUI, go to **Admin Panel** > **Settings** > **Documents**.
4. In the **General** section, select **PaddleOCR-vl** for **Content Extraction Engine**.
5. In **API Base URL**, enter the PaddleOCR endpoint URL. 
6. In **API Token**, enter any text. Do not leave this field empty.
   
   ![PaddleOCR config in Open WebUI](/images/manual/use-cases/openwebui-paddleocr-config.png#bordered)

7. Click **Save**.
