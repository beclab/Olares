---
outline: [2, 3]
description: Deploy KoboldCpp on Olares to run local GGUF models with text generation, multimodal reasoning, image generation, and voice capabilities.
head:
  - - meta
    - name: keywords
      content: Olares, KoboldCpp, local LLM, GGUF, AI inference, text generation, multimodal
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-05-15"
---

# Deploy KoboldCpp for local AI inference

KoboldCpp is a lightweight AI inference server built on llama.cpp. It runs GGUF-format models locally and provides a web interface alongside OpenAI-compatible APIs.

On Olares, KoboldCpp comes pre-configured with text generation, image understanding, text-to-image, speech recognition, and text-to-speech capabilities.

## Learning objectives

In this guide, you will learn how to:
- Install KoboldCpp and configure the Hugging Face token.
- Use the KoboldCpp web interface for text generation, multimodal prompts, and voice features.
- Call the KoboldCpp API from other applications.
- Switch to a different GGUF model through the admin panel.
- Configure multiuser mode for shared access.

## Prerequisites

- Admin privileges on your Olares device.
- A Hugging Face account and access token for downloading model files.

## Install KoboldCpp

1. Open Market and search for "KoboldCpp".
2. Click **Get**, then **Install**, and wait for installation to complete.
   <!-- ![KoboldCpp](/images/manual/use-cases/koboldcpp.png#bordered) -->

## Configure the Hugging Face token

KoboldCpp downloads models from Hugging Face during its first startup. To allow this, you must provide your access token as an environment variable.

:::tip How to get a Hugging Face token
If you do not have a token, create one in your Hugging Face account settings. For detailed instructions, see [Hugging Face Access Tokens](https://huggingface.co/docs/hub/en/security-tokens).
:::

1. Open Olares Settings, then navigate to **Advanced** > **System environment variables**.
2. Click **Edit**, enter `OLARES_USER_HUGGINGFACE_TOKEN` with your Hugging Face token value, then click **Confirm**.
3. Return to the System environment variables page and click **Apply** to make the change take effect.

## Complete initial setup

After installation, open KoboldCpp from Launchpad. The app shows a download progress screen while it fetches the default model files in the background. The main service starts only after all required files finish downloading.

:::info First startup duration
The initial download might take some time depending on your network speed and disk performance. The progress screen might persist for several minutes. This is normal.
:::

Once downloads complete, the KoboldCpp web interface loads automatically. You can also verify downloaded files in **Files** at `Home/Huggingface/koboldcpp`.

## Generate text

The main interface loads with the default Qwen3.5-4B model ready for conversation.

1. Type your prompt into the input box at the bottom of the screen, then press **Enter** to send.
2. To regenerate the last response, click **Retry**. To explore alternative replies, click **Branch**.
3. To review or edit earlier messages, click **Context**. To undo a recent edit, click **Undo**.
4. To change generation parameters such as temperature or context size, click **Settings** in the top navigation bar.
5. To start a new conversation, click **New Session**. This clears the current context.

:::tip Save and load conversations
Click **Save / Load** in the top navigation bar to export your chat history or import a previous session.
:::

:::tip Load a scenario template
Click **Scenarios** in the top navigation bar to load a preset template, such as role-play or Q&A.
:::

## Analyze images with multimodal prompts

KoboldCpp supports image understanding when the required multimodal projection model (mmproj) is loaded.

1. Click **Add File** in the toolbar above the input box.
2. Select or upload an image.
3. Enter a text prompt about the image, then send.

The model processes both the image and your text to generate a combined response.

## Generate images from text

KoboldCpp includes a text-to-image interface powered by Stable Diffusion.

1. Append `/sdui/` to your KoboldCpp URL. For example, if your KoboldCpp address is `https://example.olares.com`, navigate to `https://example.olares.com/sdui/`.
2. Enter a prompt and adjust generation parameters.
3. Click the generate button to create the image.

The image generation model (`picX_real`) and its runtime parameters inherit from the global KoboldCpp startup configuration. No additional setup is required.

## Use voice input and output

KoboldCpp supports three voice-related features:

- **Speech recognition (STT)**: Click the microphone icon in the input area and select **Toggle-To-Talk** to convert speech to text during conversations.
- **Speech-to-text (Whisper)**: Upload an audio file or record directly in supported interfaces to transcribe speech.
- **Text-to-speech (TTS)**: Enable TTS in supported interfaces to have model responses read aloud. This feature uses the Qwen3-TTS model and continues to occupy the GPU until audio generation completes.

## Call the API

KoboldCpp exposes OpenAI-compatible endpoints on the same port as the web interface. You can find the endpoint URL in Olares Settings under **Applications** > **KoboldCpp** > **Shared entrances** or **Entrances**.

Example API calls:

```bash
BASE_URL="http://your-shared-endpoint.shared.olares.com"

# List available models
curl -sS "$BASE_URL/v1/models"

# Chat completion
curl -sS "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello, please introduce yourself in one sentence."}
    ],
    "temperature": 0.7,
    "max_tokens": 128
  }'

# Streaming response
curl -N "$BASE_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d '{
    "stream": true,
    "messages": [{"role":"user","content":"Give me three study suggestions."}]
  }'
```

:::tip Internal vs external access
Use the shared entrance URL for requests from other apps within Olares. Use the regular entrance URL for requests from outside your Olares network.
:::

## Switch models via admin mode

KoboldCpp runs in `--admin` mode. When enabled, KoboldCpp reads preset configuration files from `/models/admindir` inside the container. Each preset is a `.kcpps` file in JSON format that defines a set of startup parameters, such as model path, context size, and GPU layer count. You can switch between presets from the web interface without manually restarting the server.

:::warning Admin mode security
The admin panel exposes configuration controls. Only enable this mode in trusted network environments, and protect access with authentication or a reverse proxy.
:::

By default, only one preset is auto-generated: `qwen3.5-4b.kcpps`, which points to the built-in Qwen3.5-4B model. To use any other GGUF model, you must prepare the model file and create a matching `.kcpps` preset, then load it through the admin panel.

The following example uses the `gemma-3-4b-it` model. Adapt the steps for any GGUF model you want to use.

### Step 1: Prepare the GGUF file

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

1. Open Control Hub and navigate to **Browse** > **Shared** > `koboldcppserver-shared` > **Deployments** > `koboldcpp-engine`.
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

### Step 2: Create and upload the preset file

Create a `.kcpps` preset file on your local computer. The file contains the startup parameters for the new model.

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

### Step 3: Load the preset in KoboldCpp

1. Open KoboldCpp and click **Admin** in the top navigation bar.
2. In the **Select New Model or Config** field, choose the `.kcpps` file you uploaded.
3. Click **Reload KoboldCpp** to load the new model.

:::tip About base config
The **Select Base Config** field is optional. If you choose a base config, KoboldCpp merges it with the new preset. Parameters defined in the new preset override matching fields in the base config. Unspecified parameters inherit from the base config.
:::

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

## Learn more

- [Download and run local AI models via Ollama](ollama.md): An alternative way to run local LLMs on Olares.
- [Chat with Local LLMs Using Open WebUI](openwebui.md): A graphical chat interface that connects to model backends.
