---
outline: [2, 3]
description: Connect an AI client app to a model service by choosing the connection source and API format, copying the Base URL, and entering the model name and API key.
head:
  - - meta
    - name: keywords
      content: Olares, AI apps, Model Console, LLM service, API format, Base URL, Ollama, OpenAI-Compatible
---

# Connect an AI app to a model service <Badge type="tip" text="^ 1.12.6" />

On Olares, an AI service app provides AI capabilities over an API, and an AI client app (such as LobeHub) provides the chat interface you use. To make them work together, gather the connection details from the service app and enter them in the client app.

## Before you begin

- Install both the AI service app and the AI client app.
- For an LLM service app, open it from the Launchpad and make sure the model shows **READY** and the engine shows **RUNNING**.

## Choose the connection source and API format

Open the LLM service app from the Launchpad to launch its **Model Console**, and then select the options that match your client app:

- **Connection source**: Select the option that matches where your client app runs. For example, select **Apps in Olares** when the client is installed in the same Olares cluster.
- **API format**: Select the format your client app supports, such as **OpenAI-Compatible** or **Ollama**. The Model Console displays the Base URL that matches your selection.

:::info
Non-LLM services like PaddleOCR do not use these generic formats. They communicate using their own tool-specific protocols, so you do not need to configure a provider format for them.
:::

## Copy the Base URL

The Base URL is the network address where the service app receives and processes requests.

- **For LLM service apps**: Copy the **Base URL** displayed in the Model Console. Copy it exactly as shown, including any path suffix such as `/v1`.
- **For other AI service apps**: Open Olares Settings, go to **Applications** > **[AppName]** > **Entrances**, and copy the **Endpoint URL**. Ensure the entrance's **Authentication level** is set to **Internal** so other apps can access it without a login barrier.

    :::tip Multiple entrances
    Some apps expose more than one entrance. Choose the entrance that matches your client's protocol or use case. For example, use the main entrance for web UI access and a dedicated API entrance for programmatic integrations.
    :::

## Enter the model name and API key

- **Model name**: Copy the **Model name** from the Model Console exactly as displayed. Do not abbreviate it or remove repository prefixes (such as `unsloth/`) or quantization tags (such as `UD-Q4_K_XL`), otherwise the client might return an error like "Model not found".
- **API key**: AI service apps deployed locally on Olares trust requests from other apps in the same cluster, so a real API key is usually not required. If the client app still requires a value in this field, enter any placeholder text such as `olares` or `local`.

## Verify the connection

In the client app, save the provider settings and run its connectivity check. For example, in LobeHub, click **Fetch models** next to **Model List** to load the model, enable it, and then select it next to **Connectivity Check** and click **Check**. When the check passes, the connection is established.

:::warning Disable Client Request Mode in LobeHub
Do not enable **Use Client Request Mode** in LobeHub. Enabling this forces the application to make frontend browser calls, which can trigger cross-origin (CORS) blocks or Olares security authentication prompts. Keeping it disabled ensures secure, direct backend-to-backend communication.
:::

## Fix common connection errors

| Symptom | Likely cause and fix |
|---|---|
| The client reports "Model not found" | The model name was abbreviated or missing prefixes. Copy the full model name from the Model Console. |
| The connectivity check fails or the Base URL is unreachable | The connection source does not match where the client runs. Reopen the Model Console, select the matching **Connection source**, and copy the Base URL again. |
| CORS errors or authentication prompts appear in the client | For LobeHub, disable **Use Client Request Mode** so requests go directly between the apps. |

## FAQ

### How do I connect non-AI apps?

The same internal-entrance pattern applies when connecting non-AI apps to each other. For example:
- The *Arrs media stack uses internal entrance URLs to connect Sonarr, Radarr, Prowlarr, Bazarr, and qBittorrent. See [Manage your media library with the *Arrs ecosystem](/use-cases/arrs.md).
- SearXNG itself is not an AI model, but it can be connected to an AI client such as Vane for private, enhanced search. See [Connect SearXNG to Vane](/use-cases/perplexica.md).

## Learn more

- [How do AI apps connect on Olares?](../help/usage.md#how-do-ai-apps-connect-on-olares)
- [Host local large language models with Engine Base apps](../../use-cases/llm-base-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
