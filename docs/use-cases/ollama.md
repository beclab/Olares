---
outline: [2, 3]
description: The standalone Ollama app is removed from Olares Market in version 1.12.6. Existing installations keep working, but new deployments should use Ollama Engine Base.
head:
  - - meta
    - name: keywords
      content: Olares, Ollama, private llm, run llama locally, ollama vs lm studio, self-hosted llm, ollama on olares
doc_version: "1.1"
doc_updated: "2026-07-28"
---

# Download and run local AI models via Ollama

:::warning Ollama app removed in Olares 1.12.6
The standalone Ollama app is no longer available in Olares Market starting with Olares 1.12.6.

- If you already installed Ollama on Olares 1.12.5 or earlier, the app keeps working after you update to 1.12.6. You can still pull and run models through the Ollama CLI and use the Ollama API endpoint. However, you cannot update the Ollama app itself from Olares Market.
- For new deployments on Olares 1.12.6 and later, use [Ollama Engine Base](./llm-base-apps.md) instead. It provides the same Ollama inference engine through Olares Model Console.
:::

Ollama is a lightweight platform that allows you to run open-source AI models like `gemma3` and `deepseek-r1` directly on your machine. Within Olares, you can integrate Ollama with graphical interfaces like Open WebUI or other agents to add more features and simplify interactions.

## Learning objectives

In this guide, you will learn how to:
- Use the Ollama CLI on Olares to manage local LLMs.
- Configure Ollama as an API service for internal and external apps.

## Prerequisites

- Olares admin privileges.
- The Ollama app already installed, if you are on Olares 1.12.6. New users should deploy [Ollama Engine Base](./llm-base-apps.md) instead.

## Install Ollama

:::info Not available for new installations on Olares 1.12.6+
Starting with Olares 1.12.6, the standalone Ollama app is removed from Market. If you already have it installed, skip to the next section. If you are setting up a new model backend, see [Host local large language models with Engine Base apps](./llm-base-apps.md).
:::

On Olares 1.12.5 and earlier:

1. Open Market, and search for "Ollama".
2. Click **Get**, then **Install**, and wait for installation to complete.
   ![Install Ollama](/images/manual/use-cases/ollama.png#bordered)

## Use Ollama Engine Base for new deployments

If you are starting fresh on Olares 1.12.6 or later, install [Ollama Engine Base](./llm-base-apps.md) from Market. Engine Base apps are managed through Model Console and give you the same Ollama engine without relying on the removed standalone Ollama app.

## Manage models with the Ollama CLI

Ollama CLI allows you to manage and interact with AI models directly from the Olares terminal. Below are the key commands.

### Download a model

:::tip Check Ollama library
If you are unsure which model to download, check the [Ollama Library](https://ollama.com/library) to explore available models.
:::

To download a model, use the following command:

```bash
ollama pull [model]
```

### Run a model

:::tip
If the specified model has not been downloaded yet, the `ollama run` command will automatically download it before running.
:::

To run a model, use the following command:

```bash
ollama run [model]
```

After running the command, you can enter queries directly into the CLI, and the model will generate responses.

When you're finished interacting with the model, type:

```bash
/bye
```

This will exit the session and return you to the standard terminal interface.

### Stop model

To stop a model that is currently running, use the following command:

```bash
ollama stop [model]
```

### List models

To view all models you have downloaded, use:

```bash
ollama list
```

### Remove a model

If you need to delete a model, you can use the following command:

```bash
ollama rm [model]
```

### Show information for a model

To display detailed information about a model, use:

```bash
ollama show [model]
```

### List running models

To see all currently running models, use:

```bash
ollama ps
```

## Configure Ollama API Access

To use Ollama as the backend for other applications on the same local network, make sure the API uses the correct access setting and copy the endpoint from the Ollama API page.

1. On Olares, open Settings, then go to **Application** > **Ollama**.
2. Under **Entrances**, click **Ollama API**.
3. Make sure that **Authentication level** is set to **Internal**. If you change the setting, click **Submit**.
4. Under **Endpoint settings**, copy the URL displayed in **Endpoint**. Use this address as the Base URL or Host in your application's settings.

   ![Ollama endpoint](/images/manual/use-cases/lp-endpoint-entrances-ollama-api.png#bordered){width=90%}

:::tip For OpenAI-compatible endpoint
Some apps expect an OpenAI-compatible API. If the standard endpoint fails, try appending `/v1` to your address. For example:

```
https://a5be22681.{YOURUSERNAME}.olares.com/v1
```
:::

If the application forces you to enter an API Key, you can usually enter any string (for example, `ollama`).

### Ensure network connectivity

Because you set the authentication to "Internal", your application must be on the same local network as Olares to connect.

- **Local network**: If your device (or the app running on Olares) is on the same Wi-Fi or LAN, you can connect directly using the frontend address.
- **Remote access**: If you are accessing Olares remotely, you must enable [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your client device. This creates a secure tunnel that effectively places your device on the same network as Olares.

## Learn more

- [Host local large language models with Engine Base apps](./llm-base-apps.md)
- [Run Ollama models with Open WebUI](./openwebui.md)
- [Build a local AI agent with LobeHub](./lobechat.md)
- [Integrate Ollama with DeerFlow 2.0](./deerflow2.md)
