---
outline: [2,3]
description: Learn how to install and configure Hermes Agent on Olares and connect it to Discord.
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, Hermes Agent, autonomous AI, self-improving AI, Discord bot, self-hosted
app_version: "1.3.8"
doc_version: "2.0"
doc_updated: "2026-05-29"
---

# Set up a self-directed AI agent with Hermes

Hermes Agent is a self-directed AI assistant that connects to your local models to execute system tasks, generate code, and manage workflows. It retains memory across sessions and creates reusable skills based on your interactions. By integrating it with messaging platforms like Discord, you can interact with your local AI agent remotely.

## Learning objectives

In this guide, you will learn how to:
- Install Hermes Agent on Olares.
- Configure Hermes Agent to connect to a local model.
- Interact with your agent directly via the terminal.
- Integrate with Discord for remote chat.

## Prerequisites

- Local model: [Ollama is installed](./ollama.md) with at least one tool-capable model downloaded and running. This tutorial uses `qwen3.5:9b`.
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.

## Install Hermes Agent

1. Open Market and search for "Hermes".

   ![Install Hermes Agent](/images/manual/use-cases/hermes-agent.png#bordered)

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:

    - Hermes CLI: The command line interface
    - Dashboard: The graphical dashboard

    ![Hermes entry points](/images/manual/use-cases/hermes-entry-points.png#bordered){width=50%}

:::tip Run multiple Hermes agents
Olares supports app cloning. If you want to run multiple independent AI agents for different tasks, you can clone the Hermes Agent app. For more information, see [Clone applications](../manual/olares/market/clone-apps.md).
:::

## Configure Hermes Agent

Run a quick setup to connect Hermes Agent to your local model.

### Step 1: Get model and endpoint details

1. Open the Ollama app from the Launchpad.
2. Check the installed models by running the following command:

    ```bash
    ollama list
    ```
3. Copy and save your model name exactly as shown in the **NAME** column. For example, `qwen3.5:9b`.
4. Check the context window of the model by running the following command:

    ```bash
    ollama ps
    ```

5. Copy and save the value as shown in the **CONTEXT** column. For example, `32768`.
6. Open Settings, go to **Applications** > **Ollama** > **Shared entrances** > **Ollama API**, and then copy the endpoint address. For example, `http://d54536a50.shared.olares.com`.

    ![Obtain Ollama API](/images/manual/use-cases/ollama-endpoint1.png#bordered){width=65%}

### Step 2: Run the setup wizard

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
    | API key  | Enter any text as a placeholder value, such as `ollama-local`.<br>The input remains hidden for security. |
    | Available models | Enter the number corresponding to your target model<br> from the generated list. |
    | Context length in tokens | <ul><li>If your model's context window is less than `65536`,<br>enter a value greater than `65536`.</li><li>If your model's context window is more than `65536`,<br>leave this field blank to auto-detect.</li></ul> |
    | Display name |  Enter a name for easy identification, such as `ollama-local`.|
    | Connect a messaging platform | Select **Skip - set up later with `hermes setup gateway`**. |

4. Keep the Hermes CLI window open for the next step.

## Interact with Hermes Agent

### Option 1: Terminal chat

The Terminal User Interface (TUI) runs directly in the Hermes CLI with no extra setup. It is ideal for quick tests.

1. When the setup wizard completes, it prompts you to launch `hermes chat`. Type `y` and press **Enter** to start the TUI.

   :::tip
   If you exited the wizard, manually start the chat by entering `hermes chat` in the Hermes CLI.
   :::

    ![Hermes setup complete](/images/manual/use-cases/hermes-setup-complete.png#bordered)

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

    b. Scroll down to **Bot Permissions**, set the permissions as shown in the following image. You can modify the settings later.

    ![Bot permissions](/images/manual/use-cases/hermes-bot-permissions.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server from **Add to server**, click **Continue**, and then click **Authorize**. 

    ![Add Discord bot to server](/images/manual/use-cases/hermes-add-bot.png#bordered){width=50%}

    The bot is authorized and added to your server.

    ![Bot added to server](/images/manual/use-cases/hermes-bot-added.png#bordered)

#### Step 3: Configure the messaging platform

Connect Hermes Agent to your Discord bot by configuring and running the Hermes gateway.

1. Open the Hermes CLI, and then enter the following command to start the configuration wizard:

   ```bash
   hermes gateway setup
   ```

2. Select **Discord** as your messaging platform.
3. Follow the prompts to configure the bot integration:

   - **Discord bot token**: Enter the bot token generated from your Discord Developer Portal. The input remains hidden.
   - **Allowed user IDs or usernames**: Enter your Discord user ID to restrict access to yourself.
   - (Optional) **Home channel ID**: Enter the ID of the Discord channel where the bot operates. 

4. Select **Done**.
5. When prompted to **Restart the gateway to pick up changes**, enter `y`.
6. Check your bot status in Discord. A green status icon confirms the bot is online, which means the gateway restarted successfully and the configuration is correct.

#### Step 4: Authorize your account

For security, the bot does not talk to unauthorized users. You must pair your Discord account with the bot.

1. Send a Direct Message to your new bot. The bot will reply with an error message containing a pairing code.
2. Open the Hermes CLI and enter the following command:

    ```bash
    hermes pairing approve discord {Your-Pairing-Code}
    ```

3. Once approved, you can start chatting with your agent in Discord. If you are talking to it in a channel, mention it first.

## Advanced configuration

To manually adjust parameters, edit the configuration files directly, and then restart the Hermes CLI to apply the changes. 

1. Open the Files app from the Launchpad.
2. Go to **Data** > **hermesagent** > **home**, and locate the configuration files, such as `config.yaml` and `.env`. 

The file structure and configuration options match the official defaults. For detailed parameter descriptions, see the [Hermes configuration guide](https://hermes-agent.nousresearch.com/docs/user-guide/configuration).


## Manage Olares with your Hermes Agent

Install the Olares CLI skills in Hermes Agent so your agent can manage files and applications on your Olares device. For example, ask it to list files, read logs, or install apps from Olares Market.

1. Open the Hermes CLI from the Launchpad.
2. Run the following commands one by one to install the required skills:

   ```bash
   hermes skills install clawhub/olares-shared --yes
   hermes skills install clawhub/olares-market --yes --force
   hermes skills install clawhub/olares-dashboard --yes --force
   hermes skills install clawhub/olares-settings --yes --force
   hermes skills install clawhub/olares-files --yes --force
   ```
   
3. After the installation finishes, run the following command to log in to your Olares account. Replace {your-olares-id} with your Olares ID:

   ```bash
   olares-cli profile login --olares-id {your-olares-id}
   ```

## Call Hermes Agent via OpenAI-compatible API

Hermes Agent provides an OpenAI-compatible API, allowing you to integrate it with other applications like OpenWebUI or Hermes Workspace.

### Step 1: Enable the Gateway API

1. Open Olares Settings, and then go to **Applications** > **Hermes Agent** > **Manage environment variables**.
2. Find **API_SERVER_ENABLED**, click click <i class="material-symbols-outlined">edit_square</i>, and then set its value to `true`.
3. Click **Confirm**.
4. Find **HERMES_API_SERVER_KEY**, click click <i class="material-symbols-outlined">edit_square</i>, and then enter the value. Ensure the key meets the following requirements:

    - It must be at least 8 characters long.
    - It can only contain letters, numbers, and these symbols (`-`, `_`, `.`).
    - It cannot be a predictable placeholder value, such as `your_api_key`.
  
5. Click **Confirm**, and then click **Apply**.

### Step 2: Get the Gateway API URL

1. Go to **Applications** > **Hermes Agent** > **Hermes Gateway API**.
2. Copy the endpoint URL. For example:
   ```
   https://baf3d7172.olaresdemo.olares.com
   ```

### Step 3: Verify the API is running

Open your browser and go to your gateway URL appended with `/health`:

```
https://{your-gateway-url}/health
```

If the API is enabled successfully, the page displays a response confirming the service is active. For example, `{"status": "ok", "platform": "hermes-agent"}`.


### Step 4: Connect your application

The following steps demonstrate how to connect with Open WebUI.

1. Open WebUI, click your profile icon and select **Admin Panel**.
2. Go to **Settings** > **Connections**.
3. On the right of **Manage OpenAI API Connections**, click <i class="material-symbols-outlined">add</i> to add a new connection.
4. Configure the settings as follows:
    - **API Base URL**: Enter your Hermes Gateway API URL and append `/v1`. For example, `https://baf3d7172.olaresdemo.olares.com/v1`.
    - **Auth**: Select **Bearer**, and then enter the `HERMES_API_SERVER_KEY` you set previously.
5. Click <i class="material-symbols-outlined">refresh</i> to check the connection, and then click **Save**.
6. Go to the **New chat** page and verify that **hermes-agent** appears in the model drop-down list. You can start chatting with it now.

## FAQs

### How to manually restart the Hermes gateway?

You can restart the gateway manually using one of the following methods:

- **Use the Hermes CLI**

    This is the fastest method. Open the Hermes CLI from the Launchpad, and then run the following command:

    ```bash
    restart-gateway
    ```
- **Use the Hermes Dashboard**

    This is the most intuitive method. Open the Hermes Dashboard from the Launchpad, and then select **Restart Gateway** in the left sidebar.

- **Use Control Hub**

    Open Control Hub, and then go to **Browse** > **{username}** > **hermesagent-{username}** > **Deployments** > **hermesagent**, and then click **Restart** in the upper-right corner. This method completely restarts the entire service, which takes slightly longer.

### Why does the gateway fail to start after configuring the API?

Your gateway will fail to start if the configured API key does not meet Hermes security requirements. When this happens:
- The Gateway API displays an `upstream connect error`.
- The system logs show the message `[Api_Server] Refusing to start: API_SERVER_KEY is set to a placeholder value`.

To resolve this issue, reset your `HERMES_API_SERVER_KEY`:
1. Open Olares Settings, and then go to **Applications** > **Hermes Agent** > **Manage environment variables**.
2. Set a new key for `HERMES_API_SERVER_KEY` that meets the following requirements:

    - It must be at least 8 characters long.
    - It can only contain letters, numbers, and these symbols (`-`, `_`, `.`).
    - It cannot be a predictable placeholder value, such as `your_api_key`.
    
 3. Click **Save**, and then click **Apply**.

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
- [How to find the Channel ID number](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID)
- [OpenClaw](openclaw.md)
