---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/hermes
outline: [2,3]
description: 在 Olares 上将 Hermes Agent 连接到本地模型，在浏览器工作区中聊天、委派任务，并使用 Olares 技能管理文件和应用程序。
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, Hermes Agent, Webapp, 委派, 本地模型, Discord 机器人, 自托管
app_version: "1.3.51"
doc_version: "3.3"
doc_updated: "2026-11-05"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/hermes.md)为准。
:::

# 用 Hermes 搭建一个自主工作的 AI 助手

<VersionRouteSelect />

:::info 应用版本
下文介绍的 Hermes Webapp 入口及内置的委派插件需要 Chart 1.3.51（Hermes Agent 0.22.7）。请在应用市场的应用详情页查看 **Chart version**。旧版本只有两个快捷方式：Hermes CLI 和仪表盘。
:::

Hermes Agent 是一个能够自主执行任务的 AI 助手。接入本地模型后，它可以执行系统任务、编写代码、管理工作流程。它能够在多次对话之间保持记忆，还能根据与你的互动，自动生成可复用的技能。

在 Olares 中，你可以使用内置的 Olares 技能，让助手管理设备上的文件和应用程序。你还可以通过 Discord 等聊天软件远程与它对话，或将其接入 Open WebUI 等其他应用。

## 学习目标

- 在 Olares 上安装 Hermes Agent
- 配置 Hermes Agent，连接本地模型
- 通过 Hermes CLI 或 Hermes Webapp 与助手对话
- 将任务委派给子代理并查看结果
- 接入 Discord，实现远程聊天
- 使用 Olares 技能，让助手帮你管理文件和应用程序
- 开启 Gateway API，将 Hermes Agent 与 Open WebUI 等应用对接

## 前提条件

开始前，你需要：

<!--@include: ../reusables/ai-service-connections.md#router-prerequisite-->
- **Discord 账号**：用来创建机器人应用。
- **Discord 服务器**：确保你在这个服务器上有添加机器人的权限。
- 以下模型：
    | 模型类型 | 模型 | 获取方式 |
    | :--- | :--- | :--- |
    | 聊天 | Qwen3.8-27B (llama.cpp) | 从应用市场安装 |
    :::tip
    Hermes Agent 要求模型的上下文窗口至少为 64K tokens。
    :::

<!--@include: ../reusables/ai-service-connections.md#use-different-model-->

## 安装 Hermes Agent

