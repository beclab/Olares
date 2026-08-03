---
outline: [2, 3]
description: Set up Bifrost on Olares as an AI gateway. Aggregate Ollama and single-model apps behind one endpoint, then connect clients like OpenCode and Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, Bifrost, AI gateway, LLM proxy, Ollama, OpenCode, Open WebUI, self-hosted
app_version: "1.0.11"
doc_version: "2.0"
doc_updated: "2026-08-03"
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

Before you begin, you need the following model:
| Model type | Model | How to get it |
| :--- | :--- | :--- |
| Chat | Qwen3.6-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install Bifrost

1. Open Market and search for "Bifrost".

   ![Bifrost in Market](/images/manual/use-cases/bifrost.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Add model providers in Bifrost

In Bifrost, a model provider represents the engine hosting your AI models. You configure a provider by supplying the endpoint URL of the application running the models. 

You can connect the Ollama application to route every model running inside it, or connect a single-model application to expose just that specific model.

In this tutorial, since both the example models run on the Ollama engine, select **Ollama** as the provider type for both scenarios.

### Get model connection details

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Configure

1. Open Bifrost from the Launchpad, go to **Models** > **Model Providers** > **Add provider**, and then select **Custom provider**.

   ![Select Custom provider](/images/manual/use-cases/bifrost-add-provider.png#bordered)

2. In the **Add Custom Provider** panel, configure the following settings:

   - **Name**: such as local-qwen36
   - **Base Format**: Select **OpenAI**.
   - **Base URL**: Enter the **Base URL** you copied from the Model Console, excluding `/v1`. For example, `https://e46e044d.laresprime.olares.com`.
   - **Allow Private Network**: Enable 
   - **Is Keyless**: Enable to indicate yes

   ![Edit custom provider config](/images/manual/use-cases/bifrost-single-model-config2.png#bordered){width=90%}

3. Click **Add**.

## Obtain the Bifrost endpoint

Client applications connect to Bifrost through the Bifrost endpoint URL, not the model provider URLs you configured earlier.

1. Open Olares **Settings**, go to **Applications** > **Bifrost** > **Entrances** > **Bifrost**, and then copy the endpoint URL. For example:

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

In OpenCode, register Bifrost as a custom provider and add your example models under it.

### Step 1: Connect OpenCode to Bifrost

1. Open OpenCode, go to **Settings** > **Providers** > **Custom provider**, and then click **Connect** on the right.
2. Enter the following details:
   - **Provider ID**: A unique identifier for the provider. For example, `olares-bifrost`.
   - **Display name**: The name shown in the providers or models list for selection. For example, `Olares Bifrost`.
   - **Base URL**: The Bifrost endpoint URL with `/v1` appended. For example, `https://44039dc0.laresprime.olares.com/v1`.
   - **model-id**: Enter the **Model name** you copied from the Model Console. For example, `unsloth/Qwen3.6-27B-GGUF:Q4_K_M`.
   - **Display Name**: Specify a friendly label to identify the model, such as `Qwen3.6 27B`.
   - To add multiple models, click **Add model**.

   ![Add models in OpenCode](/images/manual/use-cases/bifrost-opencode-add-model1.png#bordered){width=70%}

3. Click **Submit**. A message is displayed to notify that the provider is connected.
5. Go to **Settings** > **Models** > **Olares Bifrost**, and then verify the model you added is enabled.

   ![Added models enabled in OpenCode](/images/manual/use-cases/bifrost-opencode-add-model-enabled1.png#bordered){width=70%}

### Step 2: Chat and verify

1. Start a new session in OpenCode, and select the Bifrost-managed model to begin a chat.

   ![Chat in OpenCode](/images/manual/use-cases/bifrost-opencode-chat1.png#bordered)

2. Open Bifrost, and then go to **Observability** > **LLM Logs**.

   Each request you send appears as a log entry, which confirms that Bifrost routes the traffic successfully.

   ![Bifrost LLM logs](/images/manual/use-cases/bifrost-llm-logs1.png#bordered)

## Route models to Open WebUI

In Open WebUI, add Bifrost as a direct external connection and add both example models under it.

### Step 1: Connect Open WebUI to Bifrost

1. In Open WebUI, click your user avatar, and then select **Admin Panel**.
2. Click the **Settings** tab, and then select **Connections**.
3. Locate the **AI** section on the left sidebar, and then click **Connections** under it.
4. Enable **Direct Connection**, and then click <span class="material-symbols-outlined">add</span> on the right of **Manage OpenAI Connections**.

   ![Direct connection toggle](/images/manual/use-cases/bifrost-openwebui-direct-connection.png#bordered)

4. In the **Add Connection** window, specify the following settings:
   - **URL**: Paste the Bifrost endpoint URL with `/v1` appended.
   - **Auth**: Select **None**.
   - **Add a Model ID**: Expand **Advanced**, enter the model ID copied from the Model Console, and then click <span class="material-symbols-outlined">add</span>. 

   ![Open WebUI connection form](/images/manual/use-cases/bifrost-openwebui-connection-form.png#bordered){width=50%}

   - Tags: enter Bifrost

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

### Why do I get errors when calling a model through Bifrost in OpenCode?

Certain models have their own native output formats such as custom tags or reasoning blocks, or lack support for features the client expects, such as tool calling. When Bifrost routes these requests, the models might return responses that OpenAI-compatible clients like OpenCode fail to parse, resulting in failures.

If you encounter this issue:
- Review the model documentation for special output formats or capability limitations.
- Verify the model supports the specific features your client requests.
- Switch to a model that fully complies with the OpenAI API standard.

## Learn more

- [Download and run local AI models via Ollama](ollama.md): Install Ollama and pull models for Bifrost to route to.
- [Set up OpenCode as your AI coding agent](opencode.md): Full OpenCode setup and project workflow.
- [Chat with local LLMs using Open WebUI](openwebui.md): Configure Open WebUI against your Olares-hosted models.
- [Use LiteLLM as a unified AI model gateway](litellm.md): Compare with Bifrost to choose the right gateway for your stack.
- [Bifrost official documentation](https://docs.getbifrost.ai): Full reference for providers, MCP, caching, and governance features.
