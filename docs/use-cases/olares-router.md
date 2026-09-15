---
outline: [2, 3]
title: Use Olares Router as your AI gateway
description: Learn what Olares Router is and how to call LLM, audio, creative, and tool capabilities through one unified access layer, with the right credential for each caller.
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI gateway, OpenAI-compatible API, model endpoint, API key, multimodal
---

# Use Olares Router as your AI gateway

Olares Router is the AI gateway built into Olares, shipping with v1.12.7. It exposes every AI capability on your Olares, local or cloud, through a single access layer. Clients never connect to local models or cloud vendors directly: they connect to Router, and Router routes every request to the right backend.

## What is Olares Router

To understand how Router simplifies your AI workflows, it helps to look at where it sits within the platform and how it unifies different AI workloads into standard capabilities.

### Where Router sits

Router functions as the central dispatch layer for your entire AI ecosystem. Rather than interacting with fragmented backends directly, all traffic flows through this single gateway.

![Olares Router architecture](/images/manual/use-cases/router-archi1.png#bordered)

As shown in the architecture above, the system operates through three core layers:
1. **Callers (The inbound request)**: Whether the request comes from an app inside Olares, a signed-in Olares user, or a third party outside it, it connects to Router. Callers never interact with the backend infrastructure directly.
2. **Router (The gateway)**: Once a request arrives, Router handles the caller's authentication, quota management, and capability routing. For external cloud models, Router uses credentials securely stored within the platform, ensuring these secrets are never exposed to the caller. 
3. **Model Backends (The outbound routing)**: Router acts as a traffic director, dispatching the request to the correct destination:
   * **Local Models:** Requests for local capabilities are routed to installed model applications (like vLLM, SGLang, or llama.cpp). Each application includes a `model_console` adapter that standardizes lifecycles and API protocols, and pulls model weights directly from the platform's shared storage.
   * **Cloud Providers:** Requests for remote models are securely routed across the internet to providers like OpenAI or Anthropic, authenticated by Router's stored credentials.

### One gateway for every AI capability

While LLMs have converged on standard APIs because a single model usually completes the task, multi-modal and functional tools lack this standardization. For example, an audio workflow typically requires separate models for voice separation, enhancement, speech recognition, and text to speech. These models come from different providers with different API structures, and there is no universal interface.

Router solves this fragmentation by acting as a universal abstraction layer for your AI ecosystem. It takes language models, audio/video models, and utility tools, and exposes them as unified system capabilities behind a single gateway. This provides one standard format for all your AI workloads.

To keep your ecosystem organized and accessible, Router groups these unified capabilities into four core categories:

| Category | What lives here | Examples |
| --- | --- | --- |
| **LLM** | Chat and response models, local or remote | DeepSeek, OpenAI, Anthropic, Google Gemini, a local Qwen model running on your own GPU |
| **Audio** | Speech recognition, synthesis, and the rest of the audio pipeline | Speech-to-text (STT), text-to-speech (TTS), voice activity detection (VAD), diarization, alignment, speaker embedding, audio enhancement, sound FX, voice changer |
| **Creative** | Image and video generation | Text-to-image, text-to-video, music generation, 3D |
| **Tools** | Everything an agent reaches for around the model | Embedding, web search, web fetch, reranking, optical character recognition (OCR), translation |

![Capability categories in Olares Router](/images/manual/use-cases/router-capabilities.png#bordered)

## Why Olares Router

Router is integrated directly into the Olares platform alongside your applications and models. It functions as a platform-level gateway that manages access, lifecycles, and multi-modal routing:

- **Identity and access**: Router utilizes the Olares identity system. Internal callers authenticate using their platform identity, and external callers use Router-issued API keys. Underlying cloud provider credentials are kept isolated within Router and are never exposed to callers.
- **Model observability**: Router syncs automatically with the Model Console to discover installed models. It provides a UI to tune or restart underlying engines and tracks all request metrics by caller on the Usage page.
- **Unified multi-modal capabilities**: Router aggregates language, audio, video models, as well as utility tools (like search and embeddings), behind a single access layer. Clients use one standard interface and can invoke any system capability by its name.

## How callers authenticate with Router

Every call through Router has an identified caller, and what the caller needs to provide depends on who that caller is.

| Caller | How Router identifies it | <nobr>Credential needed</nobr> |
| --- | --- | --- |
| <nobr>**Olares apps**</nobr> | The platform stamps an identity header `X-Olares-App-ID` onto every request from an app inside Olares, so the caller needs to provide nothing else. | None |
| <nobr>**Olares users**</nobr> | The platform stamps an identity header `X-BFL-USER` onto every request from the signed-in Olares users, so the caller needs to provide nothing else. | None |
| <nobr>**Third parties**</nobr> | A caller that is neither an Olares app nor an Olares user presents a Router-issued API key (`Authorization: Bearer <api-key>`). Router validates the key and attributes the call to its owner. Delete the key and the access will end.<br><br>This covers your own devices outside Olares and anyone you share the models with. | Router-issued<br>API key |

## Get connection details from Router

Connecting to Router takes three parameters:

- **Base URL**: Open a capability page such as **LLM**, find the model, and click the **View connection example** icon on the right.

  ![The View connection example icon on a model row](/images/manual/use-cases/router-view-connection-examp.png#bordered)

  ![The How to call this model window](/images/manual/use-cases/router-how-to-call-model.png#bordered)

- **Model name**: Copy the model name from the **How to call this model** window. Or set a default model for each capability on the **Default models** page, and use the system name like `default-chat` instead of a specific model name.
- **API key**: Created on the **API keys** page. Required only for callers from the LAN or the internet. Apps in Olares do not need to enter API keys.

## Call models through Router

The examples below show how each type of caller connects to and calls models through Router.

### Apps in Olares

OpenClaw, an app running on Olares, connects to Router in its custom provider settings. Its platform identity covers authentication, so only the base URL and a model name are needed:
- **API Base URL**: `https://router.<your-olares-id>.olares.com/v1`
- **Model ID**: `default-chat`

The session then answers as `default-chat`.

### Callers outside Olares

Anyone outside Olares needs two things: a base URL matching their location, and a Router-issued API key. The base URL differs between the LAN and the internet, but the key works from anywhere.

:::tip
For callers reaching Router over the internet, set the Router entrance's **Authentication level** to **Public** in Olares Settings. LAN access is not affected.
:::

This example uses the public URL. On a device in the same LAN, use the `.local` base URL obtained from the connection details.

```bash
curl https://router.<your-olares-id>.olares.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-api-key>" \
  -d '{"model": "default-chat", "messages": [{"role": "user", "content": "Hello"}]}'
```

## Manage models and access

Beyond routing requests, Router gives you one place to manage what is exposed and who can reach it:

- **Default models**: Map each capability to a model behind a system name like `default-chat`. Apps call the system name, so you can swap, upgrade, or relocate the backing model without touching app configurations.
- **API keys**: Issue and revoke keys for callers outside Olares. Every call is attributed to its key, so you can trace usage per caller and end access by deleting the key.
- **Usage**: See every call by caller, model, and token consumption.

Model app lifecycle, installation, engine parameters, and restarts, is managed in Model Console.
