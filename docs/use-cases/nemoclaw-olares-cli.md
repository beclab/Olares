---
outline: [2, 3]
description: Use the Olares CLI skills in NemoClaw to manage files and apps on your Olares device through natural language.
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, Olares CLI, ClawHub, skills, natural language, file management, app installation
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-05-11"
---

# Manage Olares with Olares CLI

The Olares CLI skills let your NemoClaw agent manage files and apps on your Olares device through natural language. For other skills, see [Manage skills and plugins](openclaw-skills.md).

## Prerequisites

- NemoClaw installed and running on Olares.
- Your Olares ID and login password.

## Step 1: Log in to Olares CLI

Olares CLI requires your account password. Sign in from the NemoClaw CLI before the agent can use the Olares CLI skills.

1. Open the NemoClaw CLI app from Launchpad.
2. Connect to the runtime sandbox:

   ```bash
   nemoclaw my-assistant connect
   ```

   Wait until the terminal shows the sandbox prompt, such as `sandbox@my-assistant:~$`.

3. Log in to your Olares account. Replace `<your-olares-id>` with your Olares ID:

   ```bash
   olares-cli profile login --olares-id <your-olares-id>
   ```

   For example:

   ```bash
   olares-cli profile login --olares-id laresprime@olares.com
   ```

   Follow the prompts to enter your Olares login password.

4. Verify you have been logged in with the profile.

   ```bash
   olares-cli profile list
   ```

   Example output:

   ```text
   NAME                   OLARES-ID              STATUS
   *  laresprime@olares.com  laresprime@olares.com   logged-in (23h59m)
   ```

## Step 2: Install Olares skills from ClawHub

1. Open the OpenClaw Web UI and go to **Skills**.
2. In the ClawHub search box, enter `olares` to find Olares skills.

   ![Olares skills in ClawHub](/images/manual/use-cases/nemoclaw-install-olares-skills.png#bordered)

3. Install the **Olares Shared** skill first because it's the foundation of the other Olares skills.
4. Install the remaining Olares skills, such as **Olares Files** and **Olares Market**.
5. Open the chat page in the OpenClaw Web UI and run `/new` to start a new session so the agent picks up the newly installed skills. If you've configured channels such as Discord, also run `/new` in each channel conversation.

:::info Retry on 429 errors
If you see a 429 error when downloading a skill, wait a moment and try again.
:::

## Step 3: Chat with the agent in natural language

Open the OpenClaw Web UI or the OpenClaw TUI and ask the agent in natural language. For example:

- To list all files and folders under `/drive/Home/`:

  ```text
  List drive/Home/
  ```

  ![List files result](/images/manual/use-cases/nemoclaw-openclaw-olares-cli-list-files.png#bordered)

- To read a file:

  ```text
  Read the last 10 lines of the nemoclaw.log file in the Home directory.
  ```

  ![Read file result](/images/manual/use-cases/nemoclaw-openclaw-olares-cli-read-file.png#bordered)

- To install an app from Olares Market:

  ```text
  Install Firefox
  ```

- To uninstall an app from Olares Market:

  ```text
  Uninstall Firefox
  ```

## Learn more

- [Run NemoClaw with a local LLM](nemoclaw.md): Set up NemoClaw with a local model.
- [Manage skills and plugins](openclaw-skills.md): Install and manage other OpenClaw skills.
