---
outline: [2, 3]
---
# Olares CLI

:::warning Note
Use this version of Olares CLI if your Olares version is 1.12.X.
:::

The Olares CLI is a versatile command-line tool designed to help developers and system administrators manage and troubleshoot Olares systems. It offers a wide range of features, from installation and configuration to resource management and diagnostics.

With the Olares CLI, you can streamline tasks such as verifying system compatibility, downloading resources, managing nodes, collecting logs, and more. This guide provides an overview of the CLI's syntax and details the commands available for different operations.

:::info Root privileges required
Most `olares-cli` commands require root privileges. Use the root user or prepend commands with `sudo`.
:::

:::info Using Olares CLI with WSL
If you installed Olares using the WSL (Windows Subsystem for Linux) method, you need to use `olares-cli` inside the WSL environment.

To enter the WSL environment, run the following command in PowerShell:

```powershell
wsl -d Ubuntu
```
:::

## Syntax
The Olares CLI uses the following syntax:

> `olares-cli command [subcommand] [option]`

where
- `command`: Specifies the main operation you want to perform. For example, `olares-cli install`.
- `subcommand`: Further specifies the task for commands that support additional operations. For example, `wizard` or `component`.
- `option`: Optional arguments that modify the behavior of the `command`. Options include flags and options with arguments.

Olares CLI allows you to temporarily override certain Olares default settings. Each option applies only to the command in which it is used.

For example, if you use the `--base-dir` option with `olares-cli download wizard`, it will only affect the wizard downloading process and will not change the base directory for other commands, such as during the "install" phase.

To get detailed help for any command, run `olares-cli help`.

## Available CLI commands

| Operation  | Syntax   | Description     |
|--|--|--|
| `gpu` | `olares-cli gpu <subcommand> [option]`  | Manages GPU-related operations. |
| `info`   | `olares-cli info <subcommand> [option]`  | Displays general information about the operating system of the current device. |
| `node`  | `olares-cli node <subcommand> [option]` | Manages node-related operations. |
| `backups` | `olares-cli backups <subcommand> [option]` | Manages backup-related operations.  |
| `change-ip` | `olares-cli change-ip [option]` | Changes the IP address of the Olares OS.  |
| `download`  | `olares-cli download <subcommand> [option]` | Downloads specific resources. |
| `info` | `olares-cli info [option]` | Displays general information about the downloaded Olares OS. |
| `install`  | `olares-cli install [option]`  | Deploys system-level and user-level components of Olares. |
| `logs` | `olares-cli logs [option]` | Collects logs from Olares system components for debugging and troubleshooting.  |
| `precheck`  | `olares-cli precheck [option]` | Verifies whether the system environment meets all requirements for Olares installation.  |
| `prepare`  | `olares-cli prepare [option]` | Prepares the environment for the installation process, including setting up essential services and configurations of Olares. |
| `release`   | `olares-cli release [option]` | Packages Olares installation resources for distribution or deployment.  |
| `start`  | `olares-cli start [option]` | Starts Olares services and components. |
| `stop` | `olares-cli stop [option]` | Stops Olares services and components.  |
| `uninstall` | `olares-cli uninstall [option]` | Uninstalls Olares completely, or roll back the installation to a specific phase. |
| `disk` | `olares-cli disk <subcommand>` | Manages storage resources in the Olares system. |
| `user` | `olares-cli user <subcommand> [option]`| Manages users in the Olares system |
| `upgrade` | `olares-cli upgrade <subcommand> [option]` | Upgrades Olares and checks upgrade readiness and compatibility.|