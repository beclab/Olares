---
outline: [2, 3]
description: 将双节点 Olares 集群从 1.12.5 升级到 1.12.6。
head:
  - - meta
    - name: keywords
      content: Olares One, 多节点, 双节点, 升级, 1.12.5, 1.12.6
---

# 升级双节点 Olares 集群

:::warning
当前文档由 AI 翻译生成，若发现术语或表述不准确，请查看[英文原文](../../one/upgrade-multi-node-cluster.md)。
:::

本教程介绍如何手动将双节点 Olares 集群从 1.12.5 升级到 1.12.6。升级过程包括在两个节点上下载升级包、升级 Olares CLI 和守护进程，以及分别升级 master 节点和 worker 节点。

## 准备工作

**集群**
- 你有一个双节点 Olares 集群，包含一个 master 节点和一个 worker 节点。
- 两个节点都已开机，并连接到同一局域网。
- 你知道两个节点的本地 IP 地址。

**访问**
- 你可以通过 SSH 以具有 `sudo` 权限的用户访问两个节点。

## 步骤 1：连接到两个节点

在你的电脑上打开两个独立的终端窗口或标签页，分别通过 SSH 连接到两个节点。

1. 在第一个窗口中，使用 master 节点的本地 IP 地址通过 SSH 连接：

   ```bash
   ssh olares@<master-node-ip-address>
   ```

2. 在第二个窗口中，使用 worker 节点的本地 IP 地址通过 SSH 连接：

   ```bash
   ssh olares@<worker-node-ip-address>
   ```

3. 在整个过程中保持两个 SSH 连接窗口处于打开状态。某些步骤需要在两个节点上分别执行，某些步骤则需要在两个节点都可访问时执行。

## 步骤 2：在两个节点上下载升级包

先下载升级文件，但暂不安装。在 master 和 worker 两个 SSH 窗口中分别执行以下命令。

1. 切换到 root 用户：

   ```bash
   sudo su
   ```

2. 加载 Olares 环境变量：

   ```bash
   source /etc/olares/release
   ```

3. 创建升级目标文件以触发下载：

   ```bash
   echo '{"version":"1.12.6", "downloadOnly": true}' > $OLARES_BASE_DIR/upgrade.target
   ```

4. 检查下载进度：

   ```bash
   curl localhost:18088/system/status | jq | grep -i download
   ```

5. 等待 `upgradingDownloadState` 的值变为 `completed`。

:::warning 等待下载完成
进入步骤 3 之前，请确保两个节点上的 `upgradingDownloadState` 都显示为 `completed`。
:::

## 步骤 3：在两个节点上升级 Olares CLI 和守护进程

下载完成后，导入新镜像并更新核心管理工具，包括 Olares CLI 和 `olaresd` 守护进程。在 master 和 worker 两个 SSH 窗口中分别执行以下命令。

1. 确保你当前仍以 root 用户登录：

   ```bash
   sudo su
   ```

2. 确保环境变量已加载：

   ```bash
   source /etc/olares/release
   ```

3. 将 Olares CLI 更新到新版本：

   ```bash
   cp -f $OLARES_BASE_DIR/pkg/components/olares-cli-v1.12.6 /usr/local/bin/olares-cli
   ```

4. 导入新容器镜像：

   ```bash
   olares-cli prepare images
   ```

5. 升级 `olaresd` 守护进程：

   ```bash
   olares-cli prepare olaresd
   ```

## 步骤 4：升级 master 节点

两个节点都准备好后，先升级 master 节点。

1. 在 master 节点的 SSH 窗口中，开始升级：

   ```bash
   sudo olares-cli upgrade
   ```

   升级脚本执行完毕后，master 节点会自动重启，SSH 连接会断开。

2. 重启后，重新通过 SSH 连接到 master 节点：

   ```bash
   ssh olares@<master-node-ip-address>
   ```

3. 验证所有系统服务是否已成功重启并处于 `Running` 状态：

   ```bash
   kubectl get pod -o wide -A
   ```

4. 临时将集群版本号改回 `1.12.5`，以便升级 worker 节点：

   ```bash
   kubectl patch terminus terminus --type=merge -p '{"spec":{"version":"1.12.5"}}'
   ```

## 步骤 5：升级 worker 节点

master 节点升级完成并临时修改版本号后，即可升级 worker 节点。

1. 在 worker 节点的 SSH 窗口中，开始升级：

   ```bash
   sudo olares-cli upgrade
   ```

2. 升级完成后，worker 节点会重启。等待重启完成。

## 步骤 6：在 master 节点上恢复集群版本

worker 节点升级完成后，将 master 节点上的版本号恢复为 `1.12.6`，以完成整个升级过程。

如果到 master 节点的 SSH 连接已超时，请先重新连接再继续。

1. 在 master 节点的 SSH 窗口中，执行以下命令：

   ```bash
   kubectl patch terminus terminus --type=merge -p '{"spec":{"version":"1.12.6"}}'
   ```

   此时，双节点集群已运行在 Olares 1.12.6 上。

2. 要验证集群中两个节点的最终状态，请执行以下命令：

   ```bash
   kubectl get nodes
   ```
