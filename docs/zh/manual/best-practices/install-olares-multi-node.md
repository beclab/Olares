---
outline: [2, 3]
description: 使用 Olares One、NVIDIA DGX Spark 或其他兼容的 Linux 设备，将单节点 Olares 扩展为多节点集群。
head:
  - - meta
    - name: keywords
      content: Olares One, NVIDIA DGX Spark, 多节点集群, JuiceFS, 工作节点, 主节点
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../../manual/best-practices/install-olares-multi-node.md)。
:::

# 设置多节点 Olares 集群 <Badge type="warning" text="Preview" />

对于需要更多计算资源和分布式存储的工作负载，你可以为现有单节点 Olares 添加工作节点，组成多节点集群。此流程适用于 Olares One、NVIDIA DGX Spark，以及其他满足 Olares 系统要求的 Linux 设备。本文以 Olares One 作为主节点、NVIDIA DGX Spark 作为工作节点进行演示。

:::warning 预览功能
此流程需要使用尚未正式发布的 Olares 1.12.7。请勿在 Olares 1.12.6 上执行以下步骤。

如需在 Olares 1.12.6 上使用两台 Olares One 组建集群，请参阅[当前手动设置流程](../../one/connect-two-olares-one.md)。

不要使用本文步骤升级现有多节点集群。请按照当前 Olares 版本对应的更新指南操作。
:::

## 学习目标

完成本指南后，你将学会：

- 在现有 Olares 主节点上启用 JuiceFS。
- 生成并妥善保管工作节点加入命令。
- 添加工作节点，并检查集群是否就绪。

## 开始之前

开始前，先确定两台设备的角色：

- 现有单节点 Olares 系统作为 `master`（主节点）。
- 准备加入的设备作为 `worker`（工作节点）。

为两台设备分别打开一个终端窗口。在整个设置过程中，让每个窗口始终连接同一台设备。

:::warning 在正确的节点上运行命令
操作过程中需要在主节点和工作节点的终端之间切换。运行每条命令前，请检查终端提示符中的 hostname（主机名）。如果在错误的节点上运行迁移或卸载命令，可能会导致 Olares 无法正常使用。
:::

## 前提条件

确保满足以下条件：

- 主节点已安装并激活 Olares 1.12.7 或更高版本，且 Olares 正在运行。
- 可以通过 SSH 访问两台设备并运行 `sudo` 命令。主节点已启用 SSH 密码登录。
- 工作节点可以访问主节点的 SSH 地址，以及主节点生成的加入脚本地址。

:::warning 备份主节点
启用 JuiceFS 会停止 Olares，并迁移其本地文件系统。继续前请备份重要数据。迁移过程中不要关闭主节点。
:::

## 步骤 1：启用 JuiceFS <Badge type="tip" text="在主节点上操作" />

多节点集群需要使用 JuiceFS 提供共享文件系统。

1. 通过 SSH 连接主节点。

   对于已经运行 Olares 的设备，可以在 LarePass 的 **Settings** > **System** > 设备 > **Network** > **Intranet IP** 中查看其局域网 IP 地址。如果使用 Olares One，但不知道如何获取 SSH 密码，请参阅[通过 SSH 访问 Olares One](../../one/access-terminal-ssh.md)。

   ```bash
   ssh <用户名>@<主节点 IP 地址>
   ```

2. 停止 Olares。

   ```bash
   sudo olares-cli stop
   ```

3. 启用 JuiceFS，并迁移现有 Olares 文件系统。

   ```bash
   sudo olares-cli node enable-juicefs
   ```

   迁移完成后，Olares 会自动重启。等待命令提示 JuiceFS 已启用，且主节点可以接受工作节点。

   输出示例：

   ```plain
   [Job] Enable JuiceFS on the master node and migrate rootfs execute successfully!!! (...)
   JuiceFS is enabled on this master node (version <version>); it is now ready to accept worker nodes.
   Run 'olares-cli node join-command' on this master to print the command for a worker node.
   ```

## 步骤 2：生成工作节点加入命令 <Badge type="tip" text="在主节点上操作" />

1. 在主节点上运行：

   ```bash
   sudo olares-cli node join-command
   ```

