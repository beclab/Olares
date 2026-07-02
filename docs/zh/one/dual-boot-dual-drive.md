---
outline: [2,3]
description: 了解如何在 Olares One 的第二块 SSD 上安装 Windows，创建双系统启动配置。
head:
  - - meta
    - name: keywords
      content: 双系统启动, Windows, NVMe SSD, BIOS, Windows 安装
---

:::warning
本页面内容经 AI 翻译生成，仅供参考。具体细节请以[英文原文](../../one/dual-boot-dual-drive.md)为准。
:::

# 在第二块 SSD 上双系统启动 Windows <Badge type="tip" text="30 min" />

如果你需要运行竞技类游戏或 Windows 独占软件，可以为 Olares One 配置第二块 NVMe SSD，实现双系统启动。

双硬盘方案在物理层面隔离了两个操作系统。这确保了 Olares OS 的稳定性和安全性，同时为你的 Windows 应用提供完整的原生性能。

## 学习目标

阅读本指南后，你将学会如何：
- 制作可启动的 Windows 安装 U 盘。
- 配置 BIOS 设置以在安装期间隔离 Olares OS。
- 在第二块 SSD 上安装 Windows。
- 配置 GRUB 以检测并在两个操作系统之间切换。

## 前提条件

**硬件**<br>
- Olares One 已物理安装第二块 NVMe M.2 SSD。
- 一个 U 盘（8 GB 或更大）用于制作 Windows 安装介质。
- 有线键盘和鼠标已连接到 Olares One。
- 显示器已连接到 Olares One。
- 网线已连接到 Olares One。
- 一个 Microsoft 账户，用于完成 Windows 设置流程。

## 步骤 1：制作可启动的 Windows U 盘

