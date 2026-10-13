---
outline: [2, 3]
description: Acceptance cases for Pterodactyl on Olares covering access control, game operation, scheduling, file transfer, and backup recovery.
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl, acceptance testing, Minecraft, Terraria, TShock, permissions, recovery"
app_version: "0.0.12"
doc_version: "1.1"
doc_updated: "2026-10-13"
---

# Validate a Pterodactyl game server setup

Use this checklist after installing Pterodactyl Panel from Market, changing a game template, or preparing a server for players. It defines repeatable steps and pass criteria for an administrator. It also distinguishes previous runtime evidence from checks that still require execution.

## Scope and execution rules

The checklist covers Panel administration, the local Wings node, a primary Terraria/TShock server, and a separate optional Minecraft Java server. A healthy application status, an accepted API request, or a responsive TCP port alone is not a pass for game operation or recovery.

Run recovery and permission tests on disposable servers. Record the application version, game version, template, time, expected result, actual result, and evidence for every case. Do not attach passwords, access tokens, session cookies, or a complete Wings configuration to the results.

| Priority | Meaning |
| --- | --- |
| P0 | Access control, game startup, or recoverability; a failure blocks use of the affected workflow |
| P1 | Core management operation; investigate a failure before routine operation |
| P2 | Compatibility or capacity qualification for a particular environment |

In the tables, Recorded means equivalent behavior was observed in a previous runtime verification. Required means the described case has not been demonstrated end to end. Recorded evidence does not replace a fresh acceptance run of the installed release.

## Environment and test data

| Item | Baseline |
| --- | --- |
| Documentation target | Application release `0.0.12` |
| Panel / Wings | `1.15.1` / `1.13.3` |
| Verified runtime architecture | `amd64`; `arm64` image availability is not runtime qualification |
| Minecraft | Java Edition `1.21.1`, Java 21, matching client version |
| Example resources | TShock: CPU `100%`, memory `1024 MiB`, disk `1024 MiB`, backup limit `2` |
| Example world | New disposable small Terraria world, maximum two players |
| Port mapping | Internal TCP `30000` → device TCP `43000`; SFTP TCP `42022` |
| Primary TShock scenario | TShock `6.2.1`, Terraria `1.4.5.8`, a compatible template and diagnostic plugin |
| Accounts | Administrator, ordinary Panel account, and a server subuser with restricted permissions |

The resources above are a verification profile. Do not use them as a multiplayer capacity claim. Start with the [setup guide](pterodactyl.md) for the Terraria/TShock template requirements and connection instructions.

Only one server can own the default allocation. Run the Minecraft and Terraria scenarios sequentially, or use another port whose device exposure has already been configured.

## Acceptance cases

### Installation and authentication

| ID / priority | Procedure | Pass criteria | Evidence status |
| --- | --- | --- | --- |
| PT-01 / P0 | On a clean installation from Market, set the administrator username and password. Open Panel after installation and log in with those values. | The supplied account can log in; there is no unexpected preset credential. | Required for clean release `0.0.12` installation |
| PT-02 / P0 | Log in with valid credentials. Log out. Revisit a protected page and repeat with an incorrect password. | Valid login works; logout invalidates access; the incorrect password does not create an authenticated session. | Valid login and logout recorded; repeat the complete negative case |
| PT-03 / P0 | As an administrator, create an ordinary account, update its details, and log in with it. Attempt to open the administrator area. Delete the disposable account afterwards. | Account operations persist; the ordinary account cannot access administration. | Recorded |
| PT-04 / P1 | Create a narrowly scoped API key. Perform an allowed request, revoke the key, and repeat the same request. | The request works while authorized and is rejected after revocation. | Key authentication and revocation recorded; repeat scope checks |

### Node, installation, and game operation

