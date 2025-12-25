---
outline: [2, 3]
description: 控制管理和优化 Olares显卡资源，支持单节点或多节点环境下的时间分片、独占和显存分片分配模式。
---

# 管理 GPU 使用

:::info
只有 Olares 管理员可以修改 GPU 模式。这样可以避免资源冲突，保证系统整体的 GPU 性能稳定可预测。
:::

Olares 让你充分发挥 GPU 的算力，加速 AI、图像和视频生成及游戏等高负载任务。你可以在 Olares 的**设置**页面，统一管理 GPU 资源。

本文将为你介绍：
- 如何选择合适的 GPU 模式。
- 分配 GPU 前需要了解的基本规则。
- 如何逐步配置 GPU 模式。

## 选择合适的 GPU 模式

如果你不确定该使用哪种模式，可以从本节开始。

### 时间分片
**适用于**：多个应用共享一块 GPU。

- Olares 默认的 GPU 分配模式。
- 多个应用可轮流共享同一块 GPU。
- 任意时刻，只有一个应用占用 GPU 的算力和显存。
- 适合通用场景和多个轻量级应用同时运行。
- 未指定 GPU 的应用会自动调度到处于时间分片模式的 GPU 上。

### 应用独占
**适用于**：为单个高负载应用提供最强、最稳定的性能。

- 一个应用独享单块 GPU 的全部算力和显存。
- 绑定期间，其他应用无法使用该 GPU。
- 适合大语言模型、高负载渲染或高性能游戏等场景。

### 显存分片
**适用于**：同时运行多个指定应用，并限制每个应用的显存用量。

- GPU 显存被划分为固定的配额。
- 每个应用只能使用分配给它的显存额度。
- 多个应用可以同时运行，无需轮流等待。
- 所有显存配额之和不能超过 GPU 的物理显存上限。

## 了解 GPU 分配规则

分配 GPU 之前，请先了解以下规则。

### GPU 与节点

在 Olares 中，每块 GPU 都归属于 Olares 管理的一个特定节点。你可以在**设置** > **GPU** 中查看每块 GPU 所在的节点。

请遵循以下规则以避免报错：

- **仅限同节点**：一个应用只能使用位于同一节点上的 GPU。
- **不可跨节点**：不能将不同节点上的 GPU 分配给同一个应用。

### 切换 GPU

当你将应用切换到另一块 GPU 时，应用会重启。  
具体行为取决于目标 GPU 所在的节点：

| 目标 GPU 位置 | 重启后的结果 |
| --- | --- |
| 同一节点 | 应用可能同时绑定到两块 GPU |
| 不同节点 | 应用会被迁移，并只绑定到目标 GPU |

## 硬件要求

Olares 仅支持 **NVIDIA GPU**。GPU 必须是 **Turing 架构或更新版本**，包括 Turing、Ampere、Ada Lovelace 和 Blackwell。

### 检查 GPU 类型

你可以通过以下方式确认 GPU 是否受支持：

| 已知信息 | 建议 |
| --- | --- |
| GTX 或 RTX 16 系列或更新 | 你的显卡大概率受支持 |
| 其他 NVIDIA 型号 | 请查阅[兼容 GPU 列表](https://github.com/NVIDIA/open-gpu-kernel-modules?tab=readme-ov-file#compatible-gpus) |
| 不清楚型号 | 运行`lspci \| grep -i nvidia`，并对照兼容列表检查 |

:::warning AI 性能
即使 GPU 架构受支持，如果显存不足，某些 AI 应用仍可能无法运行。  
请务必确认GPU 的显存容量足以支撑你的实际工作负载。
:::

## 查看 GPU 状态

查看当前 GPU 及其使用情况：

1. 进入**设置** > **GPU**。
2. 查看列表中的 GPU 型号、所属节点、总显存和当前模式。
   ![GPU 概览](/images/zh/manual/olares/gpu-overview.png#bordered)
3. 点击某块 GPU 进入详情页。

:::tip
如果你的 Olares 只有一块 GPU，Olares 可能会直接打开 GPU 详情页。
:::

## 配置 GPU 模式

按照以下步骤更改 GPU 的使用方式：

1. 进入**设置** > **GPU**。
2. 点击需要配置的 GPU。
3. 在 **GPU 模式**下拉菜单中选择一种模式。

:::warning 重启提示
更改 GPU 模式会解绑该 GPU 上的应用并重启其容器。

重启后，如果仍有 GPU 处于**时间分片**模式，未绑定 GPU 的应用可能会被自动调度并绑定到该 GPU。
:::

### 时间分片

 ![时间分片](/images/zh/manual/olares/gpu-time-slicing.png#bordered)

在**时间分片**模式下，你可以执行以下操作。

<Tabs>
<template #绑定应用>

1. 在**绑定应用**区域，点击**绑定应用**。
2. 选择目标应用并点击**确认**。
</template>

<template #切换-GPU>

1. 在**绑定应用**区域，找到要切换的应用。
2. 点击 <i class="material-symbols-outlined">repeat</i>，选择目标 GPU 并确认。

| 目标 GPU 位置 | 结果 |
| --- | --- |
| 同一节点 | 应用重启后可能会使用多块 GPU |
| 不同节点 | 应用被迁移并仅绑定到目标 GPU |
</template>

<template #解绑应用>

1. 在**绑定应用**区域，找到目标应用。
2. 点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。
</template>
</Tabs>

### 应用独占

![应用独占](/images/zh/manual/olares/gpu-app-exclusive.png#bordered)

在**应用独占** 模式下，你可以执行以下操作。

<Tabs>
<template #绑定应用>

1. 在**选取独占应用**区域，点击**绑定应用**。
2. 选择目标应用并点击**确认**。

该应用将独占此 GPU。
</template>

<template #切换应用>

1. 在**选取独占应用**区域，点击**切换应用**。
2. 选择新的应用并确认。

原应用将被解绑，新应用成为独占应用。
</template>

<template #切换-GPU>

1. 在**选取独占应用**区域，点击 <i class="material-symbols-outlined">repeat</i>。
2. 选择目标 GPU 并确认。
</template>

<template #解绑应用>

1. 在**选取独占应用**区域，点击 <i class="material-symbols-outlined">link_off</i>。
2. 点击**确认**。
</template>
</Tabs>

### 显存分片

![显存分片](/images/zh/manual/olares/gpu-memory-slicing.png#bordered)

在**显存分片**模式下，你可以绑定应用并为其设置显存上限：

1. 在**分配显存**区域，点击**绑定应用**。
2. 选择目标应用，设置显存大小（GB），并点击**确认**。
3. 如有需要，重复以上步骤为其他应用分配显存。

所有显存配额之和不能超过 GPU 的物理显存上限。

## 了解更多
- [监控 Olares 中的显卡使用情况](../resources-usage.md)