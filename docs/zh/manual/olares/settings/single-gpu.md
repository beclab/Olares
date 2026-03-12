---
outline: [2,3]
description: 当 Olares 只有一块 GPU 时，配置 GPU 模式并管理应用访问权限。
---
# 管理单 GPU 资源

本文为你介绍 Olares 仅有一块 GPU 时，如何管理 GPU 模式和应用访问权限。

开始之前，请阅读[GPU 资源管理](./gpu-resource.md)，以了解 Olares 中 GPU 的管理机制。

## 打开 GPU 设置

此页面为你展示 GPU 详情，你也可以在此切换 GPU 模式。

1. 前往**设置** > **GPU**。
    ![时间分片-单 GPU](/images/zh/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

2. 在 **GPU 模式**下拉菜单中选择一种模式。

:::info GPU 调度延迟
切换 GPU 模式或恢复应用后，应用与 GPU 的绑定关系可能不会立刻更新到界面中。

如果某个运行中的应用暂时未出现在列表中，请先等待几秒，再点击对应区域中的 <i class="material-symbols-outlined">sync</i> 刷新列表。系统通常会在 GPU 调度完成后自动绑定应用。

**绑定应用**按钮仅在存在待绑定应用时显示。
:::

## 时间分片 

![时间分片-单 GPU](/images/zh/manual/olares/gpu-time-slicing-single.png#bordered){width=90%}

### 添加应用

大多数情况下，正在运行的应用会在 GPU 调度完成后自动绑定，并显示在列表中。

如果目标应用暂时未出现：
1. 等待几秒。
2. 在**绑定应用**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果应用仍未自动绑定，点击**绑定应用**手动添加。

### 移除应用的 GPU 访问权限

1. 暂停该应用。
2. 返回**设置** > **GPU**。
3. 在**绑定应用**区域，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

## 显存分片 <Badge type="tip" text="支持 DGX Spark" />

![显存分片-单 GPU](/images/zh/manual/olares/gpu-mem-slicing-single.png#bordered){width=90%}

### 调整显存分配

1. 在**分配显存**区域，找到目标应用。
2. 点击显存数值旁边的 <i class="material-symbols-outlined">edit_square</i>。
3. 在**编辑 VRAM 分配**对话框，输入所需的显存大小（单位为 GB），然后点击**确认**。

:::warning 注意
分配给所有应用的显存总量不能超过 GPU 的总显存。

如果输入的数值低于应用的最小显存需求，**确认**按钮将不可用。
:::

### 添加应用并分配显存

大多数情况下，正在运行的应用会在 GPU 调度完成后自动绑定，并显示在列表中。

如果目标应用暂时未出现：
1. 等待几秒。
2. 在**分配显存**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果应用仍未自动绑定，点击**绑定应用**，选择应用并分配显存。

### 移除应用的显存分配

1. 先暂停该应用。
2. 返回**设置** > **GPU**。
3. 在**分配显存**区域，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

## 应用独占 <Badge type="tip" text="支持 DGX Spark" />

![应用独占-单 GPU](/images/zh/manual/olares/gpu-app-exclusive-single.png#bordered){width=90%}

### 更改独占应用

1. 先暂停当前独占应用，
2. 在**选取独占应用**区域点击 <i class="material-symbols-outlined">link_off</i>，解绑当前应用。
3. 确保新的目标应用已恢复运行。
4. 等待几秒。
5. 在**选取独占应用**区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
6. 如果系统仍未自动选定新的独占应用，点击**绑定应用**手动设置。

### 设置独占应用

大多数情况下，Olares 会在 GPU 调度完成后自动选择一个正在运行的应用并赋予其独占访问权限。

如果没有显示任何应用：
1. 等待几秒。
2. 在**选取独占应用** 区域点击 <i class="material-symbols-outlined">sync</i> 刷新列表。
3. 如果仍未自动选定应用，点击**绑定应用**手动设置。

### 移除某个应用的独占访问权限

1. 先暂停该应用。
2. 返回**设置** > **GPU**。
3. 在**选取独占应用**区域中，点击 <i class="material-symbols-outlined">link_off</i>，然后点击**确认**。

## 了解更多

- [了解 GPU 管理](./gpu-resource.md)
- [在 Olares 中监控 GPU 使用情况](../resources-usage.md)
- [管理多 GPU 资源](./multi-gpu.md)