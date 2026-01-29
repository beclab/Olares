---
outline: [2, 3]
description: Learn how to set up Open WebUI on your Olares One to chat with local LLMs using Ollama.
head:
  - - meta
    - name: keywords
      content: Open WebUI, Ollama, local LLM, chatbot, AI
---

# Chat with local LLMs using Open WebUI  <Badge type="tip" text="20 min" />
Open WebUI provides an intuitive interface for managing Large Language Models (LLMs) that supports both Ollama and OpenAI-compatible APIs.

You can easily install Ollama and chat with your local LLM using the Open WebUI.

## Learning objectives
- Download models directly from the interface of Open WebUI.
- Start your first local AI conversation.

## Prerequisites
**Hardware** <br>
- Olares One connected to a stable network.
- Sufficient disk space to download models.

**User permissions**
- Admin privileges to install Ollama from the Market.

## Step 1: Install Ollama and Open WebUI
1. Open Market, and search for "Ollama."
   ![Install Ollama](/images/manual/use-cases/ollama.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

3. Repeat the same steps to install "Open WebUI."
   ![Install Open WebUI](/images/one/open-webui.png#bordered)

## Step 2: Create Open WebUI admin account
1. Open the Open WebUI app.
2. On the welcome page, click **Get started**. 
3. Enter your name, email, and password to create the account.
   ![Create account](/images/one/open-webui-create-account.png#bordered)
   :::info
   All your data, including login details, is locally stored on your device.
   :::
   :::tip First account is Admin
   The first account created on Open WebUI gains Administrator privileges, controlling user management and system settings.
   :::
## Step 3: Download models
Because the Ollama API is pre-configured on Olares OS for Open WebUI, you can download models directly within the Open WebUI interface without using the command line.

:::tip Check Ollama library
If you are unsure which model to download, check the [Ollama Library](https://ollama.com/library) to explore available models.
:::
<tabs>
<template #Download-from-dropdown>

1. On the main chat page, click **Select a model** in the top-left corner.
2. In the search field, enter the model name. For example: `llama3.2`. 
3.  Click the option that says **Pull "llama3.2" from Ollama.com**. The download starts automatically.
   ![Download models from homepage](/images/one/open-webui-download-from-homepage.png#bordered)

</template>
<template #Download-from-Settings>

1. Click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Click <span class="material-symbols-outlined">download_2</span> in the top right to open the **Manage Models** dialog.
3. Under **Pull a model from Ollama.com**, enter the model name. For example: `llama3.2`.
4. Click <span class="material-symbols-outlined">download</span> to initiate the download.
   ![Download models from Settings](/images/one/open-webui-download-from-settings.png#bordered)

</template>
</tabs>

## Step 4: Chat with your local LLM
1. On the main chat page, click the model selector in the top-left and choose the model you just downloaded.
2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with your local LLM](/images/one/open-webui-chat.png#bordered)

## Troubleshooting

### Download progress disappears
When downloading a model via the dropdown menu, the progress bar might sometimes disappear before completion.

To resume the download:
1. Click the model selector again.
2. Enter the exact same model name.
3. Select **Pull from Ollama.com**. The download will resume from where it left off.

## Resources
- [Open WebUI Documentation Hub](https://docs.openwebui.com/getting-started/)
- [Switch GPU mode](gpu.md)
- [Download and run local AI models via Ollama](../use-cases/ollama.md)
- [Deep research with DeerFlow](deerflow.md)