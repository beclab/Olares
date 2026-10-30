---
outline: [2, 3]
description: Keep project notes and conversation memory on Olares with MemPalace. Import documents, find the original wording, and share the same memory between Hermes Agent and OpenClaw through MCP.
head:
  - - meta
    - name: keywords
      content: Olares, MemPalace, AI memory, MCP, Hermes Agent, OpenClaw, semantic search, local memory
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-28"
---

# Share project memory across AI agents with MemPalace

MemPalace is a local memory service for AI agents. It stores notes and conversations in their original wording and helps you find relevant passages by meaning.

On Olares, your imported memory persists independently of an agent's chat sessions. Hermes Agent and OpenClaw can use the same memory through MCP, while Files holds your source documents. The default text embedding and search run locally without a cloud API key.

This guide follows a fictional kitchen renovation project called Lakeview. You will import its brief, use Hermes Agent to search the imported file and save a meeting note, then retrieve that note in a new conversation and from OpenClaw. The MCP connection gives each agent memory tools. This workflow does not configure automatic saving or recall: you explicitly ask the agents to search and save records.

## Learning objectives

- Import a project brief from Files and search its original text.
- Ask Hermes Agent to read the brief and save a meeting note.
- Retrieve the meeting note in a new Hermes conversation.
- Ask OpenClaw to retrieve the note saved by Hermes.

## Prerequisites

- **Olares OS:** Version 1.12.7 or later.

To make project information easier to organize and find with AI agents, set up these apps before following their sections:

