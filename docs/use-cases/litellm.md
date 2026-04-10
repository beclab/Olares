---
outline: [2, 3]
description: Set up LiteLLM on Olares to unify multiple AI model providers behind a single OpenAI-compatible API, then connect it to apps like Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, LiteLLM, AI gateway, model proxy, OpenAI-compatible, Ollama, Open WebUI, self-hosted
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-09"
---

# Use LiteLLM as a unified AI model gateway

LiteLLM is an AI gateway that unifies APIs from different model providers (such as OpenAI, Anthropic, Google, and local engines like Ollama) into a single OpenAI-compatible interface. It automatically converts request parameters to the format each provider expects and routes requests to the correct backend.

Running LiteLLM on Olares gives you a central place to manage all your model configurations, switch freely between remote and local providers, and expose a single API endpoint for other apps to consume.

## Learning objectives

In this guide, you will learn how to:
- Install LiteLLM.
- Add and configure AI models from providers like Ollama in LiteLLM.
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

- One or more model apps installed from the Market. This tutorial uses the **Qwen3.5 9B Q4_K_M (Ollama)** app as an example.
- Olares admin privileges.

## Install LiteLLM

1. Open Market and search for "LiteLLM".

   ![LiteLLM in Market](/images/manual/use-cases/litellm.png#bordered)

2. Click **Get**, and then click **Install**.
3. When prompted, set the environment variables:

   - **UI_USERNAME**: Specify the username for admin account.
   - **UI_PASSWORD**: Specify the password for admin account.
4. Click **Confirm** and wait for the installation to finish.

## Add a model

This example uses the model app "Qwen3.5 9B Q4_K_M (Ollama)". The process is similar for other providers.

1. Open the Qwen3.5 9B Q4_K_M (Ollama) app from the Launchpad, and then note down the model name exactly as shown. In this case, it is `qwen3.5:9b`.

   ![Model name on the model app page](/images/manual/use-cases/litellm-model-name.png#bordered){width=55%}

2. Open **Settings**, go to **Applications** > **Qwen3.5 9B Q4_K_M (Ollama)**, click the model name under **Shared entrances**, and then note down the endpoint URL. In this case, it is `http://bd5355000.shared.olares.com`.

   ![Model endpoint on Settings page](/images/manual/use-cases/litellm-model-endpoint.png#bordered){width=80%}

3. Open LiteLLM from the Launchpad, and then log in with the admin credentials you set during installation.

   <!--![LiteLLM login](/images/manual/use-cases/litellm-login.png#bordered){width=50%}-->

4. Select **Models + Endpoints** from the left sidebar, and then click the **Add Model** tab.

   ![Add Model tab](/images/manual/use-cases/litellm-add-model-tab.png#bordered)

5. Configure the following settings:

   - **Provider**: Select the engine that powers the model app. For example, if the model app name includes "Ollama", select **Ollama**.
   - **LiteLLM Model Name(s)**: Enter the exact model name that you noted down. In this case, it is `qwen3.5:9b`.
   - (Optional) **Public Model Name**: Specify a shorter alias for the model to use in external client apps.
   - **API Base**: Enter the model app's shared endpoint URL that you noted down. In this case, it is `http://bd5355000.shared.olares.com`.

      :::warning
      Do not append `/v1` to the API Base URL. Adding it will cause the connection to fail.
      :::

6. Click **Test Connect** at the bottom of the page.
7. When the **Connection Test Results** window shows a connection success message, close the window.

   ![Test connection](/images/manual/use-cases/litellm-test-connection.png#bordered){width=60%}

8. Click **Add Model** next to **Test Connect**. You can now view your newly added model on the **All Models** tab.

   ![All models](/images/manual/use-cases/litellm-all-models.png#bordered)

## Test the model

1. Select **Playground** from the left sidebar.
2. On the **Chat** tab, configure the following settings:
   - **Virtual Key Source**: Keep the default **Current UI Session**.
   - **Custom Proxy Base URL**: Leave this empty. Filling it in will cause errors.
   - **Endpoint Type**: Select the mode that matches your model. For chat models, select **v1/chat/completions**.
   - **Select Model**: Select the model you just added. In this case, it is **qwen3.5:9b**.

   ![Playground configuration](/images/manual/use-cases/litellm-playground.png#bordered)

3. On the **Test Key** panel, send a prompt in the chat to evaluate the model's performance.

   For example:

   ```text
   Write a 3-paragraph sci-fi story about a robot discovering a forgotten library
   ```

   You can review metrics such as Time to First Token (TTFT), total latency, and input/output token counts.
   
   ![Playground test results](/images/manual/use-cases/litellm-playground-test.png#bordered)

4. To check the model's supported features and parameters, select **AI Hub** from the left sidebar, and then click **Details** on the **Model Hub** tab.

   ![View model details](/images/manual/use-cases/litellm-view-model-details.png#bordered)

   You can see the details on the model overview page.

   ![Model overview](/images/manual/use-cases/litellm-model-overview.png#bordered)   

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
4. In the **Save your Key** window, copy the virtual key for later use. In this case, it is `sk-ZSkc399qrcc3VXutDfxhpA`.
   
   ![Copy virtual key](/images/manual/use-cases/litellm-copy-key.png#bordered){width=60%}

### Obtain the LiteLLM API endpoint

1. Open Settings, go to **Applications** > **LiteLLM** > **Entrances** > **LiteLLM API**.
2. Copy the **Endpoint** URL. In this case, it is `https://6aead52a1.laresprime.olares.com`.

   ![LiteLLM API entrance](/images/manual/use-cases/litellm-api-entrance.png#bordered){width=80%}

:::info Internal vs. public access
The **Authentication Level** of the LiteLLM API endpoint is set to **Internal** by default, which means only apps on the same local network can access it. If you need to access LiteLLM from outside your local network, change the authentication level to **Public**. LiteLLM's API key authentication will control access.
:::

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

   ![Models in Open WebUI](/images/manual/use-cases/litellm-openwebui-models.png#bordered)

### Chat and monitor usage

1. Start a new chat in Open WebUI and select your LiteLLM-managed model to verify that it responds correctly in the conversation.

   ![Chat in Open WebUI](/images/manual/use-cases/litellm-openwebui-chat.png#bordered)

2. Return to LiteLLM to monitor your usage data.

   - To view graphical usage statistics, select **Usage** from the left sidebar.

   ![LiteLLM usage statistics](/images/manual/use-cases/litellm-usage.png#bordered)

   - To view detailed API request records, select **Logs** from the left sidebar.

   ![LiteLLM logs](/images/manual/use-cases/litellm-logs.png#bordered)

## Learn more

- [Download and run local AI models via Ollama](ollama.md)
- [Chat with local LLMs using Open WebUI](openwebui.md)
- [LiteLLM official documentation](https://docs.litellm.ai/docs/)
