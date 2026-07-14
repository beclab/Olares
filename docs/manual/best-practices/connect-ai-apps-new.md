---
outline: [2, 3]
description: Learn how to connect AI applications, IDEs, and workflow automation tools to your self-hosted large language models on Olares using the Model Console.
---

# Connect AI apps

Many AI applications on Olares follow a standard connection pattern: one AI service app provides AI capabilities over an API, and another AI client app provides the chat interface you interact with every day. Once you understand this pattern, you can connect almost any compatible combination of apps.

This tutorial explains the core concepts and walks you through a few practical examples using **Qwen3.6-27B (llama.cpp)** as the AI service app.

:::tip For Olares 1.12.5 and earlier
This guide applies to Olares v1.12.6 and later. The way AI apps connect changed in v1.12.6 with the introduction of the Model Console. If you are on Olares 1.12.5 or earlier, please refer to the [legacy AI apps connection guide](./connect-ai-apps.md).
:::

## Learning objectives

By the end of this tutorial, you will be able to:

- Distinguish between AI service apps and AI client apps.
- Understand the types of AI service apps available in Olares v1.12.6.
- Obtain the correct **Base URL** and **Model name** from the Model Console.
- Connect common client apps to a model service app.

## Core concepts

### AI service apps vs. AI client apps

- **AI service apps** act as the backend engine. They provide AI capabilities for compatible clients over an API, and they often run as services without a user-facing graphical interface of their own. For example, model instances created from Engine Base apps, pre-built model apps, and other AI model apps such as OCR or speech.
- **AI client apps** act as the user-facing apps. They provide the chat interface or workflow canvas you interact with directly, but they rely on an AI service app to generate responses. For example, LobeHub, Open WebUI, and n8n.

### Types of AI service apps

In Olares v1.12.6, AI service apps mainly fall into the following categories:

- **LLM service apps**: Apps that host large language models for text generation, code completion, and chat. They include:
    - **[Engine Base apps](/use-cases/llm-base-apps.md)**: Four reusable base applications (Ollama Engine Base, vLLM Engine Base, SGLang Engine Base, and llama.cpp Engine Base). You clone them into independent model instances and configure each instance yourself.
    - **Pre-built model apps**: Nine ready-to-use apps that package a specific model with a specific engine, such as Qwen3.6-27B (llama.cpp) and Gemma 4 26B (Ollama).
- **Other AI service apps**: Apps that provide other AI capabilities, such as OCR or speech recognition, which follow the same connection pattern.

### Model Console and Base URL

Olares v1.12.6 streamlines app-to-app communication by introducing the **Model Console**. Instead of manually configuring network routing and authentication levels, you can obtain an optimized **Base URL** directly from the Model Console of the AI service app and then link your clients.

## Step 1: Obtain the connection details from the Model Console

Connecting any AI client to a local model service involves two steps. First, get the endpoint from the dedicated console of the model service app.

1. Open **Qwen3.6-27B (llama.cpp)** from the Launchpad to launch its built-in model console.
2. Ensure that the **Model** shows **Ready** and the **Engine** shows **Running**.
3. Select the **Connection source** based on where your client app is deployed:

    - **Apps in Olares**: Choose this if the client app is installed directly inside your Olares cluster (for example, LobeHub or n8n). This leverages low-latency cluster-internal routing.
    - **Devices on your network**: Choose this if the client is running on a laptop or device connected to the same local network as Olares.
    - **Remote**: Choose this if you are connecting an external client over the public internet (requires your LarePass VPN to be active).

4. Select the **API format** required by your client.
5. Note down the **Model name** and **Base URL**.

## Step 2: Configure AI client apps

### Example 1: Connect a model service app to LobeHub

In this scenario, LobeHub runs as a client application inside Olares, using an active model service app.

