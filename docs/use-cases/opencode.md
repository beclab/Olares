---
outline: deep
description: Set up OpenCode on Olares to run an AI coding agent. Connect it to Ollama-hosted models via browser or local CLI, and use natural language to write, test, and manage code.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, AI coding agent, Ollama, self-hosted, code generation, TUI
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-04-02"
---

# Set up OpenCode as your AI coding agent

OpenCode is an AI-powered coding agent that lets you write, test, and manage code through natural language. It supports multiple AI providers and can run shell commands, create files, and install development environments from a chat interface.

On Olares, you can use OpenCode in two ways:

- **Browser**: Install OpenCode as an app on Olares and access it through your browser.
- **Local CLI**: Install OpenCode on your computer and connect it to Ollama on Olares for a native terminal experience.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install OpenCode on Olares and connect it to an Ollama-hosted model.
- Create projects and run coding tasks through the chat interface or the terminal-based UI (TUI).
- Use OpenCode from your local computer over the LarePass VPN or from inside VS Code.
- Edit the OpenCode configuration file to manage providers, models, and tools.

## Prerequisites

- An Olares device with sufficient disk space and memory
- [Ollama installed](./ollama.md) on Olares with at least one model downloaded
- Admin privileges to install apps from Market
- LarePass VPN enabled on your computer (for local CLI usage only)

## Run OpenCode in the browser

This option installs OpenCode as an application on your Olares device. You access it through your browser.

### Install OpenCode

