---
outline: [2, 3]
description: Install Open WebUI on Olares and connect it to a local model backend. Use Ollama to pull models from the Ollama Registry, or use a pre-configured model app.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, local LLM, AI chatbot, Ollama
app_version: "1.0.20"
doc_version: "2.0"
doc_updated: "2026-05-13"
---

# Set up Open WebUI for local AI chat

Open WebUI provides a user-friendly chat interface for local models on your Olares device. It does not include any models by default, so you need to connect it to a model backend. This guide covers two options: Ollama for pulling models from the Ollama Registry, or a dedicated model app such as Qwen3.5 27B Q4_K_M.

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Install Open WebUI

1. Open Market and search for "Open WebUI".
   ![Open WebUI](/images/one/open-webui.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Create an admin account

1. Open the Open WebUI app.
2. On the welcome page, click **Get started**.
   ![Create account](/images/one/open-webui-create-account.png#bordered)
3. Enter your name, email, and password to create the account.

   :::info First account is admin
   The first account created has full administrator privileges for managing models and settings.
   :::

   :::info Local account only
   This account is stored locally on your Olares device and does not connect to external services.
   :::

## Configure model backend

Choose one of the following options to configure your model backend.

### Option A: Ollama (recommended)

Ollama lets you pull and switch between different models from the Ollama Registry, offering more flexibility.

#### Install Ollama

1. Open Market and search for "Ollama".
   ![Ollama](/images/manual/use-cases/ollama.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

3. Open Open WebUI.

#### Download a model

:::tip Browse models first
Visit [Ollama Library](https://ollama.com) to browse available models and get the exact model name before downloading. Model names must match exactly to pull successfully.
:::

<Tabs>
<template #From-homepage>

1. In Open WebUI, click the model dropdown at the top of the chat page.
2. Enter the model name. For example: `llama3.2`.
3. Click the option that says **Pull "llama3.2" from Ollama.com**. The download starts automatically.
   ![Download from homepage](/images/one/open-webui-download-from-homepage.png#bordered)

Wait for the download to complete. Progress appears in the interface.

</template>
<template #From-Settings>

1. Click your profile icon and select **Admin Panel**.
2. Navigate to **Settings** > **Models**.
3. Click <span class="material-symbols-outlined">download_2</span> in the top right to open the **Manage Models** dialog.
4. Under **Pull a model from Ollama.com**, enter the model name. For example: `llama3.2`.
   ![Download from settings](/images/one/open-webui-download-from-settings.png#bordered)

5. Click <i class="material-symbols-outlined">download</i> to start the download.

</template>
</Tabs>

:::tip Download time
Models range from 2 GB to 20+ GB. Download time depends on your network speed.
:::

#### Verify the connection

With Ollama, Open WebUI connects automatically. When you see the model in the dropdown on the chat page, the connection is established.

### Option B: Model app

Model apps package a specific model with pre-configured settings. The model downloads automatically after you install the app.

#### Install the model app

1. Open Market and search for your desired model.
2. Click **Get**, then **Install**, and wait for installation to complete.
   ![Install model app](/images/one/qwen3.5-27b.png#bordered)

#### Download the model

1. Open the model app you just installed.
2. View the model downloading progress.
   ![Downloading model](/images/one/qwen3.5-27b-downloading.png#bordered)
3. Once you see the completion screen, the model is ready.
   ![Model downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)
4. To let Open WebUI access this model, you need to get its shared endpoint URL.

   a. Open Olares Settings, then navigate to **Applications** > **[Model App Name]**.

   b. In **Shared entrances**, select the model to view the endpoint URL.
      ![Get shared endpoint](/images/one/qwen3.5-27b-shared-entrance.png#bordered){width=90%}

   c. Copy the shared endpoint. For example:
      ```plain
      http://94a553e00.shared.olares.com
      ```
   You will need this URL in a later step.

   :::tip Why not use the URL shown on the model page?
   The URL shown on the model app page is user-specific and relies on browser-based frontend calls. If your device and Olares are not on the same local network, those calls may trigger Olares sign-in and you may encounter cross-origin restrictions (CORS). To avoid these issues, use the shared endpoint URL.
   :::

#### Connect to Open WebUI

1. In Open WebUI, click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Navigate to **Settings** > **Connections**.
3. Click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **URL** field, paste the shared endpoint URL you copied earlier.
5. Click **Save**. Open WebUI verifies the connection automatically.
   ![Connection established](/images/one/open-webui-connection-established.png#bordered)

When you see "Ollama API settings updated", the connection is established.

## Start chatting

1. On the main chat page, confirm that your model is selected in the dropdown.
   ![Select model](/images/one/open-webui-qwen3.5-27b.png#bordered)
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with LLM](/images/one/open-webui-chat1.png#bordered)

## Learn more

- [Set up multi-user access](openwebui-multiuser.md): Share Open WebUI with other users on your Olares device.
- [Configure audio](openwebui-audio.md): Enable speech-to-text and text-to-speech.
- [Enable web search](openwebui-search.md): Add web search capabilities to your chats.
- [Use knowledge base](openwebui-knowledge.md): Upload documents and create a knowledge base for RAG.
