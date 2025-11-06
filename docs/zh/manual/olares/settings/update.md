---
description: 了解如何升级 Olares 版本，保持系统功能和安全性。
---
# 更新 Olares

Olares 定期发布新版本，带来功能改进和用户体验优化。本文档说明如何检查和安装系统更新。

:::info 仅管理员可以升级
只有 Olares 管理员可以执行系统更新。更新将应用于同一 Olares 集群内的所有成员。
:::

:::tip 提示
有关 Olares 的版本控制实践及当前跨次版本升级（比如从 `1.10.5` 升到 `1.11.0`）的限制，请参阅 [Olares 版本说明](../../../developer/install/versioning.md)。
:::


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