---
outline: [2, 3]
description: 使用 Pterodactyl Panel 在 Olares 上创建和管理 Minecraft Java 服务器，配置控制台、计划任务、访问权限和备份恢复。
head:
  - - meta
    - name: keywords
      content: "Olares, Pterodactyl Panel, Wings, Minecraft, 游戏服务器, 备份, 服务器管理"
app_version: "0.0.12"
doc_version: "1.0"
doc_updated: "2026-10-13"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/pterodactyl.md)为准。
:::

# 使用 Pterodactyl Panel 管理游戏服务器

Pterodactyl Panel 是开源的游戏服务器管理应用，可通过浏览器集中管理控制台、文件、计划任务、备份和访问权限。Wings 负责运行游戏进程，并处理服务器文件与备份。

本文以本地 Wings 节点上的独立 Minecraft Java 1.21.1 服务器为例，不导入或接管 Olares 上已有的 Minecraft 应用。

## 学习目标

- 配置本地节点和游戏端口分配。
- 使用兼容模板创建并运行 Minecraft 服务器。
- 授予有限访问权限并配置例行任务。
- 备份世界并检查恢复后的服务器。

## 前提条件

- 已从 Market 安装 MariaDB。
- 有权限安装应用的 Olares 账户，以及用于配置节点和服务器的 Panel 管理员账户。
- 可访问互联网，以下载游戏运行环境和官方 Minecraft 服务器文件。
- 与 Olares 位于同一局域网的电脑；进行玩家连接检查时，电脑上需要 Minecraft Java Edition 1.21.1。

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
| Minecraft | Olares 局域网 IP、TCP `43000` | 在 Minecraft Java Edition 中输入 `OLARES_LAN_IP:43000` |
| 内部端口分配 | `0.0.0.0:30000` | 创建服务器时选择，不用于玩家连接 |
| SFTP | Olares 局域网 IP、TCP `42022` | 使用 SFTP 客户端；用户名从服务器的 **Settings** > **SFTP Details** 复制 |

将 `OLARES_LAN_IP` 替换为设备的实际局域网地址。Panel 的 HTTPS 入口用于浏览器管理，游戏客户端和 SFTP 客户端使用各自的协议与端口。本文覆盖局域网访问；从其他网络访问时，需要单独配置并验证通往游戏端口的网络路径。

## 创建 Minecraft 服务器

### 导入示例模板

Pterodactyl 将游戏模板称为 Egg。Egg 定义运行镜像、安装脚本、启动命令和服务器变量。

1. 下载 [Minecraft 1.21.1 示例 Egg](/files/use-cases/pterodactyl/vanilla-minecraft-1.21.1.json)。

2. 在管理员界面打开 **Nests** > **Import Egg**，选择 Minecraft 分类和下载的 JSON 文件，然后导入。

示例使用 Java 21，并以 UID/GID 1000 创建文件；初始配置为 `eula=false`。模板基于功能验证时使用的模板和参数。导入模板不会代替用户接受 Minecraft EULA。

### 设置服务器参数

1. 打开 **Servers** > **Create New**，输入服务器名称，例如 `Minecraft Friends`。

2. 在 **Server Owner** 中选择自己的账户，选择 **Olares Local Wings** 节点，并在 **Default Allocation** 中选择 `0.0.0.0:30000`。

3. 选择导入的 **Vanilla Minecraft 1.21.1 (Olares)** Egg 及其 Java 21 镜像，使用以下示例参数：

   | 字段 | 值 |
   | --- | --- |
   | CPU Limit | `100` percent |
   | Memory | `1536` MiB |
   | Disk Space | `2048` MiB |
   | Backup Limit | `2` |
   | Server Jar File | `server.jar` |
   | Server Version | `1.21.1` |

   这些参数用于小型功能验证，不代表适用于所有世界或玩家数量的推荐容量。示例将 Java 最大堆内存限制为 1024 MiB，最多允许两名玩家。增加资源和玩家数量前，应先验证实际负载。

4. 关闭 **Start Server when Installed**，创建服务器并等待安装完成。打开服务器的 **Files** 页面，确认已生成 `server.jar`、`server.properties` 和 `eula.txt`。

### 接受 EULA 并启动世界

