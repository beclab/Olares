---
outline: [2, 3]
description: 了解 eGPU、查看已知硬件组合的兼容状态，并将 NVIDIA eGPU 连接到 Olares One。
head:
  - - meta
    - name: keywords
      content: Olares One, eGPU, 外置显卡, 雷电 5, NVIDIA
---

# 将 eGPU 连接到 Olares One

需要更多 GPU 算力时，可以通过雷电接口给 Olares One 连接 eGPU。

:::danger 在 Olares OS 上连接前先关机
不要在 Olares OS 运行时连接或断开 eGPU。先关闭 Olares One，再给 eGPU 通电并连接线材，最后启动 Olares One。
:::

## 查看已知兼容状态

下表列出目前已有的兼容结果，不代表产品推荐，也无法涵盖所有组合。实际表现还取决于 dock 或 enclosure、供电、线材、操作系统和 NVIDIA 驱动。未列出的组合可能可以使用，但尚未验证。

| 外置显卡硬件 | Olares OS | Windows 11 |
|---|---|---|
| AOOSTAR EG02 eGPU dock + RTX 4060 Ti | **设置后可用。** 安装 [Gen1 临时方案](./egpu-olares-os.md#必要时安装临时方案)。 | **设置后可用。** 连接 eGPU 后重新安装 NVIDIA 驱动。 |
| Razer Core X V2 eGPU enclosure + RTX 4090 | **验证中。** 未安装 Gen1 临时方案时开机不稳定。安装后的表现尚未验证。 | **仅在实测的多 eGPU 配置中完成验证。** 尚未单独验证。 |
| Razer Core X eGPU enclosure + RTX 4060 | **尚未验证。** | **仅在实测的多 eGPU 配置中完成验证。** 尚未单独验证。 |
| eGPU dock 或 eGPU enclosure + 桌面版 RTX 5090 | **暂不支持。** NVIDIA 驱动无法初始化，目前没有可用方案。 | **尚未验证。** |

表中的**可用**和**已验证**表示显卡完成过实际负载测试。仅在 Dashboard、设备管理器或 `nvidia-smi` 中识别到显卡，不视为兼容性验证通过。

:::info 多 eGPU
如果需要使用多台 eGPU，建议使用 Windows 11。Olares OS 目前每次只支持连接一台 eGPU。

我们在 Windows 11 上验证了一种通过 Razer Thunderbolt 5 Dock 连接三台 eGPU 的配置：

- AOOSTAR EG02 + RTX 4060 Ti
- Razer Core X V2 + RTX 4090
- Razer Core X + RTX 4060

在该配置中，内置 RTX 5090M 的功耗约为 95 W，而不是 175 W。

多 eGPU 的兼容性取决于 Dock 和连接拓扑。建议先完成单 eGPU 设置，再逐台连接并检查其他设备。
:::

## 选择硬件

准备以下任一种设备：

- 已安装 GPU 的雷电外置显卡设备及其电源适配器。
- 雷电 eGPU dock 或 eGPU enclosure、采用 Turing 或更新架构的桌面版 NVIDIA 显卡，以及符合显卡要求的电源。

:::info 雷电 5
使用认证的雷电 5 线材，建议优先使用外置显卡设备附带的线材。

旧款雷电外置显卡设备可能也能使用，但建议选择雷电 5 设备。
:::

## 设置 eGPU

根据操作系统选择指南：

- [在 Olares OS 上设置 eGPU](./egpu-olares-os.md)
- [在 Windows 11 上设置 eGPU](./egpu-windows.md)

## 获取帮助

如果 Olares One 无法启动、系统看不到 eGPU，或显卡驱动报错，请参考[排查 eGPU 问题](./ts-egpu.md)。
