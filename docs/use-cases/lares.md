---
outline: [2, 3]
description: Get started with Lares on Olares. Connect a model through Router, choose an access level, run your first task, and find guides for web research and app automation.
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Olares Router, local AI agent, AI assistant, MCP, natural language, Olares 1.12.7
app_version: "0.26.60"
doc_version: "2.0"
doc_updated: "2026-11-07"
---

# Get started with Lares

Lares is the official AI assistant for Olares. It uses models connected through Router and can work with Olares apps, files, and system services through tools that you control.

Start with a small read-only task to check the model and permissions. You can then add web research providers or connect apps such as Blender and YuE2 Studio for task-specific workflows.

## Learning objectives

In this guide, you will learn how to:

- Prepare a chat model for Lares through Router.
- Choose an access level and run your first task.
- Find the right setup guide for web research or an app connection.

## Prerequisites

- Olares 1.12.7 or later.
- Lares 0.26.60 and Router installed from Market.
- A chat model that supports tool calls, installed and available in Router. This guide uses Qwen3.8-27B (llama.cpp).

## Understand how Lares connects to Olares

Lares combines three types of capability. Configure each one where it belongs.

| Capability | Connection | Where to configure it |
|:---|:---|:---|
| Chat and planning | A chat model provided by Router | **Router** > **Default models** |
| Olares tasks | Built-in Olares tools and skills | The access level selected in Lares |
| App-specific actions | An app's MCP server or a built-in integration | **Lares** > **Settings** > **MCP**, or the app-specific guide |

Router supplies the model, while Lares decides when to call an available tool. Your selected access level limits what Lares can do on your behalf. An MCP connection adds tools from one app without changing the permissions of unrelated apps.

## Prepare the model

1. Open Router from Launchpad and go to **Default models**.
2. Select Qwen3.8-27B (llama.cpp), or another tool-capable model, as the default chat model.
3. Confirm the model shows **Callable** before opening Lares.

If you use another model, check that it supports tool calls. A chat-only model can answer questions but cannot reliably operate Olares or a connected MCP server.

## Run your first task

1. Open Lares from Launchpad.
2. Select the chat model and review the access level for the conversation.

   ![Lares chat interface](/images/manual/use-cases/lares1.png#bordered)

3. Start with a read-only request:

   ```text
   Check this device's configuration. Summarize the operating system version,
   processor, memory, storage, and accelerator resources. Do not change anything.
   ```

4. Review any proposed tool call before allowing it. Check that the response describes your device without changing its settings.

:::warning Run one demanding task at a time
When Qwen3.8-27B (llama.cpp) runs on Olares One, wait for the current request to finish before starting another model or media-generation task. Concurrent workloads can compete for memory and make a model unresponsive.
:::

## Add the capability you need

Keep the Lares setup focused on the task you want to complete. Each guide below covers the required app, connection, and first useful result.

- [Research the web with Lares](lares-web-research.md): Find public sources and synthesize them in a conversation.
- [Create a Blender scene with Lares](blender-lares.md): Build and render a 3D scene.
- [Use FlowStudio through Lares](flowstudio-lares.md): Generate images and videos from prepared workflows.
- [Create a song with YuE2 Studio through Lares](yue2-studio-lares.md): Generate a song through the YuE2 Studio MCP server.

An app-specific guide is the source of truth for its endpoint, authentication, readiness checks, and output location. Do not reuse an MCP URL or token from another app.

## Learn more

- [Use Olares Router as your AI gateway](olares-router.md): Manage models and providers used by Lares.
- [Run local LLMs on Olares](llm-base-apps.md): Install and manage a local chat model.
