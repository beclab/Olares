---
outline: [2, 3]
description: 介绍云端 Olares 实例的管理功能，包含系统监控面板、工作节点添加方法和共享 GPU 使用方案，助力提升系统运行效率。
---

# 管理 Olares

本页介绍如何在 Olares Space 中管理 Olares，包括监控系统数据、添加工作节点和管理云服务。

## 查看系统状态

你可以通过 **Olares Space** 监控 Olares 的系统状态：

1. 在 LarePass 应用中，进入**设置** > **集成**。
2. 点击右上角的 <i class="material-symbols-outlined">add</i>，将 Olares Space 账号与 Olares 设备关联，授权 Olares Space 访问系统数据。
3. 登录 [**Olares Space**](https://space.olares.com/)。
4. 在 **Olares** 页面的系统面板中查看**存储使用量**和**流量消耗**。

![系统面板](/images/how-to/space/my_olares.jpg#bordered)

:::info 注意
对于自托管 Olares 用户，请重点关注内网穿透服务的**流量统计**和备份服务的**存储使用量**。这些服务可能会根据使用情况产生费用。
:::

## 添加工作节点

云端 Olares 用户可以通过添加工作节点来提升性能：

1. 点击右上角的 <i class="material-symbols-outlined">more_horiz</i>，选择**添加工作节点**。
2. 在引导页面选择所需的硬件配置。
3. 查看存储和流量费用。
4. 确认订单并提交。

## 销毁 Olares

如果不再需要 Olares 服务，可以按以下步骤销毁实例：

1. 点击右上角的 <i class="material-symbols-outlined">more_horiz</i>。
2. 选择**销毁 Olares**。
3. 确认操作并结算使用费用：
   - 如果符合退款条件，退款金额将返还到账户余额
   - 如需补充支付，请确认并完成支付

## 共享 GPU 方案

目前我们不提供包含 GPU 的云实例。不过，对于需要 GPU 功能的用户，我们通过 rCuda 提供共享 GPU 方案。这个方案特别适合 Stable Diffusion 等应用，每张图片的成本约为 0.02 美元。

::: info 注意
对于大语言模型（LLM），共享 GPU 方案仍在开发中，可能需要进一步优化。
:::

如果你需要 GPU 支持，请通过 [Discord](https://discord.com/invite/BzfqrgQPDK) 联系我们。