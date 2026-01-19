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

> `olares-cli command [subcommand] [argument] [options] `

where
- `command`: Specifies the main operation you want to perform. For example, `olares-cli install`.
- `subcommand`: Further specifies the task for commands that support additional operations. For example, `wizard` or `component`.
- `argument`: Specifies the target resource or input data for the command, typically an ID, name, or file path. For example, in `olares-cli user activate <Olares ID> [options]`, `<Olares ID>` is the argument.
- `options`: Optional arguments that modify the behavior of the `command`. Options include flags and options with arguments.

Olares CLI allows you to temporarily override certain Olares default settings. Each option applies only to the command in which it is used.

For example, if you use the `--base-dir` option with `olares-cli download wizard`, it will only affect the wizard downloading process and will not change the base directory for other commands, such as during the "install" phase.

To get detailed help for any command, run `olares-cli help`.

## Available CLI commands

| Operation  | Syntax   | Description     |
|--|--|--|
| `backups` | `olares-cli backups <subcommand> [options]` | Manages backup-related operations.  |
| `change-ip` | `olares-cli change-ip [options]` | Changes the IP address of the Olares OS.  |
| `disk` | `olares-cli disk <subcommand>` | Manages storage resources in the Olares system. |
| `download`  | `olares-cli download <subcommand> [options]` | Downloads specific resources. |
| `gpu` | `olares-cli gpu <subcommand> [options]`  | Manages GPU-related operations. |
| `info` | `olares-cli info [options]` | Displays general information about the downloaded Olares OS. |
| `install`  | `olares-cli install [options]`  | Deploys system-level and user-level components of Olares. |
| `logs` | `olares-cli logs [options]` | Collects logs from Olares system components for debugging and troubleshooting.  |
| `node`  | `olares-cli node <subcommand> [options]` | Manages node-related operations. |
| `osinfo`   | `olares-cli osinfo <subcommand> [options]`  | Displays general information about the operating system of the current device. |
| `precheck`  | `olares-cli precheck [options]` | Verifies whether the system environment meets all requirements for Olares installation.  |
| `prepare`  | `olares-cli prepare [options]` | Prepares the environment for the installation process, including setting up essential services and configurations of Olares. |
| `release`   | `olares-cli release [options]` | Packages Olares installation resources for distribution or deployment.  |
| `start`  | `olares-cli start [options]` | Starts Olares services and components. |
| `stop` | `olares-cli stop [options]` | Stops Olares services and components.  |
| `uninstall` | `olares-cli uninstall [options]` | Uninstalls Olares completely, or roll back the installation to a specific phase. |
| `upgrade` | `olares-cli upgrade <subcommand> [options]` | Upgrades Olares and checks upgrade readiness and compatibility.|
| `user` | `olares-cli user <subcommand> [options]`| Manages users in the Olares system |