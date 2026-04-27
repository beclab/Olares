---
outline: deep
description: Set up OpenCode on Olares to run an AI coding agent. Connect it to Ollama-hosted models or OpenAI, and use natural language to write, test, and manage code.
head:
  - - meta
    - name: keywords
      content: Olares, OpenCode, AI coding agent, Ollama, OpenAI, ChatGPT, self-hosted, code generation, TUI
app_version: "1.0.11"
doc_version: "1.1"
doc_updated: "2026-04-27"
---

# Set up OpenCode as your AI coding agent

OpenCode is an AI-powered coding agent that lets you write, test, and manage code through natural language. It supports multiple AI providers and can run shell commands, create files, and install development environments from a chat interface.

On Olares, you can use OpenCode in two ways:

- **Browser**: Install OpenCode as an app on Olares and access it through your browser.
- **Local CLI**: Install OpenCode on your computer and connect it to Ollama on Olares for a native terminal experience.

## Learning objectives

By the end of this tutorial, you will learn how to:
- Install OpenCode on Olares and connect it to an Ollama-hosted model or OpenAI.
- Create projects and run coding tasks through the chat interface or the terminal-based UI (TUI).
- Use OpenCode from your local computer over the LarePass VPN or from inside VS Code.
- Edit the OpenCode configuration file to manage providers, models, and tools.

## Prerequisites

- An Olares device with sufficient disk space and memory
- [Ollama installed](./ollama.md) on Olares with at least one model downloaded, if you plan to use local models
- A ChatGPT Plus/Pro account or an OpenAI API key, if you plan to use OpenAI
- Admin privileges to install apps from Market
- LarePass VPN enabled on your computer (for local CLI usage only)

## Run OpenCode in the browser

This option installs OpenCode as an application on your Olares device. You access it through your browser.

### Install OpenCode

