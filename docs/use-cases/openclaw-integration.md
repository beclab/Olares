---
outline: [2, 3]
description: Learn how to integrate OpenClaw with Discord by creating a bot, configuring channels, and authorizing your account.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, channel integration, Discord integration
app_version: "1.0.0"
doc_version: "1.1"
doc_updated: "2026-04-24"
---

# Integrate with Discord

To chat with your agent remotely, connect it to a Discord bot.

## Prerequisites

- Discord account: Required to create the bot application.
- Discord server: A server where you have permissions to add bots.

## Step 1: Create a Discord bot

1. Log in to the [Discord Developer Portal](https://discord.com/developers/applications) with your Discord account.
2. Click **New Application**.
    ![New application in Discord developer portal](/images/manual/use-cases/new-app.png#bordered){width=90%}

3. Enter a name for the new app, agree to terms, and then click **Create**.

    ![Create an application window](/images/manual/use-cases/create-app.png#bordered){width=40%}

4. From the left sidebar, select **Bot**.
5. Scroll down to the **Privileged Gateway Intents** section and enable the following settings:

    - Presence Intent
    - Server Members Intent
    - Message Content Intent
6. Click **Save Changes**.
7. Scroll up to the **Token** section, click **Reset Token**, and then copy the generated token for your Discord bot. You need the token for channel configuration later in Control UI.

    ![Reset token](/images/manual/use-cases/reset-token.png#bordered)

## Step 2: Invite the bot to server

1. From the left sidebar, select **OAuth2**, and then find the **OAuth2 URL Generator** section:

    a. In **Scopes**, select **Bot** and **applications.commands**.

    ![OAuth2 URL Generator](/images/manual/use-cases/oauth21.png#bordered)

    b. Scroll down to **Bot Permissions**, set as the following image. You can modify the settings later.

    ![Bot permissions](/images/manual/use-cases/bot-permissions1.png#bordered)

2. Copy the **Generated URL** at the bottom.
3. Paste the URL into a new browser tab, select your Discord server from **Add to server**, click **Continue**, and then click **Authorize**. 

    The bot is authorized and added to your server.

    ![Bot added to server](/images/manual/use-cases/bot-added.png#bordered)

## Step 3: Configure the channel

Connect OpenClaw to your Discord bot by adding the channel settings to the configuration file.

1. Open the Files app, and then go to **Data** > **clawdbot** > **config**.
2. Double-click the `openclaw.json` file to open it.
3. Click <i class="material-symbols-outlined">edit_square</i> in the upper-right corner to enter the edit mode.
4. Add the following `channels` section to the configuration file. 

    This configuration enables Discord Direct Messages (DMs) and sets the DM policy to `pairing` for security.

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
5. Replace `{YOUR_BOT_TOKEN}` with your Discord bot token.
6. Click <i class="material-symbols-outlined">save</i> in the upper-right corner to save the changes.
7. Return to the Control UI, and then select **Channels** from the left sidebar. On the Discord card, a **Probe ok** status indicates the connection is successful.   

<!--
Connect OpenClaw to your Discord bot by adding its configuration in the Control UI.

:::info About channel configuration
This tutorial provides the basic setup to get your bot running in Discord quickly. For more detailed configurations, see the official [OpenClaw documentation](https://docs.openclaw.ai/channels).
:::

1. Open the Control UI, select **Config** from the left sidebar, and then switch to the **Raw** tab.
2. Click <i class="material-symbols-outlined">visibility_off</i> to reveal the configuration fields.

    ![Reveal configuration blocks](/images/manual/use-cases/click-hide-icon.png#bordered)
3. Add the following `channels` section to the configuration file. 

    This configuration enables Discord Direct Messages (DMs) and sets the DM policy to pairing for security.

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

    ![Discord channel added](/images/manual/use-cases/channels1.png#bordered)

3. Replace `{YOUR_BOT_TOKEN}` with your Discord bot token.
4. Click **Save**.
5. From the left sidebar, select **Channels**. On the Discord card, **Probe ok** indicates successful connection.
-->

## Step 4: Authorize your account

For security, the bot does not talk to unauthorized users. You must pair your Discord account with the bot.

1. Open Discord and send a Direct Message to your new bot. The bot will reply with an error message containing a Pairing Code.
2. Open the OpenClaw CLI and enter the following command:

    ```bash
    openclaw pairing approve discord {Your-Pairing-Code}
    ```

3. Once approved, you can start chatting with your agent in Discord.

## Next steps

- [Optional: Enable web search](openclaw-web-access.md)
- [Manage skills and plugins](openclaw-skills.md)