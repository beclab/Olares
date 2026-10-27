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

## 学习目标

通过本教程，你将学习如何：

- 安装时选择 Minecraft 游戏版本和服务器类型。
- 通过局域网或 LarePass VPN 连接服务器。
- 安装模组，配置正版验证、作弊命令和人数上限。
- 使用 Clone 创建独立服务器。

## 准备工作

- Olares 1.12.6 或更高版本。
- 每位玩家的电脑上已安装 Minecraft Java 版。Bedrock、主机和移动版无法直接连接。

:::info 客户端须与服务器匹配
客户端的 Minecraft 游戏版本必须与服务器一致。使用 Forge 或 Fabric 时，每位玩家还需要安装整合包要求的加载器和客户端模组。
:::

## 安装 Minecraft

1. 打开 Market，搜索 "Minecraft"。
   ![Market 中的 Minecraft](/images/manual/use-cases/minecraft.png#bordered)

2. 点击 **Get**，然后点击 **Install**。选择要游玩的 Minecraft **VERSION**，原版游玩保持 **TYPE** 为 `VANILLA`。使用模组时，展开下方的游戏版本和服务器类型说明进行配置，然后等待安装完成。

:::: details 选择游戏版本和服务器类型

在安装或创建 Clone 时选择 **VERSION**。Olares 会自动选择服务器的 Java 运行时，无需手动配置。

| VERSION | 服务器 Java 版本 |
|:---|:---|
| `1.12.2` | Java 8 |
| `1.16.5` | Java 8 |
| `1.18.2` | Java 17 |
| `1.19.2` | Java 17 |
| `1.20.1` | Java 17 |
| `1.21.1` | Java 21 |
| `26.2` | Java 25 |
| `LATEST` | Java 25 |

选择加载器和模组支持的固定版本。Minecraft `1.12.2` 不支持 Fabric。`LATEST` 会选择最新正式版，服务器重启时游戏版本可能更新。

选择服务器类型及对应的加载器设置：

| 字段 | 值 |
|:---|:---|
| **TYPE** | `VANILLA`（默认）、`FORGE` 或 `FABRIC` |
| **FORGE_VERSION** | 仅用于 Forge。填写 `recommended`（默认）、`latest` 或整合包要求的版本。 |
| **FABRIC_LOADER_VERSION** | 仅用于 Fabric。填写 `latest`（默认）或整合包要求的版本。 |

这些设置及 **VERSION** 在安装后不能修改。需要其他游戏版本或加载器时，使用 [Clone](../manual/olares/market/clone-apps.md) 创建服务器。

:::warning 保护已有世界
迁移世界或重新安装前，请备份实例的整个数据目录。使用其他版本重新安装时，可能继续使用原有数据目录。升级可能转换世界数据。降级可能导致无法启动，或丢失区块、物品和实体。切换加载器或移除模组也可能造成存档不兼容。不要用低版本服务器启动已被高版本保存的世界。
:::
::::

首次启动会下载服务器和加载器资源，根据网络情况可能需要几分钟。原版游玩可直接继续[连接服务器](#连接服务器)。使用模组时，请先[安装模组](#安装-forge-或-fabric-模组)再加入。

## 配置 Minecraft

按需完成以下可选配置，或直接[连接服务器](#连接服务器)。

### 可选：安装 Forge 或 Fabric 模组 {#安装-forge-或-fabric-模组}

例如，选择 `VERSION=1.20.1`，将 `TYPE` 设为 `FORGE` 或 `FABRIC`，再填写模组要求的加载器版本。

1. 在 Market 或 Settings 中停止目标 Minecraft 实例。
2. 打开 Files，找到 `Data/<app-name>/data`。主实例为 `Data/minecraft/data`。Clone 实例请使用 Market 详情页 URL 中的应用名称，具体方法参见[找到实例的数据目录](#find-the-instances-data-directory)。
3. 如果没有 `mods` 文件夹，先创建它，再上传服务端模组的 `.jar` 文件及所需依赖。模组必须匹配所选的 Minecraft 版本和加载器。如果模组需要 Fabric API，也将其放入该文件夹。仅限客户端的模组留在客户端。
4. 将整合包需要的配置文件放入同一 `data` 目录下的对应文件夹。仅上传整合包压缩包不会自动安装。
5. 恢复运行实例，检查控制台日志，确认没有缺少依赖或模组不兼容的错误。
6. 使用匹配的游戏版本、兼容的加载器和必需的客户端模组连接。客户端的具体要求以模组说明为准。

### 可选：关闭正版验证，配置私服 {#关闭正版验证-配置私服}

正版验证默认开启。如果要为可信玩家提供离线模式私服：

1. 进入 **Settings** > **Applications**，选择目标 Minecraft 实例。
2. 在 **Environment variables** 下点击 **Manage environment variables**。
3. 编辑 **ONLINE_MODE**，选择 `false`（**Disabled (trusted players only)**），点击 **Confirm**。
4. 点击 **Apply**，等待服务器重启。

此设置会关闭 Minecraft 账号验证和安全档案强制校验。玩家仍需使用可访问的服务器地址和兼容的客户端。

:::warning 仅用于可信玩家
离线模式不会验证玩家身份，其他人可能冒用已有玩家的名称。请限制为可信玩家访问。在线与离线模式切换会改变玩家 UUID，可能影响背包和权限，切换前请备份数据。
:::

### 可选：开启作弊命令 {#开启作弊命令}

作弊命令默认关闭。按以下步骤开启：

1. 进入 **Settings** > **Applications**，选择 Minecraft 实例。
2. 在 **Environment variables** 下点击 **Manage environment variables**。
3. 将 **ALLOW_CHEATS** 设为 `true`，点击 **Confirm**。
4. 点击 **Apply**，等待服务器重启。

开启后，每个加入的玩家都能使用 `/give`、`/gamemode` 等命令，同时启用命令方块。此设置授予 2 级管理员权限。停止服务器等命令请在控制台执行。请仅对可信玩家开启。

关闭时，将 **ALLOW_CHEATS** 设为 `false` 并应用。此操作会清空玩家管理员列表，包括手动添加的管理员，并重启服务器。

### 可选：设置最大在线人数 {#设置最大在线人数}

默认最多允许 8 人同时在线。按以下步骤修改：

1. 进入 **Settings** > **Applications**，选择 Minecraft 实例。
2. 在 **Environment variables** 下点击 **Manage environment variables**。
3. 将 **MAX_PLAYERS** 设为 `1` 到 `100` 的整数，点击 **Confirm**。
4. 点击 **Apply**，等待服务器重启。

更高的人数上限需要更多资源，使用模组时尤其如此。

## 连接服务器

等待服务器启动完成后再加入。从 Launchpad 打开实例，即可查看控制台日志。

在 Minecraft 启动器中打开 **Installations**，选择与服务器相同的游戏版本。再选择下方的连接方式，在 Minecraft 中添加服务器。

<tabs>
<template #局域网>

Overlay gateway 会为服务器分配专用的本地 IP 地址。此方式要求 Olares 运行在原生 Linux 主机上，并使用有线以太网。Wi-Fi 和 WSL 不支持 Overlay gateway。

1. 打开 **Settings** > **Network** > **Overlay gateway**。
2. 打开 **Enable overlay gateway**。此系统级服务需由 Super admin 启用，之后 Admin 和 Member 可以为自己的应用启用网关。
3. 在 **Applications** 中，为目标 Minecraft 实例启用 Overlay gateway。点击 **Confirm**，等待实例重启并恢复为 **Running**。
4. 复制 **Minecraft Java** 旁显示的地址，包括端口。例如，`192.168.1.100:25565`。

请使用页面当前显示的地址。应用重启或网络变化后，本地 IP 可能改变。

</template>
<template #LarePass-VPN>

通过 LarePass VPN，可以从其他网络连接服务器。使用 VPN 时可以保持 Overlay gateway 开启。

1. 在运行 Minecraft 的电脑上启用 [LarePass VPN](../manual/get-started/local-access.md#使用-larepass-专用网络)。

   ![在电脑上启用 LarePass VPN](/images/manual/get-started/larepass-vpn-desktop.png#bordered)

2. 打开 **Settings** > **Applications**，选择 Minecraft 实例，再打开 **Export ports**。记下 Minecraft 的 **Exported port**，使用此值而非内部的 **Port** 值。
3. 使用 Olares 域名和导出端口组成服务器地址。例如，Olares ID 为 `alex@olares.com` 时，使用 `alex.olares.com:<exported-port>`。将 `<exported-port>` 替换为上一步的值。

每个实例的导出端口可能不同。请使用目标实例显示的端口。

</template>
</tabs>

获取地址后：

1. 打开 Minecraft Java 版，点击 **Multiplayer**。

   ![Minecraft 多人游戏菜单](/images/manual/use-cases/minecraft-multiplayer-menu.png#bordered)

2. 点击 **Add Server**。
3. 在 **Server Name** 中填写服务器名称。
4. 在 **Server Address** 中填写所选连接方式对应的地址，然后点击 **Done**。

   ![添加 Minecraft 服务器](/images/manual/use-cases/minecraft-add-server.png#bordered)

5. 选择服务器，点击 **Join Server**。

   ![加入 Minecraft 服务器](/images/manual/use-cases/minecraft-join-server.png#bordered)

## 使用 Clone 创建另一台服务器

[Clone](../manual/olares/market/clone-apps.md) 可以在同一 Olares 设备上运行不同版本或不同模组配置的服务器。它创建具有独立配置和数据目录的新实例，不会复制原实例的世界和模组。

1. 先安装 Minecraft 主实例，然后在 Market 中打开 **My Olares**。
2. 找到 Minecraft，点击 **Open** 旁的下拉箭头，选择 **Clone**。
3. 填写唯一的 **New app title** 和 **Desktop shortcut name**，点击 **Confirm**。
4. 在 **Configure Environment Variables** 中，为新服务器选择 **VERSION**、**TYPE** 和加载器配置，点击 **Confirm**，等待安装完成。
5. 单独管理新实例，将它的模组上传到 `Data/<app-name>/data/mods`。需要局域网访问时，为该实例开启 Overlay gateway。

请使用目标实例的连接信息。Overlay gateway 地址使用端口 `25565`，VPN 连接所用的导出端口则单独分配。同时运行多个服务器会增加内存和 CPU 占用。

### 找到实例的数据目录 {#find-the-instances-data-directory}

1. 打开 Market，进入 **My Olares**。
2. 选择 Minecraft 实例，打开详情页。
3. 在页面 URL 中，找到来源名称后面的应用名称。例如，`/app/<source>/minecraftabc123` 中的应用名称为 `minecraftabc123`。
4. 打开 Files，进入 `Data/<app-name>/data`。在此示例中，目录为 `Data/minecraftabc123/data`。

即使实例使用了不同的显示标题，数据目录仍使用 URL 中的应用名称。

### 复制已有世界

要在另一个实例中继续游玩已有世界，请先保持游戏、加载器和模组版本一致。

1. 停止两个实例，分别备份整个数据目录。
2. 在 Files 中，用原实例 `data` 目录中的全部内容替换新实例的 `data` 目录内容，包括隐藏文件。
3. 检查新实例的环境变量。这些设置独立于复制的文件。保持 **ONLINE_MODE** 一致，以保留玩家 UUID 和背包的对应关系。检查复制过来的管理员、白名单和封禁名单。
4. 恢复运行新实例，使用它自己的连接信息加入。

:::warning 保留备份
两个实例使用独立的数据目录，游戏进度不会同步。修改游戏版本、加载器或模组前，请保留备份。需要回退升级时，恢复升级前的备份。不要将升级后的世界复制回低版本服务器。
:::

## 管理服务器

要查看日志或执行服务器命令，请从 Launchpad 打开 Minecraft 实例，使用控制台终端。

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
- 地址使用 Olares 域名，以及 **Settings** > **Applications** > 目标实例 > **Export ports** 中的 **Exported port**。
- 服务器已就绪，客户端匹配其游戏版本与模组要求。
- LarePass VPN 已启用。

### 升级应用会怎样？

升级应用会重启服务器，当前在线玩家会被断开连接。

## 了解更多

- [管理应用的 Overlay 网关](../manual/olares/settings/overlay-gateway.md)：为支持的应用配置局域网访问。
- [管理应用环境变量](../manual/olares/settings/manage-app-env.md)：修改设置，重启后生效。
- [克隆应用](../manual/olares/market/clone-apps.md)：创建并管理独立应用实例。
