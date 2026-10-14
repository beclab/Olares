---
outline: [2, 3]
description: 使用 Pterodactyl Panel 在 Olares 上创建 Terraria 游戏服务器，与好友联机，并管理服务器文件、访问权限、定时存档和备份。
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl Panel, Wings, Terraria, TShock, 游戏服务器, 备份, 服务器管理"
app_version: "0.0.12"
doc_version: "2.2"
doc_updated: "2026-10-14"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/pterodactyl.md)为准。
:::

# 使用 Pterodactyl Panel 管理游戏服务器

Pterodactyl Panel 是一款开源的游戏服务器管理应用。在 Olares 上，你可以通过浏览器启动和停止服务器、管理文件、邀请好友协助管理，以及备份游戏世界。

本文以 Terraria（泰拉瑞亚）为例。这是一款沙盒冒险游戏，玩家可以在同一个世界里探索、建造和战斗。你将在 Olares 上运行游戏服务器，再从自己电脑上的 Terraria 进入这个世界。

## 认识各个组件

配置过程中会遇到以下名称：

| 名称 | 在本文中的作用 |
| --- | --- |
| Terraria | 你和好友玩的游戏。每位玩家在自己的电脑上运行游戏，连接到 Olares 上的同一个服务器。 |
| TShock | 为 Terraria 服务器增加管理命令、玩家权限和插件支持。本文创建的服务器会安装 TShock。 |
| Pterodactyl Panel | 创建和管理服务器的网页界面。 |
| Wings | Pterodactyl 在 Olares 上运行的后台服务，接收 Panel 的指令并运行游戏服务器。你通过 Panel 管理它。 |
| TShock Egg | 导入 Panel 的 JSON 模板，告诉 Panel 如何安装和启动带有 TShock 的 Terraria 服务器，包括使用哪个运行镜像和哪些设置。 |

你在 Panel 中导入 Egg 并创建服务器，Wings 负责在 Olares 上运行它。玩家通过 Terraria 进入游戏，你则通过 Panel 管理服务器。

