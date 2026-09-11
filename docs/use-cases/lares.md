---
outline: [2, 3]
title: Lares
description: Meet Lares, the official AI assistant for Olares. Manage apps, files, and your system, or dive into research, all through natural language.
head:
  - - meta
    - name: keywords
      content: Olares, Lares, Olares Router, Olares local AI, local AI agent, AI assistant, natural language, Olares 1.12.7
---

# Lares

Lares is the official AI assistant for Olares, introduced in v1.12.7. You state a goal in plain language, and Lares plans and carries out the task on your device: checking system status, installing apps, organizing files, and more. From the first request to the final result, everything happens in one conversation.

## How Lares works

- **Models and tools come from Router**: Lares reaches the AI models and tools you have configured through Router. It authenticates with its in-cluster app identity, so no personal login or API key is required.
- **Powered by Olares CLI Agent Skills**: Every action Lares takes on apps, files, or device states runs through these skills, under your account permissions, so it can only do what you can do yourself.
- **One task at a time**: Local AI models share accelerator resources through time slicing and can only process one request at a time.
    - **Queuing in Lares**: Lares handles one task at a time. If a task is running, new requests wait in line and start automatically when the current one finishes.
    - **Running concurrently**: To run Lares and other agents at the same time, connect them to different models.

## Get started

With Router and a connected model in place, open Lares and ask your first question:

```text
I'm new to Olares. Check this device's configuration first.
```

:::tip One request at a time
Lares works on one request at a time. If you send another message before the current task finishes, it waits in line and is handled right after. If you are running another agent at the same time, give it a different model.
:::

## What Lares can do

**Check on your device.** Inspect system status and configuration, and get plain-language findings and suggested fixes. *"Check this device and flag anything that needs attention."*

**Install and deploy apps.** Install from Market, or hand Lares a GitHub repository and get back a running Olares app. *"Install NocoDB from Market and tell me when it's ready."*

**Work with your files.** Read, organize, archive, and convert files in Olares Files without leaving the conversation. *"Organize my Downloads folder by file type."*

**Operate your apps.** Work through the apps you have installed, from controlling Home Assistant lights, climate, and media players to building automations. *"Turn off all the lights in the living room."*

**Create media.** Generate or transcode images, video, and audio with the apps on your Olares. *"Create a short clip from my latest trip photos."*

For guided walkthroughs on bigger tasks:
