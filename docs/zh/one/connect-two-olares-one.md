---
title: 将两台 Olares One 连接为集群
outline: [2, 3]
description: 在现有 Olares One 上启用 JuiceFS，生成工作节点加入命令，将两台设备组成集群并检查节点状态。
head:
  - - meta
    - name: keywords
      content: Olares One, 多节点集群, Kubernetes, JuiceFS, 工作节点, 主节点
---

:::warning
本文档由 AI 自动翻译，仅供参考。涉及关键操作或信息时，请以[英文原文](../../one/connect-two-olares-one.md)为准。
:::

# 连接两台 Olares One <Badge type="tip" text="1.5 h" />

对于需要更多计算资源和分布式存储的高级场景，你可以连接两台 Olares One，组成一个统一的集群。

:::warning 需要 Olares 1.12.7 或更高版本
主节点必须已激活，以单节点模式运行 Olares 1.12.7 或更高版本。开始前请[备份重要数据](../manual/olares/settings/backup.md)，文件系统迁移期间不要关闭主节点。
:::

## 学习目标

完成本指南后，你将学会：

- 在主节点上启用共享存储。
- 生成并妥善保管工作节点加入命令。
- 解决主机名冲突，将工作节点加入集群。
- 验证两个节点是否就绪。

## 开始之前

- 出厂 SSH 用户名和密码均为 `olares`。设备激活后，请使用 LarePass Vault 中保存的 SSH 密码。请参阅[通过 SSH 访问 Olares One](./access-terminal-ssh.md#步骤-2从-vault-获取登录密码)。
- SSH 访问权限允许你控制系统，请妥善保管登录凭据。

## 前提条件

**硬件**<br>
- 两台 Olares One 已连接同一局域网，并可联网下载加入脚本和安装文件。
- 一台有 SSH 客户端的电脑。

**经验**<br>
- 熟悉终端命令和 Kubernetes 节点管理的基础操作。

**软件**<br>
- 手机已安装 LarePass。
- 可以通过 SSH 访问两台设备并运行 `sudo` 命令。工作节点可以使用密码通过 SSH 登录主节点。

## 步骤 1：设置主节点

### 启用 JuiceFS

现有已激活的设备作为主节点。JuiceFS 为集群提供共享存储，迁移命令默认安装 MinIO 作为存储后端，并迁移现有数据。

1. 使用局域网 IP 地址连接主节点。

   在 LarePass 中，前往**设置** > **系统** > 你的设备 > **节点** > **网络** > **内网 IP**，查看地址。使用 LarePass Vault 中保存的 SSH 密码。

   ```bash
   ssh olares@<主节点 IP 地址>
   ```

2. 停止 Olares：

   ```bash
   sudo olares-cli stop
   ```

3. 启用 JuiceFS，并迁移现有文件系统：

   ```bash
   sudo olares-cli node enable-juicefs
   ```

   迁移时间取决于数据量。数据较多时，可能需要较长时间，期间不显示迁移进度。请保持终端打开，等待迁移完成。

   迁移完成后，Olares 会自动重启。等待命令提示 JuiceFS 已启用、主节点可以接受工作节点后，再继续操作。

   输出示例：

   ```plain
   [Job] Enable JuiceFS on the master node and migrate rootfs execute successfully!!! (...)
   JuiceFS is enabled on this master node (version <version>); it is now ready to accept worker nodes.
   Run 'olares-cli node join-command' on this master to print the command for a worker node.
   ```

### 生成工作节点加入命令

1. 在主节点终端中运行：

   ```bash
   sudo olares-cli node join-command
   ```

2. 检查显示的 SSH 地址和用户名。输入 `y` 确认，或按提示修改用户名。默认用户名为 `olares`。
3. 输入 SSH 密码。命令会在主节点本机检查 SSH 登录和 `sudo` 权限。如果认证失败，会再次提示输入用户名和密码。按 `Enter` 保留提示中的用户名，或输入其他用户名。如果仍无法连接，请检查地址、SSH 服务和防火墙设置。
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

## 步骤 2：设置工作节点

第二台设备作为工作节点。单独打开一个终端窗口，继续前检查设备的 IP 地址。

:::warning 工作节点需要干净的安装环境
工作节点上已有的 Olares 必须先卸载，包括出厂预装的系统。卸载前请[备份重要数据](../manual/olares/settings/backup.md)。不要卸载主节点上的 Olares。
:::

1. 使用局域网 IP 地址连接工作节点：

   ```bash
   ssh olares@<工作节点 IP 地址>
   ```

2. 在工作节点上，粘贴并运行[生成工作节点加入命令](#生成工作节点加入命令)中得到的完整命令。命令格式如下：

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

4. 如果出现提示，输入唯一的主机名，例如 `olares-worker`。两台设备默认都使用 `olares`，但 Kubernetes 要求每个节点的名称唯一。如果当前主机名格式有效且没有重名，则无需修改。

   主机名必须：

   - 只包含小写字母、数字、连字符（`-`）或句点（`.`）。
   - 以字母或数字开头和结尾。
   - 在集群中保持唯一。

加入流程会检查与主节点的连接、按需更新主机名、准备匹配的 Olares 版本，并将工作节点添加到集群。

如果工作节点仅完成安装准备（prepare），尚未完整安装 Olares，加入流程会复用匹配的版本。如果准备的版本与主节点不同，会自动清理旧的准备状态，并准备匹配的版本。

等待命令显示类似以下内容的成功提示：

```plain
[Job] Add Worker Node To The Cluster execute successfully!!! (...)

This node joined the Olares cluster at <主节点 IP 地址> as "<工作节点名称>".
Verify it on the master with: sudo /usr/local/bin/kubectl get nodes
```

## 步骤 3：验证集群

加入命令完成后，在主节点上运行以下命令，检查两个节点的状态：

```bash
sudo /usr/local/bin/kubectl get nodes
```

确认两个节点均已显示，且状态为 `Ready`。例如：

```plain
NAME            STATUS   ROLES                         AGE   VERSION
olares          Ready    control-plane,master,worker   1h    v1.33.3+k3s1
olares-worker   Ready    worker                        1m    v1.33.3+k3s1
```

## 资源

- [设置多节点 Olares 集群](../manual/best-practices/install-olares-multi-node.md)：查看适用于 Linux 设备的通用设置指南。
- [Olares 环境变量](../developer/install/environment-variables.md)：了解高级配置选项。
