---
outline: [2, 3]
description: 在 Olares 上托管 Minecraft Java 版服务器，与好友通过局域网或 VPN 联机游玩。
head:
  - - meta
    - name: keywords
      content: Olares, Minecraft, Minecraft Java, 游戏服务器, 自托管, Overlay gateway, VPN, 游戏
app_version: "1.0.0"
doc_version: "1.0"
doc_updated: "2026-08-19"
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../use-cases/minecraft.md)为准。
:::

# 在 Olares 上与好友联机游玩 Minecraft Java 版

Olares 上的 Minecraft 运行的是官方原版 Minecraft Java 版专用服务器。你通过控制台终端管理服务器，玩家则使用官方 Minecraft Java 版客户端连接。你可以邀请好友通过 Overlay gateway 在本地网络中联机，也可以通过 LarePass VPN 让他们远程加入。

## 学习目标

通过本教程，你将学习如何：

- 从 Market 安装 Minecraft 应用。
- 启用 Overlay gateway，让同一本地网络中的玩家可以连接。
- 通过本地网络或 VPN 连接到服务器。

## 准备工作

- Olares 版本为 1.12.6 或更高。
- Olares 设备运行在原生的 Linux 主机上，并使用有线以太网连接。Overlay gateway 在 Wi-Fi 或 WSL 环境下无法工作。
- 需要管理员权限，才能在 Olares 上安装 Minecraft 并启用 Overlay gateway 服务。
- 每位玩家的电脑上已安装 Minecraft Java 版。Bedrock、主机和移动版无法连接。客户端版本必须与 Market 页面上显示的 Minecraft **App version** 一致。

## 安装 Minecraft

1. 打开 Market，搜索 "Minecraft"。
2. 点击 **Get**，然后点击 **Install**，等待安装完成。
   ![Minecraft in Market](/images/manual/use-cases/minecraft.png#bordered)

3. 打开应用，等待服务器状态变为 **Running**。首次启动会下载服务器资源，可能需要几分钟。

## 为 Minecraft 启用 Overlay gateway

Overlay gateway 会为 Minecraft 分配一个专用的本地 IP 地址，让同一网络中的玩家可以直接连接。

1. 打开 Olares **Settings**，进入 **Network** > **Overlay gateway**。
2. 打开 **Enable overlay gateway** 开关。
3. 在 **Applications** 列表中找到 **Minecraft**，确认其状态为 **Running**，然后为该应用启用 Overlay gateway。
4. 在 **Minecraft Java** 右侧，复制显示的地址。例如 `192.168.50.219:25565`。

:::info LAN IP 是动态的
LAN IP 由 Overlay gateway 动态分配，应用重启或网络变化后可能会改变。请始终使用当前页面上显示的地址。
:::

## 从同一本地网络连接

与 Olares 设备处于同一本地网络的玩家，可以通过 Overlay 地址连接。

1. 从 **Settings** > **Network** > **Overlay gateway** 复制 Overlay 地址。例如 `192.168.50.219:25565`。
2. 打开 Minecraft Java 版，点击 **Multiplayer**。

   ![Minecraft multiplayer menu](/images/manual/use-cases/minecraft-multiplayer-menu.png#bordered)

3. 点击 **Add Server**。
4. 输入服务器信息，然后点击 **Done**：

   - **Server Name**：输入一个便于识别的名称。
   - **Server Address**：输入刚才复制的 Overlay gateway 地址。

   ![Minecraft add server](/images/manual/use-cases/minecraft-add-server.png#bordered)

5. 选择服务器，然后点击 **Join Server**。

   ![Minecraft join server](/images/manual/use-cases/minecraft-join-server.png#bordered)

## 通过 VPN 连接

当玩家与 Olares 设备不在同一本地网络时，可以使用此方法。

:::tip 保持 Overlay 开启
你可以保持 Minecraft 的 Overlay gateway 开启，它与 VPN 连接不冲突。
:::

1. 确保已启用 [LarePass VPN](../manual/get-started/local-access.md#using-larepass-vpn)。
2. 打开 Minecraft Java 版，点击 **Multiplayer**。
3. 点击 **Add Server**。
4. 在 **Server Address** 中，按以下格式输入：

   ```text
   <local-name>.olares.com:25565
   ```

   将 `<local-name>` 替换为你的 Olares ID 中的本地名称，即 `@` 前面的部分。例如，如果你的 Olares ID 是 `olarestest001@olares.com`，则服务器地址为 `olarestest001.olares.com:25565`。

5. 保存服务器，然后点击 **Join Server**。

## 管理服务器

Minecraft 应用没有 Web 管理界面。要查看日志或执行服务器命令，请从 Launchpad 打开该应用，使用内置的控制台终端。

## 常见问题

### 为什么我的 LAN IP 与截图不同？

Overlay gateway 会动态分配 LAN IP。连接时，请使用 **Settings** > **Network** > **Overlay gateway** > **Minecraft** 中当前显示的地址。

### 为什么客户端提示版本不匹配？

请将 Minecraft Java 版客户端更新到与 Market 页面上服务器版本一致的版本。服务器和客户端版本必须匹配。

### 为什么无法通过 VPN 地址连接？

请检查以下事项：

- 你使用的是 Minecraft Java 版。
- 服务器地址格式为 `<local-name>.olares.com:25565`，其中 `<local-name>` 是 Olares ID 中 `@` 之前的部分。
- 你的 LarePass VPN 连接已启用。

### 升级应用会怎样？

升级应用会重启服务器，当前在线的玩家会被断开连接。

## 了解更多

- [管理应用的 Overlay 网关](/zh/manual/olares/settings/overlay-gateway.md)：为支持的应用配置局域网访问。