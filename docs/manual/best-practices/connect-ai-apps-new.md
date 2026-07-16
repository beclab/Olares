---
outline: [2, 3]
description: Learn how to connect AI client apps to AI service apps on Olares using the Model Console and application entrances.
---

# Connect AI apps <Badge type="tip" text="^ 1.12.6" />

When you use AI on Olares, you typically work with two separate applications: one AI service app provides AI capabilities in the background, and another AI client app provides the chat interface you interact with directly. To make them work together, you must connect them.

This guide explains how to understand your apps, gather the required connection details, and configure the connection.

## Learning objectives

By the end of this tutorial, you will be able to:
- Distinguish between AI service apps and AI client apps.
- Gather the essential connection parameters.
- Locate the correct endpoint for different types of AI service apps.
- Connect common AI apps.

## Identify your AI apps

Before configuring a connection, determine which app is providing the AI capabilities and which app is consuming them. This helps you know where to look for your network parameters.

- **AI client apps**: They provide the chat interface or workflow canvas you interact with directly, such as LobeHub and Open WebUI. They rely on an AI service app (often called a **provider**) to perform AI tasks, such as generating text and recognizing images.
- **AI service apps**: They provide AI capabilities for compatible clients over an API, such as chat, search, and speech recognition. Some AI service apps have their own web interface for management, while others run primarily as headless backend services.

    On Olares, AI service apps fall into two categories:
    - **LLM service apps**: Apps that host large language models for text generation, code completion, and chat. They include eight pre-built model apps, and custom model instances created on [Engine Base apps](/use-cases/llm-base-apps.md).
    - **Other AI service apps**: Utility apps that provide non-LLM features, such as speech recognition (Speaches) and text extraction (PaddleOCR).

## Gather connection parameters

Most AI client apps require four pieces of information to establish a connection. Missing or mismatching any of these will cause the connection to fail. Gather these details before setting up your client.

### Provider and API format <Badge type="tip" text="LLM services only"/>

In most AI client apps, a **provider** is the service or vendor that supplies the LLM (such as OpenAI, Anthropic, or Ollama). On Olares, your local **LLM service app** acts as this provider. Instead of sending requests to a cloud vendor, your client app sends them to your local LLM service app.

Because different providers use different communication rules, they rely on specific **API formats** (the "language" the apps use to talk to each other). The two most common formats are **OpenAI-Compatible** and **Ollama**.

When configuring an LLM connection, check your client app to see which providers it supports. Then, in the Model Console, you must select the exact **API format** that matches the provider you chose in your client app.

:::info
Non-LLM services like PaddleOCR do not use these generic formats. They communicate using their own tool-specific protocols, so you do not need to configure a provider format for them.
:::

### Base URL

The Base URL is the network address (or endpoint) where the AI service app receives requests.

How you find it depends on the type of the service app:

- **For LLM service apps**: Open the app to launch its **Model Console**. Select the **Connection source** that matches where your client runs, choose the **API format**, and then copy the generated **Base URL**.
- **For other AI service apps**: Open Olares Settings, go to **Applications** > **[AppName]** > **Entrances**, and then copy the **Endpoint URL**. Ensure the entrance's **Authentication level** is set to **Internal** so other apps can access it without a login prompt.

    :::tip Multiple entrances
    Some apps expose more than one entrance. Choose the entrance that matches your client's protocol or use case. For example, use the main entrance for web UI access and a dedicated API entrance for programmatic integrations.
    :::

### Model name <Badge type="tip" text="LLM services only"/>

The model name is the exact identifier of the model. The client sends this ID with every request so the service knows which model file to process.

In the **Model Console**, copy the **Model name** exactly as displayed. Do not remove any repository prefixes (like `unsloth/`) or quantization tags (like `UD-Q4_K_XL`). Otherwise, the client might return a "Model not found" error.

### API key

An API key (also called "Auth Token" or "API Token") is a security credential used to authenticate requests.

For AI service apps deployed locally on Olares, an API key is usually not required. The internal entrance already trusts requests from other apps in the same cluster.

If the client app forces you to enter an API key:

- Try leaving the field blank.
- If the app does not allow an empty value, enter any placeholder string such as `olares`.

## Examples

The following examples focus on how to connect AI service apps to AI client apps. They assume that the relevant apps are already installed and configured.

### Connect an LLM service app to LobeHub

In this example, the pre-built model app Gemma 4 26B (Ollama) is the LLM service app, and LobeHub is the client app.

1. Open Gemma 4 26B (Ollama) from the Launchpad to launch its Model Console.
2. Ensure that the **Model** shows **Ready** and the **Engine** shows **Running**.
3. Select the connection options that match your client app to get the correct Base URL:

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


### Connect PaddleOCR to Open WebUI

In this example, PaddleOCR is the AI service app that provides capabilities, and Open WebUI is the client app.
<!--I will fill in this section tmrw.-->

## FAQs

### How to connect non-AI apps?

The same internal-entrance pattern applies when connecting non-AI apps to each other. For example:
- The *Arrs media stack uses internal entrance URLs to connect Sonarr, Radarr, Prowlarr, Bazarr, and qBittorrent. See [Manage your media library with the *Arrs ecosystem](/use-cases/arrs.md).
- SearXNG can be connected to Vane for private, enhanced search capabilities. See [Connect SearXNG to Vane](/use-cases/perplexica.md).

## Learn more

- [Use cases](../../use-cases/index.md)
- [Host local large language models with Engine Base apps](../../use-cases/llm-base-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
