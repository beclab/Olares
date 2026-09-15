---
outline: [2, 3]
description: 通过 Olares 设置或 Olares Space 网页查看系统状态、主机、云服务配额和流量用量。
head:
  - - meta
    - name: keywords
      content: Olares, Olares Space, 系统状态, 活跃主机, 备份空间, 流量用量, 仪表板
---
# 查看 Olares 状态与 Olares Space 用量

需要远程查看 Olares 状态，或检查备份空间、反向代理流量等云服务用量时，可以使用 Olares Space。如需排查 CPU、内存、磁盘、网络、Pod、GPU 或应用的实时资源使用，请改用 Olares [仪表板](../olares/resources-usage.md)。

不同入口提供的信息不同：

| 入口 | 适合查看 |
|---|---|
| Olares **设置** > 左上角头像 > **Olares Space** | 推荐的登录入口，以及已关联账户、套餐、备份空间和流量用量摘要 |
| [Olares Space](https://www.olares.com/space) | 远程系统状态、活跃主机、近期流量和计费周期用量 |
| Olares **仪表板** | 本机资源和应用的详细诊断信息 |

## 在设置中连接 Olares Space

请从 Olares 设置中的专属页面连接 Olares Space。连接成功后，可以在该页面查看用量，Olares Space 也会自动出现在**设置** > **集成**中。

:::warning 从 Olares Space 页面连接
不要从**设置** > **集成** > **添加账户** > **Olares Space** 开始操作。该入口暂不支持绑定账户。
:::

1. 打开 Olares **设置**。
2. 点击左上角的头像。
3. 点击 **Olares Space** 卡片。
4. 页面显示二维码后，使用移动设备上的 LarePass 扫码。
5. 等待页面显示 Olares Space 账户和用量信息。

## 在设置中查看摘要

1. 打开 Olares **设置**。
2. 点击左上角的头像。
3. 点击 **Olares Space** 卡片。
4. 查看已关联的账户、订阅套餐、备份空间、反向代理服务和流量用量。

该入口适合快速确认配额。如需查看状态详情、指定时间段或计费周期历史，请打开 Olares Space 网页。

## 在 Olares Space 中查看资源用量

Olares Space 的概览页提供 CPU、内存和磁盘用量摘要。无法在本地访问 Olares 时，可以通过这里远程确认系统是否在线、资源是否接近上限。

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

## 查看近期流量

检查近期流量使用情况，发现突增并避免超出套餐限制。

:::info
对于自托管 Olares 用户，请关注反向代理服务的流量统计。如果超出月度配额，速度将限速至 5 Mbps。免费的替代方案是使用 [LarePass VPN](../larepass/private-network.md) 或自行托管 FRP 服务器。
:::

1. 在 **Olares** 页面，选择 **Usage statistics** 标签页。

   ![Olares Space 流量使用](/images/how-to/space/olares_usage_statistics1.png#bordered)

2. 找到 **Traffic Usage** 区域。默认展示所有用户最近 12 小时的流量。
3. 要更改时间范围，从 **Last 12 hours** 下拉菜单中选择一个。
4. 要查看特定用户的流量，从 **All Users** 下拉菜单中选择该账号。

## 查看计费周期流量

查看月度流量使用情况，了解当前计费周期内已消耗多少数据。

1. 从左侧导航栏选择 **Usage & billing**。

   ![Olares Space 流量详情](/images/one/olares-space-traffic-usage.png#bordered)

2. 在 **Usage** 标签页，找到 **Traffic details** 区域。默认展示最新计费周期的流量详情。

   - **进度条**：显示已消耗数据量与套餐限制的对比。例如，`0.05 GB / 2.0 GB`。
   - **每日图表**：按天展示数据使用量的柱状图，帮助你发现活动量的突然增加。

3. 要查看之前的计费周期，从日期范围下拉菜单中选择一个。

   ![Olares Space 按月查看流量](/images/one/space-traffic-filter.png#bordered)

4. 要查看特定用户的流量，从 **All Users** 下拉菜单中选择该账号。
