---
outline: [2, 3]
description: 在 Olares Space 中监控 Olares 的系统状态，包括存储使用量和流量消耗。
---
# 在 Olares Space 中监控 Olares

本页介绍如何在 Olares Space 中监控 Olares 的系统状态，包括存储使用量和流量消耗。

## 查看系统状态

你可以通过 **Olares Space** 监控 Olares 的系统状态：

1. 在 LarePass 应用中，进入**设置** > **集成**。
2. 点击右上角的 <i class="material-symbols-outlined">add</i>，将 Olares Space 账号与 Olares 设备关联，授权 Olares Space 访问系统数据。
3. 登录 [**Olares Space**](https://space.olares.com/)。
4. 在 **Olares** 页面的系统面板中查看**存储使用量**和**流量消耗**。

![系统面板](/images/how-to/space/my_olares.jpg#bordered)

:::info
对于自托管 Olares 用户，请重点关注内网穿透服务的**流量统计**和备份服务的**存储使用量**。这些服务可能会根据使用情况产生费用。
:::

## 检查流量使用情况

监控你已使用的远程访问数据量，以避免超出订阅计划的月度限制。

1. 登录 [Olares Space](https://space.olares.com/)。
2. 从左侧导航面板点击 **Usage & billing**。

   ![Olares Space 流量详情](/images/one/olares-space-traffic-usage.png#bordered)

3. 在 **Usage** 选项卡上，查看 **Traffic details** 部分。默认显示最新计费周期的流量详情。

   - **进度条**：显示你已消耗的数据量与计划限制的对比。例如，0.05 GB/2.0 GB。
   - **每日图表**：柱状图显示你每日的数据使用量，帮助你发现活动量的突然增加。

4. 如果你需要查看之前计费周期的使用量：

   a. 点击日期范围下拉列表。

      ![Olares Space 按月查看流量](/images/one/space-traffic-filter.png#bordered)

   b. 选择特定的计费周期以查看该月的流量历史。例如，2025-12-02 ~ 2026-01-02。