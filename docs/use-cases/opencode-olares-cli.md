---
outline: [2, 3]
description: Use the Olares CLI skills in OpenCode so your coding agent can manage files and apps on your Olares device.
head:
  - - meta
    - name: keywords
    - content: Olares, OpenCode, Olares CLI skills, AI agent, natural language, file management, app installation
---

# Manage Olares with Olares CLI

OpenCode comes with Olares CLI [Agent Skills](/developer/cli-agent-skills.md) built in, so your agent can manage files and applications on your Olares device through natural language. For example, ask it to list files, read logs, or install apps from Olares Market.

## Prerequisites

- OpenCode installed and running on Olares, with a model connected.
- Your Olares ID and login password.

## Step 1: Authenticate the Olares CLI with your Olares ID

Before OpenCode can run Olares CLI Agent Skills on your behalf, authenticate the Olares CLI with your Olares ID.

1. From the Launchpad, click OpenCode Terminal. It opens straight into a command line.
2. Run the following command to confirm that `olares-cli` is properly installed:

   ```bash
   olares-cli -v
   ```

   Example output:

   ```text
   olares-cli version 1.12.6
   Git commit: d30eca705df2fb614bf2bbea95daa2e6998adeeb
   Build time: 2026-07-06T06:33:00Z
   ```

3. Run the following command to log in to your Olares account. Replace `<your-olares-id>` with your actual Olares ID.

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   Example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

4. When prompted, enter your Olares login password. The password is hidden as you type.
5. If two-factor authentication is enabled on your Olares account, the CLI prompts you for a two-factor code for this Olares ID. Enter the 6-digit code from LarePass, and then press **Enter**.
6. Run the following command to verify that the profile is created and logged in:

   ```bash
   olares-cli profile list
   ```

   Example output (`*` marks the current profile):

   ```text
      NAME                   OLARES-ID              STATUS     VERSION
   *  laresprime@olares.com  laresprime@olares.com  logged-in  1.12.6
   ```

   :::info
   This login keeps OpenCode authenticated for up to 30 days. When it expires, you will need to log in once more.
   :::

## Step 2: Direct your agent to execute tasks

1. From the Launchpad, click OpenCode.
2. Start a new session, and select the model you connected earlier.
3. Send your request to the agent in natural language. 

    For example, ask it to install an app from Olares Market:

   ```text
   Install Firefox from Olares Market and tell me when it is ready
   ```

    You can also ask it to manage files, read logs, check system status, and more.

## Learn more

- [Set up OpenCode as your AI coding agent](opencode.md): Install OpenCode and connect it to a model.
- [Olares CLI and AI agents, explained](https://www.olares.com/blog/olares-cli-ai-agents-explained): How agents use the Olares CLI to manage your system.
