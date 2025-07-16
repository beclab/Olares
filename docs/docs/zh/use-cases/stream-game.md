---
outline: [2, 3]
description: 在 Olares 上安装 Steam Headless，配置串流服务，并使用 Moonlight 从本地或远程网络串流 Steam 游戏。
--- 

# 使用 Steam Headless 串流喜爱的游戏

想要利用 Olares 的强大性能放松一下？没有问题。借助 Steam Headless 应用，Olares 轻松化身为 Steam 串流服务器，让你可以通过 Moonlight 或 Steam Link 客户端在任意兼容设备上畅玩你最爱的游戏大作。

本教程将带你完成 Steam Headless 应用的安装、串流服务的配置、以及通过 Moonlight 客户端进行游戏串流。

## 目标
通过本教程，你将学习：

- 在 Olares 上安装 Steam Headless，并在 Steam 客户端上打开 Windows 游戏的兼容性。  
- 配置 Sunshine 串流服务并与 Moonlight 客户端上的主机配对。  
- 通过 Moonlight 客户端在本地或远程网络上进行游戏串流。 
  
## 准备工作

开始前，请确保以下条件已满足：
- Olares 已在配有 NVIDIA 显卡的主机上运行。
- 串流设备已安装 Moonlight 客户端。可访问 [Moonlight 官网](https://moonlight-stream.org/)下载适合你设备的客户端并安装。
- 确保串流设备已安装 LarePass 并[启用专用网络](../manual/larepass/private-network.md#在-larepass-中启用专用网络)。
- 拥有一个有效的 Steam 账号以访问你的游戏。

## 安装 Steam Headless

1. 打开 Olares 应用市场，在“娱乐”分类下找到 Steam Headless 并点击**获取**。
2. 安装完成后，打开应用，然后点击 **Connect** 进入 Steam Headless 的控制台。
3. 点击 **Install** 按钮以安装并更新 Steam 客户端。安装完毕后会自动跳转至 Steam 登录页面。
   ![安装 Steam](/images/manual/tutorials/install-steam-client.png#bordered)

4. 登录你的 Steam 账号并完成基本设置。

   ![Steam 登录界面](/images/zh/manual/tutorials/steam-login.png#bordered)

::: tip 重试安装
由于国内网络环境限制，Steam 客户端安装和更新可能失败，此时可从 Steam Headless 控制台左上角进入 **Applications** > **Internet** > **Steam** 以重新安装。多次尝试一般可以成功。
:::

## 设置 Steam 游戏兼容性

Olares 运行于 Linux 环境，需要通过 [Proton](https://github.com/ValveSoftware/Proton) 兼容层运行 Steam 上的 Windows 平台游戏。

1. 在 Steam 客户端页面左上角，点击 **Steam** > **设置**。
2. 点击**兼容性**选项，打开**为所有其他产品启用 Steam Play**。
   ![Steam 设置](/images/zh/manual/tutorials/steam-setting.png#bordered)
3. 保存设置后重启 Steam 客户端即可查看全部游戏库。

## 配置串流服务

Steam Headless 集成了 Sunshine 串流服务器。要使用 Moonlight 客户端串流，还需在 Moonlight 上将游戏主机与 Sunshine 配对。

### 准备配对

1. 从浏览器获取 Steam Headless 页面的 URL，并添加端口号 `:47990`，如 `https://139ebc4f0.<你的olares ID>.olares.cn:47990`。通过该网址访问 Sunshine 串流服务器的控制页面。
   
   ![Sunshine 控制台](/images/manual/tutorials/access-sunshine.png#bordered)
   
2. 首次访问时，请使用以下默认凭据登录：  
   - 用户名：`sam`  
   - 密码：`password`
3. 点击 **Pin** 标签进入配对页面，你将看到输入配对码的提示。
   
   ![Sunshine 配对页面](/images/manual/tutorials/pin-sunshine.png#bordered)

### 在 Moonlight 端添加主机

1. 在串流设备上打开 Moonlight 客户端，点击右上角 <i class="material-symbols-outlined">add_to_queue</i> 按钮添加主机。
2. 输入主机 IP 地址，即 Steam 的本地 URL：`139ebc4f0.<你的olares ID>.olares.cn`。
   
   ::: tip 注意
   仅需填入 URL 部分（无需 `https://`）。
   :::

3. 点击**确定**，界面上会出现一个锁定状态的主机图标。
4. 点击主机图标获取配对码。
   
   ![获取配对码](/images/manual/tutorials/get-pin-code.png#bordered)

### 完成配对

1. 在 Sunshine 的配对页面中输入配对码和设备名称。 
2. 点击 **Send** 完成配对。配对成功后，你将看到提示信息："Success! Please check Moonlight to continue"。
3. 返回 Moonlight 界面，主机图标应变为激活状态。  
   
   ![配对成功](/images/manual/tutorials/active-host-moonlight.png#bordered)  

## 开始串流

配置完成后，尽情享受串流游戏的乐趣吧。

### 本地串流

如果你的串流设备与 Olares 处于同一局域网中：

1. 在串流设备上打开 Moonlight 客户端。  
2. 点击主机图标，之后点击 Steam 图标进入 Steam Big Picture 模式开始游戏。  
   
   ![进入串流](/images/manual/tutorials/stream-success.png#bordered) 

### 远程串流

借助 Olares 的专用网络，即使不在同一网络中，也能获得流畅的游戏串流体验。

启用专用网络的步骤如下：

<!--@include: ./remote.reusables.md{4,24}-->

开启专用网络后，操作步骤与本地串流一致。

## 常见问题

### 为什么我看到的画面不是全屏？

可能是分辨率设置问题。你可以尝试以下方式：

- 在 Moonlight 中进入**设置** > **基本设置** > **分辨率和帧率**进行调整。
- 从 Steam Headless 控制台左上角进入 **Applications** > **Settings** > **Display** 中修改分辨率。  
   
   ![调整显示设置](/images/manual/tutorials/set-steam-display.png#bordered)

### 如何在全屏模式下退出串流？
   
要退出串流的游戏画面：
- **Windows**：请使用快捷键组合 **Ctrl + Alt + Shift + Q**。
- **Mac**：请使用快捷键组合 **Control (^) + Option (⌥) + Shift + Q**. 
- **移动设备**：请使用手柄按键组合 **Start + Select + L1 + R1**。

游戏结束后，建议退出 Steam Big Picture 模式以释放 Olares 系统资源。

### 我下载的游戏存在哪里？

默认情况下，默认情况下，游戏下载目录为：
 
 `/Cache/olares/steam-headless/c0/.steam/debian-installation/steamapps/common`。
 
建议不要修改默认下载路径。

