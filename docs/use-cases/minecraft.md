---
outline: [2, 3]
description: Host a Minecraft Java Edition server on Olares, choose a game version and Vanilla, Forge, or Fabric, and configure private servers and independent instances for friends.
head:
  - - meta
    - name: keywords
      content: Olares, Minecraft, Minecraft Java, game server, Forge, Fabric, mods, offline mode, clone, Overlay gateway, VPN, Games
app_version: "0.1.22"
doc_version: "2.0"
doc_updated: "2026-10-10"
---

# Play Minecraft Java Edition with friends on Olares

Minecraft on Olares hosts a Minecraft Java Edition dedicated server with Vanilla, Forge, and Fabric options. You manage it through the console terminal, while players connect using a compatible Java Edition client. Invite friends to play over your local network through the overlay gateway, or let them connect remotely using LarePass VPN.

## Learning objectives

In this guide, you will learn how to:

- Choose a Minecraft version and server type during installation.
- Connect to the server over the local network or LarePass VPN.
- Install mods and configure account verification, cheat commands, and the player limit.
- Create independent servers with Clone.

## Prerequisites

- Olares 1.12.6 or later.
- Minecraft Java Edition installed on each player's computer. Bedrock, console, and mobile editions cannot connect directly.

:::info Match the client to the server
The client must use the same Minecraft game version as the server. For Forge or Fabric, each player also needs the loader and client mods required by the modpack.
:::

## Install Minecraft

