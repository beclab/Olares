---
outline: [2, 3]
description: Guide to using Open WebUI in Olares for managing Large Language Models. Learn about model management, voice interactions, and image generation capabilities with both Ollama and OpenAI-compatible APIs.
---

# Open WebUI

Open WebUI provides an intuitive interface for managing Large Language Models (LLMs) that supports both Ollama and OpenAI-compatible APIs. This page helps you set up and configure Open WebUI in Olares for:

* Model management
* Voice interactions (speech-to-text and text-to-speech)
* Image generation capabilities

:::tip Enable LarePass VPN
Make sure your device has LassPass installed and [VPN enabled](../manual/larepass/private-network.md#enable-vpn-on-larepass) to ensure a fast and stable access. 
:::

## Installation
Ollama is required prior to launching Open WebUI:
* **For admin**: Install both "Ollama" and "Open WebUI".
* **For team members**: Ensure that "Ollama" is already installed by Olares admin, and then install "Open WebUI" only.

![Install Ollama and Open WebUI](/images/manual/use-cases/install-open-webui.png){width=30%}

:::info
First-time users need to create a local Open WebUI account. This account is specifically for your Olares installation and doesn't connect to external services. Note that existing Open WebUI accounts from other installations cannot be used here - you'll need to create a new one.
:::

## Download models
:::tip
Browse available models on [Hugging Face's Chatbot Arena Leaderboard](https://huggingface.co/spaces/lmsys/chatbot-arena-leaderboard) and verify model names in the [Ollama Library](https://ollama.com/library) before downloading.
:::

Recommended starter models for optimal performance (13B parameters or smaller):

* `gemma2`: Google's latest  efficient and powerful language model
* `llama3.2`: Meta's latest open-source vision-language model

### Quick download
1. Click the dropdown menu, enter the model name (e.g., `llama3.2`).
2. Select **Pull from Ollama.com**. The download starts automatically.
### Download from settings
1. Click your username in the bottom left.
2. Navigate to **Admin Panel** > **Settings** > **Models**.
3. Under **Pull a model from Ollama.com**, enter the model name (e.g., `llama3.2`).
4. Click <i class="material-symbols-outlined">download</i> to initiate the download.
## Configure speech features
### Speech-to-text
1. Install Faster Whisper from Market based on your role:
   - Admin: Install both "Faster Whisper For Cluster" and "Faster Whisper".
   - Team members: Ensure that "Faster Whisper For Cluster" is already installed by Olares admin, and install "Faster Whisper" only.

   ![Install Faster Whisper](/images/manual/use-cases/install-faster-whisper.png){width=40%}
2. Open WebUI, and navigate to **Admin Panel** > **Settings** > **Audio**.
3. Select **OpenAI** as the speech-to-text engine, with the following configurations:
   - API Base URL: `http://whisper.whisper-{admin's local name}:8000/v1`. For example: `http://whisper.whisper-alice123:8000/v1`.
   - API key: enter any character.
4. Enter a model variant (default: `whisper-1`). You can select from the following:
   - `tiny.en`
   - `tiny`
   - `base.en`
   - `base`
   - `small.en`
   - `small`
   - `medium.en`
   - `medium`
   - `large-v1`
   - `large-v2`
   - `large-v3`
   - `large`
   - `distil-large-v2`
   - `distil-medium.en`
   - `distil-small.en`
   - `distil-large-v3`
5. Click **Save**.
6. Launch Faster Whisper after configuration. You'll see that the model you configured in Open WebUI is automatically loaded. At this point, you can either:
   - Upload audio directly to start transcription
   - Adjust parameters including:
      - Select different sub-models
      - Choose task types
      - Configure **Temperature** settings

### Text-to-speech
1. The admin installs OpenedAI Speech from Market. This launches the service within the cluster.
   :::info
   "OpenedAI Speech" is a shared application and can only be installed by Olares admin. If you are a team member, ensure that the Olares admin has already installed "OpenedAI Speech".
2. Open WebUI, and navigate to **Admin Panel** > **Settings** > **Audio**.
3. Select OpenAI as the text-to-speech engine, with the following configurations:
    - API Base URL: `http://openedaispeech.openedaispeech-{admin's local name}:8000/v1`. For example: `http://openedaispeech.openedaispeech-alice123:8000/v1`.
    - API key: enter any character.
4. Click **Save**.

### Text-to-image
With [SD Web UI Shared installed in your Olares environment](stable-diffusion.md#install-sd-web-ui), you can leverage Stable Diffusion's powerful image generation capabilities directly through Open WebUI.

1. The admin installs SD Web UI Shared from Market. This launches the Stable Diffusion service within the cluster.
2. Open Open WebUI, and navigate to **Admin Panel** > **Settings** > **Images**.
3. Select **Automatic1111** as the image generation engine, with the base URL:  `http://sdwebui.sdwebui--{admin's local name}:7860`. For example: `http://sdwebui.sdwebui-alice123:7860`.
4. Click <i class="material-symbols-outlined">cached</i> to verify the connection.
5. Turn on **Image Generation (Experimental)**, and select your preferred text-to-image model checkpoint.
6. Click **Save**.
