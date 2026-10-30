---
outline: [2, 3]
description: Use Hermes Workspace on Olares to draft an event checklist with Router and keep related files together.
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI chat, workspace files"
app_version: "0.0.9"
doc_version: "1.0"
doc_updated: "2026-10-30"
---

# Plan a library event with Hermes Workspace

Hermes Workspace brings AI conversations and workspace files into your browser. On Olares, the packaged workspace connects to Router through `default-chat` and keeps your conversations and files in application storage.

## Learning objectives

In this guide, you will learn how to:

- Draft a short event checklist with Router.
- Keep planning notes in a workspace file.
- Return to a saved conversation.

## Prerequisites

- Olares 1.12.7 or later.
- A working `default-chat` route in [Router](olares-router.md).

## Install Hermes Workspace

1. Open Market and search for "Hermes Workspace".

   <!-- ![Hermes Workspace](/images/manual/use-cases/hermes-workspace.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete. Set **WORKSPACE_PASSWORD** to a password of at least 12 characters when prompted.

## Draft an event checklist

1. Open Hermes Workspace from Launchpad and sign in with the password you set during installation.
2. Click **New Hermes chat**. Keep **Default Chat** selected.
3. Enter this example request and send it:

   ```text
   Create a concise checklist for organizing a small community library event.
   Use three bullet points in English. Do not call tools or modify files.
   ```

4. Read the reply and adapt the date, venue, staffing, and activities to your event. Model responses vary with the model configured in Router.

![A saved conversation in Hermes Workspace](/images/manual/use-cases/hermes-workspace-chat.jpg#bordered)

## Keep related notes

Open **Workspace** and click **New file**. Name the file `library-event.md`. Open it, add your planning notes, and save your changes. You can also use **Upload** to add an existing file.

Files stay in this application's workspace. Adding a file does not automatically send its contents to the model. Use the chat attachment control when you want to include a supported file in a conversation.

## Return to a conversation

Select the conversation under **SESSIONS** to reopen its history. Conversations, workspace files, and local settings remain available after the application restarts.

## Connect Hermes Agent for tool tasks

Install and configure [Hermes Agent](hermes.md) with Router `default-chat`. In its application environment settings, enable **API_SERVER_ENABLED** and set **HERMES_API_SERVER_KEY** to a strong key of at least 16 characters. Keep this key private.

For Hermes Agent 0.22.2, keep the API settings in its persistent profile as well. In Files, open **Data** > **hermesagent** > **home** > **.env**, then set `API_SERVER_ENABLED=true` and `API_SERVER_KEY` to the same key. Preserve other entries and keep this file private. In the Hermes CLI terminal, run:

```bash
hermes config set platforms.api_server.enabled true
hermes gateway restart
```

Wait for the gateway to restart before connecting Workspace. This prevents the API from remaining disabled after an Agent restart.

1. In Settings, open **Applications** > **Hermes Agent** > **Entrances**. Obtain the URLs for **Hermes Gateway API** and **Dashboard** from this instance.
2. Open **Applications** > **Hermes Workspace** > **Environment**, then set these fields:

   | Field | Value |
   | :--- | :--- |
   | **HERMES_GATEWAY_URL** | The Hermes Gateway API entrance URL |
   | **HERMES_DASHBOARD_URL** | The Dashboard entrance URL |
   | **HERMES_API_TOKEN** | The same value as Hermes Agent's HERMES_API_SERVER_KEY |

3. Wait for the application to apply the changes, then reopen Hermes Workspace. Agent skills and scheduled jobs become available through the connected backend.
4. Start a new conversation and ask the agent to create a small event checklist in its own working directory, read the file back, and report the result. Review the tool results along with the reply.

Hermes Agent executes tools in its own container. Its files and persistent memory are separate from the browser workspace files and the Workspace **Memory** notes. Use the Hermes Agent storage and Dashboard to inspect agent-created files and memory.

The connected backend supports tool tasks, skill management, persistent memory, and scheduled jobs. A job can execute model requests while unattended. Pause a trial job after checking its output.

A delegated child can finish and return a result even when the parent conversation has already ended its turn. Ask the parent to continue if its final report has not appeared. Full Workspace multi-agent orchestration and MCP management depend on backend capabilities beyond this walkthrough.
