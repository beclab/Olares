---
outline: [2, 3]
description: 在运行 Windows 11 的 Olares One 上设置 NVIDIA eGPU，并处理内置显卡驱动报错。
---

# 在安装 Windows 的 Olares One 上设置 eGPU

首次在 Windows 11 上设置 eGPU，或连接 eGPU 后内置显卡报错时，使用本指南。

:::warning 连接前先完成清理
首次设置时，运行 CleanupTool 并重启 Windows 之前，不要连接 eGPU。如果 eGPU 已连接，请关闭 Olares One，断开 eGPU，再重新启动 Windows。
:::

## 开始前

需要准备：

- 已通过 Windows 更新升级到当前可用最新版本的 Windows 11。实测配置使用 Windows 11 24H2。
- 已接通电源的雷电外置显卡设备和经过认证的雷电线材。设备可以预装 GPU，也可以通过 eGPU dock 或 eGPU enclosure 安装桌面版显卡。
- Windows 管理员权限。

如果尚未安装 Windows，请先参阅[在主硬盘上安装 Windows](./install-windows-primary-drive.md)。仅为添加 eGPU，无需重装 Windows，也无需使用指定的 Windows 镜像。

## 首次设置 eGPU

1. 打开 **设置** > **Windows 更新**，安装所有可用更新。完成后再次检查，直到 Windows 提示系统已是最新状态。
2. 下载并运行 [`CleanupTool_1.0.21.0`](https://cdn.olares.cn/common/CleanupTool_1.0.21.0.exe)，删除现有的 NVIDIA 应用、显卡驱动和相关工具。

3. CleanupTool 提示重启时，选择 **是**，等待 Windows 重新启动。
4. 关闭 Olares One。
5. 准备 eGPU 并接通电源：

   - 如果设备已经安装 GPU，请连接它的电源适配器。
   - 如果使用 eGPU dock 或 eGPU enclosure，请装入桌面版显卡，并接好显卡所需的全部供电线。

6. 使用认证的雷电线材，将 eGPU 直接连接到 Olares One 的雷电 5（USB-C）接口。
7. 启动 Olares One。
8. 安装 NVIDIA App。
9. 打开 NVIDIA App，下载并安装适用于当前显卡的最新驱动。
10. 重启 Windows。

连接其他雷电设备前，先[检查设置结果](#检查设置结果)。

## 恢复内置显卡

连接 eGPU 后，如果内置显卡在设备管理器中显示警告图标，或从 NVIDIA App 中消失，请按以下步骤操作。

操作期间保持 eGPU 连接。

1. 打开 **NVIDIA App** > **驱动程序** > **重新安装**。
2. 选择 **自定义安装**。
3. 勾选 **执行清洁安装**，完成安装。
4. 重启 Windows。

## 检查设置结果

1. 打开 **设备管理器** > **显示适配器**。
2. 检查内置显卡和所有已连接的 eGPU 是否都已显示，并且没有警告图标。
3. 打开 NVIDIA App，检查其中是否显示预期的全部 GPU。

本指南介绍推荐的单 eGPU 设置方式。[查看已知兼容状态](./egpu.md#查看已知兼容状态)中汇总了一种通过验证的三 eGPU 连接方式。本指南不包括热插拔。

## 已知问题

**内置 RTX 5090M 功耗受限**

在实测的多 eGPU 配置中，内置 RTX 5090M 的功耗约为 95 W，未达到 175 W。该限制影响内置显卡，不影响已连接的 eGPU。

## 显卡缺失或报错

请参考[排查 eGPU 问题](./ts-egpu.md)。重新安装驱动前，先记录设备管理器中的错误代码，这有助于判断问题原因。

## 相关资源

- [将 eGPU 连接到 Olares One](./egpu.md)
- [排查 eGPU 问题](./ts-egpu.md)
