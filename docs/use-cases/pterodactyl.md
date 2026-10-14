---
outline: [2, 3]
description: Create and manage a Terraria/TShock server on Olares with Pterodactyl Panel, including console access, scheduled tasks, permissions, and backup recovery.
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl Panel, Wings, Terraria, TShock, game server, backup, server management"
app_version: "0.0.12"
doc_version: "2.1"
doc_updated: "2026-10-14"
---

# Manage game servers with Pterodactyl Panel

Pterodactyl Panel is an open-source game server management application. It brings server consoles, files, scheduled tasks, backups, and access permissions into one browser interface. Wings runs the game processes and handles their files and backups. Wings has no separate web management interface: use Panel to configure nodes and manage their servers.

This guide creates a new Terraria 1.4.5.8 server with TShock 6.2.1 on the local Wings node. TShock adds server administration and plugin support to Terraria.

:::info Which servers can Panel manage?
Panel manages game servers created through Wings. Applications installed separately from Olares Market, including Minecraft, remain managed by Olares and do not appear in Panel. Installing Panel does not discover or take over those applications. See [Existing Minecraft applications](#can-panel-manage-my-existing-minecraft-application) before planning a migration.
:::

## Learning objectives

- Configure a local node and its game port allocation.
- Create and operate a Terraria/TShock server with a compatible template.
- Grant limited access and schedule routine tasks.
- Back up a world and check that a restored server works.

## Prerequisites

- MariaDB installed from Market.
- An Olares account that can install applications and a Panel administrator account for node and server configuration.
- Internet access for downloading the game runtime and official TShock server files.
- A computer on the same local network as Olares, with Terraria 1.4.5.8 for a player connection check.

The local Wings integration is experimental. Its Docker service uses a privileged container. Game templates must be compatible with UID/GID 1000 and the available network ports. Check each new template before using it for a persistent world.

## Install Pterodactyl Panel

1. Open Market and search for "Pterodactyl Panel".
   <!-- ![Pterodactyl Panel in Market](/images/manual/use-cases/pterodactyl.png#bordered) -->

2. Click **Get**, then **Install**. In the installation form, keep **ENABLE_WINGS** enabled, set **ADMIN_USERNAME** and **ADMIN_PASSWORD**, and wait for installation to complete. The password must contain at least eight characters.

## Configure the local node

1. Open Pterodactyl Panel from Launchpad. Log in with the username and password you set during installation.

2. Open the administrator area and go to **Nodes** > **Olares Local Wings** > **About**. Check that **Daemon Version** and **System Information** load. If they remain unavailable, resolve the node connection before creating a server.

   ![Local Wings node reporting its version and available resources](https://cdn.olares.com/images/2026/10/7e42f94f2ec59d6abab793d424308add.png#bordered)

3. Open **Allocation** and create an allocation with IP address `0.0.0.0` and port `30000`.

An allocation tells Wings which internal port a server can use. The application exposes this port as TCP `43000` on the Olares device. Adding another allocation in Panel alone does not expose another port on the device.

| Purpose | Address or port | How to use it |
| --- | --- | --- |
| Panel | Application entrance | Open it from Launchpad; obtain its URL from **Settings** > **Applications** > **Pterodactyl Panel** > **Entrances** |
| Terraria | Olares LAN IP, TCP `43000` | Use Join via IP in Terraria; enter the IP and port `43000` separately |
| Internal allocation | `0.0.0.0:30000` | Select it when creating the server; it is not the player's connection address |
| SFTP | Olares LAN IP, TCP `42022` | Use an SFTP client with the username shown in the server's **Settings** > **SFTP Details** |

Replace `OLARES_LAN_IP` with your device's actual local address. The Panel HTTPS entrance is for browser administration; game clients and SFTP clients use their own protocols and ports. This guide covers local network access. Access from another network requires a separately configured and verified route to the game port.

## Create the Terraria/TShock server

### Import a compatible template

Pterodactyl calls a game template an Egg. An Egg defines the runtime image, installation script, startup command, and server variables.

Choose the JSON for your device's CPU architecture. Intel and AMD x86-64 devices use amd64; aarch64 devices use ARM64.

| Architecture | Download | China mirror |
| --- | --- | --- |
| amd64 | [TShock 6.2.1 Egg](https://cdn.olares.com/apps/demos/egg-tshock-6.2.1-olares-amd64.json) | [Download](https://cdn.olares.cn/apps/demos/egg-tshock-6.2.1-olares-amd64.json) |
| ARM64 | [TShock 6.2.1 Egg](https://cdn.olares.com/apps/demos/egg-tshock-6.2.1-olares-arm64.json) | [Download](https://cdn.olares.cn/apps/demos/egg-tshock-6.2.1-olares-arm64.json) |

1. Download the matching Egg JSON file.

2. In the administrator area, create a Terraria nest if needed. Open **Nests** > **Import Egg**, select that nest and the downloaded JSON file, and import it.

The Egg downloads and verifies the official [TShock 6.2.1 release for Terraria 1.4.5.8](https://github.com/Pryaxis/TShock/releases/tag/v6.2.1). It uses public Docker Hub images mirrored under beclab with fixed tags. You do not need to build an image or upload game binaries yourself.

| Purpose | Fixed image reference |
| --- | --- |
| .NET 9 runtime | `beclab/parkervcp-yolks:sha256-cb0eb2542645dd54bd1c965cc0b7e71249b9f125f9812b21e1191abe7bf8e099` |
| Installation tools | `beclab/parkervcp-installers:sha256-a4fbf7119883a179a47b5e5fa1a60e7e0925f9b2ccbaf0548cfb261705304056` |

The installer stages the release archive in the server volume rather than Wings' small `/tmp` filesystem, then removes its staging directory. Runtime files use UID/GID 1000. Worlds are stored in `worlds/`, TShock configuration and its SQLite database in `tshock/`, and additional plugins in `plugins/`. The example keeps the TShock REST API disabled.

:::info Verification scope
The amd64 Egg was imported through Panel's import service and installed by Wings on Olares. Startup, browser console commands, shutdown, restart, the exposed game protocol, and Wings backup recovery were verified. The beclab mirrors match the original image configurations and uncompressed layer identifiers; the mirrored amd64 runtime and installer were also pulled and started on Olares.

ARM64 image manifests, the official release archive, and its checksum were checked, but the ARM64 Egg has not been tested on an ARM64 Olares device. Real player joins, multiplayer load, and third-party plugin compatibility remain unverified.
:::

### Set the server configuration

1. Open **Servers** > **Create New** and enter a name such as `Terraria Friends`.

2. Select your account in **Server Owner**, choose **Olares Local Wings**, and select `0.0.0.0:30000` in **Default Allocation**.

3. Choose the imported architecture-specific TShock Egg and its listed .NET 9 runtime image. Use these small verification-server settings:

   | Field | Value |
   | --- | --- |
   | CPU Limit | `100` percent |
   | Memory | `1024` MiB |
   | Disk Space | `1024` MiB |
   | Backup Limit | `2` |
   | Game / TShock | `1.4.5.8` / `6.2.1` |
   | Internal game port | `30000` |
   | Maximum players | `2` |

   Keep **Tshock Version** at `v6.2.1`, **World Size** at `1`, **World Name** at `OlaresFriends`, and **Max Players** at `2` for this example. These values are a verification profile, not a capacity recommendation.

4. Leave **Start Server when Installed** disabled, create the server, and wait for installation to finish. Check **Files** for the writable `worlds/`, `tshock/`, and `plugins/` directories.

### Start and connect to the world

1. Open **Console** and click **Start**. Allow the first start to generate the world. Wait for the log to report `Server started`.

2. Enter `version`, then `playing` in the console command field. Confirm the version and online-player response. TShock console commands do not need the in-game slash prefix.

3. On a computer on the same network, open Terraria 1.4.5.8 and choose **Multiplayer** > **Join via IP**. Select a player, enter `OLARES_LAN_IP`, and enter port `43000` when prompted. Confirm that the player enters the world.

A successful protocol handshake is only a reachability check. The historical verification did not include a real player join; perform that check before inviting players.

## Operate the server

### Use the console and files

Use **Console** for commands and **Start**, **Restart**, and **Stop** for lifecycle control. Use **Stop** for normal shutdown so the game can save its world; reserve forced termination for an unresponsive process.

In **Files**, inspect `worlds/` and `tshock/`. Stop the server before editing `tshock/config.json`, save the change, and start it again. Keep the game port at `30000`; command-line settings can override the configuration file. Add only version-compatible plugins to `plugins/` after testing them on a disposable server.


For SFTP, copy the server-specific username from **Settings** > **SFTP Details**, connect to the Olares LAN IP on port `42022`, and use the corresponding Panel account password. Check access with a small disposable text file before transferring a world. Do not use plain FTP.

### Schedule a routine command

1. Open **Schedules** and create a schedule. Give it a descriptive name and set the minute, hour, day, month, and weekday fields for the intended time.

2. Add a command task with `playing`. Enable **Only When Server Is Online** for this example.

3. For an initial check, run the command once per minute by setting all five timing fields to `*`. Confirm that the player-count response appears in **Console**, then disable the schedule or change it to the intended frequency.

Check the schedule's timezone before choosing an operating time. Avoid leaving a temporary per-minute schedule enabled.

### Grant limited access

Open **Users** on the server and add a subuser. Select only the permissions needed for the person's task, such as viewing the console or accessing specified file operations. Test the account in a separate browser session and check both an allowed action and a denied action.

Panel subuser permissions control server management. They do not replace the game's own player permissions or plugin permission groups.

### Create and restore a backup

1. Stop the server and wait for its status to become offline.

2. Open **Backups**, click **Create backup**, and give the backup a name that includes its purpose or date. Wait until the backup finishes; a requested or queued backup is not a completed backup.

3. Before a change, keep a completed backup. To recover, keep the server offline, choose **Restore** for the intended backup, review the replacement options, and confirm the operation.

4. After restoration finishes, start the server. Check that it listens on port `30000`, that a matching Terraria client can join, and that the expected world state is present. Include the TShock SQLite database when recovering game accounts and permission groups.

:::warning Back up the correct scope
A restore can overwrite files. Validate recovery with a disposable server before relying on it for a valued world.

Game-server backups cover that server's files. They do not back up Panel's users, nodes, schedules, or database. Before uninstalling Panel, arrange a separate database export and a backup of its application data, including the encryption key. Keeping application data alone does not guarantee that the middleware database will survive reinstallation.
:::

## Check the setup

Before inviting players, check that the server starts, accepts a console command, retains its world after a normal restart, and can recover from a completed backup. Perform a real player join check and grant users only the permissions they need.

These management checks do not establish multiplayer capacity or compatibility with arbitrary mods and plugins.

## FAQs

### Can Panel manage my existing Minecraft application?

Panel cannot directly manage Minecraft installed as a separate Market application. Olares manages that application as a Kubernetes workload, while Wings creates and controls its own game containers. Wings has no mechanism in this integration to adopt an existing Olares application.

To keep using the Market application, continue managing it through Olares and its own interface. To move to Panel, create a separate Minecraft server with a compatible Egg, match the game version and any mod/plugin runtime, stop the original server, and copy its world and configuration into the new server. Validate player access, world contents, and backups before retiring the original. Panel and game-account permissions need separate configuration. This is a data migration, not linking or taking over the existing application.

### Why is the displayed address different from the player address?

Panel can show the internal allocation `0.0.0.0:30000`. Players use the Olares device's reachable LAN IP and exposed port `43000`. Entering the Panel URL or the internal allocation in the game client does not select the exposed game endpoint.

### Why does an imported Egg fail to install or start?

Check its runtime version, download access, startup command, file ownership, and port settings. This integration runs Wings and game processes with UID/GID 1000. A template that leaves files owned by another user might need adaptation. Check the installation log separately from the game console.

### Can I reset an existing administrator by changing installation values?

The initialization values create the first administrator. They do not overwrite an administrator that already exists. Use Panel's account recovery procedure; email-based recovery requires working SMTP configuration.

## Learn more

- [TShock documentation](https://github.com/Pryaxis/TShock/tree/v6.2.1/docs): Game commands, startup parameters, and permissions.
- [Pterodactyl documentation](https://pterodactyl.io/): Upstream administration and Egg configuration.
