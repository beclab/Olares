---
outline: [2, 3]
title: Use Olares Router as your AI gateway
description: Learn what Olares Router is and how to call LLM, audio, creative, and tool capabilities through one OpenAI-compatible endpoint, with the right credential for each caller.
head:
  - - meta
    - name: keywords
      content: Olares, Router, AI gateway, OpenAI-compatible API, model endpoint, API key, multimodal
---

# Use Olares Router as your AI gateway

Olares Router is the AI gateway built into Olares, shipping with v1.12.7. It exposes every AI capability on your Olares, local or cloud, through a single OpenAI-compatible endpoint. Clients never connect to local models or cloud vendors directly: they connect to Router, and Router routes every request to the right backend.

## Why Olares Router

Many users add a standalone LLM proxy such as LiteLLM when they need to connect applications to multiple AI service providers. With Router, that extra layer is no longer needed. Router solves the same core problem from inside Olares and goes further: 

- **Built in, not an add-on**: Router arrives with Olares itself. It brings AI access and management into the same system that already runs your applications and local AI workloads. A LiteLLM-style proxy is one more component to deploy, configure, and maintain. 
- **Native visibility into local AI**: Router discovers and synchronizes installed model and utility apps automatically. A proxy requires every model to be registered before it can be used.
- **Olares-aware access**: Apps running inside Olares can use their platform identity instead of storing an API key. Clients outside Olares use Router-issued API keys, while cloud-provider credentials remain stored inside Router. With a proxy, every caller is another key to issue, store, and rotate.
- **Tune in one place**: Every local model in Router opens a model card where you can edit engine arguments and tunable parameters, or restart the engine in the UI. With a proxy, launch parameters live in the model server's own config, and every experiment means a manual edit and restart on the server.
- **New AI capabilities, same endpoint**: New models, modalities, and AI utilities supported by Olares can be exposed through Router with nothing new to deploy. Clients use the same access layer, while Router handles the platform-side integration. A proxy can grow too, but every addition is another round of configuration.

## One gateway for every AI capability

LLM service apps and AI utility apps installed on your Olares appear in Router automatically. Cloud vendors appear once added, and are managed in the same place.

Router groups all of these AI capabilities into four categories:

| Category | What lives here | Examples |
| --- | --- | --- |
| **LLM** | Chat and response models, local or remote | DeepSeek, OpenAI, Anthropic, Google Gemini, a local Qwen model running on your own GPU |
| **Audio** | Speech recognition, synthesis, and the rest of the audio pipeline | Speech-to-text (STT), text-to-speech (TTS), voice activity detection (VAD), diarization, alignment, speaker embedding, audio enhancement, sound FX, voice changer |
| **Creative** | Image and video generation | Text-to-image, text-to-video, music generation, 3D |
| **Tools** | Everything an agent reaches for around the model | Embedding, web search, web fetch, reranking, optical character recognition (OCR), translation |

