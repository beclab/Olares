---
outline: [2, 3]
description: Host RustDesk ID and relay services on Olares, configure two computers through LarePass VPN or local Overlay, and start a remote desktop session.
head:
  - - meta
    - name: keywords
      content: Olares, RustDesk Server, remote desktop, self-hosted, ID server, relay, LarePass VPN, Overlay gateway
app_version: "0.0.17"
doc_version: "1.0"
doc_updated: "2026-10-28"
---

# Connect to your computers with RustDesk Server on Olares

RustDesk Server lets RustDesk clients find each other and relay desktop traffic when a direct connection is unavailable. Run the server on Olares and install a RustDesk client on each computer. For example, connect a laptop to a home computer through LarePass VPN. The Olares app opens a connection guide, while the remote desktop session runs in the RustDesk client.

## Learning objectives

In this guide, you will learn how to:

- Install RustDesk Server and find the client connection settings.
- Configure both computers through LarePass VPN or local Overlay.
- Start a remote desktop session and recognize a relay-address mismatch.

## Prerequisites

- Olares 1.12.7 or later.
- The [RustDesk client](https://rustdesk.com/download) installed on both computers.
- LarePass installed and logged in to the Olares device on both computers for the VPN example. See [LarePass private network access](../manual/larepass/private-network.md).

## Install RustDesk Server

1. Open Market and search for "RustDesk Server".
2. Click **Get**, then **Install**, and wait for installation to complete.
3. Open RustDesk Server from Launchpad. The page shows the server addresses and a generated public key.

The globe icon in the upper-right corner changes the page language. The page remembers your selection.

## Configure both computers

Use the same connection mode on both computers for the first connection. The server addresses differ between VPN and local Overlay, but the Key is the same.

:::warning Both computers must reach the negotiated relay address
RustDesk can pass the remote computer’s configured relay address to the controlling computer. If one uses a VPN-only domain and the other uses an Overlay IP that the VPN computer cannot reach, the connection can fail even though both computers reach the ID server. Using the same connection mode on both computers avoids this address mismatch.
:::

### Connect through LarePass VPN

1. On both computers, open LarePass and enable **VPN connection** from the avatar menu. See [Enable LarePass VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass) for platform-specific instructions.
2. In the RustDesk Server page, select **VPN connection**.
3. On each computer, open RustDesk and go to **Settings** > **Network** > **ID/Relay server**. Unlock the settings if the client asks.
4. Fill in the four fields below, then confirm the changes.

   | Client field | Value |
   | --- | --- |
   | **ID Server** | Copy the full address from the page, including the `:31116` port |
   | **Relay Server** | Copy the full address from the page, including the `:31117` port |
   | **API Server** | Leave empty |
   | **Key** | Copy the public key from the page |

![English VPN connection settings in RustDesk Server](/images/manual/use-cases/rustdesk-vpn-settings.png#bordered)

Copy the addresses from your own instance. Do not use the example domain or Key shown in the screenshot. Keep LarePass VPN enabled during the session.

### Connect through local Overlay

Use this alternative when both computers are on the same network as Olares. Overlay requires a native Linux Olares host with wired Ethernet. See [Manage overlay gateway](../manual/olares/settings/overlay-gateway.md) for the availability and permission requirements.

1. In Olares, open **Settings** > **Network** > **Overlay gateway**. Enable the system gateway if it is not already enabled.
2. Under **Applications**, enable RustDesk Server and confirm the restart. Wait for the app to return to **Running**, then copy its assigned local IP from this settings page.
3. Open the RustDesk Server page and select **Local Overlay**. `Overlay_IP` is a placeholder, not an address detected by this page.
4. On both computers, open RustDesk **Settings** > **Network** > **ID/Relay server** and fill in the following fields.

   | Client field | Value |
   | --- | --- |
   | **ID Server** | The assigned Overlay IP followed by `:21116` |
   | **Relay Server** | The same Overlay IP followed by `:21117` |
   | **API Server** | Leave empty |
   | **Key** | Copy the public key from the RustDesk Server page |

![English local Overlay instructions in RustDesk Server](/images/manual/use-cases/rustdesk-overlay-settings.png#bordered)

Confirm the settings on both computers. If the app’s Overlay address changes, copy the current address from Settings and update both clients.

## Start a remote desktop session

1. Keep RustDesk running on the home computer and make sure its service is enabled.
2. On the laptop, enter the home computer’s RustDesk ID and start the connection.
3. Approve the request on the home computer, or enter the access password shown or configured in its RustDesk client.
4. When the remote desktop appears, control the home computer from the laptop. End the session in RustDesk when finished.

Leave API Server empty because RustDesk Server OSS does not provide the Pro management API. This does not prevent remote desktop connections.

The server’s public Key identifies the server. It is not the home computer’s access password. A relayed session does not require the two computers to connect directly to each other, but both must reach the relay address negotiated for that session.

## Resolve a relay connection error

If RustDesk reports "Failed to connect to relay server: Please try later", check the selected mode and relay address on both computers first.

For the VPN example, keep LarePass VPN enabled on both computers and use the domain addresses copied from the app page. For local Overlay, use the current Overlay IP on both computers. Keep **Use WebSocket**, **Allow insecure TLS fallback**, and **Disable UDP** off for this configuration.

Making the Olares web entrance public only removes the login requirement from the connection guide. It does not publish the RustDesk TCP/UDP ports or make a local Overlay IP reachable through a VPN. Public internet port forwarding is a separate network configuration.

## Learn more

- [LarePass private network access](../manual/larepass/private-network.md): Connect computers to your Olares private network.
- [Manage overlay gateway](../manual/olares/settings/overlay-gateway.md): Enable an app’s local address and check feature availability.
- [RustDesk Server OSS documentation](https://rustdesk.com/docs/en/self-host/rustdesk-server-oss/): Learn about the ID and relay services.
