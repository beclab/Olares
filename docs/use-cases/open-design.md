---
outline: [2, 3]
description: Use Open Design on Olares to turn prompts into HTML prototypes, landing pages, and slide decks with a local or cloud AI model.
head:
  - - meta
    - name: keywords
      content: Olares, Open Design, AI design studio, AI prototyping, landing page generator, presentation generator, OpenAI-compatible, local LLM, Qwen3.6, self-hosted
app_version: "0.22.1"
doc_version: "1.0"
doc_updated: "2026-09-18"
---

# Create design files with Open Design

Open Design is an open-source AI design studio that turns natural-language briefs into working design files. Its built-in OpenCode agent uses a model provider you configure to create previewable HTML prototypes, landing pages, dashboards, wireframes, and slide decks.

On Olares, you can connect Open Design to a local model through Model Console or bring an API key for a cloud provider. Projects keep their prompts, references, previews, and generated files together so you can iterate on a design before exporting it.

:::warning Current output limitation
Open Design 0.22.1 on Olares does not generate standalone images from text. Use it to create pages, prototypes, dashboards, or slide decks instead.
:::

## Learning objectives

In this guide, you will learn how to:

- Install Open Design on Olares.
- Connect a local model through its OpenAI-compatible API.
- Create, refine, and export a design project.
- Troubleshoot common model connection and generation issues.

## Prerequisites

- Olares 1.12.6 or later.
- Qwen3.6-27B (llama.cpp) installed from Market and ready in Model Console. This guide uses it as the local model.

To deploy a different local model, see [Host local large language models with Engine Base apps](llm-base-apps.md).

## Install Open Design

1. Open Market and search for "Open Design".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Connect a local model

### Get the model connection details

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

In this guide, Open Design connects to Qwen3.6-27B (llama.cpp) using the **OpenAI-Compatible** API format:

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Add the model to Open Design

1. Open Open Design from Launchpad.
2. Open **Settings**. On the **Models & providers** page, select **API provider**.
3. Under **Provider preset**, select **Custom provider**.
4. Configure the provider:

   - **Base URL**: Paste the Base URL copied from Model Console, including the trailing `/v1`.
   - **API key**: Enter any non-empty value, such as `olares`. Local model apps do not require a real key for requests from another app in the same Olares cluster.
   - **Model**: Enter the exact Model name copied from Model Console.
   - **Max tokens (optional)**: Enter `65536`. You can use `32768` for a smaller context window.

   <!-- ![Configure an API provider in Open Design](/images/manual/use-cases/open-design-api-provider.png#bordered) -->

5. Click **Test** to verify the connection. A green **Connected** status confirms that the connection is established. Changes are saved automatically.

:::tip Start with a small test
Ask Open Design to create a single title slide before starting a multi-page deck or landing page. A successful preview confirms that the Base URL, model name, and token limit work together.
:::

## Create a design project

1. On the Home page, select a project type, such as **Prototype**, **Slide deck**, **Document**, or **Website clone**.
2. Choose a Skill preset that matches your task, such as **Blog Post**. You can also open one of the examples below the prompt to use it as a starting point.
3. Optionally, select a **Design system** and **Working directory**.

   <!-- ![Select a project type, Skill, and design system](/images/manual/use-cases/open-design-skill-design-system.png#bordered) -->

4. Enter a specific brief. Include the intended audience, content, visual direction, and output format. For example:

   ```text
   Create a title slide for a product launch presentation.
   Use a dark background, a bold geometric title, and a teal accent.
   Include the product name "Northstar" and the subtitle "Find your next move."
   ```

5. Check that the intended model is selected, then click the arrow to start generating.
6. Wait for the workspace to open and the preview to render.

   <!-- ![Open Design workspace and preview](/images/manual/use-cases/open-design-workspace-preview.png#bordered) -->

7. Refine the result in the conversation. You can drag in a reference image or type `@` to attach a project file, then describe the part you want to change.
8. When the result is ready, open the export menu and download it as ZIP, PDF, or PPTX, depending on the project type.

   <!-- ![Export a project from Open Design](/images/manual/use-cases/open-design-export.png#bordered) -->

## Optional: Connect a cloud model

You can use a cloud provider instead of the local model. Open **Settings**, select **API provider** on the **Models & providers** page, then choose a provider preset or configure a custom provider with the matching protocol, Base URL, API key, and model.

| Provider | Protocol | Base URL |
|:---------|:---------|:---------|
| OpenAI | OpenAI | `https://api.openai.com/v1` |
| OpenRouter | OpenAI | `https://openrouter.ai/api/v1` |
| DeepSeek | OpenAI | `https://api.deepseek.com` |
| Anthropic | Anthropic | `https://api.anthropic.com` |
| Qwen | OpenAI | `https://dashscope.aliyuncs.com/compatible-mode/v1` |
| SiliconFlow | OpenAI | `https://api.siliconflow.cn/v1` |
| Kimi | OpenAI | `https://api.moonshot.cn/v1` |

Cloud usage is billed by the selected provider. Use that provider's exact model name and a valid API key.

## Troubleshooting

### The connection test fails or returns 404

Paste the Base URL from Model Console again without changing its path. For the local model in this guide, select **OpenAI-Compatible** in Model Console and use the URL exactly as displayed, including `/v1`.

### Generation stops before the design is complete

Set **Max tokens** to `65536` or `32768`. Multi-page slide decks need more output tokens than a single page or cover slide.

### Open Design does not respond

Return to the model app's Model Console. Continue only when **Model** shows **READY** and **Engine** shows **RUNNING**.

### Open Design uses the wrong model

Compare the **Model** value in Open Design with the **Model name** in Model Console. The values must match exactly.

### Open Design does not create a standalone image

This is a limitation of Open Design 0.22.1 on Olares. Select a Skill for a page, slide deck, dashboard, wireframe, or interactive prototype instead.

## Learn more

- [Host local large language models with Engine Base apps](llm-base-apps.md): Deploy and manage models through Model Console.
- [Connect AI apps to model services](/manual/best-practices/connect-ai-apps.md): Understand model names, API formats, and Base URLs.
- [Open Design website](https://open-design.ai/): Explore the upstream project and its capabilities.
