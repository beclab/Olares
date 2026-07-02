---
outline: [2, 3]
description: 监控你的 Olares 系统健康状况。了解如何查看 CPU 和内存使用情况、管理磁盘存储，以及识别资源占用高的应用。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, monitor system, system resources, app status, CPU usage, memory usage, disk space, fan speed
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/dashboard.md)。
:::

# 监控系统与应用状态 <Badge type="tip" text="10 min" />

Dashboard 应用提供了一个集中、实时的系统状态视图。你可以用它检查可用存储空间、监控硬件温度，以及识别哪些应用正在占用最多资源。

本指南将带你完成维护健康系统的最常见任务。

## 学习目标

- 查看系统状态概览。
- 检查特定硬件的详细使用情况。本指南以"磁盘"为例。
- 监控风扇速度和温度。
- 识别并管理资源密集型应用。

## 开始之前

熟悉用于衡量系统性能的关键指标。

| 指标 | 说明 | 重要性 |
|:-----------------|:----------------------------------------|:-------------------------------------------------------------------|
| CPU usage | 处理器使用率百分比 | 持续高使用率会使系统变慢且无响应。 |
| Memory usage | 内存使用百分比 | 如果内存已满，应用可能会崩溃或冻结。 |
| Average CPU load | 活跃进程平均数 | 高负载表示系统过载。 |
| Disk usage | 存储空间填充百分比 | 空间不足会阻止保存新文件或安装应用。 |
| Inode usage | 索引节点（inode）使用百分比 | 耗尽会阻止新文件创建。 |
| Disk throughput | 数据传输速率（MB/s） | 对大文件传输很重要。 |
| IOPS | 每秒输入/输出操作数 | 对小文件或随机数据访问至关重要。 |
| Network traffic | 互联网使用量（Mbps） | 高流量会降低远程访问和下载速度。 |
| Pod status | 活跃应用容器数量 | 指示你的应用正在运行、挂起还是失败。 |
| Fan speed | 散热风扇转速（RPM） | 较高转速表示系统正在努力降温。 |

## 检查系统健康状况

**Overview** 页面让你一目了然地查看设备健康状况。

1. 从 Launchpad 打开 Dashboard 应用。

    ![查找 Dashboard 应用](/images/one/find-dashboard.png#bordered)   

    默认进入 **Overview** 页面。

   ![Dashboard 概览](/images/one/dashboard-overview.png#bordered)

2. 查看 **Cluster's physical resources** 部分。此部分的卡片提供了硬件状态的即时快照：

    - **CPU core**：系统的"大脑"。高百分比表示处理繁重。
    - **Memory Gi**：运行应用的"工作空间"。如果已满，系统可能会变慢或无响应。
    - **Disk**：本地存储空间使用情况。
    - **Pods**：系统上运行的活跃应用单元总数。
    - **GPU**：图形处理能力，用于 AI 任务或媒体渲染。
    - **Network**：实时上传和下载速度。
    - **Fan**：当前散热状态。

## 查看资源详情

点击任意资源卡片查看详细指标。以常见的"管理存储（磁盘）"任务为例。

1. 在 **Overview** 页面的 **Cluster's physical resources** 部分，点击 **Disk Gi** 卡片。

    ![Disk 卡片](/images/one/dashboard-disk-card.png#bordered)   

2. 在 **Disk details** 面板中，你可以查看以下信息：

    - 身份和状态：磁盘名称（例如 nvme0n1）、类型（SSD）和整体健康状态（例如 Normal）。
    - 存储使用情况：可视化条形图，显示已用空间与可用空间的精确数量。
    - 硬件规格：技术细节，包括型号名称、序列号、接口协议（例如 NVMe）和总容量。
    - 健康指标：统计数据，例如当前温度、总通电时长和总写入数据量。

    ![Disk 详情](/images/one/dashboard-disk-details.png#bordered)    

3. 要查看具体哪些文件夹占用了空间，点击右上角的 **Occupancy analysis**。

    ![Disk 详情-占用分析](/images/one/dashboard-occupancy.png#bordered) 

    此视图列出了每个文件系统，帮助你精确查看存储空间的分配情况。

    ![Disk 详情-存储使用](/images/one/dashboard-storage-usage.png#bordered){width=90%} 

你可以按照相同的模式检查其他资源。

## 监控硬件状态

专门的 **Fan** 面板帮助你确保 Olares One 没有过热。

1. 在 **Overview** 页面，找到 **Fan** 卡片。

    ![Dashboard Fan 卡片](/images/one/dashboard-fan-card.png#bordered) 

2. 点击它查看实时统计信息：
  
    - Fan speed：当前 RPM（每分钟转数）。
    - Temperature：主要硬件组件的当前温度。
    - Power：GPU 的当前功耗。

    ![Dashboard Fan 详情](/images/one/dashboard-fan-details.png#bordered) 

    :::tip 调整散热模式
    要更改风扇配置文件（例如从 **Silent mode** 切换到 **Performance mode**），请前往 **Settings** > **My hardware** > **Power mode**。
    :::

## 追踪应用性能

如果系统感觉变慢，可能是某个特定应用消耗了过多资源。

### 快速排名

在 **Overview** 页面，向下滚动到 **Usage ranking** 部分。这里列出了当前使用最多 CPU 或内存的前 5 个应用。

![Dashboard 使用排名](/images/one/dashboard-usage-ranking.png#bordered) 

### 应用详细列表

要查看所有运行中服务的完整视图：

1. 从左导航窗格点击 **Applications**。
2. 使用右上角的排序下拉菜单对列表进行排序：

    - 按 CPU usage 排序：找出占用处理器最多的应用。
    - 按 Memory usage 排序：找出占用内存最多的应用。
    - 按 Inbound traffic 排序：找出下载数据最多的应用。
    - 按 Outbound traffic 排序：找出上传数据最多的应用。

    ![Applications dashboard](/images/one/dashboard-applications.png#bordered)   

## 后续步骤

如果你发现某个应用消耗了过多资源，可以采取措施恢复系统速度。例如：
- 重启应用：应用通常会因为临时错误而消耗过多资源。重启通常可以解决问题。
- 停止或卸载应用：如果某个应用持续拖慢系统速度且你不需要它，可以停止或完全卸载它以释放资源用于其他任务。

## 资源

- [卸载应用](../../manual/olares/market/market.md#uninstall-applications)
- [我的硬件](../../manual/olares/settings/my-olares.md#my-hardware)
