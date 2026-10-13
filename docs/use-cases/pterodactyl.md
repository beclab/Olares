---
outline: [2, 3]
description: Create and manage a Minecraft Java server on Olares with Pterodactyl Panel, including console access, scheduled tasks, permissions, and backup recovery.
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl Panel, Wings, Minecraft, game server, backup, server management"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

# Manage game servers with Pterodactyl Panel

Pterodactyl Panel is an open-source game server management application. It brings server consoles, files, scheduled tasks, backups, and access permissions into one browser interface. Wings runs the game processes and handles their files and backups.

This guide creates a separate Minecraft Java 1.21.1 server on the local Wings node. It does not import or take over a Minecraft application that is already installed on Olares.

## Learning objectives

- Configure a local node and its game port allocation.
- Create and operate a Minecraft server with a compatible template.
- Grant limited access and schedule routine tasks.
- Back up a world and check that a restored server works.

## Prerequisites

- MariaDB installed from Market.
- An Olares account that can install applications and a Panel administrator account for node and server configuration.
- Internet access for downloading the game runtime and official Minecraft server files.
- A computer on the same local network as Olares, with Minecraft Java Edition 1.21.1 for a player connection check.

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
| Minecraft | Olares LAN IP, TCP `43000` | Enter `OLARES_LAN_IP:43000` in Minecraft Java Edition |
| Internal allocation | `0.0.0.0:30000` | Select it when creating the server; it is not the player's connection address |
| SFTP | Olares LAN IP, TCP `42022` | Use an SFTP client with the username shown in the server's **Settings** > **SFTP Details** |

Replace `OLARES_LAN_IP` with your device's actual local address. The Panel HTTPS entrance is for browser administration; game clients and SFTP clients use their own protocols and ports. This guide covers local network access. Access from another network requires a separately configured and verified route to the game port.

## Create the Minecraft server

### Import the example template

Pterodactyl calls a game template an Egg. An Egg defines the runtime image, installation script, startup command, and server variables.

1. Download the [Minecraft 1.21.1 example Egg](/files/use-cases/pterodactyl/vanilla-minecraft-1.21.1.json).

2. In the administrator area, open **Nests** > **Import Egg**. Select the Minecraft nest, choose the downloaded JSON file, and import it.

The example uses Java 21, creates files with UID/GID 1000, and starts with `eula=false`. It is based on the template and settings used for functional verification. Importing the template does not accept the Minecraft EULA.

### Set the server configuration

1. Open **Servers** > **Create New** and enter a name such as `Minecraft Friends`.

2. Select your account in **Server Owner**, choose **Olares Local Wings**, and select `0.0.0.0:30000` in **Default Allocation**.

3. Choose the imported **Vanilla Minecraft 1.21.1 (Olares)** Egg and its Java 21 image. Use these example settings:

   | Field | Value |
   | --- | --- |
   | CPU Limit | `100` percent |
   | Memory | `1536` MiB |
   | Disk Space | `2048` MiB |
   | Backup Limit | `2` |
   | Server Jar File | `server.jar` |
   | Server Version | `1.21.1` |

   These values describe a small verification server, not a recommended capacity for every world or player count. The example limits the Java heap to 1024 MiB and allows two players. Increase resources and player limits only after checking the workload.

4. Leave **Start Server when Installed** disabled, create the server, and wait for installation to finish. Open the server's **Files** page and check that `server.jar`, `server.properties`, and `eula.txt` exist.

### Accept the EULA and start the world

