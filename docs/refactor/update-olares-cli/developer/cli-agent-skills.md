---
outline: [2, 3]
description: Install and use the Olares CLI Agent skill bundles from AI runtimes such as Cursor and Claude Code, or from Olares apps such as NemoClaw and Openclaw. Covers the bundles, the shared-first install order, and end-to-end usage.
---

# Install and use Agent skills

User mode and in-cluster mode of `olares-cli` are designed for AI agents, not for interactive use. To support this, `olares-cli` includes a set of Agent skill bundles, one per group of commands. Each bundle teaches an agent what each command does, which flags matter, how authentication works, and how to recover from common errors.

## Understand Agent skills

An Agent skill is a `SKILL.md` bundle that an AI runtime loads as a tool definition. When the agent receives a natural-language request such as "list the files in my Olares Home folder", it consults the loaded skill to know the corresponding command is `olares-cli files ls drive/Home/`, and runs it on your behalf.

The bundles are authored alongside the CLI itself and are located in [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) in the Olares repository. They are the authoritative reference for what each group of commands does.

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
Every other bundle assumes `olares-shared` is already loaded. It owns the profile model, the token refresh logic, and the auth-error recovery hints that the business skills rely on. An agent that loads only `olares-files`, for example, encounters auth errors with no recovery path.
:::

## Install Agent skills

The following uses NemoClaw as an example. The exact steps depend on your AI runtime such as Claude Code.

1. Open the OpenClaw Web UI and go to **Skills**.
2. In the ClawHub search box, enter `olares` to find Olares skills.
3. Install **Olares Shared** first because it's the foundation of the other Olares skills.
4. Install the remaining Olares skills, such as **Olares Files** and **Olares Market**.
5. Open the chat page and run `/reset` to start a new session so the agent picks up the newly installed skills.

## Use Agent skills

Once the bundles are loaded, drive Olares in natural language. The agent determines which CLI command to run. For example:


```plain
# List files using the olares-files skill
List the files in my Olares Home folder

# Install an app using the olares-market skill
Install Firefox from Market and tell me when it's ready

# Check resource usage using the olares-dashboard skill
Show me which apps are using more than 1 GB of memory
```

For the full surface of any command group, run `olares-cli <command> --help`.
