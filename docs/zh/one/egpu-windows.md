---
outline: [2, 3]
description: 在运行 Windows 11 的 Olares One 上设置并验证 NVIDIA eGPU，并处理内置显卡驱动报错。
---

# 在安装 Windows 的 Olares One 上设置 eGPU

本文介绍如何在运行 Windows 11 的 Olares One 上连接并验证 NVIDIA eGPU，以及接入 eGPU 后内置显卡报错时的恢复方法。

:::warning 暂时不要连接 eGPU
先卸载现有 NVIDIA 软件，再连接 eGPU。扩展坞需使用独立电源，并通过认证的雷电 5 线材连接。
:::

## 开始前

可以使用现有 Windows 系统，也可以先安装 Windows。不要仅为执行本文而重装 Windows。

单 eGPU 流程已使用以下配置测试：

- Olares One
- Windows 11 24H2
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti
- NVIDIA 驱动 610.74

## 准备 Windows

1. 可选：安装 Windows 11 24H2。测试使用的一键安装包为 `AGBOX4_WIN1124H2EN20260527`。
2. 打开 **设置** > **Windows 更新**，安装所有可用的 Windows 更新。等待更新完成后再继续。
3. 运行 `CleanupTool_1.0.21.0`，卸载所有 NVIDIA 应用和驱动。

## 连接 eGPU 并安装驱动

1. 将显卡安装到扩展坞，并给扩展坞通电。
2. 使用认证的雷电 5 线材，将扩展坞连接到 Olares One 的雷电 5（USB-C）接口。
3. 运行 `NVIDIA_APP_11.0.5.420_1146713`，安装 NVIDIA App。
4. 打开 NVIDIA App，下载并安装显卡驱动。测试使用了版本 610.74。
5. 驱动安装完成后，继续[确认连接状态](#确认连接状态)。

## 恢复内置显卡

接入 eGPU 后，如果内置显卡出现警告或从 NVIDIA App 消失：

1. 打开 **NVIDIA App** > **驱动程序** > **重新安装**。
2. 选择 **自定义安装**。
3. 勾选 **执行清洁安装**，完成安装。
4. 重启 Windows。

   :::warning 不要跳过重启
   在实测配置中，重启前 eGPU 会保持在 Gen1。重启 Windows，让链路重新协商为 Gen4。
   :::

## 确认连接状态

在 **设备管理器** > **显示适配器** 中，确认内置显卡和 eGPU 均已显示，且没有警告图标。

## 实测性能

在实测配置中，RTX 4060 Ti 运行 FurMark 满载时达到 165 W，链路稳定在 Gen4，且未观察到 PCIe 错误。

实际结果可能因显卡、扩展坞、供电、驱动和工作负载而异。完成设置不要求运行 FurMark。

## 已知问题

- **内置 RTX 5090M 功耗较低**
  - **现象**：内置 RTX 5090M 满载约 95 W，未达到 175 W。
  - **说明**：这是已与 NVIDIA 确认的已知问题。实测中 eGPU 链路可稳定运行在 Gen4，eGPU 使用不受影响。

- **eGPU 重启前保持在 Gen1**
  - **现象**：完成驱动恢复后，链路可能仍为 Gen1。
  - **处理**：重启 Windows，让链路重新协商为 Gen4。

## 相关资源

- [Olares One eGPU 支持概览](./egpu.md)
- [排查 eGPU 问题](./ts-egpu.md)
