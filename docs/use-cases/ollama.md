---
outline: [2, 3]
description: Learn how to install Ollama on Olares, manage models using the CLI, and configure it as a central AI service for other applications.
---

# Download and run local AI models via Ollama
Ollama is a lightweight platform that allows you to run open-source AI models like `gemma3` and `deepseek-r1` directly on your machine. Within Olares, you can integrate Ollama with graphical interfaces like Open WebUI or other agents to add more features and simplify interactions.

## Learning objectives
In this guide, you will learn how to:
- Use the Ollama CLI on Olares to manage local LLMs. 
- Configure Ollama as an API service for internal and external apps.
## Before you begin
Before you start, ensure that you have Olares admin privileges.

## Install Ollama
1. Open **Market**, and search for "Ollama".
2. Click **Get**, then **Install**, and wait for installation to complete.
   ![Install Ollama](/images/manual/use-cases/ollama.png#bordered)

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
To use Ollama as the backend for other applications (such as DeerFlow inside Olares, or Obsidian on your laptop), you must configure the API to allow access from the local network.

### Verify authentication level
By default, the API's authentication level is set to **Internal**, allowing applications on the same local network to access the API without a login check.
1. Open Settings, then navigate to **Applications** > **Ollama** > **Ollama API**.
2. Confirm that **Authentication level** is set to **Internal**.
3. Click **Submit** if you made changes.
   ![Verify authentication level](/images/manual/use-cases/ollama-authentication-level.png#bordered)

### Get the endpoint
1. On the same settings page, click **Set up endpoint**.
2. Copy the frontend address displayed in the dialog. Use this address as the Base URL or Host in your application's settings.
   ![Get Ollama endpoint](/images/manual/use-cases/ollama-endpoint.png#bordered)

:::tip For OpenAI-compatible endpoint
Some apps expect an OpenAI-compatible API. If the standard endpoint fails, try appending `/v1` to your address. For example:
```
https://39975b9a1.{YOURUSERNAME}.olares.com/v1
```
:::
If the application forces you to enter an API Key, you can usually enter any string (e.g., `ollama`).

### Ensure network connectivity
Because you set the authentication to "Internal", your application must be on the same local network as Olares to connect.
* **Local Network**: If your device (or the app running on Olares) is on the same Wi-Fi or LAN, you can connect directly using the frontend address.
* **Remote Access**: If you are accessing Olares remotely, you must enable [LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your client device. This creates a secure tunnel that treats your device as if it were on the local network.


## Learn more
- [Run Ollama models with Open WebUI](./openwebui.md)
- [Integrate Ollama with DeerFlow](./deerflow.md)