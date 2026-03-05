---
outline: [2, 3]
description: Learn how to connect AI applications on Olares using shared endpoints, with Ollama as a practical example.
---

# Connect AI apps

Many AI applications on Olares follow the same pattern: one app provides AI capabilities over an API, and another app provides the interface you use every day. Once you understand this pattern, you can apply the same steps to connect almost any compatible combination of apps.

This tutorial explains the core concepts and walks you through practical examples using Ollama as the AI service.

## Objectives

By the end of this tutorial, you will be able to:

- Identify shared applications that can act as AI services.
- Locate the correct endpoint for a shared AI service.
- Choose the right endpoint format for different client apps.
- Connect common client apps such as LobeChat, n8n, and Continue.dev to Ollama.

## How it works

Connecting a client app to a shared AI service on Olares usually involves three steps:

1. In Olares Settings, find the shared application's API entrance and set the **Authentication level** to **Internal**.
2. Copy the endpoint shown for that entrance.
3. In your client app, paste this endpoint into the model or API configuration page. If the connection fails, adjust the endpoint according to the rules in [Which endpoint to use](#which-endpoint-to-use).

## Core concepts

### Shared applications

A shared application is an app that is installed once and shared by all users on the device, rather than belonging to a single user. In Olares Market, shared applications are marked with a group badge.

Not every shared application exposes an API for other apps to call. This tutorial focuses on shared applications that provide an API entrance, for example, Ollama or ComfyUI Shared. In the connection examples, they usually act as the AI service side. These apps commonly appear in two forms:

- **Backend-only service**: Runs as a headless service with no built-in interface, like Ollama. It exposes API endpoints that compatible client apps can use. To work with this type, install a supported client app and point it to the shared endpoint.
- **Service with built-in interface**: Includes both a backend service and its own user interface, like ComfyUI Shared. You can use it directly in the browser, or connect compatible third-party clients to its API.

### Service apps and client apps

- **Service apps** provide AI capabilities over an API. In the examples in this tutorial, a shared app such as Ollama plays this role.
- **Client apps** are the apps you use directly. To work, they need to know which AI service to use and which address to call. LobeChat and Open WebUI are examples: they provide the chat interface but rely on a service like Ollama to generate responses.

### Endpoints

An endpoint is the URL through which an application's entrance can be reached. When a shared app exposes an API entrance, you will usually see both of the following endpoint types:

| Type | Format | Description |
|------|--------|-------------|
| User endpoint | `https://{route-ID}.{OlaresID}.olares.com` | Tied to a specific user, typically used for browser-facing access |
| Shared endpoint | `http://{route-ID}.shared.olares.com` | System-wide access, not tied to any user |

### Frontend calls vs. backend calls

Client apps send API requests in one of two ways:

- **Backend calls**: The client app's server process makes the request. This is the most common approach for AI integrations.
- **Frontend (browser) calls**: The request is sent directly from your browser, which requires HTTPS and may be subject to CORS restrictions.

Most AI client apps use backend calls. Some lightweight or browser-based tools send requests from the browser instead. Check the client app's documentation if you are unsure which it uses.

### Authentication levels

Olares provides three access levels for each application entrance:

- **Public**: Open to anyone on the internet. Not recommended for private services.
- **Internal**: Allows access within the LAN or via LarePass VPN. This is the recommended setting for API entrances when connecting apps in this tutorial.
- **Private**: Requires user authentication.

## Which endpoint to use

This tutorial covers connections using the `olares.com` domain. If your client device is on the same local network as Olares, the same approach applies using `.local` addresses.

When configuring a client app to use a shared AI service:

1. **Try the shared endpoint first** (`http://{route-ID}.shared.olares.com`). Shared endpoints are designed for direct app-to-app API access. They do not require user credentials and are generally the most reliable option.

2. **Fall back to the user endpoint** (`https://{route-ID}.{OlaresID}.olares.com`). If the shared endpoint is unavailable, or the client app sends requests from the browser rather than its own server, use the user endpoint. Set its **Authentication level** to **Internal** (recommended) so it can be accessed without a login prompt but is not exposed publicly.

3. **Adjust the path suffix if needed**. Many client apps expect the base URL to end with `/v1` for OpenAI-compatible APIs, or `/api` for other formats. If the connection fails, try appending the appropriate suffix. For example: `http://{route-ID}.shared.olares.com/v1`. This applies to both endpoint types.

If a client app requires an API key but the service does not use one, enter any placeholder text such as `ollama` to satisfy the required field.

## Example 1: Connect Ollama to LobeChat

In this example, Ollama Shared acts as the AI service, and LobeChat is the client app that provides the chat interface.

This example uses `qwen2.5:1.5b` as the model. Make sure you have downloaded it before starting.

1. On Olares, open Settings, then go to **Application** > **Ollama**.
2. In **Shared entrances**, select **Ollama API**.
   ![Ollama shared entrance](/images/manual/tutorials/api-ollama-shared.png#bordered)
   
3. Copy the shared endpoint URL.
4. Open LobeChat, then go to **Settings** > **AI Service Provider** > **Ollama**.
5. In the **Interface proxy address** field, paste the Ollama shared endpoint you copied.
   :::warning
   If you are using local Ollama models, do not enable **Use Client Request Mode**.
   :::
   ![Enter shared endpoint](/images/manual/tutorials/api-lobechat-enter-url.png#bordered)
6. Verify the connection:

   a. Click **Fetch models**. Models you have downloaded in Ollama will appear in the list.
   ![Fetch models](/images/manual/tutorials/api-lobechat-fetch-models.png#bordered)

   b. Enable the models you want to use. For example, enable **Qwen2.5 1.5B**.

   c. For **Connectivity Check**, select **qwen2.5:1.5b** from the dropdown, then click **Check**.

      When **Check Passed** appears, the connection is established.
      ![Check passed](/images/manual/tutorials/api-lobechat-check-passed.png#bordered)

## Example 2: Connect Ollama to n8n

n8n makes requests from the browser rather than its server, so it requires a user endpoint. Configure the authentication level to **Internal** so it can be accessed without a login prompt.

1. On Olares, open Settings, then go to **Application** > **Ollama**.
2. In **Entrances**, click **Ollama API**.
3. Set the **Authentication level** to **Internal**.
4. Click **Set up endpoint**, then copy the endpoint URL displayed.
5. Create a new Ollama credential in n8n:

   a. In n8n, select **+** > **Credential** from the left navigation bar.

   b. In the **Add new credential** dialog, select **Ollama** from the dropdown, then click **Continue**.

   c. Paste the Ollama endpoint URL you copied.

   d. Click **Save**. n8n will automatically test the connection.

      When **Connection tested successfully** appears, the connection is established.
      
      ![n8n Ollama connected](/images/manual/tutorials/api-n8n-connected.png#bordered)

## Example 3: Connect Ollama to Continue.dev (outside Olares)

You can connect your local IDE to Ollama running on your Olares system, so that AI assistance and code completion are powered by your own hardware rather than a third-party cloud.

This example uses `llama3.1:8b`, `qwen2.5-coder:7b`, and `qwen2.5-coder:1.5b`. Make sure you have downloaded these models before starting.

1. On Olares, open Settings, then go to **Application** > **Ollama**.
2. In **Entrances**, click **Ollama API**.
3. Set the **Authentication level** to **Internal**.
4. Click **Set up endpoint**, then copy the endpoint URL displayed.
5. In your local IDE (for example, IntelliJ IDEA), open the Continue panel.
6. Configure models in Continue to use Ollama:

   a. Click **Local Config** to open the **Configs** menu, then click the Settings icon next to **Local Config**.
   ![Open local config](/images/manual/tutorials/api-continue-local-config.png#bordered){width=60%}

   b. In the `config.yaml` that opens, update the model entries with the Ollama endpoint you copied:
   ```yaml
   name: Local Config
   version: 1.0.0
   schema: v1
   models:
   - name: Llama3.1-8B
      provider: ollama
      model: llama3.1:8b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - chat
   - name: Qwen2.5-Coder 7B
      provider: ollama
      model: qwen2.5-coder:7b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - edit
         - embed
         - rerank
   - name: Qwen2.5-Coder 1.5B
      provider: ollama
      model: qwen2.5-coder:1.5b
      apiBase: https://<your-ollama-endpoint>
      roles:
         - autocomplete
         - apply
   ```
7. Enable the VPN on your LarePass desktop client.
   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

8. In the Continue chat panel, enter a prompt to test the connection. For example:
   ```plain
   Write a hello world python script
   ```
   ![Enter prompt](/images/manual/tutorials/api-continue-prompt.png#bordered){width=60%}

   Continue will route the request to Ollama on your Olares system and return the result. With LarePass VPN enabled, your IDE can reach the Ollama endpoint as if it were on the same private network.
   ![Result](/images/manual/tutorials/api-continue-hello-world.png#bordered){width=60%}
## Learn more

- [Applications](../../developer/concepts/application.md)
- [Network](../../developer/concepts/network.md)
- [Manage application entrances](../olares/settings/manage-entrance.md)

