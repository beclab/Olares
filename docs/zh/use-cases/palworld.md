---
outline: [2, 3]
description: 在 Olares 上托管 Palworld 专用服务器，与好友通过局域网或 VPN 联机游玩。
head:
  - - meta
    - name: keywords
      content: Olares, Palworld, 游戏服务器, 自托管, Overlay gateway, VPN, 游戏
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-08-19"
---
:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/palworld.md)为准。
:::

# 在 Olares 上与好友联机游玩 Palworld

Olares 上的 Palworld 运行的是专用 Palworld 服务器。你通过控制台终端管理服务器，玩家则使用官方 Palworld 客户端连接。你可以邀请好友通过 Overlay gateway 在本地网络中联机，也可以通过 LarePass VPN 让他们远程加入。

## 学习目标

通过本教程，你将学习如何：

- 从 Market 安装 Palworld 应用。
- 启用 Overlay gateway，让同一本地网络中的玩家可以连接。
- 通过本地网络或 VPN 连接到服务器。

## 准备工作

- **Olares OS**：Olares 版本为 1.12.6 或更高。
- **硬件与网络**：Olares 设备运行在原生的 Linux 主机上，并使用有线以太网连接。Overlay gateway 在 Wi-Fi 或 WSL 环境下无法工作。
- **权限**：需要 Super admin 开启系统级的 Overlay gateway 服务；服务开启后，Admin 或 Member 可以为 Palworld 启用 Overlay gateway。
- **客户端要求**：每位玩家的电脑上需有支持手动输入服务器地址的正版 Palworld 客户端。例如，Windows、Mac 或 Linux（通过 Steam Proton）上的 Steam 客户端均可。
    :::info 不支持的客户端
    Xbox、PS5 和 PC Game Pass 版只能通过游戏内的社区服务器列表加入服务器，且不能手动输入地址。该服务器未注册到社区服务器列表，因此这些版本无法加入。
    :::

## 安装 Palworld

1. 打开 Market，搜索 "Palworld"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。
   ![Palworld in Market](/images/manual/use-cases/palworld.png#bordered)

3. 打开应用，等待应用在 Launchpad 或 Settings 中的状态变为 **Running**。首次启动会下载数 GB 的服务器资源，可能需要一段时间。

## 为 Palworld 启用 Overlay gateway

Overlay gateway 会为 Palworld 分配一个专用的本地 IP 地址，让同一网络中的玩家可以直接连接。

1. 打开 Olares **Settings**，进入 **Network** > **Overlay gateway**。
2. 确认 **Enable overlay gateway** 开关已打开。这是系统级服务开关；如未开启，需由 Super admin 打开。
3. 在 **Applications** 列表中找到 **Palworld**，确认其状态为 **Running**，然后为该应用启用 Overlay gateway。
4. 在 **Palworld** 右侧，复制显示的地址。例如 `192.168.50.118:8211`。

:::info 本地 IP 地址 是动态的
本地 IP 地址 由 Overlay gateway 动态分配，应用重启或网络变化后可能会改变。请始终使用当前页面上显示的地址。
:::

## 从同一本地网络连接

与 Olares 设备处于同一本地网络的玩家，可以通过 Overlay 地址连接。

1. 从 **Settings** > **Network** > **Overlay gateway** 复制 Overlay 地址。例如 `192.168.50.118:8211`。
2. 启动 Palworld，在主菜单选择 **Join Multiplayer Game**。

   ![Palworld join multiplayer](/images/manual/use-cases/palworld-join-multiplayer.png#bordered){width=60%}

3. 在地址栏中输入 Overlay 地址。
4. 点击 **Connect**。

   <!-- ![Palworld connect](/images/manual/use-cases/palworld-connect.png#bordered) -->

## 通过 VPN 连接

当玩家与 Olares 设备不在同一本地网络时，可以使用此方法。

:::warning 使用 VPN 前需关闭 Overlay
Palworld 不能同时使用 Overlay gateway 和 VPN。通过 VPN 连接前，管理员必须在 **Settings** > **Network** > **Overlay gateway** 中关闭 Palworld 的 Overlay gateway。Palworld 应用本身需保持 **Running**。
:::

1. 确保已启用 [LarePass VPN](../manual/get-started/local-access.md#using-larepass-vpn)。
2. 启动 Palworld，选择 **Join Multiplayer Game**。
3. 在地址栏中，按以下格式输入：

   ```text
   <local-name>.olares.com:8211
   ```

   将 `<local-name>` 替换为你的 Olares ID 中的本地名称，即 `@` 前面的部分。例如，如果你的 Olares ID 是 `olarestest001@olares.com`，则服务器地址为 `olarestest001.olares.com:8211`。

4. 点击 **Connect**。

:::tip UDP 端口
Palworld 使用 UDP 端口 `8211`。Windows 上的 TCP 工具（如 `Test-NetConnection`）无法检测该 UDP 端口是否可达。
:::

## 管理服务器

Palworld 应用没有 Web 管理界面。要查看日志或执行服务器命令，请从 Launchpad 打开该应用，使用内置的控制台终端。你也可以通过游戏内聊天命令或容器的 REST API 管理部分游戏内操作。

## 常见问题

### 为什么我的 本地 IP 地址 与截图不同？

Overlay gateway 会动态分配 本地 IP 地址。连接时，请使用 **Settings** > **Network** > **Overlay gateway** > **Palworld** 中当前显示的地址。

### 为什么可以通过本地网络连接，但无法通过 VPN 连接？

请检查以下事项：

- 管理员已关闭 Palworld 的 Overlay gateway。
- 服务器地址格式为 `<local-name>.olares.com:8211`，其中 `<local-name>` 是 Olares ID 中 `@` 之前的部分。
- 你的 LarePass VPN 连接已启用。

### 升级应用会怎样？

升级应用会重启服务器，当前在线的玩家会被断开连接。重启可能还会触发 SteamCMD 更新，耗时比普通启动更长。

### 如何运行服务器命令？

从 Launchpad 打开 Palworld 应用，使用内置的控制台终端。部分命令也可以作为游戏内聊天命令输入，或者通过容器的 REST API 发送。

## 了解更多

- [管理应用的 Overlay 网关](/zh/manual/olares/settings/overlay-gateway.md)：为支持的应用配置局域网访问。