:::info 已从 Market 安装 Minecraft？
Panel 管理通过 Wings 创建的服务器。从 Market 安装的 Minecraft 可以继续通过 Olares 管理。迁移方式见[已有 Minecraft 应用的说明](#panel-能管理已有的-minecraft-应用吗)。
:::

## 前提条件

- 已从 Market 安装 MariaDB。
- 有权限安装应用的 Olares 账户。
- 可访问互联网，以下载服务器软件。
- 与 Olares 位于同一局域网的电脑，已安装 Terraria 1.4.5.8。TShock 6.2.1 支持此游戏版本。

:::info 实验性集成
本地 Wings 集成属于实验性功能。请使用下文提供的 Olares 模板，其中已包含本例所需的设置和运行镜像。
:::

## 安装 Pterodactyl Panel

1. 打开 Market，搜索“Pterodactyl Panel”。
   <!-- ![Market 中的 Pterodactyl Panel](/images/manual/use-cases/pterodactyl.png#bordered) -->

2. 点击**获取**，然后点击**安装**。在安装表单中保持 **ENABLE_WINGS** 启用，设置 **ADMIN_USERNAME** 和 **ADMIN_PASSWORD**，等待安装完成。密码至少包含 8 个字符。

## 配置本地节点

节点是运行 Wings 和游戏服务器的机器。本文中，节点就是你的 Olares 设备，在 Panel 中显示为 **Olares Local Wings**。

1. 从启动台打开 Pterodactyl Panel，使用安装时设置的用户名和密码登录。

2. 进入管理员界面，打开 **Nodes** > **Olares Local Wings** > **About**。等待 **Daemon Version** 和 **System Information** 加载完成，再创建服务器。

   ![本地 Wings 节点显示版本和可用资源](/images/manual/use-cases/pterodactyl-local-node.png#bordered)

3. 打开 **Allocation**，创建 IP 地址为 `0.0.0.0`、端口为 `30000` 的端口分配。

   <!-- ![创建游戏端口分配](/images/manual/use-cases/pterodactyl-create-allocation.png#bordered) -->

端口分配指定 Wings 内部供服务器使用的端口。应用将该端口映射到 Olares 设备的 TCP `43000`。

玩家使用 Olares 设备的局域网 IP 地址和端口 `43000` 加入游戏，通过 Panel 的网页地址管理服务器。

## 创建 Terraria 服务器

### 下载并导入 Egg

根据 Olares 设备的 CPU 架构选择 Egg。Intel、AMD 的 x86-64 设备使用 amd64，采用 aarch64 处理器的设备使用 ARM64。

| 架构 | 下载 |
| --- | --- |
| amd64 | [TShock 6.2.1 Egg](https://cdn.olares.cn/apps/demos/egg-tshock-6.2.1-olares-amd64.json) |
| ARM64 | [TShock 6.2.1 Egg](https://cdn.olares.cn/apps/demos/egg-tshock-6.2.1-olares-arm64.json) |

:::info ARM64 支持
ARM64 模板属于实验性功能。请先在 ARM64 Olares 设备上创建新世界试用。
:::

1. 从上表下载与设备架构匹配的 Egg JSON 文件。

2. 在管理员界面中打开 **Nests**。Nest 用于将相关模板归为一组。需要添加 Terraria 分类时，点击 **Create New**，将名称设为 `Terraria`。

   ![Nests 页面中的 Create New 和 Import Egg 按钮](/images/manual/use-cases/pterodactyl-nests.png#bordered)

3. 点击 **Import Egg**，选择 Terraria 分类和下载的 JSON 文件，然后导入。

   <!-- ![将 TShock Egg 导入 Terraria 分类](/images/manual/use-cases/pterodactyl-import-egg.png#bordered) -->

Egg 已包含镜像设置。创建服务器时，Wings 会自动下载镜像，并安装适用于 Terraria 1.4.5.8 的 TShock 6.2.1。

### 设置服务器参数

1. 在管理员界面的左侧导航中打开 **Users**，根据用户名找到自己的 Panel 账户，复制该账户显示的邮箱地址。

   <!-- ![在 Users 中查看 Panel 账户的邮箱地址](/images/manual/use-cases/pterodactyl-find-account-email.png#bordered) -->

2. 打开 **Servers** > **Create New**，输入服务器名称，例如 `Terraria Friends`。

3. 在 **Server Owner** 中输入刚才复制的邮箱地址，再从搜索结果中选择自己的账户。

   <!-- ![使用邮箱地址搜索并选择服务器所有者](/images/manual/use-cases/pterodactyl-select-server-owner.png#bordered) -->

4. 在 **Node** 中选择 **Olares Local Wings**，在 **Default Allocation** 中选择 `0.0.0.0:30000`。 游戏使用此端口分配中的 `30000` 端口。

5. 选择已导入的 Egg。amd64 对应 **TShock 6.2.1 (Olares)**，ARM64 对应 **TShock 6.2.1 (Olares ARM64)**。保留其中列出的 .NET 9 运行镜像，为示例世界设置以下资源：

   | 字段 | 值 |
   | --- | --- |
   | CPU Limit | `100`% |
   | Memory | `1024` MiB |
   | Disk Space | `1024` MiB |
   | Backup Limit | `2` |

   :::tip 调整服务器资源
   使用更大的世界、增加玩家人数或安装更多插件时，请根据服务器需求提高 CPU、内存和磁盘配额。
   :::

6. 保留以下 Egg 启动设置，创建一个可供两名玩家游玩的小型世界：

   | 字段 | 值 | 用途 |
   | --- | --- | --- |
   | Tshock Version | `v6.2.1` | 安装 TShock 6.2.1。 |
   | World Size | `1` | 创建小型世界。 |
   | World Name | `OlaresFriends` | 设置世界及其存档文件的名称。 |
   | Max Players | `2` | 最多允许两名玩家加入。 |

   <!-- ![配置 TShock 服务器资源和启动参数](/images/manual/use-cases/pterodactyl-configure-server.png#bordered) -->

7. 关闭 **Start Server when Installed**，创建服务器并等待安装完成。在 **Files** 中可以找到 `worlds/`、`tshock/` 和 `plugins/` 目录。世界文件会在首次启动时生成。

### 启动并进入世界

本文覆盖局域网联机。从其他网络连接时，需要网络管理员配置通往游戏端口的网络路径。

1. 打开 **Console**，点击 **Start**。首次启动需要生成世界，等待日志出现 `Server started`。

   <!-- ![控制台显示 Terraria 服务器已启动](/images/manual/use-cases/pterodactyl-start-server.png#bordered) -->

2. 在控制台命令输入框中依次输入 `version` 和 `playing`，查看版本和在线玩家信息。在控制台中直接输入命令名称即可。

3. 在同一局域网的电脑上打开 Terraria 1.4.5.8，选择 **Multiplayer** > **Join via IP**，然后选择自己的角色。

4. 输入 Olares 设备的局域网 IP 地址，在端口提示中输入 `43000`。连接成功后，角色会进入共享世界。同一局域网的好友可以使用相同的 IP 地址和端口加入。

   <!-- ![Terraria 角色进入共享世界](/images/manual/use-cases/pterodactyl-join-world.png#bordered) -->

## 管理服务器

### 使用控制台和文件管理

通过 **Console** 发送游戏命令，通过 **Start**、**Restart** 和 **Stop** 按钮控制服务器。正常关服时使用 **Stop**，让游戏保存世界。

在 **Files** 中浏览 `worlds/` 中的世界文件和 `tshock/` 中的 TShock 设置。修改 `tshock/config.json` 时，先停止服务器，编辑并保存文件，再重新启动。内部游戏端口保持为 `30000`，与端口分配一致。

使用 SFTP 客户端传输文件：

1. 打开服务器的 **Settings** > **SFTP Details**，复制服务器专属用户名。

2. 在 SFTP 客户端中选择 SFTP，输入 Olares 设备的局域网 IP 地址和端口 `42022`。

3. 输入复制的用户名和你的 Panel 账户密码，然后连接。

附加 TShock 插件放在 `plugins/` 中。选择与你的 TShock 版本兼容的插件，并先在另一台服务器上试用，再添加到与好友共用的世界。

### 定时保存世界

你可以让 Panel 每 30 分钟发送一次 TShock 的 `save` 命令，保存游戏世界。

1. 打开 **Schedules**，点击 **Create schedule**。

   ![创建计划任务的表单](/images/manual/use-cases/pterodactyl-create-schedule.png#bordered)

2. 在 **Schedule Name** 中输入 `Save world`。将 **Minute** 设为 `*/30`，将 **Hour**、**Day of Month**、**Month** 和 **Day of Week** 均设为 `*`。

3. 启用 **Only When Server Is Online** 和 **Schedule Enabled**，然后保存计划任务。

4. 添加一项命令任务，将命令设为 `save`。保存任务，并保持计划任务启用。

   <!-- ![设置每 30 分钟保存世界的计划任务](/images/manual/use-cases/pterodactyl-schedule-save.png#bordered) -->

这会保存当前世界文件。要保留可供日后恢复的独立副本，请[创建备份](#创建和恢复备份)。

### 授予有限访问权限

要让好友协助管理服务器，打开 **Users**，将对方添加为子用户。选择对方所需的权限，例如查看控制台或管理文件。一起玩的好友通过 Terraria 加入游戏。

在 Panel 中设置服务器管理权限，在 TShock 中设置游戏内的玩家权限。

### 创建和恢复备份

1. 停止服务器，等待状态变为离线。

2. 打开 **Backups**，点击 **Create backup**，使用包含用途或日期的备份名称。将 `tshock/` 目录纳入备份，以保留 TShock 设置、游戏账户和权限组。等待备份完成后，再修改服务器。

   <!-- ![Panel 中已完成的游戏服务器备份](/images/manual/use-cases/pterodactyl-create-backup.png#bordered) -->

3. 需要恢复时，停止服务器，找到目标备份并选择 **Restore**，查看将被替换的文件，然后确认。

4. 等待恢复完成，再启动服务器并通过 Terraria 加入。此时世界应恢复为备份时的状态。

:::warning 备份范围
恢复备份可能覆盖较新的游戏进度。恢复前，请为需要保留的世界另存一份副本。

游戏服务器备份保存该服务器的文件。卸载 Panel 前，如需保留 Panel 的用户、节点和计划任务，还应导出 Panel 数据库，并备份包含加密密钥的应用数据。
:::

## 常见问题

### Panel 能管理已有的 Minecraft 应用吗？

从 Market 安装的 Minecraft 通过 Olares 和 Minecraft 应用自身的界面管理。Panel 管理通过 Wings 创建的服务器。

要迁移到 Panel，需使用兼容 Egg 新建独立的 Minecraft 服务器，匹配游戏版本和模组、插件运行环境，停止原服务器，再复制世界和配置。确认玩家可以连接、世界内容正确且备份可用后，再停用原服务器。新服务器的 Panel 访问权限和游戏账户权限需要分别配置。

### 为什么显示地址与玩家连接地址不同？

Panel 可能显示内部端口分配 `0.0.0.0:30000`。玩家应使用 Olares 设备可达的局域网 IP 和对外端口 `43000`。

### 为什么导入的 Egg 无法安装或启动？

选择与 Olares 设备 CPU 架构匹配的 Egg，并保留其中列出的运行镜像。下载或安装出错时，查看安装日志。点击 **Start** 后出错时，查看 **Console**。

### 可以用这个 Egg 重装或升级服务器吗？

此 Egg 仅安装 TShock 6.2.1。如果发现已有的 `TShock.Server` 文件，安装程序会停止并提示 `Existing installation: back up and upgrade explicitly`。

先备份世界和 `tshock/` 目录，再使用支持目标版本的 Egg 准备独立服务器。确认新服务器能够加载世界存档的副本后，再让玩家迁移过去。

### 修改安装参数能重置已有管理员吗？

安装参数用于创建首个管理员账户。重置已有账户时，请使用 Panel 的账户恢复流程。配置 SMTP 后，可以通过邮件恢复账户。

## 了解更多

- [TShock 文档](https://github.com/Pryaxis/TShock/tree/v6.2.1/docs)：游戏命令、启动参数和权限。
- [Pterodactyl 文档](https://pterodactyl.io/)：上游管理功能和 Egg 配置说明。
