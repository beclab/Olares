---
outline: [2, 3]
description: Draft an event checklist, keep planning notes in workspace files, and resume past conversations with Hermes Workspace, then connect Hermes Agent for tool tasks.
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI chat, workspace files"
app_version: "0.0.9"
doc_version: "1.2"
doc_updated: "2026-11-05"
---

# Chat, save files, and run agent tasks with Hermes Workspace

Hermes Workspace pairs AI chat sessions with file storage on your Olares. Conversations and files remain available after the application restarts.

This guide uses a small library event to show three everyday moves: drafting a checklist with your model, keeping related notes in a workspace file, and reopening a past conversation. An optional final section connects Hermes Agent for tool tasks.

## Learning objectives

In this guide, you will learn how to:

- Draft a short event checklist with Router.
- Keep planning notes in a workspace file.
- Resume a saved conversation with a follow-up.
- Connect Hermes Agent for tool tasks.

## Prerequisites

- Olares 1.12.7 or later.
- A working `default-chat` route in [Router](olares-router.md).

## Install Hermes Workspace

Only one setting is required during installation: the password you will use to sign in.

1. Open Market and search for "Hermes Workspace".
2. Click **Get**, then **Install**, and wait for installation to complete. Set **WORKSPACE_PASSWORD** to a password of at least 12 characters when prompted.

## Draft an event checklist

Send one chat request to see how the workspace talks to your model, then use the reply as a starting draft for your own event.

1. Open Hermes Workspace from Launchpad and sign in with the password you set during installation.
2. Click **New Hermes chat**. Keep **Default Chat** selected.
3. Enter this example request and send it:

   ```text
   Create a concise checklist for organizing a small community library event.
   Use three bullet points. Do not call tools or modify files.
   ```

4. Read the reply. It is a generic checklist that reads like a finished plan, so go through each item and adjust whatever does not match your event, such as the venue, helpers, or promotion plan.

## Keep related notes

Store planning notes in a workspace file so they sit next to the conversation that produced them.

1. Go to **Workspace** > **Files**, and create a new file named `library-event.md`.
2. Open the file, add your planning notes, and save your changes. You can also use **Upload** to add an existing file.

   ```text
   # Library event plan

   - Date: Saturday, March 14, 2:00-4:00 pm
   - Venue: Community room, Central Library
   - Staff: 2 desk volunteers, 1 setup helper
   - Activities: storytime, book sale, sign-up desk
   - To confirm: guest speaker reply due March 1
   ```

Files stay in this application's workspace. They are not visible to the model until you attach them in a chat.

## Return to a conversation

Every chat is saved as a session. Reopen one to pick up a past conversation where you left off.

1. Go to **Chat** > **Hermes**, and then select the conversation under **SESSIONS** to reopen its history.
2. Send a follow-up request to continue where you left off, such as asking for a schedule for the event day based on the checklist.

## Connect Hermes Agent for tool tasks

By default, the workspace sends plain chat requests to your model. This section connects a Hermes Agent backend so chats can also run tool tasks.

### Configure Hermes Agent

1. Install and configure [Hermes Agent](hermes.md) with the Router `default-chat` route.
2. Open Olares Settings, go to **Applications** > **Hermes Agent** > **Manage environment variables**:

   a. Enable **API_SERVER_ENABLED**.
   
   b. Set **HERMES_API_SERVER_KEY** to a strong key of at least 16 characters. Keep this key private.

3. For Hermes Agent 0.22.2, keep the API settings in its persistent profile as well:

   a. Open Olares Files, go to **Data** > **hermesagent** > **home**.
   
   b. Open `.env`, then set `API_SERVER_ENABLED=true` and `API_SERVER_KEY` to the same key.

4. Open Hermes CLI from the Launchpad and run:

   ```bash
   hermes config set platforms.api_server.enabled true
   hermes gateway restart
   ```

   If the chat interface opens instead of a terminal, type `/quit` to return to the terminal, then run the commands.

   Wait for the gateway to restart before connecting Workspace. This prevents the API from remaining disabled after an Agent restart.

### Connect the workspace

1. Open Olares Settings, go to **Applications** > **Hermes Agent** > **Entrances**.
2. Obtain the endpoint URLs for **Hermes Gateway API** and **Dashboard**.
3. Open **Applications** > **Hermes Workspace** > **Manage environment variables**, then set these fields:

   | Field | Value |
   | :--- | :--- |
   | **HERMES_GATEWAY_URL** | The Hermes Gateway API endpoint URL |
   | **HERMES_DASHBOARD_URL** | The Dashboard endpoint URL |
   | **HERMES_API_TOKEN** | The same value as Hermes Agent's `HERMES_API_SERVER_KEY` |

4. Click **Apply** and wait for the app restart to finish.
5. Reopen Hermes Workspace. Agent skills and scheduled jobs become available through the connected backend.
6. Start a new conversation and ask the agent to create a small event checklist in its own working directory, read the file back, and report the result. Review the tool results along with the reply.

   ```text
   Create a small event checklist in your working directory for a community library event.
   Save it as library-event-checklist.md, then read the file back and report what you wrote.
   ```

Hermes Agent executes tools in its own container. Its files and persistent memory are separate from the browser workspace files and the Workspace **Memory** notes. Use the Hermes Agent storage and Dashboard to inspect agent-created files and memory.

The connected backend supports tool tasks, skill management, persistent memory, and scheduled jobs. A job can execute model requests while unattended. Pause a trial job after checking its output.

A delegated child can finish and return a result even when the parent conversation has already ended its turn. Ask the parent to continue if its final report has not appeared. Full Workspace multi-agent orchestration and MCP management depend on backend capabilities beyond this walkthrough.

## Learn more

- [Hermes Agent](hermes.md): The CLI-based agent that provides tool tasks, skills, and scheduled jobs for the connected workspace.
- [Router](olares-router.md): The model routes behind `default-chat`, including how to configure the route used in this guide.
