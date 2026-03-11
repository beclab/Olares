---
outline: deep
description: Set up Open WebUI on Olares with Ollama as the model backend to pull and run models from the Ollama Registry.
head:
  - - meta
    - name: keywords
      content: Olares, Open WebUI, Ollama, local LLM, AI chatbot
app_version: "1.0.20"
doc_version: "1.0"
doc_updated: "2026-03-11"
---

# Set up Open WebUI with Ollama

This guide shows how to use Open WebUI with Ollama on Olares. Ollama lets you pull and switch between different models from the Ollama Registry, offering more flexibility than individual model apps.

:::warning GPU time slicing limitation
When using Ollama to manage multiple models, GPU time slicing is not supported between models.
:::

## Learning objectives

By the end of this guide, you will be able to:
- Use Open WebUI with the Ollama backend on Olares.
- Pull and manage models from Ollama.

## Prerequisites

- An Olares device with sufficient disk space and memory
- Admin privileges to install shared apps from Market

## Install Ollama and Open WebUI

1. Open **Market** and search for "Ollama".
   ![Install Ollama](/images/manual/use-cases/ollama.png#bordered)

2. Click **Get**, then click **Install**.
3. Search for "Open WebUI" and install it.
   ![Install Open WebUI](/images/one/open-webui.png#bordered)

4. Wait for both installations to complete.

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

## Download a model

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

## Start chatting

1. On the main chat page, confirm that your model is selected in the dropdown.
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with LLM](/images/one/open-webui-chat.png#bordered)

## Troubleshooting

### Download progress disappears

When downloading a model via the dropdown menu, the progress bar might sometimes disappear before completion.

To resume the download:
1. Click the model selector again.
2. Enter the exact same model name.
3. Select **Pull from Ollama.com**. The download will resume from where it left off.
