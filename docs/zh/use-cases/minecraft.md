---
outline: [2, 3]
description: 在 Olares 上托管 Minecraft Java 版服务器，选择游戏版本与 Vanilla、Forge 或 Fabric，配置私服并为好友创建独立实例。
head:
  - - meta
    - name: keywords
      content: Olares, Minecraft, Minecraft Java, 游戏服务器, Forge, Fabric, 模组, 离线模式, 克隆, Overlay gateway, VPN, 游戏
app_version: "0.1.22"
doc_version: "2.0"
doc_updated: "2026-10-10"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/minecraft.md)为准。
:::

# 在 Olares 上与好友联机游玩 Minecraft Java 版

Olares 上的 Minecraft 托管 Java 版专用服务器，支持 Vanilla、Forge 和 Fabric。你通过控制台终端管理服务器，玩家使用兼容的 Java 版客户端连接。好友可以通过 Overlay gateway 在局域网中联机，也可以通过 LarePass VPN 远程加入。

:::info 本教程对应的应用版本
本教程对应 Minecraft 应用包 0.1.22。应用包版本与安装时选择的 Minecraft 游戏版本不同。这些功能已在测试市场提供；如果你的 Market 来源仍提供 0.1.14，请等待应用更新后，再使用下文的配置与 Clone 功能。
:::

## 学习目标

通过本教程，你将学习如何：

- 安装时选择 Minecraft 游戏版本和服务器类型。
- 安装模组，配置正版验证、作弊命令和人数上限。
- 使用 Clone 创建独立服务器。
- 启用 Overlay gateway，让局域网中的玩家连接。
- 通过局域网或 VPN 连接服务器。

## 准备工作

- Olares 版本为 1.12.6 或更高。
- 使用原生 Linux 主机和有线以太网进行本地 Overlay 访问。Overlay gateway 在 Wi-Fi 或 WSL 环境下无法工作。
- 由 Super admin 开启系统级 Overlay gateway 服务。开启后，Admin 或 Member 可以为 Minecraft 启用该功能。
- 每位玩家的电脑上已安装 Minecraft Java 版。Bedrock、主机和移动版不能直接连接。客户端须匹配所选的 Minecraft **VERSION**，而不是应用包版本。

## 安装 Minecraft

