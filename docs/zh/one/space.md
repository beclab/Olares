---
outline: [2,3]
description: 了解如何将你的设备关联到 Olares Space，并监控云流量和备份存储使用情况。
head:
  - - meta
    - name: keywords
      content: Olares, Olares One, Olares Space, 监控流量使用, 账单
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/space.md)。
:::

# 在 Olares Space 中监控流量 <Badge type="tip" text="10 min" />

Olares Space 是设备的云端管理平台。虽然你的 Olares 设备在本地运行，但 Olares Space 允许你从任何网页浏览器监控云流量使用情况和管理账单。

本指南介绍如何将你的 Olares 设备关联到 Olares Space 并跟踪流量使用情况。

:::info 使用和账单
如果你使用云服务（如远程访问（从家庭网络外部访问设备）或云备份），监控使用量非常重要。如果超出订阅计划的限制，这些服务可能会产生费用。
:::

## 学习目标

- 将你的 Olares 设备关联到 Olares Space 账户。
- 登录 Olares Space。
- 监控云流量消耗。

## 开始之前

确保你已在移动设备上安装 LarePass 应用，并使用你的 Olares ID 登录。

## 步骤 1：关联 Olares Space

在 Olares Space 中查看统计数据之前，你必须授权 Olares Space 访问设备的状态。你可以通过 LarePass 手机应用完成此操作。

1. 在手机上打开 LarePass 应用，进入 **Settings** > **Integration**。

   ![LarePass 设置](/images/one/larepass-settings-integration.png#bordered)

2. 点击右上角的 <i class="material-symbols-outlined">add</i>。

   ![集成](/images/one/larepass-integration.png#bordered)

3. 从列表中点击 **Space**。

   ![添加 Space 集成](/images/one/larepass-integration-add.png#bordered)

   完成后，你的 Olares Space 账户已关联到物理设备，Olares Space 账户会显示在 **Integration** 列表中。

   ![Olares Space 账户已集成](/images/one/larepass-integration-add-space.png#bordered)

## 步骤 2：登录 Olares Space

通过网页浏览器访问 Olares Space 仪表板。

1. 访问 https://space.olares.com/。

   ![Olares Space 登录页面](/images/one/olares-space-login.png#bordered)

2. 登录 Olares Space：

    a. 在 LarePass 应用中，进入 **Settings**，然后点击扫描图标。

      ![扫码登录](/images/one/scan-olares-space.png#bordered)

    b. 扫描电脑屏幕上的二维码，点击 **Confirm**。现在你已经登录了 Olares Space。

<!--## 步骤 3：监控系统使用情况

Olares Space 将本地硬件监控与云流量使用情况分开。
### 检查设备健康状态

登录页面是 **Olares** 选项卡。使用 **Resource Monitor** 部分检查已连接设备的实时状态：
- CPU 和内存：显示设备当前使用的计算能力。
- 磁盘：显示设备硬盘上使用的总存储空间。
- 活动主机：列出你特定设备的连接状态和运行时间。

![Olares Space Olares 选项卡](/images/one/olares-space-olares-tab.png#bordered)-->

## 步骤 3：检查流量使用情况

监控你已使用的远程访问数据量，以避免超出订阅的月度限制。

1. 从左侧导航面板点击 **Usage & billing**。

      ![Olares Space 流量详情](/images/one/olares-space-traffic-usage.png#bordered)

2. 在 **Usage** 选项卡上，查看 **Traffic details** 部分。默认显示最新计费周期的流量详情。

   - 进度条：准确显示你已消耗的数据量与计划限制的对比。例如，0.05 GB/2.0 GB。
   - 每日图表：柱状图显示你每日的数据使用量，帮助你发现活动量的突然增加。

3. 如果你需要查看之前计费周期的使用量：

   a. 点击日期范围下拉列表。

      ![Olares Space 按月查看流量](/images/one/space-traffic-filter.png#bordered)

   b. 选择特定的计费周期以查看该月的流量历史。例如，2025-12-02 ~ 2026-01-02。

## 资源

- [Olares Space 计划和定价](https://space.olares.com/plans)
- [管理 Olares](../manual/space/manage-olares.md)
- [Olares Space 账单](../manual/space/billing.md)
