---
outline: [2, 3]
description: Install and update olares-cli with npm or npx. Covers the first-run wizard, a persistent install, one-off runs, and the special case of a machine that already runs Olares OS.
---

# Install olares-cli

Olares ships `olares-cli` at `/usr/local/bin/olares-cli` on the host. To use the CLI on another machine, or to use the user-mode agent commands that the host bundle may not include yet, install the standalone CLI with npm.

:::info
Apps that run inside Olares, such as OpenClaw, ship with `olares-cli` preinstalled in their container image. You don't need to install the CLI manually for in-cluster use.
:::

## Choose an install method

Pick the path that fits how you work.

### Set up the CLI and Agent Skills

Use this path when you plan to control Olares from an AI agent. 

Open your terminal and run the following command to install the CLI and the Agent Skills with an interactive wizard:

```bash
npx @olares/cli@latest install
```

Example output:

```bash
┌  Setting up Olares CLI...
│
◇  Installed globally
│
◇  Skills installed
│
└  You are all set!

Next:
  olares-cli profile login --olares-id <your-olares-id>   # authenticate (browser/password + optional TOTP)
  olares-cli profile current                              # verify

Then tell your AI agent: "Load the olares-shared skill, then use olares-cli to ..."
```

:::info
The wizard runs `npm install -g @olares/cli` and then installs the Agent Skills. It does not install Olares OS, and it does not log you in.
:::

### Install the CLI only

Use this path for a persistent CLI when you plan to install the Agent Skills later.

Open your terminal and run the following command:

```bash
npm install -g @olares/cli
```

To add the Agent Skills using `npx skills`, see [Install and use Agent Skills](./cli-agent-skills.md).

### Run without installing

Use this path to run a single command without a persistent install.

```bash
npx @olares/cli files ls /drive/Home
```

:::info The npm CLI acts as an Olares user
The CLI you install with npm or run with npx works on behalf of an Olares user. It can manage files, apps, settings, and the cluster on a running Olares, but it can't install or maintain Olares OS itself. Host commands such as `upgrade`, `node`, `gpu`, and `disk` run only from the CLI bundled with Olares OS at `/usr/local/bin/olares-cli`.
:::

## Update olares-cli

To update an npm-installed CLI to the latest release, run one of the following commands.

Install the latest version explicitly:

```bash
npm install -g @olares/cli@latest
```

Or use npm's built-in update command:

```bash
npm update -g @olares/cli
```

Verify the update:

```bash
olares-cli --version
```

The output shows the installed version. The exact version depends on the release you have installed.

:::info
Updating the CLI does not update the Agent Skills. To update the skills after updating the CLI, see [Update the skills](./cli-agent-skills.md#update-the-skills).
:::

## Special case: a Linux host running Olares OS

Olares OS runs on Linux and ships a bundled CLI at `/usr/local/bin/olares-cli`. On such a host, installing the npm CLI can conflict with this bundled binary. On macOS and Windows, use the [methods above](#choose-an-install-method).

On a Linux Olares host, `npm install -g @olares/cli` exits with `EEXIST`. npm refuses to overwrite a binary it does not manage, so the system `olares-cli` remains unchanged. To install the npm copy in parallel, use a separate prefix:

```bash
npm install -g @olares/cli --prefix ~/.olares-cli-npm
export PATH="$HOME/.olares-cli-npm/bin:$PATH"
```

:::warning Do not overwrite the system `olares-cli`
Don't run `npm install -g @olares/cli --force` on an Olares host. That overwrites the OS-managed `/usr/local/bin/olares-cli`, which breaks the version chain with `olaresd` and the cluster. The OS bundle is updated only through `olares-cli upgrade`.
:::

## Next step

[Log in to Olares](./cli-log-in.md) to create a profile, then drive Olares from your agent.
