---
outline: [2, 3]
description: Set up LiteLLM on Olares to unify multiple AI model providers behind a single OpenAI-compatible API, then connect it to apps like Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, LiteLLM, AI gateway, model proxy, OpenAI-compatible, Ollama, Open WebUI, self-hosted
app_version: "1.0.8"
doc_version: "1.1"
doc_updated: "2026-07-29"
---

# Use LiteLLM as a unified AI model gateway

LiteLLM is an AI gateway that unifies APIs from different model providers (such as OpenAI, Anthropic, Google, and local engines like Ollama) into a single OpenAI-compatible interface. It automatically converts request parameters to the format each provider expects and routes requests to the correct backend.

Running LiteLLM on Olares gives you a central place to manage all your model configurations, switch freely between remote and local providers, and expose a single API endpoint for other apps to consume.

## Learning objectives

In this guide, you will learn how to:
- Install LiteLLM.
- Add and configure an OpenAI-compatible local model in LiteLLM.
- Test model connection using the built-in Playground.
- Generate virtual keys and connect LiteLLM to Open WebUI.
- Monitor API call logs and model usage statistics.

## Understand the LiteLLM gateway

LiteLLM sits between your apps and model providers, acting as a proxy layer:
- **Unified interface**: LiteLLM normalizes the different API formats from providers like OpenAI, Anthropic, Google, and local engines (Ollama, vLLM) into a single OpenAI-compatible standard.
- **Automatic format conversion**: When you send a request using the standard parameters, LiteLLM translates them into the specific parameter names and data structures the target provider expects.
- **Request routing**: Based on the model name in your request, LiteLLM determines whether to forward it to a remote cloud provider or a local model server.

