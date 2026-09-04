---
outline: [2, 3]
description: Learn how to use Lares and Olares CLI Agent Skills to manage Olares through natural language.
head:
  - - meta
    - name: keywords
      content: Olares One, Lares, Router, Agent Skills, natural language, AI agent, Olares CLI
---

# Manage Olares through natural language <Badge type="tip" text="30 min" />

Lares is Olares' built-in AI assistant. With a local model and Router, it can understand your text or voice requests and turn them into real device management actions through Olares CLI Agent Skills. For example, you can ask Lares to check system status, install apps, manage files, or troubleshoot issues.

This guide walks you through your first Lares session. You will check whether your environment is ready, start a conversation, and try a few common tasks.

:::info
Lares is the recommended entry point, but it is not the only way to use Olares CLI Agent Skills. You can also use the same skills from other agent apps on Olares, or from a local agent such as Codex or Cursor, to manage your Olares device through natural language.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:

- Check whether your environment is ready for Lares.
- Start your first conversation with Lares.
- Use natural language in Lares to manage Olares.

## Prerequisites

- **System**: Olares OS v1.12.7 or later.
- **AI components**: Lares, Router, and a usable model. The model can come from a local model app or from a provider configured in Router.
- **User permissions**: Admin privileges to install shared apps from Market.

## Step 1: Prepare your environment

What you need to do first depends on how your Olares device was installed or upgraded.

| Starting point | What's preinstalled | Next step |
| --- | --- | --- |
| Olares One v1.12.7 factory image | Lares, Router, and Qwen3.8-27B | Open Lares and start. |
| Upgraded or self-hosted Olares v1.12.7 | Router only. | Install Lares, then install a model app or add a provider in Router. |

## Step 2: Start your first Lares conversation

1. Open Lares from the Launchpad.
2. Select the workspace you want Lares to use.
3. Grant the permissions Lares needs for the tasks you want it to perform.
4. Check that the model is selected and available.

    ![Lares chat interface](/images/one/lares-chat.png#bordered)

5. Send your first `Hello`. Once your first message goes through, you are ready to manage Olares by chatting with Lares.

    ![Lares chat response](/images/one/lares-chat-response.png#bordered)

## Step 3: Try common tasks

The following examples cover common scenarios.

### Check your device configuration

Start with a basic question:

```text
I'm new to Olares. Check this device's configuration first.
```

![Check device configuration in Lares](/images/one/onboard-scenario-question.png#bordered)

### Install an app from Market

Ask Lares to install an app for you:

```text
Install Code Server from the Olares Market and tell me when it's ready.
```

![Install an app in Lares](/images/one/onboard-scenario-install2.png#bordered)

### Deploy an app to Olares

For a more advanced task, ask Lares to deploy a project from a GitHub repository. The example below uses "Wealthfolio", a finance app.

```text
Deploy this app to Olares: https://github.com/wealthfolio/wealthfolio
and make sure it has a desktop icon.
```

Lares will inspect the source app, prepare the Olares app chart, and update the required manifest files. Depending on the app, this might take a few minutes. When it finishes, Lares will tell you how to verify the result.

![Deploy an app in Lares](/images/one/onboard-scenario-porting1.png#bordered)

You can then find the app on the Launchpad and in My Olares.

![Deployed app in Lares](/images/one/onboard-scenario-ported1.png#bordered)

## Resources

- [Install and use Agent Skills](../developer/cli-agent-skills.md)
- [Explore Olares use cases](../use-cases/index.md)