1. Open Market and search for "OpenCode".
   ![Install OpenCode](/images/manual/use-cases/opencode.png#bordered)

2. Click **Get**, then **Install**, and wait for installation to complete.

After installation, you will see two icons on Launchpad:
- OpenCode: The main interface for OpenCode.
- OpenCode Terminal: A command-line terminal for OpenCode. Use this if you prefer working in the TUI.

Open OpenCode from Launchpad. On first launch, OpenCode needs to download dependency packages. This might take 10 to 30 minutes depending on your network conditions.

To track the download progress:
1. Open Control Hub and select the OpenCode project from the sidebar.
2. Navigate to **Deployments** > **opencode** and click the running pod.
3. Under **Containers**, locate the **init-packages** container, and click <i class="material-symbols-outlined">article</i> to open the log window.
   ![Check initialization progress in Control Hub](/images/manual/use-cases/opencode-init-package.png#bordered)

### Get the model endpoint

Olares offers two ways to serve local models:

- **Ollama app**: One app that hosts multiple models behind a single shared endpoint.
- **Single-model app**: Each app packages one specific model and exposes its own shared endpoint.

OpenCode connects to both through a shared entrance URL.

:::tip Planning to use oh-my-openagent?
For multi-local-model setups under [oh-my-openagent](opencode-omo.md), use single-model apps. Each model gets its own endpoint, which lets you register it as a separate provider and tune per-model settings like concurrency limits.
:::

#### Ollama

To connect OpenCode to Ollama, get the shared entrance URL:

1. Open Settings, then navigate to **Applications** > **Ollama**.
2. In **Shared entrances**, select **Ollama API** to view the shared endpoint URL.
   ![Ollama shared entrance in Settings](/images/manual/use-cases/ollama-shared.png#bordered){width=80%}

3. Copy the shared endpoint. For example:
   ```plain
   http://d54536a50.shared.olares.com
   ```

#### Single-model apps

To connect OpenCode to a single-model app, get its shared entrance URL. The example below uses Qwen3.5 9B Q4_K_M:

1. Open Settings, then navigate to **Applications** > **Qwen3.5 9B Q4_K_M**.
2. In **Shared entrances**, select **Ollama API** to view the shared endpoint URL.
   ![Qwen3.5 9B Q4_K_M shared entrance in Settings](/images/manual/use-cases/ollama-shared.png#bordered){width=80%}

3. Copy the shared endpoint. For example:
   ```plain
   http://bd5355000.shared.olares.com
   ```

### Connect to a custom provider

Add your endpoint as a custom provider in OpenCode. The steps are the same for the Ollama app and single-model apps, but the Model ID you enter differs based on the app type.

1. In OpenCode, click <i class="material-symbols-outlined">settings</i> in the bottom-left corner.
   ![Open OpenCode settings](/images/manual/use-cases/opencode-settings.png#bordered)

2. Select **Providers**, then scroll down and select **Connect** next to **Custom Provider**.
   ![Select custom provider](/images/manual/use-cases/opencode-custom-provider.png#bordered)

3. Enter the following details:
   - **Provider ID**: A unique identifier for this provider. For example, `olares-ollama` for the Ollama app or `ollama-9b` for a single-model app.
   - **Display name**: The name shown in the provider list. For example, `Olares Ollama` or `Ollama 9B`.
   - **Base URL**: The endpoint URL you copied above, with `/v1` appended.
   - **Models**:
     - **Model ID**: The model to use. For the Ollama app, enter the ID of any model you've downloaded, such as `qwen3.5:9b`. For a single-model app, enter the exact model name shown for the app since the app serves only that model.
     - **Display Name**: The name shown for this model. For example, `Qwen3.5 9B`.

   ![Provider configuration](/images/manual/use-cases/opencode-provider-config.png#bordered){width=70%}

4. To add more models under the same Ollama app provider, click **Add model** and enter the model ID and display name. Skip this step for single-model app providers.

5. Click **Submit** to save the configuration. Your newly added provider will appear in the provider list.
   ![Provider list](/images/manual/use-cases/opencode-provider-list.png#bordered)

### Connect to OpenAI

OpenCode supports two practical ways to connect OpenAI:

- **Browser sign-in**: Recommended for ChatGPT Plus or Pro accounts.
- **API key**: Simple and direct. Usage is billed separately by OpenAI based on token usage, even if you also have a ChatGPT subscription.

:::warning Do not use headless sign-in
The headless sign-in method is prone to known failures caused by OpenAI security checks. After a failed attempt, the sign-in flow might not recover for a short period of time. Use browser sign-in or an API key instead.
:::

<tabs>
<template #Browser-sign-in>

Use this method when you want OpenCode to connect through your ChatGPT Plus or Pro account.

Because OpenCode runs inside an Olares container, the browser callback URL that starts with `localhost:1455` cannot reach OpenCode directly from your browser. Start the browser flow, then relay the final callback URL back to OpenCode manually.

:::warning Keep the authorization session active
- Complete the steps within 5 minutes. If the session expires, start over.
- Keep the **Connect OpenAI** dialog open in OpenCode until the connection succeeds.
- Each callback URL is one-time use. If the attempt fails, close the dialog and repeat the flow to get a new URL.
:::

1. In OpenCode, click <i class="material-symbols-outlined">settings</i> > **Providers**, then scroll down and select **Connect** next to **OpenAI**.
   ![Add the OpenAI provider](/images/manual/use-cases/opencode-openai-provider.png#bordered)

2. For the sign-in method, select **ChatGPT Pro/Plus (browser)**.
   ![Select browser sign-in for OpenAI](/images/manual/use-cases/opencode-openai-browser-signin.png#bordered){width=70%}

3. In the authorization dialog, click the authorization link to open it in your browser.
   ![Open the OpenAI authorization link](/images/manual/use-cases/opencode-openai-authorization-link.png#bordered){width=70%}

4. Sign in with your ChatGPT account.

   After sign-in, the browser redirects to a URL that starts with `http://localhost:1455` and might show that the page cannot be reached. This is expected. Do not close the page.

5. Copy the full URL from the browser address bar, from `http://localhost:1455` through the end of the URL.
   ![Open the OpenAI authorization link](/images/manual/use-cases/opencode-openai-auth-url.png#bordered)

6. Open OpenCode Terminal from Launchpad.

7. Run `curl` with the full URL you copied:

   ```bash
   curl '<paste-the-full-url-you-copied>'
   ```
   :::warning
   Wrap the copied URL in single quotes (`''`) when running `curl`. Otherwise, the shell treats `&` as a special character and truncates the URL.
   :::

8. Wait until the terminal returns `Authorization Successful`. The waiting dialog in OpenCode should change to connected automatically.
   ![OpenAI authorization succeeds in OpenCode Terminal](/images/manual/use-cases/opencode-openai-terminal-success.png#bordered)

9. Refresh the OpenCode page to load new OpenAI models.
   ![OpenAI provider connected](/images/manual/use-cases/opencode-openai-connected.png#bordered){width=70%}
</template>

<template #API-key>

Use this method when you want a direct OpenAI API connection. It is usually easier to set up than browser sign-in, but OpenAI bills API usage separately by token.

1. Create or copy an OpenAI API key from your OpenAI account.
2. In OpenCode, click <i class="material-symbols-outlined">settings</i> > **Providers**, then scroll down and select **Connect** next to **OpenAI**.
   ![Add the OpenAI provider](/images/manual/use-cases/opencode-openai-provider.png#bordered)

3. For the sign-in method, select **API key**.
   ![Select API key sign-in for OpenAI](/images/manual/use-cases/opencode-openai-api-key-signin.png#bordered){width=70%}

4. Paste your OpenAI API key and click **Continue**.
   ![Enter the OpenAI API key](/images/manual/use-cases/opencode-openai-api-key.png#bordered){width=70%}

5. Refresh the OpenCode page to load new OpenAI models.
   ![OpenAI provider connected](/images/manual/use-cases/opencode-openai-connected.png#bordered){width=70%}
</template>
</tabs>

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

OpenCode also offers a terminal-based UI (TUI). You can launch it in two ways:

- **Inside the OpenCode UI**: Open a terminal panel at the bottom of the OpenCode UI, similar to VS Code's integrated terminal. The agent chat stays visible above while the TUI runs in the panel.
- **From OpenCode Terminal**: Open OpenCode Terminal from Launchpad. It opens straight into a command line, without the chat UI.

<tabs>
<template #Inside-the-OpenCode-UI>

1. In OpenCode, click <i class="material-symbols-outlined">terminal_2</i> in the top-right to open a terminal panel at the bottom of the window.
   ![Open the terminal panel in OpenCode](/images/manual/use-cases/opencode-web-terminal.png#bordered)

2. In the terminal panel, run `opencode` to launch the TUI.
   ![Launch TUI in the OpenCode UI](/images/manual/use-cases/opencode-web-launch-tui.png#bordered)

3. Use the `/models` command to switch to Qwen3.5 9B.
   ![Select model in TUI](/images/manual/use-cases/opencode-web-tui-select-model.png#bordered)

4. Type your prompt directly. For example, ask OpenCode to improve the code.
   ![Chat in TUI](/images/manual/use-cases/opencode-web-tui-chat.png#bordered)

</template>
<template #From-OpenCode-Terminal>

1. Open OpenCode Terminal from Launchpad.

2. Run `opencode` to launch the TUI.
   ![Launch TUI in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-launch-tui.png#bordered)

3. Use the `/models` command to switch to Qwen3.5 9B.
   ![Select model in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-select-model.png#bordered)

4. Type your prompt directly. For example, use `@` to mention a file and ask OpenCode about it.
   ![Chat in OpenCode Terminal](/images/manual/use-cases/opencode-terminal-mention.png#bordered)

</template>
</tabs>

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
