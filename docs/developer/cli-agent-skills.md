---
outline: [2, 3]
description: Install and use the Olares CLI Agent Skill bundles from AI runtimes such as Cursor and Claude Code, or from Olares apps such as Hermes Agent and OpenClaw. Covers the bundles, the shared-first install order, and end-to-end usage.
---

# Install and use Agent Skills

The user and in-cluster modes of `olares-cli` are built for AI agents rather than interactive use. To support this, `olares-cli` includes a set of Agent Skill bundles, one per group of commands. Each bundle teaches an agent what each command does, which flags matter, how authentication works, and how to recover from common errors.

## Understand Agent Skills

An Agent Skill is a `SKILL.md` bundle that an AI runtime loads as a tool definition. When the agent receives a natural-language request like "list the files in my Olares Home folder", it consults the loaded skill to find the corresponding command (`olares-cli files ls /drive/Home`), and runs it on your behalf.

The bundles are located in [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) in the Olares repository. Each bundle contains a `SKILL.md` and a `references/` folder. `SKILL.md` carries routing logic, cross-cutting concepts, and an error-to-fix matrix. `references/` holds one file per non-trivial subcommand for details that are too long for `SKILL.md` but not in `--help`.

## Available Agent Skills

| Skill | Description |
|-------|--------|
| `olares-chart` | Deploy charts, package app images, validate manifests, configure wiring, and fix installs. |
| `olares-cluster` | Read and modify pods, workloads, nodes, jobs, cronjobs, and<br>  middleware passwords. |
| `olares-dashboard` | Overview and application metrics, with a stable JSON schema. |
| `olares-doctor` | Diagnose runtime issues for misbehaving apps and identify root causes. |
| `olares-files` | List, upload, download, share, mount SMB/NFS, and compress files. |
| `olares-market` | Browse, install, upgrade, uninstall, and upload local charts. |
| `olares-search` | Search globally across files, content, and apps with `olares-cli search`. |
| `olares-settings` | Read and modify settings that the web UI exposes. |
| `olares-shared` | Profile model, log-in flows, token storage, automatic refresh, and<br> auth-error recovery. Foundation for every other skill. |

:::warning Always install `olares-shared` first
All other bundles assume `olares-shared` is already loaded. It owns the profile model, the token refresh logic, and the auth-error recovery hints that the other skills rely on. An agent that loads only `olares-files`, for example, encounters auth errors with no recovery path.
:::

## Install the skills manually

If you set up the CLI with `npx @olares/cli@latest install`, the skills are already installed and you can skip this step.

Otherwise, install all bundles into your active agent with the following command:

```bash
npx skills add beclab/Olares -y -g
```

This installs the skills into the agent you are using, such as Cursor or Claude Code. The agent then loads the matching skill when you mention an Olares task. Because `olares-shared` is part of the same bundle, the shared-first requirement is satisfied automatically.

:::tip
The skills are also published on ClawHub. Both channels read the same `SKILL.md` files, so you only need to install from one. If your agent integrates with ClawHub, you can add them from there instead.
:::

Some AI agent apps on Olares bundle these skills, so the agent can manage Olares out of the box. To use the skills from such an app, see [Manage Olares with your Hermes Agent](../use-cases/hermes.md#manage-olares-with-your-hermes-agent) or [Manage Olares with your OpenClaw agent](../use-cases/openclaw-olares-skills.md).

## Use the skills

Once the bundles are loaded, control Olares with natural language. The agent determines which CLI command to run. For example:

```plain
# List files using the olares-files skill
List the files in the Home folder on my Olares device

# Install an app using the olares-market skill
Install Firefox from Market and tell me when it's ready

# Check resource usage using the olares-dashboard skill
Show me which apps are using more than 1 GB of memory
```

:::tip
If the agent doesn't load the Olares skills, explicitly invoke them with a slash command.
:::
