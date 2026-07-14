---
outline: [2, 3]
title: Edit Penpot files with Cursor via MCP
description: Use Penpot on Olares as a self-hosted design workspace, then connect Cursor through MCP to inspect and modify an active Penpot design file.
head:
  - - meta
    - name: keywords
      content: Olares, Penpot, MCP, Model Context Protocol, Cursor, design collaboration, prototype, self-hosted design tool
app_version: "1.0.15"
doc_version: "1.0"
doc_updated: "2026-05-22"
---

# Use Cursor to inspect and edit Penpot files with MCP

Penpot is an open-source, web-based design and prototyping tool for UI design, interactive prototypes, component systems, and developer handoff. It uses open standards such as CSS, SVG, and HTML, which makes it a practical bridge between design files and frontend implementation.

On Olares, you can run Penpot as a self-hosted design workspace and connect it to Cursor through Penpot MCP. This guide walks through a complete workflow: open a Penpot file, let Cursor read its structure, ask Cursor to add a card-style component, and check the result in Penpot.

## Learning objectives

In this guide, you will learn how to:

- Install Penpot from Market.
- Get the Penpot MCP endpoints from Olares Settings.
- Connect an active Penpot file to the MCP server.
- Configure Cursor to read the connected Penpot file.
- Use Cursor to inspect frames and add a new design element.
- Review the change in Penpot and refine it with follow-up prompts.

## Prerequisites

- Cursor installed on your computer.
- A Penpot file created or imported in Penpot.

## Install Penpot

