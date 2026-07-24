---
outline: [2, 3]
description: 在 Olares 上使用 Obsidian LiveSync 托管私有 CouchDB 同步后端，让 Obsidian 笔记库在电脑和手机之间保持同步。
head:
  - - meta
    - name: keywords
      content: Olares, Obsidian LiveSync, Obsidian, Self-hosted LiveSync, CouchDB, Markdown notes, knowledge management, vault sync, cross-device sync
app_version: "1.0.16"
doc_version: "1.0"
doc_updated: "2026-07-24"
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../use-cases/obsidian-livesync.md)为准。
:::

# 使用 Obsidian LiveSync 同步 Obsidian 笔记

Obsidian 是一款灵活的 Markdown 笔记应用，适用于私密笔记、双链知识库、日记和项目资料管理。Olares 上的 Obsidian LiveSync 为 Self-hosted LiveSync 社区插件提供自托管 CouchDB 后端，让你的笔记库可以跨设备同步，而无需依赖 Obsidian Sync 或其他云服务。

本指南使用电脑作为主设备，手机作为第二台设备。示例笔记库命名为 `Olares`。

## 学习目标

在本指南中，你将学习如何：

- 在 Olares 上安装 Obsidian LiveSync。
- 为 Obsidian 笔记库创建 CouchDB 数据库。
- 在主 Obsidian 设备上连接 Self-hosted LiveSync 插件。
- 在另一台设备上导入同一套同步配置。
- 启用 LiveSync 模式并检查跨设备同步结果。

## 准备工作

