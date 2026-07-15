---
outline: [2, 3]
description: Learn how to connect AI client apps to AI service apps on Olares using the Model Console and application entrances.
---

# Connect AI apps

Many AI applications on Olares follow a standard connection pattern: one AI service app provides AI capabilities over an API, and another AI client app provides the chat interface you interact with every day. Once you understand this pattern, you can connect almost any compatible combination of apps.

This tutorial explains the core concepts and walks you through a few practical examples.

:::tip For Olares 1.12.5 and earlier
This guide applies to Olares v1.12.6 and later. The way AI apps connect changed in v1.12.6 with the introduction of the Model Console. If you are on Olares 1.12.5 or earlier, refer to the [legacy AI apps connection guide](./connect-ai-apps.md).
:::

## Learning objectives

By the end of this tutorial, you will be able to:
- Distinguish between AI service apps and AI client apps.
- Understand the types of AI service apps available in Olares.
- Identify the correct endpoint source for different types of AI service apps.
- Connect common AI apps.

## Core concepts

### AI service apps vs. AI client apps

- **AI service apps**: They provide AI capabilities for compatible clients over an API, such as chat, search, and speech recognition. Some AI service apps have their own web interface for management, while others run primarily as headless backend services.
- **AI client apps**: They provide the chat interface or workflow canvas you interact with directly, but they rely on an AI service app to perform AI tasks such as generating text, searching the web, or recognizing text. For example, LobeHub (formerly LobeChat) and Open WebUI.

### Types of AI service apps

In Olares, AI service apps mainly fall into the following categories:

- **LLM service apps**: Apps that host large language models for text generation, code completion, and chat. They include:
    - **[Engine Base apps](/use-cases/llm-base-apps.md)**: Four reusable base applications (Ollama Engine Base, vLLM Engine Base, SGLang Engine Base, and llama.cpp Engine Base). You clone them into independent model instances and configure each instance yourself.
    - **Pre-built model apps**: Nine ready-to-use apps that package a specific model with a specific engine, such as Qwen3.6-27B (llama.cpp) and Gemma 4 26B (Ollama).
- **Other AI service apps**: Apps that provide other AI capabilities beyond LLMs, such as SearXNG (search) and PaddleOCR (OCR).

### How AI service apps expose endpoints

An endpoint is the URL through which an application's entrance can be reached. In Olares, AI service apps expose their endpoints through two different paths, depending on the type of capability they provide:

| Service type | Endpoint location | Description | Examples |
| :--- | :--- | :--- | :--- |
| **LLM services** | Model Console | Provides dynamic, network-optimized<br>APIs depending on whether your client<br>is inside Olares, on your local network,<br>or remote.<br><br>Open the app to launch the Model Console, and then get the **Base URL**. | <ul><li>Qwen3.6-27B<br>(llama.cpp)</li><li>Gemma 4 26B (Ollama)</li></ul> |
| **Other AI services** | Application Settings | Uses standard HTTPS endpoints.<br><br>Open Olares **Settings**, go to **Applications** > **[AppName]** > **Entrances**, and then copy the **Endpoint URL**. | <ul><li>SearXNG</li><li>PaddleOCR</li></ul> |

### Authentication levels

Olares provides the following access levels for application entrances:

- **Internal (recommended)**: Allows apps to communicate without login prompts. It also allows access via your local network or via LarePass VPN.
- **Public**: Open to anyone on the internet. Not recommended for private services.

How you apply these access levels depends on the type of AI service app:
- For LLM service apps, the Model Console handles access control through the **Connection source** you select.
- For other AI service apps whose endpoints come from **Settings > Applications > [App Name] > Entrances**, set the entrance's **Authentication level** to **Internal** before connecting a client.

:::info Non-AI apps use the same pattern
The same internal-entrance pattern also applies when connecting non-AI apps to each other. For example, the *Arrs media stack uses internal entrance URLs to connect Sonarr, Radarr, Prowlarr, Bazarr, and qBittorrent. See [Manage your media library with the *Arrs ecosystem](/use-cases/arrs.md).
:::

## Examples

The following examples focus on how to connect AI service apps to AI client apps. They assume that the relevant apps are already installed and configured.

### Connect an LLM service app to LobeHub

In this example, the pre-built model app Gemma 4 26B (Ollama) is the LLM service app, and LobeHub is the client app.

1. Open Gemma 4 26B (Ollama) from the Launchpad to launch its Model Console.
2. Ensure that the **Model** shows **Ready** and the **Engine** shows **Running**.
3. Specify the following settings:

    - **Connection source**: Select **Apps in Olares**, because LobeHub is installed directly in the Olares cluster.
    - **API format**: Select **Ollama**, because LobeHub's Ollama provider expects this format.

4. Note down the following details:

    - **Model name**: `gemma4:26b`
    - **Base URL**: `https://74bfa5ee.alice2026.olares.com`

    ![LobeHub connected to a model instance](/images/manual/tutorials/connect-app-exp-model-console.png#bordered)

5. Open **LobeHub**, and then go to **Settings** > **AI Service Provider** > **Ollama**.
6. In the **Interface proxy address** field, paste the **Base URL** you copied.
7. Ensure **Use Client Request Mode** is disabled.

    :::warning Disable Client Request Mode
    Do not enable **Use Client Request Mode** in LobeHub. Enabling this forces the application to make frontend browser calls, which can trigger cross-origin (CORS) blocks or Olares security authentication prompts. Keeping it disabled ensures secure, direct backend-to-backend communication.
    :::

8. Next to **Model List**, click **Fetch models**. The model name `gemma4:26b` appears in the list.
9. Toggle on `gemma4:26b` to enable it.
10. On the right of **Connectivity Check**, select the model name from the drop-down list, and then click **Check**. When **Check Passed** appears, the connection is established.

    ![LobeHub connected to a model instance](/images/manual/tutorials/connect-app-eg-lobehub2.png#bordered)

### Connect SearXNG to Vane

In this example, SearXNG is the AI service app that provides web search capabilities, and Vane (formerly Perplexica) is the client app.

1. Open Olares Settings, and then go to **Applications** > **SearXNG** > **Entrances** > **SearXNG**.
2. In the **Access policies** section, ensure the **Authentication level** is set to **Internal**.
3. In the **Endpoint settings** section, copy the endpoint URL, such as `https://84a93c3c.alice2026.olares.com`.

    ![SearXNG endpoint in Settings](/images/manual/tutorials/connect-apps-searxng-endpoint.png#bordered){width=70%}

4. On the Vane home page, click <i class="material-symbols-outlined">settings</i> in the lower-left corner, and then select **Search**.
5. Enter the SearXNG internal endpoint URL you copied, such as `https://84a93c3c.alice2026.olares.com`.

    ![SearXNG settings in Vane](/images/manual/tutorials/connect-apps-searxng-vane.png#bordered)

6. In the Vane chat box, ask a question that requires web search, such as `What is the latest news about OpenAI?`.

    If SearXNG is connected correctly, Vane shows **Sources** and **Found X results**, and then returns an answer with web search citations.

    ![Vane answering with SearXNG search results](/images/manual/tutorials/connect-apps-searxng-vane-working.png#bordered)

## Learn more

- [Host local large language models with Engine Base apps](../../use-cases/llm-base-apps.md)
- [Shared applications](../olares/market/shared-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
- [Network](../../developer/concepts/network.md)