| ID / priority | Procedure | Pass criteria | Evidence status |
| --- | --- | --- | --- |
| PT-05 / P0 | Open **Nodes** > **Olares Local Wings** > **About**. Check the daemon version and system information. Request the Wings API once with authorized access and once without credentials. | Node details load; authorized access works; anonymous access is rejected. | Recorded; anonymous request returned `401` |
| PT-06 / P0 | Create allocation `0.0.0.0:30000` and assign it to a disposable server. Probe the device's TCP `43000` with the matching game protocol. | The response is from the intended game server, not merely an open port. | Minecraft and Terraria protocol responses recorded |
| PT-07 / P0 | Install the validated TShock template, check world/configuration directories and their ownership, and start it. | Files are usable by UID/GID 1000; startup reports `Listening on port 30000`; the world exists. | Adapted TShock installation and startup recorded; repeat with the supplied Egg and image |
| PT-08 / P0 | From **Console**, issue `version` and `playing`, then stop the server normally. Start it again. | Commands reach the game; stop saves the world and exits normally; restart loads the existing world without an out-of-memory failure. | Console and lifecycle recorded; repeat these two example commands |
| PT-09 / P1 | On the disposable Minecraft server, run `gamerule doDaylightCycle false`, `setworldspawn 12 80 34`, and `save-all flush`. Stop and restart it, then check the persisted world settings. | The daylight-cycle rule remains false and the spawn remains `12,80,34`. | Recorded using persisted world data |
| PT-10 / P1 | Create a disposable text file through **Files**. Read and edit it. Upload a second file through SFTP and download it again. Compare its contents. | Allowed operations work; the downloaded file matches the original; the correct server-specific SFTP account is used. | File API and SFTP round trips recorded; repeat browser editing |
| PT-11 / P1 | Schedule `playing` once per minute, with the server online. Wait for execution, inspect the console and schedule state, then disable the schedule. | The command executes at the intended time, processing clears, and the disabled schedule stops producing command responses. | Execution and cleared processing recorded; repeat disabled-state observation |
| PT-12 / P1 | Check the configured CPU and memory limits against the actual running container and observe resource use while the world loads. | Limits match the server allocation; the game can start within them. | Recorded; this does not establish sustained-load capacity |

### Management and game permissions

| ID / priority | Procedure | Pass criteria | Evidence status |
| --- | --- | --- | --- |
| PT-13 / P0 | Grant a server subuser console-view access without file deletion or backup restoration. Log in as that subuser in a separate browser session and attempt all three operations. | Console viewing works; deletion and restoration are rejected; no forbidden side effect occurs. | Required; ordinary-account administration denial does not substitute for this case |
| PT-14 / P0 | On an isolated TShock server, create a restricted group and an account assigned to it. Load a diagnostic command with a specific permission. Dispatch it with the restricted group and with superadmin. Restart and repeat. | The restricted group cannot execute the command; superadmin can; account and group data persist. | Recorded with actual TShock command dispatch and simulated REST players |
| PT-15 / P0 | Repeat PT-14 by logging in from a matching Terraria client as the restricted account. Attempt the command, then repeat with an authorized account. | Real player authentication works and permission enforcement matches the intended groups. | Required; simulated REST players are not client-login evidence |

### Backup and recovery

| ID / priority | Procedure | Pass criteria | Evidence status |
| --- | --- | --- | --- |
| PT-16 / P0 | Stop the disposable server. Record SHA-256 hashes for the files in the recovery scope. Create a backup and wait for completion. | Backup finishes without an error, has a nonzero size, and represents the intended offline snapshot. | Recorded |
| PT-17 / P0 | With the server offline and a completed backup available, remove only the disposable world's files. Restore the backup. Compare hashes before starting the game, then start it and check the world settings. | Snapshot files match their original hashes; the restored world loads and retains the checked settings. | Recorded; all 19 Minecraft world files matched |
| PT-18 / P0 | On the isolated TShock server, include the world, configuration, and SQLite permissions database in an offline backup. Simulate their loss, restore, compare hashes, start the game, and repeat the permission probe. | Files match before startup; the world loads; account/group data and permission enforcement recover. | Recorded; 19 regular files matched |
| PT-19 / P0 | Back up Panel's database and application data through the administrator's approved procedure. Reinstall from Market in a disposable environment, restore both scopes, and check the original administrator, node, and server records. | The original records and encryption-dependent configuration remain usable after recovery. | Database export/restore recorded; repeat node/server recovery. Application-data retention alone was insufficient |

