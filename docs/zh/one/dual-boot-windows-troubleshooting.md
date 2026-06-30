---
outline: [2, 3]
description: Olares One 上 Olares OS 与 Windows 双系统启动的常见问题和解决方案。
head:
  - - meta
    - name: keywords
      content: Olares One, 双系统启动, Windows, 故障排除, 显示器闪烁, HDMI, 外接显示器
---

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/dual-boot-windows-troubleshooting.md)。
:::

# 故障排除

使用本页面识别和解决 Olares One 上 Olares OS 与 Windows 双系统启动时的常见问题。

## 连接两台外接显示器启动 Windows 时显示器闪烁

### 适用条件

如果你同时遇到以下所有情况，则本问题适用：

- Olares One 已配置为双系统启动 Olares OS 和 Windows。
- Windows 启动时连接了两台外接显示器。
- 一台显示器通过 HDMI 连接，另一台通过 USB-C 扩展坞或适配器连接。
- 一台或两台显示器在 Windows 启动期间或 Windows 进入桌面后不久出现闪烁。

如果你断开并重新连接受影响的显示器后显示恢复正常，则可以确认此问题。

### 原因

Windows 启动时，会先加载一个基础显示驱动来点亮屏幕。该驱动足以启动 Windows，但在这种双系统配置中，可能无法可靠处理高分辨率 HDMI 输出和多显示器检测。

Windows 进入桌面后，NVIDIA 驱动接管显示输出。然而，如果显示器连接在启动期间已进入不稳定状态，NVIDIA 驱动可能无法自动恢复。重新连接显示器会强制 Windows 重新检测该显示设备。

这通常是一个临时的显示初始化问题，并非显示器或 GPU 损坏的迹象。

### 解决方案

如果显示器当前正在闪烁：

1. 将受影响的显示器从 Olares One 或 USB-C 扩展坞上断开。
2. 等待几秒。
3. 重新连接显示器，等待 Windows 重新检测。

要防止问题再次发生：

1. 启动进入 Windows 之前，断开 USB-C 显示器或扩展坞，仅保留 HDMI 显示器连接。
2. 启动 Windows 并等待桌面完全加载。
3. 通过 USB-C 扩展坞或适配器连接第二台显示器。