1. 阅读 [Minecraft EULA](https://www.minecraft.net/en-us/eula)，仅在同意条款后继续。

2. 在 **Files** 中打开 `eula.txt`，将 `eula=false` 改为 `eula=true`，然后点击 **Save Content**。

3. 打开 **Console**，点击 **Start**。首次启动会生成世界；出现 `Done` 消息后再连接玩家。

4. 在控制台命令输入框中输入 `list`。返回在线玩家数量，说明控制台输入已到达运行中的游戏进程。

   ![Minecraft 运行中的控制台输出和资源统计](https://cdn.olares.com/images/2026/10/e6b8791ebba2f51e2cdc6e783a6f0a58.png#bordered)

5. 在同一局域网的电脑上打开 Minecraft Java Edition 1.21.1，选择**多人游戏** > **添加服务器**，输入 `OLARES_LAN_IP:43000`。确认列表中能显示服务器，然后加入世界。

服务器列表返回状态只能说明协议可达。成功进入世界是独立的检查项，不能由端口可达或 Panel 中的运行状态推断。

## 管理服务器

### 使用控制台和文件管理

通过 **Console** 执行命令，通过 **Start**、**Restart** 和 **Stop** 控制生命周期。正常关服时使用 **Stop**，让游戏保存世界；仅在进程无响应时考虑强制终止。

在 **Files** 中打开 `server.properties` 修改游戏设置。编辑持久化配置前先停止服务器，保存后再启动。此端口分配应保持 `server-port=30000`。将插件上传到 Vanilla 服务器，不会让它自动获得插件支持；其他游戏版本需要相应的 Egg 和运行环境。

![Panel 文件管理器中的服务器文件](https://cdn.olares.com/images/2026/10/3d08eb1c79172687b0622fcea9247940.png#bordered)

使用 SFTP 时，从 **Settings** > **SFTP Details** 复制服务器专属用户名，连接 Olares 局域网 IP 的 `42022` 端口，并使用对应 Panel 账户的密码。传输世界前，先用一个可删除的小文本文件检查访问是否正常。不要使用普通 FTP。

### 配置例行命令

1. 打开 **Schedules** 并创建计划任务。设置清晰的名称，以及分钟、小时、日、月和星期字段。

2. 添加命令任务，内容为 `say Scheduled maintenance reminder`。本例启用 **Only When Server Is Online**。

3. 首次验证时，将五个时间字段均设为 `*`，每分钟运行一次。在 **Console** 中确认出现消息后，关闭计划任务或改为实际需要的执行频率。

选择执行时间前先确认计划任务使用的时区。不要让临时的每分钟任务持续运行。

### 授予有限访问权限

打开服务器的 **Users**，添加子用户。仅授予职责所需权限，例如查看控制台或执行指定文件操作。使用独立的浏览器会话登录该账户，分别验证一项允许操作和一项禁止操作。

Panel 子用户权限控制服务器管理操作，不替代游戏内部的玩家权限或插件权限组。

### 创建和恢复备份

1. 停止服务器，等待状态变为离线。

2. 打开 **Backups**，点击 **Create backup**，使用包含用途或日期的备份名称。等待备份完成；已提交请求或正在排队不代表备份已完成。

3. 修改前保留一份已完成的备份。恢复时保持服务器离线，找到目标备份并选择 **Restore**，检查文件替换选项后确认操作。

4. 等待恢复完成，再启动服务器。确认出现 `Done`，匹配版本的 Minecraft 客户端可以连接，并且所需世界状态已恢复。

:::warning 备份范围
恢复可能覆盖文件。依赖备份保护重要世界前，应先在可删除的服务器上验证恢复流程。

游戏服务器备份覆盖该服务器的文件，不包含 Panel 的用户、节点、计划任务或数据库。卸载 Panel 前，应单独导出数据库，并备份包含加密密钥的应用数据。仅保留应用数据，不保证重新安装后中间件数据库仍然存在。
:::

## 检查部署结果

邀请玩家前，确认服务器能够启动、接受控制台命令、在正常重启后保留世界，并能从已完成的备份中恢复。完成真实玩家加入检查，并只授予用户所需权限。

详细验收步骤、预期结果、已有证据和覆盖限制见 [Pterodactyl 验收检查清单](pterodactyl-test-cases.md)。基础检查通过不代表已验证多人容量或任意模组、插件的兼容性。

## 常见问题

### Panel 能管理已有的 Minecraft 应用吗？

Panel 管理通过兼容 Egg 在 Wings 节点上创建的服务器，不会自动接管从 Market 单独安装的 Minecraft 应用。应将世界副本迁移到新的 Panel 服务器并完成验证，再停用原服务器。

### 为什么显示地址与玩家连接地址不同？

Panel 可能显示内部端口分配 `0.0.0.0:30000`。玩家应使用 Olares 设备可达的局域网 IP 和对外端口 `43000`。在 Minecraft 中输入 Panel URL 或内部端口分配，不能访问对外的游戏入口。

### 为什么导入的 Egg 无法安装或启动？

检查运行环境版本、下载连接、启动命令、文件所有者和端口设置。此集成使用 UID/GID 1000 运行 Wings 和游戏进程。若模板将文件留给其他用户所有，可能需要适配。安装日志与游戏控制台日志应分别检查。

### 修改安装参数能重置已有管理员吗？

初始化参数用于创建首个管理员，不会覆盖已经存在的管理员。应使用 Panel 的账户恢复流程；通过邮件恢复账户需要可用的 SMTP 配置。

## 了解更多

- [Pterodactyl 验收检查清单](pterodactyl-test-cases.md)：管理、游戏运行、权限和恢复的可重复验证步骤。
- [Pterodactyl 文档](https://pterodactyl.io/)：上游管理功能和 Egg 配置说明。
