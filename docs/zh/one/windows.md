---
outline: [2, 4]
description: 在 Olares 上安装和运行 Windows 虚拟机的完整指南。了解如何配置初始凭据、通过基于浏览器的 VNC 或 Microsoft 远程桌面（RDP）连接，以及在计算机和 VM 之间传输文件。
head:
  - - meta
    - name: keywords
      content: Olares One, Windows VM
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/windows.md)为准。
:::

# 在 Olares One 上运行 Windows 虚拟机 <Badge type="tip" text="20 min" />

Olares 允许你直接在设备上运行完整的 Windows 虚拟机，为你提供一个个人化的、始终可用的 Windows 环境，可从 macOS、Windows 或 Linux 访问。

:::info 系统能力
- Olares 支持运行基本的 Windows 应用程序。
- 工作流仅限于 **CPU 或集成显卡性能**。尚不支持 GPU 直通，这意味着重度 GPU 加速应用程序可能无法达到最佳性能。
- 仅通过远程桌面（RDP）连接时支持音频输出。
:::

本指南将引导你完成安装 Windows 虚拟机、启用安全网络，以及使用浏览器或远程桌面进行连接。

:::warning 升级到 1.12.5 后的已知问题
如果你从 Olares 1.12.4 升级到 1.12.5，并在设备重启后无法远程连接 Windows，请参阅[已知问题](/use-cases/windows-issues.md#windows-cannot-be-remotely-accessed-after-upgrading-from-1124-to-1125)。
:::

## 学习目标

完成本教程后，你将学会如何：
- 在 Olares One 上安装并设置 Windows 虚拟机。
- 使用基于浏览器的 VNC 查看器或 Microsoft 远程桌面（RDP）访问 Windows 虚拟机。
- 从虚拟机内部更改 Windows 登录密码。
- 在计算机和 Windows 虚拟机之间无缝传输文件。

## 安装并配置 Windows 虚拟机

Windows 在 Olares Market 中以应用形式提供。

### 安装 Windows
1. 打开 Market，搜索 "Windows"。
   ![安装 Windows](/images/one/win-market.png#bordered)

2. 点击 **Get**，然后点击 **Install**。
3. 出现提示时，设置环境变量：
    - `USERNAME`：创建用于访问 Windows 的用户名。
    - `PASSWORD`：设置相应的密码。
    - `VERSION`：从下拉列表中选择你偏好的 Windows 版本。
    - `DISK_SIZE`：使用 `G` 为单位为 Windows 分配磁盘空间，例如 `128G`。

    ![设置环境变量](/images/manual/use-cases/win-set-env-var1.png#bordered){width=70%}

4. 等待几分钟，让安装和初始化完成。

### 设置 Windows

安装完成后，从 Launchpad 打开 Windows 以首次启动虚拟机。

Olares 将自动下载并安装相应 Windows 版本的系统镜像。这可能需要几分钟，具体取决于你的网络速度。

![下载 Windows 11](/images/one/win-download.png#bordered){width=90%}
## 访问 Windows 虚拟机

你可以通过两种方式访问虚拟机：
- [浏览器](#方法-1从浏览器访问-vnc)：用于初始设置或快速任务。
- [远程桌面](#方法-2使用远程桌面客户端-rdp-访问)：推荐日常使用，支持音频、文件传输功能，交互更流畅。

### 方法 1：从浏览器访问（VNC）

从 Launchpad 打开 Windows 应用，以直接在浏览器中使用 VNC 启动虚拟机。
:::info
VNC（Virtual Network Computing）提供即时、无需客户端的访问，无需额外软件。它非常适合初始设置、故障排查或无法使用 RDP 时的紧急访问。然而，它可能感觉响应较慢，且缺少音频重定向和高性能图形等高级功能。
:::
### 方法 2：使用远程桌面客户端访问（RDP）

RDP（Remote Desktop Protocol）提供更流畅、类似原生的体验，具有更好的性能、音频支持和文件传输功能。

#### 前提条件

连接之前，确保你已准备好以下内容：
- 安全隧道：在设备上[启用 LarePass VPN](../../manual/larepass/private-network.md#enable-vpn-on-larepass)。
- RDP 客户端：在你要连接的设备上下载 Microsoft 远程桌面客户端。
    - **Windows**：无需安装。
    - **macOS / iOS**：从 App Store 下载 [Windows App](https://apps.apple.com/us/app/windows-app/id1295203466)。
    - **Android**：从 Google Play 下载 [Windows App](https://play.google.com/store/apps/details?id=com.microsoft.rdc.androidx)。

#### 查找 Windows 端口号

:::warning 多个 Windows 实例
如果你有多个 Windows 实例，每个实例使用不同的端口。
确保你检查的是要连接的特定实例的 **ACLs**。
:::

1. 打开 Settings，导航至 **Application** > **Windows**。
2. 在 **Permissions** 下，点击 **ACLs**。
3. 记下 **Port** 列中列出的端口号。连接步骤中需要用到它。
   ![查找端口号](/images/one/win-port-number.png#bordered)

#### 连接到 Windows

:::info
以下步骤展示 macOS 界面，但所有平台的操作流程相似。
:::

1. 从 Launchpad 打开 Windows。从地址栏复制域名（排除 `https://` 和域名后的任何文本）。
   ![域名地址](/images/one/win-url.png#bordered)
2. 在设备上打开 Windows App（或 Remote Desktop）。
3. 将你的 Windows 虚拟机添加为 RDP 连接。

    a. 点击右上角的 <i class="material-symbols-outlined">add</i> 并选择 **Add PC**。

    b. 在 **PC name** 中，输入上一步获取的域名，后跟冒号和端口号。

      例如，如果你的 URL 是 `https://0f4137ed.<username>.olares.com`，端口是 `47374`，则输入：
      ```text
      0f4137ed.<username>.olares.com:47374
      ```

   ![添加 PC](/images/one/win-add-pc.png#bordered)

    c. 点击 **Add**。

4. 连接到 Windows 虚拟机。

   a. 双击你保存的 PC 条目，或点击 **⋯**，然后选择 **Connect**。
   ![连接到 PC](/images/one/win-connect.png#bordered){width=90%}

   b. 出现提示时，输入你之前创建的 **Username** 和 **Password**。
   ![登录 PC](/images/one/win-login.png#bordered){width=90%}

   c. 如果出现安全警告，点击 **Continue**。
   ![继续登录](/images/one/win-warning.png#bordered){width=90%}

你现在已通过 RDP 连接到你的 Windows 虚拟机。
![Windows 虚拟机](/images/one/win-desktop.png#bordered)

## 常见任务

### 更改 Windows 登录密码

你可以直接从虚拟机内部更新 Windows 登录密码：
1. 点击 Windows 任务栏中的搜索框并输入 "password"。
2. 选择 **Change your password**。
3. 点击 **Change** 设置你的新密码。

### 在计算机和 Windows 之间传输文件

RDP 支持基于剪贴板的文件传输。

通过 RDP 剪贴板传输文件：
1. 在你的 Mac 或 PC 上复制文件。
2. 将其粘贴到 Windows 虚拟机中。

文件将立即出现在 Windows 中，随时可用。

### 断开与 Windows 虚拟机的连接

要结束 RDP 会话，只需关闭 RDP 窗口。

Windows 虚拟机继续在 Olares 设备上运行，随时等待你重新连接。

## 常见问题

### Windows 虚拟机显示空白屏幕或无桌面

浏览器可能因不活动而暂停了 VNC 连接以节省系统资源。
    ![重新连接虚拟机](/images/manual/use-cases/win-vnc-reconnect.png#bordered){width=90%}

点击 **Connect** 恢复会话。

### Windows 系统镜像下载失败

如果 Windows 系统镜像在设置期间下载失败：

- 稍等片刻，然后重启应用：
    1. 从 Launchpad 打开 Control Hub。
    2. 选择 Windows 项目。
    3. 在 **Deployment** 下，点击 **windows**。
    4. 点击 **Restart**。
    ![重启虚拟机](/images/one/win-restart.png#bordered)

  重启后，系统镜像下载将自动重试。

- 如果反复失败，你的 IP 可能因短时间内多次下载尝试而被 Microsoft 暂时屏蔽。
  等待 **24 小时**，然后重启或重新安装应用并重试。
- 如果问题持续存在，请联系我们获取帮助。

### 可以安装其他 Windows 版本或语言吗？

目前支持以下 Windows 版本：
- **Windows 11**：Pro, LTSC, Enterprise
- **Windows 10**：Pro, LTSC, Enterprise
- **Legacy Desktop**：Windows 8.1, 7 Ultimate, Vista Ultimate, 2000 Professional
- **Windows Server**：2025, 2022, 2019, 2016, 2012, 2008, 2003

Windows 安装完成后，你可以使用标准 Windows 语言设置更改显示语言。

## 资源

- [为 Windows 虚拟机启用 Intel 集成 GPU 直通](/use-cases/windows-intel-gpu-passthrough.md)：安装 Intel 显卡驱动并验证 Windows 中集成 GPU 的检测。
- [克隆应用](../../manual/olares/market/clone-apps.md)：了解如何在 Olares 上运行多个 Windows 实例。
