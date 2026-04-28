---
outline: [2, 3]
description: Learn how to install and configure Hermes Agent on Olares and connect it to Discord.
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, AI agent, Discord bot, Olares, self-hosted
app_version: "0.1.1"
doc_version: "1.0"
doc_updated: "2026-04-28"
---

# Set up your terminal AI agent with Hermes

Hermes is a powerful terminal-based AI agent that connects to local models and executes various tasks, from coding to system operations. By integrating it with messaging platforms like Discord, you can interact with your local AI agent remotely.

## Learning objectives

In this guide, you will learn how to:
- Install Hermes Agent on Olares.
- Configure Hermes Agent to connecto to a local model.
- Interact with your agent directly via the terminal.
- Integrate with Discord for remote chat.

## Prerequisites

- Local model: [Ollama is installed](./ollama.md) with at least one tool-capable model downloaded and running. This tutorial uses `qwen3.5:9b`.
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.

## Install Hermes Agent

1. From the Olares Market, search for "Hermes".

   ![Install Hermes Agent](/images/manual/use-cases/hermes-agent.png#bordered)

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:

    - Dashboard: The graphical dashboard
    - Hermes CLI: The command line interface

    ![Hermes entry points](/images/manual/use-cases/hermes-entry-points.png#bordered){width=50%}

:::tip Run multiple Hermes agents
Olares supports app cloning. If you want to run multiple independent AI agents for different tasks, you can clone the Hermes Agent app. For more information, see [Clone applications](../manual/olares/market/clone-apps.md).
:::

## Initialize Hermes Agent

Run a quick setup to connect the agent to your local environment.

### Step 1: Get the model endpoint

To connect Hermes Agent to Ollama, get the shared entrance URL and your model details:

1. Check the installed models by running the following command:

    ```bash
    ollama list
    ```
2. Copy and save your model name exactly as shown in the **NAME** column. For example, `qwen3.5:9b`.
3. Check the context window of the model by running the following command:

    ```bash
    ollama ps
    ```

4. Copy and save the value as shown in the **CONTEXT** column. For example, `128000`.
5. Open Settings, go to **Applications** > **Ollama** > **Shared entrances** > **Ollama API**, and then copy the endpoint address. For example, `http://d54536a50.shared.olares.com`.

    ![Obtain Ollama API](/images/manual/use-cases/ollama-endpoint1.png#bordered){width=65%}

### Step 2: Run the setup wizard

Configure Hermes to connect to your local model using the interactive setup wizard.

1. Open the Hermes CLI app from the Launchpad.
2. Enter the following command to start the configuration wizard:

   ```bash
   hermes setup
   ```

3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press **Enter** to confirm.

    | Settings   | Option   |
    |:-----------|:---------|
    | How would you like to set up Hermes | Select **Quick setup - provider, model & messaging (recommended)**. |
    | Select provider | Select **Custom endpoint (enter URL manually)**.  |
    | API base URL  | Enter your model's API address and append `/v1` to the end.<br>For example, `http://d54536a50.shared.olares.com/v1`.  |
    | API key  | Enter a placeholder value, such as `ollama-local`.<br>The input remains hidden for security. |
    | Available models | Enter the number corresponding to your target model<br> from the generated list. |
    | Context length in tokens | Enter a value greater than `65536` or the exact context window<br> of your running model. Do not leave this field blank. |
    | Display name |  Enter a name for easy identification, such as `ollama-local`.|
    | Connect a messaging platform | Select **Skip - set up later with `hermes setup gateway`**. |

4. Keep the Hermes CLI window open for the next step.

## Interact with Hermes Agent

### Option 1: Terminal chat

The Terminal User Interface (TUI) runs directly in the Hermes CLI. It requires no extra setup and is ideal for quick tests or when you are already working on Olares.

1. When the setup wizard completes, it prompts you to launch hermes chat. Type `y` and press **Enter** to start the TUI.

   :::tip
   If you exited the wizard, manually start the chat by entering `hermes chat` in the Hermes CLI.
   :::

    ![Hermes setup complete](/images/manual/use-cases/hermes-setup-complete.png#bordered){width=70%}

2. Send a message, such as `what can you do`, to verify that the agent responds correctly.

    ![Hermes TUI chat](/images/manual/use-cases/hermes-tui.png#bordered)

3. To close the TUI and return to the standard CLI, type `/exit`, and then press **Enter**.

### Option 2: Remote chat via Discord

To chat with your agent remotely, connect it to a Discord bot.

#### Step 1: Create a Discord bot

1. Log in to the [Discord Developer Portal](https://discord.com/developers/home) with your Discord account.
2. Select **Applications** from the left sidebar, and then click **New Application**.

    ![New application in Discord developer portal](/images/manual/use-cases/hermes-new-app.png#bordered)

3. Enter a name for the new app, agree to the terms, and then click **Create**.

    ![Create an application window](/images/manual/use-cases/hermes-create-app.png#bordered){width=40%}

4. From the left sidebar, select **Bot**.
5. Scroll down to the **Privileged Gateway Intents** section and enable the following settings:

    - Presence Intent
    - Server Members Intent
    - Message Content Intent

    ![Privileged gateway intents](/images/manual/use-cases/hermes-privileged-gateway-intents.png#bordered)

6. Click **Save Changes**.
7. Scroll up to the **Token** section, click **Reset Token**, and then copy the generated token for your Discord bot. You need the token for messaging platform configuration later in the Hermes CLI.

    ![Reset token](/images/manual/use-cases/hermes-reset-token.png#bordered)

#### Step 2: Invite the bot to your server

1. From the left sidebar, select **OAuth2**, and then find the **OAuth2 URL Generator** section:

    a. In **Scopes**, select **Bot** and **applications.commands**.

    ![OAuth2 URL Generator](/images/manual/use-cases/hermes-oauth21.png#bordered)

    b. Scroll down to **Bot Permissions**, set as the following image. You can modify the settings later.

    ![Bot permissions](/images/manual/use-cases/hermes-bot-permissions.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server from **Add to server**, click **Continue**, and then click **Authorize**. 

    ![Add Discord bot to server](/images/manual/use-cases/hermes-add-bot.png#bordered){width=50%}

    The bot is authorized and added to your server.

    ![Bot added to server](/images/manual/use-cases/hermes-bot-added.png#bordered)

#### Step 3: Configure the messaging platform

Connect Hermes Agent to your Discord bot by adding its configuration in the Hermes CLI.

1. Open the Hermes CLI, and then enter the following command:

   ```bash
   hermes gateway setup
   ```

2. Select **Discord** as your messaging platform.
3. Follow the prompts to configure the bot integration:
   - **Bot token**: Enter the bot token generated from your Discord Developer Portal. The input remains hidden.
   - **Allowed user IDs or usernames**: Enter your Discord user ID to restrict the access to yourself.
   - **Home channel ID**: Enter the ID of the Discord channel where the bot operates. 
   
        For more information, see [How to find the Channel ID number](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID).

4. Enter the following command to start the gateway:

   ```bash
   hermes gateway run
   ```
   :::warning
   The gateway runs in the foreground. Leave the **Hermes CLI** window open to keep the Discord bot online. If you close the window, the bot disconnects.
   :::

5. Send a Direct Message to your new bot. You are now communicating with your Hermes agent remotely.

## Advanced configuration

To manually adjust parameters, edit the configuration files directly. 

1. Open the Files app from the Launchpad.
2. Go to **Data** > **hermesagent** > **home**, and locate the configuration files, such as `config.yaml` and `.env`. 

The file structure and configuration options match the official defaults. For detailed parameter descriptions, see the [Hermes configuration guide](https://hermes-agent.nousresearch.com/docs/user-guide/configuration).

## Next steps

After you complete the basic setup, explore the official resources to expand your agent's capabilities:
- For advanced terminal commands, see [CLI Interface](https://hermes-agent.nousresearch.com/docs/user-guide/cli).
- For integration with Slack, Telegram, and other platforms, see [Messaging Platforms](https://hermes-agent.nousresearch.com/docs/user-guide/messaging/).
- For installation of skills, see [Skills System](https://hermes-agent.nousresearch.com/docs/user-guide/features/skills).
- For installation of tools, see [Tools & Toolsets](https://hermes-agent.nousresearch.com/docs/user-guide/features/tools).
- For information about best practices and optimization, see [Tips & Best Practices](https://hermes-agent.nousresearch.com/docs/guides/tips).

:::info Note on sudo commands
The Hermes CLI in the Olares environment does not support `sudo` commands. Ignore any steps in the official documentation that require `sudo` privileges.
:::

## Learn more

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
- [OpenClaw](openclaw.md)
