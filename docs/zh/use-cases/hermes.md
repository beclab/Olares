---
outline: [2,3]
description: 在 Olares 上安装 Hermes Agent，将其接入 Discord，安装 Olares 技能，并通过 Gateway API 与其他应用打通。
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, Hermes Agent, 自主 AI, 自我进化 AI, Discord 机器人, 自托管
app_version: "1.3.9"
doc_version: "2.0"
doc_updated: "2026-06-04"
---

# 用 Hermes 搭建一个自主工作的 AI 助手

Hermes Agent 是一个能够自主执行任务的 AI 助手。接入本地模型后，它可以执行系统任务、编写代码、管理工作流程。它能够在多次对话之间保持记忆，还能根据与你的互动，自动生成可复用的技能。

在 Olares 中，你可以为助手安装专门的 Olares 技能，让它管理设备上的文件和应用程序。你还可以通过 Discord 等聊天软件远程与它对话，或将其接入 Open WebUI 等其他应用。

## 学习目标

- 在 Olares 上安装 Hermes Agent
- 配置 Hermes Agent，连接本地模型
- 直接在终端中与它对话
- 接入 Discord，实现远程聊天
- 安装 Olares 技能，让助手帮你管理文件和应用程序
- 开启 Gateway API，将 Hermes Agent 与 Open WebUI 等应用对接

## 前提条件

开始之前，确保满足以下要求：
- **本地模型**：你已经安装好 [Ollama](./ollama.md)，并至少下载了一个能调用工具的模型，且模型正在运行。这篇教程将以 `qwen3.5:9b` 为例。
- **Discord 账号**：用来创建机器人应用。
- **Discord 服务器**：确保你在这个服务器上有添加机器人的权限。

## 安装 Hermes Agent

