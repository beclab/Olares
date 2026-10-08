---
connectionVersion: "1.12.7"
connectionLatestPath: /manual/best-practices/connect-ai-apps
outline: [2, 3]
description: Understand how AI clients connect to local and remote models through Olares Router.
head:
  - - meta
    - name: keywords
      content: Olares, AI client, model service, LarePass VPN, Router, default-chat, Base URL, API key
---

# How AI apps connect to Olares models

<VersionRouteSelect />

Many AI apps focus on the interface or workflow, while a separate model service performs inference. LobeHub, OpenCode, and Claude Code all follow this pattern.

Connecting an AI app means telling the app which API to use, where to send requests, which model to call, and how to authenticate.

## Client apps, Router, and model services

An AI connection on Olares involves three parts:

- **Client app**: The chat interface, coding tool, or workflow app that sends model requests.
- **Router**: The shared entry point for AI capabilities on Olares. It authenticates the caller and routes each request.
- **Model service**: A model running in a local model app or a remote model provider connected to Router.

The client uses Router as a shared gateway. Router sends the request to the selected model and returns the response. Local and remote models use the same Router service, while the Base URL used by the client depends on the request origin.

A local model app exposes an in-cluster shared entrance for Router to call its inference engine. That address belongs to the connection between Router and the model backend. Client apps use the Base URL provided by Router.

## The request origin determines the connection

Choose a connection according to the component that sends the API request and its network location.

- **Server-side request**: An app installed on Olares usually sends requests from its own server process. The request stays within the Olares network.
- **Direct client request**: A desktop app, CLI, IDE extension, or browser client sends requests from your computer. The request reaches Olares through the local network or LarePass VPN.

Some web apps can switch between server-side and browser-side requests. For example, enabling a setting such as **Client Request Mode** moves the request origin from the Olares app to the browser. Follow the app-specific tutorial when such an option is available.

| Request origin | Router connection option | Access method |
| --- | --- | --- |
| An app process in Olares | **Apps in Olares** | The platform injects the app identity |
| A computer on the same local network | **Devices in LAN** | Direct LAN access with a Router API key |
| A computer outside the local network | **Remote** | LarePass VPN with a Router API key |

### Connecting from outside the local network

A computer on the same local network can reach Router directly. A computer on another network uses the connection information under **Remote** and first joins the Olares private network through LarePass VPN.

The VPN and API key solve different parts of the connection. LarePass VPN provides an encrypted network path to Olares. The Router API key identifies the external client and grants access to model capabilities. A remote client needs both.

## Connection parameters

AI apps use different labels, but a model connection usually contains the following values:

| Parameter | Purpose |
| --- | --- |
| Provider or API format | Defines the request format expected by the client and Router |
| Base URL | Identifies the Router address and API path |
| Model name or model ID | Identifies the model or routing rule to call |
| API key | Authenticates a client outside Olares |

### Provider and API format

“Provider” refers to different objects in the client and in Router:

- In a client, the provider is usually an adapter that formats requests. For example, an OpenAI-compatible provider can call a model running locally on Olares.
- In Router, a provider is the backend to which Router forwards requests. It can be a local model app or a cloud provider.

Use the provider or adapter specified in the app tutorial. Different formats can use different paths and request structures even when they call the same model.

### Base URL

The Base URL tells the client where to send requests. Router provides a different URL for **Apps in Olares**, **Devices in LAN**, and **Remote** because each request reaches Router through a different network path.

Copy the complete URL for the request origin, including a path such as `/v1` when shown. Some clients append their own API path, so the app tutorial may instruct you to remove or change the suffix.

### API key

The platform authenticates requests from apps in Olares. Clients on a computer use a key created in Router, including clients on the same local network.

When an Olares app requires a value in the key field, use the placeholder given in that app's tutorial.

## Model selection

The model name determines how Router handles the request:

- **`default-chat`** routes the request to the model selected for chat on Router's **Default models** page. Apps using this name automatically follow later changes to the default model.
- **A full model name** sends the request to one specific model. Use it when the app must remain on that model.

`default-chat` is a routing name, while the model-list API returns individual models. If a client builds its model menu from that API, add `default-chat` manually. Clients restricted to models returned in the list use the full model name.

## App-specific tutorials

The following tutorials show the provider, URL format, and fields required by each client:

- [Build your local AI agent with LobeHub](/use-cases/lobechat.md)
- [Set up OpenCode as your AI coding agent](/use-cases/opencode.md)
- [Write code using Claude Code](/use-cases/claude-code.md)

## Learn more

- [Use Olares Router as your AI gateway](/use-cases/olares-router.md)
- [Connect to your Olares network with LarePass VPN](../larepass/private-network.md)
