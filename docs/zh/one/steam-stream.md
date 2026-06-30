---
outline: [2, 3]
description: 关于在 Olares 上串流 Steam 游戏的完整教程。学习如何安装 Steam Headless、配置串流服务，以及在 Moonlight 上通过本地和远程网络进行游戏串流。
head:
  - - meta
    - name: keywords
      content: Steam, Steam Headless, Moonlight, Sunshine, 游戏串流, 远程畅玩, 云游戏
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/steam-stream.md)为准。
:::

# 将 Steam 游戏串流至任意设备 <Badge type="tip" text="30 min" />

Olares One 配备了 NVIDIA 5090 移动 GPU，配合 Steam Headless 应用，可化身为强大的游戏串流服务器。你可以通过 Moonlight 在任何兼容设备上畅玩喜爱的游戏。

## 学习目标
- 在 Olares One 上设置 Steam Headless。
- 配置 Sunshine 串流服务。
- 通过 Moonlight 在本地或远程网络配对 Olares One 并串流游戏。

## 前提条件
**硬件** <br>
- Olares One 已连接至稳定的网络（建议使用以太网）。
- 足够的磁盘空间用于下载游戏。

**软件** <br>
- 客户端设备已安装 [Moonlight](https://moonlight-stream.org/)。
- 有效的 Steam 账户。

**LarePass**（远程访问必需）
- 如果你计划在家庭网络外进行串流，客户端设备（桌面或移动端）需要安装 LarePass。

## 步骤 1：安装 Steam Headless

1. 打开 Market，搜索 "Steam"。
2. 点击 **Get**，然后点击 **Install**。
   ![安装 Steam Headless](/images/manual/use-cases/steam-install-steam-headless1.png#bordered)

3. 将出现提示，要求你设置环境变量。这将创建 Sunshine Web UI 的登录凭据：
    - `SUNSHINE_USER`：设置 Sunshine 访问的用户名。
    - `SUNSHINE_PASS`：设置安全密码。
      :::tip 记住你的登录凭据
      这是你首次登录 Sunshine 的初始凭据，首次登录时必须使用。
      :::
4. 等待安装完成。

## 步骤 2：安装 Steam 客户端
Headless 应用运行后，你需要在其中初始化 Steam 客户端。
1. 打开 Steam Headless，点击 **Connect**。
   ![连接 Steam](/images/manual/use-cases/steam-connect-to-steam.png#bordered)

2. Steam 客户端将自动开始下载和安装。
   ![安装 Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![更新 Steam](/images/manual/use-cases/steam-update-steam.png#bordered)

3. 安装完成后，将出现 Steam 登录界面。使用你的 Steam 账户登录。
   ![登录 Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

::: tip 安装失败时重试
如果 Steam 安装或更新因网络问题失败，请前往 Steam Headless 控制台左上角的菜单，导航至 **Applications** > **Internet** > **Steam** 重新启动安装。
:::

## 步骤 3：将 Sunshine 与 Moonlight 配对
Steam Headless 使用 Sunshine 进行视频串流。你必须将其与游玩设备上的 Moonlight 应用配对。

### 访问 Sunshine 控制台

你需要访问运行在 Olares One 上的 Sunshine Web UI，以输入配对 PIN 码。

1. 复制当前 Steam Headless 浏览器标签页的 URL。
2. 打开新的浏览器标签页，修改 URL 以访问端口 `47990`。地址因网络环境而异。

    - **同一网络**：使用 HTTPS 搭配 `.local` 地址。以下两种格式均可（主机名中使用点或连字符）：

   :::info 使用 HTTPS 访问 Sunshine
   大多数 Olares 服务可以通过 `.local` 地址使用 HTTP 访问。

   然而，Sunshine Web UI 需要 HTTPS 以确保本地通信安全。
   :::

   ```plain
   https://139ebc4f0.<your Olares ID>.olares.local:47990
   https://139ebc4f0-<your Olares ID>-olares.local:47990
   ```

    - **不同网络（通过 VPN）**：在设备上启用 LarePass VPN，然后追加 `:47990`。例如：

   ```plain
   https://139ebc4f0.<your Olares ID>.olares.com:47990
   ```

3. 按 **Enter** 打开 Sunshine 控制台页面。
4. 使用之前创建的 `SUNSHINE_USER` 和 `SUNSHINE_PASS` 凭据登录。
   ![登录 Sunshine](/images/manual/use-cases/steam-sign-in-to-sunshine.png#bordered)
5. 点击 **PIN** 标签页。页面将等待配对 PIN 码。
   ![Sunshine 上的 PIN](/images/manual/use-cases/steam-pin-on-sunshine.png#bordered)

### 在 Moonlight 中添加主机

当 Olares 主机与 Moonlight 处于同一本地网络时，Moonlight 通常会自动检测到它。

如果未出现，或者你正在通过不同网络连接，请按照以下步骤手动添加主机。以下流程使用 macOS 版 Moonlight 客户端。

1. 在客户端设备上打开 Moonlight。
2. 点击 **Add Host** 按钮，图标为带加号的电脑。
3. 输入你的 Olares 域名，不带 `https://` 前缀。例如：

   ```plain
   139ebc4f0.<your Olares ID>.olares.com
   ```
4. 点击 **OK**。将出现一个新的锁定主机图标。
5. 点击锁定图标。Moonlight 将显示一个 4 位配对 PIN 码。
   ![获取配对 PIN](/images/manual/use-cases/steam-get-pairing-pin.png#bordered)

### 完成配对

1. 返回浏览器中的 **Sunshine PIN** 页面。
2. 输入 Moonlight 中显示的 PIN 码，并为你的设备命名。
   ![输入配对 PIN](/images/manual/use-cases/steam-enter-pairing-pin.png#bordered)

3. 点击 **Send**。
4. 配对成功后，你将看到确认消息，Moonlight 中的锁定图标将消失。
   ![Moonlight 中的主机](/images/manual/use-cases/steam-host-in-moonlight.png#bordered)

配对完成后，即可开始串流。

## 步骤 4：串流游戏
:::tip 优化远程畅玩体验
为获得最佳的远程串流体验：
1. 将客户端设备连接至 5GHz Wi-Fi 或以太网。
2. 确保 **LarePass VPN** 已激活。
:::
你可以根据网络设置，通过 Moonlight 在本地或远程串流游戏。

以下步骤演示本地串流。

1. 在客户端设备上打开 Moonlight。
2. 选择你的 Olares 主机，点击 **Steam** 图标。
   ![Moonlight 中的 Steam](/images/manual/use-cases/steam-in-moonlight.png#bordered)

3. Steam **Big Picture Mode** 将启动。从库中选择游戏并开始游玩。

## 常见问题

### 为什么无法使用 `.local` 地址访问 Sunshine Web UI？

对于大多数 Olares 服务，你可以使用 HTTP 配合 `.local` 地址访问。Sunshine 是例外，它需要 HTTPS 来确保本地通信安全。如果你使用 `http://` 配合 `.local` URL，Sunshine 页面将无法加载。

请在浏览器地址栏中使用 `https://`（例如，`https://139ebc4f0.<your Olares ID>.olares.local:47990`）。

### 为什么游戏没有全屏显示？

这可能是由分辨率设置引起的。尝试调整分辨率：

- **在 Moonlight 中**：前往 **Settings** > **Basic Settings** > **Resolution and FPS**。
  ![Moonlight 中的显示设置](/images/manual/use-cases/steam-display-in-moonlight.png#bordered)
- **在 Steam 控制台页面**：前往 **Applications** > **Settings** > **Display**。
  ![Steam Headless 中的显示设置](/images/manual/use-cases/steam-display-in-steam-hd.png#bordered)

### 如何退出全屏串流？

使用以下快捷键：
- **Windows**：`Ctrl + Alt + Shift + Q`
- **Mac**：`Control (^) + Option (⌥) + Shift + Q`
- **移动端**：`Start + Select + L1 + R1`

结束后，退出 Steam Big Picture 模式以释放 Olares 上的系统资源。
![退出 Steam Big Picture 模式](/images/manual/use-cases/steam-exit-big-pic.png#bordered)

### 下载的游戏存储在哪里？

你可以在 Files 应用中查看已下载的游戏。默认情况下，游戏保存在：

```plain
/Cache/olares/steam-headless/c0/.steam/steam/steamapps/common
```

建议不要更改此默认目录。

### 为什么在 Moonlight 中重新配对主机时会报错？

如果你在 Moonlight 中删除了 Olares 主机并尝试重新配对，可能会遇到以下错误：

- **The PIN from the PC didn't match. Please try again.**
- **Request timed out (Error 4)**
- **Connection closed (Error 2)**

这通常发生在 Sunshine 服务无响应时。
要解决此问题，只需在 Olares 中重启 Steam Headless，然后重新尝试配对：

1. 从 Launchpad 打开 Control Hub。
2. 导航至 **Browser** > **steamheadless** > **Deployments** > **steamheadless** > **Restart**。
   ![重启 Steam Headless](/images/manual/use-cases/steam-restart.png#bordered)

3. 在确认提示中，输入 `steamheadless` 并点击 **Confirm**。
   ![确认重启](/images/manual/use-cases/steam-confirm-restart.png#bordered){width=80%}

4. 重启完成后，在 Moonlight 中再次与 Sunshine 配对。

### 如何更改 Sunshine 用户名或密码？

你可以直接从 Sunshine Web 控制台更改凭据。

1. 使用本地地址在浏览器中打开 Sunshine，例如：`https://139ebc4f0.<your Olares ID>.olares.local:47990`。
2. 使用当前用户名和密码登录。
3. 前往 **Change Password** 标签页。
   ![更改 Sunshine 密码](/images/manual/use-cases/steam-change-sunshine-pw.png#bordered)

4. 输入新密码（如需也可更改用户名），然后点击 **Save**。

## 资源
- [在 Olares One 上本地玩 Steam 游戏](steam-direct-play.md)
