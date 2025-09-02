---
outline: [2, 3]
description: Olares 多节点集群部署教程，包括主节点配置、工作节点添加和网络变更处理，助你搭建可扩展的分布式环境。
---

# 安装多节点 Olares 集群 <Badge type="warning" text="Alpha" />

默认情况下，Olares 的安装会部署单节点集群。从 v1.11.3 开始，Olares 支持添加子节点。本教程将指导你如何配置主节点并添加子节点，以创建一个可扩展的多节点 Olares 集群。

:::warning Alpha 功能
此功能目前处于 **Alpha** 阶段，可能存在性能问题并需要额外的手动配置，不建议用于生产环境。如果遇到任何问题，请在 [Olares 的 GitHub 仓库](https://github.com/beclab/Olares/issues)中提交 Issue。
:::
:::info 仅支持 Linux
当前仅支持 Linux 系统节点加入 Olares 集群。
:::

## 目标

通过本教程，你将学习：

- 在主节点上安装支持 JuiceFS 的 Olares。
- 向集群中添加子节点。
- 处理可能的网络变化，确保集群能够持续高效运行。

## 准备工作

在开始之前，请确保满足以下条件：

- 熟悉 Kubernetes 和系统管理。
- 主节点和子节点必须在同一个本地网络中。
- 主节点和子节点必须有唯一的主机名，以避免冲突。
- 子节点必须能够通过 SSH 连接到主节点。这意味着：
  - 如果使用 root 用户或具有 `sudo` 权限的用户：需要将子节点的 SSH 公钥添加到主节点的 `authorized_keys` 文件中。
  - 如果使用非 root 用户：需要在主节点上启用基于密码的 SSH 身份验证。

## 第一步：设置主节点

::: tip 卸载已有的 Olares 集群
如果你已经使用默认的安装命令在当前节点上安装了 Olares 集群，运行 `olares-cli uninstall --all` 命令将其卸载。
:::

在主节点上运行以下命令以启用 JuiceFS 支持：

```bash
export JUICEFS=1 \
&& curl -sSfL https://cn.olares.sh | bash -
```

此命令将安装 Olares，并内置一个 MinIO 实例作为后端存储。安装过程与单节点安装相同，系统会提示你输入域名并提供 Olares ID 的用户名。

:::tip 自定义存储
如果你已经有自己的 MinIO 集群，或有一个 S3（或 S3 兼容）存储桶，可以将 Olares 配置为使用这些存储，而不是内置的 MinIO 实例。
:::

## 第二步：向集群添加子节点
1. 在子节点上，使用以下方式下载 `joincluster.sh`：
::: code-group

```bash [curl]
# 使用 Curl 方式下载
curl -fsSL https://raw.githubusercontent.com/beclab/Olares/refs/heads/main/build/base-package/joincluster.sh -o joincluster.sh
```

```bash [wget]
# 使用 wget 方式下载
wget https://raw.githubusercontent.com/beclab/Olares/refs/heads/main/build/base-package/joincluster.sh
```
:::

2. 使用必要的环境变量运行 `joincluster.sh` 脚本。这些变量用于告诉子节点如何连接到主节点。必须要设置 `MASTER_HOST` 变量，该变量指定主节点的 IP 地址：
   ```bash
   export MASTER_HOST=192.168.1.15
   ./joincluster.sh
   ```

下面是可能需要设置的变量列表：

| **变量**                      | **描述**                                                                                                                |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `MASTER_HOST`                 | 主节点的 IP 地址。<br/>必填项。                                                                                         |
| `MASTER_NODE_NAME`            | 主节点的 Kubernetes 节点名称。<br/>如果未指定，脚本会交互式提示你确认所需参数。<br/>可选项。                            |
| `MASTER_SSH_USER`             | 用于通过 SSH 登录主节点的用户名。<br/>默认是 root。                                                                     |
| `MASTER_SSH_PASSWORD`         | SSH 用户的密码。<br/>如果未使用 SSH 密钥，则必填。                                                                      |
| `MASTER_SSH_PRIVATE_KEY_PATH` | 用于身份验证的私有 SSH 密钥路径。<br/>如果未指定，脚本会交互式提示你确认所需参数。<br/>默认路径为 `/root/.ssh/id_rsa`。 |
| `MASTER_SSH_PORT`             | 主节点 SSH 服务的端口号。<br/>默认值为 `22`。                                                                           |

:::info

- 非 root 用户必须提供密码以用 `sudo` 执行命令。因此，如果使用非 root 用户作为 `MASTER_SSH_USER`，且未指定 `MASTER_SSH_PASSWORD`，将无法继续。
- 使用 `export` 设置的环境变量会在当前终端会话中保持有效。切换不同配置时，需清除（`unset`）任何冲突的变量。
  `bash
  unset MASTER_SSH_PRIVATE_KEY_PATH
  `
  :::

## 使用示例

以下是一些实际示例，帮助你理解在不同场景下如何使用 `joincluster.sh` 脚本。

### 示例 1：默认设置

如果主节点的 IP 是 `192.168.1.15`，使用默认用户（`root`）和端口（`22`），主节点已在 `/root/.ssh/authorized_keys` 中包含当前节点的公钥 `/root/.ssh/id_rsa.pub`，运行：

```bash
export MASTER_HOST=192.168.1.15
./joincluster.sh
```

### 示例 2：自定义 SSH 密钥路径

如果主节点的 IP 是 `192.168.1.15`，SSH 端口是 `22`，用户是 `root`，而子节点使用位于 `/home/olares/.ssh/id_rsa` 的自定义 SSH 密钥，运行：

```bash
export MASTER_HOST=192.168.1.15 \
MASTER_SSH_PRIVATE_KEY_PATH=/home/olares/.ssh/id_rsa
./joincluster.sh
```

### 示例 3：使用非 root 用户和密码

如果主节点的 IP 是 `192.168.1.15`，SSH 端口是 `22`，用户是具有 `sudo` 权限的 `olares`，并且密码是 `olares`，运行：

```bash
export MASTER_HOST=192.168.1.15 \
MASTER_SSH_USER=olares \
MASTER_SSH_PASSWORD=olares
./joincluster.sh
```

## 卸载子节点

在子节点上运行以下命令：

```bash
olares-cli olares uninstall
```

## 处理网络变化

集群设置完成后，网络配置的变化可能会中断主节点与子节点的通信。

### 如果主节点网络发生变化

- **如果主节点切换到另一个局域网**：Olares 系统守护进程（olaresd）会检测到这一变化，触发 `olares-cli` 调用 `changeip` 命令。此时主节点将继续工作，但子节点无法与主节点通信，导致无法正常运行。

- **如果主节点的 IP 在同一局域网内发生变化**：子节点同样会失去通信，因为它们无法自动检测新的 IP。为解决此问题，可以在子节点上使用 `olares-cli` 命令更新主节点的 IP 地址并重启相关服务：

  ```bash
  sudo olares-cli olares change-ip -b /home/olares/.olares --new-master-host 192.168.1.18
  ```

  其中：

  - `-b /home/olares/.olares`：指定 Olares 的基础目录（默认值为 `$HOME/.olares`）。
  - `--new-master-host 192.168.1.18`：指定主节点的新 IP 地址。

### 如果子节点网络发生变化

- **如果子节点切换到另一个局域网**：子节点将失去与主节点的通信，无法正常运行。

- **如果子节点的 IP 在同一局域网内发生变化**：olaresd 会自动将新 IP 上报给主节点，无需手动干预。

## 了解更多

- [Olares 系统架构](../concepts/system-architecture.md#分布式存储)：了解支持 Olares 的分布式文件系统，确保可扩展性、高可用性以及无缝的数据管理。
- [系统守护进程](../../developer/install/installation-overview.md#系统守护进程olaresd)：olaresd：了解 orchestrates 和管理 Olares 核心功能的中央系统进程。
- [数据](../concepts/data.md#juicefs)：探索 Olares 如何利用 JuiceFS 提供统一文件系统，实现高效的数据存储和检索。
- [Olares CLI](../../developer/install/cli/olares-cli.md)：深入了解用于管理 Olares 安装的命令行工具。
- [Olares 环境变量](../../developer/install/environment-variables.md)：了解支持 Olares 高级配置的环境变量。
- [安装 Olares](../get-started/install-olares.md)：了解安装与激活 Olares 的过程。
