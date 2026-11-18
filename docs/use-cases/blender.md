---
outline: [2, 3]
description: Install Blender on Olares, open its streamed desktop, and save an editable project to Olares Files.
head:
  - - meta
    - name: keywords
      content: Olares, Blender, 3D creation, streamed desktop, Olares Files, self-hosted
app_version: "0.1.27"
doc_version: "2.0"
doc_updated: "2026-11-07"
---

# Start a Blender project on Olares

Blender is a 3D creation suite for modeling, animation, and rendering. On Olares, it runs as a streamed desktop that you open in a browser while project files remain on your device.

This guide installs Blender, opens the desktop, and saves a starter project to Olares Files. You can then continue in the Blender interface or let Lares work on the same scene through MCP.

## Learning objectives

In this guide, you will learn how to:

- Choose a compute mode and install Blender.
- Open the streamed Blender desktop.
- Save and reopen a project through Olares Files.

## Prerequisites

- Olares 1.12.7 or later.

## Install Blender

1. Open Market and search for "Blender".

   <!-- ![Blender in Market](/images/manual/use-cases/blender.png#bordered) -->

2. Click **Get**, then **Install**.
3. Select a compute mode that matches your device and workload:

   - **Intel:** Use an available Intel integrated GPU for the viewport and desktop stream. This option leaves a discrete GPU available for another app.
   - **CPU:** Use this mode when no suitable GPU is available. Desktop streaming uses more CPU because the stream is software-encoded.
   - **NVIDIA:** Assign an NVIDIA GPU for heavier rendering work. Other apps using the same GPU can reduce performance.

4. Confirm the selection and wait for installation to complete.

You can change the assigned accelerator later through Olares resource settings. Save the current project before restarting Blender with a different allocation.

## Open Blender

1. Open Blender from Launchpad.
2. Wait for the default scene to appear. A new scene normally contains `Camera`, `Cube`, and `Light`.

The browser displays a streamed desktop. Enable LarePass VPN on your computer before opening Blender when you want a lower-latency connection.

:::tip On the same local network?
You can use the Olares `.local` address instead. See [Access Olares services locally](../manual/best-practices/local-access.md) for the address format and Windows setup.
:::

## Save the starter project

1. In Files, create **Home** > **blender-projects**.
2. In Blender, click **File** > **Save As**.
3. In the file browser, open `/config/Home/blender-projects`.
4. Set the filename to `starter-scene.blend` and click **Save As**.
5. If Blender asks for confirmation, click **Save As** again.
6. Open Files and go to **Home** > **blender-projects**. Confirm that `starter-scene.blend` appears.

Blender's `/config/Home` path corresponds to **Home** in Files. Save projects and outputs there when you want to preview, download, or reuse them outside the Blender desktop.

## Reopen a saved project

1. Open Blender from Launchpad.
2. Click **File** > **Open**.
3. Open `/config/Home/blender-projects/starter-scene.blend`.
4. Continue editing, then click **File** > **Save** to update the same project.

## Learn more

- [Create a Blender scene with Lares](blender-lares.md): Connect Blender through MCP and build a rendered solar-system illustration from a conversation.
- [Access Olares services locally](../manual/best-practices/local-access.md): Reduce streamed-desktop latency on your local network.
- [Manage accelerator resources](../manual/olares/settings/gpu-resource.md): Change the GPU resources assigned to Blender.
