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

:::info App version covered
This guide covers Minecraft app package 0.1.22. The package version is separate from the Minecraft game version selected during installation. These features are available in the test market; if your Market source still offers 0.1.14, wait for the updated package before following the configuration and Clone steps below.
:::

## Learning objectives

In this guide, you will learn how to:

- Choose a Minecraft version and server type during installation.
- Install mods and configure account verification, cheat commands, and the player limit.
- Create independent servers with Clone.
- Enable overlay gateway for local network play.
- Connect to the server over the local network or VPN.

## Prerequisites

- **Olares OS**: Olares version 1.12.6 or later.
- **Hardware and network**: A native Linux host with a wired Ethernet connection for the Olares device. Overlay gateway does not work on Wi-Fi or WSL.
- **Permissions**: A super admin must toggle on the system-level overlay gateway service. After the service is on, an admin or member can enable overlay gateway for Minecraft.
- **Client requirements**: Minecraft Java Edition installed on each player's computer. Bedrock, console, and mobile editions cannot connect directly. The client version must match the Minecraft **VERSION** selected during installation, rather than the app package version shown on the Market page.

## Install Minecraft

1. Open Market and search for "Minecraft".
   ![Minecraft in Market](/images/manual/use-cases/minecraft.png#bordered)

2. Click **Get**, then **Install**. In the environment variables dialog, select **VERSION** and **TYPE** using the reference below, configure any loader settings required by your mods, and wait for installation to complete.

The first start downloads the server and loader resources, which might take a few minutes depending on your network. Wait for the server to be ready before joining. Open Minecraft from Launchpad to view startup logs in the console.

## Configure Minecraft

### Choose the game version and server type

Select **VERSION** during installation or when creating a Clone. Olares selects the Java runtime automatically; players do not need to configure Java on the server.

| VERSION | Server Java runtime | Typical use |
|:---|:---|:---|
| `1.12.2` | Java 8 | Older Forge mods and modpacks. Fabric is not supported. |
| `1.16.5` | Java 8 | Forge or Fabric modpacks built for 1.16.5. |
| `1.18.2` | Java 17 | Mods and modpacks targeting 1.18.2. |
| `1.19.2` | Java 17 | Mods and modpacks targeting 1.19.2. |
| `1.20.1` | Java 17 | Mods and modpacks targeting 1.20.1. |
| `1.21.1` | Java 21 | Mods and modpacks targeting 1.21.1. Check the required loader. |
| `26.2` | Java 25 | A fixed newer release for Vanilla play. Check loader support before using mods. |
| `LATEST` | Java 25 | The latest release. The game version might update on restart. |

Choose the version required by your modpack. A version appearing in this list does not guarantee that every loader or mod supports it. For a long-running world, use a fixed version instead of `LATEST`.

The installation fields are:

- **TYPE**: `VANILLA` for the original game without a mod loader, `FORGE` for Forge mods, or `FABRIC` for Fabric mods. The default is `VANILLA`. NeoForge is not offered by this package.
- **FORGE_VERSION**: Used only with `FORGE`. The default is `recommended`; you can also enter `latest` or the exact Forge version required by your modpack.
- **FABRIC_LOADER_VERSION**: Used only with `FABRIC`. The default is `latest`; enter an exact loader version when your modpack requires one. Fabric API is a separate mod.

**VERSION**, **TYPE**, and the loader versions are fixed after installation. To run another version or loader, create a separate instance with Clone. In the official Java launcher, open **Installations**, create an installation, and select the same game version as the server. For modded play, also install the compatible loader and required client mods.

:::warning Protect existing worlds
Back up the instance's entire data directory before migrating a world or reinstalling. Reinstalling with another version can reuse the existing data directory. Upgrading can convert world data, while downgrading can prevent startup or lose chunks, items, and entities. Changing loaders or removing mods can also make a world incompatible. Never start an older server against a world saved by a newer version.
:::

### Install Forge or Fabric mods

For example, choose `VERSION=1.20.1` and `TYPE=FORGE` or `TYPE=FABRIC`, then set the loader version required by your mods.

1. Stop the target Minecraft instance from Market or Settings.
2. Open Files and locate `Data/<instance-name>/data`. For the primary instance, this is `Data/minecraft/data`. Use the instance's data folder name, which can differ from its display title.
3. Create a `mods` folder if it does not exist, then upload the server-side mod `.jar` files and their required dependencies into it. Use mods for the selected Minecraft version and loader. Install Fabric API here when a mod requires it; keep client-only mods on the client.
4. Add any configuration files required by the modpack to the corresponding folders under the same `data` directory. Uploading a modpack archive alone does not install it.
5. Resume the instance and check the console logs for missing dependencies or incompatible mods.
6. Connect using the matching game version, compatible loader, and the mods required on the client. Some server-only mods allow an unmodified client; follow each mod's requirements.

### Disable account verification for a private server

Account verification is enabled by default. For an offline-mode private server with trusted players:

1. Go to **Settings** > **Applications** and select the target Minecraft instance.
2. Under **Environment variables**, click **Manage environment variables**.
3. Edit **ONLINE_MODE**, select `false` (**Disabled (trusted players only)**), and click **Confirm**.
4. Click **Apply** and wait for the server to restart.

This disables Minecraft account verification and secure profile enforcement. It supports offline-mode private servers but does not change the game's client licensing requirements or network access settings. Players still need the correct game version and a reachable server address.

:::warning Trusted players only
Offline mode does not verify player identities, so another player can use a name belonging to someone else. Restrict access to trusted players. Switching between online and offline mode changes player UUIDs and can affect inventories and permissions; back up the data first.
:::

### Allow cheat commands

In the same **Manage environment variables** page, set **ALLOW_CHEATS** to `true`, click **Confirm**, then **Apply**. The server restarts. The default is `false`.

When enabled, every player who joins receives permission to use commands such as `/give` and `/gamemode`, and command blocks are enabled. This grants level 2 operator permissions; server-stop commands remain available through the console. Use this setting only with trusted players.

To disable cheat access, set **ALLOW_CHEATS** to `false` and apply the change. This clears the player operator list, including manually added operators, and restarts the server.

### Set the maximum player count

The default is 8 players online at once. In **Manage environment variables**, edit **MAX_PLAYERS** to an integer from `1` to `100`, click **Confirm**, then **Apply**. The server restarts to apply the limit. Higher limits require more resources, especially with mods.

### Create another server with Clone

Clone lets you run separate versions or mod setups on the same Olares device. It creates a new instance with its own configuration and data directory; it does not copy the original world's data or mods.

1. Install the primary Minecraft instance first, then open **My Olares** in Market.
2. Find Minecraft, click the drop-down arrow next to **Open**, and select **Clone**.
3. Enter a unique **New app title** and **Desktop shortcut name**, then click **Confirm**.
4. In **Configure Environment Variables**, select **VERSION**, **TYPE**, and the loader settings for the new server, then click **Confirm** and wait for installation.
5. Manage the new instance separately. Upload its mods to `Data/<instance-name>/data/mods` and enable Overlay for that instance when using LAN access.

Each instance receives its own external port for VPN access and its own Overlay address when enabled. Copy the address for the instance you want to join using the connection steps below. The Overlay port remains `25565`; the external VPN port is assigned separately. Running several servers also increases memory and CPU use.

#### Reuse the original instance's data

You can copy an existing world to a Clone to continue playing or test another version. With matching game, loader, and mod versions, the world can usually be reused. Upgrading might convert the world and requires compatible mods. Downgrading or changing loaders can prevent startup or lose world content. You can choose to test compatibility in the Clone while keeping the original instance as a fallback.

1. In Market > **My Olares**, open the target Clone's details. Its App ID is the URL segment after `/app/<market-source>/` and before `?`. For example, `/app/market.test/minecraftb5764c?source=…` identifies `minecraftb5764c`, whose data directory is `Data/minecraftb5764c/data`. Check the original instance's ID the same way.
2. Stop both instances and back up their entire data directories. Replace the Clone's `data` contents with a copy of the original instance's `data` contents.
3. Check the Clone's mods and dependencies against the target version. Review its environment variables separately; they are not copied with the files. Keep **ONLINE_MODE** unchanged to preserve player UUIDs and inventory associations, and review the copied operator, whitelist, and ban lists.
4. Start the Clone, check its logs, then join using its own address. Verify builds, inventories, entities, and mod behavior. Stop it if compatibility errors appear.

:::warning Keep a backup
Each instance must use a separate data directory. Progress is not synchronized between them. To roll back, restore the backup from before the upgrade; do not copy an upgraded world back to an older server.
:::

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
4. Open **Settings** > **Applications** > **Minecraft** (or the target Clone) > **Ports**. Copy the address and external port shown for that instance and enter them in **Server Address**. The external port is assigned per instance; do not assume it is `25565` or reuse another instance's address.

5. Save the server and click **Join Server**.

## Manage the server

The Minecraft app does not have a web management interface. To view logs or run server commands, open it from Launchpad and use the built-in console terminal.

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
- You copied the address and external port from **Settings** > **Applications** > the target instance > **Ports**.
- The server is ready, and your client matches its game version and mod requirements.
- Your LarePass VPN connection is enabled.

### What happens when I upgrade the app?

Upgrading the app restarts the server. Any players currently online will be disconnected.

## Learn more

- [Manage overlay gateway for applications](/manual/olares/settings/overlay-gateway.md): Configure LAN access for supported apps.
- [Manage application environment variables](../manual/olares/settings/manage-app-env.md): Change settings and apply them with a restart.
- [Clone applications](../manual/olares/market/clone-apps.md): Create and manage independent app instances.