- 已在电脑和手机上安装 Obsidian。你可以从[官方网站](https://obsidian.md/download)下载。
- 一个用于保存主 Obsidian 笔记库的本地文件夹。
- 已在电脑和手机上安装 LarePass。

## 安装 Obsidian LiveSync

1. 打开 Market 并搜索 "Obsidian LiveSync"。

   ![Market 中的 Obsidian LiveSync](/images/manual/use-cases/obsidian-livesync.png#bordered)

2. 点击 **Get**，然后点击 **Install**，等待安装完成。

## 在 Olares 上配置 Obsidian LiveSync

在配置 Obsidian 客户端之前，先在 Obsidian LiveSync 中创建数据库，并允许 Self-hosted LiveSync 插件访问该数据库。

### 创建数据库

1. 从 Launchpad 打开 Obsidian LiveSync。

2. 使用默认凭据登录：

   - **Username**：`admin`
   - **Password**：`password`

   ![登录 Obsidian LiveSync](/images/manual/use-cases/obsidian-livesync-sign-in.png#bordered)

3. 修改默认密码（可选）：

   a. 在左侧边栏中，点击用户管理图标。

   ![在 Obsidian LiveSync 中修改密码](/images/manual/use-cases/obsidian-change-password.png#bordered)

   b. 在 **Change Password** 标签页中，输入并确认新密码，然后点击 **Change**。

4. 在左侧边栏中，点击 <i class="material-symbols-outlined">database</i> 图标。

5. 点击右上角的 **Create Database**。

6. 输入名称以创建数据库。

   :::warning 数据库命名规则
   只能使用小写字母、数字和连字符。Self-hosted LiveSync 依赖 CouchDB，而 CouchDB 严格要求数据库名使用小写字母。如果使用大写字母，后续同步可能会静默失败。
   :::

   本示例创建名为 `olares` 的数据库。

   ![创建 Obsidian LiveSync 数据库](/images/manual/use-cases/obsidian-livesync-create-database.png#bordered)

数据库创建完成后，Obsidian LiveSync 会跳转到数据库配置页面。

### 允许插件访问

Self-hosted LiveSync 插件会直接连接到 Obsidian LiveSync 托管的 CouchDB endpoint。更新应用入口配置，让插件可以通过 LarePass VPN 连接。

1. 打开 Olares 设置，然后进入 **Applications**、**Obsidian LiveSync**、**Entrances**。

2. 将 **Authentication level** 设置为 **Internal**，然后点击 **Submit**。

   ![设置 Obsidian LiveSync 认证级别](/images/manual/use-cases/obsidian-livesync-authentication-level.png#bordered)

   :::warning Public 访问风险
   将 **Authentication level** 设置为 **Public** 也可以让连接正常工作，但任何拿到 endpoint 的人都可以尝试连接 CouchDB。除非有明确需要，否则请保持 **Authentication level** 为 **Internal**。
   :::

3. 从同一个入口页面复制 **Endpoint** URL。你稍后会在 Obsidian 中把它作为 CouchDB URL 使用。

   endpoint 通常类似于：

   ```text
   https://8591294e.{username}.olares.com
   ```

## 配置主设备

使用电脑作为主设备。这台设备会初始化远程数据库，因此请从你希望作为权威数据源的笔记库开始。

### 准备 Obsidian

:::info 需要开启 LarePass VPN
配置桌面端 Obsidian 前，请在 LarePass 桌面端启用 LarePass VPN。
:::

1. 在电脑上打开 Obsidian。
2. 创建或打开要同步的笔记库。本指南会在笔记库选择页面创建一个名为 `Olares` 的新笔记库：

   a. 在 **Create new vault** 旁，点击 **Create**。

   ![创建笔记库](/images/manual/use-cases/obsidian-vault-selection.png#bordered)

   b. 在 **Vault name** 中输入 `Olares`。

   c. 点击 **Browse**，然后选择用于保存笔记库的本地文件夹。

   d. 点击 **Create**。

3. 在 Obsidian 中，进入 **Settings**、**Community plugins**，然后点击 **Browse**。

   ![浏览插件](/images/manual/use-cases/obsidian-livesync-plugin-browse.png#bordered)

4. 搜索 "Self-hosted LiveSync"，然后点击 **Install**。
5. 点击 **Enable** 以启用插件。

   ![在桌面端安装 Self-hosted LiveSync](/images/manual/use-cases/obsidian-livesync-enable-plugin.png#bordered)

6. 出现提示时，点击 **No, please take me back**，退出插件设置向导。

### 连接到 CouchDB

1. 进入 **Settings**、**Self-hosted LiveSync**。
2. 点击远程配置图标。它位于插件设置页面顶部，从左数第 4 个图标。
3. 在 **E2EE Configuration** 下，点击 **Configure And Change Remote**。

   ![Configure And Change Remote](/images/manual/use-cases/obsidian-livesync-configure-change-remote.png#bordered)

4. 在 **End-to-End Encryption** 中，选择是否启用加密：

   - 如需启用，选择 **End-to-End Encryption**，输入加密口令，并将口令保存在安全位置。
   - 如需跳过，保持 **End-to-End Encryption** 未选中。
   - 可选：如果还希望在远程服务器上隐藏文件路径、大小、时间戳等元数据，选择 **Obfuscate Properties**。

5. 点击 **Proceed**。
6. 在 **Enter Server Information** 中，选择 **CouchDB**，然后点击 **Continue to CouchDB setup**。

   ![选择 CouchDB 作为同步服务器](/images/manual/use-cases/obsidian-livesync-select-couchdb.png#bordered){width=70%}

7. 在 **CouchDB Configuration** 中，输入数据库连接信息：

   - **URL**：输入 **Settings**、**Applications**、**Obsidian LiveSync**、**Entrances** 中的 Obsidian LiveSync endpoint URL。例如：
     ```text
     https://8591294e.{username}.olares.com
     ```
   - **Username**：输入 Obsidian LiveSync 用户名。如果保留默认值，输入 `admin`。
   - **Password**：输入 Obsidian LiveSync 密码。
   - **Database Name**：输入之前创建的数据库名称。本示例中为 `olares`。

      ![在 Self-hosted LiveSync 中配置 CouchDB](/images/manual/use-cases/obsidian-livesync-configure-couchdb.png#bordered){width=70%}

8. 点击 **Test Settings and Continue**。如果连接配置正确，Obsidian 会进入下一步设置。

:::tip 稍后配置加密
如果你在远程设置向导中跳过了端到端加密，也可以稍后再配置。打开 **Settings**、**Self-hosted LiveSync**，点击远程配置图标，然后点击 **E2EE Configuration** 下的 **Configure**。

在 **End-to-End Encryption** 对话框中，选择 **End-to-End Encryption**，输入加密口令，可选选择 **Obfuscate Properties**，然后保存配置。

连接到同一同步目标的每台设备都必须使用相同的端到端加密设置和加密口令。口令不匹配可能会导致同步数据无法读取。
:::

### 初始化服务器

:::warning 数据覆盖
初始化步骤会覆盖远程服务器上的现有数据，并基于当前设备重建服务器。仅在使用新的 Obsidian LiveSync 数据库，或你明确希望当前设备成为权威数据源时执行这些步骤。继续前，请将 Obsidian 笔记库文件夹复制到安全位置。
:::

1. 出现 **Mostly Complete: Decision Required** 时，选择 **I am setting up a new server for the first time / I want to reset my existing server**，然后点击 **Proceed to the next step**。

   本指南中，远程数据库是新数据库，因此主设备应使用当前笔记库数据初始化服务器。

   ![从主设备初始化远程服务器](/images/manual/use-cases/obsidian-livesync-initialize-server.png#bordered){width=70%}

2. 出现 **Setup Complete: Preparing to Initialise Server** 时，点击 **Restart and Initialise Server**。

   Obsidian 会重启，并将当前设备上的笔记库数据作为主副本上传到服务器。

   ![重启并初始化 LiveSync 服务器](/images/manual/use-cases/obsidian-livesync-restart-initialise-server.png#bordered){width=70%}

3. 出现 **Final Confirmation: Overwrite Server Data with This Device's Files** 时，确认使用当前设备覆盖服务器：

   a. 仔细阅读警告。

   b. 勾选全部 3 个确认复选框。

   c. 在 **Have you created a backup before proceeding?** 下，选择 **I have created a backup of my Vault**。

   d. 点击 **I Understand, Overwrite Server**。

   ![确认覆盖 LiveSync 服务器](/images/manual/use-cases/obsidian-livesync-overwrite-server-confirmation.png#bordered){width=70%}

### 启用 LiveSync 模式

Obsidian 重启并完成服务器初始化后，回到 Self-hosted LiveSync 设置页面。

1. 打开 **Settings**、**Self-hosted LiveSync**。
2. 点击同步配置图标。它位于插件设置页面顶部，从左数第 5 个图标。
3. 将 **Sync Mode** 设置为 **LiveSync**。

   ![启用 LiveSync](/images/manual/use-cases/obsidian-livesync-enable-livesync.png#bordered){width=70%}

:::info 妥善保存加密口令
如果启用了端到端加密，请将加密口令保存在安全位置。添加另一台设备时，需要使用同一个口令。
:::

## 配置另一台设备

使用主设备上的 setup URI，在手机上导入同一套 LiveSync 配置。

### 准备手机

:::info 需要开启 LarePass VPN
配置移动端 Obsidian 前，请在 LarePass 移动端启用 LarePass VPN。
:::

1. 在手机上打开 Obsidian，并为同步笔记创建一个新的本地笔记库。
2. 在 **Settings**、**Community plugins** 中安装并启用 Self-hosted LiveSync。操作流程与主设备相同。

### 从主设备复制 setup URI

1. 在电脑上打开 **Settings**、**Self-hosted LiveSync**。
2. 点击快速设置图标。它位于插件设置页面顶部，从左数第 2 个图标。
3. 在 **To setup other devices** 中，点击 **Copy**。
4. 输入用于加密 setup URI 的密码，然后点击 **Ok**。Obsidian 会将 setup URI 复制到剪贴板。

   ![从主设备复制 setup URI](/images/manual/use-cases/obsidian-livesync-copy-setup-uri.png#bordered)

5. 通过安全渠道将 setup URI 发送到手机。

### 加入现有服务器

1. 在手机上打开 **Settings**、**Self-hosted LiveSync**。
2. 点击快速设置图标。
3. 在 **Connect with Setup URI** 旁，点击 **Use**。
4. 粘贴 setup URI，输入用于加密它的密码，然后确认。
5. 如果 Obsidian 询问如何处理远程服务器，选择 **My remote server is already set up. I want to join this device**。

   这样第二台设备会从服务器拉取现有同步数据，而不是覆盖服务器。

6. 打开 **Settings**、**Self-hosted LiveSync**。
7. 点击同步配置图标。
8. 将 **Sync Mode** 设置为 **LiveSync**。

当两台设备都使用 LiveSync 模式后，在一个笔记库中所做的更改会实时同步到另一台设备。

## 检查同步结果

1. 在主设备上创建或编辑一条笔记。
2. 等待片刻。
3. 在手机上打开同一个笔记库。该笔记应显示最新更改。
4. 在手机上编辑这条笔记，然后检查该更改是否出现在电脑上。你可以在右上角状态栏中查看同步状态。

   ![检查 Obsidian LiveSync 同步结果](/images/manual/use-cases/obsidian-livesync-sync-result.png#bordered)

:::tip 首次同步
如果笔记库包含大量笔记或附件，首次同步可能需要更长时间。请让两台设备保持在线，直到首次同步完成。
:::

## 了解更多

- [Obsidian Help](https://help.obsidian.md)：Obsidian 功能和设置的官方指南。
- [Self-hosted LiveSync repository](https://github.com/vrtmrz/obsidian-livesync)：插件文档和故障排查说明。
