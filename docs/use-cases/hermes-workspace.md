---
outline: [2, 3]
description: Draft an event checklist, keep planning notes in workspace files, and resume past conversations with Hermes Workspace, then connect Hermes Agent so chats can take actions.
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI chat, workspace files"
app_version: "0.0.10"
doc_version: "1.0"
doc_updated: "2026-11-05"
---

# Chat, save files, and take actions with Hermes Workspace

Hermes Workspace is an AI workspace on your Olares. You chat with your model in it and keep related files in the same place. Conversations and files remain available after the application restarts.

This guide uses a small library event as the running example. An optional final section connects Hermes Agent, so your chats can take actions like creating files and running commands.

## Learning objectives

In this guide, you will learn how to:

- Draft an event checklist in Hermes Workspace.
- Keep planning notes in a workspace file.
- Resume a saved conversation with a follow-up.
- Connect Hermes Agent so chats can take actions.

## Prerequisites

- Olares 1.12.7 or later.
- A working `default-chat` route in [Router](olares-router.md). Hermes Workspace chats through this route, so plain chat needs no extra backend setup.

## Install Hermes Workspace

Only one setting is required during installation: the password you will use to sign in.

1. Open Market and search for "Hermes Workspace".
2. Click **Get**, then **Install**, and wait for installation to complete. Set **WORKSPACE_PASSWORD** to a password of at least 12 characters when prompted.

## Chat with your model

Send one chat request to see how Hermes Workspace talks to your model, then use the reply as a starting draft for your own event.

1. Open Hermes Workspace from Launchpad and sign in with the password you set during installation.
2. On the welcome screen, click **Skip setup** to start with plain chat.
3. Click **New Hermes chat**. Keep **Default Chat** selected.
4. Enter this example request and send it:

   ```text
   Create a concise checklist for organizing a small community library event.
   Use three bullet points. Do not call tools or modify files.
   ```

5. Read the reply. In this example, the result is a three-bullet checklist that covers the venue and date, the program, and promotion. It reads like a finished plan, so go through each item and adjust whatever does not match your event.

## Store files in the workspace

You adjusted the checklist with your own event details in the previous step. Save those details as a workspace file, so they are there when you come back to this workspace.

1. Go to **Workspace** > **Files** and create a new file named `library-event.md`.
2. Double-click the file, click **Edit** in the upper-right corner, add your planning notes, and then click **Save**. You can also use **Upload** to add an existing file.

   ```text
   # Library event plan

   - Date: Saturday, March 14, 2:00-4:00 pm
   - Venue: Community room, Central Library
   - Staff: 2 desk volunteers, 1 setup helper
   - Activities: storytime, book sale, sign-up desk
   - To confirm: guest speaker reply due March 1
   ```

Saving a file here does not share it with the model. To let the model read a file, attach it in a chat.

## Continue past conversations

Every chat is saved as a session. Reopen one to pick up a past conversation where you left off.

1. Go to **Chat** > **Hermes**, and then select the conversation under **SESSIONS** to reopen its history.
2. Send a follow-up request to continue where you left off, such as asking for a schedule for the event day based on the checklist.

## Connect Hermes Agent

By default, Hermes Workspace sends plain chat requests to the model behind your `default-chat` route. This section connects a Hermes Agent backend so chats can also take actions. Sidebar features such as **Jobs**, **Skills**, and **MCP** stay locked until this backend is connected.

### Configure Hermes Agent

1. Install and configure [Hermes Agent](hermes.md) with the Router `default-chat` route.
2. Open Olares Settings, go to **Applications** > **Hermes Agent** > **Manage environment variables**:

   a. Enable **API_SERVER_ENABLED**.
   
   b. Set **HERMES_API_SERVER_KEY** to a strong key of at least 16 characters. Keep this key private.

   c. Click **Apply** and wait for Hermes Agent to restart.

3. For Hermes Agent 0.22.2, keep the API settings in its persistent profile as well:

   a. Open Olares Files, go to **Data** > **hermesagent** > **home**.
   
   b. Open `.env`, then set `API_SERVER_ENABLED=true` and `API_SERVER_KEY` to the same key.

4. Open Hermes CLI from the Launchpad and run:

   ```bash
   hermes config set platforms.api_server.enabled true
   hermes gateway restart
   ```

   If the chat interface opens instead of a terminal, type `/quit` to return to the terminal, then run the commands.

   Wait for the gateway to restart before connecting Hermes Workspace. This prevents the API from remaining disabled after a Hermes Agent restart.

### Connect Hermes Workspace

1. Open Olares Settings, go to **Applications** > **Hermes Agent** > **Entrances**.
2. Click **Hermes Gateway API** and copy its URL from **Endpoint**. Repeat this for **Dashboard** to obtain its URL.
3. Open **Applications** > **Hermes Workspace** > **Manage environment variables**, then set these fields:

   | Field | Value |
   | :--- | :--- |
   | **HERMES_GATEWAY_URL** | The **Hermes Gateway API** URL you copied in step 2 |
   | **HERMES_DASHBOARD_URL** | The **Dashboard** URL you copied in step 2 |
   | **HERMES_API_TOKEN** | The same key you set for Hermes Agent's `HERMES_API_SERVER_KEY` |

4. Click **Apply** and wait for the app restart to finish.
5. Reopen Hermes Workspace. The Hermes Agent backend is now connected.
6. Start a new conversation and ask Hermes Agent to create a small event checklist in its own working directory, read the file back, and report the result. Review the tool results along with the reply.

   ```text
   Create a small event checklist in your working directory for a community library event.
   Save it as library-event-checklist.md, then read the file back and report what you wrote.
   ```

7. Optional: To locate a file after the conversation, check the store that owns it.

   - **Workspace files**: Files you create or upload in Hermes Workspace are stored under **Data** > **hermesworkspace** > **data** > **files** in Files, and appear under **Workspace** > **Files** in the app.
   - **Agent files**: Files Hermes Agent created are under **Data** > **hermesagent** > **home** in Files.

## Learn more

- [Hermes Agent](hermes.md): The CLI-based agent backend for the connected workspace.
- [Router](olares-router.md): The model routes behind `default-chat`, including how to configure the route used in this guide.
