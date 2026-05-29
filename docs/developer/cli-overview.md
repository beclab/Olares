---
outline: [2, 3]
description: How to use olares-cli for cluster management, system diagnostics, and AI agent integration. Covers host, user, and in-cluster modes of operation.
---

# Olares CLI

`olares-cli` is the command-line tool for Olares. It is included when you install Olares, and you can also install the standalone Olares CLI on macOS, Windows, or Linux without installing the full OS. You can use it for the following tasks:

- **Manage the cluster**: install Olares on a new machine, upgrade between releases, add or remove nodes, and uninstall completely.
- **Diagnose and repair the system**: run pre-installation checks, collect logs, query node status, and troubleshoot issues.
- **Drive Olares from an AI agent**: let a personal agent manage files, apps, and settings on your behalf through natural language.

## Modes of Olares CLI
Olares CLI is available through three modes that differ in where they run and how they authenticate.

| Mode | Where it runs | How it authenticates |
|------|---------------|----------------------|
| Host mode | On the same machine as the Olares OS | Host root and kubeconfig. No login required.|
| User mode <Badge type="tip" text="^1.12.5" /> | On any machine with `olares-cli` installed,<br> on behalf of a logged-in user | Profile and access token via the same HTTP API as the web UI and LarePass |
| In-cluster mode <Badge type="tip" text="^1.12.6" /> | Inside an Olares app's container | Credentials injected as environment variables, with scope set by the app's `OlaresManifest` |

:::tip
To use Olares CLI in user mode, install the CLI with npm. If the `olares-cli` already on your host predates the version shown in the table, the agent commands may not be available. See [Install olares-cli](./cli-install.md).
:::

## Drive Olares from an AI agent

To let an AI agent run Olares CLI on your behalf:

1. Install the CLI and the Agent Skills. The recommended `npx @olares/cli install` sets up both at once. See [Install olares-cli](./cli-install.md).
2. Log in with your Olares ID. See [Log in to Olares](./cli-log-in.md).
3. Drive Olares in natural language through your agent.

## Learn more

- [Cluster management](./install/index.md): install Olares on a fresh machine, upgrade between releases, manage disks and GPUs, collect logs, etc.
- [Install olares-cli](./cli-install.md): install the standalone CLI on your machine.
- [Log in to Olares](./cli-log-in.md): create a profile so the CLI can act as your Olares user.
- [Install and use Agent Skills](./cli-agent-skills.md): add the skills your agent needs to drive Olares.
