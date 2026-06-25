---
outline: [2, 3]
description: 在 Olares 上安装 Pool CLI，通过自然语言读取代码、运行终端命令和编辑文件。支持连接 Poolside 云端 API 或本地模型。
head:
  - - meta
    - name: keywords
      content: Olares, Pool CLI, AI coding, terminal, TUI, self-hosted, MCP
app_version: "0.1.0"
doc_version: "1.0"
doc_updated: "2026-06-04"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/pool-cli.md)为准。
:::

# 使用 Pool CLI 进行编码

Pool CLI 是一个基于终端的编码助手，可以帮助你通过自然语言读取代码、运行终端命令和编辑文件。在 Olares 上，这个命令行界面运行在一个基于浏览器的终端中，该终端配备了一个预配置的 Ubuntu 开发环境。

## 学习目标

在本指南中，你将学习如何：

- 从 Olares Market 安装 Pool CLI。
- 使用 Poolside 的云端 API 或本地模型将 Pool CLI 连接到模型。
- 执行基本的自然语言编码工作流。
- 为开发工作区配置目录访问权限。
- 安全地管理软件依赖。

## 前提条件

- 一台具有足够磁盘空间和内存的 Olares 设备。
- 一个 Poolside 账户，如果你计划使用云端 API 服务。
- 一台在 Olares 设备上运行的、针对编码优化的本地模型，如果你计划在本地运行任务。

   你可以通过以下方法之一安装本地模型：
   - **单模型应用**：一个运行特定模型的应用。本指南使用 **Qwen3-Coder 30B (Ollama)**。
   - **Ollama 应用**：一个托管多个模型的应用。确保已安装 [Ollama](ollama.md)，并且至少下载了一个模型，例如 `qwen3-coder:30b`。

## 安装 Pool CLI

