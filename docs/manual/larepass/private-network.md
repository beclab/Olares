---
outline: [2, 3]
description: Learn how to securely access your Olares from anywhere.
---
# Access Olares services remotely via LarePass VPN
Your Olares device hosts critical applications intended for personal or internal use, such as Vault and Ollama. To ensure security, these applications are accessed via [private or internal entrances](../../developer/concepts/network.md#private-entrance).

To ensure the best connection to these apps, it's recommended to enable LarePass VPN. Once enabled, LarePass uses Tailscale to establish a secure network and automatically selects the fastest route based on location:

- **At home**: The app connects directly via the local network for maximum speed.
- **Remote**: The app creates a direct, encrypted P2P tunnel to the device.

If the VPN is disabled, traffic routes through standard public internet tunnels using Cloudflare or FRP.

## Enable VPN on LarePass
:::info iOS and macOS setup
On iOS or macOS, you may be prompted to add a VPN Configuration to your system settings the first time you enable the feature. Allow this to complete the setup.
:::

<tabs>
<template #On-LarePass-mobile-client>

1. Open the LarePass app, and go to **Settings**.
2. In the **My Olares** card, toggle on the VPN switch.

   ![Enable LarePass VPN on mobile](/images/manual/get-started/larepass-vpn-mobile.png#bordered)
</template>
<template #On-LarePass-desktop-client>

1. Open the LarePass app, and click your avatar in the top-left corner to open the user menu.
2. Toggle on the switch for **VPN connection**.

   ![Enable LarePass VPN on desktop](/images/manual/get-started/larepass-vpn-desktop.png#bordered)
</template>
</tabs>

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
