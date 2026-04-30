---
outline: [2, 3]
description: Run NemoClaw on Olares with a local LLM such as Qwen3.5. Set up an always-on AI agent backed by the NVIDIA OpenShell runtime, with no cloud API required.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, NVIDIA, OpenShell, Nemotron, OpenClaw, local LLM, AI assistant, self-hosted AI, Qwen, Discord, Brave Search, web search, ClawHub, skills, plugins, Lark
app_version: "1.0.3"
doc_version: "1.1"
doc_updated: "2026-04-30"
---

# Run NemoClaw with a local LLM

NemoClaw is an open-source reference stack from NVIDIA that runs OpenClaw always-on assistants with a single command. It bundles the NVIDIA OpenShell runtime for policy-based privacy and security guardrails, giving you more control over your agent's behavior and data handling.

This guide walks you through running NemoClaw on Olares with the Qwen3.5 27B Q4_K_M model app as the backend LLM.

## Prerequisites

- A local model app installed and running on your Olares device.
- Admin privileges to install apps from Market and edit application settings.

## Get the model name and endpoint URL

NemoClaw needs the model name and its shared endpoint URL during installation.

1. Open your model app from Launchpad and note the model name shown on the page. In this example, it's `qwen3.5:27b-q4_K_M`.

   ![Model name shown in the model app](/images/one/qwen3.5-27b-downloaded.png#bordered)

2. Open Settings, then go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)**.
3. In **Shared entrances**, select **Qwen3.5 27B Q4_K_M** to view the endpoint URL.

   ![Get shared endpoint](/images/manual/use-cases/deerflow2-shared-entrance.png#bordered){width=90%}

4. Note the shared endpoint. For example:

   ```plain
   http://94a553e00.shared.olares.com
   ```

   :::tip Why the shared endpoint?
   The URL on the model app's main page is user-specific and routes through the browser. The shared endpoint is reachable from other apps on Olares without sign-in or CORS issues, which is what NemoClaw needs.
   :::

## Install NemoClaw

1. Open Market and search for "NemoClaw".

   ![NemoClaw](/images/manual/use-cases/nemoclaw.png#bordered)

2. Click **Get**, then **Install**.
3. When prompted, set the environment variables:

   - **NEMOCLAW_ENDPOINT_URL**: Enter or paste the shared endpoint URL, and append `/v1`, such as `http://94a553e00.shared.olares.com/v1`.
   - **NEMOCLAW_MODEL**: Enter or paste the model name, such as `qwen3.5:27b-q4_K_M`.

   :::tip
   You can change these environment variables later in **Settings** > **Applications** > **NemoClaw** > **Manage environment variables**.
   :::

   ![Set environment variables for NemoClaw](/images/manual/use-cases/nemoclaw-set-environment-variables.png#bordered){width=70%}

4. Click **Confirm** and wait for installation to complete.

   Installation takes about 15 minutes, depending on your network. During this time, NemoClaw installs the NVIDIA OpenShell runtime and runs the initial agent onboarding.

When the installation finishes, two shortcuts appear on Launchpad:

- **NemoClaw CLI**: The terminal interface for running NemoClaw and OpenClaw commands.
- **OpenClaw Web UI**: The browser-based dashboard for OpenClaw.

## Optional: Keep the model loaded

By default, the local LLM unloads from memory after 5 minutes of inactivity, and the next reply has to wait for the model to reload. For an always-on agent, enable the keep-alive setting on the model app to keep it resident in memory.

1. Open Settings and go to **Applications** > **Qwen3.5 27B Q4_K_M (Ollama)** > **Manage environment variables**.
2. Find **KEEP_ALIVE**, click <i class="material-symbols-outlined">edit_square</i>, set the value to **true**, and click **Confirm**.

   ![Enable KEEP_ALIVE for the model app](/images/manual/use-cases/keep-alive-enable.png#bordered){width=80%}

3. Click **Apply**.

:::tip When to leave KEEP_ALIVE unset
Keeping the model loaded consumes VRAM continuously. If you only use the agent occasionally and don't mind the cold-start delay, leave **KEEP_ALIVE** unset.
:::

## Start your first chat

NemoClaw lets you chat with your agent in either the OpenClaw Web UI or the OpenClaw TUI inside the NemoClaw CLI. Because the model and endpoint were configured during installation, you can skip the manual onboarding and go straight to a session.

<tabs>
<template #In-OpenClaw-Web-UI>
1. Open the OpenClaw Web UI app from Launchpad. You will be taken directly to the Chat interface.

2. Send a test message, such as `Hi`.

   The first response can take about 30 seconds while the model loads into memory. Subsequent replies are much faster. Once the model is loaded, you can ask the agent about itself to confirm the configuration. For example:

   ```text
   How are you, and what model are you running on?
   ```

   ![NemoClaw chat in OpenClaw Web UI](/images/manual/use-cases/nemoclaw-openclaw-chat-test.png#bordered)

</template>

<template #In-NemoClaw-CLI>
1. Open the NemoClaw CLI app from Launchpad.
2. Run the following command to connect to the sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

   Wait until the terminal shows the sandbox prompt.

   ![Sandbox connected](/images/manual/use-cases/nemoclaw-connect.png#bordered)

3. Launch the OpenClaw TUI:

   ```bash
   openclaw tui
   ```

   ![Launch the OpenClaw TUI in the sandbox shell](/images/manual/use-cases/nemoclaw-tui.png#bordered)

4. Send a test message, such as `Hi`.

   The first response can take about 30 seconds while the model loads into memory. Subsequent replies are much faster. Once the model is loaded, you can ask the agent about itself to confirm the configuration. For example:

   ```text
   How are you, and what model are you running on?
   ```

   ![NemoClaw chat session](/images/manual/use-cases/nemoclaw-chat-test.png#bordered)

</template>
</tabs>

## Integrate with Discord

To chat with your NemoClaw agent remotely, connect it to a Discord bot. You need a Discord account and a server where you have permission to add bots.

### Step 1: Create a Discord bot

1. Log in to the [Discord Developer Portal](https://discord.com/developers/applications) with your Discord account.
2. Click **New Application**.

   ![New application in Discord developer portal](/images/manual/use-cases/new-app.png#bordered){width=90%}

3. Enter a name for the new app, agree to the terms, and click **Create**.

   ![Create an application window](/images/manual/use-cases/create-app.png#bordered){width=40%}

4. From the left sidebar, select **Bot**.
5. Scroll down to the **Privileged Gateway Intents** section and enable the following settings:

   - Presence Intent
   - Server Members Intent
   - Message Content Intent

6. Click **Save Changes**.
7. Scroll up to the **Token** section, click **Reset Token**, and copy the generated token. You need this token in Step 3.

   ![Reset token](/images/manual/use-cases/reset-token.png#bordered)

### Step 2: Invite the bot to your server

1. From the left sidebar, select **OAuth2** and find the **OAuth2 URL Generator** section.

    a. In **Scopes**, select **Bot** and **applications.commands**.

    ![OAuth2 URL Generator](/images/manual/use-cases/oauth21.png#bordered)

    b. Scroll down to **Bot Permissions** and configure them as shown. You can adjust these later.

    ![Bot permissions](/images/manual/use-cases/bot-permissions1.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server from **Add to server**, click **Continue**, and click **Authorize**.

   The bot is authorized and added to your server.

   ![Bot added to server](/images/manual/use-cases/bot-added.png#bordered)

### Step 3: Configure the Discord channel

NemoClaw runs OpenClaw inside a sandboxed runtime, so channel configuration must be done from within the runtime shell.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

   Wait until the terminal shows the sandbox prompt, such as `sandbox@my-assistant:~$`.

3. Run the channel configuration wizard:

   ```bash
   openclaw configure --section channels
   ```

4. Follow the prompts to add Discord:

   - **Channel**: Discord
   - **Bot token**: Paste the token from Step 1.
   - **Channel access**: Open
   - **DM policy**: Pairing

### Step 4: Authorize your Discord account

For security, the bot doesn't respond to unauthorized users. You must pair your Discord account with the bot.

1. Open Discord and send a Direct Message (DM) to your bot.

   The bot replies with a pairing code, such as `M9ZEHYT7`.

   ![Pairing code from the bot DM](/images/manual/use-cases/nemoclaw-discord-pairing-code.png#bordered)

2. Switch back to the NemoClaw CLI sandbox shell and approve the pairing:

   ```bash
   openclaw pairing approve discord {Your-Pairing-Code}
   ```

   For example:

   ```bash
   openclaw pairing approve discord M9ZEHYT7
   ```

3. After approval, you can chat with your agent directly in Discord.

   ![Chat with the agent in Discord](/images/manual/use-cases/nemoclaw-discord-chat.png#bordered)

## Enable web search

By default, the agent answers only from its training data. To let it fetch real-time internet information, enable the web search tool with Brave Search.

You need a [Brave Search API](https://brave.com/search/api/) key to proceed. The free "Data for Search" tier is usually enough for personal use.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

3. Run the web tool configuration wizard:

   ```bash
   openclaw configure --section web
   ```

4. Configure as follows:

   | Settings | Option |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Enable web_search | Yes |
   | Search provider | Brave Search |
   | Brave Search API key | Your Brave Search API key |
   | Enable web_fetch (keyless HTTP fetch) | Yes |

5. To verify, ask your agent a question that requires real-time information. For example:

   ```text
   What are today's top tech news headlines?
   ```

   The agent should fetch and cite live web results.

## Install skills

Skills add new capabilities to the agent. Install the `clawhub` skill first, which lets you search and install more skills from the [ClawHub](https://clawhub.ai/) registry.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

3. Start the configuration wizard:

   ```bash
   openclaw config
   ```

4. Navigate the prompts to install `clawhub`:

   | Settings | Option |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Select settings to configure | Skills |
   | Configure skills now | Yes |
   | Install missing skill dependencies | Navigate to **clawhub**, press **Space** to select, then press **Enter**. |
   | Preferred node manager for skill installs | npm. Wait for `Installed clawhub` before proceeding. |
   | Set [API_KEY] for [skill] | Select **No** for all. |

5. Select **Continue** at **Select sections to configure**. The `Configure complete` message indicates the setup is finished.

After `clawhub` is installed, run `openclaw skills search <name>` and `openclaw skills install <id>` inside the runtime to add more skills. For details, see [Manage skills and plugins](openclaw-skills.md).

## Install plugins

Plugins extend OpenClaw with additional channels and integrations.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

3. Install the Lark plugin:

   ```bash
   npx -y @larksuite/openclaw-lark install
   ```

For other plugins, use the standard `openclaw plugins list` and `openclaw plugins install <name>` commands inside the runtime. For details, see [Manage skills and plugins](openclaw-skills.md).

## Known issues

### Duplicate messages

Each message you send and each agent reply might appear twice in the chat. This is a known issue in the current NemoClaw version on Olares and doesn't affect the underlying agent state.

## Learn more

- [NVIDIA NemoClaw](https://build.nvidia.com/nemoclaw): Official reference stack and documentation from NVIDIA.
- [OpenClaw](openclaw.md): Set up OpenClaw features such as persona setup.
- [Common issues](openclaw-common-issues.md): Troubleshoot model timing and context window settings.
