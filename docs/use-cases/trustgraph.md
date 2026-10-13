---
outline: [2, 3]
description: Turn documents into searchable knowledge with TrustGraph on Olares. Configure Router models, process files, ask questions with source context, and explore entity relationships.
head:
  - - meta
    - name: keywords
      content: Olares, TrustGraph, knowledge graph, Document RAG, Graph RAG, local models, document search
app_version: "0.1.21"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

# Turn documents into searchable knowledge with TrustGraph

TrustGraph is an open-source semantic intelligence platform that turns source data into traceable knowledge. It combines document retrieval with a knowledge graph, so you can ask questions about your files and inspect the entities and relationships extracted from them.

On Olares, TrustGraph stores documents in Olares-managed storage and calls models through Router. This example uses local models to process a short laboratory description, answer a question, and explore its graph.

## Prerequisites

- Olares 1.12.7 or later on an amd64 or arm64 device. TrustGraph requires 10 GiB of application memory and 20 GiB of disk space, in addition to the resources needed by your models.
- Access to [Olares Router](olares-router.md), with a chat model, EmbeddingGemma, and BGE Reranker v2 M3 installed and callable.

## Check model availability

TrustGraph uses three model capabilities. The chat model extracts knowledge and writes answers. The embedding model finds relevant content. The reranker orders retrieved results before an answer is generated.

1. Open the installed model apps from Launchpad and wait for their downloads and startup to finish.
2. Open Router from Launchpad. On **Default models**, select your chat model for `default-chat` and BGE Reranker v2 M3 for `default-rerank`.
3. Check that these models are callable in Router:

   | Capability | Model used by the default TrustGraph flow |
   | :--- | :--- |
   | Chat | `default-chat`, routed to your selected chat model |
   | Embedding | `Olares/embeddinggemma-300m`, producing 768-dimensional vectors |
   | Reranking | `default-rerank`, routed to BGE Reranker v2 M3 |

   The embedding model is selected explicitly by TrustGraph. Changing Router's default embedding model does not change this flow.

## Install TrustGraph

1. Open Market and search for "TrustGraph".
2. Click **Get**, then **Install**. When prompted, set `ADMIN_PASSWORD` to a password of 12–128 characters without line breaks, and wait for installation to complete.

## Sign in

1. Open TrustGraph from Launchpad.
2. Sign in with username `admin` and the password you chose during installation.

The administrator password is separate from your Olares password and Router API key. You do not need a TrustGraph API key to sign in through this form. Restarts and upgrades preserve the existing account password; the installation field does not reset it.

## Use TrustGraph

### Upload and process a document

A document must be submitted for processing after it is uploaded. A flow defines how it is processed, and a collection groups the resulting searchable knowledge.

1. Save the following text as `orion-laboratory.txt` on your computer:

   ```text
   Orion laboratory is located in Suzhou. It is directed by Dr. Chen and studies hydrogen fuel cells.
   ```

2. On **Welcome**, open **Document Ingestion**. Click **+ Add Document**, select **File**, and choose the text file.
3. Set **TITLE** to `Orion laboratory`, then click **Upload**.
4. Select the uploaded document and click **Submit for Processing →**.
5. Select the `default` flow. Choose **+ Create New Collection**, enter `orion-notes` in **ID** and `Orion notes` in **NAME**, then click **Next →**.
6. Check the document, flow, and collection in the confirmation panel, then click **Submit for Processing**.
7. Wait for processing to finish before asking questions. Set the top **COL** selector to `orion-notes` and **FLOW** to `default` for the following steps.

The ingestion view shows how documents connect to flows and storage. An uploaded file or submission count alone does not confirm that its knowledge is ready to query.

:::tip Start with a short text file
You can upload TXT, PDF, and DOCX documents. The first PDF or DOCX request can take several minutes while the document decoder downloads its parsing models. Uploading and processing are separate operations, so wait for processing before submitting the same file again.
:::

### Ask questions about your documents

1. Open **Workflows** and select **Document RAG Query**.
2. Confirm that **COL** is `orion-notes` and **FLOW** is `default`.
3. Keep **Timeline** selected, enter the following question, and click **Query**:

   ```text
   Where is Orion laboratory located, and who directs it?
   ```

