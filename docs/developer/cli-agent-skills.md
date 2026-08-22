---
outline: [2, 3]
description: Install Olares CLI Agent Skills for Cursor, Claude Code, Hermes Agent, and OpenClaw. Learn bundle contents, install order, and end-to-end usage.
head:
  - - meta
    - name: keywords
      content: Olares, Agent Skills, olares-cli, Cursor, Claude Code, AI agent integration, install Agent Skills
---

# Install and use Agent Skills

The user and in-cluster modes of `olares-cli` are built for AI agents rather than interactive use. To support this, `olares-cli` includes a set of Agent Skill bundles, one per group of commands. Each bundle teaches an agent what each command does, which flags matter, how authentication works, and how to recover from common errors.

## Understand Agent Skills

An Agent Skill is a `SKILL.md` bundle that an AI runtime loads as a tool definition. When the agent receives a natural-language request like "list the files in my Olares Home folder", it consults the loaded skill to find the corresponding command (`olares-cli files ls /drive/Home`), and runs it on your behalf.

The bundles are located in [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills) in the Olares repository. Each bundle contains a `SKILL.md` and a `references/` folder. `SKILL.md` carries routing logic, cross-cutting concepts, and an error-to-fix matrix. `references/` holds one file per non-trivial subcommand for details that are too long for `SKILL.md` but not in `--help`.

## Available Agent Skills

| Skill | Description |
|-------|--------|
| `olares-shared` | Log in to Olares, manage profiles and tokens, and recover from auth errors. Load this skill first. |
| `olares-chart` | Turn your own repo, docker-compose, or Helm chart into an Olares app and deploy it. |
| `olares-files` | Manage Olares files: upload, download, compress, extract, share, and mount SMB/NFS. Covers drive, sync, cache, and external namespaces. |
| `olares-market` | Install, upgrade, uninstall, clone, stop, resume, and restart Olares apps; browse the catalog; check status; and upload local charts. |
| `olares-settings` | Change Olares settings such as users, apps, VPN, network, backup, integrations, GPU, and search. |
| `olares-dashboard` | View system resource usage including CPU, memory, disk, network, pods, GPU, fan, and application ranking. |
| `olares-cluster` | Inspect K8s runtime state including pods, containers, workloads, jobs, cronjobs, nodes, and middleware. Read logs, exec, scale, restart, and suspend/resume cronjobs. |
| `olares-doctor` | Find out why an app is broken, such as stuck installs, crashes, image pull failures, running-but-unreachable apps, or slowdowns. Pulls evidence from cluster, dashboard, and market. |
| `olares-router` | Configure and use AI models: add a cloud vendor's models, install a local model app, pick the defaults, issue keys and quotas, read usage, and call a model for chat, embeddings, transcription, speech, or OCR. |
| `olares-knowledge` | Download to Olares: create, inspect, pause, and cancel URL, yt-dlp, aria2, torrent, and Hugging Face tasks, including the ones Wise starts. |
| `olares-search` | Search files and apps. Full-content search across Drive files and Sync libraries, plus search installed apps by title. |
| `olares-publish` | Publish an app that already runs locally to the public Olares Market: release targets, listing metadata, icon and screenshots. |

:::warning Always install `olares-shared` first
All other bundles assume `olares-shared` is already loaded. It owns the profile model, the token refresh logic, and the auth-error recovery hints that the other skills rely on. An agent that loads only `olares-files`, for example, encounters auth errors with no recovery path.
:::

## Install the skills manually

If you set up the CLI with `npx @olares/cli@latest install`, the skills are already installed and you can skip this step.

Otherwise, have the CLI write them out:

```bash
olares-cli skills install
```

The bundles are compiled into the binary, so this reads no network and fetches no version: what lands on disk is what the `olares-cli` you just ran was built from. One copy goes to `~/.agents/skills`, and every agent skills directory that already exists on the machine — Cursor, Claude Code, Codex, OpenClaw — is linked to it. Directories are never created for an agent you don't have. The agent then loads the matching skill when you mention an Olares task, and because `olares-shared` is written along with the rest, the shared-first requirement is satisfied automatically.

If `olares-cli` is not on the machine yet, [install it first](./cli-install.md) — the skills come with it.

:::tip
The skills are also listed on ClawHub, but that copy predates the move into the binary and is no longer updated. Install from the CLI instead; a registry copy has no way to tell you it disagrees with the verbs your binary actually has.
:::

Some AI agent apps on Olares bundle these skills, so the agent can manage Olares out of the box. To use the skills from such an app, see [Manage Olares with your Hermes Agent](../use-cases/hermes.md#manage-olares-with-your-hermes-agent) or [Manage Olares with your OpenClaw agent](../use-cases/openclaw-olares-skills.md).

## Update the skills

`olares-cli` and the Agent Skills are updated frequently. When a new version is available, use the path that matches your setup.

### Update skills bundled with an Olares agent app

AI agent apps on Olares, such as OpenCode and Hermes Agent, ship with the Olares CLI Agent Skills built in. Updating the app also updates the bundled skills. You don't need to run any extra CLI commands.

### Update locally installed skills

Update `olares-cli` first ([how](./cli-install.md#update-olares-cli)), then run the install command again:

```bash
olares-cli skills install
```

The skills come from the binary, so this order matters: the new CLI writes the skills that document it, replacing each bundle whole. Updating the CLI on its own leaves the previous skills in place, which is why every command other than `skills` prints one line on standard error while the two disagree. Set `OLARES_CLI_NO_SKILL_NOTICE=1` to silence that line where extra output can't be tolerated.

:::tip
You can also ask your AI agent to run these commands for you.
:::

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
If the agent doesn't load the Olares skills, explicitly invoke them with a slash (`/`) command.
:::