- **Hermes Agent:** Follow the [Hermes Agent setup guide](hermes.md#configure-hermes-agent) and select Router's `default-chat` model.
- **OpenClaw:** Follow the [OpenClaw setup guide](openclaw.md#initialize-openclaw) with OpenClaw 2026.10.5 or later, using Router's `default-chat` model.

## How MemPalace organizes memory

MemPalace organizes stored text in three levels:

| Level | What it holds | In this guide |
| --- | --- | --- |
| **Wing** | Records for a project, person, or other subject | The imported brief goes in `lakeview-home`. Hermes saves the meeting note in `lakeview-agent`. |
| **Room** | Records on a topic within a wing | The brief goes in `general`. The meeting note goes in `client-meeting`. |
| **Drawer** | The original text of a record or part of an imported file | Search results can identify a passage from `brief.md` or the saved meeting note by its drawer ID. |

Both wings are in the same MemPalace memory store and are available to connected agents. Use `--wing` to search one wing. A search limited to `lakeview-home` will not include the meeting note in `lakeview-agent`.

## Install MemPalace

1. Open Market and search for "MemPalace".
2. Click **Get**, then **Install**, and wait for installation to complete.

::: tip First import may take longer
The first import or memory write downloads a local embedding model of about 80 MB. If the download times out, retry the operation when your connection recovers. MemPalace caches the model for later use.
:::

## Import a project from Files

1. Save the following example note as `brief.md` on your computer:

   ```markdown
   # Lakeview kitchen renovation

   The approved budget is EUR 18400.
   Cabinet delivery is booked for 10 November 2026.
   Installation starts on 12 November 2026.
   The delivery contact is Olivia Carter.
   The client prefers oak cabinetry and warm white, dimmable lighting.
   Northstar Joinery will supply the cabinets.
   ```

2. Open Olares Files. In **Home**, create a folder named `MemPalace-demo` and upload `brief.md` into it.
3. Open MemPalace Terminal from Launchpad. The folder at `/Files/Home/MemPalace-demo/` in Olares Files is mounted at `/home/app/userdata/Home/MemPalace-demo` in the terminal. Import it with:

   ```bash
   mempalace mine /home/app/userdata/Home/MemPalace-demo --wing lakeview-home
   ```

   Wait for `Done.` and the number of filed drawers. The Home mount is read-only. MemPalace stores the imported memory in its own application data.

4. Search the imported note for the cabinet delivery date:

   ```bash
   mempalace search "When is cabinet delivery?" --wing lakeview-home --results 3
   ```

   Check that a returned passage gives the delivery date as 10 November 2026 and names `brief.md` as its source.

::: tip About this search
- `--wing lakeview-home` limits the search to the imported brief's wing.
- `--results 3` returns up to three matches. Change `3` to another number to request a different limit.

Use the same wing for later imports of the Lakeview brief.
:::

## Connect Hermes Agent to the same memory

You can use Hermes CLI without a messaging channel. This section shows Hermes reading the imported brief, then saving a note for later conversations.

### Copy the MCP endpoint

1. Open **Settings** > **Applications** > **MemPalace**.
2. Under **Entrances**, select **MemPalace MCP**.
3. Check that **Authentication level** is set to **Internal**.
4. Copy the **Endpoint** URL.

   ![Copy the Endpoint from the MemPalace MCP entrance](/images/manual/use-cases/mempalace-endpoint.png#bordered)

### Add the connection in Hermes

1. Open Olares Files and locate **Data** > **hermesagent** > **home** > `config.yaml`.
2. Click <i class="material-symbols-outlined">edit_square</i> to edit.
3. Add the `mempalace` entry under `mcp_servers`. Replace `<mempalace-endpoint-url>` with the full URL copied above:

   ```yaml
   mcp_servers:
     mempalace:
       url: "<mempalace-endpoint-url>/mcp"
       tools:
         include:
           - mempalace_status
           - mempalace_search
           - mempalace_add_drawer
           - mempalace_get_drawer
           - mempalace_diary_write
   ```

4. Add `mcp-mempalace` to the existing `platform_toolsets.cli` list:

   ```yaml
   platform_toolsets:
     cli:
       - hermes-cli
       - mcp-mempalace
   ```

   Add these entries to the existing lists in `config.yaml` without removing your other MCP servers or CLI toolsets.

5. Click <i class="material-symbols-outlined">save</i> to save the changes.

### Read the imported brief

1. Start a new conversation in Hermes CLI:

   ```bash
   hermes chat
   ```

2. Ask Hermes to search the `lakeview-home` wing:

   ```text
   Search MemPalace's lakeview-home wing for the Lakeview kitchen brief.
   What's the approved budget and who supplies the cabinets? Include the source.
   ```

### Save a meeting note and recall it later

The meeting note repeats the delivery details from the brief and adds a project reference.

1. In your current Hermes CLI conversation, ask Hermes to save the meeting note:

   ```text
   Use MemPalace to save this exact project note in wing lakeview-agent,
   room client-meeting:
   "Lakeview client meeting: cabinet delivery is booked for 10 November 2026.
   Delivery contact is Olivia Carter. Project reference: LV-7421."
   Confirm after the memory tool succeeds.
   ```

   Hermes might confirm the save like this:

   ```text
   Confirmed. The note is in MemPalace under wing lakeview-agent, room client-meeting.
   ```

2. Type `/exit` and press **Enter** to leave Hermes CLI. Then run `hermes chat` to start a fresh conversation.

3. In the new conversation, ask Hermes to search the shared memory:

   ```text
   Search MemPalace for the Lakeview client meeting.
   What delivery date, contact, and project reference did we record?
   Include the source drawer ID.
   ```

4. Check that Hermes finds the meeting note and cites its source drawer.

## Use the memory from OpenClaw

OpenClaw can use the same MemPalace service to retrieve the meeting note saved by Hermes.

1. Open OpenClaw CLI and register the MemPalace MCP endpoint. Replace `<mempalace-endpoint-url>` with the full URL from [Copy the MCP endpoint](#copy-the-mcp-endpoint):

   ```bash
   openclaw mcp add mempalace \
     --url "<mempalace-endpoint-url>/mcp" \
     --transport streamable-http \
     --include mempalace_search,mempalace_add_drawer,mempalace_get_drawer,mempalace_status,mempalace_diary_write
   ```

2. Run `openclaw mcp probe`. The output should show five tools for `mempalace`.
3. In Market, open **My Olares** and find OpenClaw. Open the operation menu, select **Stop**, then select **Resume**. Wait until OpenClaw is running.
4. Open Control UI from Launchpad and select **New Session**.
5. Ask OpenClaw:

   ```text
   Search MemPalace for the Lakeview client meeting.
   What delivery date, contact, and project reference did we record?
   Include the source drawer ID.
   ```

OpenClaw should return the delivery date, contact, and project reference from the meeting note saved by Hermes.

## Add documents and conversation exports

Upload the files to a folder in Olares Files. Then, in MemPalace Terminal, run the command for that file type:

| Source | Command |
| --- | --- |
| Project notes, Markdown, and code | `mempalace mine <folder> --wing <project>` |
| Conversation exports | `mempalace mine <folder> --mode convos --wing <project>` |
| PDF, DOCX, PPTX, XLSX, RTF, and EPUB | `mempalace mine <folder> --mode extract --wing <project>` |

Replace the placeholders before running a command:

- `<folder>`: The folder's mounted path in MemPalace Terminal.
- `<project>`: The wing where MemPalace will store the imported records. Use `lakeview-home` to add files to the brief's wing.

::: tip Document imports
The `extract` mode imports searchable text, not a document's layout or images. Keep the original file in Olares Files, and run OCR on scanned pages before importing them. The required extraction tools are included with MemPalace.
:::

## Beyond the basic workflow

### Review memory organization

Run the following command in MemPalace Terminal to view an organization report for the entire memory store:

```bash
mempalace audit
```

The scorecard rates how well your memory is organized, not search accuracy. Check the findings below the scores for specific issues to review. `n/a` means that an aspect has no score to report.

:::: details Optional: Organize project memory with Router

As a project accumulates records, you can group them into rooms by topic. MemPalace can use a model through Olares Router to suggest rooms for a wing. Review the suggestions and preview where records would move before applying the changes.

The Lakeview example has only one imported brief, so the proposal may include rooms with no records. A wing with several notes on different topics gives MemPalace more records to classify.

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

1. Get the Router endpoint. Go to Olares **Settings** > **Applications** > **Router** > **Entrances**. Select **Router**, then copy the **Endpoint** URL.

   Use this URL in place of `<router-endpoint-url>` below. The command adds `/v1` and uses the model selected as the default chat model in Router.

2. Finish active agent tasks first. The maintenance command pauses MCP while it works and automatically resumes the service afterward. Run the following command in MemPalace Terminal to suggest rooms for `lakeview-home`:

   ```bash
   mempalace-maintenance rooms propose --wing lakeview-home \
     --llm-provider openai-compat \
     --llm-model default-chat \
     --llm-endpoint "<router-endpoint-url>/v1" \
     --accept-external-llm
   ```

   `--accept-external-llm` permits sending sampled notes to the specified model endpoint. Whether the model runs locally or in the cloud depends on Router's current `default-chat` route.

3. After `rooms propose` finishes, find the proposal file path on the `Saved to` line. In MemPalace Terminal, run `cat` followed by that path to review the suggested room names and descriptions. Continue only if they fit your project.

4. Preview how records would be assigned to the proposed rooms:

   ```bash
   mempalace-maintenance rooms apply --wing lakeview-home --from all --show 3
   ```

   This command previews changes without applying them. `--from all` includes records from every room in `lakeview-home`, and `--show 3` displays up to three examples per destination room. Check that the proposed assignments match the records' content.

5. Apply the changes after reviewing the preview:

   ```bash
   mempalace-maintenance rooms apply --wing lakeview-home --from all --yes
   ```

::::

### Back up your memory

The source files remain in Olares Files. Imported memory, diaries, and graph data persist in MemPalace's application storage. For a manual copy, stop MemPalace from Market, copy **Data** > **mempalace** > **mempalace** to your backup location in Files, then resume the app. Keep the source documents in your normal Files backup as well.

The default embedding model and search run on Olares. Agent conversations and optional model-assisted organization follow the model provider configured through Router. Select a local model there when you want those model calls to stay on your device.

## Learn more

- [Official MemPalace guide](https://mempalaceofficial.com/guide/getting-started)
