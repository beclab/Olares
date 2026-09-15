---
description: 在设置中查看 Olares One 硬件详情，并管理工作模式、CPU 频率限制和自动开机。
head:
  - - meta
    - name: keywords
      content: Olares One, 硬件设置, 工作模式, CPU 频率, 自动开机, 硬件
---

# 管理硬件设置

在**设置**的**我的 Olares** > **硬件**页面，你可以查看硬件详情，并管理 Olares One 的性能选项。

![硬件](/images/zh/manual/olares/my-hardware-1.12.6.png#bordered)

## 查看硬件详情

查看**型号**、**设备状态**、**设备标识符**、**CPU** 和 **GPU** 等信息。

## 切换工作模式

Olares One 支持两种性能档位：

- **静音模式**：限制 CPU/GPU 功耗，满足日常负载并保持安静。
- **性能模式**：释放 CPU/GPU 最大性能，适合 AI 推理、游戏等高负载场景。

## 限制 CPU 频率

开启后，将 CPU 频率上限从 5.4 GHz 降至 5.0 GHz。关闭后恢复原有最高频率。

## 设置自动开机

开启后，设备在接通电源或停电后恢复供电时会自动开机。

:::info
需要 Olares OS 1.12.6 或更高版本，以及 EC 固件 1.03 或更高版本。如果任一条件未满足，开关会显示但无法操作。

如需查看当前 EC 固件版本或升级 EC 固件，可参考[管理 BIOS 和 EC](update-firmware.md)。
:::
