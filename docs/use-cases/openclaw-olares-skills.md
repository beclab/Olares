---
outline: [2, 3]
description: Use the Olares CLI skills in OpenClaw to manage files and apps on your Olares device through natural language.
head:
  - - meta
    - name: keywords
      content: Olares, OpenClaw, Olares CLI skills, ClawHub, natural language
app_version: "1.0.2"
doc_version: "1.0"
doc_updated: "2026-05-28"
---

# Manage Olares through natural language

Install the Olares CLI skills in OpenClaw to manage files and applications on your Olares device through natural language. For example, ask your agent to list files, read logs, or install apps from Olares Market.

## Learning objectives

In this guide, you will learn how to:
- Authenticate with the Olares CLI in the OpenClaw CLI.
- Install Olares skills from ClawHub.
- Use natural language to manage files and apps on your Olares device.

## Prerequisites

- OpenClaw installed and running on Olares.
- Your Olares ID and login password.

## Step 1: Authenticate with the Olares CLI

To authorize your agent to perform system actions, you must first log in to the Olares CLI using your account credentials.

1. Open the OpenClaw CLI from the Launchpad.
2. Log in to your Olares account. Replace `<your-olares-id>` with your Olares ID:

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   For example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

3. Follow the prompt to enter your Olares login password. For security, the password you enter is hidden.
4. Verify your login status.

   ```bash
   olares-cli profile list
   ```

   Example output:

   ```text
   NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com   logged-in
   ```

## Step 2: Install Olares skills

Give your agent the ability to manage your device by installing Olares skills from ClawHub.

1. Open the Control UI, and then select **Skills** from the left sidebar.
2. Under **ClawHub**, in the search box, enter `olares` to find the Olares skills.

   ![Olares skills in ClawHub](/images/manual/use-cases/openclaw-install-olares-skills1.png#bordered)

3. Install the **Olares Shared** skill first because it's the foundation for the other Olares skills.
4. Install the remaining Olares skills, such as **Olares Files** and **Olares Market**.
5. Go to the chat page, and then run `/reset` to start a new session so the agent picks up the newly installed skills. If you've configured channels such as Discord, also run `/reset` in each channel conversation.

:::info Retry on 429 errors
If you see a 429 error when downloading a skill, wait a moment and try again.
:::

## Step 3: Chat with the agent in natural language

Open the Control UI and send your request to the agent in natural language.

For example, ask it to install an app from Olares Market:

```text
Install Firefox
```

![Install an app](/images/manual/use-cases/openclaw-olares-cli-install-app.png#bordered)

## Learn more

- [Manage skills and plugins](openclaw-skills.md): Install and manage other OpenClaw skills.
