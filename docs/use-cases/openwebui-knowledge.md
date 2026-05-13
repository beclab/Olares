---
outline: deep
description: Upload documents and create a knowledge base in Open WebUI on Olares for retrieval-augmented generation (RAG).
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, knowledge base, RAG, document upload, PDF
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-05-13"
---

# Chat with documents and knowledge bases

Open WebUI supports retrieval-augmented generation (RAG), which lets the model answer questions based on uploaded documents or a curated knowledge base. This guide covers both one-off document uploads and persistent knowledge collections.

## Prerequisites

- [Open WebUI and a model backend installed](openwebui.md) on Olares
- An embedding model service installed, such as qwen3-embedding
- Admin privileges on Open WebUI

## Configure embedding model

Document understanding requires an embedding model to convert text into vectors.

1. In Open WebUI, click your **profile icon** and select **Admin Panel**.
2. Navigate to **Settings** > **Documents**.
3. Set **Embedding Model Engine** to **Ollama**.
   <!-- ![Embedding settings](/images/manual/use-cases/openwebui/embedding-settings-kb.png#bordered) -->
4. Get the embedding service endpoint:

   a. Open Olares Settings, then navigate to **Applications** > **[Embedding App]**.

   b. In **Shared entrances**, copy the endpoint URL.

   c. Open the embedding app from Launchpad and note the model name shown on the main page.

5. Return to Open WebUI and fill in the fields:
   - **Ollama Base URL**: Paste the embedding endpoint URL.
   - **Embedding Model**: Enter the model name you noted.
6. Click **Save**.

## Upload documents in chat

You can attach documents directly to a chat session for one-off analysis.

1. Start a new chat in Open WebUI.
2. Click the attachment icon and upload a PDF or text file.
   <!-- ![PDF upload](/images/manual/use-cases/openwebui/pdf-upload.png#bordered) -->
3. Ask the model to summarize or analyze the document. For example:
   ```plain
   Summarize the main points of this document.
   ```
4. If the response includes file citations, the document has been successfully added to the context.
   <!-- ![PDF summary](/images/manual/use-cases/openwebui/pdf-summary.png#bordered) -->

## Create a knowledge base

For documents you want to reuse across multiple chats, create a knowledge base.

1. In Open WebUI, navigate to **Workspace** > **Knowledge**.
2. Click **Create Knowledge** and give it a name.
   <!-- ![Create knowledge](/images/manual/use-cases/openwebui/create-knowledge.png#bordered) -->
3. Upload files to populate the knowledge base.

## Attach knowledge in chat

1. Start a new chat.
2. Click the attachment icon and select the **Knowledge** tab.
3. Choose the knowledge collection you want to use.
   <!-- ![Attach knowledge left](/images/manual/use-cases/openwebui/attach-knowledge-left.png#bordered) -->
4. Ask questions related to the knowledge base content. The model will retrieve relevant passages and cite them in its response.
   <!-- ![Attach knowledge right](/images/manual/use-cases/openwebui/attach-knowledge-right.png#bordered) -->

## Switch content extraction engine

By default, Open WebUI uses a simple text extraction engine. For complex document layouts, you can switch to PaddleOCR-vl for better accuracy.

:::warning Performance impact
PaddleOCR-vl is slower and requires more GPU VRAM than the default engine. Use it only when document layout quality is critical.
:::

1. Install the PaddleOCR app from Market.
2. Get the PaddleOCR endpoint URL:

   a. Open Olares Settings, then navigate to **Applications** > **PaddleOCR**.

   b. In **Shared entrances**, copy the endpoint URL.

3. In Open WebUI Admin Panel, navigate to **Settings** > **Documents**.
4. Under **Content Extraction Engine**, select **PaddleOCR-vl**.
5. Enter the PaddleOCR API URL. The API key field can be filled with any placeholder value.
   <!-- ![PaddleOCR config](/images/manual/use-cases/openwebui/paddleocr-config.png#bordered) -->
6. Click **Save**.
