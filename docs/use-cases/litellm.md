---
outline: [2, 3]
description: Set up LiteLLM on Olares to unify multiple AI model providers behind a single OpenAI-compatible API, then connect it to apps like Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, LiteLLM, AI gateway, model proxy, OpenAI-compatible, Ollama, Open WebUI, self-hosted
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-08"
---

# Use LiteLLM as a unified AI model gateway

LiteLLM is an AI gateway that unifies APIs from different model providers (such as OpenAI, Anthropic, Google, and local engines like Ollama) into a single OpenAI-compatible interface. It automatically converts request parameters to the format each provider expects and routes requests to the correct backend.

Running LiteLLM on Olares gives you a central place to manage all your model configurations, switch freely between remote and local providers, and expose a single API endpoint for other apps to consume.

## Learning objectives

In this guide, you will learn how to:
- Add and configure AI models from different providers in LiteLLM.
- Test model connections using the built-in Playground.
- Generate virtual keys and connect LiteLLM to Open WebUI.
- Monitor model usage and API call logs.

## Prerequisites

- Model apps installed from Market (for example, an Ollama-based model app)
- Olares admin privileges

## Install LiteLLM

1. Open Market and search for "LiteLLM".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Understand the LiteLLM gateway

LiteLLM sits between your apps and model providers, acting as a proxy layer:

- **Unified interface**: LiteLLM normalizes the different API formats from providers like OpenAI, Anthropic, Google, and local engines (Ollama, vLLM) into a single OpenAI-compatible standard.
- **Automatic format conversion**: When you send a request using the standard parameters, LiteLLM translates them into the specific parameter names and data structures the target provider expects.
- **Request routing**: Based on the model name in your request, LiteLLM determines whether to forward it to a remote cloud provider or a local model server.

<!-- ![LiteLLM gateway diagram](/images/manual/use-cases/litellm-gateway.png#bordered) -->

This means your apps only need one API endpoint to access all your models.

## Add a model

This example uses a model app installed from Market. The process is similar for other providers.

1. Open LiteLLM and log in with the default admin credentials shown on the login page.
   <!-- ![Log in to LiteLLM](/images/manual/use-cases/litellm-login.png#bordered) -->
2. Open the model app to find its model name. You will need it in a later step.
   <!-- ![Model name on the model app page](/images/manual/use-cases/litellm-model-name.png#bordered) -->
3. In LiteLLM, navigate to **Models + Endpoints** > **Add Model**.
4. For **Provider**, select the engine that powers the model app. For example, if the model app name includes "Ollama", select **Ollama**.
5. In **LiteLLM Model Name(s)**, enter the model name shown on the model app's page.
   <!-- ![Add model](/images/manual/use-cases/litellm-add-model.png#bordered) -->
6. (Optional) Under **Model Mappings**, set a **Public Model Name** to give the model a shorter alias for external calls.
   <!-- ![Model mappings](/images/manual/use-cases/litellm-model-mappings.png#bordered) -->
7. For **API Base**, enter the shared endpoint URL of the model app.

   To find the URL, open Settings and navigate to **Applications** > **[Your model app]**. In **Shared entrances**, select the model app to view and copy the endpoint URL.

   :::warning
   Do not append `/v1` to the API Base URL. Adding it will cause the connection to fail.
   :::

8. Click **Test Connection** at the bottom of the page. Once you see a success message, click **Add Model**.
   <!-- ![Test connection](/images/manual/use-cases/litellm-test-connection.png#bordered) -->

You can view all added models under **All Models**.

<!-- ![All models](/images/manual/use-cases/litellm-all-models.png#bordered) -->

## Test a model in Playground

1. Navigate to **Playground** > **Chat** > **Configuration**.
2. Configure the following settings:
   - **Virtual Key Source**: Keep the default **Current UI Session**.
   - **Endpoint Type**: Select the mode that matches your model. For chat models, select `v1/chat/completions`.
   - **Select Model**: Choose the model you just added.

   :::warning
   Leave **Custom Proxy Base URL** empty. Filling it in will cause errors.
   :::

   <!-- ![Playground configuration](/images/manual/use-cases/litellm-playground.png#bordered) -->

3. Enter a prompt and send it. You can evaluate the model's performance, including tokens per second, latency, and response quality.
   <!-- ![Playground test results](/images/manual/use-cases/litellm-playground-test.png#bordered) -->

:::tip View model details
To check a model's supported features and parameters, navigate to **AI Hub** > **Details**.
:::

## Use LiteLLM with Open WebUI

This section uses Open WebUI as an example. The same approach applies to other apps that support OpenAI-compatible APIs.

### Generate a virtual key

1. Navigate to **Virtual Keys** and click **Create New Key**.
2. Enter a **Key Name**.
3. Select the models this key can access.
4. Keep the other options as default, then click **Create Key**.
   <!-- ![Create virtual key](/images/manual/use-cases/litellm-create-key.png#bordered) -->
5. Copy the virtual key for later use.
   <!-- ![Copy virtual key](/images/manual/use-cases/litellm-copy-key.png#bordered) -->

### Get the LiteLLM API endpoint

1. Open Settings and navigate to **Applications** > **LiteLLM** > **Entrances** > **LiteLLM API**.
2. Copy the URL.
   <!-- ![LiteLLM API entrance](/images/manual/use-cases/litellm-api-entrance.png#bordered) -->

:::info Internal vs. public access
The LiteLLM API endpoint is set to **Internal** by default, meaning only apps on the same local network can access it. If you need to access LiteLLM from outside your local network, change the access level to **Public**. LiteLLM's API key authentication will control access.
:::

### Connect Open WebUI to LiteLLM

1. In Open WebUI, go to **Admin Panel** > **Settings** > **Connections**.
2. Under **OpenAI API**, click <span class="material-symbols-outlined">add</span> to add a new connection.
3. Enter the LiteLLM API URL and the virtual key you copied earlier.
4. Set **Connection type** to **Local**, then click the test button to verify the connection.
   <!-- ![Open WebUI connection](/images/manual/use-cases/litellm-openwebui-connection.png#bordered) -->
5. Navigate to **Models** to confirm that the models configured in LiteLLM are now available, with the public names you set earlier.
   <!-- ![Models in Open WebUI](/images/manual/use-cases/litellm-openwebui-models.png#bordered) -->

### Chat and monitor usage

1. Start a chat in Open WebUI and select a LiteLLM-managed model.
   <!-- ![Chat in Open WebUI](/images/manual/use-cases/litellm-openwebui-chat.png#bordered) -->
2. Back in LiteLLM, check **Logs** and **Usage** to view detailed call records and usage statistics.
   <!-- ![LiteLLM usage logs](/images/manual/use-cases/litellm-usage-logs.png#bordered) -->

## Learn more

- [Download and run local AI models via Ollama](ollama.md): Install models that LiteLLM can aggregate.
- [Chat with local LLMs using Open WebUI](openwebui.md): Connect Open WebUI directly to model apps without LiteLLM.
- [LiteLLM official documentation](https://docs.litellm.ai/docs/): Advanced features including team management, usage monitoring, and access control.
