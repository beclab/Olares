---
outline: [2, 3]
description: Olares 游戏串流教程。了解如何安装 Steam Headless，配置串流服务，并通过 Moonlight 在本地和远程网络上畅玩游戏。
--- 
# 使用 Steam Headless 串流你的游戏

想在 Olares 上畅享游戏乐趣吗？没有问题。通过 Steam Headless 应用，Olares 轻松变身强大的游戏串流服务器。现在你可以通过 Moonlight，在任何兼容设备上畅玩你喜欢的 Steam 游戏。

本教程将带你完成在 Olares 上安装 Steam Headless、配置 Steam 客户端、配对串流服务，以及连接 Moonlight 客户端开始游玩的完整流程。

## 目标
通过本教程，你将学习：

- 在 Olares 设备上安装并设置 Steam Headless。
- 配置 Sunshine 串流服务，并与 Moonlight 客户端配对。
- 在本地和远程网络下串流你的 Steam 游戏。

## 准备工作

开始前，请确保以下条件已满足：
- Olares 运行在配备 NVIDIA GPU 的机器上。
- 串流设备已安装 Moonlight 客户端。请访问 [Moonlight 官网](https://moonlight-stream.org/) 下载并安装相应版本。
- 拥有一个有效的 Steam 账号以访问你的游戏。
- 如果计划在家庭网络之外进行串流，请在客户端设备（桌面或移动端）上[启用 LarePass VPN](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
:::tip
仅在本地串流无需启用 LarePass VPN。
:::

## 安装并配置 Steam Headless

从 Olares 市场安装 Steam Headless，并在 Steam 客户端内完成初始设置。

### 安装 Steam Headless

1. 打开应用市场，搜索 “Steam”。
2. 点击**获取**，然后点击**安装**。
   ![安装 Steam Headless](/images/zh/manual/use-cases/steam-install-steam-headless.png#bordered)
3. 系统会弹出提示，要求你配置环境变量。这些变量将作为 Sunshine Web UI 的登录凭据：
    - `SUNSHINE_USER`：设置 Sunshine 访问用户名。
    - `SUNSHINE_PASS`：设置相应的密码。
      :::tip 记住你的登录凭据
      这是你登录 Sunshine 的初始凭据。稍后访问 Sunshine 控制台时会用到它们。
      :::
4. 等待安装完成。

### 完成 Steam 初始设置
登录 Steam 完成设置并访问你的游戏库。

1. 安装完成后，打开 Steam Headless 并点击 **Connect**。
   ![连接到 Steam](/images/zh/manual/use-cases/steam-connect-to-steam.png#bordered)
2. Steam 客户端将自动安装并更新。此过程可能需要几分钟。
   ![安装 Steam](/images/manual/use-cases/steam-install-steam.png#bordered)
   ![更新 Steam](/images/manual/use-cases/steam-update-steam.png#bordered)
3. 安装完成后会出现 Steam 登录界面。使用你的 Steam 账户登录。
   ![登录 Steam](/images/manual/use-cases/steam-sign-in-to-steam.png#bordered)

:::tip 安装失败请重试
如果由于网络问题导致 Steam 安装或更新失败，请在 Steam Headless 控制台左上角的菜单，导航至 **Applications** > **Internet** > **Steam** 重新启动安装。
:::

Steam 准备就绪后，你可以通过 Sunshine 连接 Moonlight 进行串流。

## 将 Sunshine 与 Moonlight 配对

Steam Headless 集成了开源串流服务器 Sunshine。要串流游戏，你需要先把它与你的 Moonlight 客户端配对。

### 访问 Sunshine 控制台

访问 Sunshine 控制台，将你的 Moonlight 客户端与 Olares 配对。

1. 在 Steam Headless 浏览器标签页中，复制其 URL。
2. 在浏览器中输入 Sunshine 地址。不同网络环境下地址不同。
   - <b>同一局域网</b>：将协议`https`改为`http`，将域名后缀更改为`.local`，并追加`:47990`。例如：
   ```
   http://139ebc4f0.<your Olares ID>.olares.local:47990
   ```

   - <b>不同网络（通过 VPN）</b>：在你的设备上[启用 LarePass VPN](../manual/larepass/private-network.md#在-larepass-中启用专用网络)，然后在地址末尾添加`:47990`。例如：
   
   ```
   https://139ebc4f0.<your Olares ID>.olares.com:47990
   ```

3. 按回车键打开 Sunshine 控制台页面。
4. 使用你之前创建的`SUNSHINE_USER`和`SUNSHINE_PASS`凭据登录。
   ![登录 Sunshine](/images/manual/use-cases/steam-sign-in-to-sunshine.png#bordered)
5. 点击 **PIN** 标签页。页面会进入等待配对 PIN 的状态。
   ![Sunshine 上的 PIN](/images/manual/use-cases/steam-pin-on-sunshine.png#bordered)

### 在 Moonlight 中添加主机
当 Moonlight 与你的 Olares 设备处于同一局域网时，通常会自动检测到你的 Olares 主机。
如果没有显示，或者你需要跨网络连接，请按以下步骤手动添加主机。以下示例使用 macOS Moonlight 客户端。

:::info 从不同网络连接？
如果你的 Moonlight 客户端和 Olares 设备不在同一网络，请先在运行 Moonlight 的设备上[启用 LarePass VPN](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
:::

1. 在你的串流设备上打开 Moonlight。
2. 点击 **Add Host** 按钮（外观为带有加号的电脑图标）。
3. 输入你的 Olares 域名，不要包含`https://`前缀。例如：

   ```
   139ebc4f0.<your Olares ID>.olares.com
   ```
4. 点击 **OK**。此时会出现一个带锁的主机图标。
5. 点击该带锁图标。Moonlight 将显示一个 4 位数的配对 PIN 码。
![获取配对 PIN 码](/images/manual/use-cases/steam-get-pairing-pin.png#bordered) 

### 完成配对
1. 回到浏览器中的 **Sunshine PIN** 页面。
2. 输入 Moonlight 显示的 PIN 码，并为你的设备命名。
![输入配对 PIN 码](/images/manual/use-cases/steam-enter-pairing-pin.png#bordered)
3. 点击 **Send** 。配对成功后，你将看到：

   ```text
   Success! Please check Moonlight to continue
   ```
   ![配对成功](/images/manual/use-cases/steam-pairing-succeeds.png#bordered)
4. 返回 Moonlight，你的主机图标应显示为已解锁的可用状态。
   ![Host in Moonlight](/images/manual/use-cases/steam-host-in-moonlight.png#bordered)

配对完成后，就可以开始串流了。

## 串流你的游戏

根据你的网络设置，你可以在局域网或远程网络中，通过 Moonlight 串流游戏。 

下面以局域网串流为例：

1. 在客户端设备上打开 **Moonlight**。
2. 选择已解锁的主机，然后点击 **Steam** 图标。
   ![Moonlight 里的 Steam](/images/manual/use-cases/steam-in-moonlight.png#bordered)

此时，Steam 将进入大屏幕模式，你可以开始畅玩游戏了。

:::tip 远程串流
要在家庭网络之外进行游戏，请在打开 Moonlight 之前[启用 LarePass VPN](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。

连接成功后，按照上述相同步骤即可远程串流你的游戏。
:::

## 常见问题

### 为什么游戏不能全屏显示？

这通常与分辨率设置有关。你可以尝试调整分辨率：

- <b>在 Moonlight 中</b>：进入 **Settings** > **Basic Settings** > **Resolution and FPS**。
![Moonlight 显示](/images/manual/use-cases/steam-display-in-moonlight.png#bordered)
- <b>在 Steam 控制台页面</b>：进入 **Applications** > **Settings** > **Display**.  
![Steam Headless 显示](/images/manual/use-cases/steam-display-in-steam-hd.png#bordered)

### 如何退出全屏串流？

使用以下快捷键：
- **Windows**：`Ctrl + Alt + Shift + Q`  
- **Mac**：`Control (^) + Option (⌥) + Shift + Q` 
- **移动端**：`Start + Select + L1 + R1`

结束游戏后，请退出 Steam 大屏幕模式，释放 Olares 上的系统资源。
   ![退出 Steam 大屏幕模式](/images/manual/use-cases/steam-exit-big-pic.png#bordered)

### 下载的游戏存储在哪里？

你可以在 Olares 文件管理器中查看已下载的游戏。默认保存路径为：

```
/Cache/olares/steam-headless/c0/.steam/steam/steamapps/common
```

建议不要更改此默认目录。

### 为什么在 Moonlight 里重新配对主机会报错？

如果你在 Moonlight 里删除 Olares 主机后再次配对，可能会遇到以下错误：

- **The PIN from the PC didn't match. Please try again.**
- **Request timed out (Error 4)**
- **Connection closed (Error 2)**

这通常是因为 Sunshine 服务未响应。

解决此问题，请重启 Olares 上的 Steam Headless，然后重新配对：

1. 从启动台打开控制面板。
2. 导航至**浏览** > **steamheadless** > **部署** > **steamheadless** > **重启**。
   ![重启 Steam Headless](/images/zh/manual/use-cases/steam-restart.png)
3. 在确认提示框中，输入`steamheadless`，点击 **Confirm**。
   ![确认重启](/images/zh/manual/use-cases/steam-confirm-restart.png#bordered){width=70%}
4. 重启完成后，在 Moonlight 与 Sunshine 重新配对。

### 如何修改 Sunshine 用户名或密码？

你可以在 Sunshine Web 控制台中直接修改凭据。

1. 使用本地地址在浏览器中打开 Sunshine，例如：`http://139ebc4f0.<your Olares ID>.olares.local:47990`。
2. 使用当前的用户名和密码登录。
3. 转到 **Change Password** 标签页。
   ![Change Sunshine password](/images/manual/use-cases/steam-change-sunshine-pw.png#bordered)
4. 输入新密码（如需更改用户名也可输入新用户名），然后点击 **Save**。