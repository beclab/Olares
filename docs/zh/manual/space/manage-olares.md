---
outline: [2, 3]
description: 在 Olares Space 中监控 Olares 的系统状态和流量使用情况。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, 监控 Olares, 系统状态, 资源使用, 流量使用
---
# 监控 Olares 状态与流量

本页介绍如何在 Olares Space 中监控 Olares 的系统状态和流量使用情况。

## 开始之前

要在 Olares Space 中监控 Olares，你必须先授权 Olares Space 访问你的系统数据。在 LarePass 中将 Olares Space 账号与 Olares 设备关联：

1. 在移动设备上打开 LarePass 应用，进入**设置** > **集成**。
2. 点击右上角的 <i class="material-symbols-outlined">add</i>，然后选择 **Olares Space**。

## 监控资源用量

检查 CPU、内存和磁盘用量，确保 Olares 有足够资源。

1. 在 **Olares** 页面，选择 **Overview** 标签页。

   ![Olares 页面 Overview 标签页](/images/how-to/space/olares_page_overview.png#bordered)

2. 找到 **Resource Monitor** 区域。它展示 CPU、内存和磁盘的实时用量。

   | 指标 | 说明 |
   | ---- | ---- |
   | **CPU (Cores)** | 当前 CPU 使用量及可用核心总数。 |
   | **Memory (GB)** | 当前内存使用量及总可用内存。 |
   | **Disk (GB)** | 当前磁盘使用量及总可用磁盘空间。 |

## 查看活跃主机

检查 Olares 集群中当前运行的主机及其状态。

1. 在 **Olares** 页面，选择 **Overview** 标签页。
2. 找到 **Active hosts** 区域。它展示当前 Olares 集群中运行的主机。

## 查看流量使用

检查近期流量使用情况，发现突增并避免超出套餐限制。

:::info
对于自托管 Olares 用户，请重点关注内网穿透服务的流量统计。这些服务可能会根据使用情况产生费用。
:::

1. 在 **Olares** 页面，选择 **Usage statistics** 标签页。

   ![Olares Space 流量使用](/images/how-to/space/olares_usage_statistics1.png#bordered)

2. 找到 **Traffic Usage** 区域。默认展示所有用户最近 12 小时的流量。
3. 要更改时间范围，从 **Last 12 hours** 下拉菜单中选择一个。
4. 要查看特定用户的流量，从 **All Users** 下拉菜单中选择该账号。

## 查看账单周期流量

查看月度流量使用情况，了解当前计费周期内已消耗多少数据。

1. 从左侧导航栏选择 **Usage & billing**。

   ![Olares Space 流量详情](/images/one/olares-space-traffic-usage.png#bordered)

2. 在 **Usage** 标签页，找到 **Traffic details** 区域。默认展示最新计费周期的流量详情。

   - **进度条**：显示已消耗数据量与套餐限制的对比。例如，`0.05 GB / 2.0 GB`。
   - **每日图表**：按天展示数据使用量的柱状图，帮助你发现活动量的突然增加。

3. 要查看之前的计费周期，从日期范围下拉菜单中选择一个。

   ![Olares Space 按月查看流量](/images/one/space-traffic-filter.png#bordered)

4. 要查看特定用户的流量，从 **All Users** 下拉菜单中选择该账号。
