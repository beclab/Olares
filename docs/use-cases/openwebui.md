---
outline: [2, 4]
description: Self-host Open WebUI on Olares as a private ChatGPT alternative. Connect Ollama to chat with local LLMs while keeping conversations on your device.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, self-hosted chatgpt, open webui alternative, local LLM, Ollama, open webui on olares
app_version: "1.0.29"
doc_version: "2.0"
doc_updated: "2026-05-13"
---

# Set up Open WebUI for local AI chat

Open WebUI provides a user-friendly chat interface for local models on your Olares device. It does not include any models by default, so you need to connect it to a model backend.

This guide covers two configuration options: using Ollama to pull models directly from the Ollama Registry, or using a dedicated pre-configured model app such as Qwen3.5 27B Q4_K_M (Ollama).

## Learning objectives

In this guide, you will learn how to:

- Install Open WebUI On Olares.
- Create an admin account.
- Configure a model backend using Ollama or a dedicated model app.
- Connect the model to Open WebUI and start a chat session.

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Install Open WebUI

1. Open Market and search for "Open WebUI".

   ![Open WebUI](/images/one/open-webui.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Create an admin account

The first time you launch Open WebUI, you need to create a local administrator account to manage your models and settings.

1. Open Open WebUI from the Launchpad.
2. On the welcome page, click **Get started**.
3. Enter your name, email, and password to create the admin account.

   ![Create account](/images/one/open-webui-create-account.png#bordered)

## Configure model backend

Open WebUI requires a backend model to generate responses. Choose one of the following options to configure your model backend.

:::tip Recommendation for multiple models
While Ollama (Option A) offers flexibility, hosting multiple models simultaneously within a single Ollama instance might lead to resource scheduling conflicts.

For optimal performance and stability when using multiple models, install independent model apps (Option B) instead of using the general Ollama application.
:::

### Option A: Use Ollama

Use Ollama to pull and switch between different models from the Ollama Registry for greater flexibility.

#### Install Ollama

1. Open Market and search for "Ollama".
   ![Ollama](/images/manual/use-cases/ollama.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

#### Download a model

With Ollama installed, you can pull models directly through the Open WebUI interface.

:::tip Browse models first
Visit [Ollama Library](https://ollama.com) to browse available models and get the exact model name before downloading. Model names must match exactly to pull successfully.
:::

<Tabs>
<template #From-Open-WebUI-homepage>

1. Open Open WebUI.
2. Click the model dropdown at the top of the chat page, and then enter the model name in the search field. For example: `llama3.2`.
3. Click **Pull "llama3.2" from Ollama.com**. The download starts automatically.

   ![Download from homepage](/images/one/open-webui-download-from-homepage.png#bordered)

</template>
<template #From-Open-WebUI-Settings>

1. Open Open WebUI.
2. Click your profile icon and select **Admin Panel**.
3. Go to **Settings** > **Models**.
4. Click **Manage** in the top right to open the **Manage Models** window.
5. Under **Pull a model from Ollama.com**, enter the model name. For example: `llama3.2`.

   ![Download from settings](/images/one/open-webui-download-from-settings1.png#bordered)

6. Click <i class="material-symbols-outlined">download</i>. The download starts automatically.

</template>
</Tabs>

:::tip Download time
Models range from 2 GB to 20+ GB. Download time depends on your network speed.
:::

#### Verify the connection

Open WebUI automatically detects and connects to your local Ollama installation. The connection is successful when your downloaded model appears in the model drop-down list at the top of the chat page.

### Option B: Use model apps

A model app packages a specific model with pre-configured settings. This option is best if you want a ready-to-use model without managing the Ollama Registry.

#### Install the model app

1. Open Market and search for your desired model.
2. Click **Get**, and then click **Install**. Wait for the installation to finish.

   ![Install model app](/images/one/qwen3.5-27b.png#bordered)

#### Download the model

Open the model app you just installed. The model download starts automatically.

![Downloading model](/images/one/qwen3.5-27b-downloading.png#bordered)

When you see the completion screen, the model is ready.

![Model downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)

#### Obtain model app endpoint

To allow Open WebUI to communicate with this specific model, you need to obtain its shared endpoint URL.

1. Open Olares Settings, and then go to **Applications** > **[Model App Name]**.
2. In **Shared entrances**, select the model to view its endpoint URL.

   ![Get shared endpoint](/images/one/qwen3.5-27b-shared-entrance.png#bordered){width=70%}

3. Copy the endpoint URL. For example, `http://94a553e00.shared.olares.com`.

:::tip Why not use the URL shown on the model page?
The URL shown on the model app page is user-specific and relies on browser-based frontend calls. If your device and Olares are not on the same local network, those calls might trigger Olares sign-in and you might encounter cross-origin restrictions (CORS). To avoid these issues, use the shared endpoint URL.
:::

#### Connect the model app to Open WebUI

Now, return to Open WebUI to link the model using the endpoint URL you just copied.

1. In Open WebUI, click your profile icon and select **Admin Panel**.
2. Go to **Settings** > **Connections**.
3. On the right of **Manage Ollama API Connections**, click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **URL** field, paste the model app's endpoint URL you copied earlier.
5. Click **Save**. 

   Open WebUI verifies the connection automatically. When you see the "Ollama API settings updated" message, the connection is established.

   ![Connection established](/images/one/open-webui-connection-established.png#bordered)

## Start chatting

Once you connect a model, you are ready to use the chat interface.

1. In the model drop-down list, select your configured model.

   ![Select model](/images/one/open-webui-qwen3.5-27b.png#bordered)

2. Enter your prompt in the text box, and then press **Enter** to start your conversation.

   ![Chat with LLM](/images/one/open-webui-chat1.png#bordered)

## Learn more

- [Set up multi-user access](openwebui-multiuser.md): Share Open WebUI with other users on your Olares device.
- [Configure audio](openwebui-audio.md): Enable speech-to-text and text-to-speech.
- [Enable web search](openwebui-search.md): Add web search capabilities to your chats.
- [Use knowledge base](openwebui-knowledge.md): Upload documents and create a knowledge base for RAG.
