---
outline: [2, 3]
description: Install and use the Olares CLI agent skill bundles from AI runtimes such as Cursor and Claude Code, or from Olares apps such as NemoClaw and Openclaw. Covers the bundles, the shared-first install order, and end-to-end usage.
---

# Install and use agent skills

The user and in-cluster modes of `olares-cli` are built for AI agents rather than interactive use. To support this, `olares-cli` includes a set of agent skill bundles, one per group of commands. Each bundle teaches an agent what each command does, which flags matter, how authentication works, and how to recover from common errors.

## Understand agent skills

An agent skill is a `SKILL.md` bundle that an AI runtime loads as a tool definition. When the agent receives a natural-language request like "list the files in my Olares Home folder", it consults the loaded skill to find the corresponding command (`olares-cli files ls drive/Home/`), and runs it on your behalf.

The bundles are located in [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) in the Olares repository. They are the authoritative reference for what each group of commands does.

## Agent skill bundles

| Skill | Description |
|-------|--------|
| `olares-shared` | Profile model, log-in flows, token storage, automatic refresh, and<br> auth-error recovery. Foundation for every other skill. |
| `olares-files` | List, upload, download, edit, share, mount SMB, and manage Sync repos. |
| `olares-market` | Browse, install, upgrade, uninstall, and upload local charts. |
| `olares-settings` | Read and modify settings that the web UI exposes. |
| `olares-dashboard` | Overview and application metrics, with a stable JSON schema. |
| `olares-cluster` | Read and modify pods, workloads, nodes, jobs, cronjobs, and<br>  middleware passwords. |

:::warning Always install `olares-shared` first
All other bundles assume `olares-shared` is already loaded. It owns the profile model, the token refresh logic, and the auth-error recovery hints that the other skills rely on. An agent that loads only `olares-files`, for example, encounters auth errors with no recovery path.
:::

## Install Olares skills

The following steps use Claude Code as an example. Exact steps differ depending on your AI runtime.

1. Open a terminal on your computer, and run the following command to install the ClawHub CLI.
    ```bash
    npm i -g clawhub
    ```
2. Install the agent skills to `~/.claude/skills/`.

    a. Install `olares-shared` first.

    ```bash
    clawhub install --workdir ~/.claude/skills --dir . olares-shared
    ```

    b. Install the remaining skills, such as `olares-files`.

    ```bash
    clawhub install --workdir ~/.claude/skills --dir . olares-files
    ```
3. Launch a Claude Code session and type `/` to see a list of available skills, or run `/skills` to confirm they have loaded.


To use Olares CLI with AI agents on Olares, such as NemoClaw or OpenClaw, refer to [Manage Olares with Olares CLI](../use-cases/nemoclaw-olares-cli.md).

## Use Olares CLI as agent skills

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