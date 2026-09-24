---
connectionVersion: "1.12.7"
connectionLatestPath: /manual/best-practices/connect-ai-apps
outline: [2, 3]
description: Connect an AI client to Olares Router. Get the correct Base URL, choose a model name, configure credentials, and test the connection.
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI apps, default-chat, OpenAI-compatible API, Base URL, API key
---

# Connect AI apps through Olares Router

<VersionRouteSelect />

Use this guide when you need to configure an AI client, such as LobeHub or OpenCode, to call a model through Olares Router. It covers the connection values that most clients request: API format, Base URL, model name, and API key.

This guide applies to Olares 1.12.7 and later and uses Qwen3.8-27B (llama.cpp) as the chat model. For the exact settings in a particular client, follow its [app-specific tutorial](#app-specific-tutorials).

:::info Looking for Router concepts?
See [Use Olares Router as your AI gateway](/use-cases/olares-router.md) for its architecture, capability categories, authentication model, and naming rules. This page focuses only on connecting a client.
:::

## Before you begin

1. Install your AI client and Qwen3.8-27B (llama.cpp) from Market.
2. Open Router from Launchpad. On **LLM**, confirm that Qwen3.8-27B (llama.cpp) shows **Callable**. If it is unavailable, check the reason shown below its status.
3. On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model.

Setting the default determines where `default-chat` requests are routed. It does not start a stopped model.

## Choose an API format

The client's **Provider** or **Engine** setting determines the API format it uses. Choose a format supported by both the client and Router.

- For an OpenAI-compatible connection, look for **Custom Provider**, **Custom Endpoint**, **OpenAI**, or **OpenAI-Compatible**.
- Choose **Ollama** only when the app-specific tutorial or Router connection example uses the Ollama API. An Ollama backend does not require every client to use the Ollama format.
- For another API or a tool-specific integration, follow the app-specific tutorial. It might use the tool app's own endpoint instead of Router.

Cloud provider credentials are stored in Router. Selecting **OpenAI** in a client to use its API format does not mean that you need to enter an OpenAI API key in the client.

## Get the Router Base URL

1. In Router, open the relevant capability page, find the model, and click its **View connection example** icon. For Qwen3.8-27B, open **LLM**.

   ![View the Qwen3.8-27B connection example](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. In **How to call this model**, select the tab that matches the client's location:

   | Client location | Tab |
   | --- | --- |
   | An app installed in Olares | **Apps in Olares** |
   | A computer or device on the same local network | **Devices in LAN** |
   | A device connecting from outside the local network | **Remote** |

   ![Router connection details for apps in Olares](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. Copy the **Base URL** from that tab. Use the address shown on your device. The screenshot contains an example address.

Keep the path required by the client's API format. OpenAI-compatible clients usually need the trailing `/v1`. If a client appends `/v1` itself, enter the Router root URL instead. For example, Claude Code's `ANTHROPIC_BASE_URL` uses the root URL because its SDK appends `/v1/messages`.

## Choose a model name

For general chat and agent clients, use `default-chat`. It routes requests to the default chat model selected in Router. You can change that model later without updating clients that use this name.

`default-chat` does not appear in the model-list API. Add it manually if the client supports custom model IDs. If the client only allows models returned by the API, copy the full model name from **How to call this model**, including its provider prefix. For the Qwen model in this guide, the full name is `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`.

Use the appropriate model or default system name for non-chat capabilities. For example, search can use `default-search` after a default search provider is configured. `default-chat` cannot replace an embedding, speech, or search model.

## Set the API key

The required credential depends on where the request originates:

| Caller | What to enter |
| --- | --- |
| An app in Olares | Leave the field empty. If the client requires a value, enter a placeholder such as `olares`. |
| A signed-in Olares user making a request through the platform | No additional Router API key is required. |
| A client on the LAN or internet | Create a key on Router's **API keys** page and enter it in the client. A VPN connection alone does not provide an Olares user identity. |

Cloud provider keys belong in Router's provider configuration, not in clients that connect through Router.

## Check the configured context size {#check-context-window}

Some clients, such as Hermes, ask for a context size. Check the active model configuration in Router:

<!--@include: ../../reusables/ai-service-connections.md#model-context-window-->

## Configure and test the client

Open the client's provider or model settings and enter:

| Client setting | Value |
| --- | --- |
| Provider or API format | The format selected in [Choose an API format](#choose-an-api-format) |
| Base URL | The Router URL for the client's location, including the path expected by the client |
| Model name or model ID | `default-chat` or the full model name copied from Router |
| API key | Empty or a placeholder for apps in Olares. Use a Router-issued key for external clients. |

1. Save the settings and run the client's connection test if available.
2. Send a short request. For a chat model, start a new conversation and ask a simple question.
3. Open **Usage** in Router and confirm that the request reached the expected model.

## Fix common connection errors

| Symptom | What to check |
| --- | --- |
| `default-chat` is missing from the model list | Add it manually. The model-list API returns individual models, not default system names. |
| The default model is unavailable | Check which model is selected on **Default models**, then inspect its status on **LLM**. Setting a default does not start a stopped model. |
| The client reports "Model not found" | Use `default-chat`, or copy the full model name from Router, including its provider prefix. |
| Authentication fails | External clients need a valid Router-issued API key with access to the requested model. |
| The URL is unreachable or returns 404 | Copy the URL from the tab that matches the client's location. Check whether the client expects `/v1` in its Base URL. |
| A browser reports a CORS error or opens an Olares login page | Check whether the client sends requests from its server or directly from the browser. Follow its tutorial for the correct request mode. |

## Connect other app services

Some clients connect to an app's own API, such as a workflow server, gateway, or document processor. Use the endpoint and credentials in that app's tutorial. Router Base URLs and `default-chat` apply only to capabilities called through Router.

Tools such as SearXNG and Firecrawl must be registered in Router before a client can call them through the gateway. For a complete search setup example, see [Run a deep research task with Lares](/use-cases/lares.md#run-a-deep-research-task).

## App-specific tutorials

- [Build your local AI agent with LobeHub](/use-cases/lobechat.md)
- [Set up Open WebUI for local AI chat](/use-cases/openwebui.md)
- [Customize your local AI assistant using Dify](/use-cases/dify.md)
- [Set up OpenCode as your AI coding agent](/use-cases/opencode.md)
- [Write code using Claude Code](/use-cases/claude-code.md)

## Learn more

- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Understand Router's architecture, capabilities, identities, and naming rules.
- [Run local LLMs with Engine Base apps](/use-cases/llm-base-apps.md): Deploy and manage local inference engines.
- [Manage application entrances](../olares/settings/manage-entrance.md): Find service endpoints and configure access policies.
