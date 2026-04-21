---
outline: [2, 3]
description: Set up Bifrost on Olares as an AI gateway. Aggregate Ollama and single-model apps behind one endpoint, then connect clients like OpenCode and Open WebUI.
head:
  - - meta
    - name: keywords
      content: Olares, Bifrost, AI gateway, LLM proxy, Ollama, OpenCode, Open WebUI, self-hosted
app_version: "1.0.5"
doc_version: "1.0"
doc_updated: "2026-04-20"
---

# Use Bifrost as an AI model gateway

Bifrost is an AI gateway that sits between your client apps and multiple model providers (such as OpenAI, Anthropic, and local engines like Ollama). It exposes a single OpenAI-compatible endpoint and routes each request to the right backend based on the model name.

Olares offers multiple AI gateways. Bifrost is a good fit when you need any of the following:
- High request throughput with minimal added latency.
- A built-in MCP gateway to give models access to external tools.
- Response caching based on semantic similarity, not exact match.
- Automatic fallback across providers when one is unavailable.
- Per-key, per-team, or per-customer budgets and rate limits.

For a simpler setup, consider a lighter-weight gateway such as [LiteLLM](litellm.md).

## Learning objectives

In this guide, you will learn how to:

- Install Bifrost.
- Add Ollama or a single-model app as a model provider in Bifrost.
- Route models from Bifrost to OpenCode.
- Route models from Bifrost to Open WebUI.
- Verify model connections using Bifrost's observability logs.

## Prerequisites

- [Ollama installed](ollama.md) on Olares with at least one model downloaded. This tutorial uses `llama3.1:8b` as an example.
- At least one single-model app installed from Market. This tutorial uses the **Qwen3.5 9B Q4_K_M (Ollama)** app as an example.
- Olares admin privileges.

## Install Bifrost

1. Open Market and search for "Bifrost".

   ![Bifrost in Market](/images/manual/use-cases/bifrost.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Register providers in Bifrost

A Bifrost provider points at one backend URL. Connect an Ollama app once to route every model inside it, or connect a single-model app to expose just that model.

Both flows below use **Ollama** as the provider type, since the model apps run on the Ollama engine.

### From an Ollama app

This flow routes every model in your Ollama instance through Bifrost.

1. Open **Settings**, go to **Applications** > **Ollama** > **Entrances** > **Ollama API**, and then copy the endpoint URL. For example:

   ```plain
   https://a5be22681.laresprime.olares.com
   ```

   ![Ollama endpoint in Settings](/images/manual/use-cases/bifrost-ollama-endpoint.png#bordered){width=80%}

2. Open Bifrost from the Launchpad, go to **Models** > **Model Providers** > **Add provider**, and then select **Ollama**.

   ![Select Ollama as provider](/images/manual/use-cases/bifrost-add-provider-ollama.png#bordered)

3. Click **Edit Provider Config** in the upper-right corner.
4. In **Base URL**, enter the Ollama endpoint URL you copied.

   ![Edit provider config for Ollama](/images/manual/use-cases/bifrost-config-provider-ollama.png#bordered)

5. Click **Save Network Configuration**. The message "Provider configuration updated successfully" is displayed.
6. Close the **Ollama Provider configuration** window.

### From a single-model app

Use this flow when the model runs as its own Olares app, such as **Qwen3.5 9B Q4_K_M (Ollama)**.

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

   ![Edit provider config for single-model app](/images/manual/use-cases/bifrost-single-model-config.png#bordered)

5. Click **Save Network Configuration**. The message "Provider configuration updated successfully" is displayed.
6. Close the **Ollama Provider configuration** window.

## Get the Bifrost endpoint

Clients connect to Bifrost through its own endpoint URL, not the backend provider URLs you configured above.

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

## Use Bifrost with OpenCode

In OpenCode, register Bifrost as a custom provider and declare both example models under it: one from Ollama and one from a single-model app.

### Connect OpenCode to Bifrost

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
   - Always append `/v1` to the Bifrost URL. Without it, OpenCode returns an error.
   - The `ollama/` prefix on model IDs is required. Without it, calls fail.
   - The model name must match the name downloaded in Ollama. You can verify available models by running `ollama list` in the Ollama terminal.
   :::

4. Click **Submit**. The message "Olares Bifrost connected" is displayed.
5. Return to OpenCode, and then go to **Settings** > **Models** > **Olares Bifrost**.
6. Ensure the models you just added are enabled.

   ![Added models enabled in OpenCode](/images/manual/use-cases/bifrost-opencode-add-model-enabled.png#bordered){width=70%}

### Chat and verify

1. Start a new session in OpenCode and select one of the Bifrost-managed models to begin chatting.

   ![Chat in OpenCode](/images/manual/use-cases/bifrost-opencode-chat.png#bordered)

2. Return to Bifrost, and then go to **Observability** > **LLM Logs**. Each request you send appears as a log entry, which confirms Bifrost is routing the traffic.

   ![Bifrost LLM logs](/images/manual/use-cases/bifrost-llm-logs.png#bordered)

:::tip
Some models behave poorly as coding agents in OpenCode (for example, `deepseek-r1:latest`). If a model does not work well, try a different one.
:::

## Use Bifrost with Open WebUI

In Open WebUI, add Bifrost as a direct external connection and register both example models under it.

### Connect Open WebUI to Bifrost

1. Open Open WebUI, click the model selector in the upper-left corner, and then select **Manage Connections**.

   <!-- ![Manage connections in Open WebUI](/images/manual/use-cases/bifrost-openwebui-manage-connections.png#bordered) -->

2. Go to **Settings** > **External Connection**, enable **Direct Connection**, and then click <span class="material-symbols-outlined">add</span> next to **Manage OpenAI interface connections**.

   <!-- ![Direct connection toggle](/images/manual/use-cases/bifrost-openwebui-direct-connection.png#bordered) -->

3. In the connection form, enter the following details:
   - **URL**: Paste the Bifrost endpoint URL with `/v1` appended.
   - **Auth**: Select **None**.
   - **Model IDs**: Enter each model ID in the `ollama/<model-name>` format, then click <span class="material-symbols-outlined">add</span>. For example:
     - `ollama/llama3.1:8b` (from Ollama)
     - `ollama/qwen3.5:9b` (from a single-model app)

   <!-- ![Open WebUI connection form](/images/manual/use-cases/bifrost-openwebui-connection-form.png#bordered) -->

4. Click the refresh icon to verify the connection, and then click **Save**.

### Chat and verify

1. In Open WebUI, select one of the models you just added, and then start a conversation.

   <!-- ![Open WebUI chat](/images/manual/use-cases/bifrost-openwebui-chat.png#bordered) -->

2. Return to Bifrost and check **Observability** > **LLM Logs** to confirm the request was routed through Bifrost.

   <!-- ![Bifrost log for Open WebUI](/images/manual/use-cases/bifrost-openwebui-log.png#bordered) -->

## Learn more

- [Download and run local AI models via Ollama](ollama.md): Install Ollama and pull models for Bifrost to route to.
- [Set up OpenCode as your AI coding agent](opencode.md): Full OpenCode setup and project workflow.
- [Chat with local LLMs using Open WebUI](openwebui.md): Configure Open WebUI against your Olares-hosted models.
- [Use LiteLLM as a unified AI model gateway](litellm.md): Compare with Bifrost to choose the right gateway for your stack.
- [Bifrost official documentation](https://docs.getbifrost.ai): Full reference for providers, MCP, caching, and governance features.
