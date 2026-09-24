---
connectionVersion: "1.12.7"
connectionLatestPath: /manual/best-practices/connect-ai-apps
outline: [2, 3]
description: Connect AI apps through Olares Router. Set a default chat model, copy the Router URL, and configure model names and API keys for local or external clients.
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI apps, default-chat, OpenAI-compatible API, Base URL, API key
---

# Connect AI apps through Olares Router

<VersionRouteSelect />

Olares Router gives your AI apps one connection point for local and cloud models. Set a default chat model in Router, then configure your clients with the Router Base URL and `default-chat`. When you change the default model, those clients use the new model without changing their connection settings.

This guide covers Olares 1.12.7 and later and uses Qwen3.8-27B (llama.cpp) as the local chat model. For an app's exact fields and buttons, follow its [use case](#app-specific-tutorials).

## Before you begin

- Install your AI client app and Qwen3.8-27B (llama.cpp) from Market.
- Open Router from Launchpad. Before sending a request, confirm that the model shows **Callable** on **LLM**. If it is unavailable, check the reason shown below its status.
- On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model.

## Get the Router Base URL

1. In Router, open **LLM**, find the model, and click its **View connection example** icon.

   ![View the Qwen3.8-27B connection example](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. In **How to call this model**, select the tab for your client:

   | Client location | Tab |
   | --- | --- |
   | An app installed in Olares | **Apps in Olares** |
   | A computer or device on the same local network | **Devices in LAN** |
   | A device connecting from outside the local network | **Remote** |

   ![Router connection details for apps in Olares](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. Copy the **Base URL** from that tab. Use your own Router address; the screenshot shows an example device.

For OpenAI-compatible clients, keep the trailing `/v1`. Clients that append `/v1` themselves need the Router root URL instead. For example, Claude Code's `ANTHROPIC_BASE_URL` uses the root URL because its SDK appends `/v1/messages`. Follow the client tutorial for this field.

## Choose the model name

Use `default-chat` for general chat and agent examples. It routes requests to the default chat model configured in Router. Changing that default affects every client that uses this name, so choose a model that supports the clients' needs, such as tool calling or image input.

`default-chat` is not returned by the model-list API. If a client fetches available models, add `default-chat` manually. If it only allows selecting a listed model, choose the full model name shown in Router instead.

To keep a client on a specific model, copy the full model name from **How to call this model**, including its provider prefix. For the model shown above, it is `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`.

For embeddings, speech, or other capabilities, use a model or default route for that capability. `default-chat` cannot replace an embedding model. Keep the embedding model consistent when creating and searching a knowledge base.

## Check the configured context size {#check-context-window}

Some clients, such as Hermes, ask you to enter the context size manually. Check the model configuration in Router:

<!--@include: ../../reusables/ai-service-connections.md#model-context-window-->

## Set the API key

Router authenticates the caller, so the key requirement depends on where the request comes from:

| Caller | What to enter |
| --- | --- |
| An app in Olares | No Router API key is needed. Leave the field empty where possible. If the client requires a value, use a placeholder such as `olares`. |
| A signed-in Olares user making a request through the platform | The platform supplies the user identity. No additional Router API key is needed. |
| An external client on the LAN or internet | Create a key on Router's **API keys** page and enter that key in the client. Connecting through a VPN alone does not supply an Olares user identity. |

Cloud provider keys belong in Router's provider configuration. Clients connecting through Router use the caller credentials described above.

## Configure and test the client

Open the client's provider or model settings and enter:

| Client setting | Value |
| --- | --- |
| Provider or API format | **OpenAI-compatible** for most clients; use the format required by the app-specific tutorial |
| Base URL | The Router URL for the client's location, with the path suffix expected by that client |
| Model name or model ID | `default-chat`, added manually if necessary |
| API key | Empty or a placeholder for apps in Olares; a Router-issued key for external clients |

Save the settings, run the client's connection test, and send a short message. In Router, check **Usage** to confirm that the request reached the expected model.

## Fix common connection errors

| Symptom | What to check |
| --- | --- |
| `default-chat` is missing from the model list | Add it manually. The list contains individual models, not default routes. |
| The default model is unavailable | Check **Default models** and confirm that the selected chat model is **Callable** on **LLM**. |
| Model not found | Use `default-chat`, or copy the full model name from Router, including the provider prefix. A raw engine model name may not identify the correct Router provider. |
| Authentication failed | External clients need a Router-issued API key. Check that the key is valid and allows the requested model. |
| The URL is unreachable or returns 404 | Copy the URL from the correct connection tab. Check whether the client expects `/v1` in its Base URL. |
| A browser reports a CORS error or opens an Olares login page | Check whether the client sends requests from its server or directly from the browser. Follow its tutorial for the correct request mode. |

## Connect other app services

Some tutorials connect to an app's own API, such as a gateway, workflow server, or document processor. For those connections, use the endpoint and authentication instructions in that app's tutorial. `default-chat` applies only to chat requests through Router.

## App-specific tutorials

- [Build your local AI agent with LobeHub](/use-cases/lobechat.md)
- [Set up Open WebUI for local AI chat](/use-cases/openwebui.md)
- [Customize your local AI assistant using Dify](/use-cases/dify.md)
- [Set up OpenCode as your AI coding agent](/use-cases/opencode.md)
- [Write code using Claude Code](/use-cases/claude-code.md)

## Learn more

- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Learn about capabilities, caller identities, and model naming.
- [Run local LLMs with Engine Base apps](/use-cases/llm-base-apps.md): Deploy and manage local inference engines.