1. Read the [Minecraft EULA](https://www.minecraft.net/en-us/eula). Continue only if you agree to its terms.

2. Open `eula.txt` in **Files**, change `eula=false` to `eula=true`, and click **Save Content**.

3. Open **Console** and click **Start**. Wait for the `Done` message before connecting a player. The first start generates the world.

4. Enter `list` in the console command field. A response showing the number of online players confirms that console input reached the running game process.

   ![Minecraft running with live console output and resource statistics](https://cdn.olares.com/images/2026/10/e6b8791ebba2f51e2cdc6e783a6f0a58.png#bordered)

5. On a computer on the same local network, open Minecraft Java Edition 1.21.1, select **Multiplayer** > **Add Server**, and enter `OLARES_LAN_IP:43000`. Check that the server appears, then join it.

A server-list response confirms protocol reachability. Successfully joining the world is a separate check; it is not implied by a responsive port or a running status in Panel.

## Operate the server

### Use the console and files

Use **Console** for commands and **Start**, **Restart**, and **Stop** for lifecycle control. Use **Stop** for normal shutdown so the game can save its world; reserve forced termination for an unresponsive process.

In **Files**, open `server.properties` to change game settings. Stop the server before editing persistent configuration, save the change, and start it again. Keep `server-port=30000` for this allocation. Uploading plugins does not turn a Vanilla server into a plugin-compatible server; use the appropriate Egg and runtime for that game variant.

![Server files shown in Panel's file manager](https://cdn.olares.com/images/2026/10/3d08eb1c79172687b0622fcea9247940.png#bordered)

For SFTP, copy the server-specific username from **Settings** > **SFTP Details**, connect to the Olares LAN IP on port `42022`, and use the corresponding Panel account password. Check access with a small disposable text file before transferring a world. Do not use plain FTP.

### Schedule a routine command

1. Open **Schedules** and create a schedule. Give it a descriptive name and set the minute, hour, day, month, and weekday fields for the intended time.

2. Add a command task with `say Scheduled maintenance reminder`. Enable **Only When Server Is Online** for this example.

3. For an initial check, run the command once per minute by setting all five timing fields to `*`. Confirm that the message appears in **Console**, then disable the schedule or change it to the intended frequency.

Check the schedule's timezone before choosing an operating time. Avoid leaving a temporary per-minute schedule enabled.

### Grant limited access

Open **Users** on the server and add a subuser. Select only the permissions needed for the person's task, such as viewing the console or accessing specified file operations. Test the account in a separate browser session and check both an allowed action and a denied action.

Panel subuser permissions control server management. They do not replace the game's own player permissions or plugin permission groups.

### Create and restore a backup

1. Stop the server and wait for its status to become offline.

2. Open **Backups**, click **Create backup**, and give the backup a name that includes its purpose or date. Wait until the backup finishes; a requested or queued backup is not a completed backup.

3. Before a change, keep a completed backup. To recover, keep the server offline, choose **Restore** for the intended backup, review the replacement options, and confirm the operation.

4. After restoration finishes, start the server. Check that it reaches `Done`, that a matching Minecraft client can connect, and that the expected world state is present.

:::warning Back up the correct scope
A restore can overwrite files. Validate recovery with a disposable server before relying on it for a valued world.

Game-server backups cover that server's files. They do not back up Panel's users, nodes, schedules, or database. Before uninstalling Panel, arrange a separate database export and a backup of its application data, including the encryption key. Keeping application data alone does not guarantee that the middleware database will survive reinstallation.
:::

## Check the setup

Before inviting players, check that the server starts, accepts a console command, retains its world after a normal restart, and can recover from a completed backup. Perform a real player join check and grant users only the permissions they need.

Use the [Pterodactyl validation checklist](pterodactyl-test-cases.md) for detailed acceptance cases, expected results, recorded evidence, and coverage limits. Basic checks do not establish multiplayer capacity or compatibility with arbitrary mods and plugins.

## FAQs

### Can Panel manage my existing Minecraft application?

Panel manages servers created on a Wings node with a compatible Egg. It does not automatically take over a separate Minecraft application installed from Market. Migrate a copy of the world into a new Panel-managed server and validate it before retiring the original.

### Why is the displayed address different from the player address?

Panel can show the internal allocation `0.0.0.0:30000`. Players use the Olares device's reachable LAN IP and exposed port `43000`. Entering the Panel URL or the internal allocation in Minecraft does not select the exposed game endpoint.

### Why does an imported Egg fail to install or start?

Check its runtime version, download access, startup command, file ownership, and port settings. This integration runs Wings and game processes with UID/GID 1000. A template that leaves files owned by another user might need adaptation. Check the installation log separately from the game console.

### Can I reset an existing administrator by changing installation values?

The initialization values create the first administrator. They do not overwrite an administrator that already exists. Use Panel's account recovery procedure; email-based recovery requires working SMTP configuration.

## Learn more

- [Pterodactyl validation checklist](pterodactyl-test-cases.md): Repeatable checks for administration, game operation, permissions, and recovery.
- [Pterodactyl documentation](https://pterodactyl.io/): Upstream administration and Egg configuration.
