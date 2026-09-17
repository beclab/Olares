---
outline: [2, 3]
description: 在运行 Windows 11 的 Olares One 上设置并验证 NVIDIA eGPU，并处理内置显卡驱动报错。
---

# 在安装 Windows 的 Olares One 上设置 eGPU

本文介绍如何在运行 Windows 11 的 Olares One 上连接并验证 NVIDIA eGPU，以及接入 eGPU 后内置显卡报错时的恢复方法。

:::warning 暂时不要连接 eGPU
使用全新安装流程时，先卸载现有 NVIDIA 软件，再连接 eGPU。扩展坞需使用独立电源，并通过认证的雷电 5 线材连接。
:::

## 开始前

本文包含两种场景：

- 如果准备在 Olares One 上安装 Windows，请按照[从全新 Windows 安装开始设置](#从全新-windows-安装开始设置)操作。
- 在现有 Windows 系统上设置 eGPU 的流程尚未验证。不要为了执行本文而直接重装 Windows。如果接入 eGPU 后内置显卡报错，请按照[恢复内置显卡](#恢复内置显卡)操作。

本文已在以下配置中测试：

- Olares One
- Windows 11 24H2
- AOOSTAR EG02
- NVIDIA GeForce RTX 4060 Ti
- NVIDIA 驱动 610.74

测试使用了驱动 610.74。安装其他版本前，请先确认 Olares One 当前推荐的驱动版本。

<!-- TODO(tech-review):
确认支持的 Windows 设置路径，包括：
- 是否必须安装指定 Windows 镜像，或可使用现有 Windows 11 24H2
- 已安装 Windows 和 NVIDIA 驱动时的完整设置步骤
- 首次设置时是否应在 Windows 运行期间连接 eGPU
- Windows 日常热插拔是否受支持
- CleanupTool 执行后是否需要重启，以及应在连接 eGPU 前还是连接后重启
- 全新安装显卡驱动后是否必须重启 Windows
- Windows Update 是否可能在此过程中自动重新安装 NVIDIA 驱动
-->

<!-- TODO(tech-review):
确认以下文件获准公开的版本、用途、提供方、下载地址和完整性校验方式：
- AGBOX4_WIN1124H2EN20260527
- CleanupTool_1.0.21.0
- NVIDIA_APP_11.0.5.420_1146713
- NVIDIA 驱动 610.74
同时确认 Olares 是否建议内置 RTX 5090M 使用 OEM 认证驱动。
-->

## 从全新 Windows 安装开始设置

:::warning 备份数据
安装 Windows 可能会删除 Windows 分区中现有的操作系统、应用、设置和文件。继续操作前，请备份需要保留的数据。
:::

1. 安装 Windows 11 24H2。实测一键安装包为 `AGBOX4_WIN1124H2EN20260527`。
2. 运行 `CleanupTool_1.0.21.0`，卸载所有 NVIDIA 应用和驱动。

   <!-- TODO(tech-review): 说明为什么实测流程需要先删除 Windows 镜像中的 NVIDIA 软件，再重新安装。 -->

3. 安装显卡并给扩展坞通电。
4. 使用认证的雷电 5 线材，将扩展坞连接到 Olares One 的雷电 5（USB-C）接口。

   仅在该实测初始设置步骤中，于 Windows 运行期间连接 eGPU。日常热插拔尚未验证。

5. 运行 `NVIDIA_APP_11.0.5.420_1146713` 安装 NVIDIA App。
6. 在 NVIDIA App 中安装显卡驱动。测试使用了版本 610.74。
7. 驱动安装完成后，继续[确认连接状态](#确认连接状态)。

## 恢复内置显卡

接入 eGPU 后，如果内置显卡出现警告或从 NVIDIA App 消失：

1. 打开 **NVIDIA App** > **驱动程序** > **重新安装**。
2. 选择 **自定义安装**。
3. 勾选 **执行清洁安装**，完成安装。

   :::warning 不要跳过重启
   在实测配置中，重启前 eGPU 会保持在 Gen1。重启 Windows，让链路重新协商为 Gen4。
   :::

4. 重启 Windows。

<!-- TODO(tech-review): 确认上述 NVIDIA App 界面名称与获准公开的版本一致。 -->

## 确认连接状态

- 在 **设备管理器** > **显示适配器** 中，确认内置显卡和外置显卡均已显示，且没有警告图标。
- 在 GPUMon 中，确认 eGPU 链路运行在 Gen4。

<!-- TODO(tech-review): 提供获准使用的 GPUMon 来源或下载地址，并说明用户需要检查的具体字段。 -->

## 实测性能

在实测配置中，RTX 4060 Ti 运行 FurMark 满载时达到 165 W。测试期间，GPUMon 显示链路为 Gen4，且没有 PCIe 错误。

实际结果可能因显卡、扩展坞、供电、驱动和工作负载而异。完成设置不要求运行 FurMark。

## 已知问题

| 现象 | 状态 |
|---|---|
| 内置 RTX 5090M 满载约 95 W，未达到 175 W | 这是已与 NVIDIA 确认的已知问题，目前等待上游修复。实测配置中，外置 RTX 4060 Ti 可稳定运行在 Gen4，使用未受影响。 |
| 接入 eGPU 后内置显卡报错 | 执行清洁安装并重启。 |
| 连接两张或更多外置 GPU | 多 eGPU 配置尚未验证。 |

<!-- TODO(tech-review): 确认两张或更多外置 GPU 的支持状态。内置显卡加一张外置显卡不属于多 eGPU 配置。请同时明确此前测试仅确认设备可识别，还是也完成了负载和稳定性验证。 -->
<!-- TODO(tech-review): 确认此处的 175 W 是内置 RTX 5090M 的预期满载功耗目标。 -->

## 相关文档

- [Olares One eGPU 支持概览](./egpu.md)
- [排查 eGPU 问题](./ts-egpu.md)