4. Check that the answer identifies Suzhou and Dr. Chen. Inspect the timeline and source links to compare the retrieved text with your original document.

   ![Document RAG answer and processing timeline in TrustGraph](/images/manual/use-cases/trustgraph-query-document.png#bordered)

   Your collection name and answer wording can differ from the screenshot.

**Simple** shows the answer without the explanation view. **Timeline** shows the retrieval and answer steps. **Full DAG** presents those steps as a graph you can inspect. Source context helps you check an answer; it does not guarantee that every answer is correct.

### Explore relationships in the knowledge graph

1. Open **Workflows** and select **Graph Navigator**. Keep the same collection selected.
2. Select **Full Explorer**, search for `Orion`, and choose a matching entity.
3. Select nodes to inspect their properties and navigate connected entities. Look for the laboratory's location and director relationships.

   ![TrustGraph graph navigation with entities and relationships](/images/manual/use-cases/trustgraph-explore-graph.png#bordered)

   The entities shown depend on the documents in your collection.

4. To ask a question using graph context, return to **Workflows**, select **Graph RAG Query**, and submit the same question with the same collection and flow.

Document RAG retrieves relevant document text. Graph RAG retrieves graph context built from extracted knowledge. Relationship labels and extraction results depend on your chat model and source content.

## Change the model connection

To change the chat or reranking model for every app using these Router defaults, update `default-chat` or `default-rerank` on Router's **Default models** page. Keep the replacement models callable.

When you need to change the model API URL or API key, open **Settings** > **Applications** > **TrustGraph** > **Manage environment variables**. Update `LLM_BASE_URL` or `LLM_API_KEY`, then click **Confirm** and **Apply**.

Keep a collection's embedding model consistent. The default flow uses 768-dimensional EmbeddingGemma vectors. To experiment with the supported 1024-dimensional Jina CLIP v2 model, create a separate flow and a new collection, then process the documents again. Do not mix these vector dimensions in one collection.

## FAQs

### Why does a document upload successfully but have no answers yet?

Check that you submitted it for processing, that the chat and embedding models are callable, and that the query uses the same collection and flow. For the first PDF or DOCX, allow time for decoder model downloads. Begin with the short TXT example to distinguish parsing delays from model connection problems.

### Why does Graph RAG report that no information is available?

Graph RAG needs successfully extracted knowledge in the selected collection. Uploading a file and submitting it do not by themselves confirm that extraction succeeded.

1. Check that the query uses the collection and flow selected during processing.
2. Check that the chat and embedding models are callable. In Control Hub, inspect the `text-completion` and `triples` logs for extraction failures. Repeated output at the token limit, JSONL parsing failures, or an unterminated string can indicate an incomplete model response.
3. After correcting the model or application configuration, submit the affected documents for processing again. Configuration changes do not automatically recreate missing knowledge.

With reasoning models such as Qwen, a high default reasoning effort can leave less output capacity for extracted facts. Choose a model configuration suitable for structured extraction. In this package, extraction explicitly requests low reasoning effort and allows up to 8192 output tokens. This is a ceiling, not a guarantee that every document will fit; model services can also impose their own limits.

:::details Administrator note: extraction settings
The TrustGraph OpenAI processor reads `max_output` in its launch configuration. `max_output_tokens` is not the launch configuration field for this version. Under the `openai` variant, `thinking: off` omits the reasoning parameter; it does not explicitly disable the model's reasoning. `thinking: low` sends `reasoning_effort: low`. The model service must accept that parameter.

The application chart supplies these settings. Ordinary users do not need to edit a ConfigMap. When diagnosing an older installation, ensure it includes the corrected chart configuration before resubmitting documents.
:::

### Why do model calls report an authentication error?

Check the API key's permissions and expiry in Router. When a valid key is required, update `LLM_API_KEY` in **Settings** > **Applications** > **TrustGraph** > **Manage environment variables**, then apply the change.

### What should I do if an upload fails?

Refresh the page and sign in again. Check whether the document was already created before retrying. Persistent WebSocket connection failures can interrupt uploads; inspect the app entrance and runtime logs.

## Learn more

- [Olares Router](olares-router.md): Manage callable models, defaults, and API credentials.
- [TrustGraph documentation](https://docs.trustgraph.ai/): Explore knowledge graphs, provenance, and advanced workflows.
