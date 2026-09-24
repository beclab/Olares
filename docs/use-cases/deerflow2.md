---
connectionVersion: "1.12.7"
connectionLatestPath: /use-cases/deerflow2
outline: deep
title: Set up DeerFlow 2.0 for AI research
description: Set up DeerFlow 2.0 on your Olares device and configure it with a local model app for deep research.
head:
  - - meta
    - name: keywords
      content: Olares, DeerFlow, AI agent, deep research, multi-agent, self-hosted, LLM
doc_version: "1.2"
app_version: "1.0.6"
doc_updated: "2026-09-23"
---

# Set up DeerFlow 2.0 for AI-powered research and tasks

<VersionRouteSelect />

DeerFlow is an open-source agent harness by ByteDance, built on LangGraph and LangChain. It orchestrates sub-agents, memory, and sandboxes to handle complex tasks through extensible skills.

DeerFlow 2.0 is a ground-up rewrite of the original [DeerFlow](https://github.com/bytedance/deer-flow). While version 1.0 was a deep research framework, version 2.0 is a general-purpose agent platform.

## Learning objectives

In this guide, you will learn how to:
- Install DeerFlow 2.0 on Olares and configure it with a local model.
- Run tasks such as deep research.

## Prerequisites

Before you begin, you need:

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- The following model:

  | Model type | Model | How to get it |
  | :--- | :--- | :--- |
  | Chat | Qwen3.8-27B (llama.cpp) | Install from Market |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## Install DeerFlow 2.0

1. Open Market and search for "DeerFlow 2.0".
   ![DeerFlow 2.0](/images/manual/use-cases/deerflow2.png#bordered)

2. Click **Get**, then click **Install**, and wait for the installation to complete.

## Configure the model

DeerFlow 2.0 uses a `config.yaml` file for its core configuration. To connect it to a local model, add the model and its connection details to this file.

### Get model connection details

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

For Qwen3.8-27B (llama.cpp):

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

### Edit config.yaml

1. Open Files and navigate to the DeerFlow 2.0 app data directory: `/Data/deerflowv2/config/`.
2. Open `config.yaml`, and click <span class="material-symbols-outlined">edit_square</span> in the top-right corner to open the editor.
3. Under the `models:` section, add the model configuration below. Replace `PASTE_ROUTER_BASE_URL` with the Base URL copied from Router.

   ```yaml
   models:
     - name: default-chat      # Unique identifier for the model
       display_name: Qwen3.8-27B      # Name shown in the UI
       use: langchain_openai:ChatOpenAI      # LangChain class for OpenAI-compatible APIs
       model: default-chat      # Model ID
       api_key: olares      # Use any non-empty text
       base_url: PASTE_ROUTER_BASE_URL
       supports_thinking: true      # Set to true if the model supports extended thinking
   ```

   <!--
   TODO: 素材清单 04，待补 Router 截图：deerflow2-config-router.png；替换下方旧图后再取消注释。
   ![Edit config.yaml](/images/manual/use-cases/deerflow2-edit-config-yaml1.png#bordered)
   -->


4. Click <span class="material-symbols-outlined">save</span> to save the changes.

### Restart to apply changes

1. Open Control Hub and select the DeerFlow 2.0 project.
2. Under **Deployments**, locate the backend container and click **Restart**.

   ![Restart DeerFlow 2.0](/images/manual/use-cases/deerflow2-restart.png#bordered)

3. In the confirmation dialog, confirm the restart.
4. Wait for the status icon to turn green.

## Use DeerFlow 2.0

Once the model is configured, you can start using DeerFlow 2.0.

1. Open DeerFlow 2.0 from Launchpad and click **Get Started with 2.0**.

2. On the first launch, create the administrator account used to access and manage the current DeerFlow 2.0 instance:
    - **Email:** Enter the email address for the administrator account.
    - **Password:** Enter a password with at least eight characters.
    - **Confirm Password:** Enter the password again.

   Account data, including login details, stays within the current DeerFlow 2.0 instance. Account creation and sign-in do not connect to an external account service.

   Click **Create Admin Account** to access the chat interface.

   ![Create the DeerFlow administrator account](/images/manual/use-cases/deerflow2-create-admin-account.png#bordered)

3. Select your preferred execution mode.

   ![Select execution mode](/images/manual/use-cases/deerflow2-select-mode.png#bordered)

   DeerFlow 2.0 offers several execution modes that control how the agent processes your request, from quick single-pass answers to multi-step research with sub-agents.

4. Enter your prompt in the chat box, or select a suggested topic for inspiration.

   For example, you can conduct deep research on a topic:
   ![Deep research example](/images/manual/use-cases/deerflow2-research.png#bordered)

   You can also upload attachments and ask DeerFlow to use them as input:
   ![Upload attachments](/images/manual/use-cases/deerflow2-write.png#bordered)

## FAQs

### DeerFlow 2.0 does not generate a response

If the agent fails to start or hangs:

- **Check model compatibility**: Ensure the model you selected is properly configured in `config.yaml`. Verify the endpoint URL is correct.
- **Check connection details**: Use `default-chat` as the model and the Base URL copied from Router. Check the default chat model on **Default models**.

### How do I enable follow-up suggestions?

By default, follow-up suggestions are turned off in DeerFlow 2.0 on Olares to reduce unnecessary GPU usage after a response is generated.

To enable it:

1. Open Control Hub and select the DeerFlow 2.0 project.
2. Under **Deployments**, click the **deerflowv2-frontend** deployment.
3. Click <span class="material-symbols-outlined">edit_square</span> to edit the YAML.
4. Find the `ENABLE_FOLLOWUP_SUGGESTIONS` environment variable and change its value to `'true'`.
   ![Enable follow-up suggestions](/images/manual/use-cases/deerflow2-enable-followup-suggestions.png#bordered)

5. Click **Confirm** to apply the changes.
