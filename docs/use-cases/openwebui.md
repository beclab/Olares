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

# Chat with Local LLMs Using Open WebUI

Open WebUI provides a user-friendly chat interface for local models on your Olares device. It does not include any models by default, so you need to install a dedicated model app that acts as the backend. This guide walks you through the setup process using the Qwen3.5 27B Q4_K_M model app as an example.

:::tip Alternative method
If you prefer to manage multiple models through the Ollama app, see [Set up Open WebUI with Ollama](openwebui-ollama.md).
:::

## Learning objectives

In this guide, you will learn how to:
- Run a model app from the Market with Open WebUI on Olares.
- Use the shared endpoint URL to connect the model to Open WebUI.
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
4. To let Open WebUI access this model, you need to get its shared endpoint URL.

   a. Open Olares Settings, then navigate to **Application** > **Qwen3.5 27B Q4_K_M (Ollama)**.

   b. In **Shared entrances**, select **Qwen3.5 27B Q4_K_M** to view the endpoint URL.
      
      ![Get shared endpoint](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered){width=90%}

   c. Copy the shared endpoint. For example:
      ```plain
      http://94a553e00.shared.olares.com
      ```
   You will need this URL in a later step.


   :::tip Why not use the URL shown on the model page?
   The URL shown on the model app page is user-specific and relies on browser-based frontend calls. If your device and Olares are not on the same local network, those calls may trigger Olares sign-in and you may encounter cross-origin restrictions (CORS). To avoid these issues, use the shared endpoint URL.
   :::

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
4. In the **URL** field, paste the shared endpoint URL you copied earlier.
5. Click **Save**. Open WebUI verifies the connection automatically.
   ![Connection established](/images/one/open-webui-connection-established.png#bordered)

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
