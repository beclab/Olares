---
outline: [2, 3]
description: Learn how Olares CLI Agent Skills help AI agents manage Olares, which skills are available, and how they are installed and updated.
head:
  - - meta
    - name: keywords
      content: Olares, Olares CLI, Agent Skills, olares-shared, skill catalog, AI agent integration
---

# Olares CLI Agent Skills

Olares CLI includes Agent Skills that teach AI agents how to manage Olares through the `olares-cli` command. This page lists the available skills and explains how to install, update, and inspect them.

If you are setting up Olares CLI for the first time, follow [Use Olares CLI with AI agents](./cli-install.md). That guide covers installation, login, and your first AI agent task in one flow.

## Available Agent Skills

| Skill | Description |
|-------|-------------|
| `olares-shared` | Routes Olares requests, explains profiles and authentication, and provides login recovery. |
| `olares-chart` | Turns a repository, Docker Compose project, or Helm chart into an Olares app and deploys it. |
| `olares-files` | Manages Olares files, including upload, download, compression, sharing, and SMB or NFS mounts. |
| `olares-market` | Browses Olares Market and installs, upgrades, uninstalls, clones, stops, resumes, or restarts apps. |
| `olares-settings` | Manages Olares settings such as users, apps, VPN, network, backups, integrations, GPU, and search. |
| `olares-dashboard` | Reads CPU, memory, disk, network, pod, GPU, fan, and application resource usage. |
| `olares-cluster` | Inspects Kubernetes runtime state, logs, containers, workloads, jobs, nodes, and middleware. |
| `olares-doctor` | Diagnoses failed installs, crashes, image pull errors, unreachable apps, and performance problems. |
| `olares-router` | Configures cloud or local AI models, defaults, keys, quotas, usage, and model calls. |
| `olares-knowledge` | Manages URL, yt-dlp, aria2, torrent, Hugging Face, and Wise download tasks. |
| `olares-search` | Searches Drive files, Sync libraries, and installed apps. |
| `olares-publish` | Publishes a locally running app to the public Olares Market, including listing metadata and assets. |

## Install or export the skills

Write the bundled skills to the standard locations:

```bash
olares-cli skills install
```

One copy is written to `~/.agents/skills`. Olares CLI then links it into every compatible agent skills directory that already exists, including directories used by Cursor, Claude Code, Codex, and OpenClaw. It does not create directories for agents that are not installed. On systems where symbolic links are unavailable, Olares CLI copies the files instead.

To write the suite to another location, such as a project directory or container image, export it:

```bash
olares-cli skills export ./skills
```

## Keep the skills matched to Olares CLI

For an npm-installed copy, update Olares CLI and its installed skills together:

```bash
olares-cli update
```

If you update the CLI manually or install another build, run:

```bash
olares-cli skills install
```

When the installed skills were written by a different CLI build, Olares CLI prints a notice on standard error. Set `OLARES_CLI_NO_SKILL_NOTICE=1` only when an automation cannot tolerate the extra output.

Install skills from the Olares CLI binary rather than a registry or a Git branch. Registry and repository copies can drift away from the commands supported by the installed binary. The ClawHub copy predates the embedded bundles and is no longer updated.

## How Agent Skills work

Each skill is a `SKILL.md` bundle that an AI runtime loads when a request matches its description. The skill maps a natural-language task to the appropriate Olares CLI commands and explains important flags, authentication behavior, and common error recovery.

The source bundles are in [`cli/skills/`](https://github.com/beclab/Olares/tree/main/cli/skills). Each bundle contains a `SKILL.md` file and, when needed, a `references/` directory with detailed instructions for individual commands.

The released bundles are compiled into the Olares CLI binary. Commands such as `skills install`, `skills export`, and `skills read` therefore use the same skill content as the CLI version that runs them.

## Inspect the bundled skills

List the skills carried by the current Olares CLI binary:

```bash
olares-cli skills list
olares-cli skills list -o json
```

Read a skill or one of its reference files without installing it:

```bash
olares-cli skills read olares-shared
olares-cli skills list olares-router/references
olares-cli skills read olares-router references/olares-router-calling.md
```

These commands read the skills bundled with your installed CLI. They do not connect to Olares or check which skills your agent can access. To check skill availability in the agent, follow [Run your first task with an AI agent](./cli-install.md#run-your-first-task-with-an-ai-agent).
