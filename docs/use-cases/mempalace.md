---
outline: [2, 3]
description: Keep project notes and conversation memory on Olares with MemPalace. Import documents, find the original wording, and share the same memory between Hermes Agent and OpenClaw through MCP.
head:
  - - meta
    - name: keywords
      content: Olares, MemPalace, AI memory, MCP, Hermes Agent, OpenClaw, semantic search, local memory
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-25"
---

# Share project memory across AI agents with MemPalace

MemPalace is a local memory service for AI agents. It keeps the original wording of notes and conversations, retrieves relevant passages by meaning, and organizes records into projects and topics. Its knowledge graph can also track how facts change over time.

On Olares, your imported memory persists independently of an agent's chat sessions. Hermes Agent and OpenClaw can use the same memory through MCP, while Files holds your source documents. The default text embedding and search run locally without a cloud API key.

This guide uses a fictional kitchen renovation project called Lakeview. The same workflow works for a freelance client's brief, a household project, or a software team's design decisions.

## Learning objectives

- Import project notes from Files and retrieve the original passages.
- Connect Hermes Agent to MemPalace and recall a decision in a new conversation.
- Share the same project memory with OpenClaw.
- Add office documents and understand the roles of diaries, temporal facts, and organization tools.

## Where MemPalace fits

| Your task | MemPalace capability |
| --- | --- |
| Keep a client's preferences and decisions available between meetings | Original records, semantic search, and source filters |
| Hand a project from one assistant to another | Shared MCP memory and agent diaries |
| Find a detail in proposals, spreadsheets, or meeting exports | Project, conversation, and office-document imports |
| Track a changing budget, supplier, or technical decision | Temporal knowledge graph, room links, and organization audits |

MemPalace opens a web terminal. You can use its commands directly or connect an agent to its MCP endpoint.

## Prerequisites

