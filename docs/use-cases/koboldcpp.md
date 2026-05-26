---
outline: [2, 3]
description: Run KoboldCpp on Olares to use local GGUF models for AI chat, image understanding, text-to-image generation, voice features, and optional API access.
head:
  - - meta
    - name: keywords
      content: Olares, KoboldCpp, local LLM, GGUF, AI inference, text generation, multimodal
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-05-25"
---

# Run local AI models with KoboldCpp

KoboldCpp is a lightweight AI inference server built on llama.cpp for running GGUF models locally. On Olares, it provides a web interface for local chat, image understanding, text-to-image generation, and voice features, with optional OpenAI-compatible API access.

This guide uses the default Qwen3.5-4B model for the main workflow. Switch models only when you need another GGUF model or a feature that requires a speech-capable model preset.

## Learning objectives

In this guide, you will learn how to:
- Install KoboldCpp and configure the Hugging Face token.
- Use the KoboldCpp web interface for text generation, multimodal prompts, and voice features.
- Switch to a different GGUF model when needed.
- Optionally call the KoboldCpp API from other applications.
<!-- - Configure multiuser mode for shared access. -->

## Prerequisites

- Admin privileges on your Olares device.
- A Hugging Face account and access token for downloading model files.

## Install KoboldCpp

1. Open Market and search for "KoboldCpp".

   ![KoboldCpp](/images/manual/use-cases/koboldcpp.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

## Configure the Hugging Face token

KoboldCpp downloads models from Hugging Face during its first startup. To allow this, you must provide your access token as an environment variable.

:::tip How to get a Hugging Face token
If you do not have a token, create one in your Hugging Face account settings. For detailed instructions, see [Hugging Face Access Tokens](https://huggingface.co/docs/hub/en/security-tokens).
:::

1. Open Olares Settings, then navigate to **Advanced** > **System environment variables**.
2. Find `OLARES_USER_HUGGINGFACE_TOKEN`, click <i class="material-symbols-outlined">edit_square</i> to edit the environment variable. 
3. Enter your Hugging Face token value, then click **Confirm**.

   ![Edit environment variable](/images/manual/use-cases/koboldcpp-edit-env.png#bordered)

4. Return to the System environment variables page and click **Apply** to make the change take effect.

## Start KoboldCpp for the first time

After installation, open KoboldCpp from Launchpad. 

During the first startup, KoboldCpp downloads the required model files in the background. The main service starts only after all required files finish downloading.

:::info First startup duration
The initial download might take some time depending on your network speed and disk performance. The progress screen might persist for several minutes. This is normal.
:::

Once downloads complete, the KoboldCpp web interface loads automatically. You can also verify downloaded files in **Files** at `Home/Huggingface/koboldcpp`.


## Use KoboldCpp

When KoboldCpp is ready, it opens the Lite chat interface. Use this interface to chat with the default model, attach images or files, adjust generation settings, and access optional media features. The following table summarizes the main interface areas.

| Area | Purpose |
|---|---|
| **Top navigation bar** | Access global actions, model management, session<br> management, and settings. |
| **Main conversation area**| Check the current runtime status and view the conversation. |
| **Conversation toolbar** | Manage the current conversation and attach files or images. |
| **Input area** | Enter prompts and control editing or grouping behavior. |

![KoboldCpp interface](/images/manual/use-cases/koboldcpp-interface.png#bordered)

### Chat with the default model

The main interface loads with the default Qwen3.5-4B model ready for conversation.
1. Type your prompt in the input box at the bottom of the screen.
2. Click **Submit** to send the message, or press **Enter**.

Use the top navigation bar for session-level actions:

| Control | What it does |
|:--------|:-------------|
| **New Session** | Start a new conversation. |
| **Scenarios** | Load a preset scenario, such as role-play or Q&A. |
| **Save / Load** | Export the current session or import a previous one. |
| **Settings** | Adjust generation options, such as temperature, context size, <br>and other model parameters. |

Use the toolbar above the input box to manage the current conversation:

| Control | What it does |
|:--------|:-------------|
| **Context** | View or edit the conversation context. |
| **Undo** / **Redo** | Revert or restore recent changes. |
| **Retry** | Regenerate the latest response. |
| **Branch** | Create an alternative conversation path from the current point. |
| **Add File** | Attach a file to the conversation. |

![KoboldCpp lite chat](/images/manual/use-cases/koboldcpp-lite-chat.png#bordered)

### Analyze images with multimodal prompts

KoboldCpp supports image understanding when the required multimodal projection model, or mmproj, is loaded.

1. Click **Add File** in the toolbar above the input box.
2. Select **Upload a File**.
3. Enter a text prompt about the image, then send.

The model processes both the image and your text to generate a combined response.

 ![Analyze image](/images/manual/use-cases/koboldcpp-analyze-image.png#bordered)

### Generate images from text

KoboldCpp includes a text-to-image interface powered by Stable Diffusion.

1. Append `/sdui/` to your KoboldCpp URL. For example, if your KoboldCpp address is `https://example.olares.com`, navigate to `https://example.olares.com/sdui/`.
2. Enter a prompt and adjust generation parameters.
3. Click the generate button to create the image.

   ![Generate image](/images/manual/use-cases/koboldcpp-generate-image.png#bordered)

The image generation model (`picX_real`) and its runtime parameters inherit from the global KoboldCpp startup configuration. No additional setup is required.

### Use voice input and output

KoboldCpp supports voice input, text-to-speech output, and audio transcription.

| Feature | Use it to | Requirement |
|:--------|:----------|:------------|
| Voice input with STT | Dictate a message during a chat | Default setup |
| Text-to-speech output | Read model responses aloud | Default setup |
| Speech recognition | Transcribe uploaded or recorded audio | Speech-capable model preset |

**To use voice input with STT:**

1. Click **Settings** > **Media**.
2. In **Audio Input**, select **Toggle-To-Talk**, then click **OK**.

   ![Input audio](/images/manual/use-cases/koboldcpp-audio-input.png#bordered){width=90%}

3. Back in the input area, click the microphone button to start recording. Click it again to send the voice input.

**To use text-to-speech output:**

1. Click **Settings** > **Media**.
2. In **Audio Output**, select **KoboldCPP TTS API**, then click **OK**.

   ![Output audio](/images/manual/use-cases/koboldcpp-audio-output.png#bordered){width=90%}

This feature uses the Qwen3-TTS model and continues to occupy the GPU until audio generation completes.

**To transcribe audio with speech recognition:**

Speech recognition for uploaded or recorded audio requires a speech-capable model preset. To use it, first [switch to another GGUF model](#switch-to-another-gguf-model), then return to this section.

1. Click **Add File** in the toolbar above the input box.
2. Select **Microphone** to record audio, or select **Upload a File** to upload an audio file.
3. Enter a text prompt for the transcription task, then send.

   ![Add audio file](/images/manual/use-cases/koboldcpp-add-audio.png#bordered){width=60%}

## Switch to another GGUF model

Use admin mode when you want to load a GGUF model that is not included by default. KoboldCpp reads model presets from `/models/admindir` inside the container. Each preset is a `.kcpps` file that tells KoboldCpp which model file to load and which runtime settings to use.

:::warning Admin mode security
The admin panel exposes configuration controls. Only enable this mode in trusted network environments, and protect access with authentication or a reverse proxy.
:::

By default, only one preset is auto-generated: `qwen3.5-4b.kcpps`, which points to the built-in Qwen3.5-4B model. To use another GGUF model, prepare the model file, create a matching `.kcpps` preset, and load it from the Admin panel.

The following example switches KoboldCpp to the `gemma-3-4b-it` model. When you use a different model, keep the workflow the same and update the model filename in the preset.

### Prepare a GGUF model file

Get the model file into the KoboldCpp container. Choose one of the following methods.

<tabs>
<template #Upload-via-LarePass>

1. Download the GGUF file to your local computer. For example:
   ```
   https://huggingface.co/unsloth/gemma-3-4b-it-qat-GGUF/resolve/main/gemma-3-4b-it-qat-Q4_K_M.gguf
   ```

2. Open the LarePass desktop client and upload the `.gguf` file to **Files** > `Home/Huggingface/koboldcpp`.

</template>
<template #Download-directly-inside-the-container>

1. Open Control Hub and navigate to **Browse** > **System** > `koboldcppserver-shared` > **Deployments** > `koboldcpp-engine`.
2. In the container terminal, run:
   ```bash
   wget -O /models/gemma-3-4b-it-qat-Q4_K_M.gguf \
     "https://huggingface.co/unsloth/gemma-3-4b-it-qat-GGUF/resolve/main/gemma-3-4b-it-qat-Q4_K_M.gguf"
   ```

   If `wget` is not available, install it first:
   ```bash
   apt update && apt install -y wget
   ```

</template>
</tabs>

### Create and upload a model preset

A model preset tells KoboldCpp which GGUF file to load. In most cases, you only need to update the `model_param` value so it points to your GGUF file in `/models`.

Create a `.kcpps` preset file on your local computer:

::: code-group

```bash [macOS/Linux]
cat > gemma3-4b.kcpps <<'EOF'
{
    "model_param": "/models/gemma-3-4b-it-qat-Q4_K_M.gguf",
    "port": 5001,
    "port_param": 5001,
    "host": "",
    "launch": false,
    "threads": 4,
    "contextsize": 4096,
    "gpulayers": 99,
    "usecublas": ["normal", "0"],
    "multiuser": true,
    "skiplauncher": true
}
EOF
```

```powershell [Windows]
'@
{
    "model_param": "/models/gemma-3-4b-it-qat-Q4_K_M.gguf",
    "port": 5001,
    "port_param": 5001,
    "host": "",
    "launch": false,
    "threads": 4,
    "contextsize": 4096,
    "gpulayers": 99,
    "usecublas": ["normal", "0"],
    "multiuser": true,
    "skiplauncher": true
}
'@ | Set-Content -Encoding utf8 ".\gemma3-4b.kcpps"
```

:::

Then upload the `.kcpps` file through LarePass to **Files** > `Home/Huggingface/koboldcpp/admindir`.

### Load the new model preset

1. Open KoboldCpp and click **Admin** in the top navigation bar.
2. In the **Select New Model or Config** field, choose the `.kcpps` file you uploaded.
3. Click **Reload KoboldCpp** to load the new model.

:::tip About base config
The **Select Base Config** field is optional. If you choose a base config, KoboldCpp merges it with the new preset. Parameters defined in the new preset override matching fields in the base config. Unspecified parameters inherit from the base config.
:::

## Optional: Call KoboldCpp through the API

KoboldCpp provides OpenAI-compatible API endpoints under its app entrance URL. Use this optional section when you want to connect KoboldCpp to scripts, automation tools, or OpenAI-compatible clients.

To find the API base URL, open Olares Settings and go to **Applications** > **KoboldCpp**. Copy the URL that matches where the request comes from:

| Scenario | Use this URL |
|:--|:--|
| Call from another app inside Olares. | **Shared entrances** |
| Call from your local computer or another external client. | **Entrances** |

The examples below use macOS/Linux shell syntax. On Windows, run them in WSL or Git Bash, or adapt the environment variable syntax for PowerShell.

1. Set `BASE_URL` to the KoboldCpp entrance URL you copied.

   ```bash
   BASE_URL="https://your-koboldcpp-entrance-url"
   ```

2. Check which model ID the API exposes.

   ```bash
   curl -sS "$BASE_URL/v1/models"
   ```

   If the command returns a model list, the API is ready to use. Use the returned model ID in chat requests. The examples below use `koboldcpp`.

3. Send a chat request.

   ```bash
   curl -sS "$BASE_URL/v1/chat/completions" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "koboldcpp",
       "messages": [
         {"role": "user", "content": "Hello, please introduce yourself in one sentence."}
       ],
       "temperature": 0.7,
       "max_tokens": 128
     }'
   ```

To receive the response token by token, send a streaming request:

```bash
curl -N "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "koboldcpp",
    "stream": true,
    "messages": [
      {"role": "user", "content": "Give me three study suggestions."}
    ]
  }'
```

<!--
## Manage multiuser mode

KoboldCpp enables multiuser mode by default. When multiple users send requests simultaneously, the server queues them instead of running them in parallel. This prevents resource exhaustion and keeps the service stable.

### Configure via the preset file

Add the `multiuser` field to your `.kcpps` preset file:

```json
{
    "model_param": "/models/gemma-3-4b-it-qat-Q4_K_M.gguf",
    "port": 5001,
    "port_param": 5001,
    "host": "",
    "launch": false,
    "threads": 4,
    "contextsize": 4096,
    "gpulayers": 99,
    "usecublas": ["normal", "0"],
    "multiuser": true,
    "skiplauncher": true
}
```

- `true`: Enable multiuser mode. The server queues requests automatically.
- `false`: Disable multiuser mode. Only one request runs at a time.

### Configure via the command line

When starting KoboldCpp manually, pass the `--multiuser` flag:

```bash
# Enable with default queue behavior
./koboldcpp --model /models/xxx.gguf --multiuser

# Limit the queue to 10 requests
./koboldcpp --model /models/xxx.gguf --multiuser 10

# Disable multiuser mode
./koboldcpp --model /models/xxx.gguf --multiuser 0
```

### Verify the setting

1. Open Control Hub and navigate to **Browse** > **Shared** > `koboldcppserver-shared` > **Deployments** > `koboldcpp-engine`.
2. In the container terminal, run:
   ```bash
   ps aux | grep koboldcpp
   ```
3. Look for `--multiuser` in the output to confirm the mode is active.

:::tip Adjust the queue size for your hardware
If your GPU is powerful, you can allow a larger queue (for example, `--multiuser 10`). If you run on modest hardware, keep the queue small (for example, `--multiuser 5`) to avoid delays.
:::
-->

## Learn more

- [Download and run local AI models via Ollama](ollama.md): An alternative way to run local LLMs on Olares.
- [Chat with Local LLMs Using Open WebUI](openwebui.md): A graphical chat interface that connects to model backends.
