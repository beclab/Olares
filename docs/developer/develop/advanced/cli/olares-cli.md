---
outline: [2, 3]
---
# Olares CLI

Olares provides Olares CLI, a command-line tool for developers and system administrators to customize or troubleshoot the Olares installation process.

The recommended [one-liner installation command](../../../../manual/get-started/install-olares.md) retrieves a shell script from https://olares.sh/ that downloads and installs Olares CLI. Once installed, the CLI orchestrates the remainder of the setup.

In general, Olares CLI manages installation through three main phases:
1. **Download**: Olares CLI fetches the necessary components.
2. **Prepare**: Olares CLI prepares the environment for installation.
3. **Install**: Olares CLI installs the core services of Olares.

This page explains the Olares CLI syntax and describes the command operations.

:::info Root privileges required
Most `olares-cli` commands require root privileges. Use the root user or prepend commands with `sudo`.
:::

## Syntax
The Olares CLI uses the following syntax:

> `olares-cli command [subcommand] [option]`

where
- `command`: Specifies the main operation you want to perform. For example, `olares download`.
- `subcommand`: Further specifies the task of `command`. For example, `wizard` or `component`.
- `option`: Optional arguments that modify the behavior of the `command`. Options include flags and options with arguments.

Olares CLI allows you to temporarily override certain Olares default settings. Each option applies only to the command in which it is used.

For example, if you use the `--base-dir` option with `olares-cli olares download wizard`, it will only affect the wizard downloading process and will not change the base directory for other commands, such as during the "install" phase.

To get detailed help for any command, run `olares-cli help`.

## Available CLI commands

| Operation          | Syntax                                             | Description                                                                                                                  |
|--------------------|----------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|
| `olares info`      | `olares-cli olares info [option]`                  | Displays general information about the downloaded Olares OS.                                                                 |
| `olares download`  | `olares-cli olares download [subcommand] [option]` | Downloads specific resources.                                                                                                |
| `olares prepare`   | `olares-cli olares prepare [option]`               | Prepares the environment for the installation process, including setting up essential services and configurations of Olares. |
| `olares install`   | `olares-cli olares install [option]`               | Deploys system-level and user-level components of Olares.                                                                    |
| `olares change-ip` | `olares-cli olares change-ip [option]`             | Changes the IP address of the Olares OS.                                                                                     |
| `olares release`   | `olares-cli olares release [option]`               | Packages Olares installation resources for distribution or deployment.                                                       |
| `olares uninstall` | `olares-cli olares uninstall [option]`             | Uninstalls Olares completely, or roll back the installation to a specific phase.                                             |



