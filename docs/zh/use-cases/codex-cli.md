---
outline: [2, 3]
description: 在 Olares 上运行 Codex CLI，检查代码仓库、编辑代码、执行命令并测试更改。你可以使用 ChatGPT、OpenAI API 密钥或本地模型连接 Codex。
head:
  - - meta
    - name: keywords
      content: Olares, Codex CLI, OpenAI, AI 编程智能体, 浏览器终端, ChatGPT, 本地大语言模型, Qwen3.6, 自托管
app_version: "1.0.18"
doc_version: "1.0"
doc_updated: "2026-09-18"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/codex-cli.md)为准。
:::

# 在 Olares 上运行 Codex CLI

Codex CLI 是 OpenAI 开源的终端编程智能体。你可以使用自然语言让它检查代码仓库、编辑文件、运行命令和测试，并帮助你理解不熟悉的代码。

在 Olares 上，Codex CLI 运行于带有预配置开发环境的浏览器终端中。你可以通过 ChatGPT 或 OpenAI API 密钥连接 Codex，也可以将其接入本地的 OpenAI 兼容模型。

## 学习目标

在本指南中，你将学习如何：

- 从 Market 安装 Codex CLI。
- 使用 ChatGPT、OpenAI API 密钥或本地模型连接 Codex。
- 运行编码任务并检查结果。
- 切换模型连接并清除已保存的凭据。
- 登录 Olares CLI，以执行 Olares 管理任务。

## 前提条件

开始前，你需要：

- 一台运行 Olares 1.12.6 或更高版本，并具备足够磁盘空间和内存的 Olares 设备。
- 如果计划使用 ChatGPT 连接，需要一个具有 Codex 使用权限的 ChatGPT 账户。
- 如果计划通过 OpenAI Platform 连接，需要一个 OpenAI API 密钥。
- 如果计划使用本地模型，需要在 Olares 设备上运行一个针对编程优化的模型。本指南使用以下模型：

   | 模型类型 | 模型 | 获取方式 |
   | :--- | :--- | :--- |
   | 对话 | Qwen3.6-27B MTP (llama.cpp) | 从 Market 安装 |

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Codex CLI

1. 打开 Market，搜索“Codex CLI”。
2. 点击 **Get**，再点击 **Install**，然后等待安装完成。

## 连接模型

选择一种连接方式。在 Olares 的浏览器终端中，使用 ChatGPT 设备代码认证最为简单。

### 使用 ChatGPT 连接

Codex 的标准浏览器登录流程会将凭据返回到 `localhost:1455` 上的回调服务。在 Olares 容器中，你电脑上的浏览器无法直接访问该回调服务，因此请改用设备代码认证。

1. 从 Launchpad 打开 Codex CLI。
2. 运行：

   ```bash
   codex login --device-auth
   ```