2. 检查检测到的 SSH 地址和用户名。输入 `y` 确认；如果信息不正确，请按提示输入正确的用户名。

3. 输入该用户的 SSH 密码。

   Olares 会检查工作节点能否使用此账号连接主节点并运行 `sudo` 命令。如果检查失败，请确认 IP 地址、用户名、密码、SSH 服务和防火墙设置，然后重试。

4. 复制生成的命令。下一步需要在工作节点上运行它。

   输出示例：

   ```plain
   SSH login and sudo access verified for <用户名>@<主节点 IP 地址>:22.

   Run the following command on the worker node:

   export MASTER_AUTH_INFO='<编码后的认证信息>' OLARES_SYSTEM_CDN_SERVICE='<Olares CDN 地址>' && curl -fsSL '<加入脚本地址>' | bash

   MASTER_AUTH_INFO is Base64-encoded, not encrypted. Anyone holding this command can recover the master's SSH credentials, so share it only with the intended worker administrator.
   ```

:::danger 保护加入命令
生成的 `MASTER_AUTH_INFO` 只经过 Base64 编码，并未加密。拿到命令的人可以还原主节点的 SSH 凭据。请只将命令提供给目标工作节点的管理员，切勿发布到公开群聊或代码仓库。
:::

## 步骤 3：添加工作节点 <Badge type="warning" text="在工作节点上操作" />

:::warning 检查工作节点状态
工作节点上不能有完整安装的 Olares。如果已安装 Olares，请先备份重要数据，再运行 `sudo olares-cli uninstall`。如果加入命令检测到已有安装，它会停止并显示该操作提示。
:::

1. 通过 SSH 连接工作节点。

2. 在工作节点上，粘贴并运行步骤 2 生成的完整命令。命令格式如下：

   ```bash
   export MASTER_AUTH_INFO='<编码后的主节点连接信息>' \
     OLARES_SYSTEM_CDN_SERVICE='<Olares CDN 地址>' \
     && curl -fsSL '<加入脚本地址>' | bash
   ```

   请完整使用系统生成的命令，不要复制上面的占位示例。

3. 如果命令提示已安装 Olares，请先卸载：

   ```plain
   Joining this machine to an Olares cluster as a worker node, using the Olares <version> installer.
   error: Olares <version> is already installed on this node; run 'sudo olares-cli uninstall' before joining it to another cluster
   ```

   ```bash
   sudo olares-cli uninstall
   ```

   卸载完成后，再次运行生成的加入命令。

4. 加入流程默认使用工作节点当前的 hostname（主机名）作为节点名称。如果主机名格式有效且在集群中没有重名，无需进行任何操作。

   如果当前主机名不可用，请按提示输入新名称。加入流程会自动更新工作节点的主机名。

   主机名必须：

   - 只包含小写字母、数字、连字符（`-`）或句点（`.`）。
   - 以字母或数字开头和结尾。
   - 在集群中保持唯一。

   例如，可以使用 `olares-worker`。

加入流程会检查主节点连接、准备匹配的 Olares 版本，并将工作节点添加到集群。

工作节点成功加入后，输出末尾会显示类似以下内容：

```plain
[Job] Add Worker Node To The Cluster execute successfully!!! (...)

This node joined the Olares cluster at <主节点 IP 地址> as "<工作节点名称>".
Verify it on the master with: sudo /usr/local/bin/kubectl get nodes
```

## 步骤 4：验证集群 <Badge type="tip" text="在主节点上操作" />

Olares 使用 Kubernetes 管理集群。在主节点上运行以下 `kubectl` 命令，检查所有节点的状态：

```bash
sudo /usr/local/bin/kubectl get nodes
```

确认主节点和工作节点均已显示，且状态为 `Ready`。

示例输出：

```plain
NAME          STATUS   ROLES                         AGE   VERSION
master-node   Ready    control-plane,master,worker   1h    v1.33.3+k3s1
worker-node   Ready    worker                        1m    v1.33.3+k3s1
```

## 资源

- [Olares 环境变量](../../developer/install/environment-variables.md)：了解用于 Olares 高级配置的环境变量。
- [在 Olares 1.12.6 上使用两台 Olares One 组建集群](../../one/connect-two-olares-one.md)：Olares 1.12.7 正式发布前，请使用当前手动流程。
