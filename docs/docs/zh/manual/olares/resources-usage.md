---
outline: [2, 3]
description: 通过 Olares 的仪表盘应用监控系统和应用状态。查看资源使用指标、应用性能数据，确保系统稳定运行。
---

# 监控系统和应用状态

Olares 的**仪表盘**应用类似于 Windows 资源管理器，为你提供系统状态的集中视图，无需技术背景。从主面板中，你可以查看资源使用模式和详细的指标数据。

:::info
当你的 Olares 应用暴露在公网时，它们会因外部访问产生 FRP 流量成本。要监控这些成本和流量，请参阅[查看系统状态](../../space/manage-olares.md#查看系统状态)。
:::

## 访问监控仪表板

通过以下专业仪表板查看系统状态：

- **概览**：显示当前资源使用情况和系统健康状态。
- **应用**：显示运行中的应用及其状态。

## 概览

### 查看物理资源

在**概览**中直接监控以下四个核心指标：
- CPU 使用率
- 内存使用量
- 磁盘使用量
- Pod 状态

![Dashboard overview](/images/manual/olares/dashboard-overview.png#bordered)

### 查看详细指标

点击**详情**，查看过去 7 天的综合监控数据。

使用右上角的下拉菜单更改时间范围，或点击 <i class="material-symbols-outlined">refresh</i> 更新监控数据。

以下指标帮助你保持系统的最佳性能：

| 指标       | 描述            | 影响                 |
|----------|---------------|--------------------|
| CPU 用量   | CPU 资源的使用百分比  | 持续高峰会导致系统变慢        |
| 内存用量     | 内存的使用百分比      | 影响应用性能和稳定性         |
| CPU 平均负载 | 活跃进程的平均数量     | 高负载表明系统过载          |
| 磁盘用量     | 磁盘空间的使用百分比    | 对数据可靠性至关重要，需防止过度使用 |
| Inode 用量 | Inode 的使用百分比  | 耗尽将阻止新文件的创建        |
| 磁盘吞吐     | 数据传输速率（MB/s）  | 对大文件传输非常重要         |
| IOPS     | 每秒输入/输出操作数    | 对小文件或随机数据访问至关重要    |
| 网络流量     | 网络使用情况（Mbps）  | 反映网络速度和质量          |
| 容器组状态    | 按状态划分的 Pod 数量 | 反映应用的健康状态          |

![Physical resource monitoring](/images/manual/olares/physical-resource-monitoring.png#bordered)

### 查看资源配额

你可以查看 Olares 管理员分配的资源配额。

![Resource quota](/images/manual/olares/resource-quota.png#bordered)

:::warning 警告
当资源配额不足时，可能会出现以下问题：

- 系统性能下降。
- 无法安装新应用。
- 资源密集型应用会自动暂停。
:::

### 追踪应用性能

**使用排名**面板显示 CPU 和内存资源消耗最高的前 5 个应用。要查看完整的应用资源使用列表，点击**更多**。

![Usage ranking](/images/manual/olares/usage-ranking.png#bordered)

## 应用

**应用**仪表面板帮助你通过多种排序和筛选选项监控应用的资源使用模式。

使用右上角的下拉菜单，根据以下资源消耗指标排序应用：
- CPU 使用率
- 内存使用率
- 入站流量
- 出站流量

![Applications](/images/manual/olares/applications.png#bordered)

在升序和降序之间切换，找出资源消耗最高或最低的应用。

对于支持多入口的应用（如 WordPress），你可以点击图标切换不同入口类型，并查看其对应的资源指标。
![Multiple entrances](/images/manual/olares/multiple-entrances.png){width=40%}

:::tip 提示
* 当应用列表较长时，可通过页面顶部的搜索框快速定位特定应用。
* 定期检查资源消耗模式可帮助你识别可能需要优化或关注的应用。
:::