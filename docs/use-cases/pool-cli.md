---
outline: [2, 3]
description: Set up Pool CLI on Olares to read code, run terminal commands, and edit files using natural language. Connect via Poolside cloud API or a local model.
head:
  - - meta
    - name: keywords
      content: Olares, Pool CLI, Poolside, AI coding, terminal, TUI, self-hosted, MCP
app_version: "0.1.0"
doc_version: "1.0"
doc_updated: "2026-06-03"
---

# Code with Pool CLI

Pool CLI is Poolside's terminal-based coding agent that helps you read code, run terminal commands, and edit files using natural language. On Olares, this command-line interface runs inside a browser-based terminal equipped with a pre-configured Ubuntu development environment.

## Learning objectives

In this guide, you will learn how to:

- Install the Pool CLI app from the Olares Market.
- Connect Pool CLI to a model using Poolside's cloud API or a local model.
- Execute basic and advanced natural language coding workflows.
- Manage software dependencies securely.

## Prerequisites

- An Olares device with sufficient disk space and memory.
- A Poolside account, if you plan to use the cloud API service.
- A local model optimized for coding running on your Olares device, if you plan to use local execution.

   You can install local models using one of the following methods:
   - **Ollama application**: One app that hosts multiple models. Ensure [Ollama](ollama.md) is installed with at least one model downloaded, such as `qwen3-coder:30b`.
   - **Single-model application**: Runs one specific model as a standalone application. This guide uses **Qwen3-Coder 30B (Ollama)**.

## Install Pool CLI

1. Open Market, and search for "Pool CLI".

   ![Pool CLI](/images/manual/use-cases/pool-cli.png#bordered)

2. Click **Get**, and then click **Install**. Wait for the installation to finish.

## Connect to a model

Start the Pool CLI and connect it to a language model. Choose one of the following connection methods.

### Connect using Poolside cloud service

Use this method to leverage Poolside's cloud-based inference API.

1. Open the Pool CLI from the Launchpad.
2. Enter the following command to enter interactive mode:

   ```bash
   pool
   ```

3. On first launch, the CLI triggers a login prompt. You can also run `pool setup` to initiate login manually.

4. Enter your Poolside api-key to authenticate.
5. Use `/quit` to exit interactive mode.

You can also execute one-shot tasks directly from the terminal. For example:

```bash
pool exec -p "say OK" --unsafe-auto-allow
```

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
6. Go to **Applications** > **Pool CLI** > **Manage environment variables**, and then specify the following environment variables:

   - **USE_LOCAL_LLM**: Set to `true` to enable local model mode.
   - **POOLSIDE_STANDALONE_BASE_URL**: Enter the model app's endpoint URL with `/v1` appended. For example, `http://609c5d0c0.shared.olares.com/v1`.
   - **POOL_MODEL**: Enter the model name you noted down earlier. For example, `qwen3-coder:30b`.

7. Click **Apply**. Wait for the Pool CLI container to restart.
8. Open the Pool CLI from the Launchpad, and then enter `pool` to start a session (no login required in local mode).

## Manage the development environments

### Default workspace
All project work happens in the `/opt/data` directory, which serves as the working directory in the container. This directory persists your files across app restarts and is located at **Data** > **pool** > **home** > **work**.


### Access Home and External directories

By default, Pool CLI operates within `/opt/data`. If you want Pool CLI to access files in the **Home** or **External** directories on Olares, configure the following environment variables in **Applications** > **Pool CLI** > **Manage environment variables**:

- **ALLOW_HOME_DIR_ACCESS**: Set to `true` to allow access to the Home directory in the Files app. This mounts the Home directory at `/home/userdata/home/`.
- **ALLOW_EXTERNAL_DIR_ACCESS**: Set to `true` to allow access to the External directory (mounted NAS or other external disk data). This mounts the External directory at `/home/userdata/external/`.

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

### How do I switch between cloud and local models?

To switch from cloud mode to local mode, set `USE_LOCAL_LLM` to `true` and configure `POOLSIDE_STANDALONE_BASE_URL` and `POOL_MODEL` in the environment variables, then restart the app.

To switch back to cloud mode, set `USE_LOCAL_LLM` to `false` and restart the app. You may need to run `pool setup` again to re-authenticate.

### Missing language or library

Determine if the missing tool is a system-level dependency:

- System-level dependencies: You cannot install these yourself. The app maintainers must add them to the base image. If you need a system-level library that is not currently available, [submit a GitHub Issue](https://github.com/beclab/apps/issues) to request it.
- User-level dependencies: Use `venv`, `npm install`, or similar local tools to install them.
