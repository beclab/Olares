---
outline: [2, 3]
description: 使用 CLI 收集 Olares 系统日志，从主机下载压缩包，并安全地提供诊断信息。
head:
  - - meta
    - name: keywords
      content: Olares, 技术支持, 收集日志, 导出日志, 系统日志, Pod 日志
---

# 收集诊断信息

如果故障排查指南未能解决问题，请收集相关上下文和日志，供后续定位原因。如果可以使用「设置」且只需要系统日志压缩包，也可以改为[在「设置」中导出系统日志](help/request-technical-support.md)。

## 记录问题上下文

收集日志前，请先记录以下信息：

- 问题发生的准确时间和时区。
- Olares 版本和相关应用版本。
- 问题发生时的操作、预期结果和实际现象。
- 能复现问题的最短操作步骤。
- 完整的错误信息，以及截图（如果有）。
- 问题发生前是否更新过 Olares 或应用、调整过网络，或重启过设备。

这些信息可以帮助排查人员将日志中的事件与用户看到的问题对应起来。

## 连接到 Olares 主机

通过 SSH 连接到设备。在 Olares One 上，SSH 用户为 `olares`：

```bash
ssh olares@{local_ip_address}
```

对于自行安装的 Olares 设备，请使用具有 `sudo` 权限的 Linux 账户。

## 收集并保存日志

在 Olares 主机上运行以下命令：

```bash
mkdir -p "$HOME/olares-logs"
sudo olares-cli logs --output-dir "$HOME/olares-logs"
```

该命令会收集 Olares 系统、Kubernetes、容器、内核和网络的近期诊断信息。完成后，最后一行会输出压缩包的完整路径，例如：

```plain
logs have been collected and archived in: /home/olares/olares-logs/olares-logs-20260827-050839.tar.gz
```

如果问题刚发生不久，可以通过指定时间范围减小压缩包体积，例如 `--since 3h`。

## 下载压缩包

在自己的电脑上，复制上一步输出的文件。将 `{local_ip_address}` 和 `{timestamp}` 替换为实际值：

```bash
scp olares@{local_ip_address}:~/olares-logs/olares-logs-{timestamp}.tar.gz .
```

末尾的句点表示将压缩包保存到电脑的当前目录。如果使用其他 SSH 账户，请将 `olares` 替换为对应的账户名。

## 安全提供日志

:::warning 日志可能包含隐私系统信息
系统日志可能包含 Olares 标识符、主机名、IP 地址、网络配置、文件路径和应用元数据。请勿将完整压缩包附加到公开的 GitHub Issue 或 Discussion 中。
:::

1. 对于不需要日志的一般性问题，请提交 [GitHub Discussion](https://github.com/beclab/Olares/discussions/new?category=q-a)。
2. 对于可复现的软件缺陷，请提交 [GitHub Issue](https://github.com/beclab/Olares/issues/new)，并附上问题上下文，但不要包含完整日志压缩包和敏感信息。
3. 当 Olares 团队要求提供日志时，请通过对方提供的非公开渠道发送，或发送至邮箱 [hi@olares.com](mailto:hi@olares.com)。如有相关 GitHub issue 编号，请一并注明。

发送前，请删除描述和截图中的密码、恢复短语、API 密钥、访问令牌等敏感信息。Olares 团队在排查问题时不需要你的密码或恢复短语。