1. 打开应用市场，搜索 “Hermes”。

   ![安装 Hermes Agent](/images/zh/manual/use-cases/hermes-agent.png#bordered)

2. 点击**获取**，然后点击**安装**。安装完成后，桌面上会出现两个快捷方式：

    - Hermes CLI：命令行界面
    - 仪表盘：图形面板

    ![Hermes 入口](/images/zh/manual/use-cases/hermes-entry-points.png#bordered){width=50%}

:::tip 运行多个 Hermes 实例
Olares 支持克隆应用。如果你想同时运行多个独立的 AI 助手，分别处理不同的任务，可以克隆 Hermes Agent 这个应用。具体步骤，见[克隆应用](/zh/manual/olares/market/clone-apps.md)。
:::

## 配置 Hermes Agent

按照快速设置向导，将 Hermes Agent 连接到你本地的模型。

### 第 1 步：获取模型和端点信息

1. 从桌面打开 Ollama 应用。
2. 运行以下命令，查看已安装的模型：

    ```bash
    ollama list
    ```
3. 复制 **NAME** 列中的模型名称并保存，例如 `qwen3.5:9b`。
4. 运行以下命令，查看模型的上下文窗口大小：

    ```bash
    ollama ps
    ```

5. 复制 **CONTEXT** 列中的数值并保存，例如 `32768`。
6. 打开设置，进入**应用** > **Ollama** > **共享入口** > **Ollama API**。

    ![获取 Ollama API 地址](/images/zh/manual/use-cases/ollama-endpoint1.png#bordered){width=70%}
    
7. 复制端点地址并保存，例如 `http://d54536a50.shared.olares.com`。

### 第 2 步：运行设置向导

1. 从桌面打开 Hermes CLI。
2. 输入以下命令，启动配置向导：

   ```bash
   hermes setup
   ```

3. 向导会逐步引导你完成设置。使用方向键移动，按**回车**键确认。

    | 配置   | 选项   |
    |:-----------|:---------|
    | How would you like to set up Hermes | 选择 **Quick setup - provider, model & messaging (recommended)**。 |
    | Select provider | 选择 **Custom endpoint (enter URL manually)**。  |
    | API base URL  | 输入你模型的端点地址，末尾加上 `/v1`。<br>例如 `http://d54536a50.shared.olares.com/v1`。  |
    | API key  | 可填写任意占位值，例如 `ollama-local`。<br>输入的内容是隐藏的。 |
    | Available models | 从列表中找到你的目标模型，输入其对应的编号。 |
    | Context length in tokens | <ul><li>如果你模型的上下文窗口小于 `65536`，<br>填写一个大于 `65536` 的数。</li><li>如果你模型的上下文窗口大于 `65536`，<br>留空，让系统自动检测。</li></ul> |
    | Display name | 填写一个便于识别的名称，例如 `ollama-local`。|
    | Connect a messaging platform | 选择 **Skip - set up later with `hermes setup gateway`**。 |

4. 请勿关闭 Hermes CLI 窗口，下一步还会用到。

## 与 Hermes Agent 对话

### 方式 1：在文本用户界面中对话

文本用户界面（TUI）直接在 Hermes CLI 里运行，无需额外配置，适合快速体验。

1. 设置向导完成后，会询问是否启动 `hermes chat`。输入 `y` 并按**回车**键，即可进入文本用户界面。

   :::tip
   如果你退出了向导，也可以在 Hermes CLI 里手动输入 `hermes chat` 命令来启动文本用户界面聊天。
   :::

    ![Hermes 设置完成](/images/manual/use-cases/hermes-setup-complete.png#bordered)

2. 发一条消息，例如 `你当前使用的模型是哪个`，查看助手是否回复正常。

    ![Hermes TUI 聊天界面](/images/manual/use-cases/hermes-tui.png#bordered)

3. 如需退出 TUI 回到普通命令行，输入 `/exit` 并按**回车**键。

### 方式 2：通过 Discord 远程聊

如果你想在手机或其他设备上与助手聊天，可以将其接入 Discord 机器人。

#### 第一步：创建一个 Discord 机器人

1. 用你的 Discord 账号登录 [Discord Developer Portal](https://discord.com/developers/home)。
2. 从左侧边栏选择 **APP**，然后点击**新 APP**。

    ![Discord 开发者门户新建应用](/images/zh/manual/use-cases/hermes-new-app.png#bordered)

3. 为新应用命名（例如 `Cool-hermes`），选择同意条款，然后点击**创建**。

    ![创建一个新 APP 窗口](/images/zh/manual/use-cases/hermes-create-app.png#bordered){width=40%}

4. 从左侧边栏，选择**机器人**。
5. 向下滚动至 **Privileged Gateway Intents** 部分，打开下面三个选项：

    - Presence Intent
    - Server Members Intent
    - Message Content Intent

    ![特权网关意图](/images/zh/manual/use-cases/hermes-privileged-gateway-intents.png#bordered)

6. 点击**保存更改**。
7. 向上滚动到**令牌**部分，点击**重置令牌**，然后复制生成的令牌。后面在 Hermes CLI 里配置消息平台时需要用到。

    ![重置令牌](/images/zh/manual/use-cases/hermes-reset-token.png#bordered)

#### 第二步：将机器人添加到你的服务器

1. 从左侧边栏中，选择 **OAuth2**，找到 **OAuth2 URL 生成器**部分：

    a. 在**范围**区域，勾选 **bot** 和 **applications.commands**。

    ![OAuth2 URL 生成器](/images/zh/manual/use-cases/hermes-oauth21.png#bordered)

    b. 向下滚动页面至**机器人权限**部分，按下面图示勾选权限。后续可以再调整。

    ![机器人权限](/images/zh/manual/use-cases/hermes-bot-permissions.png#bordered)

2. 复制页面最底部**已生成的 URL**。
3. 将这个 URL 粘贴到浏览器的新标签页中，在**添加至服务器**的下拉框中，选择你的 Discord 服务器，点击**继续**，再点击**授权**。

    ![将 Discord 机器人添加到服务器](/images/zh/manual/use-cases/hermes-add-bot.png#bordered)

    添加完成后，机器人就会出现在你对应的服务器中。

#### 第三步：配置消息平台

配置并运行 Hermes gateway，将 Hermes Agent 接入你的 Discord 机器人。

1. 打开 Hermes CLI，输入以下命令，启动配置向导：

   ```bash
   hermes gateway setup
   ```

2. **Select a platform to configure** 一项，选择 **Discord**。
3. 按提示填好机器人集成信息：

   - **Discord bot token**：填写你从 Discord Developer Portal 复制的令牌。注意，你的输是隐藏的。
   - **Allowed user IDs or usernames**：填写你的 Discord 用户 ID，仅限你自己使用。
   - （可选）**Home channel ID**：填写机器人所在频道的 ID。

4. 选择 **Done**。
5. 当提示 **Restart the gateway to pick up changes** 时，输入 `y`。
6. 前往 Discord 查看机器人状态。如果图标为绿色，表示机器人已在线，说明网关重启成功、配置正确。

#### 第四步：授权你的账号

出于安全考虑，机器人不会回应未授权用户。你需要将自己的 Discord 账号与机器人配对。

1. 向你的新机器人发送一条私信。机器人会回复一条错误消息，其中包含一个配对码。
2. 打开 Hermes CLI，输入以下命令，将配对码替换为你的配对码：

    ```bash
    hermes pairing approve discord {你的配对码}
    ```

3. 授权成功后，你就可以在 Discord 中与助手聊天了。如果在频道中发言，需要先 @ 提及机器人。

## 用 Hermes Agent 管理 Olares

给 Hermes Agent 安装 Olares CLI 技能后，它就能帮你管理 Olares 设备上的文件和应用程序。例如，列出文件、查看日志，或者从应用市场中安装应用。

1. 从桌面打开 Hermes CLI。
2. 依次运行下面两条命令，确认 olares-cli 和它的技能都已经正确安装并启用：

   ```bash
   olares-cli -v
   hermes skills list
   ```
   
3. 使用技能之前，先登录你的 Olares 账号。将 `{your-olares-id}` 替换为你的 Olares ID，例如 olaresdemo@olares.com：

   ```bash
   olares-cli profile login --olares-id {你的olares-ID}
   ```

4. 根据提示输入你的 Olares 登录密码。注意，你的输入是隐藏的。
5. 现在你可以进入 TUI 聊天界面，让助手与 Olares 环境互动了。例如，让它从应用市场安装一个应用。

## 将 Hermes Agent 接入其他应用

Hermes Agent 提供了兼容 OpenAI 的 API，你可以将其接入 Open WebUI 或 Hermes Workspace 等其他应用。

### 第 1 步：开启 Gateway API

1. 打开设置，进入 **应用** > **Hermes Agent** > **管理环境变量**。

    ![Hermes Agent 环境变量](/images/zh/manual/use-cases/hermes-env-var.png#bordered){width=70%}

2. 找到 **API_SERVER_ENABLED**，点击 <i class="material-symbols-outlined">edit_square</i>，将其值设为 `true`。
3. 点 **确认**。
4. 找到 **HERMES_API_SERVER_KEY**，点击 <i class="material-symbols-outlined">edit_square</i>，填写一个密钥。密钥需要满足以下要求：

    - 长度至少 8 个字符
    - 只能包含字母、数字和常见符号
    - 不能是常见的占位符，比如你的 API 密钥
  
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

1. 从桌面打开文件管理器。
2. 进入**数据** > **hermesagent** > **home**，找到配置文件，比如 `config.yaml` 和 `.env`。

文件的结构和配置选项与官方默认配置一致。具体参数的含义，可参阅 [Hermes 官方配置指南](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/configuration)。

## 常见问题

### 如何手动重启 Hermes 网关？

可以用下面几种方法之一手动重启：

- **使用 Hermes CLI**

    此方法速度最快。从桌面打开 Hermes CLI，运行以下命令：

    ```bash
    restart-gateway
    ```
- **使用 Hermes 仪表盘**

    此方法最为直观。从桌面打开 Hermes 的仪表盘，然后在左侧边栏中，选择 **Restart Gateway**。

- **使用控制面板**

    从桌面打开控制面板，进入**浏览** > **{用户名}** > **hermesagent-{用户名}** > **部署** > **hermesagent**，然后点击右上角的**重启**。此方法会完全重启整个服务，耗时稍长。

### 配置 API 后网关无法启动怎么办？

如果配置的 API 密钥不符合 Hermes 的安全要求，网关将无法启动。此时：
- Gateway API 会返回 `upstream connect error`。
- 系统日志中会出现 `[Api_Server] Refusing to start: API_SERVER_KEY is set to a placeholder value`。

解决办法是重新设置 `HERMES_API_SERVER_KEY`：
1. 打开设置，进入 **应用** > **Hermes Agent** > **管理环境变量**。
2. 为 `HERMES_API_SERVER_KEY` 设置一个新值，需满足以下条件：

    - 长度至少 8 个字符
    - 只能包含字母、数字和常见符号
    - 不能是常见的占位符，比如你的 API 密钥
    
 3. 点击**保存**，再点击**应用**。

## 后续步骤

基础设置完成后，可以查阅官方文档，进一步扩展助手的能力：
- 高级命令行操作，见 [CLI 界面](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/cli)
- 接入 Slack、Telegram 等平台，见[消息网关](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/messaging/)
- 安装技能，见 [Skills 系统](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/features/skills)
- 安装工具，见[工具与工具集](https://hermes-agent.nousresearch.com/docs/zh-Hans/user-guide/features/tools)
- 最佳实践和优化技巧，见[技巧与最佳实践](https://hermes-agent.nousresearch.com/docs/zh-Hans/guides/tips)

:::info 关于 sudo 命令的说明
Olares 环境里的 Hermes CLI 不支持 `sudo` 命令。官方文档中需要 `sudo` 权限的步骤，请忽略。
:::

## 了解更多

- [如何在 Discord 里创建一个服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
- [如何找到频道 ID 编号](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID)
- [OpenClaw](/zh/use-cases/openclaw.md)
