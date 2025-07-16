---
outline: [2, 3]
description: Learn how to download and manage AI models locally using Ollama CLI within the Olares environment.
---

# Download and run AI models locally via Ollama
Ollama is a lightweight platform that allows you to run open-source AI models like `deepseek-r1` and `gemma3` directly on your machine. Within Olares, you can integrate Ollama with graphical interfaces like Open WebUI to add more features and simplify interactions.

This guide will show you how to set up and use Ollama CLI on Olares.

## Before you begin
Before you start, ensure that:
- You have Olares admin privileges.
- Install LarePass and [enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) on your device.

## Install Ollama

Directly install Ollama from the Market.

Once installation is complete, you can access Ollama terminal from the Launchpad.

![Ollama](/images/manual/use-cases/ollama.png#bordered)
## Ollama CLI
Ollama CLI allows you to manage and interact with AI models directly. Below are the key commands and their usage:

### Download model
:::tip Check Ollama library
If you are unsure which model to download, check the [Ollama Library](https://ollama.com/library) to explore available models.
:::
To download a model, use the following command:
```bash
ollama pull [model]
```

### Run model
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
To view all models installed on your system, use:
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

## Learn more
- [Learn how to run Ollama models with Open WebUI](openwebui.md)