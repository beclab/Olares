---
outline: [2, 3]
description: Create a short video with Concat and Lares on Olares. Connect through MCP, turn your images into an edit, and find the exported MP4 in Files.
head:
  - - meta
    - name: keywords
      content: Olares, Concat, Lares, MCP, video editing, MP4, self-hosted
app_version: "0.2.5"
doc_version: "1.1"
doc_updated: "2026-09-24"
---

# Create videos with Concat and Lares

Concat is a video editor for arranging clips, adding titles and transitions, and exporting MP4 videos. On Olares, you can use its editor in a browser or let an AI assistant edit through Model Context Protocol (MCP).

This example uses Lares to turn a few images into a short introduction to Olares. Your source media, saved projects, and exported videos stay in the Concat folder in Files.

## Prerequisites

- Olares 1.12.7 or later on an amd64 device. This Concat package uses CPU rendering and does not require a GPU.
- [Lares](lares.md) configured with a model that can call tools, installed on the same Olares as Concat.
- Three or four images for your video.

## Install Concat

1. Open Market and search for "Concat".
   <!-- ![Concat in Market](/images/manual/use-cases/concat.png#bordered) -->

2. Click **Get**, then **Install**, and wait for installation to complete.

## Prepare your media

1. Open Files and go to **Home** > **Documents** > **Concat**.
2. Create an `assets/my-demo` folder and upload your images. Add an audio file to the same folder to include background music.
3. Create an `outputs/my-demo` folder for the finished video.

These subfolders are suggestions, not folders created automatically during installation.

| Location in Files | Path used by Concat MCP |
|:---|:---|
| `Home/Documents/Concat/assets/my-demo` | `assets/my-demo` |
| `Home/Documents/Concat/outputs/my-demo` | `outputs/my-demo` |

The browser editor sees `Home/Documents/Concat` as `/config/Projects`. Upload your source media to this folder before asking Lares to edit it.

## Configure Concat in Lares

### Get the MCP endpoint

1. Open Settings and go to **Applications** > **Concat** > **Entrances**.
2. Select **Concat API** and copy its **Endpoint** URL. Use this entrance for MCP connections. The **Concat** entrance opens the browser editor.
3. Check that the URL ends in `/mcp`. If it does not, append `/mcp` without duplicating the slash.

The MCP entrance is internal by default and is accessible to Lares on the same Olares device.

### Get the application token

The token is generated when Concat is installed. It is not your Olares password. Use either method below with an account that can access the app's data or ask your administrator for the token.

<tabs>
<template #Files>

1. Open Files and find Concat's application data under **Application** > **Data** > **concat** > **config**.
2. Open the `api-token` file as text. If text preview is unavailable, download it and open it in a text editor.
3. Copy the complete line. Use this token directly without decoding it.

The file is in `Data/concat/config/api-token`, not in `Home/Documents/Concat`. Do not edit or rename it.

</template>
<template #Control-Hub>

1. Open Control Hub and click **Browse**.
2. Expand the namespace for your Concat installation, normally `concat-<username>`.
3. Expand **Secrets** and select **concat-api**.
4. Under **Data**, find **token**. Click the visibility button to decode the value, then copy the complete plain token.

Use the decoded value, not the Base64 text from `data.token`. See [View Secrets](../manual/olares/controlhub/manage-resource.md#view-secrets) for the Control Hub interface.

</template>
</tabs>

Keep the token private. Paste it into the MCP configuration, not into a chat prompt or a shared screenshot.

### Add the MCP server

1. Open Lares and go to **Settings** > **MCP**.
2. Click **Add server** and enter these values:

   | Setting | Value |
   |:---|:---|
   | **Server name** | `concat` |
   | **Transport** | `Streamable HTTP` |
   | **MCP URL** | The URL ending in `/mcp` prepared above |
   | **Headers** | The JSON object below |

   In **Headers**, enter the following JSON. Replace `<your-token>` with the token you copied and keep one space after `Bearer`:

   ```json
   {
     "Authorization": "Bearer <your-token>"
   }
   ```

3. Click **Save**. Lares connects to the server and loads its tools automatically.
4. In a new conversation, ask:

   ```text
   Use Concat MCP to check its status and read its help. List the files in
   assets/my-demo. Do not create or change any projects yet.
   ```

Lares should call `concat_status`, `concat_help`, and `concat_list_files` and report your uploaded images. Concat exposes seven MCP tools. Seeing their names alone does not confirm that calls work.

:::tip Other assistants
The same endpoint and Authorization header work with compatible remote HTTP MCP clients, including OpenCode. For network access from a client outside Olares, see [Access Olares services locally](../manual/best-practices/local-access.md). This endpoint handles MCP POST requests, not a legacy SSE connection or an editor page.
:::

## Make a short video

1. Save any changes in the browser editor before asking Lares to edit. MCP editing temporarily closes the editor, which returns when the session finishes.
2. Send this prompt:

   ```text
   Use Concat MCP to make a polished, 12-second introduction to Olares from
   the images in assets/my-demo. Make it a horizontal 1080p video.

   Keep the style clean and bright. Show these three titles in order:
   "Your digital space", "Create your way", and "Start with Olares".
   Use slow, gentle movement, soft transitions, and enough space around
   the text. Add quiet background music only if an audio file is available.

   Read the tool help first. Create a new project with a unique name and
   leave existing projects unchanged. Check a few preview frames for
   cropped or overlapping text before exporting.

   Start with concat_begin_edit. Export an MP4 to outputs/my-demo with
   a unique filename, wait for the render to finish, then call
   concat_finish_edit. Tell me the full location of the video in Files.
   ```

3. Wait for Lares to report that the export succeeded. A render job ID confirms submission. Wait for the completed export before opening the MP4.
4. In Files, open **Home** > **Documents** > **Concat** > **outputs** > **my-demo** and play the returned MP4. Check the titles, transitions, and sound.

For another revision, describe what to change and ask for a new output filename. Concat rejects overwriting an existing output. The result depends on your images and model. Review the previews before exporting a longer video.

## Edit in the browser

You can also work directly in the editor:

1. Open Concat from Launchpad and create or open a project under `/config/Projects`.
2. Click **Import**, select your media, and arrange the clips on the timeline.
3. Add titles, transitions, and audio, then check the preview. Select **Full** preview quality to inspect small text.
4. Click **Export**, choose the destination and video settings, and wait for **Exported**.

![Concat editor with multiple tracks and a video preview](/images/manual/use-cases/concat-editor.png#bordered)

Saving a project does not export an MP4. Manual exports appear at the destination you choose, so select a folder under `/config/Projects` to find them in Files.

## FAQs

### Why does the MCP connection fail?

A login page or HTML response points to the Olares entrance or network policy. Check that you copied the **Concat API** endpoint and that your client can reach it. A Concat `401` response means the application token is missing or invalid. Check the plain token and the space after `Bearer`.

Opening `/mcp` directly in a browser sends a GET request and is not a connection test. Use the read-only prompt above.

### Why did the editor close while Lares was working?

Concat switches between browser editing and MCP editing. Wait until all rendering jobs finish, then ask Lares to call `concat_finish_edit`. Do not edit manually during an MCP session.

### Why are Chinese titles missing characters?

Ask the assistant to use `Noto Sans CJK SC`, then generate another preview. This font is included in the Olares package.

## Learn more

- [Lares](lares.md): Configure the assistant used in this example.
- [OpenCode](opencode.md): Use a coding assistant on Olares.
