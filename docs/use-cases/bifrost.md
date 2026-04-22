---
outline: [2, 3]
description: Set up Bifrost on Olares as an AI gateway. Aggregate Ollama and single-model apps behind one endpoint, then connect clients like OpenCode and Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, Bifrost, AI gateway, LLM proxy, Ollama, OpenCode, Open WebUI, self-hosted
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-04-22"
---

# Set up Bifrost as an AI model gateway

Bifrost is an AI gateway that sits between your client applications and multiple model providers, such as OpenAI, Anthropic, and local engines like Ollama. It exposes a single OpenAI-compatible endpoint and routes each request to the right backend based on the model name.

Use Bifrost to achieve high request throughput, built-in MCP gateway access, semantic response caching, and automatic provider fallbacks.

## Learning objectives

In this guide, you will learn how to:

- Install Bifrost.
- Add Ollama or a single-model app as a model provider in Bifrost.
- Locate the Bifrost endpoint URL.
- Route models from Bifrost to OpenCode.
- Route models from Bifrost to Open WebUI.
- Verify model connections using Bifrost's observability logs.

## Prerequisites

- [Ollama is installed](ollama.md) on Olares with at least one model downloaded. This tutorial uses `llama3.1:8b` as an example.
- At least one single-model app is installed from Market. This tutorial uses the **Qwen3.5 9B Q4_K_M (Ollama)** app as an example.
- You have Olares administrator privileges.

## Install Bifrost

