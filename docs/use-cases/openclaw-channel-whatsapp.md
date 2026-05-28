---
outline: [2, 3]
description: Learn how to integrate OpenClaw with WhatsApp.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, OpenClaw tutorial, OpenClaw learning, channel integration, WhatsApp integration
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-05-28"
---

# Integrate with WhatsApp

Connect your OpenClaw agent to WhatsApp to interact with it remotely.

Choose one of the following account setups to get started:
- **Use a standalone account (Recommended)**: Purchase a separate SIM card and register a dedicated WhatsApp account for your bot.
- **Use your personal account**: If you only need the bot for personal use, you can link your own WhatsApp account and chat with the bot by sending messages to yourself.

## Learning objectives

In this guide, you will learn how to:
- Link OpenClaw to WhatsApp using a standalone or personal account.
- Pair your personal WhatsApp number with a standalone bot account.
- Disconnect and fully remove the WhatsApp integration.

## Configure WhatsApp channel

Choose your preferred configuration method.

<Tabs>
<template #Standalone-account>

:::info Prerequisite
Before you begin, ensure you have registered a separate WhatsApp account for your bot and are logged into it on a mobile device.
:::

1. Open the OpenClaw CLI from the Launchpad.
2. Enter the following command to start the channel configuration wizard:

    ```bash
    openclaw config --section channels
    ```

3. Configure the settings as follows:

    - **Channel setup**: Select **Add or update channels**.
    - **Select a channel**: Select **WhatsApp (QR link)**.
    - **Install WhatsApp plugin**: Select **Download from ClawHub**.
    - **Link WhatsApp now (QR)**: Select **Yes**.

    A QR code is displayed in the terminal.

4. Open the bot's WhatsApp mobile app, go to **Linked Devices**, and then tap **Link a device** to scan the QR code.
5. Check the OpenClaw CLI for the message `Linked after restart; web session ready`. It indicates the binding is successful.
6. Continue to configure the following settings in OpenClaw CLI:

    - **WhatsApp phone setup**: Select **Separate phone just for OpenClaw**.
    - **WhatsApp DM policy**: Select **Pairing (recommended)**.
    - **WhatsApp allowFrom (optional pre-allowlist)**: Select **Unset allowFrom (default)**.
    - Select **Finished** to exit.

    OpenClaw will automatically restart to apply the configurations. Wait about 5 to 10 minutes.

7. Once restarted, use your personal WhatsApp account to send a message to the bot's number. The bot will reply with a pairing command and code.
8. Return to the OpenClaw CLI, paste the exact command provided in the chat (for example, `openclaw pairing approve whatsapp VUTNHSXS`), and then press **Enter**. 

    A success message will appear in the terminal indicating the pairing is approved.

9. Send another message to the bot on WhatsApp. You are now connected and can chat with your bot.
</template>
<template #Personal-account>

1. Open the OpenClaw CLI from the Launchpad.
2. Enter the following command to start the channel configuration wizard:

    ```bash
    openclaw config --section channels
    ```

3. Configure the settings as follows:

    - **Channel setup**: Select **Add or update channels**.
    - **Select a channel**: Select **WhatsApp (QR link)**.
    - **Install WhatsApp plugin**: Select **Download from ClawHub**.
    - **Link WhatsApp now (QR)**: Select **Yes**.

    A QR code is displayed in the terminal.

4. Open the bot's WhatsApp mobile app, go to **Linked Devices**, and then tap **Link a device** to scan the QR code.
5. Check in the OpenClaw CLI for the message `Linked after restart; web session ready`. It indicates the binding is successful.
6. Continue to configure the following settings in OpenClaw CLI:

    - **WhatsApp phone setup**: Select **This is my personal phone number**.
    - Enter your personal WhatsApp phone number when prompted.
    - Select **Finished** to exit.

7. OpenClaw will automatically restart to apply the configuration. Wait about 5 to 10 minutes.
8. Once restarted, open WhatsApp and send a message to yourself. The bot will reply to you within that same chat thread.

    :::tip Note on personal mode
    In this mode, the messages you send to the bot and the bot's replies will be mixed together in your self-chat.
    :::
</template>
</Tabs>

:::tip Manual restart
After completing the configuration, OpenClaw automatically restarts the gateway. WhatsApp integration requires this restart to fully come online. 

If you wait longer than 10 minutes and the bot is still not responding to your messages, try manually restarting the container:
1. Open Control Hub from the Launchpad.
2. Locate the `clawdbot` deployment, and then click **Restart** in the upper-right corner.
:::

## Disconnect WhatsApp

To fully disconnect WhatsApp from OpenClaw, you must remove the configuration from both the Control UI and the local system files.

1. Open the Control UI, and then select **Agents** from the left sidebar.
2. On the **Channels** tab, locate the WhatsApp configuration card, and then click **Logout**.
3. Open the OpenClaw CLI, and then run the following command:

    ```bash
    openclaw config --section channels
    ```

4. Select **Remove channel config**, select **WhatsApp**, and then select **Yes** to delete the configuration.
5. Select **Done** to exit the wizard.
6. Open the Olares Files, and then go to **Data** > **clawdbot** > **config** > **credentials**.
7. Delete the `whatsapp` folder, and delete any isolated WhatsApp credential files such as `whatsapp-pairing.json` and `whatsapp-default-allowFrom.json`.
8. Restart the OpenClaw container in Control Hub to apply the changes.

## Next steps

- [Optional: Enable web search](openclaw-web-access.md)
- [Manage skills and plugins](openclaw-skills.md)
