---
outline: [2, 3]
description: Connect YuE2 Studio to Lares through MCP, turn a musical idea into a song, and play the result in Lares or YuE2 Studio.
head:
  - - meta
    - name: keywords
      content: Olares, YuE2 Studio, Lares, MCP, AI music, song generation, local AI
app_version: "0.0.7"
doc_version: "1.5"
doc_updated: "2026-11-07"
---

# Create a song with YuE2 Studio through Lares

YuE2 Studio provides an MCP server that lets Lares turn a musical idea into a song from the conversation. You can play the result in Lares or open it in YuE2 Studio.

This guide creates a short acoustic pop song from one Lares conversation. You can replace the example description with your own musical idea.

## Learning objectives

In this guide, you will learn how to:

- Connect the YuE2 Studio MCP server to Lares.
- Describe a musical idea and generate a song from the conversation.
- Play the result in Lares or YuE2 Studio.

## Prerequisites

- [YuE2 Studio installed with one model set](yue2-studio.md#download-a-model-set). Keep the app running and wait for **Engine ready**.
- [Lares prepared with Qwen3.8-27B (llama.cpp)](lares.md), installed on the same Olares device.
- Enough accelerator memory for the chat model and YuE2 Studio. Run one music-generation request at a time when both use the same GPU.

## Connect YuE2 Studio to Lares

Add YuE2 Studio's MCP endpoint to Lares so Qwen3.8-27B can call its music tools and create songs from the conversation.

1. Open Olares **Settings** > **Applications** > **YuE2 Studio** > **Entrances**.
2. Select **YuE2 Studio MCP & API** and copy its **Endpoint** URL.
3. Open Lares and go to **Settings** > **MCP**.
4. Click **Add server** and enter the following settings:
    - **Server name:** `yue2studio`
    - **Transport:** `Streamable HTTP`
    - **MCP URL:** Paste the copied endpoint, then add `/mcp` to the end.
    - **Headers:** `{}`
5. Click **Save**. Lares connects to the server and loads its tools.

## Generate a song

1. Start a new Lares conversation and select **Qwen3.8-27B (llama.cpp)**.
2. Describe the song you want in plain language. You can mention its style, mood, theme, vocals, or instruments. For example:

   ```text
   Create a warm acoustic pop song about seeing the first light over the sea. Use warm female vocals, acoustic guitar, piano, and light drums.
   ```

3. Lares might ask for permission before calling a YuE2 Studio tool. Review the request and allow it to continue.
4. Keep the current conversation in the foreground while Lares generates the song. Wait until Lares reports that generation is complete.
5. Listen to the result in either place:
    - In the current Lares conversation, ask Lares to play the song it just created.
    - Open YuE2 Studio, go to **Library**, and play the latest generated song.

## Troubleshoot the connection

### Why can't Lares load the YuE2 Studio tools?

Confirm YuE2 Studio is running, the transport is **Streamable HTTP**, and the URL ends with `/mcp`.

## Learn more

- [Create and edit a song with YuE2 Studio](yue2-studio.md): Use the Studio interface to compose, edit, and export music.
- [Protect YuE2 Studio MCP with token authentication](yue2-studio-mcp-auth.md): Require a Bearer token for Lares and other MCP clients.
- [Get started with Lares](lares.md): Prepare Qwen3.8-27B (llama.cpp) and review how app connections work.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change an app setting and apply the restart.