3. 在浏览器中打开终端显示的链接，登录 ChatGPT，然后输入一次性代码。

   ![使用设备代码登录 Codex](/images/manual/use-cases/codex-cli-device-auth.png#bordered)

   如果设备代码登录不可用，请在 ChatGPT 安全设置中启用该功能，或请工作区管理员允许使用设备代码登录。

4. 返回终端，检查当前认证方式：

   ```bash
   codex login status
   ```

   ![检查 Codex 登录状态](/images/manual/use-cases/codex-cli-login-status.png#bordered)

::: details 备用方法：在容器内完成浏览器回调
仅在设备代码认证不可用时使用此方法。

1. 在第一个 Codex CLI 终端中运行 `codex login`，并保持该进程运行。
2. 在浏览器中完成登录。当浏览器打开以 `http://localhost:1455/auth/callback` 开头的 URL 时，从地址栏复制完整 URL。
3. 在另一个浏览器标签页中打开 Codex CLI。将 URL 中的 `localhost` 替换为 `127.0.0.1`，保持查询字符串不变，然后使用 `curl` 访问该 URL：

   ```bash
   curl -i "http://127.0.0.1:1455/auth/callback?code=<code>&scope=<scope>&state=<state>"
   ```

4. 确认响应为 `302 Found`，且 `Location` 响应头中包含 `/success`。
5. 返回第一个终端，运行 `codex login status`。

授权码和回调 URL 都是短期凭据。不要分享它们，也不要将其包含在截图中。
:::

### 使用 OpenAI API 密钥连接

使用此方式时，Codex 用量将计入你的 OpenAI Platform 账户。

1. 前往 **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**。
2. 将 **OPENAI_API_KEY** 设置为你的 OpenAI API 密钥。
3. 点击 **Apply**，等待 Codex CLI 重启。
4. 从 Launchpad 打开 Codex CLI，然后运行 `codex`。

### 使用本地模型连接

本示例通过 OpenAI 兼容 API 将 Codex CLI 连接到 Qwen3.6-27B MTP (llama.cpp)。

#### 获取模型连接信息

1. 从 Launchpad 打开 Qwen3.6-27B MTP (llama.cpp)。Model Console 会自动打开。
2. 等待 **Model** 显示 **READY**，且 **Engine** 显示 **RUNNING**。
3. 在 **Model** 下，准确复制显示的 **Model name**。
4. 在 **Engine** 下：

   a. 在 **Connection source** 中选择 **Apps in Olares**。

   b. 在 **API format** 中选择 **OpenAI-Compatible**。

   c. 准确复制显示的 **Base URL**。

#### 配置 Codex CLI

1. 前往 **Settings** > **Applications** > **Codex CLI** > **Manage environment variables**。
2. 配置以下环境变量：

   - **OPENAI_BASE_URL**：粘贴从 Model Console 复制的 Base URL，包括末尾的 `/v1`。
   - **OPENAI_API_KEY**：输入一个非空占位值，例如 `olares`。本地模型应用不会验证来自同一 Olares 集群内其他应用的真实 OpenAI 密钥，但 Codex CLI 要求该值非空。
   - **CODEX_MODEL**：输入从 Model Console 复制的准确 Model name。

   ![配置 Codex CLI 本地模型环境变量](/images/manual/use-cases/codex-cli-local-model-env.png#bordered)

3. 点击 **Apply**，等待 Codex CLI 重启。
4. 从 Launchpad 打开 Codex CLI，然后运行：

   ```bash
   codex
   ```

## 使用 Codex CLI

所有项目工作都在 `/opt/data` 目录中进行。存储在该目录中的文件会在应用重启后保留。

1. 检查已安装的 CLI 版本：

   ```bash
   codex --version
   ```

2. 启动交互式会话：

   ```bash
   codex
   ```

3. 描述一个明确的任务。例如：

   ```text
   Read the current directory and summarize the project structure.
   ```

   你也可以让 Codex 创建一个小文件：

   ```text
   Create a minimal README draft for this repository.
   ```

4. 批准前，检查 Codex 建议执行的命令和文件更改。
5. 检查结果，然后继续提问或要求 Codex 进行其他修改。

   ![Codex CLI 编码会话](/images/manual/use-cases/codex-cli-session.png#bordered)

## 使用 Olares CLI 管理 Olares

Codex 认证只授予对所选模型服务的访问权限。如果要让 Codex 安装 Olares 应用、检查集群状态或执行其他 Olares 管理任务，还需要单独登录 Olares CLI：

```bash
olares-cli profile login --olares-id <your-olares-id>
```

根据提示完成双重认证。更多信息，请参阅[登录 Olares CLI](/zh/developer/cli-log-in.md)。

## 切换连接或退出登录

### 从 ChatGPT 切换到 OpenAI API 密钥

在 **Settings** > **Applications** > **Codex CLI** > **Manage environment variables** 中设置 **OPENAI_API_KEY**，点击 **Apply**，然后等待应用重启。

### 从 OpenAI API 密钥切换到 ChatGPT

清除 **OPENAI_API_KEY**，点击 **Apply**，然后等待应用重启。接着运行 `codex login --device-auth`。

### 从本地模型切换到 OpenAI

清除 **OPENAI_BASE_URL** 和 **CODEX_MODEL**。如需使用 API 密钥认证，请保留真实的 **OPENAI_API_KEY**；如需使用 ChatGPT 认证，请一并清除该变量。点击 **Apply**，然后等待应用重启。

### 退出登录

1. 运行：

   ```bash
   codex logout
   codex login status
   ```

2. 当状态显示 Codex 已退出登录后，再继续后续操作。

退出登录会清除 Codex 保存的凭据，但不会从应用环境中移除 **OPENAI_API_KEY**，也不会退出 Olares CLI。

## 常见问题

### 为什么浏览器登录停在 `localhost:1455`？

回调服务运行在 Codex CLI 容器中，因此你电脑上的浏览器无法通过自身的 `localhost` 访问该服务。请使用 `codex login --device-auth`。如果设备代码认证不可用，请使用[使用 ChatGPT 连接](#使用-chatgpt-连接)中介绍的备用方法。

### 为什么登录 ChatGPT 后，Codex 仍然使用本地模型？

`OPENAI_BASE_URL` 和 `CODEX_MODEL` 仍会将 Codex 请求路由到本地服务。请在应用设置中清除这两个变量，点击 **Apply**，然后等待应用重启。

### 为什么本地模型返回 404？

打开模型的 Model Console，重新复制 Base URL。选择 **OpenAI-Compatible**，并准确使用界面显示的 URL，包括 `/v1`。同时确认 **Model** 显示 **READY**，且 **Engine** 显示 **RUNNING**。

### 为什么 Codex 报告 `model not found`？

将 `CODEX_MODEL` 与 Model Console 中的 **Model name** 进行比较，两者必须完全一致。

### 如何删除残留的登录文件？

先运行 `codex logout`。如果 `codex login status` 仍显示存在已保存的登录信息，请删除应用的凭据文件，然后再次检查：

```bash
rm -f /opt/data/.codex/auth.json
codex login status
```

该命令只会删除存储在应用数据中的 Codex 凭据，不会清除 **OPENAI_API_KEY**，也不会影响 Olares CLI 认证。

### Codex 报告缺少智能体执行能力时该怎么办？

更新应用数据目录中的 Codex 可执行文件：

```bash
curl -fsSL https://chatgpt.com/codex/install.sh | CODEX_INSTALL_DIR=/opt/data/.local/bin sh
```

重启 Codex CLI 会话，然后重试。

## 了解更多

- [Codex CLI 文档](https://developers.openai.com/codex/cli)：了解 Codex CLI 工作流和命令。
- [Codex 认证](https://developers.openai.com/codex/auth)：了解 ChatGPT、API 密钥和无头登录方式。
- [使用 Engine Base 应用托管本地大语言模型](llm-base-apps.md)：通过 Model Console 部署和管理模型。
- [安装 Olares CLI](/zh/developer/cli-install.md)：设置 Olares 管理命令和 Agent 技能。
