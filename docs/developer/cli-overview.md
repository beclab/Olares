---
outline: [2, 3]
description: Use olares-cli to maintain Olares OS, manage a running Olares, and let AI agents perform tasks through bundled Agent Skills.
head:
  - - meta
    - name: keywords
      content: Olares, olares-cli, CLI overview, cluster management, Agent Skills, AI agent integration
---

# Olares CLI overview

`olares-cli` is the official command-line tool for Olares. Use it to install, upgrade, and maintain Olares OS, or manage files, apps, settings, and system status on a running Olares. Its bundled Agent Skills let AI agents turn natural-language requests into CLI commands.

Install the standalone CLI on macOS, Windows, or Linux to let an AI agent manage a running Olares from another computer. Olares OS also includes a host CLI for system maintenance.

## Get started

### Manage a running Olares with an AI agent

Install Olares CLI and its Agent Skills on the computer where your AI agent runs, then log in with your Olares ID. The agent can translate natural-language requests into commands for managing files, apps, settings, models, and system status. Follow [Use Olares CLI with AI agents](./cli-install.md) to complete setup and run your first task in one flow.

AI agent apps that run on Olares, such as [Lares](../use-cases/lares.md), [Hermes Agent](../use-cases/hermes.md), and [OpenClaw](../use-cases/openclaw.md), already include Olares CLI and the Agent Skills. Follow your app's guide to get started.

### Maintain Olares OS

On the Linux host running Olares OS, use the bundled CLI at `/usr/local/bin/olares-cli` to install or upgrade Olares, manage nodes and hardware, collect logs, and diagnose system problems. These operations use the host's root access and kubeconfig, so they do not require an Olares ID login.

Start with [Common Olares CLI commands](./install/cli/olares-cli.md) for status checks, log collection, and troubleshooting. See [Cluster management](./install/index.md) for installation architecture, environment variables, and versioning.

## Learn more

- [Olares CLI Agent Skills](./cli-agent-skills.md): browse the skill catalog and learn how the bundles are distributed and kept in sync.
- [Use Any AI Agent to Manage Olares, with Olares CLI](https://www.olares.com/blog/olares-cli-ai-agents-explained/): understand how Olares, the CLI, Agent Skills, and AI agents work together.
