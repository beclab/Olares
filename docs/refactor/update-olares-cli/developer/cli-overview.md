---
outline: [2, 3]
description: Overview of olares-cli, the official command-line tool for installing and operating Olares and for driving Olares from AI agents. Explains the three modes and which docs to read for which task.
---

# Olares CLI

`olares-cli` is the official command-line tool for Olares. It has three modes: host mode for cluster operations, user mode for acting on behalf of a logged-in user, and in-cluster mode for running inside an Olares app's container.

## Three modes

| Mode | Where it runs | How it authenticates |
|------|---------------|----------------------|
| Host mode | On the same machine as the Olares OS | Host root and kubeconfig. No login required.|
| User mode <Badge type="tip" text="^1.12.5" /> | On the Olares host, on behalf of a logged-in user | Profile and access token via the same HTTP API as the web UI and LarePass |
| In-cluster mode <Badge type="tip" text="^1.12.6" /> | Inside an Olares app's container such as Openclaw or NemoClaw | Credentials injected as environment variables, with scope set by the app's `OlaresManifest` |

:::tip
Currently, to use Olares CLI in user mode requires a standalone Olares CLI. If the `olares-cli` already on your host predates the version shown in the table, download the standalone CLI. See [Install olares-cli](./cli-installation.md).
:::

## Learn more

- [Cluster management](./install/index.md): install Olares on a fresh machine, upgrade between releases, manage disks and GPUs, collect logs, etc.
- [Use with AI agents](./cli-installation.md): install `olares-cli` on the machine and use it with Agent skills.
