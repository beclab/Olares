---
connectionVersion: "1.12.7"
connectionLatestPath: /zh/use-cases/hermes
outline: [2,3]
description: 在 Olares 上安装 Hermes Agent，接入本地模型和 Discord，通过 SSH 管理局域网内的多台主机，并通过 Gateway API 与其他应用连接。
head:
  - - meta
    - name: keywords
      content: Olares, Hermes, Hermes Agent, 自主 AI, 自我进化 AI, Discord 机器人, 自托管, SSH, 内网主机管理
app_version: "1.3.49"
doc_version: "3.2"
doc_updated: "2026-10-21"
---

# 用 Hermes 搭建一个自主工作的 AI 助手

<VersionRouteSelect />

Hermes Agent 是一个能够自主执行任务的 AI 助手。接入本地模型后，它可以执行系统任务、编写代码、管理工作流程。它能够在多次对话之间保持记忆，还能根据与你的互动，自动生成可复用的技能。

在 Olares 中，你可以使用内置的 Olares 技能，让助手管理设备上的文件和应用程序。你还可以通过 Discord 等聊天软件远程与它对话，或将其接入 Open WebUI 等其他应用。

## 学习目标

- 在 Olares 上安装 Hermes Agent
- 配置 Hermes Agent，连接本地模型
- 直接在文本用户界面中与它对话
- 接入 Discord，实现远程聊天
- 使用 Olares 技能，让助手帮你管理文件和应用程序
- 通过 SSH 在内网主机上执行命令、运行脚本，并协调多台主机上的任务
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

