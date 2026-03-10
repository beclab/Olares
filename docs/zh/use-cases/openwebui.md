---
outline: deep
description: Set up Open WebUI on Olares to chat with local Large Language Models using either pre-configured model apps or the Ollama engine.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, local LLM, Ollama, AI chatbot
app_version: "1.0.20"
doc_version: "2.0"
doc_updated: "2025-03-10"
---

# Chat with local LLMs using Open WebUI

Open WebUI provides an intuitive chat interface for managing Large Language Models that supports both Ollama and OpenAI-compatible APIs. Running Open WebUI on an Olares device gives you a private, self-hosted alternative to cloud-based AI services, ensuring your conversations remain on your own hardware.

## Learning objectives

In this guide, you will learn how to:
- Install and connect a model app to Open WebUI
- Install Ollama and pull models from the Ollama library
- Start chatting with your local LLM

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Installation options

There are two ways to add models to Open WebUI. Choose the one that best fits how you plan to use local LLMs:

| Method | Best for | Pros | Cons |
|:-------|:---------|:-----|:-----|
| Pre-configured model apps | Running one model at a time | Pre-configured, GPU time slicing supported | Limited to available models in Market |
| Ollama app | Managing multiple models | Access to entire Ollama library | GPU time slicing not supported between models |

## Option 1: Use pre-configured model apps

Pre-configured model apps are the fastest way to get started. Each app packages a specific model with optimal settings for your Olares device.

### Step 1: Install the model app and Open WebUI

1. Open **Market** and search for your desired model.
2. Click **Get**, then click **Install**.
   ![Install model app](/images/one/qwen3.5-27b.png#bordered)
3. Search for "Open WebUI" and install it as well.
   ![Install Open WebUI](/images/one/open-webui.png#bordered)
4. Wait for both installations to complete.

### Step 2: Download the model

1. Open the model app you just installed.
2. View the model downloading progress.
   ![Downloading model](/images/one/qwen3.5-27b-downloading.png#bordered)
3. Once you see the completion screen, the model is ready.
   ![Model downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)

### Step 3: Get the model endpoint

1. Open **Settings**.
2. Navigate to **Application** and select your model app.
3. In **Shared entrances**, select the model to view its endpoint URL.
   ![Model shared entrance](/images/one/qwen3.5-27b-shared-entrance.png#bordered)
4. Copy the URL. You will need this in the next step.

### Step 4: Create an admin account

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
### Step 5: Configure the connection

1. Click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Navigate to **Settings** > **Connections**.
3. Click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **URL** field, paste the shared endpoint you copied in Step 3.
5. Click **Save**. Open WebUI verifies the connection automatically.
   ![Connection established](/images/one/open-webui-connection-established.png#bordered)

When you see "Ollama API settings updated", the connection is established.

### Step 6: Start chatting

1. On the main chat page, confirm that your model is selected in the dropdown.
   ![Select model](/images/one/open-webui-qwen3.5-27b.png#bordered)
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with LLM](/images/one/open-webui-chat1.png#bordered)

## Option 2: Use the Ollama app

The Ollama app gives you access to the entire Ollama model library. This option provides more flexibility if you want to experiment with different models or use specific versions.

### Step 1: Install Ollama and Open WebUI

1. Open **Market** and search for "Ollama".
   ![Install Ollama](/images/manual/use-cases/ollama.png#bordered)

2. Click **Get**, then click **Install**.
3. Search for "Open WebUI" and install it.
   ![Install Open WebUI](/images/one/open-webui.png#bordered)

4. Wait for both installations to complete.

### Step 2: Create an admin account

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

### Step 3: Download a model

:::tip Browse models first
Visit the Ollama Library at [Ollama Library](https://ollama.com/library) to find models before downloading.
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

4. Click <i class="material-symbols-outlined">download</i> to start the download.

</template>
</Tabs>

:::tip Download time
Models range from 2 GB to 20+ GB. Download time depends on your network speed.
:::

### Step 4: Start chatting

1. On the main chat page, confirm that your model is selected in the dropdown.
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with LLM](/images/one/open-webui-chat.png#bordered)

## Troubleshooting

### Pre-configured model app is stuck at “Waiting for Ollama” or “Needs attention”

If the pre-configured model app (for example, **Qwen3.5 27B Q4_K_M (Ollama)**) stays in these states for more than a few minutes:

- Go to **Settings** > **GPU**.
- If you are using **Memory slicing**, make sure the model app is bound to the GPU and has enough VRAM allocated.
- If you are using **App exclusive**, make sure the exclusive app is set to your model app.

Then restart the model app from Launchpad and check the status again.

### Download progress disappears
When downloading a model via the dropdown menu, the progress bar might sometimes disappear before completion.
 
To resume the download:
1. Click the model selector again.
2. Enter the exact same model name.
3. Select **Pull from Ollama.com**. The download will resume from where it left off.
