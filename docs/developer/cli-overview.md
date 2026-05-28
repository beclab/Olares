---
outline: [2, 3]
description: How to use olares-cli for cluster management, system diagnostics, and AI agent integration. Covers host, user, and in-cluster modes of operation.
---

# Olares CLI

`olares-cli` is the command-line tool for Olares. You can use it for the following tasks:

- **Manage the cluster**: install Olares on a new machine, upgrade between releases, add or remove nodes, and uninstall completely.
- **Diagnose and repair the system**: run pre-installation checks, collect logs, query node status, and troubleshoot issues.
- **Drive Olares from an AI agent**: let a Personal Agent manage files, apps, and settings on your behalf through natural language.

Internally, `olares-cli` is also used to integrate Olares operations into CI pipelines, run automated tests, and script repetitive maintenance tasks.

## Modes of Olares CLI
Olares CLI is available through three modes that differ in where they run and how they authenticate.

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