1. 打开 Market，搜索 "Minecraft"。
   ![Market 中的 Minecraft](/images/manual/use-cases/minecraft.png#bordered)

2. 点击 **Get**，然后点击 **Install**。在环境变量对话框中，参考下文选择 **VERSION** 和 **TYPE**，按模组要求配置加载器，等待安装完成。

首次启动会下载服务器和加载器资源，根据网络情况可能需要几分钟。等待服务器就绪后再加入。从 Launchpad 打开 Minecraft，即可在控制台查看启动日志。

## 配置 Minecraft

### 选择游戏版本和服务器类型

在安装或创建 Clone 时选择 **VERSION**。Olares 会自动选择服务器的 Java 运行时，无需手动配置。

| VERSION | 服务器 Java 版本 | 主要用途 |
|:---|:---|:---|
| `1.12.2` | Java 8 | 较早的 Forge 模组与整合包，不支持 Fabric。 |
| `1.16.5` | Java 8 | 为 1.16.5 制作的 Forge 或 Fabric 整合包。 |
| `1.18.2` | Java 17 | 适用于 1.18.2 的模组与整合包。 |
| `1.19.2` | Java 17 | 适用于 1.19.2 的模组与整合包。 |
| `1.20.1` | Java 17 | 适用于 1.20.1 的模组与整合包。 |
| `1.21.1` | Java 21 | 适用于 1.21.1 的模组与整合包，需核对加载器要求。 |
| `26.2` | Java 25 | 使用固定较新版本进行原版游玩，使用模组前需确认加载器支持。 |
| `LATEST` | Java 25 | 最新正式版，重启时游戏版本可能更新。 |

按整合包要求选择版本。版本出现在列表中，不代表所有加载器和模组都支持它。长期运行的世界建议使用固定版本，避免使用 `LATEST`。

安装时可配置以下字段：

- **TYPE**：`VANILLA` 为不带模组加载器的原版，`FORGE` 用于 Forge 模组，`FABRIC` 用于 Fabric 模组。默认值为 `VANILLA`。当前应用包不提供 NeoForge。
- **FORGE_VERSION**：仅在 `FORGE` 模式下生效，默认值为 `recommended`。也可填写 `latest`，或整合包要求的准确 Forge 版本。
- **FABRIC_LOADER_VERSION**：仅在 `FABRIC` 模式下生效，默认值为 `latest`。整合包要求固定加载器版本时，请填写准确版本。Fabric API 是需要单独安装的模组。

**VERSION**、**TYPE** 和加载器版本在安装后不能修改。需要运行其他版本或加载器时，请使用 Clone 创建独立实例。在官方 Java 启动器的 **配置（Installations）** 页面中新建配置，选择与服务器相同的游戏版本。使用模组时，还需安装兼容的加载器及必需的客户端模组。

:::warning 保护已有世界
迁移世界或重新安装前，请备份实例的整个数据目录。使用其他版本重新安装时，可能继续使用原有数据目录。升级可能转换世界数据；降级可能导致无法启动，或丢失区块、物品和实体。切换加载器或移除模组也可能造成存档不兼容。不要用低版本服务器启动已被高版本保存的世界。
:::

### 安装 Forge 或 Fabric 模组

例如，选择 `VERSION=1.20.1`，将 `TYPE` 设为 `FORGE` 或 `FABRIC`，再填写模组要求的加载器版本。

1. 在 Market 或 Settings 中停止目标 Minecraft 实例。
2. 打开 Files，找到 `Data/<instance-name>/data`。主实例通常为 `Data/minecraft/data`。请使用实例实际的数据文件夹名称，它可能与显示标题不同。
3. 如果没有 `mods` 文件夹，先创建它，再上传服务端模组的 `.jar` 文件及所需依赖。模组必须匹配所选的 Minecraft 版本和加载器。如果模组需要 Fabric API，也将其放入该文件夹。仅限客户端的模组留在客户端。
4. 将整合包需要的配置文件放入同一 `data` 目录下的对应文件夹。仅上传整合包压缩包不会自动安装。
5. 恢复运行实例，检查控制台日志，确认没有缺少依赖或模组不兼容的错误。
6. 使用匹配的游戏版本、兼容的加载器和必需的客户端模组连接。部分纯服务端模组允许原版客户端加入，请以模组说明为准。

### 关闭正版验证，配置私服

正版验证默认开启。如果要为可信玩家提供离线模式私服：

1. 进入 **Settings** > **Applications**，选择目标 Minecraft 实例。
2. 在 **Environment variables** 下点击 **Manage environment variables**。
3. 编辑 **ONLINE_MODE**，选择 `false`（**Disabled (trusted players only)**），点击 **Confirm**。
4. 点击 **Apply**，等待服务器重启。

此设置会关闭 Minecraft 账号验证和安全档案强制校验，支持离线模式私服。它不改变游戏客户端的许可要求，也不改变网络访问设置。玩家仍需使用匹配的游戏版本和可访问的服务器地址。

:::warning 仅用于可信玩家
离线模式不会验证玩家身份，其他人可能冒用已有玩家的名称。请限制为可信玩家访问。在线与离线模式切换会改变玩家 UUID，可能影响背包和权限，切换前请备份数据。
:::

### 开启作弊命令

在同一 **Manage environment variables** 页面中，将 **ALLOW_CHEATS** 设为 `true`，点击 **Confirm**，再点击 **Apply**。服务器会重启。默认值为 `false`。

开启后，每个加入的玩家都能使用 `/give`、`/gamemode` 等命令，同时启用命令方块。此设置授予 2 级管理员权限；停止服务器等命令仍需在控制台执行。请仅对可信玩家开启。

关闭时，将 **ALLOW_CHEATS** 设为 `false` 并应用。此操作会清空玩家管理员列表，包括手动添加的管理员，并重启服务器。

### 设置最大在线人数

默认最多允许 8 人同时在线。在 **Manage environment variables** 中，将 **MAX_PLAYERS** 修改为 `1` 到 `100` 的整数，点击 **Confirm**，再点击 **Apply**。服务器会重启并应用人数限制。更高的人数上限需要更多资源，使用模组时尤其如此。

### 使用 Clone 创建另一台服务器

Clone 可以在同一 Olares 设备上运行不同版本或不同模组配置的服务器。它创建具有独立配置和数据目录的新实例，不会复制原实例的世界和模组。

1. 先安装 Minecraft 主实例，然后在 Market 中打开 **My Olares**。
2. 找到 Minecraft，点击 **Open** 旁的下拉箭头，选择 **Clone**。
3. 填写唯一的 **New app title** 和 **Desktop shortcut name**，点击 **Confirm**。
4. 在 **Configure Environment Variables** 中，为新服务器选择 **VERSION**、**TYPE** 和加载器配置，点击 **Confirm**，等待安装完成。
5. 单独管理新实例，将它的模组上传到 `Data/<instance-name>/data/mods`。需要局域网访问时，为该实例开启 Overlay。

每个实例都有自己的 VPN 外部端口，启用 Overlay 后也有自己的地址。按下文步骤复制目标实例的连接地址。Overlay 端口仍为 `25565`，VPN 外部端口则单独分配。同时运行多个服务器会增加内存和 CPU 占用。

## 为 Minecraft 启用 Overlay gateway

Overlay gateway 会为 Minecraft 分配专用的本地 IP 地址，让同一网络中的玩家直接连接。

1. 打开 Olares Settings，进入 **Network** > **Overlay gateway**。
2. 确认 **Enable overlay gateway** 开关已打开。这是系统级服务开关，如未开启，需由 Super admin 打开。
3. 在 **Applications** 列表中找到 **Minecraft**，确认其状态为 **Running**，然后为该应用开启 Overlay gateway。
4. 在 **Minecraft Java** 右侧复制显示的地址，例如 `192.168.50.219:25565`。

:::info 本地 IP 地址是动态的
Overlay gateway 动态分配本地 IP 地址，应用重启或网络变化后可能改变。请始终使用当前页面上显示的地址。
:::

## 从同一局域网连接

与 Olares 设备处于同一局域网的玩家可以通过 Overlay 地址连接。

1. 从 **Settings** > **Network** > **Overlay gateway** 复制 Overlay 地址，例如 `192.168.50.219:25565`。
2. 打开 Minecraft Java 版，点击 **Multiplayer**。

   ![Minecraft 多人游戏菜单](/images/manual/use-cases/minecraft-multiplayer-menu.png#bordered)

3. 点击 **Add Server**。
4. 填写服务器信息，然后点击 **Done**：

   - **Server Name**：填写便于识别的名称。
   - **Server Address**：填写刚才复制的 Overlay gateway 地址。

   ![添加 Minecraft 服务器](/images/manual/use-cases/minecraft-add-server.png#bordered)

5. 选择服务器，点击 **Join Server**。

   ![加入 Minecraft 服务器](/images/manual/use-cases/minecraft-join-server.png#bordered)

## 通过 VPN 连接

玩家与 Olares 设备不在同一局域网时，可以使用此方法。

:::tip 可以保持 Overlay 开启
Minecraft 的 Overlay gateway 与 VPN 连接不冲突，可以同时启用。
:::

1. 确保已启用 [LarePass VPN](../manual/get-started/local-access.md#using-larepass-vpn)。
2. 打开 Minecraft Java 版，点击 **Multiplayer**。
3. 点击 **Add Server**。
4. 打开 **Settings** > **Applications** > **Minecraft**（或目标 Clone）> **Ports**，复制该实例显示的地址与外部端口，填写到 **Server Address**。外部端口按实例分配，不要假定为 `25565`，也不要复用其他实例的地址。
5. 保存服务器，点击 **Join Server**。

## 管理服务器

Minecraft 应用没有 Web 管理界面。要查看日志或执行服务器命令，请从 Launchpad 打开应用，使用内置控制台终端。

## 常见问题

### 为什么我的本地 IP 地址与截图不同？

Overlay gateway 动态分配本地 IP 地址。连接时，请使用 **Settings** > **Network** > **Overlay gateway** > **Minecraft** 中当前显示的地址。

### 为什么客户端提示版本不匹配？

客户端应选择与实例 **VERSION** 相同的 Minecraft 游戏版本。Market 中的应用包版本，例如 `0.1.22`，不是游戏版本。使用 Forge 或 Fabric 时，还需检查加载器及客户端模组要求。

### 为什么客户端提示聊天消息无法验证？

离线模式下不会验证玩家身份和安全档案，客户端可能提示聊天消息无法验证。这不代表网络连接失败。

### 为什么无法通过 VPN 地址连接？

请检查以下事项：

- 使用的是 Minecraft Java 版。
- 地址与外部端口来自 **Settings** > **Applications** > 目标实例 > **Ports**。
- 服务器已就绪，客户端匹配其游戏版本与模组要求。
- LarePass VPN 已启用。

### 升级应用会怎样？

升级应用会重启服务器，当前在线玩家会被断开连接。

## 了解更多

- [管理应用的 Overlay 网关](/zh/manual/olares/settings/overlay-gateway.md)：为支持的应用配置局域网访问。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改设置并通过重启应用。
- [克隆应用](../manual/olares/market/clone-apps.md)：创建并管理独立应用实例。
