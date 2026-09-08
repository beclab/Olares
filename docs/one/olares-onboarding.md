---
outline: [2, 3]
description: Learn how to use Lares and Olares CLI Agent Skills to manage Olares through natural language.
head:
  - - meta
    - name: keywords
      content: Olares One, Lares, Router, Agent Skills, natural language, AI agent, Olares CLI
---

# Manage Olares through natural language <Badge type="tip" text="30 min" />

Lares is Olares' built-in AI assistant. With Router and a connected model, it can understand your requests in natural language and turn them into real device management actions through Olares CLI Agent Skills. For example, you can ask Lares to check system status, install apps, manage files, or troubleshoot issues.

This guide walks you through your first Lares session. You will check whether your environment is ready, start a conversation, and try a few common tasks.

:::info
Lares is the recommended entry point, but it is not the only way to use Olares CLI Agent Skills. You can also use the same skills from other agent apps on Olares, or from a local agent such as Codex or Cursor, to manage your Olares device through natural language.
:::

## Learning objectives

By the end of this tutorial, you will learn how to:

- Prepare your environment for Lares.
- Start your first conversation with Lares.
- Manage Olares through natural language.

## Prerequisites

- **System**: Olares OS v1.12.7 or later.
- **AI components**: Lares, Router, and a usable model. The model can come from a local model app or from a provider configured in Router.
- **User permissions**: Admin privileges to install shared apps from Market.

## Step 1: Prepare your environment

Preparation depends on your starting point. Find yours in the following table.

| Starting point | <nobr>What's preinstalled</nobr> | Next step |
| --- | --- | --- |
| Olares One v1.12.7 factory image<br>(new device) | <ul><li>Lares</li><li>Router</li><li>Qwen3.8-27B</li></ul> | Open Lares and start. |
| <ul><li>Self-hosted Olares v1.12.7<br>(fresh install or upgrade)</li> <li><nobr>Olares One upgraded to v1.12.7</nobr></li></ul> | None. | Install Router and Lares, then install a model app or add a provider in Router. |

## Step 2: Start your first Lares conversation

1. Open Lares from the Launchpad.
2. Keep the default workspace, or select another one.
3. Keep the default write permission, or choose read-only or full access.
4. Check that the model is selected.

    ![Lares chat interface](/images/one/lares-chat.png#bordered)

5. Send your first `Hello`. Once the message goes through, you are ready to manage Olares by chatting with Lares.

    ![Lares chat response](/images/one/lares-chat-response.png#bordered)

## Step 3: Try common tasks

Now that you're up and running, try these common tasks, from a quick status check to a full app deployment.

### Check your device configuration

Start with a basic question:

```text
I'm new to Olares. Check this device's configuration first.
```

![Check device configuration in Lares](/images/one/onboard-scenario-question1.png#bordered)

### Install an app from Market

Ask Lares to install an app for you:

```text
Install NocoDB from the Olares Market and tell me when it's ready.
```

![Install an app in Lares](/images/one/onboard-scenario-install3.png#bordered)

### Deploy an app to Olares

For a more advanced task, ask Lares to deploy a project from a GitHub repository. The example below uses "Wealthfolio", a finance app.

```text
Deploy this app to Olares: https://github.com/wealthfolio/wealthfolio
and make sure it has a desktop icon.
```

Lares will inspect the source app, prepare the Olares app chart, and update the required manifest files. Depending on the app, this might take a few minutes. When it finishes, Lares will tell you how to verify the result.

![Deploy an app in Lares](/images/one/onboard-scenario-porting2.png#bordered)

You can then find the app on the Launchpad and in My Olares.

![Deployed app in Lares](/images/one/onboard-scenario-ported1.png#bordered)

## Resources

- [Install and use Agent Skills](../developer/cli-agent-skills.md): Details about the Olares CLI skill bundles.
- [Manage accelerator resources](../manual/olares/settings/gpu-resource.md): Learn how to check GPU usage, switch GPU modes, and release accelerator resources.
