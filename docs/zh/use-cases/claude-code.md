---
outline: [2, 3]
description: 在 Olares 上设置 Claude Code，通过自然语言编写、测试和管理代码。通过 OAuth 或本地模型连接。
head:
  - - meta
    - name: keywords
      content: Olares, Claude Code, Anthropic, AI coding, Ollama, terminal, TUI, self-hosted
app_version: "0.1.3"
doc_version: "1.0"
doc_updated: "2026-05-19"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/claude-code.md)为准。
:::

# 使用 Claude Code 编写代码

Claude Code 是一个 AI 编码助手，帮助你使用自然语言编写、测试和管理代码。在 Olares 上，这个命令行界面运行在基于浏览器的终端中，配备了预配置的 Ubuntu 开发环境。

## 学习目标

在本指南中，你将学习如何：
- 从 Olares Market 安装 Claude Code 应用。
- 使用 Anthropic 订阅或本地模型将 Claude Code CLI 连接到模型。
- 执行基本和高级的自然语言编码工作流程。
- 安全地管理软件依赖。

## 前提条件

- 具有足够磁盘空间和内存的 Olares 设备。
- 如果你计划使用远程模型连接，需要活跃的 Claude Pro 或 Max 订阅。
- 如果你计划使用本地执行，需要在 Olares 设备上运行一个针对编码优化的本地模型。

   你可以使用以下方法之一安装本地模型：
   - **Ollama 应用**：一个托管多个模型的应用。确保 [Ollama 已安装](ollama.md) 并至少下载了一个模型，例如 `qwen3-coder:30b`。
   - **单模型应用**：将特定模型作为独立应用运行。确保从 Market 安装了模型应用且模型已完全下载。本指南使用 **Qwen3-Coder 30B (Ollama)**。

## 安装 Claude Code

