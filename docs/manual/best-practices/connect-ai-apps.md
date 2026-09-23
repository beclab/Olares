---
outline: [2, 3]
description: Connect your apps to AI capabilities on Olares through Router, the AI gateway for models and tools, by copying the Base URL, entering the model name, and creating an API key for external callers.
head:
  - - meta
    - name: keywords
      content: Olares, AI apps, Router, AI gateway, Model Console, LLM service, Base URL, Ollama
---

# Connect your apps to AI capabilities <Badge type="tip" text="^ 1.12.7" />

Starting with 1.12.7, client applications on Olares access AI capabilities through Router, the Olares-native AI gateway for models and tools. Instead of connecting a client app directly to individual model endpoints, you connect it to Router, and Router routes every request to the corresponding backend. When the backend model changes, the client configuration stays the same.

This page explains the basic connection concepts and how to connect a client app to a model through Router.

This guide explains how to understand your apps, gather the required connection details, and configure the connection.

## Learning objectives

By the end of this tutorial, you will be able to:
- Distinguish between AI service apps and AI client apps.
- Understand the essential connection parameters and how to get them.
- Connect common AI apps.

## Understand the connection concepts

### AI service apps and AI client apps

When you use AI on Olares, you typically work with two separate applications:

- **AI client apps**: They provide the front-end chat interface or workflow canvas you interact with directly, such as LobeHub. They rely on an AI service app (often called a provider) to perform AI tasks, such as generating text.
- **AI service apps**: They provide AI capabilities for compatible clients over an API, such as chat, search, and speech recognition. Some AI service apps have their own web interface for management, while others run primarily as headless backend services.

    On Olares, AI service apps fall into two categories:
    - **LLM service apps**: Apps that host large language models (LLMs) for text generation, code completion, and chat. They include eight pre-built model apps, and custom model instances created on [Engine Base apps](/use-cases/llm-base-apps.md).
    - **Other AI service apps**: Utility apps that provide non-LLM functions, such as web search.

### Provider

In most AI client apps, a provider (or engine type) is the service or vendor that supplies the AI capability, such as OpenAI or Ollama. On Olares 1.12.7 and later, every request goes to Router, so the Base URL and other connection details always come from Router no matter which provider type you choose.

The provider type determines how the client formats requests to Router. Choose the type that matches the capability you are connecting:

- **For LLM models**: Select **Ollama** for models that run on the Ollama engine. For non-Ollama models, select **Custom Provider**.
- **For other tools**: First, look for **Custom Provider** or **Custom Endpoint**. If your client does not offer those options, select **OpenAI** (or **OpenAI-Compatible**).
- **For specific tools**: Some clients offer a dedicated provider or engine type for a specific tool, such as **SearXNG**. If available, select that type.

### Base URL

The Base URL is the network address (or endpoint) where the AI service app receives and processes your tasks.

You always copy it from the **How to call this model** window in Router. Note that it varies with where the client runs. Some tools must be registered in Router first before they appear on the **Tools** page. See [Connect through Router](#connect-through-router).

### Model name

The model name is the exact identifier the client sends with every request so Router knows which model to use. Router gives every callable capability a model name, including tools such as `localsearxng/search`. You have two options:

- Copy the specific model name from the **How to call this model** window. Copy it exactly as displayed, without removing repository prefixes or quantization tags.
- Set a default model for a capability on the **Default models** page in Router, and use the system name such as `default-chat` and `default-search`. With a system name, you can switch the backend model in Router later without changing the client configuration.

### API key

An API key is a credential that proves the caller's identity. Router identifies callers in two ways:

- **Apps in Olares**: Router trusts requests from apps inside the cluster, so no API key is required. If the client app requires a value in the API key field, enter any placeholder text such as `olares`.
- **LAN and remote callers**: Callers outside the cluster must present a Router-issued API key. Create one on the **API keys** page in Router.

## Connect through Router

Router lists every callable capability as a model, whether it is a chat model on the **LLM** page or a tool on the **Tools** page. The connection flow is the same for every model. The only difference is that some tools must be registered in Router first, using the app's entrance URL.

### Register a tool app first

Tool apps provide utility capabilities over their own protocols. Some tools, such as SearXNG and Firecrawl, can be installed from Market and run on your Olares. Before such a tool appears in Router, register it using its entrance URL. The following steps use SearXNG as an example.

1. Install **SearXNG** from Market.
2. Open Olares Settings, go to **Applications** > **SearXNG** > **Entrances**, and copy the **Endpoint URL**. Make sure the entrance's **Authentication level** allows access from other apps in the cluster.
3. Open Router and go to **Tools** > **Manage providers**.
4. Select **SearXNG**, enter a **Provider name** such as `localsearxng`, paste the entrance URL into **SearXNG instance URL**, and then click **Add**.
5. Enable the tool so it appears in the **Configured** list.

Once registered, the tool appears on the **Tools** page like any other capability, and you get its connection details the same way.

### Get connection details and configure the client

Router organizes capabilities into four pages: **LLM**, **Audio**, **Creative**, and **Tools**. Each model appears as a row with a **View connection example** icon. The following steps use the Qwen 3.8 27B model as an example.

1. Open Router from the Launchpad.
2. Go to the **LLM** capability page, locate the model, and click the **View connection example** icon on the right to open the **How to call this model** window.
3. Select the connection source that matches where your client app runs: **Apps in Olares**, **Devices in LAN**, or **Remote**.
4. Copy the base URL exactly as displayed.
5. Copy the model name exactly as displayed, or use a system name such as `default-chat` if you have set a default model on the **Default models** page.
6. If the client app runs outside Olares on the LAN or the internet, go to the **API keys** page in Router, create an API key, and copy it.
7. In the client app, open the provider settings and enter the collected values:

   | Client setting | Value to use |
   |---|---|
   | Provider | The type that matches the capability. See [Provider](#provider). |
   | Base URL or endpoint | The complete URL copied from the **How to call this model** window |
   | Model name or model ID | The model name or system name such as `default-chat` |
   | API key | The Router-issued API key, or a placeholder such as `olares` for apps in Olares |

:::info Connect directly through the Model Console on 1.12.6
On Olares 1.12.6 without Router, open the LLM service app from the Launchpad to launch its **Model Console**. Select the **Connection source** and **API format**, copy the **Base URL** and **Model name**, and enter a placeholder API key such as `olares` in the client app. This direct connection still works on later versions, but Router is the recommended way because it lets you switch backend models without reconfiguring clients.
:::

## Verify the connection

1. In the client app, start a new conversation, select the model you configured, and send a short question. A normal response confirms that the client can reach Router and use the model.
2. Open Router and go to the **Usage** page. The request appears in the usage records, confirming that the call went through Router.

## Fix common connection errors

| Symptom | Likely cause and fix |
|---|---|
| The client reports "Model not found" | The model name was abbreviated or missing prefixes. Copy the full model name from the **How to call this model** window. |
| The connectivity check fails or the Base URL is unreachable | The connection source does not match where the client runs. Reopen the **How to call this model** window, select the matching **Connection source**, and copy the Base URL again. |
| The client gets a 401 or unauthorized error | A caller from the LAN or the internet needs a Router-issued API key. Create one on the **API keys** page. |
| A browser reports a CORS error or Olares authentication page | The client may be sending requests from the browser instead of its server. Check the client tutorial for the correct request mode and service entrance. |

## Learn more

- [Use Olares Router as your AI gateway](../../use-cases/olares-router.md)
- [Run local LLMs with Ollama, vLLM, llama.cpp, and SGLang](../../use-cases/llm-base-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