1. 从 [Microsoft 官方网站](https://www.microsoft.com/en-us/software-download/windows11) 下载 Windows 11 ISO。
2. 下载并安装 [**balenaEtcher**](https://etcher.balena.io/)。
3. 将 U 盘插入你的电脑。
4. 打开 balenaEtcher，按以下步骤操作：

   a. 点击 **Flash from file**，选择你下载的 ISO 文件。

   b. 点击 **Select target**，选择你的 U 盘。

   c. 点击 **Flash!**，将安装镜像写入 U 盘。

   ![balenaEtcher 刷写界面](/images/one/balenaEtcher.png#bordered)

5. 等待刷写和验证完成，然后安全弹出 U 盘。

## 步骤 2：进入 BIOS 并禁用 Olares OS

在安装 Windows 之前，先在 BIOS 中禁用 Olares OS 的启动盘，以确保后续双系统配置正常工作。

1. 将 Windows 启动 U 盘插入 Olares One 的 USB 端口。
2. 打开 Olares One 电源，或如果已在运行则重启。
3. 当 Olares 标志出现时，反复按 **Delete** 键进入 BIOS 设置。

   ![BIOS 设置](/images/one/bios-setup-interface.png#bordered)

4. 进入 **Boot** 选项卡，选择 Olares OS 磁盘的启动项，然后按 **Enter**。
5. 在弹出的窗口中选择 **Disabled**，然后按 **Enter**。
6. 进入 **Save and Exit** 选项卡，在 **Boot Override** 下选择你的 Windows U 盘以直接从中启动，然后按 **Enter**。
7. 按 **F10**，然后选择 **Yes** 保存更改并退出。

   系统将自动从 U 盘启动进入 Windows 安装界面。

## 步骤 3：安装 Windows

Windows 安装向导将引导你完成新系统的配置。

1. 在 **Ventoy** 界面，选择 ISO 文件（如 **Win11_25H2_English_x64_v2.iso**），然后选择 **Boot in normal mode**。随后 **Windows 11 Setup** 向导启动。
2. 按照屏幕提示完成标准 Windows 安装。在此过程中，系统会重启数次。

   :::danger 请确认所选分区
   在 **Select location to install Windows 11** 界面，确保选择正确的第二块硬盘或分区。选错分区将永久擦除你的 Olares OS。
   :::

3. 最终配置完成后，Windows 桌面出现，表示安装成功。
4. 拔出 Windows U 盘。

## 步骤 4：在 BIOS 中重新启用 Olares OS

安装完 Windows 后，在 BIOS 中重新启用 Olares OS 的启动盘，并将其设为主启动设备。你需要重新启动到 Olares OS，以便在下一步配置 GRUB 双系统启动菜单。

1. 重启 Olares One。
2. 当 Olares 标志出现时，反复按 **Delete** 键进入 BIOS 设置。
3. 进入 **Boot** 选项卡。
4. 将 **Boot Option #1** 设为装有 Olares OS 的 SSD，将 **Boot Option #2** 设为装有 Windows 的 SSD。

   ![BIOS 启动选项优先级](/images/one/bios-boot-option-priorities.png#bordered)

5. 按 **F10**，然后选择 **Yes** 保存并退出 BIOS。Olares One 将启动进入 Olares OS。

## 步骤 5：检测 Windows 并更新 GRUB

为了在启动时选择操作系统，需要配置 GRUB 引导加载器以检测 Windows。

<Tabs>
<template #Olares-未激活>

1. 使用默认凭据登录：
   * **用户名**：`olares`
   * **密码**：`olares`

   ![登录 Olares One](/images/one/one-terminal.png#bordered)
2. 运行以下命令：

   ```bash
   sudo os-prober
   ```

   如果 Windows 安装成功，你应该会看到类似如下的条目：

   ```bash
   /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi:Windows Boot Manager:Windows:efi
   ```

3. 启用 GRUB 探测其他操作系统，并重新生成启动菜单：

   a. 为 GRUB 配置创建符号链接：
      ```bash
      sudo ln -s /boot/efi/grub /boot/grub
      ```

   b. 启用 OS prober 以检测 Windows：
      ```bash
      sudo sed -i 's|GRUB_DISABLE_OS_PROBER=true|GRUB_DISABLE_OS_PROBER=false|' /etc/default/grub
      ```

   c. 重新生成 GRUB 启动菜单：
      ```bash
      sudo update-grub
      ```

   示例输出：

   ```bash
   Sourcing file '/etc/default/grub'
   Generating grub configuration file ...
   Warning: os-prober will be executed to detect other bootable partitions.
   Its output will be used to detect bootable binaries on them and create new boot entries.
   Found Windows Boot Manager on /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi
   Adding boot menu entry for UEFI Firmware Settings ...
   done
   ```
</template>
<template #Olares-已激活>

1. 从 LarePass 获取系统密码。

   :::info
   激活 Olares 后，LarePass 应用会提示你重置 SSH 密码。该密码由系统自动生成并保存到你的 Vault 中。
   :::

   a. 在 LarePass 应用中点击 **Vault**。出现提示时，输入你的本地密码以解锁。

   b. 点击左上角的 **Authenticator** 打开侧边导航，然后点击 **All vaults** 显示所有已保存的项目。

      ![切换 Vault 筛选器](/images/one/ssh-switch-filter.png#bordered)

   c. 找到带有 <span class="material-symbols-outlined">terminal</span> 图标的项目并点击，以显示密码。

      ![在 Vault 中查看已保存的 SSH 密码](/images/one/ssh-check-password-in-vault.png#bordered)

   d. 记下该密码。


2. 使用以下凭据登录：
   * **用户名**：`olares`
   * **密码**：你在步骤 1 中获取的密码。

   ![登录 Olares One](/images/one/one-terminal.png#bordered)
3. 运行以下命令：

   ```bash
   sudo os-prober
   ```

   如果 Windows 安装成功，你应该会看到类似如下的条目：

   ```bash
   /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi:Windows Boot Manager:Windows:efi
   ```

4. 启用 GRUB 探测其他操作系统，并重新生成启动菜单：

   a. 为 GRUB 配置创建符号链接：
      ```bash
      sudo ln -s /boot/efi/grub /boot/grub
      ```

   b. 启用 OS prober 以检测 Windows：
      ```bash
      sudo sed -i 's|GRUB_DISABLE_OS_PROBER=true|GRUB_DISABLE_OS_PROBER=false|' /etc/default/grub
      ```

   c. 重新生成 GRUB 启动菜单：
      ```bash
      sudo update-grub
      ```

   示例输出：

   ```bash
   Sourcing file '/etc/default/grub'
   Generating grub configuration file ...
   Warning: os-prober will be executed to detect other bootable partitions.
   Its output will be used to detect bootable binaries on them and create new boot entries.
   Found Windows Boot Manager on /dev/nvme0n1p1@/efi/Microsoft/Boot/bootmgfw.efi
   Adding boot menu entry for UEFI Firmware Settings ...
   done
   ```
</template>
</Tabs>

## 步骤 6：在两个操作系统之间切换

每次启动 Olares One 时，使用 GRUB 启动菜单选择你想要使用的操作系统。

1. 关闭 Olares One，等待几秒，然后重新开机。

   GNU GRUB 双系统启动菜单将自动出现。

   ![启动时切换系统](/images/one/one-dual-boot.png#bordered)

2. 选择要启动的操作系统。系统将在 10 秒后自动执行高亮条目。

   - **启动 Olares OS**：选择 **Olares GNU/Linux**。
   - **启动 Windows**：选择 **Windows Boot Manager**。

3. 在已登录状态下切换到另一个操作系统：

   - **从 Olares OS 切换到 Windows**：在终端运行 `sudo reboot`，出现提示时输入密码。当 GNU GRUB 菜单出现时，选择 **Windows Boot Manager**。

      :::info
      在终端输入密码时，出于安全考虑字符不会显示。确保输入正确的密码后按 **Enter**。
      :::

   - **从 Windows 切换到 Olares OS**：重启 Windows。当 GNU GRUB 菜单出现时，选择 **Olares GNU/Linux**。

## 相关资源

- [在 Windows 上安装驱动](install-nvidia-driver.md)
- [在第二块 SSD 上双系统启动 Ubuntu](dual-boot-ubuntu-dual-drive.md)
