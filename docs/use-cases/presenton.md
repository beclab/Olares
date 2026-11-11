---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/presenton
outline: [2, 3]
description: Create and review slides with AI Router, upload text references through an assistant, and export PowerPoint or PDF on Olares
head:
  - - meta
    - name: keywords
      content: "Olares, Presenton, self-hosted, presentations, PowerPoint, PDF, AI Router, MCP, Lares, Hermes Agent"
app_version: "0.0.12"
doc_version: "1.2"
doc_updated: "2026-11-01"
---

# Create an editable presentation with Presenton

Presenton creates slide decks from a prompt or document. Choose a design, review the generated content and export an editable PowerPoint file or a PDF. This Olares package uses AI Router's default chat model and stores presentation data on your device.

## Learning objectives

- Generate a short presentation from a prompt.
- Review and edit the outline and slides.
- Export the result for further editing or sharing.
- Use a connected assistant to browse templates and upload a text reference.

## Prerequisites

- An amd64 Olares device running Olares 1.12.7 or later.
- A browser connected to your [Olares private network](../manual/larepass/private-network.md).
- AI Router with a working **default-chat** route.

The package uses a pinned upstream 0.9.12-beta container image. Image generation is disabled in this package, so you can complete the basic workflow without configuring a separate image provider. Templates may use placeholder images. Review them before sharing.

## Install Presenton

1. Open Market and search for "Presenton". Click **Get**, then **Install**, and wait for installation to complete.
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

Compare visible slide text with your source, especially names, dates, amounts and links. Check any numbers supplied by the template. An assistant can complete the generation task while the slides still contain unrelated placeholder metrics or omit source details. Fix these in the editor before sharing.

Start with one slide to check that your default model can complete the workflow. Longer decks and complex layouts depend on the model’s ability to return structured content. If generation takes time, check the current task before submitting the same prompt again.

![An edited slide in Presenton](/images/manual/use-cases/presenton-editor.png#bordered)

## Export and keep the result

Select **Export**, then **PPTX** when you want to continue editing in a presentation application, or **PDF** for reading and sharing. Open the exported file and check its layout before sending it to others.

Presentation records use Olares PostgreSQL. Uploaded documents, assets and exported files persist in private application data. Download important exports and include database and application data in your backup plan.

## Prepare references with an assistant

Presenton has a native MCP service that lets a connected assistant browse templates, upload reference files and start presentation tasks. Lares and Hermes Agent can use this service. Presenton remains the slide editor, while the assistant handles the brief and reference material.

### Connect the assistant

Use Presenton package 0.0.10 or later for the separate internal MCP entrance. To get the address, run the following command in a terminal on a computer with Olares CLI configured for your device:

```bash
olares-cli settings apps get presenton -o json
```

In `entrances`, find the entry named `mcp`. Use `https://` followed by its `url`, then append `/mcp`. Use this address rather than the main application URL.

Sign in to Presenton as the instance administrator. Open `/admin` on the Presenton web entrance and select **API keys**. Choose the account, enter a label and an expiry, then create and copy its key. Configure the assistant's MCP connection with these values:

| Setting | Value |
| --- | --- |
| Transport | Streamable HTTP |
| Server URL | The internal MCP URL supplied for your Presenton instance, ending in `/mcp` |
| Request header | `Authorization: Bearer <presenton-api-key>` |

Store the key in the assistant's private connection settings. The key belongs to Presenton, rather than AI Router. When connecting from your computer, enable [LarePass VPN](../manual/larepass/private-network.md). The internal entrance is not a public sharing address. A redirect to Olares web sign-in means the connection is using a web entrance instead of the intended MCP entrance.

After connecting, ask the assistant:

> List the built-in Presenton templates. Show their names without starting a presentation.

The assistant should call `list_templates`. Choose from the returned templates rather than asking it to invent a template identifier.

### Upload a text reference

Save the following fictional example as `garden-brief.txt`, or replace it with your own short source note:

```text
Portland Library Community Garden
Coordinator: Maya Chen
Volunteer meeting: every Saturday at 9 AM
New volunteers are welcome.
```

Give the assistant access to the actual file and ask:

> Upload garden-brief.txt to Presenton as reference material. Report whether the upload succeeded. Do not generate a presentation yet.

The assistant uses `upload_files` and receives Presenton-owned file paths. These paths can be passed unchanged to a later presentation task. If the assistant can see only a filename or attachment identifier, it needs access to the file contents before it can upload the document.

TXT references can be read by Presenton's document processing service. Review the generated slides against your original note when you subsequently use the reference to create a deck.

:::info PDF references
Uploading a PDF and extracting its contents are separate operations. PDF upload is available, but document extraction has not completed reliably in this package. For now, copy the relevant text into a TXT file and retain the original PDF for your own source review.
:::

### Review the assistant's result

For a presentation task, the assistant should keep checking the same task until it completes or reports an error. Use the returned editor link to review the deck, then export the reviewed result. Avoid starting another task just because the first one is still pending.

An uploaded reference is preparation for slide generation. It is not a completed presentation or a confirmation that the final slides retain every source fact.

## Learn more

- [Presenton documentation](https://docs.presenton.ai/)
- [Olares Router](./olares-router.md)
- [Lares](./lares.md): Work with an assistant on Olares.
- [Hermes Agent](./hermes.md): Use an agent with persistent conversations and tools.
