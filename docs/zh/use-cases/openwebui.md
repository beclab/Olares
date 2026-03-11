---
outline: deep
description: Set up Open WebUI on Olares to chat with local Large Language Models using pre-configured model apps.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, local LLM, AI chatbot
app_version: "1.0.20"
doc_version: "2.0"
doc_updated: "2026-03-11"
---

# Chat with local LLMs using Open WebUI

Open WebUI provides an intuitive chat interface for managing LLMs. Running Open WebUI on an Olares device gives you a private, self-hosted alternative to cloud-based AI services, ensuring your conversations remain on your own hardware.

This guide covers the recommended approach: using model apps from the Market.

:::tip Alternative method
If you prefer to manage multiple models through the Ollama app, see [Set up Open WebUI with Ollama](openwebui-ollama.md).
:::

## Learning objectives

In this guide, you will learn how to:
- Run a model app from the Market with Open WebUI on Olares.
- Use the model app URL to connect the model to Open WebUI.
- Start a chat using the connected local model.

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Install the model app and Open WebUI

1. Open **Market** and search for your desired model.
2. Click **Get**, then click **Install**.
   ![Install model app](/images/one/qwen3.5-27b.png#bordered)
3. Search for "Open WebUI" and install it as well.
   ![Install Open WebUI](/images/one/open-webui.png#bordered)
4. Wait for both installations to complete.

## Download the model

1. Open the model app you just installed.
2. View the model downloading progress.
   ![Downloading model](/images/one/qwen3.5-27b-downloading.png#bordered)
3. Once you see the completion screen, the model is ready.
   ![Model downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)
4. Copy the URL shown on the model page. You will need this to configure Open WebUI.

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

## Configure the connection

1. Click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Navigate to **Settings** > **Connections**.
3. Click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **URL** field, paste the URL you copied earlier.
5. Click **Save**. Open WebUI verifies the connection automatically.
   ![Connection established](/images/one/open-webui-connection-established1.png#bordered)

When you see "Ollama API settings updated", the connection is established.

## Start chatting

1. On the main chat page, confirm that your model is selected in the dropdown.
   ![Select model](/images/one/open-webui-qwen3.5-27b.png#bordered)
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with LLM](/images/one/open-webui-chat1.png#bordered)

## Troubleshooting

### Model app is stuck at "Waiting for Ollama" or "Needs attention"

If the model app stays in these states for more than a few minutes:

- Go to **Settings** > **GPU**.
- If you are using **Memory slicing**, make sure the model app is bound to the GPU and has enough VRAM allocated.
- If you are using **App exclusive**, make sure the exclusive app is set to your model app.

Then restart the model app from Launchpad and check the status again.