![LiteLLM gateway diagram](/images/manual/use-cases/litellm-gateway.png#bordered){width=80%}

Because of this unified layer, your client apps only need one API endpoint to access all your configured models.

## Prerequisites

Before you begin, you need:

- Olares admin privileges.
- The following model:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.6-27B (llama.cpp) | Install from Market |

## Install LiteLLM

1. Open Market and search for "LiteLLM".

   ![LiteLLM in Market](/images/manual/use-cases/litellm.png#bordered)

2. Click **Get**, and then click **Install**.
3. When prompted, set the environment variables:

   - **UI_USERNAME**: Specify the username for admin account.
   - **UI_PASSWORD**: Specify the password for admin account.
4. Click **Confirm** and wait for the installation to finish.

## Add a model

### Get model connection details

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For Qwen3.6-27B (llama.cpp), LiteLLM uses the OpenAI-compatible API format. In the Model Console, select **OpenAI-Compatible**, then follow these steps:

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Add the model to LiteLLM

1. Open LiteLLM from the Launchpad, and then log in with the admin credentials you set during installation.

   <!--![LiteLLM login](/images/manual/use-cases/litellm-login.png#bordered){width=50%}-->

2. Select **Models + Endpoints** from the left sidebar, and then click the **Add Model** tab.

   ![Add Model tab](/images/manual/use-cases/litellm-add-model-tab.png#bordered)

3. Configure the following settings:

   - **Provider**: Select **OpenAI**.
   - **LiteLLM Model Name(s)**: Enter the exact Model name copied from the Model Console. In this example, it is `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.
   - **Public Model Name**: Enter a recognizable alias, such as `qwen3.6-27b`.
   - **API Base**: Paste the Base URL from the Qwen3.6-27B Model Console. Use it exactly as displayed.
   - **API Key**: Enter a placeholder value, such as `local`.

4. Click **Test Connect** at the bottom of the page.
5. When the **Connection Test Results** window shows a connection success message, close the window.

6. Click **Add Model** next to **Test Connect**. You can now view your newly added model on the **All Models** tab.

## Test the model

1. Select **Playground** from the left sidebar.
2. On the **Chat** tab, configure the following settings:
   - **Virtual Key Source**: Keep the default **Current UI Session**.
   - **Custom Proxy Base URL**: Leave this empty. Filling it in will cause errors.
   - **Endpoint Type**: Select the mode that matches your model. For chat models, select **v1/chat/completions**.
   - **Select Model**: Select the model you just added. In this example, it is **qwen3.6-27b**.

3. On the **Test Key** panel, send a prompt in the chat to evaluate the model's performance.

   For example:

   ```text
   Write a 3-paragraph sci-fi story about a robot discovering a forgotten library
   ```

   You can review metrics such as Time to First Token (TTFT), total latency, and input/output token counts.
   
4. To check the model's supported features and parameters, select **AI Hub** from the left sidebar, and then click **Details** on the **Model Hub** tab.

   You can see the details on the model overview page.

## Use LiteLLM with Open WebUI

This section uses Open WebUI as an example. The same approach applies to any client app that supports OpenAI-compatible APIs.

### Generate a virtual key

1. In LiteLLM, select **Virtual Keys** from the left sidebar, and then click **Create New Key**.
2. In the Key Ownership window, configure the following settings:

   - **Key Name**: Enter a descriptive name for easy identification.
   - **Models**: Select the models this key is allowed to access.
   - Keep all other options as their defaults.

   ![Create virtual key](/images/manual/use-cases/litellm-create-key.png#bordered)
   
3. Click **Create Key**.
4. In the **Save your Key** window, copy the virtual key for later use.
   
   ![Copy virtual key](/images/manual/use-cases/litellm-copy-key.png#bordered){width=60%}

### Get the LiteLLM API endpoint

<!--@include: ../reusables/ai-service-connections.md#app-endpoint-overview-->

For LiteLLM:

1. Go to Olares **Settings** > **Applications** > **LiteLLM** > **Entrances**.
2. Select **LiteLLM API**, then copy the **Endpoint** URL.

Use this Endpoint as the **API Base URL** in Open WebUI.

### Connect Open WebUI to LiteLLM

1. Launch Open WebUI, click your user avatar in the lower-left corner, and then select **Admin Panel**.
2. Click the **Settings** tab, and then click **Connections**.

   ![Open WebUI connections page](/images/manual/use-cases/litellm-openwebui-connection.png#bordered)

3. Under **OpenAI API**, click <span class="material-symbols-outlined">add</span> to add a new connection.
4. In the **Add Connection** window, configure the following settings:

   - **Connection Type**: Click **External** to switch it to **Local**.
   - **API Base URL**: Enter the LiteLLM API URL that you noted down earlier.
   - **API Key**: Enter the virtual key you copied earlier.

   ![Open WebUI connection setup](/images/manual/use-cases/litellm-openwebui-connection-setup.png#bordered){width=60%}

5. Click <span class="material-symbols-outlined">cached</span> to verify the connection.
6. When you see the "Server connection verified" message, click **Save**. 
7. Under **Connections**, select **Models** to confirm that the model configured in LiteLLM is now available, displayed with the public model name you set earlier.

### Chat and monitor usage

1. Start a new chat in Open WebUI and select your LiteLLM-managed model to verify that it responds correctly in the conversation.

   ![Chat in Open WebUI](/images/manual/use-cases/litellm-openwebui-chat.png#bordered)

2. Return to LiteLLM to monitor your usage data.

   - To view graphical usage statistics, select **Usage** from the left sidebar.

   ![LiteLLM usage statistics](/images/manual/use-cases/litellm-usage.png#bordered)

   - To view detailed API request records, select **Logs** from the left sidebar.

   ![LiteLLM logs](/images/manual/use-cases/litellm-logs.png#bordered)

## Learn more

- [Host local large language models with Engine Base apps](llm-base-apps.md)
- [Chat with local LLMs using Open WebUI](openwebui.md)
- [LiteLLM official documentation](https://docs.litellm.ai/docs/)