1. Follow [Step 1](#step-1-obtain-the-connection-details-from-the-model-console) to open your model's console and configure the connection source:

    - **Connection source**: `Apps in Olares`
    - **API format**: `OpenAI-Compatible`
    - **Copy the Base URL** (for example, `https://e46e044d.alice2026.olares.com/v1`).

2. Open **LobeHub** from your Launchpad, and then go to **Settings** > **AI Service Provider**.
3. Select the provider that matches your API format: **OpenAI**.
4. In the **API Key** field, enter any placeholder text (such as `olares`) to satisfy the interface requirement.
5. In the **API proxy URL** field, paste the Base URL you copied.

    :::warning Disable Client Request Mode
    Do not enable **Use Client Request Mode** in LobeHub. Enabling this forces the application to make frontend browser calls, which can trigger cross-origin (CORS) blocks or Olares security authentication prompts. Keeping it disabled ensures secure, direct backend-to-backend communication.
    :::

6. Next to **Model List**, click **Fetch models**. The model name you noted down appears in the list.
7. Enable the exact **Model name**.
8. For **Connectivity Check**, select the model name from the drop-down list, and then click **Check**.

When **Check Passed** appears, the connection is established.

![LobeHub connected to a model instance](/images/manual/tutorials/connect-app-eg-lobehub.png#bordered)

### Example 2: Connect a model service app to n8n

n8n workflow automation makes requests directly from its backend environment on Olares, making it highly reliable when paired with the proper internal endpoint.

1. Follow [Step 1](#step-1-obtain-the-connection-details-from-the-model-console) to get your model's credentials:

    - **Connection source**: `Apps in Olares`
    - **API format**: `OpenAI-Compatible`
    - **Copy the Base URL**.

2. Open **n8n** from your Launchpad and create or open a workflow.
3. Go to Settings > Chat, locate OpenAI, and then click the overflow icon on the right of it, to edit the provider.
4. In Configure OpenAI, click the **Default Credential** drop-down list, and select **Create New Credential**.
5. On the Connection tab, configure the fields:

    - **API Key**: Enter any placeholder string (for example, `olares`).
    - **Base URL**: Paste the Base URL copied from your Model Console.

6. Click **Save** in the upper right corner. **Credential successfully created inside your personal space** appears.
7. Click **Confirm** in the Configure OpenAI window. Chat provider settings updated.
8. Create a new workflow or use an existing workflow, add a **Basic LLM Chain** node, and use your configured model to test the connection.

![n8n connected to a model instance](/images/manual/tutorials/connect-app-eg-n8n.png#bordered)

### Example 3: Connect your local IDE to Olares (IntelliJ IDEA)

You can connect your local IDE to an active model service app on Olares, so that AI assistance and code completion are powered by your own hardware rather than a third-party cloud. This example uses **IntelliJ IDEA** with the **Continue.dev** plugin.

1. Follow [Step 1](#step-1-obtain-the-connection-details-from-the-model-console) to expose the model outside the cluster:

    - **Connection source**: `Remote`
    - **API format**: `OpenAI-Compatible`
    - **Copy the Base URL** (for example, `https://<route-id>.<username>.olares.com/v1`).

2. Open your LarePass desktop client and toggle the VPN to **Enabled**.
3. In IntelliJ IDEA, install the **Continue.dev** plugin from the JetBrains Marketplace if you have not already, and then open the Continue panel.
4. Click the settings gear icon in the Continue panel to edit your `config.yaml`.
5. Map your active Olares models to specific roles. Update your configuration blocks using the example below:

    ```yaml
    name: Olares Remote AI Config
    version: 1.0.0
    schema: v1
    models:
      - name: Qwen3.6-35B-Chat
        provider: openai
        model: unsloth/Qwen3.6-35B-A3B-GGUF:UD-Q4_K_XL # Your exact Model Name
        apiBase: https://<your-copied-remote-base-url>/v1
        apiKey: olares # Placeholder key
        roles:
          - chat
          - edit
      - name: Qwen2.5-Coder-Autocomplete
        provider: openai
        model: qwen2.5-coder:1.5b
        apiBase: https://<your-other-model-remote-base-url>/v1
        apiKey: olares
        roles:
          - autocomplete
    ```

6. Return to the Continue chat panel and submit a prompt to test the connection (for example, `Write a clean Python singleton pattern`).

With LarePass active, your IDE securely tunnels past public barriers, hits your specific Olares model service app, and streams the response directly into your editor.

## Learn more

- [Host local large language models with Engine Base apps](../../use-cases/llm-base-apps.md)
- [Shared applications](../olares/market/shared-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
- [Network](../../developer/concepts/network.md)