1. Open Market and search for "Penpot".

   ![Penpot](/images/manual/use-cases/penpot.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Prepare the Penpot workflow

This guide uses a simple task: ask Cursor to inspect a Penpot file and add a card-style component to one frame.

1. Open Penpot from Launchpad.

2. Create a new Penpot file or open an existing one.

3. Select the page and frame you want Cursor to work with.

:::tip Keep the file open
Cursor can only read the Penpot file that is currently active and connected. Keep this browser tab open throughout the workflow.
:::

## Connect Penpot and Cursor through MCP

Get your Penpot MCP endpoints from Olares, then add them to Penpot and Cursor.

### Get the MCP endpoints

1. Open Settings, then go to **Applications** > **Penpot**.
   
   ![Penpot application endpoints](/images/manual/use-cases/penpot-entrances.png#bordered)

2. Go to the target entrance, and find the endpoint URLs for **Penpot MCP Plugin** and **Penpot MCP HTTP**.

   - **Penpot MCP Plugin**: Use this endpoint to install the Penpot plugin.

      ![Copy Penpot MCP Plugin endpoint](/images/manual/use-cases/lp-penpotmcpplugin-endpoint.png#bordered)

   - **Penpot MCP HTTP**: Use this endpoint to connect Cursor.

      ![Copy Penpot MCP HTTP endpoint](/images/manual/use-cases/lp-penpotmcphttp-endpoint.png#bordered)
   
3. Keep this page open or copy both URLs. You will use them in the next steps.

### Install and connect the MCP plugin in Penpot

With the target file open in Penpot:

1. In the Penpot editor, click <i class="material-symbols-outlined">more_vert</i> to open the main menu, then select **Plugins** > **Plugin manager**.
   
   ![Open Penpot plugin manager](/images/manual/use-cases/penpot-plugin-manager.png#bordered)

2. Append `/manifest.json` to your MCP Plugin endpoint. Enter the URL in the following format, then click **Install**:

   ```text
   <your-mcp-plugin-endpoint>/manifest.json
   ```
   
   For example:
   ```text
   https://2550d96f1.laresprime.olares.com/manifest.json
   ```

   ![Add Penpot MCP plugin](/images/manual/use-cases/penpot-add-plugin.png#bordered){width=60%}

3. Review the permission prompt, then click **Allow**.

4. The plugin now appears in the **INSTALLED PLUGINS** section.
   
   ![Penpot MCP plugin installed](/images/manual/use-cases/penpot-plugin-installed.png#bordered){width=60%}

5. In Plugin manager, click **Open** next to the MCP plugin.

6. In the plugin panel, click **CONNECT TO MCP SERVER**.

   ![Connect Penpot plugin to MCP server](/images/manual/use-cases/penpot-plugin-connect.png#bordered){width=95%}

7. Wait until the status changes to **Connected to MCP server**.
   
   ![Penpot plugin connected](/images/manual/use-cases/penpot-plugin-connected.png#bordered){width=40%}

### Configure Cursor as an MCP client

1. Open Cursor on your computer.

2. Go to **Cursor** > **Settings** > **Tools & MCPs**, then click **Add Custom MCP**.

   ![Add Custom MCP in Cursor](/images/manual/use-cases/penpot-cursor-add-mcp.png#bordered)

3. Append `/mcp` to your MCP HTTP endpoint. In `~/.cursor/mcp.json`, add the following configuration:

   ```json
   {
     "mcpServers": {
       "penpot": {
         "url": "<your-mcp-http-endpoint>/mcp"
       }
     }
   }
   ```

   ![Configure Penpot MCP in Cursor](/images/manual/use-cases/penpot-cursor-mcp-config.png#bordered)

   :::warning Check your JSON syntax
   Ensure you copy the exact format above, including all quotation marks `"` and braces `{}`. Invalid JSON will cause Cursor to fail to load the MCP server.
   :::

4. Save the file. On macOS, press `Cmd + S`. On Windows, press `Ctrl + S`.

5. In **Tools & MCPs**, enable the switch next to **penpot**. If it does not appear, restart Cursor and reopen **Tools & MCPs**.
   
   ![Enable Penpot MCP in Cursor](/images/manual/use-cases/penpot-cursor-mcp-enabled.png#bordered)

## Use Cursor to edit the Penpot file

### Inspect the file structure

Start by asking Cursor to read the design structure.

1. Keep your Penpot file open and ensure the plugin status is **Connected to MCP server**.

2. In Cursor, start a new chat and ask:

   ```text
   List all frames in the current Penpot file.
   ```

   ![Cursor lists Penpot frames](/images/manual/use-cases/penpot-cursor-list-frames.png#bordered)

:::tip Work one frame at a time
If the file has several frames, name the target frame in your next prompt. This keeps the change focused and makes it easier to review in Penpot.
:::

### Add a card-style component

After Cursor reads the file, ask it to make a concrete design change. The following example adds a reusable card-style component to the selected frame.

1. In Cursor, send a prompt like this. Replace `Home` with the frame you want to modify:

   ```text
   In the current Penpot file, add a card-style component to the Home frame.
   The card should include a title, a short description, and one primary button.
   Match the existing spacing, colors, and typography as closely as possible.
   Name the main group "Feature card" and explain what you changed.
   ```

2. If Cursor asks you to choose from several options, choose the option that best matches your layout.

3. Wait for Cursor to finish the tool calls.

4. Return to Penpot and check the active page. The new design element should appear in the target frame.

   ![Penpot file updated by Cursor](/images/manual/use-cases/penpot-result-card.png#bordered)

5. Select the new element in Penpot and check its layer name, position, text, and visual style.

### Review and refine the result

Treat Cursor's first edit as a draft. Review the result in Penpot, then ask Cursor for specific adjustments.

Use prompts like:

```text
Move the Feature card 24 px below the hero heading and align it with the left edge of the content column.
```

```text
Make the button label shorter and adjust the card width so it matches the other content blocks.
```

```text
Rename the card layers so they are easy for developers to inspect.
```

The workflow is complete when:

- Cursor can list the frames in the connected Penpot file.
- Cursor can explain which frame or layer it changed.
- The new card appears in the selected Penpot frame.
- The card uses clear layer names and fits the surrounding design.

## FAQs

### Cursor cannot see my Penpot file

#### Cause

The MCP plugin is not connected, the Penpot browser tab is closed, or a different file or page is active.

#### Solution

Open the target Penpot file, open the MCP plugin, click **Connect to server**, and wait until the status shows **Connected to MCP server**. Then retry your prompt in Cursor.

### How does the Penpot MCP connection work?

Penpot MCP connects Cursor to the Penpot file currently open in your browser.

| Component | Used by | What it does | Manual setup |
|:----------|:--------|:-------------|:-------------|
| MCP Plugin | Penpot file | Exposes the active file, page, frames, layers, components, styles, and tokens to the MCP server. | Add it in Penpot. |
| MCP HTTP | Cursor | Receives MCP requests from Cursor and forwards them to the connected Penpot file. | Configure it in Cursor. |
| MCP WebSocket | Plugin and MCP server | Keeps the Penpot file and MCP server connected in real time. | No manual setup needed. |

Cursor only needs the MCP HTTP endpoint with `/mcp` appended. The MCP WebSocket connection is used internally between the Penpot plugin and the MCP server.

The plugin must stay connected in Penpot while Cursor works with the file.

## Learn more

- [Penpot Help Center](https://help.penpot.app/): Official Penpot guides and product documentation.
- [Model Context Protocol](https://modelcontextprotocol.io/): Learn how MCP connects AI clients to external tools and data sources.
