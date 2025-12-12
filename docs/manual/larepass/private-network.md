---
outline: [2, 3]
description: Learn how to securely access your Olares from anywhere. This guide explains public vs. private entrances, when to use LarePass VPN, how to enable VPN on mobile and desktop.
---

# Access Olares anywhere

This guide explains **how to reach your Olares from anywhere**. You will learn:

1) The access paths for public vs. private entrances.
2) How to enable LarePass VPN on your mobile and desktop.
3) Interpret connection status and know when to troubleshoot.

## How access works in Olares

In Olares, you access each app or service via its own URL (`https://app.olares-id.olares.com`, for example, `https://desktop.nicholas.olares.com/`). Depending on who should reach it, there are two entrance types.

### Public entrance

  * Accessible to anyone on the internet without authentication. For example, a public blog hosted on WordPress.
  * Traffic is securely routed from the internet to Olares via Cloudflare Tunnel or FRP.

### Private entrance

Application entrances intended only for you, such as Desktop, Vault, and the management console of WordPress. Depending on where you are, there are two scenarios when accessing private entrances:

- **Remote access** (Outside your local network)
  - **With LarePass VPN (Recommended):** Traffic is routed directly and securely through the VPN (Tailscale), no matter where you are.
  - **Without LarePass VPN:** Traffic is routed through the same internet tunnel as public access (Cloudflare/FRP).

 - **Local access** (On the same network)
  
    Use the local URL (`http://app.yourname.olares.local`) for a direct, local connection that bypasses the VPN and internet tunnels.

    :::tip For macOS users
    Chrome may fail to access local URLs if macOS blocks local network permissions.
    To enable access:
    1. Open Apple menu and go to **System Settings**.
    2. Go to **Privacy & Security** > **Local Network**.
    3. Find Google Chrome and Google Chrome Helper in the list and enable the toggles.
    ![Enable local network](/public/images/manual/larepass/mac-chrome-local-access.png#bordered){width=400}

    Restart Chrome and try accessing the local URL again.
    :::

    :::info For Windows users
    Currently, local access via `.local` domains is not supported on Windows.
    :::

:::tip Always enable VPN for remote access
For the best experience with private apps when you're away from your network, enable **LarePass VPN**. It keeps your connection to Olares encrypted, direct, and fast.
:::

## Enable VPN on LarePass

:::tip
For different LarePass download options, visit [the official page](https://www.olares.com/larepass).
:::

![VPN](/images/manual/larepass/vpn.jpg)

### On LarePass mobile client
1. Open LarePass, and go to **Settings**.
2. In the **My Olares** card, toggle on the VPN switch.

### On LarePass desktop client
1. Open LarePass, click on the avatar area in the top left corner of the main interface.
2. Toggle on the switch for **VPN connection** in the pop-up panel.

Devices with activated VPN will use the VPN connection to access Olares, whether through the LarePass client or a browser.

:::info
iOS or macOS versions of LarePass will require adding a VPN configuration file to the system when turning on the VPN. Follow the prompts to complete the setup.
:::

## Understand connection status
LarePass displays the connection status between your device and Olares, helping you understand or diagnose your current network connection.

![Connection status](/images/manual/larepass/connection-status.jpg)

| Status       | Description                                      |
|--------------|--------------------------------------------------|
| Internet     | Connected to Olares via the public internet      |
| Intranet     | Connected to Olares via the local network        |
| FRP          | Connected to Olares via FRP                      |
| DERP         | Connected to Olares via VPN using DERP relay     |
| P2P          | Connected to Olares via VPN using P2P connection |
| Offline mode | Currently offline, unable to connect to Olares   |

::: info
When accessing private entrances from an external environment through VPN, if the status shows "DERP", it indicates that the VPN cannot directly connect to Olares via P2P and must use Tailscale's relay servers. This status may affect connection quality. If you constantly encounter this situation, please contact Olares support for assistance.
:::

## Troubleshoot connection issues
If you encounter connection problems, LarePass will display diagnostic messages to help you resolve the issue. Here are some common scenarios and how to address them:

![Abnormal status](/images/manual/larepass/abnormal_state.png)

| Status message                                        | Possible cause and recommended actions                                                                                                                                                                                                                                                                                                                                            |
|-------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Network issue detected. Check local network settings. | **Local network issue** <br> 1. Wait for automatic reconnection. <br/>The system will detect network recovery <br/>and sync data.<br/> 2. Check your local network settings if <br/>the issue persists.                                                                                                                                                                           |
| VPN required to connect to Olares.                    | **VPN not enabled** <br> Click the notification banner and follow <br/>prompts to enable VPN connection.                                                                                                                                                                                                                                                                        |
| Need to log in to Olares again.                       | **Session expired or authentication issue** <br> Click the notification banner and follow<br/> prompts to log in.                                                                                                                                                                                                                                                                 |
| Need to reconnect to Olares.                          | **Connection interrupted or timed out** <br> Click the notification banner and follow<br/> prompts to log in. After re-login, Vault <br/>data will sync and merge with the server.                                                                                                                                                                                                |
| No active Olares found.                               | **Temporary network issue or Olares is restarting<br/> or shutting down** <br> Wait for automatic recovery. This <br/>usually resolves shortly. <br> **Olares instance no longer exists** <br> 1. Click the notification banner and follow<br/> prompts to reactivate Olares, enable offline <br/>mode or ignore notification. <br> 2. Contact Olares Admin if the issue persists. |
