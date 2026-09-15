---
outline: [2, 3]
description: 在 Olares 设置或 LarePass 中查看已安装的 Olares OS 版本，并通过 LarePass 下载和安装系统更新。
---
# 查看并更新 Olares

可以在 Olares 设置或 LarePass 中查看当前安装的 Olares OS 版本。系统升级目前只能通过 LarePass 移动端执行。

:::info 仅管理员可以升级
只有 Olares 管理员可以执行系统更新。更新将应用于同一 Olares 集群内的所有成员。
:::

:::tip 提示
有关 Olares 的版本控制实践及当前跨次版本升级（比如从 `1.10.5` 升到 `1.11.0`）的限制，请参阅 [Olares 版本说明](../../../developer/install/versioning.md)。
:::

## 查看当前版本

<Tabs>
<template #在-Olares-设置中>

1. 打开**设置**。
2. 点击左上角的头像。
3. 在**当前版本**中查看已安装的 Olares OS 版本。

</template>
<template #在-LarePass-中>

1. 在手机上打开 LarePass，进入**设置**。
2. 在**我的 Olares** 卡片里，点击**系统**，进入 **Olares 管理**页面。
3. 点击页面顶部的设备信息区域。**系统版本**字段会显示已安装的 Olares OS 版本。

</template>
</Tabs>

## 在 LarePass 中安装更新

:::tip
更新前请阅读 [Olares 发布说明](https://github.com/beclab/Olares/releases/)，了解重要变更和版本特定说明。
:::

1. 在手机上打开 LarePass，进入**设置**。
2. 在**我的 Olares**卡片里，点击**系统**，进入 **Olares 管理**页面。
3. 点击**系统更新**。
4. 确认**新版本**字段中的可更新版本信息，然后点击**升级**。
   ![检查可用更新](/images/zh/manual/larepass/check-version1.png#bordered)
5. 在弹出的对话框中，选择升级方式：
   - **仅下载**：Olares 只下载更新包，你可以照常使用 Olares。
   - **下载并升级**：Olares 会下载更新包，并在你确认重启后开始安装。
   ![升级方法](/images/zh/manual/larepass/olares-upgrade2.png#bordered)
6. 如果你选择了**仅下载**，在**系统更新**页面点击**升级**，开始更新流程。如果你选择了**下载并升级**，在出现提示时确认重新启动，即可开始安装。
7. 等待更新和重启完成。出现成功消息表示升级已完成。
   ![升级成功提示](/images/zh/manual/larepass/olares-upgrade-success.png#bordered)
8. 刷新你的 Olares 桌面以同步最新的系统更改。

设备重启后，再次[查看当前版本](#查看当前版本)，确认版本号与所选更新一致。
