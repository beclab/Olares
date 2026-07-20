---
outline: [2, 4]
description: 在 Olares 上安装和运行 Windows 虚拟机的综合指南。学习如何配置初始凭据、通过基于浏览器的 VNC 或 Microsoft 远程桌面（RDP）连接，以及在计算机和 VM 之间传输文件。
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../use-cases/windows.md)。
:::

# 在 Olares 设备上运行 Windows VM

Olares 允许您直接在设备上运行完整的 Windows 虚拟机，为您提供可从 macOS、Windows 或 Linux 访问的个人、始终可用的 Windows 环境。

:::info 系统功能
- Olares 支持运行基本的 Windows 应用程序。
- 默认情况下，Windows VM 使用基于 CPU 的虚拟化和虚拟显示输出。
- Intel 集成显卡支持仅在支持的硬件上可用，需要额外的主机配置。请参阅[为 Windows VM 启用 Intel 集成显卡直通](./windows-intel-gpu-passthrough.md)。
- 仅通过远程桌面（RDP）连接时才支持音频输出。
:::

本指南将引导您完成安装 Windows VM、启用安全网络以及使用远程桌面获得最佳体验。

## 学习目标

完成本教程后，您将学习如何：
- 在 Olares 设备上安装和设置 Windows VM。
- 使用基于浏览器的 VNC 查看器或 Microsoft 远程桌面（RDP）访问 Windows VM。
- 从 VM 内部更改 Windows 登录密码。
- 在您的计算机和 Windows VM 之间无缝传输文件。

## 安装和配置 Windows VM

Windows 在 Olares Market 中作为应用提供。