:::warning Recovery test boundary
PT-17 through PT-19 intentionally exercise data loss and replacement. Use disposable data and keep an independent verified backup. Do not perform a destructive rehearsal on the only copy of a production world or database.
:::

### Client, compatibility, and load qualification

| ID / priority | Procedure | Pass criteria | Evidence status |
| --- | --- | --- | --- |
| PT-20 / P0 | Join the primary world with a matching Terraria client from the same network and send matching and mismatched Terraria handshakes. For the optional Minecraft scenario, join its separate world with Java Edition 1.21.1. | Terraria player login and gameplay work; matching Terraria receives a valid response and the mismatched protocol is rejected; optional Minecraft players can join. | Terraria handshake acceptance/rejection recorded; real Terraria and Minecraft player login required |
| PT-21 / P2 | Repeat the installation, startup, access, and recovery cases on an `arm64` Olares device with the same template versions. | The workflows pass on the actual architecture. | Required |
| PT-22 / P2 | Before execution, define player count, duration, join pattern, world activity, and thresholds for tick performance, memory, and disconnections. Run that load and retain measurements. | Every agreed threshold is met without crashes, uncontrolled memory growth, or unexpected disconnections. | Required; no multiplayer load evidence exists |
| PT-23 / P2 | Record a specific mod/plugin version and its required runtime. Install it on a compatible disposable server, exercise its feature, restart, and repeat backup recovery. | The selected feature works and its persistent data survives restart and recovery. | A diagnostic TShock plugin was recorded; arbitrary third-party plugins and mod packs require separate qualification |

## Recorded baseline evidence

Runtime verification on 2026-10-12 used Panel 1.15.1, Wings 1.13.3, and an amd64 device. The recorded game runs used adapted templates. These results qualify those configurations, not every Egg or every client.

| Area | Recorded result |
| --- | --- |
| Panel | 61 assertions covering authentication, accounts, authorization, API-key lifecycle, queues, and restart persistence |
| Wings | Authenticated API, anonymous rejection, WebSocket authentication, SFTP file transfer, game-port access, and recovery |
| Minecraft | Official server installation, console commands, persisted world settings, scheduling, normal stop/restart, and world recovery |
| Minecraft recovery | A completed 107,985,166-byte backup; all 19 saved world files matched their original SHA-256 hashes after restoration |
| TShock | Terraria 1.4.5.8 with TShock 6.2.1; world generation, protocol handshake, diagnostic plugin, restricted-group enforcement, and account persistence |
| TShock recovery | A completed 6,623,322-byte backup; 19 regular files matched before restored startup; permission probe passed after startup |

Backup sizes describe the recorded runs. They are not fixed acceptance thresholds. A compatible handshake does not prove a player joined; a diagnostic plugin does not qualify third-party plugins; a successful small-server run does not establish multiplayer capacity.

## Record and sign off

Use one record per case:

| Field | Required information |
| --- | --- |
| Case | ID and priority |
| Environment | Application, Panel/Wings, game, template, architecture, resource limits, and connection method |
| Execution | Operator, date/time with timezone, and exact steps performed |
| Result | Pass, Fail, Blocked, or Not run; actual result and evidence location |
| Defect | Reproduction steps and impact; never include credentials |
| Cleanup | Temporary accounts, API keys, schedules, servers, and backups removed or intentionally retained |

Complete the applicable P0 cases before relying on the server. Record unexecuted or blocked cases explicitly. Capacity, architecture, and mod/plugin compatibility claims require results from their specific qualification cases.
