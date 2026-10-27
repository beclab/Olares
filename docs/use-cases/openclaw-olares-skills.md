---
outline: [2, 3]
description: Use the Olares CLI skills in OpenClaw so your agent can manage files and apps on your Olares device.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, Olares CLI skills, ClawHub, agent
app_version: "1.0.17"
doc_version: "1.1"
doc_updated: "2026-07-31"
---

# Manage Olares with your OpenClaw agent

OpenClaw comes with Olares CLI [Agent Skills](/developer/cli-agent-skills.md) built in, so your agent can manage files and applications on your Olares device out of the box. For example, ask it to list files, read logs, or install apps from Olares Market.

## Learning objectives

In this guide, you will learn how to:
- Authenticate the Olares CLI with your Olares ID.
- Chat with your agent in natural language to perform actions on your Olares device, such as installing apps from Market.

## Prerequisites

- OpenClaw installed and running on Olares.
- Your Olares ID and login password.

## Step 1: Authenticate with the Olares CLI

Before your agent can run Olares CLI Agent Skills on your behalf, authenticate the Olares CLI with your Olares ID.

1. Open the OpenClaw CLI from the Launchpad.
2. Run the following command to confirm that both `olares-cli` and its skills are properly installed and enabled:

   ```bash
   olares-cli -v
   ```

   Example output:

   ```
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
5. If two-factor authentication is enabled on your Olares, the CLI prompts you for a two-factor code for this Olares ID. Enter the 6-digit code from LarePass, and then press **Enter**.
6. Run the following command to verify that the profile is created and logged in:

   ```bash
   olares-cli profile list
   ```

   Example output (`*` marks the current profile):

   ```text
      NAME                   OLARES-ID              STATUS     VERSION
   *  laresprime@olares.com  laresprime@olares.com  logged-in  1.12.6
   ```

## Step 2: Direct your agent to execute tasks

Open the Control UI,  start a new session, and send your request to the agent in natural language.

For example, ask it to install an app from Olares Market:

```text
Install Firefox from Olares Market and tell me when it is ready
```

## Learn more

- [Manage skills and plugins](openclaw-skills.md): Install and manage other OpenClaw skills.
