---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/cognee
outline: [2, 3]
description: Build and explore document knowledge with Cognee on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Cognee, self-hosted, knowledge graph, documents, search"
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-11-06"
---

# Turn documents into memory your AI can recall

Cognee turns documents into connected knowledge that you and your AI agents can recall. Add documents to a dataset, ask questions and check the supporting sources, or let your AI agents recall the same knowledge through MCP. On Olares, your documents and the knowledge built from them stay on your own device. Answer generation uses the default chat model configured in Olares Router.

## Learning objectives

- Create a dataset and add project documents to it.
- Ask questions in Cognee and check the supporting sources and knowledge graph.
- Connect Lares or Hermes Agent to Cognee through MCP, then ask questions in a new conversation.

## Prerequisites

- Olares 1.12.7 or later.
- Olares Router with a callable model assigned to `default-chat`.

## Install Cognee

1. Open Market and search for "Cognee".
2. Click **Get**, then **Install**. When prompted, set an initial password with at least 12 characters.
3. When the installation finishes, open Cognee from Launchpad.
4. Sign in with the built-in initial account `default_user@example.com` and the password you set during installation.

## Create a project dataset

Create a dataset and add your project documents to it. One dataset can hold one or many documents, and a search covers every document in it, so keep related documents together. You can paste text or upload a file. This guide uses the two sample documents below.

1. Open **Datasets** and click **Create dataset**. If you already have a dataset, click **New dataset**.
2. Name the dataset `project_rehearsal`. Use letters, numbers, and underscores in the name.
3. Select the new dataset.
4. Choose **Paste text** and add each document below as a separate item. Wait until the first item shows **Ready** before adding the second.

   ::: details Meeting notes

   ```text
   Project Harbor: Migration Rehearsal Meeting Notes

   This document contains fictional test data. All people, dates, and project details are invented for a documentation walkthrough.

   Project: Harbor document migration
   Meeting date: November 4, 2026
   Rehearsal date: November 12, 2026, at 10:00 AM, Asia/Shanghai

   Maya Chen coordinates the migration rehearsal. She schedules the rehearsal and collects the validation results.

   Before handoff, the team must complete the rehearsal, confirm that the source and destination file counts match, verify the SHA-256 checksums, and attach the validation report to the handoff record.

   Omar Lee is the project owner. Handoff requires his approval after he reviews the validation report.

   The migration runbook identifies the escalation contact for checksum mismatches.
   ```

   :::

   ::: details Migration validation runbook

   ```text
   Project Harbor: Migration Validation Runbook

   This document contains fictional test data. All people and project details are invented for a documentation walkthrough.

   Scope: Validation for the Harbor document migration rehearsal.

   File validation compares the source and destination file counts and SHA-256 checksums. The validation report records the results and any unresolved differences.

   Noah Zhou is the data integrity lead and the escalation contact when migration checksums do not match. The report sent to Noah must identify the affected files and include the expected and actual checksums.

   A checksum mismatch blocks handoff until it has been investigated, resolved, and checked again. Any remaining mismatch must be recorded as unresolved in the validation report.

   Maya Chen coordinates the rehearsal. Omar Lee, the project owner, approves handoff after reviewing the completed validation report.
   ```

   :::

   :::info First processing job
   The first document may take longer to process because Cognee needs to download its embedding model. This small model runs on CPU inside Cognee. Later processing jobs reuse the saved model cache. Wait for the current import to finish before submitting the same document again.
   :::

## Query a project dataset

Ask a question about your dataset, then check the answer, evidence, and graph against the original documents.

1. Open **Search**, select `project_rehearsal` in the dataset selector above the question input, and then ask:

   ```text
   Who coordinates the rehearsal, what must be complete before handoff, and who should we contact if migration checksums do not match? Distinguish the meeting notes from the runbook in your sources.
   ```

2. Verify the answer and compare the supporting text in **Evidence** with the two original documents. If an excerpt is truncated, do not treat it as a complete source check.

   In this example, the answer identifies Maya Chen as the coordinator, explains the validation and approval required before handoff, and names Noah Zhou as the checksum-mismatch contact.

   :::tip Chat model connection
   If answer generation fails, make sure a default chat model is assigned on the **Default models** page and shows **callable** status. See [Olares Router](olares-router.md).
   :::

3. Open **Graph** and select `project_rehearsal` in the **DATASET** selector in the upper-left corner. Layouts and generated relationships can vary with the model.

## Query in Lares or Hermes Agent

Connect Lares or Hermes Agent to Cognee's MCP server so you can search your datasets in a conversation without pasting the documents again.

### Get Cognee's MCP address

Copy the MCP endpoint URL from Olares Settings, then add `/mcp` to it.

1. Open Olares **Settings**, go to **Applications** > **Cognee** > **Cognee MCP**, and then copy its endpoint URL. For example, `https://ba2717462.laresprime.olares.com`.
2. Add `/mcp` to the end of the copied URL. This is the MCP URL for the assistant connections below.

:::info Default access without a key
By default, no client API key is needed. The MCP server uses the built-in initial account, `default_user@example.com`, so assistants can access the knowledge you added with that account. Connected assistants share that account's permissions. The official tools can add and delete knowledge as well as retrieve it. Review tool requests before approving them.
:::

### Connect Lares

Register Cognee as an MCP server in Lares, then ask a question that uses the `recall` tool.

1. Open Lares, open its settings, select **MCP**, then click **Add server**. Enter:

   | Field | Value |
   | --- | --- |
   | **Server name** | `cognee` |
   | **Transport** | **Streamable HTTP** |
   | **MCP URL** | The MCP URL obtained above, including `/mcp` |
   | **Headers** | Keep the default empty JSON object, `{}`. |

2. Click **Save**. Lares connects and loads the official tools, including `recall` and `remember`.
3. Start a new conversation and send the prompt below. Approve the `recall` tool when prompted.

      ```text
      Use Cognee recall to search dataset project_rehearsal for who coordinates the rehearsal and who to contact when migration checksums do not match. Set include_references to true.
      ```

   The prompt specifies three things:

   | Part | Required | What it does |
   | --- | --- | --- |
   | `recall` tool | Yes | Searches your stored knowledge in Cognee. |
   | Dataset name | Yes | `recall` searches one dataset at a time. Use the exact name shown in Cognee. |
   | `include_references` | No | Also return the source text behind the answer. Set to `true` here so you can compare it with the original in the next step. Omit it if you only need the answer. |

4. Compare the answer and the returned source text with the original documents you added.

### Connect Hermes Agent

1. Open Files and go to **Data** > **hermesagent** > **home**.
2. Double-click `config.yaml` and click the edit icon in the upper-right corner.
3. Add the following connection under `mcp_servers`, replacing any existing `cognee` entry and keeping other entries unchanged. Replace `<mcp-url>` with the MCP URL obtained above. This example enables only the `recall` tool.

   ```yaml
   mcp_servers:
     cognee:
       url: <mcp-url>
       tools:
         include:
           - recall
   ```

4. Save the file changes.
5. Open **Hermes CLI** from Launchpad.
6. If a Hermes Agent session is open, enter `/exit` to return to the shell.
7. Run `hermes` to start a new agent session and load the configuration. Use the same recall prompt as Lares.
8. Compare the answer and the returned source text with the original documents you added.

## Learn more

- [Olares Router](./olares-router.md): Manage model routes, including the `default-chat` endpoint.
- [Cognee documentation](https://docs.cognee.ai/): Official reference for datasets, graphs, and the Cognee API.
