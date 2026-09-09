---
description: 了解如何从 LarePass 或设置中升级 Olares 版本，保持系统功能和安全性。
---
# 更新 Olares

Olares 定期发布新版本，带来功能改进和用户体验优化。本文档说明如何检查和安装系统更新。

:::info 仅管理员可以升级
只有 Olares 管理员可以执行系统更新。更新将应用于同一 Olares 集群内的所有成员。
:::

:::tip 提示
有关 Olares 的版本控制实践及当前跨次版本升级（比如从 `1.10.5` 升到 `1.11.0`）的限制，请参阅 [Olares 版本说明](../../../developer/install/versioning.md)。
:::

## 检查并安装更新
:::tip 提示
更新前请阅读发布说明，了解新功能和重要变更。
:::

你可以从 LarePass 应用或 Olares 设置中更新 Olares。

<Tabs>
<template #在-LarePass-中>

1. 在手机上打开 LarePass，进入**设置**。
2. 在**我的 Olares** 卡片里，点击**系统**，进入 **Olares 管理**页面。
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

</template>
<template #在-Olares-设置中>

1. 打开**设置**，进入 **System** > **我的 Olares** > 当前版本。
2. 有可用新版本时，点击**立即升级**。

更新完成后你会看到确认消息。

</template>
</Tabs>

## 手动升级 `olaresd`

`olaresd` 是 Olares 系统的核心守护进程，负责提供多种关键系统管理功能。在某些情况下，升级 Olares 版本之后，可能还需要手动升级 `olaresd` 以解决某些服务无法正常访问的问题。

参考版本对应的[发布说明](https://github.com/beclab/Olares/releases/)，确认是否需要手动升级。

要手动升级 `olaresd`：

1. 打开控制面板，左侧点击**终端** > **Olares**。
   ![Open terminal in Olares](/images/zh/manual/tasks/olares-terminal-in-control-hub.png#bordered)
2. 在终端中执行以下命令：
   ```bash
   curl -SsfL https://cdn.olares.cn/upgrade_1_11_6.sh | bash -
   ```
   其中：
   - `1_11_6` 表示将 `olaresd` 和 `olares-cli` 升级到 `1.11.6` 版本。
