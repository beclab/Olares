---
outline: [2,3]
description: 在单块 SSD 上安装 Windows 和 Olares OS，实现双系统启动配置。
head:
  - - meta
    - name: keywords
      content: 双系统启动, 单块 SSD, Windows, Ubuntu, Olares OS, 分区
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/dual-boot-single-drive.md)。
:::

# 在单块 SSD 上双系统启动 Windows <Badge type="tip" text="45 min" />

无需添加第二块硬盘，即可在 Olares One 上同时运行 Windows 和 Olares OS。

该配置先安装 Windows，然后在其旁安装 Ubuntu Linux，最后部署 Olares OS。

## 前提条件
**硬件**<br>
- Olares One 已连接电源。
- 有线键盘和鼠标已连接到 Olares One。
- （推荐）网线将 Olares One 连接到路由器。
- 一个包含 Windows 安装介质的 U 盘。
- 一个包含 Ubuntu Server 或 Desktop（24.04 LTS 或更新版本）安装介质的 U 盘。

**网络**<br>
- 稳定的互联网连接。
- 你的手机连接到同一网络。

**软件**<br>
- [已下载 LarePass 应用并创建了 Olares ID](first-boot.md#step-1-power-on-and-install-larepass)。

:::danger 需要备份
对硬盘进行分区存在数据丢失风险。如果当前硬盘上有重要数据，请在操作前将所有数据备份到外部存储。
:::

## 步骤 1：安装 Windows
1. 将 Windows 启动 U 盘插入 Olares One 的 USB 端口。
2. 打开 Olares One 电源，或如果已在运行则重启。
3. 当 Olares 标志出现时，立即反复按 **Delete** 键进入 **BIOS 设置**。
   ![BIOS 设置](/images/one/bios-setup.png#bordered)

4. 使用键盘方向键导航到 **Boot** 选项卡。
5. 将 **Boot Option #1** 设为 Windows U 盘，然后按 **Enter**。
6. 按 **F10**，然后选择 **Yes** 保存并退出 BIOS。系统将从 U 盘重启并进入 Windows 安装界面。
7. 按照屏幕提示开始 Windows 安装。
8. 安装完成且系统重启后，拔出 Windows U 盘。

系统将自动启动进入 Windows。

## 步骤 2：为 Ubuntu 创建分区

1. Windows 运行后，右键点击 **Start** 按钮，选择 **Disk Management**。
2. 右键点击你的主 `C:` 分区，选择 **Shrink Volume**。
3. 输入要为 Olares 释放的空间大小。至少需要 150 GB。
   :::tip
   避免将硬盘分成两个相等的大小。使用不同的大小可以让你在后续 Ubuntu 安装时更容易识别正确的分区。
   :::
4. 点击 **Shrink** 创建一块未分配空间。

## 步骤 3：安装 Ubuntu

Olares OS 运行在 Linux 内核之上。你需要安装 Ubuntu 作为宿主系统。

1. 插入 Ubuntu U 盘并重启 Olares One。
2. 当 Olares 标志出现时，反复按 **Delete** 键进入 **BIOS 设置**。
3. 进入 **Boot** 选项卡，将 **Boot Option #1** 设为 Ubuntu U 盘，然后按 **Enter**。
4. 按 **F10**，然后选择 **Yes** 保存并退出 BIOS。系统将从 U 盘重启并进入 Ubuntu 安装程序。
5. 按照安装向导提示操作，直到到达 **Installation type** 界面。
   :::tip
   如果该选项没有出现，选择手动安装选项，将未分配空间手动分配给 Ubuntu。
   :::
6. 安装完成且系统重启后，拔出 Ubuntu U 盘。

系统将自动启动进入 Ubuntu。

## 步骤 4：安装 Olares OS

1. 使用你的用户名和密码登录 Ubuntu。
2. 如果使用 Ubuntu Desktop，打开终端窗口；如果使用命令行，直接使用命令行。
3. 运行官方安装脚本：
   ```bash
   curl -fsSL https://olares.sh | bash
   ```

4. 安装过程结束时，系统会提示你输入域名和 Olares ID。

   ![输入域名和 Olares ID](/images/manual/get-started/enter-olares-id.png)

   例如，如果你的完整 Olares ID 是 `alice123@olares.com`：

   - **Domain name**：按 `Enter` 使用默认域名，或输入 `olares.com`。
   - **Olares ID**：输入 Olares ID 的前缀。在本例中，输入 `alice123`。

安装完成后，屏幕上会显示初始系统信息，包括 Wizard URL 和初始登录密码。在激活阶段你将需要这些信息。
![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

## 步骤 5：激活 Olares OS

1. 在 Ubuntu 的浏览器中输入 Wizard URL，或在同一网络下的另一台电脑的浏览器中打开。
   ![打开 wizard](/images/manual/get-started/open-wizard.png#bordered)
2. 输入一次性密码并点击 **Continue**。

   ![输入密码](/images/manual/get-started/wizard-enter-password.png#bordered)
3. 选择系统语言。

   ![选择语言](/images/manual/get-started/select-language.png#bordered)
4. 选择地理位置离你最近的反向代理节点。

   ![选择 FRP](/images/manual/get-started/wizard-frp.png#bordered)

5. 使用 LarePass 应用激活 Olares。

   a. 打开 LarePass 应用，点击 **Scan QR code** 扫描 Wizard 页面上的二维码，完成激活。

   ![激活 Olares](/images/manual/get-started/activate-olares.png#bordered)

   b. 重置登录密码。


设置完成后，LarePass 应用返回主屏幕，浏览器将重定向到 Olares 登录页面。

## 在两个操作系统之间切换

你可以通过 BIOS 启动优先级在 Windows 和 Olares 之间切换。
### 切换到 Olares OS
1. 重启 Olares One。
2. 反复按 **Delete** 键进入 **BIOS 设置**。
3. 进入 **Boot** 选项卡。
4. 将 **Boot Override** 设为 Ubuntu。
5. 按 **F10** 保存并退出 BIOS。

### 切换到 Windows
1. 重启 Olares One。
2. 反复按 **Delete** 键进入 **BIOS 设置**。
3. 进入 **Boot** 选项卡。
4. 将 **Boot Override** 设为 Windows。
5. 按 **F10** 保存并退出 BIOS。

## 故障排除

### 安装过程中出现错误

如果安装过程中出现错误，请先使用以下命令卸载：

```bash
olares-cli uninstall --all
```

卸载完成后，重新运行原始安装命令重试安装。

### Ubuntu 界面卡顿
首次启动进入 Ubuntu 桌面时，界面可能会感觉卡顿。这是因为 Ubuntu 的 GPU 驱动尚未安装。

继续进行 Olares OS 安装。安装脚本会自动安装必要的 GPU 驱动，卡顿问题将随之解决。


## 相关资源
- [在 Windows 上安装驱动](install-nvidia-driver.md)
