---
outline: deep
description: 在 Olares 上运行 macOS 虚拟机。了解如何从 Market 安装 macOS、配置初始设置，以及通过基于浏览器的 VNC 或 VNC 客户端连接。
head:
  - - meta
    - name: keywords
      content: Olares, macOS, virtual machine, VNC, macOS VM
app_version: "1.0.8"
doc_version: "1.0"
doc_updated: "2026-04-08"
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../use-cases/macos.md)为准。
:::

# 在你的 Olares 设备上运行 macOS 虚拟机

Olares 允许你直接在设备上将 macOS 作为虚拟机 (VM) 运行。这使你可以通过 Web 浏览器或 VNC 客户端从任何计算机访问 Apple 特定的应用和工作流。

:::info
此虚拟机仅使用 CPU 运行。GPU 加速不可用，因此推荐用于不需要高性能图形的应用。
:::

## 学习目标

在本教程结束时，你将学习如何：
- 在你的 Olares 设备上安装和设置 macOS 虚拟机环境。
- 直接在 Web 浏览器中或通过 VNC 应用访问 macOS 虚拟机。

## 安装和配置 macOS 虚拟机

macOS 在 Olares Market 中作为应用提供。

### 安装 macOS 虚拟机应用

1. 打开 Market 并搜索 "macOS"。

   ![Market 中的 macOS 应用](/images/manual/use-cases/macos.png#bordered)

2. 点击 **获取**，然后点击 **安装**。
3. 出现提示时，设置环境变量：
   - **VERSION**：从下拉列表中选择你首选的 macOS 版本。
   - **DISK_SIZE**：为 macOS 分配磁盘空间。

   ![设置环境变量](/images/manual/use-cases/macos-set-env-var.png#bordered){width=65%}

4. 点击 **确认** 并等待安装完成。

### 格式化虚拟磁盘

1. 从 Launchpad 打开 MacOS 应用。

   :::tip 首次启动
   首次启动时，Olares 自动下载并安装系统镜像。根据你的网络速度，这可能需要几分钟。
   :::

2. 当 **恢复** 屏幕出现时，从主菜单中选择 **磁盘工具**，然后点击 **继续**。

   ![macOS 恢复菜单](/images/manual/use-cases/macos-recovery-menu.png#bordered){width=50%}

3. 在左侧边栏中，选择容量最大的 **Apple Inc. VirtIO Block Media**，然后点击工具栏上的 **抹掉**。

   ![选择磁盘工具](/images/manual/use-cases/macos-select-disk-utility.png#bordered)

4. 配置格式：

   - **名称**：输入一个名称。例如，`Macintosh HD`。
   - **格式**：选择 **APFS**。

   ![配置格式](/images/manual/use-cases/macos-configure-format.png#bordered){width=50%}

5. 点击 **抹掉**，等待过程完成，然后点击 **完成**。

   ![磁盘已格式化](/images/manual/use-cases/macos-disk-formatted.png#bordered){width=50%}

6. 关闭 **磁盘工具** 窗口以返回主菜单。

### 在虚拟磁盘上安装 macOS

1. 从主菜单中，选择 **重新安装 macOS**，然后点击 **继续**。

   ![重新安装 macOS](/images/manual/use-cases/macos-reinstall.png#bordered){width=50%}

2. 接受许可协议。
3. 选择你刚刚格式化的磁盘，然后点击 **继续**。
4. 等待安装完成，通常需要 20-40 分钟。

   ![macOS 安装进度](/images/manual/use-cases/macos-installing.png#bordered){width=60%}

### 完成初始设置

系统安装完成后：

1. 按照地区、语言和辅助功能设置的提示操作。
2. 当提示设置迁移助手时，点击左下角的 **以后**。
3. 当提示使用 Apple ID 登录时，点击左下角的 **稍后设置**。
4. 为 macOS 账户设置用户名和密码。对于剩余的设置步骤，你可以跳过或接受默认值。

   ![macOS 桌面](/images/manual/use-cases/macos-desktop.png#bordered)

## 访问 macOS 虚拟机

### 从浏览器访问

从 Launchpad 打开 macOS 应用以直接在浏览器中访问虚拟机。

用于初始设置、快速访问或故障排除。

### 使用 VNC Viewer 访问

专用的 VNC 客户端提供更好的稳定性、更低的延迟和改进的键盘映射。

#### 步骤 1：获取连接详情

:::warning 多个 macOS 实例
每个 macOS 实例使用唯一的端口。如果你克隆了 MacOS 应用，请确保检查你要访问的特定实例的 **ACLs** 部分。
:::

1. 打开 **设置**，然后前往 **应用** > **MacOS**。
2. 在 **入口** 下，点击 **MacOS**，然后记下端点地址。

   **示例**：`https://43b9d8ea.alexmiles.olares.com`。

   ![定位端点](/images/manual/use-cases/macos-endpoint.png#bordered){width=80%}

3. 返回上一页，点击 **ACLs**，然后记下端口号。

   **示例**：`49238`。

   ![定位端口号](/images/manual/use-cases/macos-port-number.png#bordered){width=80%}

4. 通过组合 **端点**（不带 `https://` 前缀）和 **端口号**，用冒号分隔，构建 VNC 连接地址。

   - **格式**：`[端点-不含-https]:[端口]`
   - **示例**：`43b9d8ea.alexmiles.olares.com:49238`

#### 步骤 2：启用 VPN 连接

你必须在 Olares 安全网络上才能通过 VNC Viewer 连接。

1. 打开 LarePass 桌面客户端。
2. 点击头像，然后启用 **VPN 连接**。

   ![在 LarePass 桌面版上启用 VPN](/images/manual/use-cases/alex-larepass-vpn-desktop.png#bordered){width=90%}

3. 确保状态显示 **P2P** 或 **内网** 后再继续。

#### 步骤 3：通过 VNC Viewer 连接

<Tabs>
<template #macOS>

1. （可选）如果你尚未安装 [Homebrew](https://brew.sh)。
2. 在你的计算机上打开终端，然后运行以下命令安装 VNC Viewer 应用：

   ```bash
   brew install --cask vnc-viewer
   ```

   消息 `vnc-viewer was successfully installed!` 表示安装成功。

3. 从你的计算机打开 VNC Viewer 并使用你的 RealVNC 账户登录。如果你没有账户，请创建一个然后登录。
4. 点击 **文件** > **新建连接**。
5. 输入从[步骤 1](#步骤-1获取连接详情)获取的地址。在本例中，它是 `43b9d8ea.alexmiles.olares.com:49238`。

   ![macOS 上 VNC Viewer 中的新连接](/images/manual/use-cases/vnc-new-connection.png#bordered){width=60%}

6. 点击 **确定**。连接保存在 VNC Viewer 中。

   ![VNC Viewer 中已连接的虚拟机](/images/manual/use-cases/vnc-vm-connected.png#bordered)

7. 双击保存的连接以连接。
8. 如果出现 "未加密连接" 警告，点击 **继续**。
9. 出现提示时，输入你之前创建的用户名和密码。

   你现在已通过 VNC Viewer 连接到你的 macOS 虚拟机。

</template>
<template #Windows>

1. 下载并安装 [RealVNC Viewer](https://www.realvnc.com/en/connect/download/viewer/)。
2. 从你的计算机打开 VNC Viewer 并使用你的 RealVNC 账户登录。如果你没有账户，请创建一个然后登录。
3. 点击 **文件** > **新建连接**。
4. 输入从[步骤 1](#步骤-1获取连接详情)获取的地址。在本例中，它是 `43b9d8ea.alexmiles.olares.com:49238`。

   ![Windows 上 VNC Viewer 中的新连接](/images/manual/use-cases/vnc-viewer-windows.png#bordered){width=60%}
5. 点击 **确定**。连接保存在 VNC Viewer 中。
6. 双击保存的连接以连接。
7. 如果出现 "未加密连接" 警告，点击 **继续**。
8. 出现提示时，输入你之前创建的用户名和密码。

   你现在已通过 VNC Viewer 连接到你的 macOS 虚拟机。

</template>
</Tabs>

#### 步骤 4：断开与 macOS 虚拟机的连接

要断开与 macOS 虚拟机的连接，请关闭 VNC Viewer 窗口。

macOS 虚拟机继续在你的 Olares 设备上运行，并保持就绪状态供你重新连接。

## 常见问题

### 我可以在此虚拟机中使用我的 Apple ID 吗？

虽然你可以在设置期间使用 Apple ID 登录，但某些 Apple 服务在虚拟化环境中可能无法正常工作。为获得最佳效果，请使用本地账户或跳过 Apple ID 设置。

### 支持哪些 macOS 版本？

目前支持的版本：
- macOS 14 Sonoma
- macOS 13 Ventura
- macOS 12 Monterey
- macOS 11 Big Sur
- macOS 10 Catalina

### 为什么 VNC Viewer 显示 "The connection closed unexpectedly"？

此问题通常发生在 LarePass VPN 被禁用时。

要解决它，请打开你的 LarePass 桌面客户端并确保 VPN 连接状态为 **P2P** 或 **内网**。然后重试连接。

## 了解更多

- [dockur/macos GitHub 仓库](https://github.com/dockur/macos)
- [在你的 Olares 设备上运行 Windows 虚拟机](./windows.md)
