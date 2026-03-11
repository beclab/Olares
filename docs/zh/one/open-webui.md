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

This guide walks you through installing Open WebUI and Qwen3.5 27B on Olares One, connecting the model, and starting your first chat. By the end, you'll have a private, local chatbot ready for everyday use.

## Learning objectives
- Use Open WebUI on Olares One to run local LLMs.
- Make the Qwen3.5 27B Q4_K_M model available to other apps.

## Prerequisites
**Hardware** <br>
- Olares One connected to a stable network.
- At least 20 GB free disk space to download the model.
- Sufficient GPU VRAM and system memory to run LLMs.

**User permissions**
- Admin privileges to install shared apps from the Market and manage GPU resources.

## Step 1: Install Open WebUI
1. Open Market, and search for "Open WebUI".
   ![Install Open WebUI](/images/one/open-webui.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Step 2: Install Qwen3.5 27B and get the shared endpoint
1. In Market, search for and install "Qwen3.5 27B Q4_K_M (Ollama)". Wait for the installation to complete.
   ![Install Qwen3.5 27B](/images/one/qwen3.5-27b.png#bordered)

2. Once installed, click **Open** to view the model downloading progress.
   ![Downloading Qwen3.5 27B](/images/one/qwen3.5-27b-downloading.png#bordered)
   :::tip
   The model file is approximately 17 GB. Download time varies depending on your network speed.
   :::

3. Once you see the following screen, the model is ready to use.
   ![Qwen3.5 27B downloaded](/images/one/qwen3.5-27b-downloaded.png#bordered)

4. Copy the URL shown on the model page. You will need this to configure Open WebUI.

## Step 3: Create an Open WebUI admin account
1. Open the Open WebUI app.
2. On the welcome page, click **Get started**.
3. Enter your name, email, and password to create the account.
   ![Create account](/images/one/open-webui-create-account.png#bordered)
   :::info
   All your data, including login details, is stored locally on your Olares One.
   :::
   :::tip First account is admin
   The first account created on Open WebUI has administrator privileges, giving you full control over user management and system settings.
   :::

## Step 4: Configure connections
1. Click your **profile icon** in the bottom-left corner and select **Admin Panel**.
2. Go to **Settings** > **Connections**.
   :::info
   By default, the local Ollama API is pre-configured and visible under **Manage Ollama API connections**.
   :::
3. Click <span class="material-symbols-outlined">add</span> to open the Add Connection dialog.
4. In the **URL** field, paste the URL you copied in Step 2, then click **Save**. Open WebUI automatically verifies the connection. When you see "Ollama API settings updated", the connection is established.
   ![Connection established](/images/one/open-webui-connection-established1.png#bordered)

## Step 5: Chat with your local LLM
1. On the main chat page, confirm that **qwen3.5:27b-q4_K_M** is selected in the model dropdown.
   ![Chat with your local LLM](/images/one/open-webui-qwen3.5-27b.png#bordered)

2. Enter your prompt in the text box and press **Enter** to start chatting.
   ![Chat with your local LLM](/images/one/open-webui-chat1.png#bordered)

## Troubleshooting
### Qwen3.5 27B is stuck at "Waiting for Ollama" or "Needs attention"

If the Qwen3.5 27B app stays in one of these states for more than a few minutes, first check your GPU mode in **Settings** > **GPU**:

- If you are in **Memory slicing** mode, make sure you have bound the Qwen3.5 27B app and allocated it sufficient VRAM.
- If you are in **App exclusive** mode, make sure the app with full GPU access is Qwen3.5 27B.

## Resources
- [Open WebUI Documentation Hub](https://docs.openwebui.com/getting-started/)
- [Switch GPU mode](gpu.md)
- [More Open WebUI features](../use-cases/openwebui.md)
