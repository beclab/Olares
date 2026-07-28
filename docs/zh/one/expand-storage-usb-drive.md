---
outline: [2, 3]
description: 了解如何使用 USB 设备扩展 Olares One 的存储空间。
head:
   - - meta
     - name: keywords
       content: Olares One, USB 存储, 扩展存储, 文件管理
---

:::warning
本页面为 AI 翻译版本，内容仅供快速参考。关键信息建议以[英文原文](../../one/expand-storage-usb-drive.md)为准。
:::

# 通过 USB 设备扩展存储 <Badge type="tip" text="5 min" />

Olares One 会自动检测并挂载插入的 USB 存储设备，让你可以即时扩展存储容量，用于存储媒体、备份或文件传输。

## 前提条件
**硬件**
- 兼容的 U 盘。
  :::warning 兼容性
  Olares One 已测试兼容 Samsung 和 SanDisk U 盘。其他品牌可能无法被系统识别。
  :::
## 连接并访问存储

1. 将 U 盘插入 Olares One 的可用 USB 端口。系统将自动挂载。
2. 打开 Files 应用，在侧边栏选择 **External** 以访问你的文件。

## 安全移除设备

虽然系统在你物理拔出设备时会自动卸载，但在数据写入过程中移除硬盘可能导致文件损坏。

建议先手动弹出设备：
1. 打开 Files 应用。
2. 右键点击侧边栏中的 U 盘，选择 **Unmount**。
3. 硬盘从列表中消失后，即可安全拔出。

## 相关资源
- [在 Olares 中管理文件](../manual/olares/files/index.md)
- [使用外接 SSD 扩展存储](expand-storage-external-ssd)
- [使用内置 SSD 扩展存储](expand-storage-internal-ssd.md)
