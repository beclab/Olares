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

Many users run a dedicated LLM proxy such as LiteLLM, or are considering one, to deal with the multi-vendor problem. Once Router is available, that extra layer is no longer needed.

Router covers the same ground and goes further in the following ways:
- **Built in, not an add-on**: A LiteLLM-style proxy is one more component to deploy, configure, and maintain. Router arrives with Olares itself, and is simply there when you need it.
- **Local models, zero registration**: With a proxy, every model has to be registered before it can be used. With Router, the models and utility apps on your Olares are already there, ready to call.
- **Identity, not just keys**: Apps inside Olares call without ever handling a key, because the platform already knows who they are. Every other caller carries one key that works from anywhere, and every call, keyless or not, lands on a real user's record.
- **Ready for what's next**: Whatever Olares adds next, a new modality or a new tool, appears behind the same endpoint, and no client has to change. A proxy can grow too, but every addition is another round of configuration.

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

The credential a caller needs depends on where the request comes from.

Every request to Router passes through the Olares platform. When the caller runs inside Olares, the platform verifies it and stamps an identity header onto the request that only the platform can add:

- `X-Olares-App-ID` for applications running in Olares.
- `X-BFL-USER` for signed-in users calling through Olares CLI.

Because the header cannot be forged, these callers need no API key. A request from a device on the local network or from the internet carries no such stamp, so it must also present an API key issued in Router.

This gives three authentication scenarios:

| Caller | Base URL | Credential |
| :--- | :--- | :--- |
| **Apps in Olares** | `https://router.<your-olares-id>.olares.com/v1` | None. The platform stamps `X-Olares-App-ID`. |
| **Devices on the local network** | macOS: `http://router.<cluster-name>.olares.local/v1`<br>Windows and Linux: `http://router-<cluster-name>-olares.local/v1` | API key |
| **Remote clients** | `https://router.<your-olares-id>.olares.com/v1` (same URL as apps) | API key |

The pattern is simple: the closer the caller is to the cluster, the simpler the credential. Apps in Olares need nothing. Every other caller needs an API key.

## Get connection details from Router

Connecting a client generally takes three parameters, and Router has a source for each.

- **Base URL**: Open a capability page such as **LLM**, find the model, and click the **View connection example** icon on the right. In the **How to call this model** dialog, each of the three caller locations has its own tab with the matching base URL, ready to copy into your client.

  ![The View connection example icon on a model row](/images/manual/use-cases/router-view-connection-examp.png#bordered)

  ![The How to call this model dialog](/images/manual/use-cases/router-how-to-call-model.png#bordered)

- **Model name**: Copy the model name from the **How to call this model** dialog. Or set a default model for each capability on the **Default models** page, and use the system name like `default-chat` instead of a specific model name.
- **API key**: Created on the **API Keys** page. Required only for callers from the LAN or the internet. Apps in Olares can use any placeholder.

## Set up a client

With the connection details ready, the examples below show how to configure a client in each caller location.

### Apps in Olares

**Configurations**: OpenClaw, an app running on Olares. In its provider settings, point the model at Router, with no API key required:

- **API Base URL**: `https://router.<your-olares-id>.olares.com/v1`
- **Model ID**: `default-chat`

**Result**: The status line shows the session answering as `default-chat`.

![OpenClaw chatting through Router with default-chat](/images/manual/use-cases/router-client-connect-openclaw.png#bordered)

### Devices on the local network

**Configurations**: OpenCode running on a Mac connected to the same network as your Olares. Point it at the `.local` address and provide an API key created in Router. The provider entry looks like this:

```jsonc
{
  "provider": {
    "router": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "Router",
      "options": {
        "baseURL": "http://router.<cluster-name>.olares.local/v1",
        "apiKey": "<your-api-key>"
      },
      "models": {
        "default-chat": { "name": "Router-chat" }
      }
    }
  }
}
```

**Result**: OpenCode on the Mac talks to Router at the `.local` address, and each build is labeled with the display name `Router-chat` from the config. These calls also appear on the **Usage** page in Router, attributed to the caller.

![OpenCode chatting through Router default-chat](/images/manual/use-cases/router-client-connect-opencode.png#bordered)

### Remote clients

For a laptop on the go, any OpenAI-compatible client works the same way: use the public base URL `https://router.<your-olares-id>.olares.com/v1` with a real API key. The same key works from anywhere.

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

3. Send a chat message. The last output line names the model that answered.

   ```bash
   olares-cli router call chat "explain what is AI gateway in one sentence"
   ```

    Sample output:

    ```text
    [reasoning] User asks: "explain what is AI gateway in one sentence". Need answer one sentence. Need final only one sentence.

    An AI gateway is a central access point that routes, secures, manages, and monitors requests to AI models and APIs.

    unsloth/Qwen3.8-27B-GGUF:UD-Q4_K_XL
    ```
