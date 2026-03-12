---
outline: [2,3]
description: 当 Olares 配置了多块 GPU 时，配置 GPU 模式、在 GPU 之间重新分配应用，并管理 GPU 访问权限。
---
# 管理多 GPU 资源

本文为你介绍 Olares 配置了多块 GPU 时，如何管理 GPU 模式和应用分配。每块 GPU 都需要单独配置。

## 打开 GPU 设置

此页面会列出所有可用 GPU，包括每块 GPU 的型号、节点、总 VRAM 和当前模式。你还可以选择某个 GPU 并更改其模式。

1. 前往 **设置** > **GPU**。
    ![GPU 概览-多 GPU](/images/zh/manual/olares/gpu-overview.png#bordered){width=90%}

2. 点击你要配置的 GPU。
3. 在 **GPU 模式**下拉菜单中选择一种模式。

:::tip 找不到目标应用？
如果应用在刷新后仍未出现在当前 GPU 上，它可能已经被分配到了同一节点上的另一块 GPU，或者其他节点上的 GPU。

请检查其他 GPU，确认该应用当前被分配的位置。如有需要，再使用**切换 GPU**将其重新分配到目标 GPU。
:::

## 时间分片

:::info
DGX Spark 当前不支持时间分片模式。
:::

![时间分片-多 GPU](/images/zh/manual/olares/gpu-time-slicing-multi.png#bordered){width=90%}

### 添加应用

大多数情况下，应用会在 GPU 调度完成后自动分配到合适的 GPU，并显示在列表中。

如果目标应用暂时未出现在当前 GPU：

1. 等待几秒钟。
2. 在**绑定应用**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果应用仍未出现在当前 GPU，请检查其他 GPU，确认它是否已被分配到其他位置。
4. 如果系统仍未自动完成分配，再点击**绑定应用**手动添加。

### 将应用重新分配到另一块 GPU

1. 在**绑定应用**区域中找到该应用。
2. 点击 <i class="material-symbols-outlined">repeat</i>。
3. 选择目标 GPU 并点击**确认**。

在此过程中，应用会持续运行，不会中断。

### 移除应用对当前 GPU 的访问权限

1. 如果该应用只使用了当前这块 GPU 的资源，请先将应用暂停。
   - 前往**应用商店** > **我的 Olares**，从下拉菜单中选择**暂停**。
   - 或前往**设置** > **应用**，选择目标应用后点击 **暂停**。
2. 返回**设置** > **GPU**。
3. 在**绑定应用**区域中，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

:::tip 从多块 GPU 解绑
如果该应用在同一节点上的其他 GPU 上仍有分配，你可以在不暂停应用的情况下，将其从当前 GPU 上移除。
:::

## 显存分片

![显存分片-多 GPU](/images/zh/manual/olares/gpu-mem-slicing-multi.png#bordered){width=90%}

### 调整显存分配

1. 在**分配显存**区域中，找到目标应用。
2. 点击显存数值旁边的 <i class="material-symbols-outlined">edit_square</i>。
3. 在**编辑 VRAM 分配**对话框中，输入所需的显存大小（单位为 GB），然后点击**确认**。

:::warning 注意
当前 GPU 上所有应用分配的显存总量不能超过该 GPU 的总显存。

如果输入值低于应用运行所需的最小值，**确认**按钮将不可用。
:::

### 添加应用

大多数情况下，应用会在 GPU 调度完成后自动分配到合适的 GPU，并显示在列表中。

如果目标应用暂时未出现在当前 GPU：

1. 等待几秒钟。
2. 在**分配显存**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果应用仍未出现在当前 GPU，请检查其他 GPU，确认应用是否已被分配到其他位置。
4. 如果系统仍未自动完成分配，再点击**绑定应用**手动添加。

### 将应用重新分配到另一块 GPU

1. 在**分配显存**区域中，找到该应用并点击 <i class="material-symbols-outlined">repeat</i>。
2. 选择目标 GPU。
3. 点击**确认**。

### 从当前 GPU 移除应用的显存分配

1. 如果该应用只使用了当前这块 GPU 的资源，请先将应用暂停。
   - 前往**应用商店** > **我的 Olares**，从下拉菜单中选择**暂停**。
   - 或前往**设置** > **应用**，选择目标应用后点击**暂停**。
2. 返回**设置** > **GPU**。
3. 在**分配显存**区域中，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

:::tip
如果该应用在同一节点上的其他 GPU 上也分配了显存，你可以在不暂停应用的情况下，从当前 GPU 移除它的显存分配。
:::

## 应用独占

![应用独占-多 GPU](/images/zh/manual/olares/gpu-app-exclusive-multi.png#bordered){width=90%}

### 将独占应用分配到另一块 GPU

1. 在**选取独占应用**区域中，点击 <i class="material-symbols-outlined">repeat</i>。
2. 选择目标 GPU。
3. 点击**确认**。

### 设置独占应用

大多数情况下，系统会在 GPU 调度完成后自动选择一个合适的运行中应用作为独占应用。

如果当前 GPU 上尚未显示独占应用：

1. 等待几秒钟。
2. 在**选取独占应用**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果应用仍未出现在当前 GPU，请检查其他 GPU，确认它是否已被分配到其他位置。
4. 如果系统仍未自动完成选择，再点击**绑定应用**手动设置。

### 移除应用对当前 GPU 的独占访问权限

1. 如果该应用只使用了当前这块 GPU 的资源，请先将应用暂停。
   - 前往**应用商店** > **我的 Olares**，从下拉菜单中选择**暂停**。
   - 或前往**设置** > **应用**，选择目标应用后点击**暂停**。
2. 返回**设置** > **GPU**。
3. 在**选取独占应用**区域中，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

:::tip
如果该应用在同一节点上的其他 GPU 上也分配了资源，你可以在不暂停应用的情况下，将其从当前 GPU 上移除。
:::

## 了解更多

- [GPU 资源管理](./gpu-resource.md)
- [在 Olares 中监控 GPU 使用情况](../resources-usage.md)
- [管理单 GPU 资源](./single-gpu.md)