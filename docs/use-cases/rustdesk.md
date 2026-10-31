---
outline: [2, 3]
description: Run RustDesk Server on Olares to access your home computer from a laptop through LarePass VPN, or connect over the same local network with Overlay.
head:
  - - meta
    - name: keywords
      content: Olares, RustDesk Server, remote desktop, self-hosted, ID server, relay, LarePass VPN, Overlay gateway
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-10-31"
---

# Access your home computer with RustDesk

Use RustDesk to access your home computer from a laptop through LarePass VPN. RustDesk Server runs on Olares to help the two computers connect and relay traffic when a direct connection is unavailable.

## Learning objectives

In this guide, you will learn how to:

- Install RustDesk Server and find your server addresses and public key.
- Configure both computers to use your server through LarePass VPN or local Overlay.
- Connect to your home computer and control it remotely.

## Prerequisites

- Olares 1.12.7 or later.
- The [RustDesk client](https://rustdesk.com/download) installed on both computers.
- LarePass installed on both computers and logged in to the same Olares device. See [LarePass private network access](../manual/larepass/private-network.md).

## Install RustDesk Server

1. Open Market and search for "RustDesk Server".
2. Click **Get**, then **Install**, and wait for installation to complete.

## Find your server settings

1. Open RustDesk Server from Launchpad. The app displays a connection guide with your server addresses and public key. You will use these values in the RustDesk clients on your computers.
2. Select **VPN connection**.

   ![VPN server addresses and public key in RustDesk Server](/images/manual/use-cases/rustdesk-vpn-settings.png#bordered)

Copy the values from your own app page when configuring the clients below.

To change the page language, click the globe icon in the upper-right corner. The page remembers your selection.

## Configure both computers for LarePass VPN

Use VPN on both computers for this walkthrough. For two computers on the same network as Olares, you can instead [use local Overlay](#connect-through-local-overlay).

Use the same connection mode on both computers. If one RustDesk client uses the VPN domain address and the other uses an Overlay IP, the connection can fail when either computer cannot reach the relay address used for the session.

1. On both computers, open LarePass and enable **VPN connection** from the avatar menu. Keep it enabled during the remote desktop session. See [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) for platform-specific setup.
2. On each computer, open RustDesk and go to **Settings** > **Network** > **ID/Relay server**. Unlock the settings if prompted.
3. Fill in the following fields, then save the settings on both computers.

   | Client field | Value |
   | --- | --- |
   | **ID Server** | Copy the full address from the app page, including the `:31116` port |
   | **Relay Server** | Copy the full address from the app page, including the `:31117` port |
   | **API Server** | Leave empty. RustDesk Server OSS does not provide the Pro management API |
   | **Key** | Copy the public key from the app page |

   The Key identifies your RustDesk server. The access password shown in the home computer's RustDesk client authorizes remote control of that computer.

## Start a remote desktop session

:::warning Remote control of a Mac
On the Mac you want to control, grant RustDesk Screen Recording and Accessibility permissions before connecting. These permissions let RustDesk share the screen and control the keyboard and mouse. See the [RustDesk macOS setup guide](https://rustdesk.com/docs/en/client/mac/).
:::

1. Keep RustDesk running on the home computer and make sure its service is enabled.
2. On the laptop, enter the home computer's RustDesk ID and start the connection.
3. Approve the request on the home computer, or enter the access password shown or configured in its RustDesk client.
4. When the home computer's desktop appears, use the laptop's keyboard and mouse to control it. End the session in RustDesk when finished.

## Connect through local Overlay

If both computers are on the same network as Olares, you can use Overlay instead of LarePass VPN. Overlay gives RustDesk Server its own local IP address.

:::warning Overlay requirements
Olares must run on a native Linux host with wired Ethernet. The Super admin must enable the system-level Overlay gateway before you can enable it for RustDesk Server. See [Manage overlay gateway](../manual/olares/settings/overlay-gateway.md) for availability and permissions.
:::

1. In Olares, open **Settings** > **Network** > **Overlay gateway**. If the system gateway is off, enable it as the Super admin or ask your Super admin to do so.
2. Under **Applications**, enable RustDesk Server and confirm the restart. Wait for the app to return to **Running**, then copy its assigned local IP from this settings page.
3. Open RustDesk Server from Launchpad and select **Local Overlay**. When configuring the clients, replace the `Overlay_IP` placeholder with the IP you copied from Settings.

   ![Local Overlay address format in RustDesk Server](/images/manual/use-cases/rustdesk-overlay-settings.png#bordered)

4. On both computers, open RustDesk **Settings** > **Network** > **ID/Relay server** and fill in the following fields.

   | Client field | Value |
   | --- | --- |
   | **ID Server** | The assigned Overlay IP followed by `:21116` |
   | **Relay Server** | The same Overlay IP followed by `:21117` |
   | **API Server** | Leave empty |
   | **Key** | Copy the public key from the RustDesk Server page |

5. Save the settings on both computers, then [start a remote desktop session](#start-a-remote-desktop-session).

If the app's Overlay address changes, copy the current address from Settings and update both clients.

## FAQs

### Why does RustDesk fail to connect to the relay server?

During connection setup, RustDesk can pass the home computer's relay address to the laptop. Both computers must be able to reach that address. If the laptop cannot reach it, RustDesk can report "Failed to connect to relay server: Please try later", even when both computers can reach the ID server.

Configure both computers for the same connection mode, then check the following:

1. Check **Relay Server** on both computers. For VPN, use the full domain address from the app page and keep LarePass VPN enabled. For Overlay, use the current local IP from Settings with port `21117`.
2. In the RustDesk network settings, keep **Use WebSocket**, **Allow insecure TLS fallback**, and **Disable UDP** off for this configuration.
3. Save any changes and try connecting again.

### What does setting the web entrance to Public change?

Public lets you open the connection guide without signing in. Internet access to RustDesk's TCP/UDP ports requires separate network configuration. For the connections in this guide, use the app's domain addresses through LarePass VPN or the Overlay IP on the same local network.

## Learn more

- [LarePass private network access](../manual/larepass/private-network.md): Connect computers to your Olares private network.
- [Manage overlay gateway](../manual/olares/settings/overlay-gateway.md): Enable an app's local address and check feature availability.
- [RustDesk Server OSS documentation](https://rustdesk.com/docs/en/self-host/rustdesk-server-oss/): Learn about the ID and relay services.
