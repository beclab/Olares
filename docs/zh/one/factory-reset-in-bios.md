---
outline: [2, 3]
description: 了解如何在 Olares One 上恢复 BIOS 默认设置，将设备恢复到初始设置状态。
head:
  - - meta
    - name: keywords
      content: Olares One, BIOS 默认设置, 恢复, BIOS 设置
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/factory-reset-in-bios.md)。
:::

# 恢复 BIOS 默认设置 <Badge type="tip" text="10 min" />

恢复 BIOS 默认设置会重置固件配置，将 Olares One 恢复到初始设置状态。如果你已连接显示器和键盘，可以直接在 BIOS 中执行此操作。

:::warning 数据丢失
这将永久删除设备上的所有账户、设置和数据。此操作无法撤销。
:::

## 前提条件
**硬件**<br>
- 有线键盘已连接到 Olares One。
- 显示器已连接到 Olares One。

## 步骤 1：在 BIOS 中加载优化默认设置

1. 打开 Olares One 电源，或如果已在运行则重启。
2. 当 Olares 标志出现时，立即反复按 **Delete** 键进入 **BIOS 设置**。
  ![BIOS 设置](/images/one/bios-setup.png#bordered)

3. 按 **F9**，然后选择 **Yes** 恢复出厂设置。
  ![加载优化默认设置](/images/one/bios-load-optimized-defaults.png#bordered)

4. 按 **F10**，然后选择 **Yes** 保存并退出。设备将自动重启。
  ![保存并退出](/images/one/bios-save-load-defaults.png#bordered)

完成后，Olares One 将重启进入初始设置阶段。

## 步骤 2：通过 LarePass 完成激活

然后你可以再次通过 LarePass 激活 Olares One。详细说明请参阅[首次启动](first-boot.md)。

