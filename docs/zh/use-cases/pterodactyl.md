---
outline: [2, 3]
description: 使用 Pterodactyl Panel 在 Olares 上创建和管理 Terraria/TShock 服务器，配置控制台、计划任务、访问权限和备份恢复。
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl Panel, Wings, Terraria, TShock, 游戏服务器, 备份, 服务器管理"
app_version: "0.0.12"
doc_version: "2.0"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/pterodactyl.md)为准。
:::

# 使用 Pterodactyl Panel 管理游戏服务器

Pterodactyl Panel 是开源的游戏服务器管理应用，可通过浏览器集中管理控制台、文件、计划任务、备份和访问权限。Wings 负责运行游戏进程，并处理服务器文件与备份。Wings 没有独立的网页管理界面；节点配置与游戏服务器管理统一在 Panel 中完成。

本文以在本地 Wings 节点上新建 Terraria 1.4.5.8 + TShock 6.2.1 服务器为例。TShock 为 Terraria 提供服务器管理和插件支持。

:::info Panel 可以管理哪些服务器？
Panel 管理通过 Wings 创建的游戏服务器。从 Olares Market 独立安装的应用，包括 Minecraft，仍由 Olares 管理，不会出现在 Panel 中。安装 Panel 不会自动发现或接管这些应用。迁移前请先阅读[已有 Minecraft 应用的说明](#panel-能管理已有的-minecraft-应用吗)。
:::

## 学习目标

- 配置本地节点和游戏端口分配。
- 使用兼容模板创建并运行 Terraria/TShock 服务器。
- 授予有限访问权限并配置例行任务。
- 备份世界并检查恢复后的服务器。

## 前提条件

- 已从 Market 安装 MariaDB。
- 有权限安装应用的 Olares 账户，以及用于配置节点和服务器的 Panel 管理员账户。
- 可访问互联网，以下载游戏运行环境和官方 TShock 服务器文件。
- 管理员提供的 TShock 6.2.1 Egg 和运行镜像，已针对此 Olares Wings 集成完成适配与验证。原版 Docker 镜像和 Egg 不一定兼容，具体要求见下文。
- 与 Olares 位于同一局域网的电脑；进行玩家连接检查时，电脑上需要 Terraria 1.4.5.8。

本地 Wings 集成属于实验性功能，其 Docker 服务使用特权容器。游戏模板必须兼容 UID/GID 1000 和已开放的网络端口。为长期使用的世界部署新模板前，应先完成兼容性检查。

## 安装 Pterodactyl Panel

1. 打开 Market，搜索“Pterodactyl Panel”。
   <!-- ![Market 中的 Pterodactyl Panel](/images/manual/use-cases/pterodactyl.png#bordered) -->

2. 点击**获取**，然后点击**安装**。在安装表单中保持 **ENABLE_WINGS** 启用，设置 **ADMIN_USERNAME** 和 **ADMIN_PASSWORD**，等待安装完成。密码至少包含 8 个字符。

## 配置本地节点

1. 从启动台打开 Pterodactyl Panel，使用安装时设置的用户名和密码登录。

2. 进入管理员界面，打开 **Nodes** > **Olares Local Wings** > **About**。确认 **Daemon Version** 和 **System Information** 已加载。如果始终无法显示，应先排查节点连接，再创建服务器。

   ![本地 Wings 节点显示版本和可用资源](https://cdn.olares.com/images/2026/10/7e42f94f2ec59d6abab793d424308add.png#bordered)

3. 打开 **Allocation**，创建 IP 地址为 `0.0.0.0`、端口为 `30000` 的端口分配。

端口分配指定 Wings 内部供服务器使用的端口。应用将该端口映射到 Olares 设备的 TCP `43000`。仅在 Panel 中增加一个端口分配，不会同时在设备上开放新的端口。

| 用途 | 地址或端口 | 使用方式 |
| --- | --- | --- |
| Panel | 应用入口 | 从启动台打开；入口 URL 可在**设置** > **应用** > **Pterodactyl Panel** > **入口**中查看 |
| Terraria | Olares 局域网 IP、TCP `43000` | 在 Terraria 中通过 IP 加入，分别输入 IP 和端口 `43000` |
| 内部端口分配 | `0.0.0.0:30000` | 创建服务器时选择，不用于玩家连接 |
| SFTP | Olares 局域网 IP、TCP `42022` | 使用 SFTP 客户端；用户名从服务器的 **Settings** > **SFTP Details** 复制 |

将 `OLARES_LAN_IP` 替换为设备的实际局域网地址。Panel 的 HTTPS 入口用于浏览器管理，游戏客户端和 SFTP 客户端使用各自的协议与端口。本文覆盖局域网访问；从其他网络访问时，需要单独配置并验证通往游戏端口的网络路径。

## 创建 Terraria/TShock 服务器

### 导入兼容模板

Pterodactyl 将游戏模板称为 Egg。Egg 定义运行镜像、安装脚本、启动命令和服务器变量。

1. 向管理员获取已验证的 TShock Egg。参考版本为 [TShock 6.2.1 / Terraria 1.4.5.8](https://github.com/Pryaxis/TShock/releases/tag/v6.2.1)。

2. 在管理员界面中按需创建 Terraria 分类。打开 **Nests** > **Import Egg**，选择该分类和管理员提供的 JSON 文件，然后导入。

模板必须兼容 UID/GID 1000，将世界保存在 `worlds/`，将 TShock 配置和 SQLite 数据库保存在 `tshock/`，并监听分配的端口。附加插件放在 `plugins/`。本例保持 TShock REST API 关闭，无需额外的管理入口。

上游 TShock Docker 镜像使用自己的启动入口和目录，需要适配 Wings 的启动命令和可写服务器目录。历史验证使用了适配镜像；本文没有提供可直接下载的通用 Egg。请勿替换成未经验证的原版镜像。

### 设置服务器参数

1. 打开 **Servers** > **Create New**，输入服务器名称，例如 `Terraria Friends`。

2. 在 **Server Owner** 中选择自己的账户，选择 **Olares Local Wings** 节点，并在 **Default Allocation** 中选择 `0.0.0.0:30000`。

3. 选择导入的 TShock Egg 及已验证的运行镜像，使用以下小型验证服务器参数：

   | 字段 | 值 |
   | --- | --- |
   | CPU Limit | `100` percent |
   | Memory | `1024` MiB |
   | Disk Space | `1024` MiB |
   | Backup Limit | `2` |
   | 游戏 / TShock | `1.4.5.8` / `6.2.1` |
   | 内部游戏端口 | `30000` |
   | 最大玩家数 | `2` |

   在 Egg 启动参数中配置新建小型世界。这些参数用于功能验证，不代表推荐容量。变量名称以管理员提供的 Egg 为准。

4. 关闭 **Start Server when Installed**，创建服务器并等待安装完成。在 **Files** 中检查可写的 `worlds/`、`tshock/` 和 `plugins/` 目录。

### 启动并连接世界

1. 打开 **Console**，点击 **Start**。首次启动需要生成世界，等待日志出现 `Listening on port 30000`。

2. 在控制台命令输入框中依次输入 `version` 和 `playing`，确认返回版本和在线玩家信息。TShock 控制台命令不需要游戏内使用的斜杠前缀。

3. 在同一局域网的电脑上打开 Terraria 1.4.5.8，选择 **Multiplayer** > **Join via IP**。选择角色，输入 `OLARES_LAN_IP`，在端口提示中输入 `43000`，确认角色能够进入世界。

协议握手成功只能说明服务可达。历史验证没有包含真实玩家加入；邀请玩家前需要完成此项检查。

## 管理服务器

### 使用控制台和文件管理

通过 **Console** 执行命令，通过 **Start**、**Restart** 和 **Stop** 控制生命周期。正常关服时使用 **Stop**，让游戏保存世界；仅在进程无响应时考虑强制终止。

在 **Files** 中查看 `worlds/` 和 `tshock/`。修改 `tshock/config.json` 前先停止服务器，保存后再启动。游戏端口保持为 `30000`；启动命令中的参数可能覆盖配置文件。仅将版本兼容、已在可删除服务器上验证的插件放入 `plugins/`。


使用 SFTP 时，从 **Settings** > **SFTP Details** 复制服务器专属用户名，连接 Olares 局域网 IP 的 `42022` 端口，并使用对应 Panel 账户的密码。传输世界前，先用一个可删除的小文本文件检查访问是否正常。不要使用普通 FTP。

### 配置例行命令

1. 打开 **Schedules** 并创建计划任务。设置清晰的名称，以及分钟、小时、日、月和星期字段。

2. 添加命令任务，内容为 `playing`。本例启用 **Only When Server Is Online**。

3. 首次验证时，将五个时间字段均设为 `*`，每分钟运行一次。在 **Console** 中确认出现在线玩家信息后，关闭计划任务或改为实际需要的执行频率。

选择执行时间前先确认计划任务使用的时区。不要让临时的每分钟任务持续运行。

### 授予有限访问权限

打开服务器的 **Users**，添加子用户。仅授予职责所需权限，例如查看控制台或执行指定文件操作。使用独立的浏览器会话登录该账户，分别验证一项允许操作和一项禁止操作。

Panel 子用户权限控制服务器管理操作，不替代游戏内部的玩家权限或插件权限组。

### 创建和恢复备份

1. 停止服务器，等待状态变为离线。

2. 打开 **Backups**，点击 **Create backup**，使用包含用途或日期的备份名称。等待备份完成；已提交请求或正在排队不代表备份已完成。

3. 修改前保留一份已完成的备份。恢复时保持服务器离线，找到目标备份并选择 **Restore**，检查文件替换选项后确认操作。

4. 等待恢复完成，再启动服务器。确认服务器监听端口 `30000`，匹配版本的 Terraria 客户端可以加入，并且所需世界状态已恢复。恢复游戏账户和权限组时，备份范围必须包含 TShock SQLite 数据库。

:::warning 备份范围
恢复可能覆盖文件。依赖备份保护重要世界前，应先在可删除的服务器上验证恢复流程。

游戏服务器备份覆盖该服务器的文件，不包含 Panel 的用户、节点、计划任务或数据库。卸载 Panel 前，应单独导出数据库，并备份包含加密密钥的应用数据。仅保留应用数据，不保证重新安装后中间件数据库仍然存在。
:::

## 检查部署结果

邀请玩家前，确认服务器能够启动、接受控制台命令、在正常重启后保留世界，并能从已完成的备份中恢复。完成真实玩家加入检查，并只授予用户所需权限。

详细验收步骤、预期结果、已有证据和覆盖限制见 [Pterodactyl 验收检查清单](pterodactyl-test-cases.md)。基础检查通过不代表已验证多人容量或任意模组、插件的兼容性。

## 常见问题

### Panel 能管理已有的 Minecraft 应用吗？

Panel 不能直接管理从 Market 独立安装的 Minecraft。该应用由 Olares 作为 Kubernetes 工作负载管理；Wings 则创建并控制自己的游戏容器。此集成没有让 Wings 接管现有 Olares 应用的机制。

继续使用市场版时，通过 Olares 和该应用自身的界面管理即可。若要迁移到 Panel，应使用兼容 Egg 新建独立的 Minecraft 服务器，匹配游戏版本和模组、插件运行环境，停止原服务器，再复制世界和配置。验证玩家连接、世界内容和备份后，再停用原服务器。Panel 管理权限和游戏账户权限需要分别配置。这是数据迁移，不是关联或接管原应用。

### 为什么显示地址与玩家连接地址不同？

Panel 可能显示内部端口分配 `0.0.0.0:30000`。玩家应使用 Olares 设备可达的局域网 IP 和对外端口 `43000`。在游戏客户端中输入 Panel URL 或内部端口分配，不能访问对外的游戏入口。

### 为什么导入的 Egg 无法安装或启动？

检查运行环境版本、下载连接、启动命令、文件所有者和端口设置。此集成使用 UID/GID 1000 运行 Wings 和游戏进程。若模板将文件留给其他用户所有，可能需要适配。安装日志与游戏控制台日志应分别检查。

### 修改安装参数能重置已有管理员吗？

初始化参数用于创建首个管理员，不会覆盖已经存在的管理员。应使用 Panel 的账户恢复流程；通过邮件恢复账户需要可用的 SMTP 配置。

## 了解更多

- [Pterodactyl 验收检查清单](pterodactyl-test-cases.md)：管理、游戏运行、权限和恢复的可重复验证步骤。
- [TShock 文档](https://github.com/Pryaxis/TShock/tree/v6.2.1/docs)：游戏命令、启动参数和权限。
- [Pterodactyl 文档](https://pterodactyl.io/)：上游管理功能和 Egg 配置说明。
