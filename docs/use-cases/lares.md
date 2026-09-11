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

## Core capabilities

Lares scales to your needs, handling everything from quick, everyday commands to complex, multi-step automations.

**Everyday tasks**

For routine system management, Lares acts as a direct, natural language interface to your device:

- **System & App Management:** Inspect device status, check logs, or manage apps from the Market.
- **File Operations:** Read, move, compress, and organize items in Olares Files.
- **App Control:** Command your connected applications, like smart home devices or media players.

**Advanced workflows**

Lares is an autonomous agent, and it isn't limited to one-off commands. It can break down complex goals, chain multiple tools together, and execute end-to-end plans across your system.

Explore the guides below to master these advanced scenarios:
- [Link: Deploy custom apps to Olares]
- [Link: Deploy and publish a website to a custom domain]
- [Link: Enable search for in-depth research]







## How Lares works

- **Seamless integration with Router**: Lares automatically accesses the AI models and tools you have configured in Router. It authenticates using its in-cluster identity, meaning no personal logins or API keys are required.
- **Permission-aware execution**: Lares executes tasks on your behalf using Olares CLI Agent Skills, but you control its scope. You can assign specific permission levels, such as Read Only, Write, or Full Access, to define exactly what it is allowed to perform on your system.
- **One task at a time**: Local AI models share accelerator resources through time slicing and can only process one request at a time.
    - **Queuing in Lares**: Lares handles one task at a time. If a task is running, new requests wait in line and start automatically when the current one finishes.
    - **Running concurrently**: To run Lares and other agents at the same time, connect them to different models.

## Get started

Once you have Router and a connected model set up, open Lares and ask your first question:

```text
I'm new to Olares. Check this device's configuration first.
```

:::tip One request at a time
Lares works on one request at a time. If you send another message before the current task finishes, it waits in line and is handled right after. If you are running another agent at the same time, give it a different model.
:::
