---
outline: [2, 3]
description: Run NemoClaw on Olares with a local LLM such as Gemma4. Set up an always-on AI agent backed by the NVIDIA OpenShell runtime, with no cloud API required.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, NVIDIA, OpenShell, OpenClaw, local LLM, AI assistant, Discord, web search, ClawHub, skills, plugins
app_version: "1.0.8"
doc_version: "1.1"
doc_updated: "2026-05-11"
---

# Run NemoClaw with a local LLM

NemoClaw is an open-source reference stack from NVIDIA that runs OpenClaw with the NVIDIA OpenShell runtime bundled.

This guide walks you through running NemoClaw on Olares with the Gemma4 26B model app as the backend LLM.

:::warning Alpha software
NemoClaw is an early preview release from NVIDIA and is not recommended for production use. For official updates and community feedback, see [NVIDIA/NemoClaw](https://github.com/NVIDIA/NemoClaw).
:::

## Learning objectives

In this guide, you will learn how to:

- Install and configure NemoClaw with a local LLM.
- Keep the model loaded for always-on responses.
- Start your first chat with the agent.
- Connect the agent to Discord for remote chat.
- Enable real-time web search.

## Prerequisites

- A local model app installed and running on your Olares device.
- Admin privileges to install apps from Market and edit application settings.

## Get the model name and endpoint URL

NemoClaw needs the model name and its shared endpoint URL during installation.

1. Open your model app from Launchpad and note the model name shown on the page. In this example, it's `gemma4:26b`.

   ![Model name shown in the model app](/images/manual/use-cases/gemma4-26b-downloaded.png#bordered)

2. Open Settings, then go to **Applications** > **Gemma4 26B Q4_K_M (Ollama)**.
3. In **Shared entrances**, select **Gemma4 26B Q4_K_M** to view the endpoint URL.

   ![Get shared endpoint](/images/manual/use-cases/gemma4-26b-shared-laresprime.png#bordered){width=90%}

4. Note the shared endpoint. For example:

   ```plain
   http://2e53d5230.shared.olares.com
   ```

   :::tip Why the shared endpoint?
   The URL on the model app's main page is user-specific and routes through the browser. The shared endpoint is reachable from other apps on Olares without sign-in or CORS issues, which is what NemoClaw needs.
   :::

## Install NemoClaw

1. Open Market and search for "NemoClaw".

   ![NemoClaw in Market](/images/manual/use-cases/nemoclaw.png#bordered)

2. Click **Get**, then **Install**.
3. When prompted, set the environment variables:

   - **NEMOCLAW_ENDPOINT_URL**: Enter or paste the shared endpoint URL, and append `/v1`, such as `http://2e53d5230.shared.olares.com/v1`.
   - **NEMOCLAW_MODEL**: Enter or paste the model name, such as `gemma4:26b`.

   ![Set environment variables for NemoClaw](/images/manual/use-cases/nemoclaw-set-environment-variables.png#bordered){width=70%}

   :::tip
   You can change these environment variables later in **Settings** > **Applications** > **NemoClaw** > **Manage environment variables**.
   :::

4. Click **Confirm** and wait for installation to complete.

   Installation takes about 15 minutes, depending on your network. During this time, NemoClaw installs the NVIDIA OpenShell runtime and runs the initial agent onboarding.

   :::warning
   Keep the model app running during installation. The initial onboarding requires the model to be reachable, and the installation won't complete if the model stops or becomes unavailable.
   :::

When the installation finishes, two shortcuts appear on Launchpad:

- **NemoClaw CLI**: The terminal interface for running NemoClaw and OpenClaw commands.
- **OpenClaw Web UI**: The browser-based dashboard for OpenClaw.

## Keep the model loaded (optional)

By default, the local LLM unloads from memory after 5 minutes of inactivity, and the next reply has to wait for the model to reload. For an always-on agent, enable the keep-alive setting on the model app to keep it resident in memory.

1. Open Settings and go to **Applications** > **Gemma4 26B Q4_K_M** > **Manage environment variables**.
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

NemoClaw runs OpenClaw inside a sandboxed runtime, so you must configure the channel from within the runtime shell.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

   Wait until the terminal shows the sandbox prompt, such as `sandbox@my-assistant:~$`.

3. Run the channel configuration wizard:

   ```bash
   openclaw config --section channels
   ```

4. Follow the prompts to add Discord:
   | Settings | Option |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Channels | Configure/link |
   | Select a channel | Discord (Bot API) |
   | How do you want to provide this Discord bot token? | Enter Discord bot token, and paste the token from Step 1. |
   | Configure Discord channels access | Yes |
   | Discord channels access | Open (allow all channels) |

5. When finished, when prompted to select a channel, select **Finished**.

6. When prompted to configure DM (Direct Message) access policies, select **Pairing**.

:::info Discord channel stuck in `startup-not-ready`
If the Discord channel shows `startup-not-ready` in the OpenClaw Web UI, restart the gateway. See [Common issues](nemoclaw-common-issues.md#discord-channel-stuck-in-startup-not-ready-state) for the steps.
:::

### Step 4: Authorize your Discord account

For security, the bot doesn't respond to unauthorized users. You must pair your Discord account with the bot.

1. Open Discord and send a direct message to your bot.

   The bot replies with a pairing code and a command.

   ![Pairing code from the bot DM](/images/manual/use-cases/nemoclaw-discord-pairing-code.png#bordered)

2. Switch back to the NemoClaw CLI sandbox shell and use the command provided by the bot to approve the pairing. For example:

   ```bash
   openclaw pairing approve discord FY6PAVY8
   ```
   When you see the following, it means the Discord is authorized:
   ```text
   Approved discord sender 1277468602303385654.
   ```

3. After approval, you can chat with your agent directly in Discord.

   ![Chat with the agent in Discord](/images/manual/use-cases/nemoclaw-discord-chat.png#bordered)

## Enable web search

By default, the agent answers only from its training data. To let it fetch real-time internet information, you can use a web search provider. The following uses SearXNG as an example, which you can install as a self-hosted instance from Olares Market.

1. Install SearXNG from Olares Market.
2. Open Settings, then go to **Applications** > **SearXNG**.
3. In **Shared entrances**, select **SearXNG** to view the endpoint URL.

   ![Get shared endpoint](/images/manual/use-cases/searxng-shared-laresprime.png#bordered){width=90%}
4. Copy the shared endpoint. For example:

   ```plain
   http://d1236e020.shared.olares.com
   ```

5. Open the NemoClaw CLI app from Launchpad.
6. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

7. Run the web tool configuration wizard:

   ```bash
   openclaw config --section web
   ```

8. Configure as follows:

   | Settings | Option |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Enable web_search | Yes |
   | Search provider | SearXNG |
   | SearXNG Base URL | Paste the shared SearXNG endpoint from Step 4 |
   | Enable web_fetch (keyless HTTP fetch) | Yes |

9. To verify, ask your agent a question that requires real-time information. For example:

   ```text
   What are today's top tech news headlines?
   ```

   The agent should fetch and cite live web results.
   ![Web search results](/images/manual/use-cases/nemoclaw-web-search-result.png#bordered){width=90%}

## Install skills

Skills add capabilities to your agent, such as managing Olares files and apps or integrating with Google Workspace.

1. Open the OpenClaw Web UI from Launchpad.
2. Go to **Skills**.
3. Search for the skill in ClawHub and click **Install**.
4. Open the chat page in the OpenClaw Web UI and run `/reset` to start a new session so the agent picks up the newly installed skill. If you've configured channels such as Discord, also run `/reset` in each channel conversation.

   :::tip
   You can also install skills from the NemoClaw CLI sandbox using `openclaw config --section skills`.
   :::

For walkthroughs of common skills, see:

- [Manage Olares with Olares CLI](nemoclaw-olares-cli.md): Let the agent operate files and apps on your Olares device through natural language.
- [Integrate with Google Workspace](nemoclaw-google-workspace.md): Connect Gmail, Calendar, and Drive via the gog skill.

For more on managing skills, see [Manage skills and plugins](openclaw-skills.md).

## Install plugins

Plugins extend OpenClaw with additional channels and integrations.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

3. Install the BlueBubbles plugin:

   ```bash
   openclaw plugins install @openclaw/bluebubbles
   ```

For other plugins, use the standard `openclaw plugins list` and `openclaw plugins install <name>` commands inside the runtime. For details, see [Manage skills and plugins](openclaw-skills.md).

## Common issues

For a list of common issues and workarounds, see [Common issues](nemoclaw-common-issues.md).

## Learn more

- [NVIDIA NemoClaw](https://build.nvidia.com/nemoclaw): Official reference stack and documentation from NVIDIA.
- [OpenClaw](openclaw.md): Set up OpenClaw features such as persona setup.