1. 打开 Market，搜索 "Pool CLI"。

   ![Pool CLI](/images/manual/use-cases/pool-cli.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 连接模型

启动 Pool CLI 并将其连接到语言模型。选择以下连接方法之一。

### 使用 Poolside 云端服务连接

使用此方法可以利用 Poolside 的云端推理 API。

1. 从 Launchpad 打开 Pool CLI。
2. 输入以下命令触发登录认证：

   ```bash
   pool setup
   ```

3. 选择 **Log in with Poolside**，然后输入你的 Poolside API key 进行认证。
4. 选择以下模式之一来运行任务：

   <Tabs>
   <template #Interactive-mode>

   当你希望与助手进行持续的、类似聊天的对话时使用此方法。这适用于多步骤任务，其中助手需要读取上下文并向你询问后续澄清。

   1. 输入以下命令启动交互式会话：

      ```bash
      pool
      ```
   2. 使用自然语言与助手交互。
   3. 输入以下命令退出会话：

      ```bash
      /quit
      ```
   </template>
   <template #Automated-mode>

   使用此方法运行单个任务并立即返回到你的正常终端。这适用于不需要来回对话的快速请求。

   输入 `pool exec` 命令发送单个提示并退出。例如：

      ```bash
      pool exec -p "Create a folder named Test"
      ```

   :::tip
   默认情况下，助手会暂停并要求你手动批准任何系统操作，例如写入文件。要绕过此手动检查并允许助手立即执行操作，请添加 `--unsafe-auto-allow` 标志。例如，`pool exec -p "Create a folder named Test" --unsafe-auto-allow`。
   :::
   </template>
   </Tabs>

### 使用本地模型连接

使用此方法可以完全离线使用本地模型运行 Pool CLI。本示例使用模型应用 **Qwen3-Coder 30B (Ollama)**。

1. 从 Market 安装模型应用 **Qwen3-Coder 30B (Ollama)**。

   ![Qwen3-Coder 30B (Ollama)](/images/manual/use-cases/qwen3-coder-30b.png#bordered)

2. 从 Launchpad 打开模型应用，等待下载完成。
3. 记下页面上显示的精确模型名称。例如，`qwen3-coder:30b`。

   ![模型应用页面上的模型名称](/images/manual/use-cases/qwen3-coder-model-name.png#bordered){width=50%}

4. 打开 Settings，然后进入 **Applications** > **Qwen3-Coder 30B (Ollama)** > **Shared entrances**。

   ![Settings 中的模型应用端点](/images/manual/use-cases/qwen3-coder-30b-endpoint.png#bordered){width=70%}

5. 点击 **Qwen3-Coder 30B**，然后记下端点 URL。例如，`http://609c5d0c0.shared.olares.com`。
6. 进入 **Applications** > **Pool CLI** > **Manage environment variables**，然后点击 <i class="material-symbols-outlined">edit</i> 配置以下变量：

   - **USE_LOCAL_LLM**：设置为 `true` 以启用本地模型模式。
   - **POOLSIDE_STANDALONE_BASE_URL**：输入模型应用的端点 URL，并追加 `/v1`。例如，`http://609c5d0c0.shared.olares.com/v1`。
   - **POOL_MODEL**：输入你之前记下的模型名称。例如，`qwen3-coder:30b`。

7. 点击 **Apply**。等待 Pool CLI 容器重启。
8. 从 Launchpad 打开 Pool CLI，然后输入以下命令启动会话。

   ```bash
   pool-local
   ```

## 使用自然语言编码

连接到模型后，你可以使用对话式提示与 Pool CLI 交互。助手会解析你的请求来编写代码、修改文件和执行终端命令。

以下场景演示如何使用 Pool CLI 生成并运行一个简单的 Python 脚本。

1. 从 Launchpad 打开 Pool CLI。
2. 输入以下命令启动交互式会话：

   ```bash
   pool
   ```

3. 输入一个自然语言请求。例如：

   ```text
   Create a Python script named greeting.py that outputs the 
   current date and time
   ```

4. 查看助手提出的代码和操作。Pool CLI 会生成脚本并请求继续的权限。
5. 选择允许操作。终端会显示你的脚本输出。

   ![Pool CLI 代码结果](/images/manual/use-cases/pool-cli-results.png#bordered)

6. 要退出交互式会话，输入以下命令：

   ```bash
   /quit
   ```

7. 要验证输出，打开 Files，然后进入 **Data** > **pool** > **home** > **work**。

   ![Pool CLI 结果验证](/images/manual/use-cases/pool-cli-results-verify.png#bordered)

## 管理开发环境

Pool CLI 在一个预配置的 Ubuntu 24.04 环境中运行。根据你的项目需求自定义目录访问权限并安装额外的工具。

### 管理目录访问权限

默认情况下，所有项目工作都在 `/opt/data` 目录中进行。该目录会在应用重启后持久化保存你的文件，在 Olares 上位于 **Files** > **Data** > **pool** > **home** > **work**。

如果你希望 Pool CLI 访问 **Home** 或 **External** 目录中的文件，请配置环境变量：

1. 打开 Settings，然后进入 **Applications** > **Pool CLI** > **Manage environment variables**。
2. 根据需要指定以下变量：

   - **ALLOW_HOME_DIR_ACCESS**：设置为 `true` 以允许访问 Files 中的 **Home** 目录。这会将 **Home** 目录挂载到 `/home/userdata/home/`。
   - **ALLOW_EXTERNAL_DIR_ACCESS**：设置为 `true` 以允许访问 **External** 目录，例如挂载的 NAS 或 USB 驱动器。这会将 **External** 目录挂载到 `/home/userdata/external/`。

3. 点击 **Apply**。

### 查看预装开发工具

在安装额外软件之前，先查看工作区中已包含的工具。容器镜像基于 Ubuntu 24.04，并预装了许多常见的开发工具。

下表列出了主要类别和示例。

| 类别 | 包含的工具 |
|:---------|:---------------|
| 语言和运行时 | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3,<br>Lua, Perl, SQLite |
| 构建工具 | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`,<br>common `-dev` headers |
| CLI 工具 | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`,<br>`zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| 数据库客户端 | `postgresql-client`, `mysql-client`, `redis-tools` |

### 安装额外软件

如果你的项目需要预装工具之外的工具或库，你必须在容器的安全边界内管理它们。

#### 你无法自行安装的内容

如果你的项目需要系统级库（例如 `libpq-dev`, `ffmpeg`, `libssl-dev`），你无法直接安装它们。这些依赖必须由应用维护者添加到基础容器镜像中。

#### 你可以在工作区中安装的内容

在你的可写目录中（主要是 `/opt/data`），你可以使用常见工具无需 root 权限安装项目级依赖：

- **Python**：创建虚拟环境并使用 `pip`。例如：

   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install <package>
   ```

- **Node.js**：在你的项目文件夹中使用 `npm`。例如：

   ```bash
   npm install <package>
   ```

- **Rust/Go**（或其他编译语言）：将二进制文件安装到用户可写的路径。例如：

   ```bash
   cargo install --root ~/.local <package>   # Rust
   go install <package>@latest               # Go (installs to ~/go/bin)
   ```

## 常见问题

### 如何在云端模型和本地模型之间切换？

- 要从云端模式切换到本地模式，请将 `USE_LOCAL_LLM` 设置为 `true`，并在环境变量中配置 `POOLSIDE_STANDALONE_BASE_URL` 和 `POOL_MODEL`，然后重启应用。
- 要从本地模式切换到云端模式，请将 `USE_LOCAL_LLM` 设置为 `false` 并重启应用。你可能需要再次运行 `pool setup` 以重新认证。

### 缺少语言或库

确定缺失的工具是否为系统级依赖：

- 系统级依赖：你无法自行安装这些。应用维护者必须将它们添加到基础镜像中。如果你需要当前不可用的系统级库，请[提交 GitHub Issue](https://github.com/beclab/apps/issues) 来请求添加。
- 用户级依赖：使用 `venv`, `npm install` 或类似的本地工具来安装它们。

## 了解更多

- [Poolside documentation](https://docs.poolside.ai/get-started/overview)
- [使用 Claude Code 编写代码](claude-code.md)
