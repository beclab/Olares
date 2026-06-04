---
outline: [2, 3]
description: Set up Pool CLI on Olares to read code, run terminal commands, and edit files using natural language. Connect via Poolside cloud API or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Pool CLI, AI coding, terminal, TUI, self-hosted, MCP
app_version: "0.1.0"
doc_version: "1.0"
doc_updated: "2026-06-04"
---

# Code with Pool CLI

Pool CLI is a terminal-based coding agent that helps you read code, run terminal commands, and edit files using natural language. On Olares, this command-line interface runs inside a browser-based terminal equipped with a pre-configured Ubuntu development environment.

## Learning objectives

In this guide, you will learn how to:

- Install Pool CLI from the Olares Market.
- Connect Pool CLI to a model using Poolside's cloud API or a local model.
- Execute a basic natural language coding workflow.
- Configure directory access for your development workspace.
- Manage software dependencies securely.

## Prerequisites

- An Olares device with sufficient disk space and memory.
- A Poolside account, if you plan to use the cloud API service.
- A local model optimized for coding running on your Olares device, if you plan to run tasks locally.

   You can install local models using one of the following methods:
   - **Single-model application**: One app that runs one specific model. This guide uses **Qwen3-Coder 30B (Ollama)**.
   - **Ollama application**: One app that hosts multiple models. Ensure [Ollama](ollama.md) is installed with at least one model downloaded, such as `qwen3-coder:30b`.

## Install Pool CLI