1. Open Market and search for "Minecraft".
   ![Minecraft in Market](/images/manual/use-cases/minecraft.png#bordered)

2. Click **Get**, then **Install**. Select the Minecraft **VERSION** you want to play and keep **TYPE** set to `VANILLA` for the original game. For modded play, expand the version and server type settings below, and wait for installation to complete.

:::: details Choose the game version and server type

Select **VERSION** during installation or when creating a Clone. Olares selects the server’s Java runtime automatically.

| VERSION | Server Java runtime |
|:---|:---|
| `1.12.2` | Java 8 |
| `1.16.5` | Java 8 |
| `1.18.2` | Java 17 |
| `1.19.2` | Java 17 |
| `1.20.1` | Java 17 |
| `1.21.1` | Java 21 |
| `26.2` | Java 25 |
| `LATEST` | Java 25 |

Choose a fixed version supported by your loader and mods. Minecraft `1.12.2` does not support Fabric. `LATEST` selects the latest release and might update the game when the server restarts.

Choose the server type and its loader settings:

| Field | Value |
|:---|:---|
| **TYPE** | `VANILLA` (default), `FORGE`, or `FABRIC` |
| **FORGE_VERSION** | Forge only. Use `recommended` (default), `latest`, or the version required by your modpack. |
| **FABRIC_LOADER_VERSION** | Fabric only. Use `latest` (default) or the version required by your modpack. |

These settings and **VERSION** cannot be changed after installation. Use [Clone](../manual/olares/market/clone-apps.md) to create a server with a different game version or loader.

:::warning Protect existing worlds
Back up the instance's entire data directory before migrating a world or reinstalling. Reinstalling with another version can reuse the existing data directory. Upgrading can convert world data, while downgrading can prevent startup or lose chunks, items, and entities. Changing loaders or removing mods can also make a world incompatible. Never start an older server against a world saved by a newer version.
:::
::::

The first start downloads the server and loader resources, which might take a few minutes depending on your network. For Vanilla play, continue to [Connect to the server](#connect-to-the-server). For modded play, [install your mods](#install-forge-or-fabric-mods) before joining.

## Configure Minecraft

Use the optional settings below to customize your server, or go straight to [Connect to the server](#connect-to-the-server).

### Optional: Install Forge or Fabric mods {#install-forge-or-fabric-mods}

For example, choose `VERSION=1.20.1` and `TYPE=FORGE` or `TYPE=FABRIC`, then set the loader version required by your mods.

1. Stop the target Minecraft instance from Market or Settings.
2. Open Files and locate `Data/<app-name>/data`. For the primary instance, this is `Data/minecraft/data`. For a Clone, use its app name from the Market details URL, as described in [Find the instance’s data directory](#find-the-instances-data-directory).
3. Create a `mods` folder if it does not exist, then upload the server-side mod `.jar` files and their required dependencies into it. Use mods for the selected Minecraft version and loader. Install Fabric API here when a mod requires it. Keep client-only mods on the client.
4. Add any configuration files required by the modpack to the corresponding folders under the same `data` directory. Uploading a modpack archive alone does not install it.
5. Resume the instance and check the console logs for missing dependencies or incompatible mods.
6. Connect using the matching game version, compatible loader, and the mods required on the client. Follow each mod’s requirements for the client.

### Optional: Disable account verification for a private server {#disable-account-verification-for-a-private-server}

Account verification is enabled by default. For an offline-mode private server with trusted players:

1. Go to **Settings** > **Applications** and select the target Minecraft instance.
2. Under **Environment variables**, click **Manage environment variables**.
3. Edit **ONLINE_MODE**, select `false` (**Disabled (trusted players only)**), and click **Confirm**.
4. Click **Apply** and wait for the server to restart.

This disables Minecraft account verification and secure profile enforcement. Players still need a reachable server address and a compatible client.

:::warning Trusted players only
Offline mode does not verify player identities, so another player can use a name belonging to someone else. Restrict access to trusted players. Switching between online and offline mode changes player UUIDs and can affect inventories and permissions. Back up the data first.
:::

### Optional: Allow cheat commands {#allow-cheat-commands}

Cheat commands are disabled by default. To enable them:

1. Go to **Settings** > **Applications** and select the Minecraft instance.
2. Under **Environment variables**, click **Manage environment variables**.
3. Set **ALLOW_CHEATS** to `true` and click **Confirm**.
4. Click **Apply** and wait for the server to restart.

When enabled, every player who joins receives permission to use commands such as `/give` and `/gamemode`, and command blocks are enabled. This grants level 2 operator permissions. Use the console for server-stop commands. Use this setting only with trusted players.

To disable cheat access, set **ALLOW_CHEATS** to `false` and apply the change. This clears the player operator list, including manually added operators, and restarts the server.

### Optional: Set the maximum player count {#set-the-maximum-player-count}

The default limit is 8 players online at once. To change it:

1. Go to **Settings** > **Applications** and select the Minecraft instance.
2. Under **Environment variables**, click **Manage environment variables**.
3. Set **MAX_PLAYERS** to an integer from `1` to `100` and click **Confirm**.
4. Click **Apply** and wait for the server to restart.

Higher limits require more resources, especially with mods.

## Connect to the server

Wait for the server to finish starting before joining. Open the instance from Launchpad to view its console logs.

In the Minecraft launcher, open **Installations** and select the same game version as the server. Choose a connection method below, then add the server in Minecraft.

<tabs>
<template #Local-network>

Overlay gateway gives the server a dedicated local IP address. This method requires Olares to run on a native Linux host with wired Ethernet. Wi-Fi and WSL do not support overlay gateway.

1. Open **Settings** > **Network** > **Overlay gateway**.
2. Turn on **Enable overlay gateway**. A super admin must enable this system-level service. Admins and members can then enable it for their own apps.
3. Under **Applications**, enable overlay gateway for the Minecraft instance you want to join. Click **Confirm** and wait for the instance to restart and return to **Running**.
4. Copy the address next to **Minecraft Java**, including the port. For example, `192.168.1.100:25565`.

Use the address currently shown on this page. The local IP can change after an app restart or a network change.

</template>
<template #LarePass-VPN>

Use LarePass VPN to connect from another network. You can keep overlay gateway enabled while using VPN.

1. Enable [LarePass VPN](../manual/get-started/local-access.md#using-larepass-vpn) on the computer running Minecraft.

   ![Enable LarePass VPN on your computer](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. Open **Settings** > **Applications**, select the Minecraft instance, and open **Export ports**. Note its **Exported port** for Minecraft. Use this value instead of the internal **Port** value.
3. Use the Olares domain and the exported port as the server address. For example, for the Olares ID `alex@olares.com`, use `alex.olares.com:<exported-port>`. Replace `<exported-port>` with the value from the previous step.

Each instance can have a different exported port. Use the port shown for the instance you want to join.

</template>
</tabs>

Once you have the address:

1. Open Minecraft Java Edition and click **Multiplayer**.

   ![Minecraft multiplayer menu](/images/manual/use-cases/minecraft-multiplayer-menu.png#bordered)

2. Click **Add Server**.
3. In **Server Name**, enter a name for the server.
4. In **Server Address**, enter the address for your connection method, then click **Done**.

   ![Minecraft add server](/images/manual/use-cases/minecraft-add-server.png#bordered)

5. Select the server, then click **Join Server**.

   ![Minecraft join server](/images/manual/use-cases/minecraft-join-server.png#bordered)

## Create another server with Clone

[Clone](../manual/olares/market/clone-apps.md) lets you run separate versions or mod setups on the same Olares device. It creates a new instance with its own configuration and data directory. It does not copy the original world's data or mods.

1. Install the primary Minecraft instance first, then open **My Olares** in Market.
2. Find Minecraft, click the drop-down arrow next to **Open**, and select **Clone**.
3. Enter a unique **New app title** and **Desktop shortcut name**, then click **Confirm**.
4. In **Configure Environment Variables**, select **VERSION**, **TYPE**, and the loader settings for the new server, then click **Confirm** and wait for installation.
5. Manage the new instance separately. Upload its mods to `Data/<app-name>/data/mods` and enable overlay gateway for that instance when using LAN access.

Use the connection details for the instance you want to join. Its overlay gateway address uses port `25565`, while its exported port for VPN access is assigned separately. Running several servers increases memory and CPU use.

### Find the instance’s data directory {#find-the-instances-data-directory}

1. Open Market and go to **My Olares**.
2. Select the Minecraft instance to open its details.
3. In the page URL, find the app name immediately after the source name. For example, in `/app/<source>/minecraftabc123`, the app name is `minecraftabc123`.
4. Open Files and go to `Data/<app-name>/data`. For this example, the directory is `Data/minecraftabc123/data`.

The folder uses the app name from the URL, even if the instance has a different display title.

### Copy an existing world

To continue an existing world in another instance, start with matching game, loader, and mod versions.

1. Stop both instances and back up their entire data directories.
2. In Files, replace the new instance’s `data` contents with a copy of the original instance’s `data` contents, including hidden files.
3. Review the new instance’s environment variables. These settings are separate from the copied files. Keep **ONLINE_MODE** the same to preserve player UUIDs and inventory associations. Review the copied operator, whitelist, and ban lists.
4. Resume the new instance and join using its own connection details.

:::warning Keep a backup
The instances have separate data directories, so progress is not synchronized. Keep a backup before changing the game version, loader, or mods. To roll back an upgrade, restore the earlier backup. Do not copy an upgraded world back to an older server.
:::

## Manage the server

To view logs or run server commands, open the Minecraft instance from Launchpad and use its console terminal.

## FAQs

### Why does my local IP address differ from the screenshots?

The overlay gateway assigns the local IP address dynamically. Always use the address shown in **Settings** > **Network** > **Overlay gateway** > **Minecraft** at the time you connect.

### Why does the client report a version mismatch?

Select the same Minecraft game version as the instance's **VERSION**. The Market app package version, such as `0.1.22`, is not a game version. For Forge or Fabric, also check the loader and client mod requirements.

### Why does the client say chat messages cannot be verified?

In offline mode, player identities and secure profiles are not verified. The client can warn that chat messages cannot be verified. This warning does not mean the network connection failed.

### Why can't I connect with the VPN address?

Check the following:

- You are using Minecraft Java Edition.
- The address uses your Olares domain and the instance’s **Exported port** from **Settings** > **Applications** > the target instance > **Export ports**.
- The server is ready, and your client matches its game version and mod requirements.
- Your LarePass VPN connection is enabled.

### What happens when I upgrade the app?

Upgrading the app restarts the server. Any players currently online will be disconnected.

## Learn more

- [Manage overlay gateway for applications](../manual/olares/settings/overlay-gateway.md): Configure LAN access for supported apps.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change settings and apply them with a restart.
- [Clone applications](../manual/olares/market/clone-apps.md): Create and manage independent app instances.