1. Open Market and search for "Bifrost".

   ![Bifrost in Market](/images/manual/use-cases/bifrost.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Add model providers in Bifrost

In Bifrost, a model provider represents the engine hosting your AI models. You configure a provider by supplying the endpoint URL of the application running the models. 

You can connect the main Ollama application to route every model running inside it, or connect a single-model application to expose just that specific model.

In this tutorial, since both model applications run on the Ollama engine, select **Ollama** as the provider type for both scenarios.

<tabs>
<template #Ollama-app>

Use this method to route every downloaded model in your Ollama instance through Bifrost.

1. Open **Settings**, go to **Applications** > **Ollama** > **Entrances** > **Ollama API**, and then copy the endpoint URL. For example:

   ```plain
   https://a5be22681.laresprime.olares.com
   ```

   ![Ollama endpoint in Settings](/images/manual/use-cases/bifrost-ollama-endpoint.png#bordered){width=80%}

2. Open Bifrost from the Launchpad, go to **Models** > **Model Providers** > **Add provider**, and then select **Ollama**.

   ![Select Ollama as provider](/images/manual/use-cases/bifrost-add-provider-ollama.png#bordered)

3. Click **Edit Provider Config** in the upper-right corner.
4. In **Base URL**, enter the Ollama endpoint URL you copied.

   ![Edit provider config for Ollama](/images/manual/use-cases/bifrost-config-provider-ollama.png#bordered){width=90%}

5. Click **Save Network Configuration**. The message "Provider configuration updated successfully" is displayed.
6. Close the **Ollama Provider configuration** window.
</template>
<template #Single-model-app>

Use this method when the model runs as its own Olares application, such as Qwen3.5 9B Q4_K_M (Ollama).

1. Open **Settings**, go to **Applications** > **Qwen3.5 9B Q4_K_M (Ollama)** > **Entrances**, click the model name under **Shared entrances**, and then note down the endpoint URL. 

   In this case, it is:

   ```plain
   http://bd5355000.shared.olares.com
   ```

   ![Model endpoint on Settings page](/images/manual/use-cases/litellm-model-endpoint.png#bordered){width=80%}

2. Open Bifrost from the Launchpad, go to **Models** > **Model Providers** > **Add provider**, and then select **Ollama**.

   ![Select Ollama as provider](/images/manual/use-cases/bifrost-add-provider-ollama.png#bordered)

3. Click **Edit Provider Config** in the upper-right corner.
4. Configure the following settings:
   - **Base URL**: Paste the endpoint URL you copied. Ensure the Base URL does not end with `/v1`.
   - **Timeout (seconds)**: Set it to `300`. Single-model apps can take longer to warm up than a running Ollama instance.

   ![Edit provider config for single-model app](/images/manual/use-cases/bifrost-single-model-config.png#bordered){width=90%}

5. Click **Save Network Configuration**. The message "Provider configuration updated successfully" is displayed.
6. Close the **Ollama Provider configuration** window.
</template>
</tabs>

## Obtain the Bifrost endpoint

Client applications connect to Bifrost through the Bifrost endpoint URL, not the model provider URLs you configured earlier.

1. Open **Settings**, go to **Applications** > **Bifrost** > **Entrances** > **Bifrost**, and then copy the endpoint URL. For example:

   ```plain
   https://44039dc0.laresprime.olares.com
   ```

   ![Bifrost endpoint in Settings](/images/manual/use-cases/bifrost-endpoint.png#bordered){width=70%}

2. When you configure a client, always append `/v1` to this Bifrost endpoint URL. For example:

   ```plain
   https://44039dc0.laresprime.olares.com/v1
   ```

:::warning
The `/v1` suffix is required for OpenAI-compatible clients. Without it, requests fail.
:::

## Route models to OpenCode

In OpenCode, register Bifrost as a custom provider and add your example models (from Ollama and the single-model app) under it.

### Step 1: Connect OpenCode to Bifrost

1. Open OpenCode, and then go to **Settings** > **Providers** > **Custom provider** > **Connect**.

   <!--![Custom provider in OpenCode](/images/manual/use-cases/bifrost-opencode-custom-provider.png#bordered)-->

2. Enter the following details:
   - **Provider ID**: A unique identifier. For example, `olares-bifrost`.
   - **Display name**: The name shown in the provider list. For example, `Olares Bifrost`.
   - **Base URL**: Paste the Bifrost endpoint URL with `/v1` appended.

3. Add one row per model. Click **Add model** to insert more rows as needed, and specify each row as follows:
   - **Model ID**: Use the format `ollama/<model-name>`, where `<model-name>` is the exact model name on the backend.
     - For an **Ollama model**, use the name shown in Ollama. For example, `ollama/llama3.1:8b`.
     - For a **single-model app**, use the model name shown on the app page. For example, `ollama/qwen3.5:9b`.
         ![Model name on the model app page](/images/manual/use-cases/litellm-model-name.png#bordered){width=55%}         
   - **Display name**: Any friendly label, such as `Llama 3.1 8B` or `Qwen3.5 9B`.
         ![Add models in OpenCode](/images/manual/use-cases/bifrost-opencode-add-model.png#bordered){width=70%}

   :::warning
   - You must append `/v1` to the Bifrost URL. Without it, OpenCode returns an error.
   - You must include the `ollama/` prefix on model IDs. Without it, API calls fail.
   - The model name you enter must exactly match the name of the downloaded model in your Ollama instance. To find the exact names of your downloaded models, run `ollama list` in the Ollama terminal.
   :::

4. Click **Submit**. The message "Olares Bifrost connected" is displayed.
5. Return to OpenCode, and then go to **Settings** > **Models** > **Olares Bifrost**.
6. Verify the models you added are enabled.

   ![Added models enabled in OpenCode](/images/manual/use-cases/bifrost-opencode-add-model-enabled.png#bordered){width=70%}

### Step 2: Chat and verify

1. Start a new session in OpenCode, and select one of the Bifrost-managed models to begin a chat.

   ![Chat in OpenCode](/images/manual/use-cases/bifrost-opencode-chat.png#bordered)

2. Open Bifrost, and then go to **Observability** > **LLM Logs**.

   Each request you send appears as a log entry, which confirms that Bifrost routes the traffic successfully.

   ![Bifrost LLM logs](/images/manual/use-cases/bifrost-llm-logs.png#bordered)

## Route models to Open WebUI

In Open WebUI, add Bifrost as a direct external connection and add both example models under it.

### Step 1: Connect Open WebUI to Bifrost

1. In Open WebUI, click your user avatar, and then select **Admin Panel**.
2. Click the **Settings** tab, and then select **Connections**.
3. Enable **Direct Connection**, and then click <span class="material-symbols-outlined">add</span> on the right of **Manage OpenAI Connections**.

   ![Direct connection toggle](/images/manual/use-cases/bifrost-openwebui-direct-connection.png#bordered)

4. In the **Add Connection** window, specify the following settings:
   - **URL**: Paste the Bifrost endpoint URL with `/v1` appended.
   - **Auth**: Select **None**.
   - **Add a Model ID**: Enter each model ID in the `ollama/<model-name>` format, and then click <span class="material-symbols-outlined">add</span> to add it. For example:
     - `ollama/llama3.1:8b`
     - `ollama/qwen3.5:9b`

   ![Open WebUI connection form](/images/manual/use-cases/bifrost-openwebui-connection-form.png#bordered){width=50%}

5. Click <span class="material-symbols-outlined">refresh</span> to verify the connection, and then click **Save**.

### Step 2: Chat and verify

1. In Open WebUI, go to the **New Chat** page.
2. Select one of the configured models, and then start a conversation.

   ![Open WebUI chat](/images/manual/use-cases/bifrost-openwebui-chat.png#bordered)

3. Open Bifrost, and then go to **Observability** > **LLM Logs**.

   Each request you send appears as a log entry, which confirms that Bifrost routes the traffic successfully.

   ![Bifrost log for Open WebUI](/images/manual/use-cases/bifrost-openwebui-log.png#bordered)

## FAQs

### Use Bifrost or LiteLLM?

Olares offers multiple AI gateways. Use Bifrost if you require high request throughput, built-in MCP gateway access, semantic caching, or advanced rate limiting. For a simpler setup without these advanced features, consider using [LiteLLM](litellm.md).

### Why does OpenCode return an error when connecting to Bifrost?

Ensure you appended `/v1` to the Bifrost endpoint URL in your client configuration. Without the `/v1` suffix, requests from OpenAI-compatible clients fail.

### Why do my model calls fail even though the connection is successful?

- **Check model IDs**: You must include the `ollama/` prefix on model IDs. For example, `ollama/llama3.1:8b`.
- **Check model names**: Ensure the model name perfectly matches the name downloaded in your Ollama instance.

### Why does the AI model behave poorly as a coding agent?

Some models, such as `deepseek-r1:latest`, might not perform well as coding agents in OpenCode. If a model generates poor responses, try switching to a different model.

## Learn more

- [Download and run local AI models via Ollama](ollama.md): Install Ollama and pull models for Bifrost to route to.
- [Set up OpenCode as your AI coding agent](opencode.md): Full OpenCode setup and project workflow.
- [Chat with local LLMs using Open WebUI](openwebui.md): Configure Open WebUI against your Olares-hosted models.
- [Use LiteLLM as a unified AI model gateway](litellm.md): Compare with Bifrost to choose the right gateway for your stack.
- [Bifrost official documentation](https://docs.getbifrost.ai): Full reference for providers, MCP, caching, and governance features.