1. Open Market, and search for "Pool CLI".

   ![Pool CLI](/images/manual/use-cases/pool-cli.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Connect to a model

Start the Pool CLI and connect it to a language model. Choose one of the following connection methods.

### Connect using Poolside cloud service

Use this method to leverage Poolside's cloud-based inference API.

1. Open the Pool CLI from the Launchpad.
2. Enter the following command to trigger the login authentication:

   ```bash
   pool setup
   ```

3. Select **Log in with Poolside**, and then enter your Poolside API key to authenticate.
4. Choose one of the following modes to run your tasks:

   <Tabs>
   <template #Interactive-mode>

   Use this method when you want a continuous, chat-like conversation with the agent. This is ideal for multi-step tasks where the agent needs to read context and ask you for follow-up clarifications.

   1. Enter the following command to start an interactive session:

      ```bash
      pool
      ```
   2. Interact with the agent using natural language.
   3. Enter the following command to exit the session:

      ```bash
      /quit
      ```
   </template>
   <template #Automated-mode>
   
   Use this method to run a single task and immediately return to your normal terminal. This is ideal for quick requests that do not require a back-and-forth conversation.

   Enter the `pool exec` command to send a single prompt and exit. For example:

      ```bash
      pool exec -p "Create a folder named Test"
      ```

   :::tip
   By default, the agent pauses and asks you to manually approve any system actions, such as writing files. To bypass this manual check and allow the agent to execute actions instantly, append the `--unsafe-auto-allow` flag. For example, `pool exec -p "Create a folder named Test" --unsafe-auto-allow`.
   :::
   </template>
   </Tabs>

### Connect using a local model

Use this method to run Pool CLI entirely offline with a local model. This example uses the model app **Qwen3-Coder 30B (Ollama)**.

1. Install the model app **Qwen3-Coder 30B (Ollama)** from Market.

   ![Qwen3-Coder 30B (Ollama)](/images/manual/use-cases/qwen3-coder-30b.png#bordered)

2. Open the model app from the Launchpad and wait for the download to complete.
3. Note down the exact model name displayed on the page. For example, `qwen3-coder:30b`.

   ![Model name on the model app page](/images/manual/use-cases/qwen3-coder-model-name.png#bordered){width=50%}

4. Open Settings, and then go to **Applications** > **Qwen3-Coder 30B (Ollama)** > **Shared entrances**.

   ![Model app endpoint in Settings](/images/manual/use-cases/qwen3-coder-30b-endpoint.png#bordered){width=70%}

5. Click **Qwen3-Coder 30B**, and then note down the endpoint URL. For example, `http://609c5d0c0.shared.olares.com`.
6. Go to **Applications** > **Pool CLI** > **Manage environment variables**, and then click <i class="material-symbols-outlined">edit</i> to configure the following variables:

   - **USE_LOCAL_LLM**: Set it to `true` to enable local model mode.
   - **POOLSIDE_STANDALONE_BASE_URL**: Enter the model app's endpoint URL with `/v1` appended. For example, `http://609c5d0c0.shared.olares.com/v1`.
   - **POOL_MODEL**: Enter the model name you noted down earlier. For example, `qwen3-coder:30b`.

7. Click **Apply**. Wait for the Pool CLI container to restart.
8. Open the Pool CLI from the Launchpad, and then enter the following command to start a session.

   ```bash
   pool
   ```

## Code with natural language

After connecting to a model, you interact with Pool CLI using conversational prompts. The agent interprets your requests to write code, modify files, and execute terminal commands.

The following scenario demonstrates how to use the Pool CLI to generate and run a simple Python script.

1. Open the Pool CLI from the Launchpad.
2. Enter the following command to start an interactive session:

   ```bash
   pool
   ```

3. Enter a natural language request. For example:

   ```text
   Create a Python script named greeting.py that outputs the 
   current date and time
   ```

4. Review the agent's proposed code and actions. Pool CLI generates the script and asks for permissions to proceed.
5. Select to allow the operations. The terminal displays the output of your script.

   ![Pool CLI code result](/images/manual/use-cases/pool-cli-results.png#bordered)

6. To exit the interactive session, enter the following command:

   ```bash
   /quit
   ```

7. To verify the output, open Files, and then go to **Data** > **pool** > **home** > **work**.

   ![Pool CLI result verify](/images/manual/use-cases/pool-cli-results-verify.png#bordered)

## Manage the development environment

Pool CLI operates within a pre-configured Ubuntu 24.04 environment. Customize your directory access and install additional tools based on your project requirements.

### Manage directory access

By default, all project work happens in the `/opt/data` directory. This directory persists your files across app restarts and is located at **Files** > **Data** > **pool** > **home** > **work** on Olares.

If you want Pool CLI to access files in your **Home** or **External** directories, configure the environment variables:

1. Open Settings, and then go to **Applications** > **Pool CLI** > **Manage environment variables**.
2. Specify the following variables as needed:

   - **ALLOW_HOME_DIR_ACCESS**: Set to `true` to allow access to the **Home** directory in Files. This mounts the **Home** directory at `/home/userdata/home/`.
   - **ALLOW_EXTERNAL_DIR_ACCESS**: Set to `true` to allow access to the **External** directory, such as mounted NAS or USB drives. This mounts the **External** directory at `/home/userdata/external/`.

3. Click **Apply**.

### Review pre-installed development tools

Before you install additional software, review the tools already included in your workspace. The container image is based on Ubuntu 24.04 and comes with many common development tools pre-installed.

The following table lists the key categories and examples.

| Category | Included tools |
|:---------|:---------------|
| Languages and runtimes | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3,<br>Lua, Perl, SQLite |
| Build tools | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`,<br>common `-dev` headers |
| CLI utilities | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`,<br>`zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| Database clients | `postgresql-client`, `mysql-client`, `redis-tools` |

### Install additional software

If your project requires tools or libraries beyond the pre-installed ones, you must manage them within the container's security boundaries.

#### What you cannot install yourself

If your project requires a system-level library (e.g., `libpq-dev`, `ffmpeg`, `libssl-dev`), you cannot install it directly. These dependencies must be added to the base container image by the application maintainer.

#### What you can install in your workspace

Inside your writable directories (primarily `/opt/data`), you can install project-level dependencies without root privileges using common tools:

- **Python**: Create a virtual environment and use `pip`. For example:

   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install <package>
   ```

- **Node.js**: Use `npm` inside your project folder. For example:

   ```bash
   npm install <package>
   ```

- **Rust/Go** (or other compiled languages): Install binaries to a user-writable path. For example:

   ```bash
   cargo install --root ~/.local <package>   # Rust
   go install <package>@latest               # Go (installs to ~/go/bin)
   ```

## FAQs

### How to switch between cloud and local models?

- To switch from cloud mode to local mode, set `USE_LOCAL_LLM` to `true` and configure `POOLSIDE_STANDALONE_BASE_URL` and `POOL_MODEL` in the environment variables, then restart the app.
- To switch from local mode to cloud mode, set `USE_LOCAL_LLM` to `false` and restart the app. You might need to run `pool setup` again to re-authenticate.

### Missing language or library

Determine if the missing tool is a system-level dependency:

- System-level dependencies: You cannot install these yourself. The app maintainers must add them to the base image. If you need a system-level library that is not currently available, [submit a GitHub Issue](https://github.com/beclab/apps/issues) to request it.
- User-level dependencies: Use `venv`, `npm install`, or similar local tools to install them.

## Learn more

- [Poolside documentation](https://docs.poolside.ai/get-started/overview)
- [Write code using Claude Code](claude-code.md)
