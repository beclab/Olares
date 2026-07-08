---
outline: [2, 3]
description: 部署可扩展 Olares 集群的技术指南。了解如何配置主节点、解决主机名冲突，以及将工作节点加入集群。
head:
  - - meta
    - name: keywords
      content: Cluster, Kubernetes, Multi-node, Worker node, Master node
---

:::warning
本文档由 AI 自动翻译，可能存在表述差异。如需核对，请参考[英文原文](../../one/connect-two-olares-one.md)。
:::

# 设置多节点 Olares 集群 <Badge type="tip" text="1.5 h" />
对于需要更高可用性和分布式存储的高级用例，你可以将两台 Olares One 设备物理连接，组成一个统一的集群。

## 学习目标
- 配置支持分布式存储的主节点。
- 解决节点之间的主机名冲突。
- 使用 `joincluster.sh` 将工作节点加入集群。

## 开始之前
- Olares One 的默认用户名和密码均为 `olares`。
  :::warning 重置默认 SSH 密码
  即使主要使用 Control Hub 终端，你也必须立即在 **Settings** > **My hardware** 中重置此密码，以防止未授权访问。
  :::
- SSH 访问赋予了对系统的强大控制权。请确保妥善保管你的凭据。

## 前提条件
**硬件**<br>
- 两台连接到同一局域网的 Olares One 设备。
- 你知道两台设备的本地 IP 地址。

**经验**<br>
- 熟悉终端命令。
- 对 Kubernetes 节点管理有基本了解。

**软件**<br>
- 手机上已安装 LarePass。

## 步骤 1：设置主节点
:::danger 需要全新安装
设置集群需要干净的环境。如果该设备上已安装 Olares OS，必须先卸载：
```bash
sudo olares-cli uninstall
```
:::

第一台 Olares One 设备将作为主节点。

1. 使用本地 IP 地址通过 SSH 访问主节点。
    ```bash
    ssh olares@<Master-IP-Address>
    ```
2. 初始化本地存储服务 MinIO，它是分布式文件系统的后端。
    ```bash
    sudo olares-cli install storage
    ```
3. 安装启用 JuiceFS 的 Olares。这允许数据在多个节点之间共享。
    ```bash
    sudo olares-cli install --with-juicefs=true
    ```
4. 安装脚本会提示你输入 Olares ID 详情。

   例如，如果你的完整 Olares ID 是 `alice123@olares.com`：

   - **Domain name**：按 `Enter` 使用默认域名，或输入 `olares.com`。
   - **Olares ID**：输入 Olares ID 的前缀。在此示例中，输入 `alice123`。

   安装完成后，屏幕上将显示初始系统信息，包括 Wizard URL 和初始登录密码。在后续激活阶段你会需要它们。

   ![Wizard URL](/images/manual/get-started/wizard-url-and-login-password.png)

5. 使用 Wizard URL 和初始一次性密码进行激活。此过程通过 LarePass 将 Olares 设备与你的 Olares ID 关联。

   a. 在浏览器中输入 Wizard URL。你将进入欢迎页面。按任意键继续。
   ![打开 wizard](/images/manual/get-started/open-wizard.png#bordered)

   b. 按照屏幕上的指示完成激活。

设置完成后，LarePass 应用返回主屏幕，Wizard 将你重定向到 Olares 登录页面。

## 步骤 2：设置工作节点

:::info 加入前的工作节点要求
工作节点必须处于以下状态之一：

- **出厂状态（预装 Olares）**：Olares 已预装，且版本与主节点匹配。

  如果版本不再匹配（例如，主节点已升级到 v1.12.5，而工作节点仍运行 v1.12.4），请先运行 `sudo olares-cli uninstall --all` 将工作节点擦除为干净的 Linux 状态。

- **干净 Linux**：未安装 Olares。
:::

1. 通过 SSH 访问工作节点。
    ```bash
    ssh olares@<Worker-IP-Address>
    ```
2. 更新主机名：
    ```bash
    sudo hostnamectl set-hostname olares-worker
    ```
    :::info
    默认情况下，所有 Olares One 设备的主机名都是 `olares`。Kubernetes 要求集群中每个节点的主机名唯一。在将工作节点加入集群之前，必须确保它具有唯一的主机名。
    :::
3. 下载 `joincluster.sh`：

    ::: code-group
    ```bash [curl]
    # 此命令适用于已安装 curl 的用户。
    curl -fsSL https://raw.githubusercontent.com/beclab/Olares/refs/heads/main/build/base-package/joincluster.sh -o joincluster.sh
    ```

    ```bash [wget]
    # 此命令适用于已安装 wget 的用户。
    wget https://raw.githubusercontent.com/beclab/Olares/refs/heads/main/build/base-package/joincluster.sh
    ```
    :::
4. 使用主节点信息运行脚本：
    ```bash
    export MASTER_HOST=<Master-IP-Address> \
        MASTER_NODE_NAME=olares \
        MASTER_SSH_USER=olares \
        MASTER_SSH_PASSWORD=<Password>
    ./joincluster.sh
    ```

   脚本会自动检测工作节点上是否已安装 Olares，针对主节点运行预安装检查，然后加入集群。

## 步骤 3：验证集群

加入命令完成后，验证节点是否通信正常。

使用以下命令检查 Kubernetes 集群中所有节点的状态：
```bash
kubectl get nodes
```

示例输出：
```bash
NAME            STATUS   ROLES                         AGE   VERSION
olares          Ready    control-plane,master,worker   2h    v1.33.3+k3s1
olares-worker   Ready    worker                        50m   v1.33.3+k3s1
```

## 资源
- [Olares CLI](../developer/install/cli/node.md)：探索 Olares CLI。
- [Olares 环境变量](../developer/install/environment-variables.md)：了解支持 Olares 高级配置的环境变量。
