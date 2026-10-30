---
outline: [2, 3]
description: Use Hermes Workspace on Olares to draft an event checklist with Router and keep related files together.
head:
  - - meta
    - name: keywords
      content: "Olares, Hermes Workspace, Router, AI chat, workspace files"
app_version: "0.0.7"
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

## Advanced agent features

This package provides basic Router conversations and workspace files. Agent tools, scheduled jobs, skills, memory, and orchestration require a separately installed and configured Hermes Agent gateway and dashboard. They are not enabled by the basic Router connection.

See [Hermes Agent](hermes.md) for the agent application. Connecting an external gateway is an advanced configuration and is outside this walkthrough.
