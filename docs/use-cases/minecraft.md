---
outline: [2, 3]
description: Play Minecraft Java Edition with friends by hosting a server on Olares that supports LAN and VPN connections.
head:
  - - meta
    - name: keywords
      content: Olares, Minecraft, Minecraft Java, game server, self-hosted, Overlay gateway, VPN, Games
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-08-19"
---

# Play Minecraft Java Edition with friends on Olares

Minecraft on Olares runs the official vanilla Minecraft Java Edition dedicated server. You manage it through the console terminal, while players connect using the official Minecraft Java Edition client. Invite friends to play over your local network through the overlay gateway, or let them connect remotely using LarePass VPN.

## Learning objectives

In this guide, you will learn how to:

- Install the Minecraft app from Market.
- Enable overlay gateway for local network play.
- Connect to the server over the local network or VPN.

## Prerequisites

- **Olares OS**: Olares version 1.12.6 or later.
- **Hardware and network**: A native Linux host with a wired Ethernet connection for the Olares device. Overlay gateway does not work on Wi-Fi or WSL.
- **Permissions**: A super admin must toggle on the system-level overlay gateway service. After the service is on, an admin or member can enable overlay gateway for Minecraft.
- **Client requirements**: Minecraft Java Edition installed on each player's computer. Bedrock, console, and mobile editions cannot connect. The client version must match the **App version** of Minecraft shown on the Market page.

## Install Minecraft

1. Open Market and search for "Minecraft".
2. Click **Get**, then **Install**, and wait for the installation to complete.
   ![Minecraft in Market](/images/manual/use-cases/minecraft.png#bordered)

3. Open the app and wait for the server status to become running. The first start downloads server resources, which may take a few minutes.

## Enable overlay gateway for Minecraft

Overlay gateway gives Minecraft a dedicated local IP address so players on the same network can connect directly.

1. Open Olares Settings and go to **Network** > **Overlay gateway**.
2. Ensure **Enable overlay gateway** is toggled on. This is the system-level service switch. If it is off, a super admin must turn it on.
3. In the **Applications** list, find **Minecraft**, confirm its status is **Running**, then enable overlay gateway for the app.
4. To the right of **Minecraft Java**, copy the address shown. For example, `192.168.50.219:25565`.

:::info Local IP address is dynamic
The local IP address is assigned by the Overlay gateway and may change when the app restarts or the network changes. Always use the address currently shown on this page.
:::

## Connect from the same local network

Players on the same local network as the Olares device can connect through the overlay address.

1. Copy the overlay address from **Settings** > **Network** > **Overlay gateway**. For example `192.168.50.219:25565`.
2. Open Minecraft Java Edition and click **Multiplayer**.

   ![Minecraft multiplayer menu](/images/manual/use-cases/minecraft-multiplayer-menu.png#bordered)

3. Click **Add Server**.
4. Enter the server information, then click **Done**:

   - **Server Name**: Enter a name for easy identification.
   - **Server Address**: Enter the overlay gateway address copied earlier.

   ![Minecraft add server](/images/manual/use-cases/minecraft-add-server.png#bordered)

5. Select the server, then click **Join Server**.

   ![Minecraft join server](/images/manual/use-cases/minecraft-join-server.png#bordered)

## Connect over VPN

This method works when the player is not on the same local network as the Olares device.

:::tip Keep Overlay enabled
You can keep Minecraft's overlay gateway enabled. It does not conflict with VPN connections.
:::

1. Ensure [LarePass VPN](../manual/get-started/local-access.md#using-larepass-vpn) is enabled.
2. Open Minecraft Java Edition and click **Multiplayer**.
3. Click **Add Server**.
4. In **Server Address**, enter in the following format. Replace `<local-name>` with the local name in your Olares ID (the text before `@`).

   For example, if your Olares ID is `olarestest001@olares.com`, the server address is `olarestest001.olares.com:25565`.

   ```text
   <local-name>.olares.com:25565
   ```

5. Save the server and click **Join Server**.

## Manage the server

The Minecraft app does not have a web management interface. To view logs or run server commands, open it from Launchpad and use the built-in console terminal.

## FAQs

### Why does my local IP address differ from the screenshots?

The overlay gateway assigns the local IP address dynamically. Always use the address shown in **Settings** > **Network** > **Overlay gateway** > **Minecraft** at the time you connect.

### Why does the client report a version mismatch?

Update the Minecraft Java Edition client to the same version shown on the Market page for the server. The server and client versions must match.

### Why can't I connect with the VPN address?

Check the following:

- You are using Minecraft Java Edition.
- The server address format is `<local-name>.olares.com:25565`, where `<local-name>` is the part of your Olares ID before `@`.
- Your LarePass VPN connection is enabled.

### What happens when I upgrade the app?

Upgrading the app restarts the server. Any players currently online will be disconnected.

## Learn more

- [Manage overlay gateway for applications](/manual/olares/settings/overlay-gateway.md): Configure LAN access for supported apps.
