---
outline: [2, 3]
description: Use Excalidraw on Olares to create hand-drawn style diagrams, wireframes, and sketches in a self-hosted virtual whiteboard.
head:
  - - meta
    - name: keywords
      content: Olares, Excalidraw, whiteboard, diagrams, wireframes, hand-drawn, sketching, self-hosted
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

# Create hand-drawn diagrams with Excalidraw

Excalidraw is an open-source virtual whiteboard with a hand-drawn aesthetic. You can use it to create diagrams, wireframes, flowcharts, or any freeform sketches directly on your Olares device.

## Install Excalidraw

1. Open Market and search for "Excalidraw".
   ![Install Excalidraw](/images/manual/use-cases/excalidraw.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Use Excalidraw

Open Excalidraw from Launchpad to access the whiteboard canvas.

![Excalidraw canvas](/images/manual/use-cases/excalidraw-canvas.png#bordered)

You can also click <i class="material-symbols-outlined">open_in_new</i> to open Excalidraw in a new browser tab.

### Create diagrams

1. Select a shape from the toolbar (rectangle, ellipse, arrow, line, etc.).
2. Customize the shape's stroke color, stroke width, stroke style, background pattern, and opacity from the style panel.
3. Click and drag on the canvas to draw the shape.

    ![Drawing on canvas](/images/manual/use-cases/excalidraw-drawing.png#bordered)

4. Select the text tool and click on the canvas to add text.

    ![Adding text](/images/manual/use-cases/excalidraw-text.png#bordered)

### Add Excalidraw libraries
Excalidraw libraries are sets of reusable graphical elements. Instead of drawing common elements like servers, databases, or user icons from scratch, you can drag and drop them from an imported library.

1. In the Excalidraw editor, click <span class="material-symbols-outlined">dock_to_left</span> in the top-right corner to open the sidebar.

2. In the library sidebar, click **Browse libraries** to open the official Excalidraw Libraries website.
    ![Browse libraries](/images/manual/use-cases/excalidraw-browse-libraries.png#bordered)

3. Search for a library you need, then click **Add to Excalidraw**.
    ![Add to Excalidraw](/images/manual/use-cases/excalidraw-add-library.png#bordered)

4. Back in the editor, the imported library appears in the sidebar. Drag any element from it onto your canvas.
    ![Imported library](/images/manual/use-cases/excalidraw-imported-library.png#bordered)

### Save your work

Excalidraw supports saving your canvas locally as an `.excalidraw` file or exporting it as an image.

- **Save to local**: Click <span class="material-symbols-outlined">menu</span> in the top-left, and then select **Save to** > **Save to disk** to save the canvas as an `.excalidraw` file that you can reopen later.

    ![Save to disk](/images/manual/use-cases/excalidraw-save-to-disk.png#bordered)

- **Export as image**: Click <span class="material-symbols-outlined">menu</span> in the top-left, and then **Export image** to save the canvas as a PNG or SVG file, or copy to clipboard.

    ![Export as image](/images/manual/use-cases/excalidraw-export-image.png#bordered)

## Known issues

### Collaboration and sharing not supported

The self-hosted version of Excalidraw does not support real-time collaboration or sharing links. The official self-hosted image includes only the frontend client and cannot connect to Excalidraw's cloud backend for collaboration and link sharing. The Excalidraw team plans to provide a fully self-hostable backend in the future. For details, see [excalidraw#1772](https://github.com/excalidraw/excalidraw/issues/1772) and [excalidraw#8195](https://github.com/excalidraw/excalidraw/issues/8195).

## Learn more

- [Excalidraw documentation](https://docs.excalidraw.com): Official Excalidraw docs and guides.