- Olares 1.12.7 or later.
- [Hermes Agent](hermes.md#configure-hermes-agent), connected to Router with `default-chat` as its model.

For the model address and authentication rules, see [Connect AI apps](../manual/best-practices/connect-ai-apps.md). This workflow uses local terminal conversations and does not require a messaging channel.

::: info Listing availability
This guide applies to the MemPalace 3.11.0 test package. Follow the installation steps when this version is available in your Market, or use the test package supplied with the app submission.
:::

## Install MemPalace

1. Open Market and search for "MemPalace".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open MemPalace Terminal from Launchpad and run:

   ```bash
   mempalace status
   ```

The first import or memory write downloads the default local embedding model, approximately 80 MB. Installation can finish before this download is needed. If a model download times out, retry the same operation after the connection recovers. The downloaded model is cached for later use.

## Import a project from Files

1. Save this fictional project note as `brief.md` on your computer:

   ```markdown
   # Lakeview kitchen renovation

   The approved budget is EUR 18400.
   Cabinet delivery is booked for 10 November 2026.
   Installation starts on 12 November 2026.
   The delivery contact is Olivia Carter.
   The client prefers oak cabinetry and warm white, dimmable lighting.
   Northstar Joinery will supply the cabinets.
   ```

2. Open Files. In **Home**, create a folder named `MemPalace-demo` and upload `brief.md` into it.
3. In MemPalace Terminal, import that folder:

   ```bash
   mempalace mine /home/app/userdata/Home/MemPalace-demo --wing lakeview-home
   ```

   Wait for `Done.` and the number of filed drawers. The Home mount is read-only. MemPalace stores the imported memory in its own application data.

4. Ask a question using the terminology from the note:

   ```bash
   mempalace search "When is cabinet delivery?" --wing lakeview-home --results 3
   ```

   Read the returned passage and its source filename. You can use different wording from the source, then check the original text before making a project decision.

   ![A project question returns the original Lakeview note](https://cdn.olares.com/images/2026/10/ee7fefa1275966f278391220fd52869a.jpg)

Wings group projects, rooms group topics within a wing, and drawers hold individual records. Re-running an import skips unchanged files. Use the same wing for subsequent imports of this project.

## Connect Hermes Agent to the same memory

### Copy the MCP endpoint

1. Open **Settings** > **Applications** > **MemPalace**.
2. Under **Entrances**, select **MemPalace MCP**.
3. Copy the **Endpoint** URL and append `/mcp` once.

   ![Copy the Endpoint from the MemPalace MCP entrance](https://cdn.olares.com/images/2026/10/dd2af69c620cb23c62900ad5f868787f.jpg)

Use the MCP entrance for this connection. Keep its **Internal** authentication level. The following examples connect agents running on the same Olares.

### Add the connection in Hermes

1. Open Files and locate **Data** > **hermesagent** > **home** > `config.yaml`.
2. Add the `mempalace` entry under `mcp_servers`. Replace the example URL with the endpoint you copied:

   ```yaml
   mcp_servers:
     mempalace:
       url: "https://<your-mempalace-mcp-endpoint>/mcp"
       tools:
         include:
           - mempalace_status
           - mempalace_search
           - mempalace_add_drawer
           - mempalace_get_drawer
           - mempalace_diary_write
   ```

3. Add `mcp-mempalace` to the existing `platform_toolsets.cli` list:

   ```yaml
   platform_toolsets:
     cli:
       - hermes-cli
       - mcp-mempalace
   ```

   Merge these entries into the existing sections. Keep your current model, other MCP servers, and existing toolsets.

4. Save the file and start a new conversation in Hermes CLI:

   ```bash
   hermes chat
   ```

### Save a decision and recall it later

Ask Hermes:

```text
Use MemPalace to save this exact project note in wing lakeview-agent,
room client-meeting:
"Lakeview client meeting: cabinet delivery is booked for 10 November 2026.
Delivery contact is Olivia Carter. Project reference: LV-7421."
Confirm after the memory tool succeeds.
```

After the save succeeds, end the conversation and start a new one. Ask it to search the shared memory:

```text
Search MemPalace for the Lakeview client meeting.
What delivery date, contact, and project reference did we record?
Include the source drawer ID.
```

Open the returned record or compare the quoted text with your note. To make a fact available in later sessions, ask the agent to save it through MemPalace or import its source file.

## Use the memory from OpenClaw

This optional path uses OpenClaw 2026.10.5 with native MCP support. First [configure OpenClaw](openclaw.md#initialize-openclaw) to use Router's `default-chat` model.

1. Open OpenClaw CLI and register the same MemPalace endpoint:

   ```bash
   openclaw mcp add mempalace \
     --url "https://<your-mempalace-mcp-endpoint>/mcp" \
     --transport streamable-http \
     --include mempalace_search,mempalace_add_drawer,mempalace_get_drawer,mempalace_status,mempalace_diary_write
   ```

2. Run `openclaw mcp probe`. The connection should list five tools for `mempalace`.
3. Restart OpenClaw from Market, wait until it is running, and open a new Control UI conversation.
4. Ask OpenClaw to search MemPalace for the Lakeview client meeting and return its delivery date, contact, project reference, and drawer ID.

Both agents now use the same records. You can also ask OpenClaw to save an installation appointment, then ask Hermes to retrieve it in a new conversation.

## Add documents and conversation exports

Upload additional source files into a folder in Files, then choose the matching import mode in MemPalace Terminal:

| Source | Command |
| --- | --- |
| Project notes, Markdown, and code | `mempalace mine <folder> --wing <project>` |
| Conversation exports | `mempalace mine <folder> --mode convos --wing <project>` |
| PDF, DOCX, PPTX, XLSX, RTF, and EPUB | `mempalace mine <folder> --mode extract --wing <project>` |

Replace `<folder>` with the mounted path, such as `/home/app/userdata/Home/MemPalace-demo`, and `<project>` with your wing name. Document extraction dependencies are included in this package.

Office-document imports extract text for retrieval. Keep the original documents in Files for their layout and images. Scanned pages need OCR before this text-import workflow. To work on one topic, add `--wing` and `--room` to a search. Run `mempalace search --help` for source and date filters.

## Keep a growing memory useful

MemPalace provides additional tools for longer projects:

| Need | Tool |
| --- | --- |
| Leave a handoff note after a meeting | Agent diary write and read tools |
| Record a revised budget while retaining the earlier fact | Knowledge-graph add, supersede, and timeline tools |
| Connect a supplier topic to its project brief | Room tunnels and graph traversal |
| Exchange task progress or an exact text artifact | Task events, acknowledgements, and artifact tools |

To use these from Hermes, add the relevant tool names to `tools.include`, or remove that filter to discover the full MemPalace tool set. For example, the temporal-fact tools are `mempalace_kg_add`, `mempalace_kg_query`, `mempalace_kg_supersede`, and `mempalace_kg_timeline`. Start a new conversation after changing the configuration.

For an organization overview, run this in MemPalace Terminal:

```bash
mempalace audit
```

::: details Optional: Propose and apply topic categories with Router
Use this after a wing has accumulated enough notes to benefit from topic categories. Copy the Router **Base URL** for **Apps in Olares**, including `/v1`, as described in [Connect AI apps](../manual/best-practices/connect-ai-apps.md).

Finish active agent tasks first. The maintenance command pauses MCP while it works and automatically resumes the service afterward.

```bash
mempalace-maintenance rooms propose --wing lakeview-home \
  --llm-provider openai-compat \
  --llm-model default-chat \
  --llm-endpoint "https://<your-router-endpoint>/v1" \
  --accept-external-llm
```

The last flag permits sending sampled notes to the model endpoint you specified. Whether the model runs locally or in the cloud depends on Router's current `default-chat` route.

Review the proposed categories, then preview the reassignment:

```bash
mempalace-maintenance rooms apply --wing lakeview-home --from all
```

Apply the plan when you are satisfied with it:

```bash
mempalace-maintenance rooms apply --wing lakeview-home --from all --yes
```
:::

## Keep your source files and memory

The source files remain in Files. Imported memory, diaries, and graph data persist in MemPalace's application storage. For a manual copy, stop MemPalace from Market, copy **Data** > **mempalace** > **mempalace** to your backup location in Files, then resume the app. Keep the source documents in your normal Files backup as well.

The default embedding model and search run on Olares. Agent conversations and optional model-assisted organization follow the model provider configured through Router. Select a local model there when you want those model calls to stay on your device.