1. 打开 Market，并搜索 "Claude Code"。
   
   ![Claude Code](/images/manual/use-cases/claude-code.png#bordered)

2. 点击 **Get**，然后点击 **Install**。等待安装完成。

## 连接到模型

启动 Claude Code CLI 并将其连接到语言模型。选择以下连接方法之一。

### 使用 Anthropic 订阅连接

如果你持有活跃的 Claude Pro 或 Max 订阅，请使用此方法。

1. 从 Launchpad 打开 Claude Code CLI。
2. 输入以下命令：

   ```bash
   claude
   ```
   
3. 选择终端主题，例如 **Dark mode**，然后按 **Enter**。
4. 选择 **Claude account with subscription** 作为登录方法，然后按 **Enter**。浏览器窗口将打开以进行登录。如果浏览器未能打开，请点击提供的 URL 手动登录。

   ![使用订阅账户登录 Claude Code](/images/manual/use-cases/claude-sign-subscription.png#bordered)

5. 在浏览器中完成登录流程，然后复制认证代码。
6. 返回终端，粘贴代码，然后按 **Enter** 完成登录。
7. 查看 **Accessing workspace: /opt/data** 安全提示，然后选择 **Yes, I trust this folder**。

   终端用户界面（TUI）自动打开。

   ![Claude Code TUI](/images/manual/use-cases/claude-code-tui.png#bordered)

### 使用本地模型连接

使用此方法在本地运行 Claude Code。此示例使用模型应用 **Qwen3-Coder 30B (Ollama)**。

1. 从 Market 安装模型应用 **Qwen3-Coder 30B (Ollama)**。

   ![Qwen3-Coder 30B (Ollama)](/images/manual/use-cases/qwen3-coder-30b.png#bordered)

2. 从 Launchpad 打开模型应用并等待下载完成。
3. 记下页面上显示的精确模型名称。例如，`qwen3-coder:30b`。

   ![模型应用页面上的模型名称](/images/manual/use-cases/qwen3-coder-model-name.png#bordered){width=50%}

4. 打开 Settings，然后前往 **Applications** > **Qwen3-Coder 30B (Ollama)** > **Shared entrances**。

   ![Settings 中的模型应用端点](/images/manual/use-cases/qwen3-coder-30b-endpoint.png#bordered){width=70%}

5. 点击 **Qwen3-Coder 30B**，然后记下端点 URL。例如，`http://609c5d0c0.shared.olares.com`。
6. 前往 **Applications** > **Claude Code** > **Manage environment variables**，然后指定以下环境变量：

   - **ANTHROPIC_AUTH_TOKEN**: 输入任何文本，例如 `ollama`。模型应用不会验证此值，但 Claude Code 需要一个填充的认证令牌。
   - **ANTHROPIC_BASE_URL**: 输入模型应用的端点 URL。例如，`http://609c5d0c0.shared.olares.com`。
   - **ANTHROPIC_MODEL**: 输入你之前记下的模型名称。例如，`qwen3-coder:30b`。

   ![Claude Code 环境变量设置](/images/manual/use-cases/claude-env-var.png#bordered){width=70%}  

7. 点击 **Apply**。等待约 10 秒让容器重启。
8. 从 Launchpad 打开 Claude Code CLI，然后在终端中输入 `claude` 以启动你的会话。

## 使用 Claude Code

所有项目工作都在 `/opt/data` 目录中进行，该目录在容器中作为 `$HOME`。此目录在应用重启之间持久保存你的文件。

以下示例演示如何与 Claude Code 交互以完成日常开发任务。

### 运行基本查询

1. 在 Claude Code CLI 中，输入以下命令：

   ```bash
   claude
   ```

2. 查看 **Accessing workspace: /opt/data** 安全提示，然后选择 **Yes, I trust this folder** 以授予 Claude Code 读取、编辑和执行权限。
3. （可选）在 TUI 中，运行 `/clear` 命令以使用空上下文开始新会话。

   ![Claude Code 首次聊天](/images/manual/use-cases/claude-first-chat.png#bordered)

   :::info 在模式之间切换
   如果你在远程和本地模型之间切换，先在 Claude Code 中运行 `/clear` 再开始新会话。这可以防止先前模型的上下文影响新工作空间。
   :::

4. 用自然语言描述你的任务。例如：

   ```text
   List the files in the current directory
   ```

   助手自动执行必要的内部命令来探索目录，并返回你的文件的详细列表。

   ![Claude Code 首次聊天结果](/images/manual/use-cases/claude-first-chat-result.png#bordered)

5. 查看结果。

### 构建全栈项目

Claude Code 可以创建多服务项目、运行测试并验证端到端集成。

以下示例演示如何使用单个 Node.js Express 服务器构建轻量级的 "Hello Olares" 网页应用，该服务器同时处理后端 API 和前端显示。

1. 在 Claude Code TUI 中，输入以下提示：

   ```text
   Create a simple full-stack "Hello Olares" application in a new directory called `hello-olares`.
   
   Please do the following:
   1. Initialize a Node.js project and install the `express` package.
   2. Create a backend API (`server.js`) that runs on port 3000 and has a single endpoint `/api/message` returning `{"message": "Hello Olares!"}`.
   3. Create a frontend (`public/index.html`) with vanilla JavaScript that fetches the message from the API and displays it on the screen. Configure the server to serve this static directory.
   4. Start the server in the background, use `curl` to verify the `/api/message` endpoint works, and then stop the server cleanly.
   ```

2. 等待 Claude Code 处理提示。助手自动初始化项目、安装 Express、编写代码、启动服务器并执行实时 curl 集成检查。
3. 当助手提示你允许继续时，选择 **Yes, and don't ask again...**。你可能需要为不同类型的操作批准多个提示。
4. 查看助手返回的最终摘要报告。它概述了新创建的项目结构、配置的后端 API、前端设置以及成功的 curl 测试结果。

      ![Claude Code 编码项目结果](/images/manual/use-cases/claude-code-report.png#bordered)

## 管理安全和开发环境

Claude Code 容器在严格的最小权限设置下运行以确保安全。

主进程和所有执行的命令使用非 root 用户（UID/GID 1000）。容器禁用 `allowPrivilegeEscalation` 并丢弃所有 Linux 功能。因此，`sudo` 和 `apt install` 等管理命令不可用。

### 查看预装的开发工具

在安装额外软件之前，请查看工作空间中已包含的工具。容器镜像基于 Ubuntu 24.04，并预装了许多常见的开发工具。

下表列出了主要类别和示例。

| 类别 | 包含的工具 |
|:---------|:---------------|
| 语言和运行时 | Python 3, Node.js, Go, Rust, Java (OpenJDK 21), Ruby, PHP 8.3,<br>Lua, Perl, SQLite |
| 构建工具 | `build-essential`, `cmake`, `ninja-build`, `clang`, `pkg-config`,<br>常见 `-dev` 头文件 |
| CLI 工具 | `git`, `git-lfs`, `curl`, `wget`, `jq`, `yq`, `openssh-client`, `unzip`,<br>`zip`, `rsync`, `tmux`, `htop`, `shellcheck` |
| 数据库客户端 | `postgresql-client`, `mysql-client`, `redis-tools` |

:::info
`ripgrep` (`rg`) 工具被有意排除，以防止与 Claude Code 的原生搜索行为冲突。
:::

### 安装额外软件

如果你的项目需要预装工具之外的工具或库，你必须在容器的安全边界内管理它们。

#### 你无法自行安装的内容

如果你的项目需要系统级库（例如 `libpq-dev`、`ffmpeg`、`libssl-dev`），你无法直接安装它们。这些依赖必须由应用维护者添加到基础容器镜像中。

#### 你可以在工作空间中安装的内容

在你的可写目录（主要是 `/opt/data`）中，你可以使用常见工具在没有 root 权限的情况下安装项目级依赖：

- **Python**: 创建虚拟环境并使用 `pip`。例如：

   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install <package>
   ```

- **Node.js**: 在你的项目文件夹中使用 `npm`。例如：

   ```bash
   npm install <package>
   ```

- **Rust/Go**（或其他编译语言）：将二进制文件安装到用户可写的路径。例如：

   ```bash
   cargo install --root ~/.local <package>   # Rust
   go install <package>@latest               # Go (安装到 ~/go/bin)
   ```

:::info
容器预配置环境变量 `PIP_BREAK_SYSTEM_PACKAGES=1`。虽然这允许系统范围的 Python 包安装，但建议使用虚拟环境以保持工作空间整洁和可靠。
:::

## 常见问题

### `claude: command not found`

等待片刻让 init 容器完成 Claude Code 的安装。验证 `$HOME/.local/bin` 目录是否存在于你的系统 `PATH` 中。

### OAuth 或安装脚本失败

验证你的 Olares 集群的出站网络连接。init 容器需要互联网访问才能从 https://claude.ai/install.sh 下载依赖。

### 缺少语言或库

确定缺失的工具是系统级依赖还是用户级依赖：

- 系统级依赖：你无法自行安装这些。应用维护者必须将它们添加到基础镜像中。如果你需要当前不可用的系统级库，请[提交 GitHub Issue](https://github.com/beclab/apps/issues) 请求添加。
- 用户级依赖：使用 `venv`、`npm install` 或类似的本地工具来安装它们。

## 了解更多

- [将 OpenCode 设置为你的 AI 编码助手](opencode.md)
- [Claude Code 官方文档](https://code.claude.com/docs)