2. 点击**获取**，然后点击**安装**。安装完成后，桌面上会出现两个快捷方式：

    - Hermes CLI：命令行界面
    - 仪表盘：图形面板

    ![Hermes 入口](/images/zh/manual/use-cases/hermes-entry-points.png#bordered){width=50%}

:::tip 运行多个 Hermes 实例
Olares 支持克隆应用。如果你想同时运行多个独立的 AI 助手，分别处理不同的任务，可以克隆 Hermes Agent 这个应用。具体步骤，见[克隆应用](/zh/manual/olares/market/clone-apps.md)。
:::

## 配置 Hermes Agent

按照快速设置向导，将 Hermes Agent 连接到你本地的模型。

### 第 1 步：获取模型连接信息

<!--@include: ../reusables/ai-service-connections.md#model-connection-overview-->

<!--@include: ../reusables/ai-service-connections.md#get-model-connection-details-->

5. 在 Router 的 **LLM** 页面，点击 Qwen3.8-27B (llama.cpp) 所在行的信息图标，打开 **Model card**。在 **Engine args** 中读取 `-c` 后的数值。例如，`-c 104448` 表示下一步应填写 `104448`。截图说明见[查看实际上下文大小](/zh/manual/best-practices/connect-ai-apps.md#check-context-window)。

### 第 2 步：运行设置向导

1. 从桌面打开 Hermes CLI。
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
    | API key  | 输入任意占位值，例如 `local`。<br>出于安全考虑，输入内容会被隐藏。 |
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

4. 保持 Hermes CLI 窗口开启，后续步骤仍需使用。

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

2. 发一条消息，例如 `你当前使用的模型是哪个`，查看助手是否回复正常。

    ![Hermes TUI 聊天界面](/images/manual/use-cases/hermes-tui1.png#bordered)

3. 如需退出 TUI 回到普通命令行，输入 `/exit` 并按**回车**键。

### 方式 B：通过 Discord 远程对话

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

## 用 Hermes Agent 管理 Olares

Hermes Agent 可以直接使用预装的 Olares CLI 技能，帮你管理 Olares 设备上的文件和应用程序。例如，列出文件、查看日志，或者从应用市场中安装应用。

1. 从桌面打开 Hermes CLI。
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

除了管理 Olares 中的文件和应用，你还可以让 Hermes 通过 SSH 连接局域网内的其他主机，执行远程命令、运行脚本、处理服务运维任务，或协调多台主机上的操作。每次在对话中指定目标主机和任务即可，无需在配置中固定一个 IP 地址。可执行的操作取决于远端账号的权限。

在 Hermes 设置中，将终端后端保持为 **Local**。这里的 Local 指 Hermes 应用容器：助手先在容器中运行 `ssh`，再由 SSH 在目标主机上执行命令。这个场景直接使用标准 SSH 命令，不需要开启 Gateway API 或额外开发 SSH 工具。

### 连接前的准备

确认以下条件：

- Olares 设备能通过网络访问目标主机的 SSH 端口，默认端口为 `22`。
- 目标主机已启用 SSH 服务，并允许你使用指定账号登录。
- Hermes 已连接可用的模型，终端后端选择 **Local**。

下面以 Linux 主机 `ubuntu@192.168.1.50` 为例。请将用户名、IP 地址和端口替换为你自己的配置。

:::tip 网络连接从哪里发起？
SSH 连接由 Olares 上的 Hermes 容器发起。即使你从 Mac 打开 Hermes，也不代表容器自动获得了 Mac 所在网络的访问能力。在本次验证环境中，无需为容器开启 hostNetwork 或特权权限即可访问内网主机。跨网段、访客网络或防火墙规则仍可能阻止连接。
:::

### 使用密码登录

目标主机允许密码认证时，可以先在 Hermes CLI 中手动登录，无需提前配置密钥。

1. 从启动台打开 Hermes CLI。如果当前处于聊天界面，输入 `/exit` 返回命令行。

2. 准备持久化的 SSH 目录，然后发起连接：

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh -o PubkeyAuthentication=no \
     -o PreferredAuthentications=password,keyboard-interactive \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50
   ```

   如果使用其他 SSH 端口，加入 `-p 端口号`。首次连接时，先向目标主机管理员核对主机指纹，确认一致后再输入 `yes`。

3. 出现 `ubuntu@192.168.1.50's password:` 提示时，输入远端账号的密码并按 **Enter**。输入内容不会显示。登录成功后，可以执行该账号有权限运行的命令；完成后输入 `exit` 退出远端会话。

这个示例用于你在 Hermes CLI 中交互登录。普通 SSH 命令不会自动回答后台任务中的密码提示；需要让助手独立执行任务时，建议按下面的步骤配置密钥登录。不要把密码直接写进命令、脚本或聊天提示词。

### 配置密钥登录

密钥登录适合由助手执行的自动化任务。完成一次公钥授权后，Hermes 可以使用私钥连接目标主机，无需每次输入账号密码。

#### 第 1 步：在持久化目录中准备密钥

1. 从启动台打开 Hermes CLI。如果当前处于聊天界面，输入 `/exit` 返回命令行。

2. 创建 SSH 目录，并生成一对密钥：

   ```bash
   mkdir -p /opt/data/.ssh
   chmod 700 /opt/data/.ssh
   ssh-keygen -t ed25519 -C "hermes-lan" -f /opt/data/.ssh/id_ed25519
   ```

   如果已经存在同名密钥，请保留原密钥，或换一个文件名，不要覆盖。自动执行的任务无法交互输入密钥口令；为此生成专用自动化密钥时，可在两次口令提示中直接按 **Enter**，将口令留空。

3. 设置私钥权限，并查看公钥：

   ```bash
   chmod 600 /opt/data/.ssh/id_ed25519
   cat /opt/data/.ssh/id_ed25519.pub
   ```

   复制输出的完整公钥行。私钥保留在应用中，不要粘贴到聊天消息里。

`/opt/data` 是 Hermes 的持久化数据目录，应用重启或容器重建后，目录中的密钥仍会保留。卸载应用或删除应用数据前，请先备份密钥。克隆应用使用独立的数据目录，需要单独配置 SSH 密钥。

:::tip 使用已有 PEM 私钥
在文件管理器中进入**数据** > **hermesagent** > **home**，显示隐藏文件，然后将已有私钥上传到 `.ssh` 目录，例如命名为 `lan-host.pem`。在 Hermes CLI 中运行：

```bash
chmod 600 /opt/data/.ssh/lan-host.pem
```

后续命令中的私钥路径改为 `/opt/data/.ssh/lan-host.pem` 即可。SSH 不要求私钥文件必须使用 `.pem` 扩展名。

如果需要新建 PEM 格式的 RSA 密钥，可以运行：

```bash
ssh-keygen -t rsa -b 3072 -m PEM -f /opt/data/.ssh/lan-host.pem
```

随后将生成的 `lan-host.pem.pub` 公钥添加到目标主机。
:::

#### 第 2 步：在目标主机上添加公钥

1. 在目标主机上，以 Hermes 将要使用的账号登录，例如 `ubuntu`。打开终端，准备该账号的 SSH 目录：

   ```bash
   mkdir -p "$HOME/.ssh"
   chmod 700 "$HOME/.ssh"
   touch "$HOME/.ssh/authorized_keys"
   chmod 600 "$HOME/.ssh/authorized_keys"
   ```

2. 用文本编辑器打开 `$HOME/.ssh/authorized_keys`，将上一步复制的公钥追加为单独一行，然后保存。保留文件中原有的公钥，并确认目录和文件归当前登录账号所有。

3. 查看目标主机的 SSH 主机密钥指纹，供首次连接时核对：

   ```bash
   ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub
   ```

#### 第 3 步：验证 SSH 连接

1. 返回 Hermes CLI，先手动连接一次：

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   首次连接会提示确认主机指纹。与目标主机上显示的指纹核对一致后，再输入 `yes`。如果 SSH 使用其他端口，在命令中加入 `-p 端口号`。

2. 再验证不需要交互输入的连接方式：

   ```bash
   ssh -i /opt/data/.ssh/id_ed25519 \
     -o IdentitiesOnly=yes \
     -o BatchMode=yes \
     -o StrictHostKeyChecking=yes \
     -o UserKnownHostsFile=/opt/data/.ssh/known_hosts \
     -o ConnectTimeout=10 \
     ubuntu@192.168.1.50 'hostname; id'
   ```

   命令应直接返回目标主机名和登录账号信息，不再要求输入密码。`BatchMode=yes` 会在无法自动认证时直接返回错误，避免任务停在密码提示处。

私钥和 `known_hosts` 都使用绝对路径，避免助手运行命令时，因主目录不同而找不到文件。主机指纹也保存在 `/opt/data/.ssh/known_hosts` 中，应用重启后无需重新确认。

### 用自然语言安排远程任务

在 Hermes 的 TUI 或 Dashboard 中发起对话，明确告诉助手目标地址、账号和密钥路径。例如：

```text
通过 SSH 连接 ubuntu@192.168.1.50，使用私钥
/opt/data/.ssh/id_ed25519 和主机指纹文件
/opt/data/.ssh/known_hosts，保留严格主机指纹校验，
使用批处理模式，连接超时设为 10 秒。
连接后，先确认目标主机和当前账号，然后根据我接下来给出的要求
执行远程命令或运行指定脚本。执行前说明操作步骤和影响范围，
完成后汇总结果；遇到权限不足或连接失败时，说明原因。
```

连接另一台主机时，先为它完成公钥授权和首次连接验证，再在对话中更换目标地址。你可以给不同主机使用不同的私钥，不需要修改 Hermes 的终端后端。

多台主机都配置好后，可以在同一次对话中安排多机任务。例如，将下面的占位内容替换为你的实际任务：

```text
在 ubuntu@192.168.1.50 和 olares@192.168.1.60 上完成以下任务：
【填写任务目标、需要执行的命令或脚本，以及主机之间的执行顺序】。
两台主机都使用 /opt/data/.ssh/id_ed25519，主机指纹文件为
/opt/data/.ssh/known_hosts，保留严格主机指纹校验。
先分别确认目标主机和登录账号，再按照指定顺序执行。
按主机名和 IP 地址整理每一步的结果；某台主机失败时，
先报告失败原因和已完成的操作，再由我决定是否继续后续步骤。
```

:::info 远端 sudo 权限
远端命令的权限取决于 SSH 登录账号。需要提权时，可以要求助手使用 `sudo -n`：如果该账号已获准免密码执行对应命令，就会直接执行；否则返回错误，不会等待输入密码。不要把 sudo 密码写进提示词。

这与在 Hermes 容器内使用 sudo 是两回事。容器本身以普通用户运行，并不妨碍 SSH 账号在远端使用已获准的 sudo 权限。
:::

### 常见连接问题

| 现象 | 检查方法 |
|:---|:---|
| 连接超时或被拒绝 | 检查 IP、SSH 端口、服务状态、防火墙，以及 Olares 到目标主机的网络路由。开启 Olares 应用入口不能解决出站 SSH 的连通性问题。 |
| `Permission denied (publickey)` | 核对远端用户名、该账号 `authorized_keys` 中的公钥和私钥路径。私钥权限应为 `600`，`.ssh` 目录权限应为 `700`。 |
| `Host key verification failed` | 使用同一个 `known_hosts` 路径手动完成首次连接。如果已保存的指纹发生变化，先在目标主机上确认原因，再更新记录，不要关闭指纹校验。 |
| `sudo: a password is required` | 当前账号没有免密码执行该命令的权限。改用账号有权执行的只读命令，或请目标主机管理员配置所需权限。 |

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
1. 打开设置，进入**应用** > **Hermes Agent** > **管理环境变量**。
2. 为 `HERMES_API_SERVER_KEY` 设置一个新值，需满足以下条件：

    - 长度至少 16 个字符
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
Hermes 容器以普通用户运行，不适用需要在容器内执行本地 `sudo` 的操作说明。通过 SSH 执行的命令取决于远端账号的权限，详见[通过 SSH 管理内网主机](#通过-ssh-管理内网主机)。
:::

## 了解更多

- [如何在 Discord 里创建一个服务器](https://support.discord.com/hc/en-us/articles/204849977-How-do-I-create-a-server)
- [如何找到频道 ID 编号](https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID)
- [OpenClaw](/zh/use-cases/openclaw.md)
