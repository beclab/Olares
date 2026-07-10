---
outline: [2, 3]
description: 了解如何在 Olares One 上管理 BIOS 和 EC 固件，包括检查固件版本、下载固件更新包、执行更新、解锁高级模式，以及排查黑屏等问题。
head:
  - - meta
    - name: keywords
      content: Olares One, 固件更新, 嵌入式控制器 (EC), BIOS
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/update-firmware.md)为准。
:::

# 管理 BIOS 和 EC

本文档介绍如何在 Olares One 设备上管理 BIOS 和嵌入式控制器（EC）固件，包括如何检查当前版本、下载和执行固件更新、解锁高级设置，以及排查由显示配置引起的黑屏问题。

## 检查固件版本

在进行任何更新之前，先检查当前系统固件版本，以确定是否需要更新。

1. 开启 Olares One 电源，或如果设备正在运行则重启。
2. 当 Olares 启动标志出现时，立即按住 **F7** 键进入启动菜单。
3. 选择 **Enter Setup** 进入 BIOS。
4. 在 **Main** 标签页中，检查以下内容：

    - **System BIOS Version**：当前 BIOS 版本。
    - **EC FW Version**：当前嵌入式控制器（EC）版本。

    ![在 BIOS 中检查当前固件版本](/images/one/check-firmware-versions-in-bios.png#bordered)

5. 按 **ESC**，然后选择 **Yes** 退出 BIOS 而不保存。

## 下载固件更新

查看以下各版本的更新日志，了解每个更新包含的功能或修复。

如果你的当前版本低于以下列出的版本，请下载相应的更新包以继续。

### BIOS 版本

<!--| [1.05 (下载)](https://cdn.olares.com/common/OlaresOne_BIOS_1.05.zip) | 2026-07-06 | <ul><li>修复雷电口功能异常的问题。此前为改善功耗异常，禁用了所有设备的 D3 (D3Cold) 低功耗状态；现在仅禁用 GPU 的 D3 状态。</li></ul> |-->

| 版本 | 发布日期 | 更新日志 |
|:--------|:-------------|:----------|
| [1.04 (下载)](https://cdn.olares.com/common/OlaresOne_BIOS_1.04.zip)| 2026-05-09 | <ul><li>在将 **Primary Display** 设置更改为 **HG** 时添加警告提示。</li><li>通过锁定 GPU PCIe 速度为 Gen4 修复 GPU 意外断开连接的问题。</li><li>通过禁用产品空闲时将 GPU 置于睡眠模式的功能，修复长时间使用后性能下降且功耗异常受限的问题。</li></ul> |
| 1.03  | 2026-03-19 | <ul><li>修复 Ubuntu 系统启动时出现的 ACPI 错误。</li><li>将 Intel CPU 微码更新至版本 0x121。</li></ul> |
| 1.01  | 2025-12-04 | <ul><li>通过禁用 SSD1 和 SSD2 的 ASPM 和 L-state 电源管理，修复 SSD 意外断开连接的问题。</li></ul> |
| 1.00 | 2025-11-28 | <ul><li>更新版本命名规范。</li></ul> |
| C400 | 2025-11-05 | <ul><li>默认隐藏高级 BIOS 选项。</li><li>移除 MCU 版本显示。</li><li>通过启用 SAGV 修复内存测试报错的问题。</li></ul> |

### EC 版本

| 版本 | 发布日期 | 更新日志 |
|:--------|:-------------|:----------|
| [1.03 (下载)](https://cdn.olares.com/common/OlaresOne_EC_1.03.zip) | 2026-05-29 | <ul><li>增加连接适配器后自动开机的端口，从而支持设置中的<strong>自动开机</strong>功能。</li><li>需要 Olares OS 1.12.6 或更高版本。</li></ul> |
| 1.02 | 2026-01-19 | <ul><li>修复键盘无法唤醒系统从睡眠模式恢复的问题。</li></ul> |
| 1.01 | 2026-01-13 | <ul><li>添加对网络唤醒（WOL）的支持。</li><li>在睡眠模式下禁用白色呼吸 LED 指示灯。</li></ul> |
| 1.00 | 2025-12-01 | <ul><li>在睡眠模式下启用白色呼吸 LED 指示灯。</li></ul> |
| C3.00 | 2025-11-25 | <ul><li>修复从睡眠模式唤醒后风扇不转的问题。</li></ul> |

## 更新固件

使用以下说明在 Olares One 上手动刷写 BIOS 或 EC 固件。

### 前提条件

- 格式化为 `FAT32` 的 U 盘。
- 显示器和 USB 键盘连接到 Olares One。
- 已下载到电脑的 EC 或 BIOS 更新包。

### 更新 BIOS

:::warning 重要提示
在 BIOS 更新过程中请勿断开电源或关闭设备。这样做可能会永久损坏系统。
:::

1. 解压下载的 BIOS 更新包。
2. 将生成的文件夹（例如 `AGBOX4_BIOS_101` 和 `EFI`）复制到 U 盘根目录。
3. 将 U 盘连接到 Olares One。
4. 开启设备电源，或如果设备正在运行则重启。
5. 当 Olares 标志出现时，立即按住 **F7** 键进入启动菜单。
6. 从列表中选择你的 U 盘，然后按 **Enter**。

    ![选择 USB 启动设备](/images/one/select-usb-boot1.png#bordered)

7. 当 EFI 启动倒计时屏幕出现时（`Press ESC in 5 seconds to skip startup.nsh`），立即按 **Enter** 进入命令行 shell。

    ![UEFI shell 启动](/images/one/uefi-shell-startup.png#bordered)

8. 依次运行以下命令，导航到 AFU 目录并启动刷写脚本：

    ```bash
    cd AGBOX4_BIOS_<version> # 例如，cd AGBOX4_BIOS_101
    cd AFU
    FlashAFU.nsh
    ```
    ![运行 BIOS 刷写脚本](/images/one/bios-flash-commands-101.png#bordered)

9. 等待脚本执行完成。

    系统将自动重启并显示蓝色的 **Flash Update** 进度屏幕。

    ![BIOS 刷写进度屏幕](/images/one/bios-update-progress.png#bordered)

10. 刷写更新达到 100% 后，**ME FW Update** 将自动开始。等待此过程完成。

    ![BIOS 刷写进度屏幕 - ME FW Update](/images/one/bios-update-progress-me.png#bordered)

11. **ME FW Update** 完成后，系统将自动重启两次。等待直到正常的 `olares login` 提示出现。

    :::info
    在重启过程中，系统将执行全面的硬件自检。此过程大约需要 2 到 3 分钟。在此期间屏幕可能保持黑屏。
    :::

12. 验证 BIOS 版本。

    a. 手动重启 Olares One。

    b. 当 Olares 标志出现时，立即按住 **F7** 进入启动菜单。

    c. 选择 **Enter Setup** 进入 BIOS。

    ![进入 BIOS 设置](/images/one/enter-setup.png#bordered)

    d. 在 **Main** 标签页中，验证 **System BIOS Version** 显示的是你的目标版本（例如 `1.01`），以确认更新成功。

    ![验证 BIOS 版本](/images/one/enter-setup-bios1.png#bordered)

### 更新 EC 固件

1. 解压下载的 EC 更新包。
2. 将生成的文件夹（例如 `AGBOX4_EC_01_02` 和 `EFI`）复制到 U 盘根目录。
3. 将 U 盘连接到 Olares One。
4. 开启设备电源，或如果设备正在运行则重启。
5. 当 Olares 标志出现时，立即按住 **F7** 键进入启动菜单。
6. 从列表中选择你的 U 盘，然后按 **Enter**。

    ![选择 USB 启动设备](/images/one/select-usb-boot.png#bordered)

7. 当 EFI 启动倒计时屏幕出现时（`Press ESC in 5 seconds to skip startup.nsh`），立即按 **Enter** 进入命令行 shell。

    ![UEFI shell 启动](/images/one/uefi-shell-startup-ec.png#bordered)

8. 输入以下命令，然后按 **Enter** 导航到 EC 目录：

    ```bash
    cd AGBOX4_EC_<version> # 例如，cd AGBOX4_EC_01_02
    ```

    ![导航到 EC 目录](/images/one/ec-cd-command.png#bordered)

9. 输入以下命令，然后按 **Enter** 执行更新工具：

    ```bash
    ECFlashTool.efi AGBOX4_EC_<version>.bin # 例如，ECFlashTool.efi AGBOX4_EC_01_02.bin
    ```

    ![运行 EC 刷写工具](/images/one/ec-flash-command.png#bordered)

    系统将显示擦除和编程闪存存储器的进度。等待更新过程完成。完成后，设备将自动重启。

    ![EC 更新进度](/images/one/ec-update-progress.png#bordered)

10. 当 Olares 标志出现时，立即按住 **F7** 进入启动菜单。
11. 选择 **Enter Setup** 进入 BIOS。

    ![进入 BIOS 设置](/images/one/enter-setup-bios.png#bordered)

12. 在 **Main** 标签页中，验证 **EC FW Version** 显示的是你的目标版本（例如 `1.02`），以确认更新成功。

    ![在 BIOS 中验证 EC 版本](/images/one/verify-ec-version.png#bordered)

## 解锁高级设置

某些高级 BIOS 选项默认隐藏，以防止意外更改配置影响系统稳定性。

如果你需要进行深度硬件配置，可以解锁隐藏的高级设置：
1. 进入 BIOS，然后前往 **Advanced** 标签页。

    ![BIOS 中的默认高级设置](/images/one/bios-advanced-default.png#bordered)

2. 在键盘上按 Ctrl + 右方向键。隐藏的配置选项（例如 **RC ACPI Settings** 和 **PCIE Configuration**）将显示在屏幕上。

    ![BIOS 中的完整高级设置](/images/one/bios-advanced-full.png#bordered)

    :::warning 请勿更改 Primary Display
    确保 **Primary Display** 设置保持为 **Discrete GPU**。

    不要将此设置更改为 **HG**（Hybrid Graphic）。选择 **HG** 会将视频输出路由到非活动接口，导致启动时完全黑屏。

    如果你不小心更改了此设置并丢失了显示输出，请参阅[故障排查](#无法进入-bios黑屏)部分执行盲操作重置。
    :::

## 故障排查

### 无法进入 BIOS（黑屏）

当你开启 Olares One 电源并尝试进入 BIOS 时，显示器屏幕保持黑屏，无法显示 BIOS 设置界面。

此问题通常发生在 BIOS 中的 `Primary Display` 设置从 `Discrete GPU` 被更改为 `HG`（Hybrid Graphics）时。当 `Primary Display` 设置为 `HG` 时，早期启动阶段的视频输出可能会被发送到显示器无法检测到的不同显示接口（例如集成显卡）。系统实际上可能已成功进入 BIOS，但你的显示器屏幕保持黑屏，导致无法看到 BIOS 设置界面。

#### 解决方案

按照以下步骤盲操作重置 BIOS 为出厂默认值。这将恢复 `Primary Display` 设置为默认值。

:::warning
- 以下步骤需要物理移除存储设备。请确保在操作前关闭设备电源并断开电源线。
- 直到步骤 3 屏幕才会恢复显示。使用键盘 Caps Lock 指示灯确认是否已进入 BIOS。
:::

##### 步骤 1：准备设备

1. 关闭 Olares One 电源。
2. 断开电源线。
3. 断开所有外部存储驱动器，例如 U 盘或外置硬盘。
4. 打开设备外壳，临时移除内部 NVMe SSD。
5. 将键盘和显示器连接到设备。
6. 重新连接电源线。

##### 步骤 2：执行盲操作重置

1. 开启 Olares One 电源，然后立即按住 **Delete** 键约 20 秒。
2. 通过按 **Caps Lock** 键确认系统正在接收键盘输入：如果键盘上的 Caps Lock 指示灯亮起和熄灭，说明已成功进入 BIOS。
3. 按 **F9** 键，然后按 **Enter**。此快捷键加载出厂默认 BIOS 设置。
4. 按 **F10** 键，然后按 **Enter**。此快捷键保存配置并提示设备重启。

##### 步骤 3：验证修复

1. 等待设备重启。Olares One 标志应出现，表示正常显示输出已恢复。
2. 关闭设备电源并断开电源线。
3. 重新安装内部 NVMe SSD 并重新连接任何外部驱动器。
4. 重新连接电源线并开启设备电源，恢复正常 BIOS 操作。
