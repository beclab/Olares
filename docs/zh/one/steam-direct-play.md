---
outline: [2, 3]
description: 了解如何通过连接显示器、键盘和鼠标，在 Olares One 上直接玩 Steam 游戏。
head:
  - - meta
    - name: keywords
      content: Steam, 本地游戏, 直接游玩, Linux 游戏
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/steam-direct-play.md)。
:::

# 在 Olares One 上本地玩 Steam 游戏 <Badge type="tip" text="15 min" />

将显示器、键盘和鼠标连接到 Olares One，直接在设备上玩 Steam 游戏。

## 前提条件

**硬件** <br>
- Olares One 连接到稳定的网络（建议使用以太网）。
- 显示器、键盘和鼠标连接到 Olares One。
- 足够的磁盘空间用于下载游戏。

**软件** <br>
- 一个有效的 Steam 账户。

## 步骤 1：安装 Steam Headless

1. 打开 Market，搜索 "Steam"。
2. 点击 **Get**，然后点击 **Install**。
   ![安装 Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. 会出现提示要求你设置环境变量：
   - `SUNSHINE_USER`：设置 Sunshine 访问的用户名。
   - `SUNSHINE_PASS`：设置一个安全的密码。
4. 等待安装完成。

## 步骤 2：安装 Steam 客户端

1. 打开 Steam Headless，点击 **Connect**。
   ![连接到 Steam](/images/manual/use-cases/steam-connect-to-steam.png#bordered)

2. Steam 客户端将自动开始下载和安装。
   ![安装 Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![更新 Steam](/images/manual/use-cases/steam-update-steam.png#bordered)

3. 安装完成后，使用你的 Steam 账户登录。
   ![登录 Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

## 步骤 3：下载并玩游戏

1. 在 Steam 中，进入 **Library** 查看你已购买的游戏。
2. 选择你想玩的游戏，点击 **Install**。
3. 等待下载完成，然后点击 **Play**。

## 常见问题

### 首次连接或 Steam 重启后键盘或鼠标无响应

如果你的键盘或鼠标在首次连接时或 Steam 重启后没有响应，请拔掉设备并重新插入。

这是 Steam 启动期间设备检测的已知问题。重新连接设备会触发再次检测，使键盘或鼠标可用。

### 为什么我的显示器在我不玩游戏时也显示 Steam 界面？

Olares One 连接到显示器时通常显示终端提示符。然而，运行 Steam 应用会激活一个接管显示的图形界面。

要将显示器恢复为标准终端视图，请通过 **Market** 或 **Settings** 停止 Steam 应用。

## 资源

- [将 Steam 游戏串流到任何设备](steam-stream.md)
