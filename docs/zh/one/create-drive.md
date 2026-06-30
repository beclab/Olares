---
outline: [2, 3]
description: 使用可启动 USB 驱动器在 Olares One 上重新安装 Olares OS，将设备恢复到干净的初始状态。
head:
  - - meta
    - name: keywords
      content: Olares One, reinstall, Olares OS, bootable USB, installation USB
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/create-drive.md)。
:::

# 使用可启动 USB 重新安装 Olares OS <Badge type="tip" text="15 min"/>

重新安装 Olares OS 会将你的 Olares One 恢复到干净的初始状态。

:::warning 数据丢失
这将永久删除设备上的所有账户、设置和数据。此操作无法撤销。
:::

## 前提条件

- 一台连接到 Olares One 的显示器和键盘。
- Olares One 附带的可启动 USB 驱动器。
 
   :::info 已安装的 OS 版本
   附带 USB 驱动器上的 OS 镜像可能是较早版本，具体取决于设备的出厂时间。你仍可以用它重新安装系统，如有需要，激活后再更新 Olares OS。

   如果你需要最新的 OS 镜像，可以从 [Olares One ISO](https://cdn.olares.com/one/olares-latest-amd64.iso) 创建新的可启动 USB 驱动器。详细步骤请参阅[创建 Olares One 可启动 USB 驱动器](create-bootable-usb.md)。

## 步骤 1：从 USB 驱动器启动

1. 将可启动 USB 驱动器插入 Olares One。
2. 打开 Olares One 电源，或如果已在运行则重新启动。
3. 当 Olares 标志出现时，立即反复按 **Delete** 键进入 **BIOS setup**。
   ![BIOS 设置](/images/one/bios-setup.png#bordered)

4. 导航到 **Boot** 标签页，将 **Boot Option #1** 设置为 USB 驱动器，然后按 **Enter**。
   ![设置启动选项](/images/one/bios-set-boot-option.png#bordered)

5. 按 **F10**，然后选择 **Yes** 保存并退出。
   ![保存并退出](/images/one/bios-save-usb-boot.png#bordered)


Olares One 将重新启动并进入 Olares 安装程序界面。

## 步骤 2：将 Olares 安装到磁盘

1. 在安装程序界面中，选择 **Install Olares to Hard Disk** 并按 **Enter**。
   ![Olares 安装程序](/images/one/olares-installer.png#bordered)

2. 当提示选择安装目标时，安装程序会显示可用磁盘列表。输入 `/dev/` 后跟列表中的磁盘名称（例如 `nvme0n1`），然后按 **Enter**。
   ![选择磁盘](/images/one/olares-installer-select-disk.png#bordered)

   例如，要安装到 `nvme0n1`，输入：
   ```bash
   /dev/nvme0n1
   ```

3. 当看到 NVIDIA GPU 驱动程序提示时，按 **Enter** 接受默认选项。
   ![安装 NVIDIA 驱动](/images/one/olares-installer-install-nvidia-drivers.png#bordered)

4. 当看到以下消息时，重新安装已完成：
   ```bash
   Installation completed successfully!
   ```

5. 移除 USB 驱动器，然后按 **Ctrl + Alt + Delete** 重新启动。

## 步骤 3：验证安装

重新启动后，系统以干净的出厂状态启动，并显示基于文本的 Ubuntu 登录提示。

1. 使用默认凭据登录：
   - **Username**：`olares`
   - **Password**：`olares`
   ![登录](/images/one/olares-login.png#bordered)

2. （可选）运行以下命令验证安装：
   ```bash
   sudo olares-check
   ```
   示例输出：
   ![Olares 检查](/images/one/olares-check.png#bordered)


## 步骤 4：通过 LarePass 完成激活

然后你可以再次通过 LarePass 激活 Olares One。详细说明请参阅[首次启动](first-boot.md)。

## 步骤 5：检查系统更新（可选）

附带的 USB 驱动器可能安装的是较早的 OS 版本。激活后，你可以在 LarePass 中更新 Olares OS。详细步骤请参阅[更新 OS](update.md)。
