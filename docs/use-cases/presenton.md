---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/presenton
outline: [2, 3]
description: Create editable presentation slides with AI Router on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Presenton, self-hosted, presentations, PowerPoint, PDF, AI Router"
app_version: "0.0.8"
doc_version: "1.0"
doc_updated: "2026-10-29"
---

# Create an editable presentation with Presenton

Presenton creates slide decks from a prompt or document. Choose a design, review the generated content and export an editable PowerPoint file or a PDF. This Olares package uses AI Router's default chat model and stores presentation data on your device.

## Learning objectives

- Generate a short presentation from a prompt.
- Review and edit the outline and slides.
- Export the result for further editing or sharing.

## Prerequisites

- An amd64 Olares device running Olares 1.12.7 or later.
- A browser connected to your [Olares private network](../manual/larepass/private-network.md).
- AI Router with a working **default-chat** route.

The package uses a pinned upstream 0.9.12-beta container image. Image generation is disabled in this package, so you can complete the basic workflow without configuring a separate image provider. Templates may use placeholder images; review them before sharing.

## Install Presenton

1. Open Market and install **Presenton**.
2. Open Presenton from Launchpad.
3. Wait for the application to finish its first startup. Create the instance account when prompted, then sign in.

The application is preconfigured to use **default-chat** through AI Router. You do not need to paste an external model key for this workflow. First use may download local indexing assets; later runs reuse the saved cache.

## Create a short deck

On the new-presentation page, select **Standard**, set the slide count to **1**, and choose **English**. Enter a specific brief. For example:

> Create one introductory slide for the Portland Library Community Garden. Volunteers meet on Saturdays at 9 AM. Maya Chen coordinates the group. The garden welcomes new volunteers. Use a short title and readable English text.

Select **Generate presentation**. Review the generated outline, check the slide count and topic order, and choose a suitable built-in template before continuing to the full deck.

![A presentation workflow in Presenton](/images/manual/use-cases/presenton-create.png#bordered)

## Review and edit

Read every slide. Double-click a text element to edit it, then click outside it. Click the presentation title at the top to rename the deck. Correct names, dates and factual claims before exporting. A model can produce a useful first draft, but it may omit details or create text that does not fit the chosen design.

Start with one slide to check that your default model can complete the workflow. Longer decks and complex layouts depend on the model’s ability to return structured content. If generation takes time, check the current task before submitting the same prompt again.

![An edited slide in Presenton](/images/manual/use-cases/presenton-editor.png#bordered)

## Export and keep the result

Select **Export**, then **PPTX** when you want to continue editing in a presentation application, or **PDF** for reading and sharing. Open the exported file and check its layout before sending it to others.

Presentation records use Olares PostgreSQL. Uploaded documents, assets and exported files persist in private application data. Download important exports and include database and application data in your backup plan.

## Learn more

- [Presenton documentation](https://docs.presenton.ai/)
- [Olares Router](./olares-router.md)
