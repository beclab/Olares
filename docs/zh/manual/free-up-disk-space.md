---
outline: [2, 3]
description: 通过删除不用的 AI 模型，并在 Olares 1.12.6 中谨慎清理未引用的容器镜像，为 Olares 释放磁盘空间。
head:
  - - meta
    - name: keywords
      content: Olares, 磁盘空间, 存储已满, 镜像清理, 模型清理, crictl, Hugging Face, Ollama
---

# 释放 Olares 磁盘空间

本指南介绍如何删除不再需要的 AI 模型和 Olares 不再引用的容器镜像，以释放存储空间。磁盘用量过高、无法安装应用，或登录时出现 **Authentication failed, disk space is full** 时，也可按照本文操作。

先清理你能明确识别的模型文件。容器镜像必须先查看未使用列表，确认后再删除。停止应用不会自动删除其模型文件或容器镜像，因为应用恢复时可能还需要使用这些内容。

## 前提条件

- 拥有 Olares 的管理员访问权限。
- 拥有 Olares 主机的终端访问权限，用于移除未使用的镜像。请参考[通过 SSH 连接 Olares](/zh/manual/access-olares-terminal)。

## 检查磁盘用量

1. 打开**仪表盘**，选择**磁盘**卡片。
2. 查看已用和可用空间。
3. 点击**占用分析**，确认哪个文件系统空间不足。

详细说明请参考[监控资源使用情况](olares/resources-usage.md#磁盘面板)。

## 删除不用的 AI 模型

:::warning 删除模型后无法撤销
只删除你能确认且不再需要的模型。使用该模型的应用将无法继续工作，直到重新下载模型。
:::

1. 打开**文件**。
2. 检查系统中存在的模型目录：
   - 在 Olares 1.12.6 中，选择**应用** > **公共**，然后打开 `huggingface` 或 `ollama`。
   - 如果模型仍保存在旧版的用户目录中，选择**首页**，查找与应用同名的目录，例如 `Huggingface` 和 `Ollama`。
3. 确认哪些模型已不再被任何应用使用。
4. 右键点击模型目录或文件，然后选择**删除**。

删除单个模型时，请保留所需的目录结构。详细说明请参考[管理共享 AI 模型](olares/files/files-common.md#管理模型文件)。

## 查看未使用的容器镜像

Olares 1.12.6 内置了检查命令，可以列出未被标准工作负载引用的本地容器镜像。删除前先运行：

```bash
olares-cli doctor images --unused
```

逐项检查输出结果。命令会按镜像大小从大到小排列候选项，并显示预计可以释放的空间。

在多节点 Olares 集群中，该列表只包含运行命令的控制节点上保存的镜像，不会统计仅保存在工作节点上的镜像。

:::warning 继续前先检查自定义工作负载
未使用镜像检查会读取 Deployment、StatefulSet、DaemonSet、Job 和 CronJob 等标准工作负载。它可能无法识别仅由裸 Pod、静态 Pod、自定义控制器使用的镜像，或正在运行的容器所固定的旧镜像摘要。如果你创建过自定义工作负载，请勿在未核实的情况下运行删除管道。
:::

## 删除已确认不用的容器镜像

确认列表中只有可以重新下载的镜像后，运行：

```bash
olares-cli doctor images --unused --no-headers \
  | awk '$1 ~ /^[0-9a-f]+$/ && length($1) == 12 {print $1}' \
  | xargs -r sudo crictl rmi
```

过滤条件只会把 12 位十六进制镜像 ID 传给 `crictl`，不会把 `no unused images` 提示误当成镜像 ID。

删除镜像不会移除个人文件或应用数据。但是，恢复或重新安装相关应用时，Olares 需要重新下载镜像，因此所需时间可能更长。

## 验证可用空间

1. 再次运行 `olares-cli doctor images --unused`。
2. 返回**仪表盘** > **磁盘**，刷新用量数据。
3. 恢复需要使用的应用，确认应用可以正常启动。

如果清理后磁盘用量仍然很高，请参阅[扩展 Olares 系统存储](best-practices/expand-storage-in-olares.md)。

## 避免直接清理所有未运行镜像

:::danger 不建议使用
请勿将 `sudo crictl rmi --prune` 作为日常清理方式。它会删除所有已停止应用的镜像，而不是只删除通过 `olares-cli doctor images --unused` 检查过的候选项。这可能延长应用恢复时间，并导致 Olares 界面暂时显示不一致的镜像信息。
:::

只有在 Olares 客服明确要求，并且你了解已停止应用可能需要重新下载镜像时，才考虑使用该命令。
