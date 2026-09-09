---
outline: [2, 3]
description: Connect an AI client app to a model service by choosing the connection source and API format, copying the Base URL, and entering the model name and API key.
head:
  - - meta
    - name: keywords
      content: Olares, AI apps, Model Console, LLM service, API format, Base URL, Ollama, OpenAI-Compatible
---

# Connect an AI app to a model service <Badge type="tip" text="^ 1.12.6" />

On Olares, an AI service app provides AI capabilities over an API, while a client app provides the interface or workflow you use. Connecting them follows the same pattern across apps: choose how the client reaches the service, match the API format, and copy the service address and model name.

This page covers that common pattern. For the exact fields and buttons in a specific client app, use the app-specific tutorial linked at the end of this page.

## Before you begin

- Install both the AI service app and the AI client app.
- For an LLM service app, open it from the Launchpad and make sure **Model** shows **Ready** and **Engine** shows **Running**.

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

## Add the service to the client app

Open the client app's model, provider, or integration settings, then enter the values collected from the service app:

| Client setting | Value to use |
|---|---|
| Provider or API format | The format selected in Model Console, such as **OpenAI-Compatible** or **Ollama** |
| Base URL or endpoint | The complete URL copied from Model Console or the app entrance |
| Model name or model ID | The complete model name shown in Model Console |
| API key | The real key required by the service, or a placeholder if the local client requires a non-empty value |

The labels vary by client. If a client asks for additional fields or changes where requests are sent from, follow that client's tutorial instead of guessing.

## Verify the connection

Save the provider settings and use the client app's connection test or model-list refresh. If the client has neither option, start a new session, select the configured model, and send a short request. A response confirms that the client can reach the service and use the selected model.

## Fix common connection errors

| Symptom | Likely cause and fix |
|---|---|
| The client reports "Model not found" | The model name was abbreviated or missing prefixes. Copy the full model name from the Model Console. |
| The connectivity check fails or the Base URL is unreachable | The connection source does not match where the client runs. Reopen the Model Console, select the matching **Connection source**, and copy the Base URL again. |
| A browser reports a CORS error or Olares authentication page | The client may be sending requests from the browser instead of its server. Check the client tutorial for the correct request mode and service entrance. |

## App-specific tutorials

- [Build your local AI agent with LobeHub](/use-cases/lobechat.md)
- [Set up Open WebUI for local AI chat](/use-cases/openwebui.md)
- [Customize your local AI assistant using Dify](/use-cases/dify.md)

## Learn more

- [How do AI apps connect on Olares?](../help/usage.md#how-do-ai-apps-connect-on-olares)
- [Run local LLMs with Ollama, vLLM, llama.cpp, and SGLang](../../use-cases/llm-base-apps.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)
