---
outline: [2, 3]
description: 在 Olares 设备终端中使用常用 Olares CLI 命令检查系统状态、收集日志、重置密码和管理存储。
head:
  - - meta
    - name: keywords
      content: Olares, olares-cli, 设备终端, 系统日志, 重置密码, 磁盘存储, GPU 状态
---

# Olares CLI 常用命令

Olares OS 在 `/usr/local/bin/olares-cli` 中内置了 `olares-cli`。本页介绍系统维护和故障排查时，最常需要你在 Olares 设备终端中亲自运行的命令。

如需通过自然语言管理文件、应用、设置等 Olares 功能，请参阅[通过 AI Agent 使用 Olares CLI](../../cli-install.md)。

## 准备工作

- 在你管理的设备上[访问 Olares 终端](../../../manual/access-olares-terminal.md)。
- 需要系统权限的命令请使用 `sudo` 运行。
- 运行 `olares-cli help` 或 `olares-cli <command> --help`，查看当前安装版本提供的命令和选项。

## 检查系统状态

检查 Olares 服务是否正常运行，并查看当前安装或升级状态：

```bash
olares-cli status
```

如果设备配有 NVIDIA GPU，可检查驱动和服务状态：

```bash
olares-cli gpu status
```

## 收集系统日志

将系统、Kubernetes、容器、内核和网络诊断信息收集到一个压缩包中：

```bash
mkdir -p "$HOME/olares-logs"
sudo olares-cli logs --max-lines 0 --output-dir "$HOME/olares-logs"
```

命令输出的最后一行会显示压缩包路径。日志可能包含系统隐私信息，请只通过私密支持渠道提供。

了解更多：[收集诊断信息](../../../manual/collect-diagnostic-information.md)。

## 重置忘记的桌面密码

将 `<username>` 替换为 Olares ID 中 `@` 之前的部分：

```bash
olares-cli user reset-password <username> -p <new-password>
```

如果命令返回权限错误，请按照[忘记桌面登录密码](../../../manual/help/ts-forget-login-password.md)中的说明检查错误，并选择下一步操作。

## 释放磁盘空间

在 Olares 1.12.6 及以上版本中，列出未被标准工作负载引用的容器镜像：

```bash
olares-cli doctor images --unused
```

删除前请逐项检查结果。自定义工作负载仍可能依赖列表中的镜像。

了解更多：[释放磁盘空间](../../../manual/free-up-disk-space.md)。

## 扩展系统存储

对于基于 LVM 的 Olares 安装，先检查哪些磁盘可以加入系统卷：

```bash
olares-cli disk list-unmounted
```

:::warning 所选磁盘上的数据会被清除
扩容流程会清除所选磁盘上的所有数据，并将磁盘合并到 Olares 系统卷中。请先备份磁盘数据。
:::

请按照[扩展 Olares 系统存储](../../../manual/best-practices/expand-storage-in-olares.md)中的步骤识别目标磁盘、执行扩容并检查结果。

## 相关指南

以下 Olares CLI 命令属于较长的操作流程，不应脱离对应指南单独运行：

- [安装特定 CUDA 版本](../../../manual/best-practices/install-specific-cuda-version.md)：自动配置无法满足需求时，安装或启用 GPU 支持。
- [通过命令行安装并激活 Olares](../../../manual/best-practices/activate-olares-using-cli.md)：在设备终端中初始化新设备。
- [搭建 Olares 多节点集群](../../../manual/best-practices/install-olares-multi-node.md)：启用共享存储并添加工作节点。
- [升级双节点 Olares 集群](../../../one/upgrade-multi-node-cluster.md)：手动协调主节点和工作节点的升级。
