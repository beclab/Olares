---
outline: [2, 3]
description: Learn how to install, configure, and integrate OpenClaw with Discord.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning
---

# OpenClaw

OpenClaw is a personal AI assistant that is designed to run on your local device. It connects directly to the messaging apps like Discord and Slack, and allows you to interact with it right in the app. 

It acts as an "always-on" operator that can execute real tasks, such as searching and sending documents, managing calendars, or browsing webpages.

## Learning objectives

By the end of this tutorial, you are be able to:

- Install and initialize the OpenClaw environment.
- Pair and connect the OpenClaw CLI and the Control UI.
- Configure OpenClaw to use the local AI model Ollama.
- Integrate OpenClaw with Discord.
- Enable the web search capability using Brave Search.

## Prerequisites

- Local model: Ensure Ollama is installed and running. You must have a tool-capable model installed, such as `glm-4.7-flash`, `qwen3`, and `llama3.1`. This tutorial uses `llama3.1:8b`.
- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.
- (Optional) Brave search API key: Required for the agent to search the web for real-time information. 

    :::tip Tip
    You can obtain a free API key from the [Brave Search API](https://brave.com/search/api/). The free tier of the "Data for Search"" plan is usually sufficient for personal use.
    :::

## Install OpenClaw

1. From the Olares Market, search for "OpenClaw".

    ![Search for OpenClaw from Market](/images/manual/use-cases/find-openclaw.png#bordered){width=90%}

2. Click **Get**, and then click **Install**. When the installation finishes, two shortcuts appear in the Launchpad:
    - **OpenClaw CLI**: The command line interface
    - **Control UI**: The graphical dashboard

    ![OpenClaw entry points](/images/manual/use-cases/openclaw-entry-points.png#bordered){width=30%}

## Initialize OpenClaw

Run a quick setup for the agent in the OpenClaw CLI.

1. Open the OpenClaw CLI app from the Launchpad.
2. Enter the following command to start the onboarding wizard.

    ```bash
    openclaw onboard
    ```
    ![OpenClaw onboarding wizard](/images/manual/use-cases/openclaw-wizard.png#bordered){width=80%}

3. The wizard guides you through a series of steps. Use the arrow keys to navigate and press Enter to confirm.

    :::tip Note on configurations
    To get you started quickly, this tutorial skips several advanced settings in the wizard. You can configure or modify them later.
    :::

    | Settings | Option |
    |:-------- |:-------|
    | I understand this is powerful and inherently <br>risky. Continue? | Yes |
    | Onboarding mode | QuickStart |
    | Config handling | Use existing values |
    | Model/auth provider | Skip for now |
    | Filter models by provider | All providers |
    | Default model | Keep current |
    | Select channel | Skip for now |
    | Configure skills now | No |
    | How do you want to hatch your bot | Do this later |
    
4. After you complete the wizard, restart OpenClaw CLI.
5. Enter the following command to view and copy the gateway token displayed:

    ```bash
    openclaw config get gateway.auth.token
    ```

6. Keep the OpenClaw CLI window open for the next step.

## Pair device

Connect the Control UI to the OpenClaw CLI to use the graphical dashboard.

1. Open the Control UI app from the Launchpad.
2. On the **Overview** page, in the **Gateway Access** panel, specify the following settings:
    - **Gateway Token**: Enter the token you copied in the previous step.
    - **Default Session Key**: Enter `agent:main:main`.
3. Click **Connect**. 

    The connection error `disconnected[1008]:pairing required` occurs. This is expected and means the device connection is waiting for approval.
4. Return to the OpenClaw CLI window and enter the following command:
    ```bash
    openclaw devices list
    ```
5. In the **Pending** table, find the **Request** ID associated with your current device.

    :::info
    The Request ID has a time limit. If the authorization fails, re-run `openclaw devices list` to obtain a new valid ID.
    :::

    ![View pending device request](/images/manual/use-cases/pending-request.png#bordered)
    
6. Authorize the device by entering the following command:

    ```bash
    openclaw devices approve {RequestID}
    ```
7. When the terminal displays `Approved {DeviceID}`, return to the Control UI. Now the **STATUS** in the **Snapshot** panel should be **Connected**.

    ![Health OK](/images/manual/use-cases/openclaw-connected.png#bordered)

## Configure local AI model

1. In the Control UI, select **Config** from the left sidebar.
2. Switch to the **Raw** tab to edit the configuration JSON file directly.
3. Find the `agents` section and update the `defaults` block to specify your primary model. Ensure that the model name matches the one installed in Ollama.

    ```json
    "agents": {
        "defaults": {
        "model": {
            "primary": "ollama/llama3.1:8b"
        },
        "workspace": "/home/node/.openclaw/workspace",
        "maxConcurrent": 4,
        "subagents": {
            "maxConcurrent": 8
        }
        }
    },
    ```
4. Click **Save** in the upper-right corner. The system validates the config and restarts automatically to apply the changes.

## Integrate with Discord

To chat with your agent remotely, connect it to a Discord bot.

### Step 1: Create a Discord bot

1. Log in to the [Discord Developer Portal](https://discord.com/developers/applications) with your Discord account.
2. Click **New Application**.
    ![Search for OpenClaw from Market](/images/manual/use-cases/new-app.png#bordered){width=90%}

3. Enter a name for the new app, agree to terms, and then click **Create**.

    ![Search for OpenClaw from Market](/images/manual/use-cases/create-app.png#bordered){width=40%}

3. From the left sidebar, select **Bot**.
4. Scroll down to the **Privileged Gateway Intents** section and enable the following settings:

    - Presence Intent
    - Server Members Intent
    - Message Content Intent
5. Click **Save Changes**.
6. Scroll up to the **Token** section, click **Reset Token**, and then copy the generated token for your Discord bot. You need the token for channel configuration later in Control UI.

    ![Reset token](/images/manual/use-cases/reset-token.png#bordered)

### Step 2. Invite the bot to server

1. From the left sidebar, select **OAuth2**, and then find the **OAuth2 URL Generator** section:

    a. In **Scopes**, select **Bot** and **applications.commands**.

    ![OAuth2 URL Generator](/images/manual/use-cases/oauth2.png#bordered)

    b. In **Bot Permissions**, set as the following image. You can modify the settings later.

    ![Bot permissions](/images/manual/use-cases/bot-permissions.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server, and then click **Authorize**. The bot is added to your server.

    ![Bot added to server](/images/manual/use-cases/bot-added.png#bordered)

### Step 3: Configure channel

Configure the Discord channel in Control UI.

1. Return to the **Control UI** > **Config** > **Raw** tab.
2. Find the `channels` section:

    a. Update with your Discord bot token.

    b. Enable Discord DM (Direct Messages) and set the Discord DM Policy to **Pairing**.

    ```json
    "channels": {
        "discord": {
        "enabled": true,
        "token": "{YOUR_BOT_TOKEN}",
        "allowBots": true,
        "dm": {
            "enabled": true,
            "policy": "pairing"
        }
        }
    },
    ```

    ![Discord channel added](/images/manual/use-cases/channels.png#bordered)

3. Click **Save**.
4. From the left sidebar, select **Channels**. On the Discord card, **Probe ok** indicates successful connection.

   ![Probe OK](/images/manual/use-cases/probe-ok.png#bordered)

### Step 4: Authorize your account

For security, the bot does not talk to unauthorized users. You must pair your Discord account with the bot.

1. Open Discord and send a Direct Message to your new bot. The bot will reply with an error message containing a Pairing Code.
2. Open the OpenClaw CLI and enter the following command:
    ```bash
    openclaw pairing approve discord {Your-Pairing-Code}
    ```
3. Once approved, you can start chatting with your agent in Discord.

## (Optional) Enable web search

By default, OpenClaw answers questions only based on its training data, which means it doesn't know about current events or real-time news. To give your agent access to the live internet, you can enable the web search tool.

OpenClaw officially recommends Brave Search. It uses an independent web index optimized for AI retrieval, ensuring your agent finds accurate information.

1. Open the OpenClaw CLI.
2. Run the following command to start the web configuration wizard:

    ```bash
    openclaw configure --section web
    ```
3. Configure the basic settings as follows:

    |Settings|Option|
    |:-------|:-----|
    |Where will the Gateway run|Local (this machine)|
    |Enable web_search (Brave Search)|Yes|
    |Brave Search API key| Your `BraveSearchAPIkey` |
    |Enable web_fetch (keyless HTTP <br>fetch) |Yes|

4. Finalize the configuration in Control UI.

    The CLI wizard sets up the API key, but you can customize specific tool parameters such as timeouts and limits in the Control UI.

    a. Return to the **Control UI** > **Config** > **Raw** tab. 

    b. Find the `tools` section and update as follows: 

    ```json
    "tools": {
        "web": {
        "search": {
            "enabled": true,
            "provider": "brave",
            "apiKey": "{Your-Brave-Search-API-Key}",
            "maxResults": 10,
            "timeoutSeconds": 30
        },
        "fetch": {
            "enabled": true,
            "timeoutSeconds": 30
        }
        }
    },
    ```

5. Now you can ask the agent in Discord to answer questions that require real-time internet data.

## FAQ

### I see the error "HTTP 503" or "event gap detected"

If you encounter these errors in Discord or the Control UI, check your Ollama and ensure it is running properly in the background.

## Resources

- [How do I create a server in Discord](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)