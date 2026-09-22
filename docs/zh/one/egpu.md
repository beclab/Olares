---
outline: [2, 3]
description: 了解 eGPU、查看已知硬件组合的兼容状态，并将 NVIDIA eGPU 连接到 Olares One。
head:
  - - meta
    - name: keywords
      content: Olares One, eGPU, 外置显卡, 雷电 5, NVIDIA
---

# 将 eGPU 连接到 Olares One

需要更多 GPU 算力时，可以通过雷电接口给 Olares One 连接 eGPU。本文汇总目前已知的硬件兼容情况，并提供 Olares OS 和 Windows 11 的设置入口。

:::danger 在 Olares OS 上连接前先关机
不要在 Olares OS 运行时连接或断开 eGPU。先关闭 Olares One，再给 eGPU 通电并连接线材，最后启动 Olares One。
:::

## 查看已知兼容状态

下表列出的是目前已有明确兼容结果的单 eGPU 硬件组合，不是产品推荐或完整的兼容列表。未列出的组合可能可以使用，但尚未验证。

| 外置显卡硬件 | Olares OS | Windows 11 |
|---|---|---|
| ROG XG Mobile (2025)（NVIDIA GeForce RTX 5070 Ti Laptop GPU） | **连接后可用。** 冷启动后无需额外设置。 | **尚未验证。** |
| AOOSTAR EG02 eGPU dock + RTX 4060 Ti | **设置后可用。** 安装 [Gen1 临时方案](./egpu-olares-os.md#必要时安装临时方案)。 | **设置后可用。** 连接 eGPU 后重新安装 NVIDIA 驱动。 |
| Razer Core X V2 eGPU enclosure + RTX 4090 | **验证中。** 未安装 Gen1 临时方案时开机不稳定。安装后的表现尚未验证。 | **验证中。** |
| eGPU dock 或 eGPU enclosure + 桌面版 RTX 5090 | **暂不支持。** NVIDIA 驱动无法初始化，目前没有可用方案。 | **尚未验证。** |

兼容性还会受到 eGPU dock 或 eGPU enclosure、供电、线材、操作系统和 NVIDIA 驱动的影响。

:::warning 先使用一台 eGPU
Olares OS 不支持多 eGPU。Windows 最多可以识别两台 eGPU，但多 eGPU 仍在验证中。串接三台 eGPU 的组合在两个系统上均未能正常工作。
:::

## 开始前的准备

- 如果使用桌面版显卡，请选择 Turing 或更新架构的 NVIDIA 显卡，并确保 eGPU dock 或 eGPU enclosure 及其电源满足显卡要求。
- 使用认证的雷电 5 线材，建议优先使用外置显卡设备附带的线材。
- 首次设置时，只连接一台 eGPU，并将它直接连接到 Olares One。断开其他高带宽雷电设备，不要经过另一台 Dock 转接。

:::info 旧款雷电设备
旧款雷电外置显卡设备可能也能使用，但建议选择雷电 5 设备。
:::

## 设置 eGPU

### Olares OS

先冷启动连接。关闭 Olares One，给 eGPU 通电并完成连接，再启动 Olares One。如果系统能识别 eGPU，且连接稳定，无需其他设置。

按照[在 Olares OS 上设置 eGPU](./egpu-olares-os.md)中的步骤操作。如果出现开机卡住、无法识别或掉卡，指南中也提供了 Gen1 临时方案。

Olares OS 不支持热插拔。

### Windows 11

可以继续使用现有的 Windows 11。首次设置时，需要在连接 eGPU 后重新安装 NVIDIA 驱动。

按照[在 Windows 上设置 eGPU](./egpu-windows.md)中的步骤操作。该指南只介绍单 eGPU 设置，不包括热插拔。

## 获取帮助

如果 Olares One 无法启动、系统看不到 eGPU，或显卡驱动报错，请参考[排查 eGPU 问题](./ts-egpu.md)。
