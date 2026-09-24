---
connectionVersion: "1.12.7"
connectionLatestPath: /manual/best-practices/connect-ai-apps
outline: [2, 3]
description: Connect apps to models and tools through Olares Router. Choose an API format, get connection details, and configure credentials for local or external clients.
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI apps, AI tools, default-chat, default-search, OpenAI-compatible API, Base URL, API key
---

# Connect your apps to AI capabilities

<VersionRouteSelect />

Olares Router connects your apps to local and cloud models, along with tools such as web search. Your client sends requests to Router, which routes them to the selected model or tool. With a default system name such as `default-chat`, you can change the backend in Router without updating each client.

This guide covers Olares 1.12.7 and later. It uses Qwen3.8-27B (llama.cpp) for the chat example and SearXNG for the search tool example. For an app's exact fields and buttons, follow its [use case](#app-specific-tutorials).

## Understand the connection concepts

### AI clients and services

An AI client provides the interface or workflow you use, such as the chat interface in LobeHub. An AI service provides a capability over an API, such as text generation, speech recognition, or web search.

On Olares, local model services include prebuilt model apps from Market and model instances created with [Engine Base apps](/use-cases/llm-base-apps.md). Tool apps such as SearXNG and Firecrawl provide search and web page content. Router makes configured capabilities available to compatible clients through one gateway.

### Provider and API format

A client's **Provider** or **Engine** setting determines which API format it uses. When connecting through Router, choose a format that both Router and the client support:

- For OpenAI-compatible connections, look for **Custom Provider**, **Custom Endpoint**, **OpenAI**, or **OpenAI-Compatible**. Set the Base URL to Router's address.
- Use **Ollama** only when following a connection example that uses the Ollama API. An Ollama backend alone does not mean the client must use this format.
- For other APIs or tools, follow the connection example in Router and the client tutorial. A tool-specific provider might require the tool app's own endpoint rather than Router's URL.

Cloud provider credentials are configured in Router. Selecting **OpenAI** in the client to use its API format does not mean you need to enter an OpenAI API key there.

### Connection parameters

| Parameter | What it controls | Where to get it |
| --- | --- | --- |
| Base URL | The Router address the client sends requests to | **How to call this model**, using the tab for the client's location |
| Model name | The model or capability Router should call | Copy the full name from **How to call this model**, or use a default system name configured on **Default models** |
| API key | The caller's identity and access | **API keys** for external callers. Apps in Olares do not need a Router-issued key. |

Router organizes capabilities under **LLM**, **Audio**, **Creative**, and **Tools**. A capability can have a model name even when it is a tool. For example, a SearXNG provider named `localsearxng` exposes search as `localsearxng/search`.

## Prepare the model or tool

### Prepare a chat model

1. Install your AI client and Qwen3.8-27B (llama.cpp) from Market.
2. Open Router from Launchpad. On **LLM**, find the model and check its status. Before sending a request, it must show **Callable**. If it is unavailable, check the reason shown below its status.
3. On **Default models**, set Qwen3.8-27B (llama.cpp) as the default chat model to use `default-chat` in this example.

Setting the default selects which model receives requests to `default-chat`. It does not start a stopped model.

### Register a tool app

Some tools must be added as providers in Router before clients can use them through the gateway. For SearXNG:

1. Install SearXNG from Market.
2. Open Settings and go to **Applications** > **SearXNG** > **Entrances**. Copy the **Endpoint** URL for the service and confirm that its access policy allows requests from other Olares apps.
3. Open Router and go to **Tools** > **Manage providers**.
4. Select **SearXNG**, then enter:

   | Setting | Value |
   | --- | --- |
   | **Provider name** | `localsearxng` |
   | **SearXNG instance URL** | The endpoint copied from Settings |

5. Click **Add**. In the **Available** list, click the tool's add icon to enable it.

   ![Enable SearXNG in Router](/images/manual/use-cases/router-search-tool-enable.png#bordered)

6. Check that the tool appears in the **Configured** list.

   ![SearXNG in the Configured list](/images/manual/use-cases/router-search-tool-enabled.png#bordered)

For a client that calls search through Router, use `localsearxng/search`, or set a default search model on **Default models** and use `default-search`. The client must support Router's search API. For using search in Lares, see [Run a deep research task](/use-cases/lares.md#run-a-deep-research-task).

## Get the Router Base URL

1. In Router, open the capability page, find the model or tool, and click its **View connection example** icon. For the Qwen example, use **LLM**. For SearXNG, use **Tools**.

   ![View the Qwen3.8-27B connection example](/images/manual/use-cases/router-view-connection-examp.png#bordered)

2. In **How to call this model**, select the tab for your client:

   | Client location | Tab |
   | --- | --- |
   | An app installed in Olares | **Apps in Olares** |
   | A computer or device on the same local network | **Devices in LAN** |
   | A device connecting from outside the local network | **Remote** |

   ![Router connection details for apps in Olares](/images/manual/use-cases/router-how-to-call-model.png#bordered)

3. Copy the **Base URL** from that tab. Use your own Router address. The screenshot shows an example device.

Copy the path shown for the API you are using. For OpenAI-compatible chat clients, keep the trailing `/v1`. Clients that append `/v1` themselves need the Router root URL instead. For example, Claude Code's `ANTHROPIC_BASE_URL` uses the root URL because its SDK appends `/v1/messages`. Follow the client tutorial for this field.

## Choose the model name

Use `default-chat` for general chat and agent examples. It routes requests to the default chat model configured in Router. Changing that default affects every client that uses this name, so choose a model that supports the clients' needs, such as tool calling or image input.

`default-chat` is not returned by the model-list API. If a client fetches available models, add `default-chat` manually. If it only allows selecting a listed model, choose the full model name shown in Router instead.

To keep a client on a specific model, copy the full model name from **How to call this model**, including its provider prefix. For the model shown above, it is `Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL`.

For search, embeddings, speech, or other capabilities, use a model or default system name for that capability, such as `default-search` for search. `default-chat` cannot replace an embedding model. Keep the embedding model consistent when creating and searching a knowledge base.

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
| Provider or API format | A format supported by Router and the client. See [Provider and API format](#provider-and-api-format). |
| Base URL | The Router URL for the client's location, with the path suffix expected by that client |
| Model name or model ID | The full capability name, or a configured default system name such as `default-chat` for chat |
| API key | Empty or a placeholder for apps in Olares. Use a Router-issued key for external clients. |

1. Save the settings and run the client's connection test if available.
2. Send a short request for the capability you configured. For chat, start a conversation and ask a short question. For search, run a search.
3. Open **Usage** in Router and check that the request reached the expected model or tool.

## Fix common connection errors

| Symptom | What to check |
| --- | --- |
| `default-chat` is missing from the model list | Add it manually. The list contains individual models, not default routes. |
| The default model is unavailable | Check which model is selected on **Default models**, then inspect its status on the capability page. For chat, check **LLM**. Setting a default does not start a stopped model. |
| Model not found | Copy the full name from Router, including the provider prefix, or use a configured default system name for that capability. A raw engine model name might not identify the correct Router provider. |
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

- [Manage application entrances](../olares/settings/manage-entrance.md): Find service endpoints and configure access policies.
- [Use Olares Router as your AI gateway](/use-cases/olares-router.md): Learn about capabilities, caller identities, and model naming.
- [Run local LLMs with Engine Base apps](/use-cases/llm-base-apps.md): Deploy and manage local inference engines.