### 安装 Windows
1. 打开 Market，搜索 "Windows"。
2. 点击 **Get**，然后点击 **Install**。
   ![Install Windows](/images/manual/use-cases/win-install1.png#bordered)

3. 当提示时，设置环境变量：
    - **USERNAME：** 创建用于访问 Windows 的用户名。
    - **PASSWORD：** 设置相应的密码。
    - **VERSION：** 从下拉列表中选择您偏好的 Windows 版本。
    - **DISK_SIZE：** 为 Windows 分配磁盘空间。

    ![Set environment variables](/images/manual/use-cases/win-set-env-var1.png#bordered){width=70%}

4. 等待几分钟，让安装和初始化完成。

### 设置 Windows

安装完成后，从 Launchpad 打开 Windows 以首次启动 VM。

Olares 将自动下载并安装相应 Windows 版本的系统镜像。这可能需要几分钟，具体取决于您的网络速度。

![Download Windows 11](/images/manual/use-cases/win-downloading-win11.png#bordered)
## 访问 Windows VM

您可以通过两种方式访问您的 VM：
- [**浏览器：**](#方法-1-从浏览器访问-vnc) 用于设置和快速任务
- [**远程桌面：**](#方法-2-使用远程桌面客户端访问-rdp) 用于最佳的日常体验

### 方法 1：从浏览器访问（VNC）

从 Launchpad 打开 Windows 应用，以直接在浏览器中使用 VNC 启动 VM。
::: info
VNC（Virtual Network Computing）提供即时、无需客户端的访问，无需任何额外软件。它非常适合初始设置、故障排除或无法使用 RDP 时的紧急访问。但是，它可能感觉响应性较差，并且缺乏音频重定向和高性能图形等高级功能。
:::
### 方法 2：使用远程桌面客户端访问（RDP）
RDP（Remote Desktop Protocol）提供更流畅、类似原生的体验，具有更好的性能、音频支持和无缝文件传输。

#### 查找 Windows 的端口号
:::warning 多个 Windows 实例
每个 Windows 实例使用唯一的端口。如果您克隆了 Windows 应用，请确保检查您想要访问的特定实例的 **ACLs** 部分。
:::
1. 打开 Settings，导航到 **Application** > **Windows**。
2. 在 **Permissions** 下，点击 **ACLs**。
3. 记下 **Port** 列中列出的端口号。连接步骤中需要用到它。
   ![Locate port number](/images/manual/use-cases/win-port-number.png#bordered){width=90%}

#### 通过 RDP 连接到 Windows
:::info
以下步骤显示 macOS 界面，但所有平台上的工作流程类似。
:::
1. 在您的设备上[在 LarePass 上启用 VPN](../manual/larepass/private-network.md#enable-vpn-on-larepass)。

    当 VPN 连接状态显示 **P2P** 或 **Intranet** 时，安全网络已激活并准备好进行远程访问。

2. 安装远程桌面客户端。
   - **Windows：** 无需安装。
   - **macOS / iOS：** 从 App Store 下载 [Windows App](https://apps.apple.com/us/app/windows-app/id1295203466)。
   - **Android：** 从 Google Play 下载 [Windows App](https://play.google.com/store/apps/details?id=com.microsoft.rdc.androidx)。

3. 在浏览器中从 Launchpad 打开 Windows。从地址栏复制域名（排除 `https://` 和域名后的任何文本）。
   ![Domain address](/images/manual/use-cases/win-url.png#bordered)

4. 将您的 Windows VM 添加为 RDP 连接。

    a. 在您的设备上打开 Windows App。

    b. 点击 **＋** 图标并选择 **Add PC**。

    c. 在 **PC name** 中，输入上一步获取的域名，后跟冒号和端口号。

      例如，如果您的 URL 是 `https://0f4137ed.<username>.olares.com`，端口是 `47374`，请输入：
      ```
      0f4137ed.<username>.olares.com:47374
      ```

   ![Add PC](/images/manual/use-cases/win-add-pc1.png#bordered)

    d. 点击 **Add**。

5. 连接到 Windows VM。

   a. 双击您保存的 PC 条目，或点击 **⋯** 并选择 **Connect**。
   ![Connect to PC](/images/manual/use-cases/win-connect-device1.png#bordered)
        
   b. 当提示时，输入您之前创建的 **Username** 和 **Password**。
   ![Log in to PC](/images/manual/use-cases/win-log-in1.png#bordered)

   c. 如果出现安全警告，点击 **Continue**。
   ![Continue to log in](/images/manual/use-cases/win-confirm-connect1.png#bordered)

您现在已通过 RDP 连接到您的 Windows VM。
![Windows VM](/images/manual/use-cases/win-vm-interface.png#bordered)

## 可选：更改 Windows 登录密码

您可以直接从 VM 内部更新 Windows 登录密码：
1. 点击 Windows 任务栏中的搜索栏并输入 "password"。
2. 选择 **Change your password**。
   ![Change your password](/images/manual/use-cases/win-change-pw.png#bordered)
3. 点击 **Change** 设置新密码。
   ![Set new password](/images/manual/use-cases/win-set-pw.png#bordered)

## 在计算机和 Windows 之间传输文件

RDP 支持基于剪贴板的文件传输。

您可以：
- 在 Mac 或 PC 上复制任何文件。
- 直接将其粘贴到 Windows VM 中。

文件会立即出现在 Windows 中并可供使用。

## 断开与 Windows VM 的连接

要结束 RDP 会话，只需关闭 RDP 窗口。

Windows VM 会继续在您的 Olares 设备上运行，并随时准备让您重新连接。

## 常见问题

### Windows VM 显示空白屏幕或无桌面

浏览器可能由于不活动而暂停了 VNC 连接以节省系统资源。
   ![Reconnect VM](/images/manual/use-cases/win-vnc-reconnect.png#bordered)

点击 **Connect** 恢复会话。

### Windows 系统镜像下载失败

如果 Windows 系统镜像在设置期间下载失败：

- 稍等片刻，然后重启应用：
    1. 从 Launchpad 打开 Control Hub。
    2. 选择 windows 项目。
    3. 在 **Deployment** 下，点击 windows。
    4. 点击 **Restart**。
    ![Restart VM](/images/manual/use-cases/win-restart.png#bordered)

  重启后，系统镜像下载将自动重试。
- 如果反复失败，您的 IP 可能由于在短时间内多次下载尝试而被 Microsoft 暂时阻止。
  等待 **24 小时**，然后重启或重新安装应用并重试。
- 如果问题仍然存在，请联系我们寻求帮助。

### 我可以安装其他 Windows 版本或语言吗？

目前支持以下 Windows 版本：
- Windows 11 Pro
- Windows 11 LTSC
- Windows 11 Enterprise
- Windows 10 Pro
- Windows 10 LTSC
- Windows 10 Enterprise
- Windows 8.1 Enterprise
- Windows 7 Ultimate
- Windows Vista Ultimate
- Windows 2000 Professional
- Windows Server 2025
- Windows Server 2022
- Windows Server 2019
- Windows Server 2016
- Windows Server 2012
- Windows Server 2008
- Windows Server 2003

安装 Windows 后，您可以使用标准 Windows 语言设置更改显示语言。

## 了解更多

- [为 Windows VM 启用 Intel 集成显卡直通](./windows-intel-gpu-passthrough.md)
- [在 Olares 设备上运行 macOS VM](./macos.md)