![Capability categories in Olares Router](/images/manual/use-cases/router-capabilities.png#bordered)

## How authentication works

How Router verifies a caller, and what credential the caller needs to provide, depends on where the request comes from.

| Caller location | How it works | <nobr>Credential needed</nobr> |
| --- | --- | --- |
| <nobr>**Apps in Olares**</nobr> | Every request inside Olares passes through the platform first, which verifies the caller and stamps an identity header onto the request: `X-Olares-App-ID` for apps running in Olares, or `X-BFL-USER` for signed-in users. Only the platform can add this header, so it cannot be forged. Router therefore always knows who is calling, and the caller needs to provide nothing else. | None |
| <nobr>**Devices in LAN**</nobr> | The request arrives over the LAN with no platform-stamped identity. The caller presents an API key (`Authorization: Bearer <api-key>`), and Router validates the key and attributes the call to the key's owner. | Router-issued<br>API key |
| **Remote** | A request from the internet also arrives without a platform-stamped identity. The caller presents an API key, and Router validates the key and attributes the call to the key's owner. | Router-issued<br>API key |

:::info Two kinds of API keys
The key you add for a cloud vendor stays inside Router and never reaches a caller. A caller only presents the key Router issues, and Router uses its stored credentials when it talks to the vendor.
:::

## Get connection details from Router

Connecting to Router generally takes three parameters, and Router has a source for each.

- **Base URL**: Open a capability page such as **LLM**, find the model, and click the **View connection example** icon on the right.

  ![The View connection example icon on a model row](/images/manual/use-cases/router-view-connection-examp.png#bordered)

  In the **How to call this model** window, each of the three caller locations has its own tab with the matching base URL, ready to be copied into your client.

  | Caller location | Base URL |
  | --- | --- |
  | Apps in Olares | `https://router.<your-olares-id>.olares.com/v1` |
  | Devices in LAN | <ul><li>Windows, Linux: `http://router-<your-olares-id>-olares.local/v1`</li><li>macOS: `http://router.<your-olares-id>.olares.local/v1`</li></ul> |
  | Remote | `https://router.<your-olares-id>.olares.com/v1` |  

  ![The How to call this model window](/images/manual/use-cases/router-how-to-call-model.png#bordered)

- **Model name**: Copy the model name from the **How to call this model** window. Or set a default model for each capability on the **Default models** page, and use the system name like `default-chat` instead of a specific model name.
- **API key**: Created on the **API Keys** page. Required only for callers from the LAN or the internet. Apps in Olares can use any placeholder.

## Set up a client

With the connection details ready, the examples below show you how to configure a client in each caller location.

### Apps in Olares

**Configurations**: OpenClaw, an app running on Olares. In the custom provider settings, point the model at Router, with no API key required:

- **API Base URL**: `https://router.<your-olares-id>.olares.com/v1`
- **Model ID**: `default-chat`

**Result**: The status line shows the session answering as `default-chat`.

![OpenClaw chatting through Router with default-chat](/images/manual/use-cases/router-client-connect-openclaw.png#bordered)

### Devices on the local network

**Configurations**: OpenCode running on a Mac connected to the same network as your Olares. In the configuration file `opencode.jsonc`, add a provider entry that points at the `.local` address with an API key created in Router:

```jsonc
{
  "provider": {
    "router": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Router",
      "options": {
        "baseURL": "http://router.<your-olares-id>.olares.local/v1",
        "apiKey": "<your-api-key>"
      },
      "models": {
        "default-chat": { "name": "Router-chat" }
      }
    }
  }
}
```

**Result**: OpenCode on the Mac talks to Router at the `.local` address, and each build is labeled with the display name `Router-chat` from the config. These calls also appear on the **Usage** page in Router, attributed to you.

![OpenCode chatting through Router default-chat](/images/manual/use-cases/router-client-connect-opencode.png#bordered)

### Remote clients

:::tip
In this setup, ensure that the Router entrance's **Authentication level** is set to **Public** in Olares Settings.
:::

**Configurations**: A client outside your local network, for example a laptop on a public Wi-Fi. Use the public base URL `https://router.<your-olares-id>.olares.com/v1` with an API key created in Router. The same key works from anywhere.

## Skip the client: Call with Olares CLI

Sometimes all you want is to check what is callable or try a model right now, not configure a client. Olares CLI is the fastest way there: no base URL, no API key, just your signed-in identity (`X-BFL-USER`). It takes three commands:

1. Authenticate the Olares CLI with your Olares ID. The CLI saves the address and your identity for later calls.

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

2. See everything callable on your Olares, with readiness.

   ```bash
   olares-cli router call models
   ```

    Sample output:


    ```text
    NAME                                        MODE    SUPPORTS                                                 READINESS  SERVED BY
    Olares/onnx-community/silero-vad            audio   vad                                                      ready      Olares
    Olares/unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL  chat    function_calling,parallel_function_calling,reasoning,+3  ready      Olares
    deepseek/deepseek-v4-flash                  chat    assistant_prefill,function_calling,native_streaming,+7   ready      deepseek
    deepseek/deepseek-v4-pro                    chat    assistant_prefill,function_calling,native_streaming,+7   ready      deepseek
    myfirecrawl/scrape                          scrape  -                                                        ready      myfirecrawl
    myjina/reader                               scrape  -                                                        ready      myjina
    mysearxng/search                            search  -                                                        ready      mysearxng
    myserper/search                             search  -                                                        ready      myserper
    mytavily/extract                            scrape  -                                                        ready      mytavily
    mytavily/search                             search  -                                                        ready      mytavily
    mytavily/search-advanced                    search  -                                                        ready      mytavily
    ```

3. Send a chat message.

   ```bash
   olares-cli router call chat "explain what is AI gateway in one sentence"
   ```

    Sample output:

    ```text
    [reasoning] User asks: "explain what is AI gateway in one sentence". Need answer one sentence. Need final only one sentence.

    An AI gateway is a central access point that routes, secures, manages, and monitors requests to AI models and APIs.

    unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL
    ```