1. Open Market and search for "OpenCode".
2. Click **Get**, then **Install**.
   ![Install OpenCode](/images/manual/use-cases/opencode.png#bordered)

3. Wait for installation to complete, then launch OpenCode from Launchpad.

   After installation, OpenCode needs to download dependency packages. This might take 10 to 30 minutes depending on your network conditions. 
   
   To track the download progress:
   1. Open Control Hub and select the OpenCode project from the sidebar.
   2. Navigate to **Deployments** > **opencode** and click the running pod.
   3. Under **Containers**, locate the **init-packages** container, and click <i class="material-symbols-outlined">article</i> to open the log window.
   ![Check initialization progress in Control Hub](/images/manual/use-cases/opencode-init-package.png#bordered)

### Get the Ollama endpoint

To connect OpenCode to Ollama, get the shared entrance URL:

1. Open Settings, then navigate to **Applications** > **Ollama**.
2. In **Shared entrances**, select **Ollama API** to view the shared endpoint URL.
   ![Ollama shared entrance in Settings](/images/manual/use-cases/ollama-shared.png#bordered){width=80%}

3. Copy the shared endpoint. For example:
   ```plain
   http://d54536a50.shared.olares.com
   ```

### Connect to Ollama

1. In OpenCode, click <i class="material-symbols-outlined">settings</i> in the bottom-left corner.
   ![Open OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. Select **Providers**, then scroll down and select **Connect** next to **Custom Provider**.
   ![Select custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. Enter the following details:
   - **Provider ID**: A unique identifier for this provider. For example: `olares-ollama`.
   - **Display name**: The name shown in the provider list. For example: `Olares Ollama`.
   - **Base URL**: The endpoint URL you copied above, with `/v1` appended.
   - **Models**
     - **Model ID**: The model to use. For example: `qwen3.5:9b`.
     - **Display Name**: The name shown for this model. For example: `Qwen3.5 9B`.

   ![Provider configuration](/images/manual/use-cases/opencode-provider-config.png#bordered){width=70%}
4. To add multiple models, click **Add model** and enter the model ID and display name.

5. Click **Submit** to save the configuration. Your newly added provider will appear in the provider list.
   ![Provider list](/images/manual/use-cases/opencode-provider-list.png#bordered)

### Create a project

The default workspace is `Home/Code` in Files. Click **+** in the left navigation bar to create a project.

![Create a project](/images/manual/use-cases/opencode-create-project.png#bordered)

To work with multiple projects, create subfolders under `Home/Code` first:

1. Open Files and navigate to `Home/Code/`.
2. Create a subfolder for each project.
   ![Create subfolders in Files](/images/manual/use-cases/opencode-create-subfolders-in-files.png#bordered)

3. Go back to OpenCode, click **+**, and open the project with the subfolder name.
   ![Open project from subfolder](/images/manual/use-cases/opencode-create-more-projects.png#bordered)

### Start coding

You can now interact with OpenCode through the chat interface.

1. Select a project to open the coding agent interface.
2. Below the chat box, select **Big Pickle** to open the model selector, and select **Qwen3.5 9B** from the list.
3. Type a coding task in natural language.
   ![Code generation](/images/manual/use-cases/opencode-code-generation.png#bordered)

4. Click <i class="material-symbols-outlined">folder_open</i> to open the file browser and review the generated code.
   ![View files](/images/manual/use-cases/opencode-view-files.png#bordered)

5. Use `@` to mention a file in the chat window and ask OpenCode to edit it.
   ![Mention file in chat](/images/manual/use-cases/opencode-mention.png#bordered)

### Use the TUI

OpenCode also offers a terminal-based UI (TUI). To use the TUI in the browser:

1. Click <i class="material-symbols-outlined">terminal_2</i> from the top-right to open the terminal.
   ![Open web terminal](/images/manual/use-cases/opencode-web-terminal.png#bordered)

2. Run `opencode` to launch the TUI:
   ![Launch TUI in browser](/images/manual/use-cases/opencode-web-launch-tui.png#bordered)

3. Use the `/models` command to switch to Qwen3.5 9B.
   ![Select model in TUI](/images/manual/use-cases/opencode-web-tui-select-model.png#bordered)

4. Type your prompt directly. For example, ask OpenCode to improve the code.
   ![Chat in TUI](/images/manual/use-cases/opencode-web-tui-chat.png#bordered)

## Run OpenCode from your computer

This option installs the OpenCode CLI on your local machine and connects it to Ollama on Olares via LarePass VPN for a native terminal experience.

### Install OpenCode CLI

Install the CLI using the official installer:

```bash
curl -fsSL https://opencode.ai/install.sh | sh
```

### Get the Ollama endpoint

The local CLI requires the Ollama API endpoint. The shared entrance URL does not work for CLI connections.

1. Open Settings, then navigate to **Applications** > **Ollama**.
2. In **Entrances**, select **Ollama API** to view the endpoint URL.
   ![Ollama API endpoint](/images/manual/use-cases/ollama-api.png#bordered){width=70%}

3. Copy the endpoint URL. For example:
   ```plain
   https://a5be22681.laresprime.olares.com
   ```

### Configure the connection

1. Enable LarePass VPN on your computer to connect to Olares.
   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

   :::tip On the same local network?
   If your computer and Olares are on the same LAN, you can skip VPN and use the `.local` domain instead. Replace `https://a5be22681.{username}.olares.com` with `http://a5be22681.{username}.olares.local` in the config below. For details, see [Use `.local` domain](../manual/best-practices/local-access.md#method-2-use-local-domain).
   :::

2. Open the OpenCode config file at `~/.config/opencode/config.json` in a text editor. Add a custom provider with your Ollama endpoint and model. The general format is:

   ```json
   {
     "$schema": "https://opencode.ai/config.json",
     "provider": {
       "<provider-id>": {
         "npm": "@ai-sdk/openai-compatible",
         "name": "<display-name>",
         "options": {
           "baseURL": "<your-endpoint>/v1"
         },
         "models": {
           "<model-id>": {
             "name": "<model-display-name>"
           }
         }
       }
     }
   }
   ```

   For example, to connect to Ollama on Olares with the Qwen3.5 9B model:

   ```json
   {
     "$schema": "https://opencode.ai/config.json",
     "disabled_providers": [
       "ollama"
     ],
     "provider": {
       "olares-ollama": {
         "name": "olares-ollama",
         "npm": "@ai-sdk/openai-compatible",
         "models": {
           "qwen3.5:9b": {
             "name": "Qwen3.5 9B"
           }
         },
         "options": {
           "baseURL": "https://a5be22681.laresprime.olares.com/v1"
         }
       }
     }
   }
   ```

   :::info Windows WSL users
   If you installed OpenCode in WSL, the config file path is `~/.local/share/opencode/config.json`.
   :::

3. Save the file.

### Launch OpenCode TUI

1. In your terminal, run `opencode` to launch the TUI:
   ![Launch OpenCode TUI in terminal](/images/manual/use-cases/opencode-terminal-tui.png#bordered)

2. Use the `/models` command to switch to Qwen3.5 9B.
   ![Select model in terminal TUI](/images/manual/use-cases/opencode-terminal-tui-select-model.png#bordered)

3. Start chatting with your self-hosted models.
   ![Chat in terminal TUI](/images/manual/use-cases/opencode-terminal-tui-chat.png#bordered)

:::info First connection
The first connection might take longer to establish.
:::

### Use OpenCode in VS Code

To work with your codebase directly, open a terminal in VS Code and run `opencode` from your project directory.

![OpenCode in VS Code](/images/manual/use-cases/opencode-vscode.png#bordered)

## Edit the config file

OpenCode stores its configuration in a JSON file. You can edit this file directly to manage providers, models, and tools.

1. Open Files and navigate to `Application/Data/opencode/.config/opencode/`.
2. Right-click `config.jsonc` and select **Rename**.
   ![Rename config file](/images/manual/use-cases/opencode-rename-config-file.png#bordered)

3. Rename the file to `config.json` so you can edit it directly in Files. OpenCode recognizes both extensions.

4. Open `config.json` and click <i class="material-symbols-outlined">edit_square</i> to edit it.
   ![Edit config file](/images/manual/use-cases/opencode-edit-config-file.png#bordered)

5. Save the changes.

6. Restart OpenCode from **Settings** > **Applications** to apply the changes.

## Learn more

- [Manage packages](opencode-packages.md): Install system-level and language-specific packages.
- [Skills and plugins](opencode-extensions.md): Add capabilities through skills and plugins.
- [Orchestrate multi-agent workflows with oh-my-openagent](opencode-omo.md): Enable OMO to run multi-agent collaboration in OpenCode.
- [Common issues](opencode-issues.md): Solutions for known problems.
- [Connect AI coding assistants to up-to-date docs with Context7](context7.md#opencode): Register Context7 as a remote MCP server in OpenCode.
- [OpenCode official documentation](https://opencode.ai/docs)
