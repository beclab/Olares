---
outline: [2, 3]
description: 在 Olares 上通过 Gemma4 等本地大语言模型（LLM）运行 NemoClaw。无需云端 API，即可部署一个基于 NVIDIA OpenShell 运行时的常驻 AI Agent。
head:
  - - meta
    - name: keywords
      content: Olares, NemoClaw, NVIDIA, OpenShell, OpenClaw, 本地 LLM, AI 助手, Discord, 网页搜索, ClawHub, skills, plugins
app_version: "1.0.8"
doc_version: "1.1"
doc_updated: "2026-05-11"
---

# 使用本地 LLM 运行 NemoClaw

NemoClaw 是 NVIDIA 开源的参考技术栈，内置 NVIDIA OpenShell 运行时，并在其中运行 OpenClaw。

本文介绍如何在 Olares 上使用 Gemma4 26B 模型应用作为后端大语言模型（LLM）来运行 NemoClaw。

:::warning Alpha 软件
NemoClaw 是 NVIDIA 发布的早期预览版本，不建议用于生产环境。如需了解官方更新和社区反馈，可参阅 [NVIDIA/NemoClaw](https://github.com/NVIDIA/NemoClaw)。
:::

## 学习目标

通过本教程，你将学习：

- 安装并配置带有本地 LLM 的 NemoClaw。
- 保持模型处于加载状态，以实现 Agent 常驻响应。
- 与 Agent 开始第一次聊天。
- 将 Agent 连接到 Discord 进行远程聊天。
- 启用实时网页搜索。

## 前提条件

- 已在 Olares 设备上安装并运行一款本地模型应用。
- 拥有从应用市场安装应用、编辑应用设置的管理员权限。

## 获取模型名称和端点 URL

安装 NemoClaw 时需要填写模型名称及其共享端点 URL。

1. 从启动台打开你的模型应用，记下页面上显示的模型名称。本文示例为 `gemma4:26b`。

   ![模型应用中显示的模型名称](/images/manual/use-cases/gemma4-26b-downloaded.png#bordered)

2. 打开设置，进入**应用** > **Gemma4 26B Q4_K_M (Ollama)**。
3. 在**共享入口**选择 **Gemma4 26B Q4_K_M**，查看端点 URL。

   ![获取共享端点](/images/manual/use-cases/gemma4-26b-shared-laresprime.png#bordered){width=90%}

4. 记下共享端点。例如：

   ```plain
   http://2e53d5230.shared.olares.com
   ```

   :::tip 为什么使用共享端点？
   模型应用主页上的 URL 与用户绑定，并通过浏览器路由访问。共享端点可供 Olares 上的其他应用直接访问，不存在登录或跨域（CORS）问题，满足 NemoClaw 使用场景。
   :::

## 安装 NemoClaw

1. 打开应用市场，搜索“NemoClaw”。

   ![应用市场中的 NemoClaw](/images/manual/use-cases/nemoclaw.png#bordered)

2. 点击**获取**，然后点击**安装**。
3. 按提示设置环境变量：

   - **NEMOCLAW_ENDPOINT_URL**：输入或粘贴共享端点 URL，并在末尾添加 `/v1`，例如 `http://2e53d5230.shared.olares.com/v1`。
   - **NEMOCLAW_MODEL**：输入或粘贴模型名称，例如 `gemma4:26b`。

   ![为 NemoClaw 设置环境变量](/images/manual/use-cases/nemoclaw-set-environment-variables.png#bordered){width=70%}

   :::tip
   之后也可在**设置** > **应用** > **NemoClaw** > **管理环境变量**中更改这些环境变量。
   :::

4. 点击 **Confirm**，并等待安装完成。

   安装大约需要 15 分钟，具体取决于你的网络状态。在此期间，NemoClaw 会安装 NVIDIA OpenShell 运行时并执行初始 Agent 引导流程。

   :::warning
   安装过程中请保持模型应用运行。初始引导需要模型处于可访问状态，如果模型停止运行或不可用，将无法完成安装。
   :::

安装完成后，启动台会出现两个快捷方式：

- **NemoClaw CLI**：用于运行 NemoClaw 和 OpenClaw 命令的终端界面。
- **OpenClaw Web UI**：OpenClaw 的网页控制台。

## 保持模型处于加载状态（可选）

默认情况下，本地 LLM 会在 5 分钟无活动后从内存中卸载，下一次回复需要等待模型重新加载。如果你希望 Agent 常驻在线，可以在模型应用上启用保持活跃（keep-alive）设置，让模型常驻在内存中。

1. 打开设置，进入**应用** > **Gemma4 26B Q4_K_M** > **管理环境变量**。
2. 找到 **KEEP_ALIVE**，点击 <i class="material-symbols-outlined">edit_square</i>，将值设置为 **true**，然后点击**确认**。

   ![为模型应用启用 KEEP_ALIVE](/images/manual/use-cases/keep-alive-enable.png#bordered){width=80%}

3. 点击**应用**。

:::tip 何时不设置 KEEP_ALIVE
保持模型加载会持续占用显存。如果你只是偶尔使用 Agent，并且可以接受冷启动延迟，可以不设置 **KEEP_ALIVE**。
:::

## 开始第一次聊天

你可以在 OpenClaw Web UI 中与 Agent 聊天，也可以在 NemoClaw CLI 内的 OpenClaw TUI 中聊天。由于模型和端点已在安装时配置完成，你可以跳过手动引导，直接进入会话。

<tabs>
<template #使用-OpenClaw-Web-UI>
1. 从启动台打开 OpenClaw Web UI 应用，你会直接进入聊天界面。

2. 发送一条测试消息，例如 `Hi`。

   模型加载到内存时，第一次回复可能需要约 30 秒。后续回复会快很多。模型加载完成后，可以询问 Agent 自身信息来确认配置。例如：

   ```text
   How are you, and what model are you running on?
   ```

   ![OpenClaw Web UI 中的 NemoClaw 聊天](/images/manual/use-cases/nemoclaw-openclaw-chat-test.png#bordered)

</template>

<template #使用-NemoClaw-CLI>

1. 从启动台打开 NemoClaw CLI 应用。
2. 运行以下命令连接到沙盒：

   ```bash
   nemoclaw my-assistant connect
   ```

   等待终端显示沙盒提示符。

   ![已连接沙盒](/images/manual/use-cases/nemoclaw-connect.png#bordered)

3. 启动 OpenClaw TUI：

   ```bash
   openclaw tui
   ```

   ![在沙盒 shell 中启动 OpenClaw TUI](/images/manual/use-cases/nemoclaw-tui.png#bordered)

4. 发送一条测试消息，例如 `Hi`。

   模型加载到内存时，第一次回复可能需要约 30 秒。后续回复会快很多。模型加载完成后，可以询问 Agent 自身信息来确认配置。例如：

   ```text
   How are you, and what model are you running on?
   ```

   ![NemoClaw 聊天会话](/images/manual/use-cases/nemoclaw-chat-test.png#bordered)

</template>
</tabs>

## 集成 Discord

如需远程与 NemoClaw Agent 聊天，可以将它连接到 Discord Bot。你需要一个 Discord 账号，以及一个你有权限添加 Bot 的服务器。

### 步骤 1：创建 Discord Bot

1. 使用你的 Discord 账号登录 [Discord Developer Portal](https://discord.com/developers/applications)。
2. 点击 **New Application**。

   ![Discord Developer Portal 中的新建应用](/images/manual/use-cases/new-app.png#bordered){width=90%}

3. 输入新应用名称，同意条款，然后点击 **Create**。

   ![创建应用窗口](/images/manual/use-cases/create-app.png#bordered){width=40%}

4. 在左侧边栏选择 **Bot**。
5. 向下滚动到 **Privileged Gateway Intents** 区域，启用以下设置：

   - Presence Intent
   - Server Members Intent
   - Message Content Intent

6. 点击 **Save Changes**。
7. 向上滚动到 **Token** 区域，点击 **Reset Token**，并复制生成的 token。步骤 3 中会用到这个 token。

   ![重置 token](/images/manual/use-cases/reset-token.png#bordered)

### 步骤 2：邀请 Bot 加入服务器

1. 在左侧边栏选择 **OAuth2**，找到 **OAuth2 URL Generator** 区域。

    a. 在 **Scopes** 中选择 **Bot** 和 **applications.commands**。

    ![OAuth2 URL Generator](/images/manual/use-cases/oauth21.png#bordered)

    b. 向下滚动到 **Bot Permissions**，并按截图所示进行配置。之后也可以再调整这些权限。

    ![Bot 权限](/images/manual/use-cases/bot-permissions1.png#bordered)

2. 复制页面底部的 **Generated URL**。
3. 将该 URL 粘贴到新的浏览器标签页，在 **Add to server** 中选择你的 Discord 服务器，点击 **Continue**，再点击 **Authorize**。

   Bot 授权完成并加入服务器。

   ![Bot 已添加到服务器](/images/manual/use-cases/bot-added.png#bordered)

### 步骤 3：配置 Discord 频道

NemoClaw 会在沙盒运行时中运行 OpenClaw，因此必须从运行时 shell 内配置频道。

1. 从启动台打开 NemoClaw CLI 应用。
2. 连接到运行时沙盒：

   ```bash
   nemoclaw my-assistant connect
   ```

   等待终端显示沙盒提示符，例如 `sandbox@my-assistant:~$`。

3. 运行频道配置向导：

   ```bash
   openclaw config --section channels
   ```

4. 按提示添加 Discord：
   | 设置 | 选项 |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Channels | Configure/link |
   | Select a channel | Discord (Bot API) |
   | How do you want to provide this Discord bot token? | 输入 Discord bot token，并粘贴步骤 1 中的 token。 |
   | Configure Discord channels access | Yes |
   | Discord channels access | Open (allow all channels) |

5. 完成后，当系统提示选择频道时，选择 **Finished**。

6. 当系统提示配置 DM（Direct Message）访问策略时，选择 **Pairing**。

:::info Discord 频道卡在 `startup-not-ready`
如果 Discord 频道在 OpenClaw Web UI 中显示 `startup-not-ready`，需重启 gateway。操作步骤参见[常见问题](nemoclaw-common-issues.md#discord-频道卡在-startup-not-ready-状态)。
:::

### 步骤 4：授权你的 Discord 账号

出于安全考虑，Bot 不会回应未授权用户。需要先将自己的 Discord 账号与 Bot 配对。

1. 打开 Discord，向你的 Bot 发送一条私信。

   Bot 会回复一个配对码和一条命令。

   ![Bot 私信返回的配对码](/images/manual/use-cases/nemoclaw-discord-pairing-code.png#bordered)

2. 切回 NemoClaw CLI 沙盒 shell，使用 Bot 提供的命令批准配对。例如：

   ```bash
   openclaw pairing approve discord FY6PAVY8
   ```
   看到以下输出时，表示 Discord 已授权：
   ```text
   Approved discord sender 1277468602303385654.
   ```

3. 授权后，你就可以直接在 Discord 中与 Agent 聊天。

   ![在 Discord 中与 Agent 聊天](/images/manual/use-cases/nemoclaw-discord-chat.png#bordered)

## 启用网页搜索

默认情况下，Agent 只会基于训练数据回答。要让它获取实时互联网信息，可以使用网页搜索提供方。下面以 SearXNG 为例，你可以从 Olares 应用市场安装一个自托管的 SearXNG 实例。

1. 从 Olares 应用市场安装 SearXNG。
2. 打开设置，进入**应用**> **SearXNG**。
3. 在**共享入口**中选择 **SearXNG**，查看端点 URL。

   ![获取共享端点](/images/manual/use-cases/searxng-shared-laresprime.png#bordered){width=90%}
4. 复制共享端点。例如：

   ```plain
   http://d1236e020.shared.olares.com
   ```

5. 从启动台打开 NemoClaw CLI 应用。
6. 连接到运行时沙盒：

   ```bash
   nemoclaw my-assistant connect
   ```

7. 运行网页工具配置向导：

   ```bash
   openclaw config --section web
   ```

8. 按如下方式配置：

   | 设置 | 选项 |
   |:---------|:-------|
   | Where will the Gateway run | Local (this machine) |
   | Enable web_search | Yes |
   | Search provider | SearXNG |
   | SearXNG Base URL | 粘贴步骤 4 中的 SearXNG 共享端点 |
   | Enable web_fetch (keyless HTTP fetch) | Yes |

9. 要检查是否生效，可以向 Agent 提一个需要实时信息的问题。例如：

   ```text
   What are today's top tech news headlines?
   ```

   Agent 应能获取并引用实时网页结果。
   ![网页搜索结果](/images/manual/use-cases/nemoclaw-web-search-result.png#bordered){width=90%}

## 安装技能

技能可以为 Agent 增加能力，例如管理 Olares 文件和应用，或集成 Google Workspace。

1. 从启动台打开 OpenClaw Web UI。
2. 进入 **Skills**。
3. 在 ClawHub 搜索框中搜索技能，然后点击 **Install**。
4. 打开 OpenClaw Web UI 的聊天页面，运行 `/reset` 开启新会话，让 Agent 识别新安装的技能。如果你已配置 Discord 等频道，也需要在每个频道会话中运行 `/reset`。

   :::tip
   你也可以在 NemoClaw CLI 沙盒中使用 `openclaw config --section skills` 安装技能。
   :::

常见技能的详细操作请参阅：

- [使用 Olares CLI 管理 Olares](nemoclaw-olares-cli.md)：让 Agent 通过自然语言操作 Olares 设备上的文件和应用。
- [集成 Google Workspace](nemoclaw-google-workspace.md)：通过 gog 技能连接 Gmail、Calendar 和 Drive。

如需了解更多技能管理方法，请参阅[管理技能和插件](openclaw-skills.md)。

## 安装插件

插件可以为 OpenClaw 扩展更多频道和集成能力。

1. 从启动台打开 NemoClaw CLI 应用。
2. 连接到运行时沙盒：

   ```bash
   nemoclaw my-assistant connect
   ```

3. 安装 BlueBubbles 插件：

   ```bash
   openclaw plugins install @openclaw/bluebubbles
   ```

如需安装其他插件，可以在运行时中使用标准的 `openclaw plugins list` 和 `openclaw plugins install <name>` 命令。详情请参阅[管理技能和插件](openclaw-skills.md)。

## 常见问题

常见问题和解决方法请参阅[常见问题](nemoclaw-common-issues.md)。

## 了解更多

- [NVIDIA NemoClaw](https://build.nvidia.com/nemoclaw)：NVIDIA 官方参考技术栈和文档。
- [OpenClaw](openclaw.md)：设置 OpenClaw 功能，例如 persona 设置。
