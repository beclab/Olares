---
outline: [2, 3]
description: Find answers to common questions about using Olares and community apps.
head:
  - - meta
    - name: keywords
      content: Olares, usage FAQ, apps, storage, multi-node cluster, Control Hub, update app
---

# Usage FAQs

Find answers to common questions about daily usage, applications, and system management.

## Applications

### What apps can I run in Olares?

The [Olares Market](https://market.olares.com/) maintains popular open-source apps like Ollama, ComfyUI, and Open WebUI.

If you have Docker experience, you can package apps not listed in the Olares Market as [Olares Application Charts](../../developer/develop/package/chart.md) and test them on your own Olares device.

### Can I play games on my Olares device?

Yes. Install the Steam Headless app to transform your Olares device into a gaming server.

* [**Streaming**](../../use-cases/steam-stream.md): Run games locally on Olares and stream them to devices like phones and tablets.
* [**Direct play**](../../use-cases/steam-direct-play.md): Connect a monitor, keyboard, and mouse directly to the Olares device to play games without streaming.

### How do I access the Windows environment in Olares?

Install and run a Windows VM from the Olares Market, and access it using any standard RDP client.

For detailed instructions, refer to [Run a Windows VM on your Olares device](../../use-cases/windows.md).

### Can I develop apps on Olares?

Yes. Build your app with your preferred local tools, package it as an [Olares Application Chart](../../developer/develop/package/chart.md), and test it on an Olares device before submitting it to the Market.

### How do AI apps connect on Olares?

When you use AI on Olares, you typically work with two separate applications: one AI service app provides AI capabilities in the background, and another AI client app provides the chat interface you interact with directly.

- **AI client apps** provide the front-end chat interface or workflow canvas, such as LobeHub. They rely on an AI service app (often called a **provider**) to perform AI tasks, such as generating text and extracting data.
- **AI service apps** provide AI capabilities for compatible clients over an API, such as chat, search, and speech recognition. Some have their own web interface for management, while others run primarily as headless backend services.

    On Olares, AI service apps fall into two categories:
    - **LLM service apps** host large language models (LLMs) for text generation, code completion, and chat. They include eight pre-built model apps and custom model instances created on [Engine Base apps](/use-cases/llm-base-apps.md).
    - **Other AI service apps** provide non-LLM functions, such as speech recognition (Speaches) and text extraction (PaddleOCR).

Because different providers use different communication rules, they rely on specific **API formats**, which you can think of as the "languages" the apps use to talk to each other. The two most common formats are **OpenAI-Compatible** and **Ollama**. To connect the apps, enter the service app's Base URL, model name, and API key in the client app. For the steps, see [Connect an AI app to a model service](../best-practices/connect-ai-apps.md).

### Can I manually update an application version?

:::tip Important
We recommend always updating applications through the Olares Market to ensure stability and compatibility.
:::

Yes. In some cases, an application might show an internal prompt for a new update before it is officially available in the Market. If you urgently need the latest features, you can [manually update the app image in Control Hub](../update-app-image.md).

Note that manual edits in Control Hub are temporary and will be overwritten by the next Market update, and the application might fail to start due to compatibility issues. Review the warnings in [Manually update an app image](../update-app-image.md) before proceeding.

## Storage## Storage

### If I add new disks to a running Olares machine, will Olares use them automatically?

It depends on the type of drive:
* **USB drives**: Yes. These are automatically mounted and will appear immediately in the Files app.
* **Internal drives**: No. Iternal HDDs or SSDs require manual configuration to join the storage pool.
* **SMB shares**: Add network storage manually. Go to **External** > **Connect to server** in the Files app.

For detailed instructions, refer to [Expand storage in Olares](../best-practices/expand-storage-in-olares.md).

## Multi-node clusters

### How do I add more machines to my cluster?

By default, Olares installs as a single-node cluster. To create a scalable, multi-node cluster, install Olares as a master node and then add worker nodes.

Note that this is currently an Alpha feature and works on Linux only. For detailed instructions, refer to [Install a multi-node Olares cluster](../best-practices/install-olares-multi-node.md).
