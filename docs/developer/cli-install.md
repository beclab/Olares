---
outline: [2, 3]
description: Install and sign in to Olares CLI, then use its Agent Skills to manage Olares with an AI agent.
head:
  - - meta
    - name: keywords
      content: Olares, Olares CLI, AI agents, Agent Skills, install Olares CLI, Claude Code, Codex, Cursor
---

# Use Olares CLI with AI agents

This guide shows you how to install Olares CLI and its matching Agent Skills on your computer, sign in to Olares, and complete your first task through an AI agent.

:::info Using an AI agent app on Olares?
[Lares](../use-cases/lares.md), [Hermes Agent](../use-cases/hermes.md), and [OpenClaw](../use-cases/openclaw.md) on Olares already include Olares CLI and its Agent Skills. You do not need to install them again inside the app. Follow your app's guide to get started.
:::

:::warning Olares hosts already include Olares CLI
If you are working directly on a Linux host that runs Olares OS, do not overwrite `/usr/local/bin/olares-cli`. See [Using Olares CLI on an Olares host](#using-olares-cli-on-an-olares-host).
:::

## Before you begin

Make sure you have:

- A computer running macOS, Windows, or Linux with Node.js 18 or later.
- An activated Olares and your Olares ID.
- An AI agent that supports Agent Skills, such as Codex, Claude Code, Cursor, or OpenClaw.

## Install Olares CLI and Agent Skills

Open a terminal and run:

```bash
npx @olares/cli@latest install
```

The setup wizard installs the same release of Olares CLI and its Agent Skills. The skills are written to `~/.agents/skills` and linked to the skill directories of compatible AI agents already installed on your computer.

Example output:

```text
┌  Setting up Olares CLI...
│
◇  Installed globally
│
◇  Skills installed
│
└  You are all set!
```

Check the installed CLI version:

```bash
olares-cli version
```

:::info
The setup wizard installs Olares CLI on your computer. It does not install Olares OS and does not sign in to your Olares account.
:::

## Log in to Olares

1. Start an interactive login with your Olares ID. For example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

   Replace `laresprime@olares.com` with your Olares ID.

2. Enter your Olares password when prompted. The password is not displayed in the terminal.

3. If two-step verification is enabled, enter the six-digit code from LarePass.

4. Check the active profile:

   ```bash
   olares-cli profile whoami
   ```

The profile tells Olares CLI which Olares to connect to and which user it should act as. Olares CLI refreshes the login token automatically.

## Run your first task with an AI agent

Open your AI agent using the same operating system account you used to install Olares CLI.

In the agent's chat window, check that it can find the Olares skills:

- If the agent lists skills as slash commands, type `/` and look for entries starting with `olares-`.
- Or ask the agent:

  ```plain
  List the available skills related to Olares.
  ```

Look for skills such as `olares-files` and `olares-market`. If they are missing, run `olares-cli skills install` again, check the reported installation locations, and start a new agent conversation.

Then ask the agent to complete your first task:

```plain
List the files in the Home folder on my Olares device.
```

The agent selects the relevant skill and runs the Olares CLI command. Check that it returns the contents of your Olares Home folder. This confirms that the agent can use the CLI to access your Olares.

Try a few more tasks:

```plain
Install Firefox from Olares Market and tell me when it is ready.

Show me which apps on my Olares are using more than 1 GB of memory.

Find the document named quarterly-report in my Olares files.
```

Your AI agent may ask for approval before running a command or changing Olares. Review the proposed action before approving it.

## Manage profiles

One Olares CLI installation can keep profiles for multiple Olares devices or identities.

| Task | Command |
|------|---------|
| List all profiles | `olares-cli profile list` |
| Show the current identity | `olares-cli profile whoami` |
| Switch to another profile | `olares-cli profile use <name>` |
| Switch back to the previous profile | `olares-cli profile use -` |
| Remove a profile and its stored token | `olares-cli profile remove <name>` |

Olares CLI stores tokens securely for the current operating system. Use `profile remove` to clear a profile and its token instead of editing credential files by hand.

## Update Olares CLI and Agent Skills

Check for an update without changing anything:

```bash
olares-cli update --check
```

Update Olares CLI and install the matching Agent Skills in one step:

```bash
olares-cli update
```

To receive every published release, including versions not yet tagged as `latest`, use the `next` channel:

```bash
olares-cli update --channel next
```

Check the installed copy and version after the update:

```bash
olares-cli version
```

## Other installation options

### Install Olares CLI first

To install the CLI without running the setup wizard:

```bash
npm install -g @olares/cli
```

Then install the Agent Skills carried by that binary:

```bash
olares-cli skills install
```

### Run one command without installing

Use `npx` to run a single command without adding Olares CLI to your `PATH`:

```bash
npx @olares/cli@latest files ls /drive/Home
```

Your profile and login state remain available for later commands.

## Using Olares CLI on an Olares host

Olares OS includes a system-managed CLI at `/usr/local/bin/olares-cli`. The system copy maintains Olares OS and is updated through Olares OS releases. Do not replace it with the npm package.

On an Olares host, the setup wizard detects the existing release-grade binary and leaves it in place. If you need the standalone npm version as well, install it under a separate prefix:

```bash
npm install -g @olares/cli@latest --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
olares-cli skills install
```

In this terminal, `olares-cli` now runs the npm copy. For Olares OS maintenance, use the full path `/usr/local/bin/olares-cli` to run the system copy.

:::warning Do not overwrite the system CLI
Do not run `npm install -g @olares/cli --force` on an Olares host. It can overwrite the OS-managed binary and break its version relationship with `olaresd` and the cluster.
:::

## Learn more

- [Olares CLI Agent Skills](./cli-agent-skills.md): Browse the complete skill catalog and learn how skills are installed, discovered, and kept in sync with the CLI.
- [Manage Olares and run research tasks with Lares](../use-cases/lares.md): Use the official Olares AI assistant with Olares CLI and Agent Skills built in.
- [Olares, Olares CLI, and AI Agents: The Relationship Explained](https://www.olares.com/blog/olares-cli-ai-agents-explained/): Learn where agents can run and how profiles connect them to Olares.