1. 打开应用市场，搜索 “Hermes”。

   ![安装 Hermes Agent](/images/zh/manual/use-cases/hermes-agent.png#bordered)

2. 点击**获取**，然后点击**安装**，等待安装完成。

Chart 1.3.51 在启动台提供三个快捷方式：

| 快捷方式 | 用途 |
| --- | --- |
| Hermes CLI | 运行初始设置、输入命令和排查问题。 |
| Hermes Dashboard | 管理模型、技能、消息渠道和网关设置。 |
| Hermes Webapp | 在浏览器中聊天、继续已保存的会话、跟进委派的任务。 |

先在下方完成模型配置，再开始你的第一次 Hermes Webapp 对话。

:::tip 运行多个 Hermes 实例
Olares 支持克隆应用。如果你想同时运行多个独立的 AI 助手，分别处理不同的任务，可以克隆 Hermes Agent 这个应用。具体步骤，见[克隆应用](/zh/manual/olares/market/clone-apps.md)。
:::

## 配置 Hermes Agent

按照快速设置向导，将 Hermes Agent 连接到你本地的模型。

### 第 1 步：获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

6. 在 Router 的 **LLM** 页面，点击 Qwen3.8-27B (llama.cpp) 所在行的信息图标，打开 **Model card**。在 **Engine args** 中读取 `-c` 后的数值。例如，`-c 104448` 表示下一步应填写 `104448`。截图说明见[查看实际上下文大小](llm-base-apps.md#在-router-中查看模型详情)。

### 第 2 步：运行设置向导

1. 从启动台打开 Hermes CLI。
2. 输入以下命令，启动配置向导：

   ```bash
   hermes setup
   ```

3. 向导会逐步引导你完成设置。使用方向键移动，按**回车**键确认。

    下表中的模型选择提示是不同分支。根据检测到的模型数量完成对应分支，再继续填写 **Context length in tokens**。

    | 配置   | 选项   |
    |:-----------|:---------|
    | How would you like to set up Hermes | 选择 **Full setup — configure every provider, tool & option yourself (bring your own keys)**。 |
    | Select provider | 选择 **Custom endpoint (enter URL manually)**。  |
    | API base URL  | 输入从 Router 复制的 **Base URL**。<br>例如 `https://router.<your-olares-domain>/v1`。  |
    | API key  | 填写 `olares`。<br>出于安全考虑，输入内容会被隐藏。 |
    | Select API compatibility mode | 输入 `1`，选择 **Auto-detect**。 |
    | Use this model? [Y/n] — 仅检测到一个模型时出现 | 输入 `n`，然后在接下来的 **Model name** 提示中填写 `default-chat`。确认提示只接受是/否，不是在这里输入模型名。 |
    | Select model [1-N] or type name — 检测到多个模型时出现 | 直接输入 `default-chat`，不要选择列表中的编号。 |
    | Model name — 未检测到模型时出现 | 输入 `default-chat`。 |
    | Context length in tokens | 填写 Router 的 **Model card > Engine args** 中的精确上下文大小。例如，`-c 104448` 对应填写 `104448`。Hermes 要求至少 `65536` token，且填写值不能超过引擎实际配置。 |
    | Display name | 输入一个便于识别该模型的名称，例如 `router-chat`。|
    | Select terminal backend | 选择 **Local - run directly on this machine**。 |
    | Select platforms to configure | 按**回车**键暂时跳过。 |
    | Tools for CLI | 按**回车**键暂时跳过。 |
    | Choose a provider | 选择 **Skip - keep defaults / configure later**。 |
    | Select Search Provider | 选择 **Skip - keep defaults / configure later**。 |

4. 继续完成其余可选服务的提示。没有配置的服务可以直接跳过。在用量统计提示处，选择 **No thanks**，不共享统计信息。
5. 等待向导显示 `Ready to go`。保持 Hermes CLI 窗口开启，后续步骤仍需使用。

## 与 Hermes Agent 对话

### 方式 A：在文本用户界面中对话

文本用户界面（TUI）直接在 Hermes CLI 里运行，无需额外配置，适合快速体验。

1. 设置向导完成后，会显示 `Ready to go` 及以下可用命令：

    - `hermes`：开始聊天。
    - `hermes gateway`：启动消息网关。
    - `hermes doctor`：检查问题。

2. 输入 `hermes` 并按**回车**键启动 TUI。

   :::tip
   如果你退出了向导，也可以在 Hermes CLI 里手动输入 `hermes` 命令来启动聊天。
   :::

3. 发一条消息，例如 `你当前使用的模型是哪个`，查看助手是否回复正常。

    ![Hermes TUI 聊天界面](/images/manual/use-cases/hermes-tui1.png#bordered)

4. 如需退出 TUI 回到普通命令行，输入 `/exit` 并按**回车**键。

### 方式 B：通过 Hermes Webapp 在浏览器中对话

Hermes Webapp 为同一个 Hermes Agent 安装实例提供浏览器工作区。它与其他入口共享模型配置和已保存的会话。

1. 从启动台打开 Hermes CLI。如果有聊天正在运行，输入 `/exit` 回到命令行。
2. 运行以下命令，查看本次安装生成的 Webapp 登录凭据：

   ```bash
   cat /opt/data/webapp/olares-login.txt
   ```

   如果你在升级前已经配置过 Hermes 的用户名和密码认证，请继续使用原凭据。升级会保留这些凭据，不会为该账号生成新的登录文件。

3. 从启动台打开 Hermes Webapp。输入用户名和密码，点击 **Sign in**。这个登录与 Olares 账号相互独立。
4. 点击 **New session**，输入并发送一条消息。例如，让助手帮你规划每周一次的文件清理。先检查它建议的操作，确认无误后再让它执行更改。

   按照上述 Router 配置，模型选择器会显示 **Default Chat**。助手会使用 Router 中为 `default-chat` 选择的模型。

5. 想继续之前的对话时，在侧边栏的 **Sessions** 中选择对应会话。

生成的登录信息和对话记录在应用重启后仍会保留。

:::tip 启动台中找不到 Hermes Webapp
请检查你的安装是否为 Chart 1.3.51。如果你之前显式禁用过 Webapp，升级后会保留该设置。要重新启用，请在 Hermes CLI 中先运行 `hermes config set webapp.enabled true`，再运行 `hermes gateway restart`。
:::

### 方式 C：通过 Discord 远程对话

如果你想在手机或其他设备上与助手聊天，可以将其接入 Discord 机器人。

#### 第 1 步：创建一个 Discord 机器人

1. 用你的 Discord 账号登录 [Discord Developer Portal](https://discord.com/developers/home)。
2. 点击右上角的用户头像，将**语言**设置为**中文**。
3. 从左侧边栏中，选择 **APP**，然后点击**新 APP**。

    ![Discord 开发者门户新建应用](/images/zh/manual/use-cases/hermes-new-app.png#bordered)

4. 为新应用命名（例如 `Cool-hermes`），选择同意条款，然后点击**创建**。

    ![创建一个新 APP 窗口](/images/zh/manual/use-cases/hermes-create-app.png#bordered){width=40%}

5. 从左侧边栏中，选择**机器人**。
6. 向下滚动页面至 **Privileged Gateway Intents** 部分，打开以下三个选项：

    - Presence Intent
    - Server Members Intent
    - Message Content Intent

    ![特权网关意图](/images/zh/manual/use-cases/hermes-privileged-gateway-intents.png#bordered)

7. 点击**保存更改**。
8. 向上滚动页面至**令牌**部分，点击**重置令牌**，然后复制生成的令牌。后续配置消息平台时需提供该令牌。

    ![重置令牌](/images/zh/manual/use-cases/hermes-reset-token.png#bordered)

#### 第 2 步：将机器人添加到你的服务器

1. 从左侧边栏中，选择 **OAuth2**，找到 **OAuth2 URL 生成器**部分：

    a. 在**范围**区域，勾选 **bot** 和 **applications.commands**。

    ![OAuth2 URL 生成器](/images/zh/manual/use-cases/hermes-oauth21.png#bordered)

    b. 向下滚动页面至**机器人权限**部分，按下面图示勾选权限。后续可以再调整。

    ![机器人权限](/images/zh/manual/use-cases/hermes-bot-permissions.png#bordered)

2. 复制页面最底部**已生成的 URL**。
3. 将这个 URL 粘贴到浏览器的新标签页中，在**添加至服务器**的下拉框中，选择你的 Discord 服务器，点击**继续**，再点击**授权**。

    ![将 Discord 机器人添加到服务器](/images/zh/manual/use-cases/hermes-add-bot.png#bordered)

    添加完成后，机器人就会出现在你对应的服务器中。

#### 第 3 步：配置消息平台

配置并运行 Hermes gateway，将 Hermes Agent 接入你的 Discord 机器人。

1. 打开 Hermes CLI，输入以下命令，启动配置向导：

   ```bash
   hermes gateway setup
   ```

2. **Select a platform to configure** 一项，选择 **Discord**。
3. 按提示填好机器人集成信息：

   - **Discord bot token**：填写你从 Discord Developer Portal 复制的令牌。注意：输入的内容是隐藏的。
   - **Allowed user IDs or usernames**：填写你的 Discord 用户 ID，仅限你自己使用。
   - （可选）**Home channel ID**：填写机器人所在频道的 ID。

4. 选择 **Done**。
5. 当提示 **Restart the gateway to pick up changes** 时，输入 `y`。
6. 前往 Discord 查看机器人状态。如果图标为绿色，表示机器人已在线，说明网关重启成功、配置正确。

#### 第 4 步：授权你的账号

出于安全考虑，机器人不会回应未授权用户。你需要将自己的 Discord 账号与机器人配对。

1. 向你的新机器人发送一条私信。机器人会回复一条错误消息，其中包含一个配对码。
2. 打开 Hermes CLI，输入以下命令，将配对码替换为你的实际配对码：

    ```bash
    hermes pairing approve discord {你的配对码}
    ```

3. 授权成功后，你就可以在 Discord 中与助手聊天了。如果在频道中发言，需要先 @ 提及机器人。

## 将任务委派给子代理

子代理可以独立处理一个任务，并把结果返回给主对话。advanced-delegation 插件提供了下文使用的 `/delegation` 命令。

:::info 首次安装插件
当插件缺失时，Chart 1.3.51 会自动安装 advanced-delegation 插件。这次安装需要访问 GitHub。如果你之前禁用过该插件，请在开始新对话前，在 Hermes CLI 中运行 `hermes plugins enable advanced-delegation`。
:::

1. 在 Hermes CLI 中开始一段对话，或在 Hermes Dashboard 中打开 **Chat**。
2. 输入以下命令，让新会话使用默认的委派模式：

   ```text
   /delegation default
   ```

3. 输入 `/new` 以该设置开始新会话。在默认模式下，子代理使用你的主模型，除非你配置了单独的委派模型。
4. 让助手委派一个小任务。例如：

   ```text
   让子代理核对一笔订单总额：137 件商品，每件 283。完成后报告总额。
   ```

   助手会启动子代理，任务完成后把结果返回到对话中。你也可以在 Hermes Webapp 中跟进委派的任务。

:::details 选择命名模型路由
要配置单独的模型路由，请在 Hermes CLI 中运行 `hermes delegation config`。切换到 `model_map` 之前，至少先配置一条路由。如需通过 Router 接入，使用前面获取的 **Apps in Olares** Base URL，模型名填 `default-chat`。如果需要密钥，填写 `olares`。

在对话中，`/delegation model_map` 会保存该模式供下一个会话使用。`/delegation model_map --now` 还会将其应用到当前对话。用 `/delegation default` 可以切回默认模式。

路由编辑器和更多选项，请参阅 [advanced-delegation 官方指南](https://hermes-agent.nousresearch.com/marketplace/advanced-delegation)。
:::

## 用 Hermes Agent 管理 Olares

Hermes Agent 可以直接使用预装的 Olares CLI 技能，帮你管理 Olares 设备上的文件和应用程序。例如，列出文件、查看日志，或者从应用市场中安装应用。

1. 从启动台打开 Hermes CLI。
2. 依次运行下面两条命令，确认 olares-cli 和它的技能都已经正确安装并启用：

   ```bash
   olares-cli -v
   hermes skills list
   ```
   
3. 使用技能之前，先登录你的 Olares 账号。将 `{你的olares-ID}` 替换为你的 Olares ID，例如 olaresdemo@olares.com：

   ```bash
   olares-cli profile login --olares-id {你的olares-ID}
   ```

4. 根据提示输入你的 Olares 登录密码。注意：输入的内容是隐藏的。
5. 现在你可以进入 TUI 聊天界面，让助手与 Olares 环境互动了。例如，让它从应用市场安装一个应用。

## 通过 SSH 管理内网主机

你还可以让 Hermes 通过 SSH 在局域网主机上运行命令和脚本。有关密钥配置、单机和多机磁盘空间检查示例，以及连接排障，请参阅[使用 Hermes Agent 通过 SSH 管理局域网主机](hermes-ssh.md)。

## 将 Hermes Agent 接入其他应用

Hermes Agent 提供了兼容 OpenAI 的 API，你可以将其接入 Open WebUI 或 Hermes Workspace 等其他应用。

### 第 1 步：开启 Gateway API

1. 打开设置，进入**应用** > **Hermes Agent** > **管理环境变量**。

    ![Hermes Agent 环境变量](/images/zh/manual/use-cases/hermes-env-var.png#bordered){width=70%}

2. 找到 **API_SERVER_ENABLED**，点击 <i class="material-symbols-outlined">edit_square</i>，将其值设为 `true`。
3. 点 **确认**。
4. 找到 **HERMES_API_SERVER_KEY**，点击 <i class="material-symbols-outlined">edit_square</i>，填写一个密钥。密钥需要满足以下要求：

    - 长度至少 16 个字符
    - 只能包含字母、数字和常见符号
    - 不能是常见的占位符，比如 `your_API_key`
  
5. 点击**确认**，然后点击**应用**。

### 第 2 步：获取 Gateway API 的端点

1. 进入**应用** > **Hermes Agent** > **Hermes Gateway API**。
2. 复制端点 URL。例如：
   ```
   https://baf3d7172.olaresdemo.olares.com
   ```

### 第 3 步：验证 Gateway API 是否正常运行

在浏览器中访问刚才获取的 Gateway API URL，后面加上 `/health`：

```
https://{Hermes-Gateway-API-URL}/health
```

如果 API 已成功启用，页面会返回服务正常的响应，例如 `{"status": "ok", "platform": "hermes-agent"}`。

### 第 4 步：接入应用

下面以 Open WebUI 为例，演示如何对接。

1. 在 Open WebUI 里，点击用户头像，选择 **Admin Panel**。
2. 进入 **Settings** > **Connections**。
3. 在 **Manage OpenAI API Connections** 右侧，点击 <i class="material-symbols-outlined">add</i> 添加一个新连接。

    ![在 Open WebUI 里配置 Hermes Agent](/images/manual/use-cases/hermes-integrate-openwebui.png#bordered)

4. 按照下面这样填：

    - **API Base URL**：填写你的 Hermes Gateway API URL，末尾加上 `/v1`。例如 `https://baf3d7172.olaresdemo.olares.com/v1`。
    - **Auth**：选 **Bearer**，然后填写你之前设置的 `HERMES_API_SERVER_KEY`。

5. 点击<i class="material-symbols-outlined">refresh</i> 测试连接，然后点击 **Save**。
6. 进入 **New chat** 页面，查看模型下拉列表中是否出现 **hermes-agent**。如果出现，就可以开始聊天了。

    ![在 Open WebUI 里配置完 Hermes Agent](/images/zh/manual/use-cases/hermes-openwebui-integrated.png#bordered)

## 高级配置

如果想手动调整参数，可以直接编辑配置文件，然后重启 Hermes CLI 使更改生效。

1. 从启动台打开文件管理器。
2. 进入**数据** > **hermesagent** > **home**，找到配置文件，比如 `config.yaml` 和 `.env`。

文件的结构和配置选项与官方默认配置一致。具体参数的含义，可参阅 [Hermes 官方配置指南](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/configuration)。

## 常见问题

### 如何手动重启 Hermes 网关？

可以用下面几种方法之一手动重启：

- **使用 Hermes CLI**

    此方法速度最快。从启动台打开 Hermes CLI，运行以下命令：

    ```bash
    restart-gateway
    ```
- **使用 Hermes 仪表盘**

    此方法最为直观。从启动台打开 Hermes 的仪表盘，然后在左侧边栏中，选择 **Restart Gateway**。

- **使用控制面板**

    从启动台打开控制面板，进入**浏览** > **{用户名}** > **hermesagent-{用户名}** > **部署** > **hermesagent**，然后点击右上角的**重启**。此方法会完全重启整个服务，耗时稍长。

### 配置 API 后网关无法启动怎么办？

如果配置的 API 密钥不符合 Hermes 的安全要求，网关将无法启动。此时：
- Gateway API 会返回 `upstream connect error`。
- 系统日志中会出现 `[Api_Server] Refusing to start: API_SERVER_KEY is set to a placeholder value`。

解决办法是重新设置 `HERMES_API_SERVER_KEY`：
1. 打开设置，进入**应用** > **Hermes Agent** > **管理环境变量**。
2. 为 `HERMES_API_SERVER_KEY` 设置一个新值，需满足以下条件：

    - 长度至少 16 个字符
    - 只能包含字母、数字和常见符号
    - 不能是常见的占位符，比如 `your_API_key`
    
3. 点击**保存**，再点击**应用**。

## 了解更多

:::info 关于 sudo 命令的说明
Hermes 容器以普通用户运行，不适用需要在容器内执行本地 `sudo` 的操作说明。通过 SSH 执行的命令取决于远端账号的权限，详见[通过 SSH 管理局域网主机](hermes-ssh.md)。
:::

- [CLI 界面](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/cli)：了解高级命令行操作。
- [消息网关](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/messaging/)：接入 Slack、Telegram 等平台。
- [Skills 系统](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/features/skills)：安装和管理技能。
- [工具与工具集](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/features/tools)：配置助手可用的工具。
- [技巧与最佳实践](https://hermes-agent.nousresearch.com/docs/zh-Hans/guides/tips)：优化助手的配置和工作流程。
- [在 Discord 中创建服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)：为机器人创建服务器。
- [查找 Discord ID](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID)：获取配置机器人所需的用户、服务器和频道 ID。
- [OpenClaw](/zh/use-cases/openclaw.md)：在 Olares 上配置另一款 AI 助手。
